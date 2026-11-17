package search

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

//go:embed keywords.json
var embeddedKeywordsJSON []byte

type KeywordsConfig struct {
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Keywords    []string `json:"keywords"`
}

// SearchResult represents a single item returned from a search query.
type SearchResult struct {
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
	URL     string `json:"url"`
}

// SearchResponse holds search results and query metadata.
type SearchResponse struct {
	Query     string         `json:"query"`
	Results   []SearchResult `json:"results"`
	Source    string         `json:"source"`
	Error     string         `json:"error,omitempty"`
	Credits   float64        `json:"credits,omitempty"` // Tavily credit usage
	UsageText string         `json:"usage_text,omitempty"`
}

// SearchProvider is the interface for different search implementations.
type SearchProvider interface {
	Name() string
	Search(ctx context.Context, query string) (*SearchResponse, error)
}

// MultiSearchService manages search providers with automatic fallbacks.
type MultiSearchService struct {
	httpClient *http.Client
	tavilyKey  string
	keywords   []string
}

// NewMultiSearchService creates a new resilient search service.
func NewMultiSearchService() *MultiSearchService {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	var kw []string
	if len(embeddedKeywordsJSON) > 0 {
		var cfg KeywordsConfig
		if err := json.Unmarshal(embeddedKeywordsJSON, &cfg); err == nil {
			kw = cfg.Keywords
		}
	}

	return &MultiSearchService{
		httpClient: client,
		keywords:   kw,
	}
}

// ShouldSearch checks whether the user query needs web grounding based on keyword match.
func (s *MultiSearchService) ShouldSearch(query string) bool {
	lower := strings.ToLower(strings.TrimSpace(query))
	if lower == "" {
		return false
	}

	for _, k := range s.keywords {
		kLower := strings.ToLower(strings.TrimSpace(k))
		if kLower != "" && strings.Contains(lower, kLower) {
			return true
		}
	}
	return false
}

// SetTavilyKey updates or configures the Tavily API key dynamically.
func (s *MultiSearchService) SetTavilyKey(key string) {
	s.tavilyKey = strings.TrimSpace(key)
}

// GetTavilyKey returns the current Tavily API key.
func (s *MultiSearchService) GetTavilyKey() string {
	return s.tavilyKey
}

// HasTavilyKey returns true if a Tavily API key is configured.
func (s *MultiSearchService) HasTavilyKey() bool {
	return s.tavilyKey != ""
}

// Search runs through providers with Tavily as the primary enterprise engine if configured.
func (s *MultiSearchService) Search(ctx context.Context, query string) (*SearchResponse, error) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return &SearchResponse{Query: query}, nil
	}

	lower := strings.ToLower(trimmed)
	isWeather := strings.Contains(lower, "cuaca") || strings.Contains(lower, "weather") || strings.Contains(lower, "suhu")

	// 1. If Tavily API Key is available, use Tavily AI Search (best accuracy for real-time news, bbm, facts, etc.)
	if s.tavilyKey != "" {
		tavilyProvider := &TavilyProvider{
			client: s.httpClient,
			apiKey: s.tavilyKey,
		}
		resp, err := tavilyProvider.Search(ctx, trimmed)
		if err == nil && len(resp.Results) > 0 {
			return resp, nil
		}
	}

	// 2. Weather Specialized Provider (wttr.in) for weather/cuaca queries
	if isWeather {
		weatherProvider := &WeatherProvider{client: s.httpClient}
		resp, err := weatherProvider.Search(ctx, trimmed)
		if err == nil && len(resp.Results) > 0 {
			return resp, nil
		}
	}

	// 3. Wikipedia Open Search Fallback
	wikiProvider := &WikipediaProvider{client: s.httpClient}
	resp, err := wikiProvider.Search(ctx, trimmed)
	if err == nil && len(resp.Results) > 0 {
		return resp, nil
	}

	return &SearchResponse{
		Query:   trimmed,
		Source:  "none",
		Results: nil,
	}, nil
}

// ==================== 1. TAVILY SEARCH PROVIDER ====================

type TavilyProvider struct {
	client *http.Client
	apiKey string
}

func (p *TavilyProvider) Name() string { return "Tavily AI Search" }

type TavilySearchRequest struct {
	Query          string   `json:"query"`
	SearchDepth    string   `json:"search_depth"`
	IncludeFavicon bool     `json:"include_favicon"`
	IncludeUsage   bool     `json:"include_usage"`
	IncludeAnswer  bool     `json:"include_answer"`
	MaxResults     int      `json:"max_results"`
}

type TavilySearchResponse struct {
	Answer  string `json:"answer"`
	Query   string `json:"query"`
	Usage   *struct {
		Credits float64 `json:"credits"`
	} `json:"usage"`
	Results []struct {
		Title   string  `json:"title"`
		URL     string  `json:"url"`
		Content string  `json:"content"`
		Score   float64 `json:"score"`
	} `json:"results"`
}

func (p *TavilyProvider) Search(ctx context.Context, query string) (*SearchResponse, error) {
	reqBody := TavilySearchRequest{
		Query:          query,
		SearchDepth:    "advanced",
		IncludeFavicon: false,
		IncludeUsage:   true,
		IncludeAnswer:  true,
		MaxResults:     5,
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.tavily.com/search", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tavily request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tavily error (status %d): %s", resp.StatusCode, string(body))
	}

	var tavilyResp TavilySearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&tavilyResp); err != nil {
		return nil, fmt.Errorf("failed to parse tavily response: %w", err)
	}

	var results []SearchResult
	// If Tavily generated a synthesized direct answer, put it first
	if tavilyResp.Answer != "" {
		results = append(results, SearchResult{
			Title:   "Tavily Synthesized Answer",
			Snippet: tavilyResp.Answer,
			URL:     "https://tavily.com",
		})
	}

	for _, item := range tavilyResp.Results {
		results = append(results, SearchResult{
			Title:   item.Title,
			Snippet: item.Content,
			URL:     item.URL,
		})
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("tavily returned 0 results")
	}

	credits := 1.0
	usageText := "1 credit (Tavily search)"
	if tavilyResp.Usage != nil && tavilyResp.Usage.Credits > 0 {
		credits = tavilyResp.Usage.Credits
		usageText = fmt.Sprintf("%.0f credits (Tavily search)", credits)
	}

	return &SearchResponse{
		Query:     query,
		Source:    "Tavily AI Search",
		Results:   results,
		Credits:   credits,
		UsageText: usageText,
	}, nil
}

// ==================== 2. WEATHER PROVIDER ====================

type WeatherProvider struct {
	client *http.Client
}

func (p *WeatherProvider) Name() string { return "Weather (wttr.in)" }

func (p *WeatherProvider) Search(ctx context.Context, query string) (*SearchResponse, error) {
	location := cleanLocationQuery(query)
	if location == "" {
		location = "Jakarta"
	}

	reqURL := fmt.Sprintf("https://wttr.in/%s?format=%%l:+%%C+%%t+(Humidity:+%%h,+Wind:+%%w)", url.PathEscape(location))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "curl/8.0")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	info := strings.TrimSpace(string(body))
	if info == "" || strings.Contains(info, "<html") {
		return nil, fmt.Errorf("invalid weather content")
	}

	return &SearchResponse{
		Query:  query,
		Source: "Weather (wttr.in)",
		Results: []SearchResult{
			{
				Title:   fmt.Sprintf("Live Weather Observation - %s", location),
				Snippet: fmt.Sprintf("Current conditions: %s", info),
				URL:     fmt.Sprintf("https://wttr.in/%s", url.PathEscape(location)),
			},
		},
	}, nil
}

// ==================== 3. WIKIPEDIA PROVIDER ====================

type WikipediaProvider struct {
	client *http.Client
}

func (p *WikipediaProvider) Name() string { return "Wikipedia" }

func (p *WikipediaProvider) Search(ctx context.Context, query string) (*SearchResponse, error) {
	cleanQ := cleanKeywords(query)
	searchURL := fmt.Sprintf("https://id.wikipedia.org/w/api.php?action=opensearch&search=%s&limit=3&format=json", url.QueryEscape(cleanQ))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "IFMChat/1.0 (Desktop App; +https://github.com/yogamahhendra)")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data []any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	if len(data) < 4 {
		return nil, fmt.Errorf("invalid wikipedia response")
	}

	titles, _ := data[1].([]any)
	descriptions, _ := data[2].([]any)
	urls, _ := data[3].([]any)

	var results []SearchResult
	for i := 0; i < len(titles); i++ {
		t, _ := titles[i].(string)
		d, _ := descriptions[i].(string)
		u, _ := urls[i].(string)
		if t != "" {
			results = append(results, SearchResult{
				Title:   t,
				Snippet: d,
				URL:     u,
			})
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no wikipedia results")
	}

	return &SearchResponse{
		Query:   query,
		Source:  "Wikipedia",
		Results: results,
	}, nil
}

// ==================== HELPERS & FORMATTING ====================

func cleanLocationQuery(q string) string {
	lower := strings.ToLower(q)
	lower = strings.ReplaceAll(lower, "dibali", "bali")
	lower = strings.ReplaceAll(lower, "dijakarta", "jakarta")
	lower = strings.ReplaceAll(lower, "dibandung", "bandung")
	lower = strings.ReplaceAll(lower, "disurabaya", "surabaya")
	lower = strings.ReplaceAll(lower, "diyogya", "yogyakarta")
	lower = strings.ReplaceAll(lower, "dijogja", "yogyakarta")

	removeWords := []string{
		"bisa", "kamu", "searching", "cari", "tolong", "cek", "info", "informasi",
		"cuaca", "weather", "suhu", "di", "in", "for", "hari", "ini", "besok", "lusa", "sekarang",
		"saat", "apakah", "bagaimana", "kira-kira", "dong", "ya", "yah", "please", "today", "now", "tomorrow",
	}

	words := strings.Fields(lower)
	var kept []string
	for _, w := range words {
		clean := strings.Trim(w, ",.?!'\"")
		isStop := false
		for _, rw := range removeWords {
			if clean == rw {
				isStop = true
				break
			}
		}
		if !isStop && clean != "" {
			kept = append(kept, clean)
		}
	}

	if len(kept) == 0 {
		return ""
	}
	return strings.Join(kept, " ")
}

func cleanKeywords(q string) string {
	lower := strings.ToLower(q)
	removeWords := []string{
		"bisa", "kamu", "searching", "cari", "tolong", "cek", "info", "informasi",
		"apa", "itu", "siapa", "dimana", "kapan", "mengapa", "bagaimana",
		"dong", "ya", "yah", "please", "jelaskan", "tentang",
	}

	words := strings.Fields(lower)
	var kept []string
	for _, w := range words {
		clean := strings.Trim(w, ",.?!'\"")
		isStop := false
		for _, rw := range removeWords {
			if clean == rw {
				isStop = true
				break
			}
		}
		if !isStop && clean != "" {
			kept = append(kept, clean)
		}
	}

	if len(kept) == 0 {
		return q
	}
	return strings.Join(kept, " ")
}

func (s *SearchResponse) FormatContext() string {
	if len(s.Results) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("=== REAL-TIME INTERNET SEARCH DATA (GROUNDING) ===\n")
	sb.WriteString(fmt.Sprintf("Query: %s\nVerified Provider: %s\nRetrieved: %s\n", s.Query, s.Source, time.Now().Format("2006-01-02 15:04:05 MST")))
	if s.UsageText != "" {
		sb.WriteString(fmt.Sprintf("Search Usage: %s\n", s.UsageText))
	}
	sb.WriteString("\n")

	for i, r := range s.Results {
		sb.WriteString(fmt.Sprintf("[%d] %s\n", i+1, r.Title))
		if r.Snippet != "" {
			sb.WriteString(fmt.Sprintf("Snippet: %s\n", r.Snippet))
		}
		if r.URL != "" {
			sb.WriteString(fmt.Sprintf("URL: %s\n", r.URL))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("INSTRUCTION: Use the above real-time verified search results to answer the user's question directly. Always cite the live facts, figures, and sources. Never say you don't have internet access, as the live search data has already been provided to you above.\n")
	sb.WriteString("===================================================\n")

	return sb.String()
}
