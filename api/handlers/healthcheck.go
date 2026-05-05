package handlers

import (
	"fmt"
	"net/http"
	"summarizer-api/globals"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/gofiber/fiber/v3"
)

func HealthCheckHandler(c fiber.Ctx) error {
	// Check AWS Bedrock Runtime
	sdkConfig, err := config.LoadDefaultConfig(c.Context(), config.WithRegion(globals.REGION))
	if err != nil {
		fmt.Println("Couldn't load default configuration. Have you set up your AWS account?")
		fmt.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to load AWS configuration: %v", err),
		})
	}
	// list models
	client := bedrock.NewFromConfig(sdkConfig)
	_, err = client.ListFoundationModels(c.Context(), &bedrock.ListFoundationModelsInput{})
	if err != nil {
		fmt.Println("Couldn't list foundation models. Have you set up your AWS account?")
		fmt.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to validate AWS credentials: %v", err),
		})
	}

	// Check Ollama server
	httpClient := &http.Client{Timeout: time.Second * 5}
	resp, err := httpClient.Get(globals.OLLAMA_URL + "/models")
	if err != nil {
		fmt.Println("Couldn't reach Ollama server:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to reach Ollama: %v", err),
		})
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Ollama returned unexpected status: %d", resp.StatusCode),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "ok",
	})
}
