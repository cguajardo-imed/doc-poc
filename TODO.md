# Summarizer API

## Local summarizer

- Golang: ["https://github.com/sugarme/transformer", ]
- Gemma4-e2b
- ollama: ["https://ollama.com/library/gemma4"]
- Huggingface: "https://huggingface.co/csebuetnlp/mT5_multilingual_XLSum"

Go-API -> receives content and passes to ollama endpoint loaded with Gemma4-e2b -> returns ollama result to Go-API -> Go-API return result to user.


## Content cache
Every time a summarization request comes to API, it makes the text extraction from input document, 
then hashes the content to get a unique identifier for that content, so the unique identifier is
used to query the database, if the identifier exists, then it returns the previously generated 
summary for that text, so with this technique we make sure to save AI resources and reduce AI costs
by avoiding sending same requests to AI provider.
