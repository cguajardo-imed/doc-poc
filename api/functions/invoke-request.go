package functions

import "github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"

type NovaRequest struct {
	Messages        []Message        `json:"messages"`
	System          []SystemContent  `json:"system,omitempty"`
	InferenceConfig *InferenceConfig `json:"inferenceConfig,omitempty"`
}

type Message struct {
	Role    string    `json:"role"` // "user" or "assistant"
	Content []Content `json:"content"`
}

type Content struct {
	Text  string `json:"text,omitempty"`
	Image *Image `json:"image,omitempty"`
}

type Image struct {
	Format string `json:"format"` // "png", "jpeg", "webp"
	Source Source `json:"source"`
}

type Source struct {
	Bytes []byte `json:"bytes"` // Standard library json will base64 encode this
}

type SystemContent struct {
	Text string `json:"text"`
}

type InferenceConfig struct {
	MaxTokens   int     `json:"maxTokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
	TopP        float64 `json:"topP,omitempty"`
	TopK        int     `json:"topK,omitempty"`
}

/*
{
	"output":{
		"message":{
			"content":[
				{"text":"..."}
			],
			"role":"assistant"
		}
	},
	"stopReason":"end_turn",
	"usage":{
		"inputTokens":16779,
		"outputTokens":1411,
		"totalTokens":18190,
		"cacheReadInputTokenCount":0,
		"cacheWriteInputTokenCount":0
	}
}
*/

type InvokeModelResponse struct {
	Output struct {
		Message Message `json:"message"`
	} `json:"output"`
	Usage types.TokenUsage `json:"usage"`
}
