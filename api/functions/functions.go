package functions

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

type LLMConfig struct {
	Type         string
	Engine       string
	UserPrompt   string
	SystemPrompt string
	Model        string
	Temperature  *float32
}

type ClaudeResponse struct {
	Completion string `json:"completion"`
}

func CallML(ctx context.Context, text string) (string, error) {

	return "NOT IMPLEMENTED YET", nil
}

// parseBedrockResponse parses the Bedrock response structure and extracts relevant fields.
func parseBedrockResponse(response *bedrockruntime.ConverseOutput) (map[string]any, error) {
	if response == nil {
		return nil, fmt.Errorf("response is nil")
	}

	parsedResponse := map[string]any{
		"Metrics":        response.Metrics,
		"Output":         response.Output,
		"StopReason":     response.StopReason,
		"Usage":          response.Usage,
		"ServiceTier":    response.ServiceTier,
		"Trace":          response.Trace,
		"ResultMetadata": response.ResultMetadata,
	}

	// Additional parsing for nested fields in ResultMetadata if needed
	if response.ResultMetadata.Has("Values") {
		parsedResponse["ResultMetadata.Values"] = response.ResultMetadata.Get("Values")
	}

	return parsedResponse, nil
}

func parseBedrockOutput(output types.ConverseOutput) string {
	if output == nil {
		fmt.Println("output is nil")
		return ""
	}

	switch v := output.(type) {
	case *types.ConverseOutputMemberMessage:
		result := ""
		for _, contentBlock := range v.Value.Content {
			result += parseTextContentBlock(contentBlock)
		}

		if result == "" {
			fmt.Println("Content field is missing or not a text block")
		}
		return result

	case *types.UnknownUnionMember:
		fmt.Printf("unknown tag: %s\n", v.Tag)
		return string(v.Value)

	default:
		fmt.Println("output is of an unknown type")
		return ""
	}
}

func parseTextContentBlock(contentBlock types.ContentBlock) string {
	if contentBlock == nil {
		fmt.Println("contentBlock is nil")
		return ""
	}

	switch v := (contentBlock).(type) {
	case *types.ContentBlockMemberText:
		return v.Value

	case *types.UnknownUnionMember:
		fmt.Printf("unknown tag: %s\n", v.Tag)
		return string(v.Value)

	default:
		fmt.Println("contentBlock is of an unknown type")
		return ""
	}
}

func GenerateHash(text string) string {
	hasher := sha256.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
