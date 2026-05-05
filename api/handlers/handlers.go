package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"summarizer-api/functions"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/gofiber/fiber/v3"
)

// llamaModelsResponse mirrors the relevant fields of the /v1/models payload
// returned by llama-server's router mode.
type llamaModelsResponse struct {
	Data []llamaModel `json:"data"`
}

type llamaModel struct {
	ID     string            `json:"id"`
	Status *llamaModelStatus `json:"status"`
}

type llamaModelStatus struct {
	Value string `json:"value"`
}

// modelSummary is the simplified shape we return to callers.
type modelSummary struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type summaryParams struct {
	engine string
	model  string
	text   string
}

type chatParams struct {
	documentUUID string
	messages     []chatMessagePayload
	userPrompt   string
	engine       string
	model        string
}

type chatMessagePayload struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (cmp chatMessagePayload) ToMessage() types.Message {
	role := types.ConversationRoleUser
	if strings.EqualFold(cmp.Role, "assistant") {
		role = types.ConversationRoleAssistant
	}

	return types.Message{
		Role: role,
		Content: []types.ContentBlock{
			&types.ContentBlockMemberText{Value: cmp.Content},
		},
	}
}

func (cp chatParams) ToMessages() []types.Message {
	messages := make([]types.Message, 0, len(cp.messages))
	for _, msg := range cp.messages {
		messages = append(messages, msg.ToMessage())
	}
	return messages
}

func parseSummaryRequest(c fiber.Ctx) (summaryParams, error) {
	log.Printf("Model: %s\n", c.Query("model"))
	model := strings.TrimSpace(c.Query("model", "gemma4:e2b"))
	if model == "" {
		return summaryParams{}, fmt.Errorf("query parameter 'model' is required")
	}

	// engine is optional; when no engine is provided it assumes "AI" as default value
	log.Printf("Engine: %s\n", c.Query("engine"))
	engine := strings.ToUpper(strings.Trim(c.Query("engine"), "ML"))

	// Get uploaded PDF file from multipart form field "file"
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return summaryParams{}, errors.New("field 'file' is required (multipart/form-data)")
	}

	f, err := fileHeader.Open()
	if err != nil {
		return summaryParams{}, errors.New("failed to open uploaded file")
	}
	defer f.Close()

	pdfBytes, err := io.ReadAll(f)
	if err != nil {
		return summaryParams{}, errors.New("failed to read uploaded file")
	}

	// Extract text from the PDF using ledongthuc/pdf (pure Go, no system deps)
	text, err := functions.ExtractTextFromPDF(pdfBytes)
	if err != nil {
		return summaryParams{}, errors.New("text extraction failed")
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return summaryParams{}, errors.New("no text could be extracted from the PDF")
	}

	return summaryParams{
		engine, model, text,
	}, nil
}

func parseChatRequest(c fiber.Ctx) (chatParams, error) {
	log.Printf("Model: %s\n", c.Query("model"))
	model := strings.TrimSpace(c.Query("model", "anthropic.claude-v2-100k:claude-v2"))
	if model == "" {
		return chatParams{}, fmt.Errorf("query parameter 'model' is required")
	}

	// engine is optional; when no engine is provided it assumes "AI" as default value
	log.Printf("Engine: %s\n", c.Query("engine"))
	engine := strings.ToUpper(strings.Trim(c.Query("engine"), "ML"))

	// Get the document UUID from the URL
	documentUUID := c.Params("document_uuid")
	if documentUUID == "" {
		return chatParams{}, fmt.Errorf("document UUID is required")
	}

	// Get the user message from the request body
	body := strings.TrimSpace(string(c.Body()))
	if body == "" {
		return chatParams{}, fmt.Errorf("request body is required")
	}

	// Parse request payload
	var messages []chatMessagePayload
	if err := json.Unmarshal([]byte(body), &messages); err != nil {
		return chatParams{}, fmt.Errorf("invalid chat history payload: %w", err)
	}
	if len(messages) == 0 {
		return chatParams{}, fmt.Errorf("chat history payload is empty")
	}

	// Normalize and validate messages. Only user/assistant turns are accepted.
	normalizedMessages := make([]chatMessagePayload, 0, len(messages))
	userPrompt := ""
	for _, msg := range messages {
		role := strings.ToLower(strings.TrimSpace(msg.Role))
		content := strings.TrimSpace(msg.Content)
		if content == "" {
			continue
		}
		if role != "user" && role != "assistant" {
			return chatParams{}, fmt.Errorf("invalid role '%s'; expected 'user' or 'assistant'", msg.Role)
		}

		normalizedMessages = append(normalizedMessages, chatMessagePayload{
			Role:    role,
			Content: content,
		})

		if role == "user" {
			userPrompt = content
		}
	}

	if len(normalizedMessages) == 0 {
		return chatParams{}, fmt.Errorf("chat history payload has no valid messages")
	}

	if userPrompt == "" {
		return chatParams{}, fmt.Errorf("chat history must include at least one 'user' message")
	}

	return chatParams{
		documentUUID, normalizedMessages, userPrompt, engine, model,
	}, nil
}
