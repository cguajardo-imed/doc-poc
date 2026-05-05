package handlers

import (
	"fmt"
	"log"

	"summarizer-api/functions"
	"summarizer-api/globals"
	"summarizer-api/prompts"

	"github.com/gofiber/fiber/v3"
)

func SummarizeHandler(c fiber.Ctx) error {
	ep, err := parseSummaryRequest(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("ML request failed: %v", err),
		})
	}
	documentHash := functions.GenerateHash(ep.model + "--" + ep.text)
	log.Printf("Document hash: %s", documentHash)
	// Check if the document already exists in the database
	document, err := globals.DB.GetDocumentByHash(documentHash)
	if err == nil {
		log.Printf("Document already exists in the database with hash: %s", documentHash)
		return c.JSON(fiber.Map{
			"summary":     document.Summary,
			"full":        ep.text,
			"document_id": document.UUID,
		})
	}

	var summary = ""
	if ep.engine == "AWS" || ep.engine == "OLLAMA" {
		userPrompt := fmt.Sprintf(prompts.UserPrompt_Summarizer, ep.text)
		config := functions.LLMConfig{
			Type:         globals.LLM_CALL_TYPE__INVOKE,
			Engine:       ep.engine,
			UserPrompt:   userPrompt,
			SystemPrompt: prompts.SystemPrompt_Summarizer,
			Model:        ep.model,
			Temperature:  &globals.SUMMARY_TEMPERATURE,
		}
		summary, err = functions.CallLLM(c.Context(), config, nil)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("LLM request failed: %v", err),
			})
		}
	} else {
		summary, err = functions.CallML(c.Context(), ep.text)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("ML request failed: %v", err),
			})
		}
	}

	// Create a new document in the database
	documentUUID, err := globals.DB.CreateDocument(documentHash, ep.text, summary)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create document: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"summary":     summary,
		"full":        ep.text,
		"document_id": documentUUID,
	})
}
