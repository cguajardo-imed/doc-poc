package handlers

import (
	"fmt"
	"log"
	"summarizer-api/functions"
	"summarizer-api/globals"
	"summarizer-api/prompts"

	"github.com/gofiber/fiber/v3"
)

func DocumentChatHandler(c fiber.Ctx) error {
	ep, err := parseChatRequest(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("ML request failed: %v", err),
		})
	}
	log.Printf("Document Chat Handler: %+v\n", ep)

	document, err := globals.DB.GetDocumentByID(ep.documentUUID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "document not found",
		})
	}

	// Get the LLM message from the request body
	llmMessage, err := functions.CallLLM(c.Context(), functions.LLMConfig{
		Type:         globals.LLM_CALL_TYPE__CHAT,
		Engine:       ep.engine,
		UserPrompt:   ep.userPrompt,
		SystemPrompt: fmt.Sprintf(prompts.SystemPrompt_Chat, document.Content),
		Model:        ep.model,
	}, ep.ToMessages())
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("LLM request failed: %v", err),
		})
	}

	// Create a new chat in the database
	_, err = globals.DB.CreateChat(ep.documentUUID, ep.userPrompt, llmMessage)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create chat: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"message": llmMessage,
	})
}
