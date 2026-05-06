package globals

import (
	"summarizer-api/db"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

var BedrockClient *bedrockruntime.Client = nil

const REGION = "us-east-1"

var OLLAMA_URL = ""

const (
	LLM_CALL_TYPE__INVOKE = "INVOKE"
	LLM_CALL_TYPE__CHAT   = "CHAT"
)

var DB *db.DB = nil

var (
	SUMMARY_TEMPERATURE float32 = 0.4
	CHAT_TEMPERATURE    float32 = 0.7
)
