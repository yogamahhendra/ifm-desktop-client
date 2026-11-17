package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"ifm-chat/internal/config"
	"ifm-chat/internal/conversation"
	"ifm-chat/internal/database"
	"ifm-chat/internal/database/repositories"
	"ifm-chat/internal/ifm"
	"ifm-chat/internal/preset"
	"ifm-chat/internal/security"
	"ifm-chat/internal/settings"
	"ifm-chat/internal/workspace"
	"ifm-chat/internal/search"
	"os"
	"path/filepath"
)

// App struct manages application state and Wails IPC methods
type App struct {
	ctx          context.Context
	db           *database.DB
	paths        *config.AppPaths
	security     *security.CredentialService
	convService  *conversation.Service
	modelRepo    *repositories.ModelRepository
	presetSvc    *preset.Service
	settingsSvc  *settings.Service
	keyRepo      *repositories.APIKeyRepository
	ifmClient    *ifm.Client
	workspaceSvc *workspace.Service
	searchSvc    *search.MultiSearchService

	// Generation cancellations per conversation ID
	genMu      sync.Mutex
	genCancels map[string]context.CancelFunc
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		genCancels: make(map[string]context.CancelFunc),
	}
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	paths, err := config.GetAppPaths()
	if err != nil {
		wailsRuntime.LogErrorf(ctx, "Failed to get app paths: %v", err)
		return
	}
	a.paths = paths

	db, err := database.Open(paths.DatabaseURL)
	if err != nil {
		wailsRuntime.LogErrorf(ctx, "Failed to initialize database: %v", err)
		return
	}
	a.db = db

	if err := db.Migrate(); err != nil {
		wailsRuntime.LogErrorf(ctx, "Database migration error: %v", err)
		return
	}

	convRepo := repositories.NewConversationRepository(db.DB)
	msgRepo := repositories.NewMessageRepository(db.DB)
	modelRepo := repositories.NewModelRepository(db.DB)
	presetRepo := repositories.NewPresetRepository(db.DB)
	settingsRepo := repositories.NewSettingsRepository(db.DB)
	sec := security.NewCredentialService()
	keyRepo := repositories.NewAPIKeyRepository(db.DB)

	a.convService = conversation.NewService(convRepo, msgRepo)
	a.modelRepo = modelRepo
	a.presetSvc = preset.NewService(presetRepo)
	a.settingsSvc = settings.NewService(settingsRepo, sec, paths)
	a.keyRepo = keyRepo
	a.security = sec
	a.workspaceSvc = workspace.NewService()
	a.searchSvc = search.NewMultiSearchService()

	// Load stored Tavily key if available
	tavilyKey, _ := sec.GetTavilyKey()
	if tavilyKey != "" {
		a.searchSvc.SetTavilyKey(tavilyKey)
	}

	// Auto-migrate legacy single key to multi-key table if present
	legacyKey, _ := sec.GetAPIKey()
	keysList, _ := keyRepo.List(context.Background())
	if len(keysList) == 0 && legacyKey != "" {
		mask := "ifm-..."
		if len(legacyKey) > 4 {
			mask = "ifm-..." + legacyKey[len(legacyKey)-4:]
		}
		item := &repositories.APIKeyItem{
			ID:       uuid.NewString(),
			Name:     "Default Key",
			KeyMask:  mask,
			IsActive: true,
		}
		_ = keyRepo.Create(context.Background(), item)
		_ = sec.SetKeyWithID(item.ID, legacyKey)
	}

	// Determine active API key
	activeKey := a.getActiveKeySecret()

	cfg, _ := a.settingsSvc.GetConfig(context.Background())
	baseURL := "https://api.ifm.ai/v1"
	timeoutSec := 60
	if cfg != nil {
		if cfg.IFMBaseURL != "" {
			baseURL = cfg.IFMBaseURL
		}
		if t, err := strconv.Atoi(cfg.RequestTimeout); err == nil && t > 0 {
			timeoutSec = t
		}
	}

	a.ifmClient = ifm.NewClient(baseURL, activeKey, time.Duration(timeoutSec)*time.Second)
	wailsRuntime.LogInfo(ctx, "IFM Chat initialized successfully")
}

func (a *App) getActiveKeySecret() string {
	if a.keyRepo != nil {
		active, _ := a.keyRepo.GetActive(a.ctx)
		if active != nil {
			key, err := a.security.GetKeyWithID(active.ID)
			if err == nil && key != "" {
				return key
			}
		}
	}
	key, _ := a.security.GetAPIKey()
	return key
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	if a.db != nil {
		_ = a.db.Close()
	}
}

// ==================== SECURITY & API KEY ====================

func (a *App) ListAPIKeys() ([]repositories.APIKeyItem, error) {
	return a.keyRepo.List(a.ctx)
}

func (a *App) AddAPIKey(name, apiKey string) (*repositories.APIKeyItem, error) {
	name = strings.TrimSpace(name)
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, errors.New("API key cannot be empty")
	}
	if name == "" {
		name = "API Key"
	}
	id := uuid.NewString()
	mask := "ifm-..."
	if len(apiKey) > 4 {
		mask = "ifm-..." + apiKey[len(apiKey)-4:]
	}

	existing, _ := a.keyRepo.List(a.ctx)
	isActive := len(existing) == 0

	item := &repositories.APIKeyItem{
		ID:       id,
		Name:     name,
		KeyMask:  mask,
		IsActive: isActive,
	}

	if err := a.keyRepo.Create(a.ctx, item); err != nil {
		return nil, err
	}
	if err := a.security.SetKeyWithID(id, apiKey); err != nil {
		return nil, err
	}

	if isActive {
		a.ifmClient.SetAPIKey(apiKey)
		_ = a.security.SetAPIKey(apiKey)
	}

	return item, nil
}

func (a *App) SetActiveAPIKey(id string) error {
	key, err := a.security.GetKeyWithID(id)
	if err != nil || key == "" {
		return errors.New("could not find credential for selected key")
	}
	if err := a.keyRepo.SetActive(a.ctx, id); err != nil {
		return err
	}
	a.ifmClient.SetAPIKey(key)
	_ = a.security.SetAPIKey(key)
	return nil
}

func (a *App) DeleteAPIKeyByID(id string) error {
	_ = a.security.DeleteKeyWithID(id)
	if err := a.keyRepo.Delete(a.ctx, id); err != nil {
		return err
	}
	active, _ := a.keyRepo.GetActive(a.ctx)
	if active == nil {
		list, _ := a.keyRepo.List(a.ctx)
		if len(list) > 0 {
			_ = a.SetActiveAPIKey(list[0].ID)
		} else {
			a.ifmClient.SetAPIKey("")
			_ = a.security.DeleteAPIKey()
		}
	}
	return nil
}

// ==================== TAVILY WEB SEARCH KEY ====================

func (a *App) SaveTavilyKey(key string) error {
	trimmed := strings.TrimSpace(key)
	if err := a.security.SetTavilyKey(trimmed); err != nil {
		return err
	}
	if a.searchSvc != nil {
		a.searchSvc.SetTavilyKey(trimmed)
	}
	return nil
}

func (a *App) GetTavilyKeyMask() string {
	key, err := a.security.GetTavilyKey()
	if err != nil || key == "" {
		return ""
	}
	if len(key) <= 8 {
		return "tvly-****"
	}
	return key[:5] + "****" + key[len(key)-4:]
}

func (a *App) DeleteTavilyKey() error {
	_ = a.security.DeleteTavilyKey()
	if a.searchSvc != nil {
		a.searchSvc.SetTavilyKey("")
	}
	return nil
}

func (a *App) SaveAPIKey(apiKey string) error {
	_, err := a.AddAPIKey("Default Key", apiKey)
	return err
}

func (a *App) HasAPIKey() bool {
	keys, err := a.keyRepo.List(a.ctx)
	if err == nil && len(keys) > 0 {
		return true
	}
	return a.security.HasAPIKey()
}

func (a *App) DeleteAPIKey() error {
	keys, _ := a.keyRepo.List(a.ctx)
	for _, k := range keys {
		_ = a.DeleteAPIKeyByID(k.ID)
	}
	_ = a.security.DeleteAPIKey()
	a.ifmClient.SetAPIKey("")
	return nil
}

func (a *App) TestConnection() error {
	apiKey := a.getActiveKeySecret()
	if apiKey == "" {
		return errors.New("API key not configured in Windows Credential Manager")
	}
	a.ifmClient.SetAPIKey(apiKey)

	// Use a 60 second timeout for 375B reasoning model connection check
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	return a.ifmClient.TestConnection(ctx)
}

// ==================== CONVERSATIONS ====================

func (a *App) CreateConversation(title, modelID, presetID, systemPrompt string) (*repositories.Conversation, error) {
	return a.convService.Create(a.ctx, title, modelID, presetID, systemPrompt)
}

func (a *App) GetConversation(id string) (*repositories.Conversation, error) {
	return a.convService.Get(a.ctx, id)
}

func (a *App) ListConversations(search string) ([]repositories.Conversation, error) {
	return a.convService.List(a.ctx, search)
}

func (a *App) UpdateConversation(id, title, modelID, presetID, systemPrompt string) error {
	return a.convService.Update(a.ctx, id, title, modelID, presetID, systemPrompt)
}

func (a *App) UpdateConversationTitle(id, title string) error {
	return a.convService.UpdateTitle(a.ctx, id, title)
}

func (a *App) UpdateConversationModel(id, modelID string) error {
	conv, err := a.convService.Get(a.ctx, id)
	if err != nil || conv == nil {
		return fmt.Errorf("conversation not found")
	}
	presetID := ""
	if conv.PresetID != nil {
		presetID = *conv.PresetID
	}
	return a.convService.Update(a.ctx, id, conv.Title, modelID, presetID, conv.SystemPrompt)
}

func (a *App) UpdateConversationPreset(id, presetID string) error {
	conv, err := a.convService.Get(a.ctx, id)
	if err != nil || conv == nil {
		return fmt.Errorf("conversation not found")
	}
	sysPrompt := conv.SystemPrompt
	if presetID != "" {
		if p, pErr := a.presetSvc.Get(a.ctx, presetID); pErr == nil && p != nil {
			sysPrompt = p.SystemPrompt
		}
	} else {
		sysPrompt = ""
	}
	return a.convService.Update(a.ctx, id, conv.Title, conv.ModelID, presetID, sysPrompt)
}

func (a *App) ExportConversationMarkdown(conversationID string) (string, error) {
	conv, err := a.convService.Get(a.ctx, conversationID)
	if err != nil || conv == nil {
		return "", fmt.Errorf("conversation not found")
	}
	messages, err := a.convService.GetMessages(a.ctx, conversationID)
	if err != nil {
		return "", err
	}

	var md strings.Builder
	md.WriteString(fmt.Sprintf("# %s\n\n", conv.Title))
	md.WriteString(fmt.Sprintf("- **Model**: `%s`\n", conv.ModelID))
	md.WriteString(fmt.Sprintf("- **Date**: %s\n\n---\n\n", conv.CreatedAt.Format("2006-01-02 15:04:05 UTC")))

	for _, m := range messages {
		if m.Role == "user" {
			md.WriteString("### 👤 User\n\n")
			md.WriteString(m.Content)
			md.WriteString("\n\n")
		} else if m.Role == "assistant" {
			md.WriteString("### 🤖 IFM Assistant\n\n")
			if m.ReasoningContent != nil && *m.ReasoningContent != "" {
				md.WriteString("<details><summary><b>Thought Process (Reasoning)</b></summary>\n\n")
				md.WriteString(*m.ReasoningContent)
				md.WriteString("\n\n</details>\n\n")
			}
			md.WriteString(m.Content)
			md.WriteString("\n\n---\n\n")
		}
	}

	return md.String(), nil
}

func (a *App) DeleteConversation(id string) error {
	return a.convService.Delete(a.ctx, id)
}

func (a *App) ClearConversationMessages(id string) error {
	return a.convService.ClearMessages(a.ctx, id)
}

// ==================== WORKSPACE FOLDER CONTEXT ====================

func (a *App) SelectFolderDialog() (string, error) {
	dir, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select Workspace Folder for Context",
	})
	if err != nil {
		return "", err
	}
	return dir, nil
}

func (a *App) AttachFolder(conversationID, folderPath string) (*workspace.WorkspaceSummary, error) {
	folderPath = strings.TrimSpace(folderPath)
	if folderPath == "" {
		return nil, errors.New("folder path cannot be empty")
	}
	summary, err := a.workspaceSvc.Scan(folderPath)
	if err != nil {
		return nil, err
	}
	if err := a.convService.UpdateFolderPath(a.ctx, conversationID, &folderPath); err != nil {
		return nil, err
	}
	return summary, nil
}

func (a *App) DetachFolder(conversationID string) error {
	return a.convService.UpdateFolderPath(a.ctx, conversationID, nil)
}

func (a *App) GetWorkspaceSummary(conversationID string) (*workspace.WorkspaceSummary, error) {
	conv, err := a.convService.Get(a.ctx, conversationID)
	if err != nil || conv == nil || conv.FolderPath == nil || *conv.FolderPath == "" {
		return nil, nil
	}
	return a.workspaceSvc.Scan(*conv.FolderPath)
}

func (a *App) OpenFolderInExplorer(folderPath string) error {
	return a.workspaceSvc.OpenInExplorer(folderPath)
}

// ==================== MESSAGES & CHAT STREAMING ====================

func (a *App) GetMessages(conversationID string) ([]repositories.Message, error) {
	return a.convService.GetMessages(a.ctx, conversationID)
}

func (a *App) DeleteMessage(messageID string) error {
	return a.convService.DeleteMessage(a.ctx, messageID)
}

func (a *App) StopGeneration(conversationID string) {
	a.genMu.Lock()
	defer a.genMu.Unlock()
	if conversationID == "" {
		for id, cancel := range a.genCancels {
			cancel()
			delete(a.genCancels, id)
		}
		return
	}
	if cancel, exists := a.genCancels[conversationID]; exists {
		cancel()
		delete(a.genCancels, conversationID)
	}
}

// SendMessage handles adding user message and streaming response asynchronously
func (a *App) SendMessage(conversationID, content string, webSearchMode string) error {
	// Add user message to DB
	_, err := a.convService.AddUserMessage(a.ctx, conversationID, content)
	if err != nil {
		return fmt.Errorf("failed to save user message: %w", err)
	}

	// Fetch conversation to see if title should be auto-updated
	conv, err := a.convService.Get(a.ctx, conversationID)
	if err == nil && conv != nil && (conv.Title == "New Chat" || conv.Title == "") {
		newTitle := conversation.GenerateTitleFromContent(content)
		_ = a.convService.UpdateTitle(a.ctx, conversationID, newTitle)
		wailsRuntime.EventsEmit(a.ctx, "conversation_title_updated", map[string]string{
			"id":    conversationID,
			"title": newTitle,
		})
	}

	// Trigger streaming generation in background with chosen search mode
	go a.generateResponse(conversationID, webSearchMode)

	return nil
}

// RegenerateMessage deletes recent assistant reply and re-runs generation
func (a *App) RegenerateMessage(conversationID, messageID string, webSearchMode string) error {
	if err := a.convService.PrepareRegenerate(a.ctx, conversationID, messageID); err != nil {
		return err
	}
	go a.generateResponse(conversationID, webSearchMode)
	return nil
}

func (a *App) generateResponse(conversationID string, webSearchMode string) {
	// Cancel any existing generation only for THIS specific conversation
	a.StopGeneration(conversationID)

	ctx, cancel := context.WithCancel(context.Background())
	a.genMu.Lock()
	if a.genCancels == nil {
		a.genCancels = make(map[string]context.CancelFunc)
	}
	a.genCancels[conversationID] = cancel
	a.genMu.Unlock()

	defer func() {
		a.genMu.Lock()
		delete(a.genCancels, conversationID)
		a.genMu.Unlock()
	}()

	conv, err := a.convService.Get(ctx, conversationID)
	if err != nil || conv == nil {
		a.emitStreamError(conversationID, "Conversation not found")
		return
	}

	dbMessages, err := a.convService.GetMessages(ctx, conversationID)
	if err != nil {
		a.emitStreamError(conversationID, "Failed to load conversation history")
		return
	}

	// Prepare chat messages for IFM API
	var apiMessages []ifm.ChatMessage

	// Build full system prompt and inject workspace folder context if attached
	fullSystemPrompt := conv.SystemPrompt
	if fullSystemPrompt == "" && conv.PresetID != nil && *conv.PresetID != "" {
		if p, pErr := a.presetSvc.Get(ctx, *conv.PresetID); pErr == nil && p != nil {
			fullSystemPrompt = p.SystemPrompt
		}
	}
	if conv.FolderPath != nil && *conv.FolderPath != "" {
		workspacePrompt, wsErr := a.workspaceSvc.BuildPrompt(*conv.FolderPath, nil, 60000)
		if wsErr == nil && workspacePrompt != "" {
			if fullSystemPrompt != "" {
				fullSystemPrompt += "\n\n" + workspacePrompt
			} else {
				fullSystemPrompt = workspacePrompt
			}
		}
	}

	// Live Web Search Grounding: determined by webSearchMode ('auto', 'on', 'off')
	var searchSource string
	var searchCredits float64
	if a.searchSvc != nil && len(dbMessages) > 0 && webSearchMode != "off" {
		lastMsg := dbMessages[len(dbMessages)-1]
		if lastMsg.Role == "user" {
			// If mode is 'auto', check keywords. If mode is 'on', always search.
			shouldTrigger := webSearchMode == "on" || a.searchSvc.ShouldSearch(lastMsg.Content)
			if shouldTrigger {
				searchResp, searchErr := a.searchSvc.Search(ctx, lastMsg.Content)
				if searchErr == nil && searchResp != nil && len(searchResp.Results) > 0 {
					searchSource = searchResp.Source
					searchCredits = searchResp.Credits
					groundingContext := searchResp.FormatContext()
					if groundingContext != "" {
						if fullSystemPrompt != "" {
							fullSystemPrompt += "\n\n" + groundingContext
						} else {
							fullSystemPrompt = groundingContext
						}
						if a.ctx != nil {
							wailsRuntime.LogInfof(a.ctx, "Injected web search grounding from: %s for query: %s (mode: %s)", searchResp.Source, lastMsg.Content, webSearchMode)
						}
					}
				}
			}
		}
	}

	if fullSystemPrompt != "" {
		apiMessages = append(apiMessages, ifm.ChatMessage{
			Role:    "system",
			Content: fullSystemPrompt,
		})
	}

	// Replay previous messages exactly, ensuring EVERY assistant message includes thinking fields for multi-turn validation!
	for _, m := range dbMessages {
		msg := ifm.ChatMessage{
			Role:       m.Role,
			Content:    m.Content,
			ToolCallID: m.ToolCallID,
		}
		if m.Role == "assistant" {
			// IFM requires a supported thinking field on every assistant message in multi-turn history
			reasoning := ""
			if m.ReasoningContent != nil {
				reasoning = *m.ReasoningContent
			}
			msg.ReasoningContent = &reasoning
			msg.Think = &reasoning
			msg.Thinking = &reasoning
		}
		apiMessages = append(apiMessages, msg)
	}

	// Load model parameters
	model, err := a.modelRepo.GetByIdentifier(ctx, conv.ModelID)
	var temp *float64
	var topP *float64
	var maxTokens *int
	var effort string

	if model != nil {
		t := model.Temperature
		p := model.TopP
		mt := model.MaxTokens
		temp = &t
		topP = &p
		maxTokens = &mt
		effort = model.ReasoningEffort
	}

	// Override with preset parameters if custom behavior is defined
	if conv.PresetID != nil && *conv.PresetID != "" {
		if p, pErr := a.presetSvc.Get(ctx, *conv.PresetID); pErr == nil && p != nil {
			if p.Temperature != nil {
				temp = p.Temperature
			}
			if p.TopP != nil {
				topP = p.TopP
			}
			if p.MaxTokens != nil {
				maxTokens = p.MaxTokens
			}
			if p.ReasoningEffort != nil && *p.ReasoningEffort != "" {
				effort = *p.ReasoningEffort
			}
		}
	}

	// Ensure active API key is set
	apiKey := a.getActiveKeySecret()
	a.ifmClient.SetAPIKey(apiKey)

	req := ifm.ChatCompletionRequest{
		Model:           conv.ModelID,
		Messages:        apiMessages,
		Temperature:     temp,
		TopP:            topP,
		MaxTokens:       maxTokens,
		ReasoningEffort: effort,
		Stream:          true,
	}

	assistantMsgID := uuid.NewString()
	inThinkTag := false
	var fullContent string
	var fullReasoning string
	var finishReason *string
	var promptTokens, completionTokens int

	err = a.ifmClient.StreamChatCompletion(ctx, req, conv.SessionID, func(chunk *ifm.StreamChunk) error {
		if chunk.Usage != nil {
			promptTokens = chunk.Usage.PromptTokens
			completionTokens = chunk.Usage.CompletionTokens
		}

		for _, choice := range chunk.Choices {
			contentDelta := choice.Delta.Content
			reasoningDelta := ""
			if choice.Delta.ReasoningContent != nil && *choice.Delta.ReasoningContent != "" {
				reasoningDelta = *choice.Delta.ReasoningContent
			} else if choice.Delta.Think != nil && *choice.Delta.Think != "" {
				reasoningDelta = *choice.Delta.Think
			} else if choice.Delta.Thinking != nil && *choice.Delta.Thinking != "" {
				reasoningDelta = *choice.Delta.Thinking
			}

			// Also handle models that stream thinking inside <think> tags in Content
			if reasoningDelta == "" && contentDelta != "" {
				if strings.Contains(contentDelta, "<think>") {
					inThinkTag = true
					parts := strings.SplitN(contentDelta, "<think>", 2)
					contentDelta = parts[0]
					after := parts[1]
					if strings.Contains(after, "</think>") {
						thinkParts := strings.SplitN(after, "</think>", 2)
						reasoningDelta = thinkParts[0]
						contentDelta += thinkParts[1]
						inThinkTag = false
					} else {
						reasoningDelta = after
					}
				} else if inThinkTag {
					if strings.Contains(contentDelta, "</think>") {
						parts := strings.SplitN(contentDelta, "</think>", 2)
						reasoningDelta = parts[0]
						contentDelta = parts[1]
						inThinkTag = false
					} else {
						reasoningDelta = contentDelta
						contentDelta = ""
					}
				}
			}

			fullContent += contentDelta
			fullReasoning += reasoningDelta
			if choice.FinishReason != nil {
				finishReason = choice.FinishReason
			}

			event := ifm.StreamEvent{
				ConversationID: conversationID,
				MessageID:      assistantMsgID,
				ContentDelta:   contentDelta,
				ReasoningDelta: reasoningDelta,
				FullContent:    fullContent,
				FullReasoning:  fullReasoning,
				FinishReason:   finishReason,
				IsDone:         false,
			}
			wailsRuntime.EventsEmit(a.ctx, "chat_stream", event)
		}
		return nil
	})

	if err != nil {
		if errors.Is(err, context.Canceled) {
			if fullContent != "" || fullReasoning != "" {
				var rc *string
				if fullReasoning != "" {
					rc = &fullReasoning
				}
				stopped := "stopped"
				_, _ = a.convService.AddAssistantMessage(context.Background(), conversationID, fullContent, rc, promptTokens, completionTokens)
				wailsRuntime.EventsEmit(a.ctx, "chat_stream", ifm.StreamEvent{
					ConversationID: conversationID,
					MessageID:      assistantMsgID,
					FullContent:    fullContent,
					FullReasoning:  fullReasoning,
					FinishReason:   &stopped,
					IsDone:         true,
				})
			}
			return
		}
		a.emitStreamError(conversationID, err.Error())
		return
	}

	// Save completed message to DB
	var rc *string
	if fullReasoning != "" {
		rc = &fullReasoning
	}
	_, err = a.convService.AddAssistantMessage(context.Background(), conversationID, fullContent, rc, promptTokens, completionTokens)
	if err != nil {
		wailsRuntime.LogErrorf(a.ctx, "Failed saving assistant message: %v", err)
	}

	// Record token usage on active API key
	if activeItem, _ := a.keyRepo.GetActive(context.Background()); activeItem != nil {
		_ = a.keyRepo.RecordUsage(context.Background(), activeItem.ID, promptTokens, completionTokens)
	}

	// Emit final completion event
	wailsRuntime.EventsEmit(a.ctx, "chat_stream", ifm.StreamEvent{
		ConversationID:   conversationID,
		MessageID:        assistantMsgID,
		FullContent:      fullContent,
		FullReasoning:    fullReasoning,
		FinishReason:     finishReason,
		IsDone:           true,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		SearchSource:     searchSource,
		SearchCredits:    searchCredits,
	})
}

func (a *App) emitStreamError(conversationID, errorMsg string) {
	wailsRuntime.EventsEmit(a.ctx, "chat_stream", ifm.StreamEvent{
		ConversationID: conversationID,
		IsDone:         true,
		Error:          errorMsg,
	})
}

// ==================== MODELS ====================

func (a *App) ListModels() ([]repositories.ModelItem, error) {
	return a.modelRepo.List(a.ctx)
}

func (a *App) CreateModel(model repositories.ModelItem) (*repositories.ModelItem, error) {
	if err := a.modelRepo.Create(a.ctx, &model); err != nil {
		return nil, err
	}
	return &model, nil
}

func (a *App) UpdateModel(model repositories.ModelItem) error {
	return a.modelRepo.Update(a.ctx, &model)
}

func (a *App) DeleteModel(id string) error {
	return a.modelRepo.Delete(a.ctx, id)
}

func (a *App) SetDefaultModel(id string) error {
	return a.modelRepo.SetDefault(a.ctx, id)
}

// ==================== PRESETS ====================

func (a *App) ListPresets() ([]repositories.PresetItem, error) {
	return a.presetSvc.List(a.ctx)
}

func (a *App) CreatePreset(preset repositories.PresetItem) (*repositories.PresetItem, error) {
	return a.presetSvc.Create(a.ctx, preset)
}

func (a *App) UpdatePreset(preset repositories.PresetItem) error {
	return a.presetSvc.Update(a.ctx, preset)
}

func (a *App) DeletePreset(id string) error {
	return a.presetSvc.Delete(a.ctx, id)
}

// ==================== SETTINGS & STORAGE ====================

func (a *App) GetAppConfig() (*settings.AppConfigResponse, error) {
	return a.settingsSvc.GetConfig(a.ctx)
}

func (a *App) SaveSettings(cfg map[string]string) error {
	if err := a.settingsSvc.SaveSettings(a.ctx, cfg); err != nil {
		return err
	}
	if baseURL, ok := cfg["ifm_base_url"]; ok && baseURL != "" {
		a.ifmClient.SetBaseURL(baseURL)
	}
	if timeoutStr, ok := cfg["request_timeout"]; ok {
		if t, err := strconv.Atoi(timeoutStr); err == nil && t > 0 {
			a.ifmClient.SetTimeout(time.Duration(t) * time.Second)
		}
	}
	return nil
}

type ExportData struct {
	Version       string                      `json:"version"`
	ExportedAt    time.Time                   `json:"exported_at"`
	Conversations []repositories.Conversation `json:"conversations"`
	Messages      []repositories.Message      `json:"messages"`
}

func (a *App) ExportConversations() (string, error) {
	convs, err := a.convService.List(a.ctx, "")
	if err != nil {
		return "", err
	}

	var allMessages []repositories.Message
	for _, c := range convs {
		msgs, err := a.convService.GetMessages(a.ctx, c.ID)
		if err == nil {
			allMessages = append(allMessages, msgs...)
		}
	}

	export := ExportData{
		Version:       "1.0",
		ExportedAt:    time.Now().UTC(),
		Conversations: convs,
		Messages:      allMessages,
	}

	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (a *App) ImportConversations(jsonData string) error {
	var export ExportData
	if err := json.Unmarshal([]byte(jsonData), &export); err != nil {
		return fmt.Errorf("invalid import format: %w", err)
	}

	convRepo := repositories.NewConversationRepository(a.db.DB)
	msgRepo := repositories.NewMessageRepository(a.db.DB)

	for _, c := range export.Conversations {
		_ = convRepo.Create(a.ctx, &c)
	}
	for _, m := range export.Messages {
		_ = msgRepo.Create(a.ctx, &m)
	}

	return nil
}

func (a *App) ClearApplicationData() error {
	convRepo := repositories.NewConversationRepository(a.db.DB)
	return convRepo.ClearAll(a.ctx)
}

// MoveDatabaseLocation moves the current SQLite database file to a new user-selected folder
func (a *App) MoveDatabaseLocation(targetDirectory string) (string, error) {
	targetDirectory = strings.TrimSpace(targetDirectory)
	if targetDirectory == "" {
		return "", errors.New("target directory cannot be empty")
	}

	if err := os.MkdirAll(targetDirectory, 0755); err != nil {
		return "", fmt.Errorf("failed creating target directory: %w", err)
	}

	currentDbPath := a.paths.DatabaseURL
	newDbPath := filepath.Join(targetDirectory, "ifm-chat.db")

	if filepath.Clean(currentDbPath) == filepath.Clean(newDbPath) {
		return currentDbPath, errors.New("database is already located in this directory")
	}

	// 1. Flush & checkpoint WAL to ensure clean copy
	if a.db != nil {
		_, _ = a.db.Exec("PRAGMA wal_checkpoint(TRUNCATE);")
		_ = a.db.Close()
	}

	// 2. Helper to move file
	moveFile := func(src, dst string) error {
		if _, err := os.Stat(src); os.IsNotExist(err) {
			return nil
		}
		data, err := os.ReadFile(src)
		if err != nil {
			return err
		}
		if err := os.WriteFile(dst, data, 0644); err != nil {
			return err
		}
		_ = os.Remove(src)
		return nil
	}

	// Move main DB file and auxiliary files (-wal, -shm)
	if err := moveFile(currentDbPath, newDbPath); err != nil {
		// Reopen old DB on failure
		oldDb, _ := database.Open(currentDbPath)
		a.db = oldDb
		return "", fmt.Errorf("failed moving database file: %w", err)
	}
	_ = moveFile(currentDbPath+"-wal", newDbPath+"-wal")
	_ = moveFile(currentDbPath+"-shm", newDbPath+"-shm")

	// 3. Open new database connection
	newDB, err := database.Open(newDbPath)
	if err != nil {
		return "", fmt.Errorf("failed opening moved database: %w", err)
	}
	a.db = newDB

	// 4. Update repository pointers to use new DB
	convRepo := repositories.NewConversationRepository(newDB.DB)
	msgRepo := repositories.NewMessageRepository(newDB.DB)
	modelRepo := repositories.NewModelRepository(newDB.DB)
	presetRepo := repositories.NewPresetRepository(newDB.DB)
	settingsRepo := repositories.NewSettingsRepository(newDB.DB)
	keyRepo := repositories.NewAPIKeyRepository(newDB.DB)

	a.convService = conversation.NewService(convRepo, msgRepo)
	a.modelRepo = modelRepo
	a.presetSvc = preset.NewService(presetRepo)
	a.settingsSvc = settings.NewService(settingsRepo, a.security, a.paths)
	a.keyRepo = keyRepo

	// 5. Persist new location in config.json
	a.paths.DatabaseURL = newDbPath
	a.paths.DatabaseDir = targetDirectory
	_ = config.SaveCustomDatabasePath(newDbPath)

	return newDbPath, nil
}

// SelectFolderForDatabase prompts directory picker dialog
func (a *App) SelectFolderForDatabase() (string, error) {
	return wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select New Folder for SQLite Database",
	})
}

func (a *App) GetAppVersion() string {
	return "1.0.0"
}
