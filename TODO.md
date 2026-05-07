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

## QA generation post summarization
Every time a document is requested for summarization, we also generate a question and answer pair set
for that document, so the user can ask the question and get the answer.
The LLM must generate all the questions and anwers as possible, this will allow us to have a good
QA set for the document, and also a good base to make a search engine for the document with AI improvements.

## Search engine (Semantic Search)
We will use the LLM to generate a dataset of Q&A pairs, then we will use this dataset to train a search engine
and get the answer to the question using semantic search.
- It will be neccessary to use an embedding model to generate the Q&A pairs, so we can use the embedding
model to generate the Q&A pairs, and then use the embedding model to train the search engine.