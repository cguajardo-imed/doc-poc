package main

import (
	"context"
	"fmt"
	"os"

	"summarizer-api/db"
	"summarizer-api/globals"
	"summarizer-api/handlers"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/static"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file")
	}

	ctx := context.Background()
	sdkConfig, err := config.LoadDefaultConfig(ctx, config.WithRegion(globals.REGION))
	if err != nil {
		fmt.Println("Couldn't load default configuration. Have you set up your AWS account?")
		fmt.Println(err)
		return
	}
	globals.BedrockClient = bedrockruntime.NewFromConfig(sdkConfig)
	globals.OLLAMA_URL = os.Getenv("OLLAMA_URL")

	globals.DB = db.NewDB()
	if globals.DB == nil {
		fmt.Println("Failed to create database connection.")
		return
	}
	if err := globals.DB.Init(); err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		return
	}

	app := fiber.New(fiber.Config{
		// Allow up to 50 MB PDF uploads
		BodyLimit: 50 * 1024 * 1024,
	})

	// Logging Request ID
	app.Use(logger.New())
	app.Get("/*", static.New("./files"))
	app.Post("/api/summarize", handlers.SummarizeHandler)
	app.Post("/api/document-chat/:document_uuid", handlers.DocumentChatHandler)
	app.Get("/api/health", handlers.HealthCheckHandler)
	app.Delete("/api/db", handlers.ClearDBHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	if err := app.Listen(":" + port); err != nil {
		panic(err)
	}
}
