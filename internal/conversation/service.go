package conversation

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"ifm-chat/internal/database/repositories"
)

type Service struct {
	convRepo *repositories.ConversationRepository
	msgRepo  *repositories.MessageRepository
}

func NewService(convRepo *repositories.ConversationRepository, msgRepo *repositories.MessageRepository) *Service {
	return &Service{
		convRepo: convRepo,
		msgRepo:  msgRepo,
	}
}

func (s *Service) Create(ctx context.Context, title, modelID, presetID, systemPrompt string) (*repositories.Conversation, error) {
	if strings.TrimSpace(title) == "" {
		title = "New Chat"
	}
	var pid *string
	if presetID != "" {
		pid = &presetID
	}

	conv := &repositories.Conversation{
		ID:           uuid.NewString(),
		Title:        title,
		ModelID:      modelID,
		PresetID:     pid,
		SessionID:    uuid.NewString(),
		SystemPrompt: systemPrompt,
	}

	if err := s.convRepo.Create(ctx, conv); err != nil {
		return nil, err
	}
	return conv, nil
}

func (s *Service) Get(ctx context.Context, id string) (*repositories.Conversation, error) {
	return s.convRepo.Get(ctx, id)
}

func (s *Service) List(ctx context.Context, search string) ([]repositories.Conversation, error) {
	return s.convRepo.List(ctx, search)
}

func (s *Service) Update(ctx context.Context, id, title, modelID, presetID, systemPrompt string) error {
	conv, err := s.convRepo.Get(ctx, id)
	if err != nil || conv == nil {
		return fmt.Errorf("conversation not found: %w", err)
	}

	conv.Title = title
	conv.ModelID = modelID
	if presetID != "" {
		conv.PresetID = &presetID
	} else {
		conv.PresetID = nil
	}
	conv.SystemPrompt = systemPrompt

	return s.convRepo.Update(ctx, conv)
}

func (s *Service) UpdateTitle(ctx context.Context, id, title string) error {
	return s.convRepo.UpdateTitle(ctx, id, title)
}

func (s *Service) UpdateFolderPath(ctx context.Context, id string, folderPath *string) error {
	return s.convRepo.UpdateFolderPath(ctx, id, folderPath)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.convRepo.Delete(ctx, id)
}

func (s *Service) ClearMessages(ctx context.Context, convID string) error {
	return s.msgRepo.ClearByConversation(ctx, convID)
}

func (s *Service) GetMessages(ctx context.Context, convID string) ([]repositories.Message, error) {
	return s.msgRepo.ListByConversation(ctx, convID)
}

func (s *Service) AddUserMessage(ctx context.Context, convID, content string) (*repositories.Message, error) {
	msg := &repositories.Message{
		ID:             uuid.NewString(),
		ConversationID: convID,
		Role:           "user",
		Content:        content,
		CreatedAt:      time.Now().UTC(),
	}
	if err := s.msgRepo.Create(ctx, msg); err != nil {
		return nil, err
	}
	_ = s.convRepo.Touch(ctx, convID)
	return msg, nil
}

func (s *Service) AddAssistantMessage(ctx context.Context, convID, content string, reasoning *string, promptTokens, completionTokens int) (*repositories.Message, error) {
	msg := &repositories.Message{
		ID:               uuid.NewString(),
		ConversationID:   convID,
		Role:             "assistant",
		Content:          content,
		ReasoningContent: reasoning,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		CreatedAt:        time.Now().UTC(),
	}
	if err := s.msgRepo.Create(ctx, msg); err != nil {
		return nil, err
	}
	_ = s.convRepo.Touch(ctx, convID)
	return msg, nil
}

func (s *Service) DeleteMessage(ctx context.Context, id string) error {
	return s.msgRepo.Delete(ctx, id)
}

// PrepareRegenerate deletes subsequent messages after messageID
func (s *Service) PrepareRegenerate(ctx context.Context, convID, messageID string) error {
	msg, err := s.msgRepo.Get(ctx, messageID)
	if err != nil || msg == nil {
		return fmt.Errorf("message not found")
	}
	// Delete this assistant message and any subsequent messages
	return s.msgRepo.DeleteAfter(ctx, convID, msg.CreatedAt)
}

// GenerateTitleFromContent creates a concise title from the initial prompt
func GenerateTitleFromContent(content string) string {
	cleaned := strings.TrimSpace(content)
	lines := strings.Split(cleaned, "\n")
	firstLine := strings.TrimSpace(lines[0])

	// Limit to ~40 characters or ~7 words
	words := strings.Fields(firstLine)
	if len(words) == 0 {
		return "New Conversation"
	}

	var titleWords []string
	currLen := 0
	for _, w := range words {
		// Clean punctuation from word borders
		trimmed := strings.TrimFunc(w, func(r rune) bool {
			return unicode.IsPunct(r) && r != '-'
		})
		if trimmed == "" {
			continue
		}
		if currLen+len(trimmed) > 36 {
			break
		}
		titleWords = append(titleWords, trimmed)
		currLen += len(trimmed) + 1
		if len(titleWords) >= 6 {
			break
		}
	}

	if len(titleWords) == 0 {
		return "New Conversation"
	}
	return strings.Join(titleWords, " ")
}
