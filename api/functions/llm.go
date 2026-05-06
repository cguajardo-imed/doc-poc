package functions

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"summarizer-api/globals"

	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// callLLM sends the extracted text to the LLM and returns the desired output.
func CallLLM(ctx context.Context, config LLMConfig, messages []types.Message) (string, error) {
	switch config.Engine {
	case "AWS":
		if config.Type == globals.LLM_CALL_TYPE__INVOKE {
			return callBedrock(ctx, config)
		}
		return callBedrockChat(ctx, config, messages)
	case "OLLAMA":
		return callOllama(ctx, config)
	default:
		return "", fmt.Errorf("unknown engine: %s", config.Engine)
	}
}

func callBedrock(ctx context.Context, config LLMConfig) (string, error) {
	log.Printf("Sending request to Bedrock with model: %s\n", config.Model)

	if config.Temperature == nil {
		config.Temperature = &globals.CHAT_TEMPERATURE
	}

	bodyMap := NovaRequest{
		Messages: []Message{
			{
				Role:    "user",
				Content: []Content{{Text: config.UserPrompt}},
			},
		},
		InferenceConfig: &InferenceConfig{
			Temperature: float64(*config.Temperature),
		},
		System: []SystemContent{
			{Text: config.SystemPrompt},
		},
	}
	bodyBytes, marshalErr := json.Marshal(bodyMap)
	if marshalErr != nil {
		return "", fmt.Errorf("failed to build Bedrock request body: %w", marshalErr)
	}

	log.Printf("Bedrock request body prepared (size=%d bytes)\n", len(bodyBytes))

	result, err := globals.BedrockClient.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(config.Model),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
		Body:        bodyBytes,
	})

	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "no such host") {
			fmt.Printf("Error: The Bedrock service is not available in the selected region. Please double-check the service availability for your region at https://aws.amazon.com/about-aws/global-infrastructure/regional-product-services/.\n")
		} else if strings.Contains(errMsg, "Could not resolve the foundation model") {
			fmt.Printf("Error: Could not resolve the foundation model from model identifier: \"%v\". Please verify that the requested model exists and is accessible within the specified region.\n", config.Model)
		} else if strings.Contains(errMsg, "ValidationException") {
			fmt.Printf("Error: Malformed input request. Please reformat your input and try again. Details: %v\n", errMsg)
		} else {
			fmt.Printf("Error: Couldn't invoke Bedrock model. Here's why: %v\n", err)
		}
		return "", fmt.Errorf("failed to invoke model: %w", err)
	}

	var output InvokeModelResponse
	err = json.Unmarshal(result.Body, &output)
	if err != nil {
		return "", fmt.Errorf("failed to parse Bedrock response: %w", err)
	}
	fmt.Printf("Token Usage: \nInput: %v\nOutput: %v\nTotal: %v\n\n",
		output.Usage.InputTokens, output.Usage.OutputTokens, output.Usage.TotalTokens)
	strOutput := parseInvokeBedrockOutput(output.Output)
	return strOutput, nil
}

func parseInvokeBedrockOutput(message struct {
	Message Message `json:"message"`
}) string {
	if message.Message.Content == nil {
		return ""
	}

	if len(message.Message.Content) == 0 {
		return ""
	}

	return message.Message.Content[0].Text
}

func callBedrockChat(ctx context.Context, config LLMConfig, messages []types.Message) (string, error) {
	log.Println("Sending chat request to Bedrock")
	inferenceConfig := &types.InferenceConfiguration{}
	if config.Temperature != nil {
		inferenceConfig.Temperature = config.Temperature
	}

	result, err := globals.BedrockClient.Converse(ctx, &bedrockruntime.ConverseInput{
		ModelId:         aws.String(config.Model),
		InferenceConfig: inferenceConfig,
		Messages:        messages,
		System: []types.SystemContentBlock{
			&types.SystemContentBlockMemberText{Value: config.SystemPrompt},
		},
	})

	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "no such host") {
			fmt.Printf("Error: The Bedrock service is not available in the selected region. Please double-check the service availability for your region at https://aws.amazon.com/about-aws/global-infrastructure/regional-product-services/.\n")
		} else if strings.Contains(errMsg, "Could not resolve the foundation model") {
			fmt.Printf("Error: Could not resolve the foundation model from model identifier: \"%v\". Please verify that the requested model exists and is accessible within the specified region.\n", config.Model)
		} else {
			fmt.Printf("Error: Couldn't invoke Anthropic Claude. Here's why: %v\n", err)
		}
		os.Exit(1)
	}

	fmt.Printf("Token Usage: \nInput: %v\nOutput: %v\nTotal: %v\n\n",
		&result.Usage.InputTokens, &result.Usage.OutputTokens, &result.Usage.TotalTokens)
	output := parseBedrockOutput(result.Output)
	return output, nil
}

func callOllama(ctx context.Context, config LLMConfig) (string, error) {
	log.Printf("Sending request to Ollama with model: %s\n", config.Model)
	client := openai.NewClient(
		option.WithBaseURL(globals.OLLAMA_URL),
		option.WithAPIKey("not-needed-for-ollama"),
	)

	params := openai.ChatCompletionNewParams{
		Model: openai.ChatModel(config.Model),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(config.SystemPrompt),
			openai.UserMessage(config.UserPrompt),
		},
	}

	// chat_template_kwargs is a llama.cpp-specific field not part of the OpenAI
	// spec, so we inject it via SetExtraFields.
	params.SetExtraFields(map[string]any{
		"stream":  false,
		"options": map[string]any{"think": false},
	})

	resp, err := client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned by the model")
	}

	return resp.Choices[0].Message.Content, nil
}
