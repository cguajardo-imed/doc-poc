# app.py
from contextlib import asynccontextmanager
from typing import Literal

import torch
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field
from transformers import AutoModelForSeq2SeqLM, AutoTokenizer

MODEL_NAME = "mrm8488/bert2bert_shared-spanish-finetuned-summarization"


class SummarizeRequest(BaseModel):
    text: str = Field(..., min_length=1)
    max_input_tokens: int = Field(512, ge=32, le=4096)
    max_new_tokens: int = Field(128, ge=8, le=512)
    min_new_tokens: int = Field(30, ge=1, le=512)
    num_beams: int = Field(4, ge=1, le=16)
    length_penalty: float = Field(1.0, ge=0.1, le=5.0)
    no_repeat_ngram_size: int = Field(3, ge=0, le=10)


class SummarizeResponse(BaseModel):
    model: str
    summary: str
    input_chars: int
    input_tokens: int


class AppState:
    tokenizer = None
    model = None
    device = None


state = AppState()


@asynccontextmanager
async def lifespan(app: FastAPI):
    state.device = "cuda" if torch.cuda.is_available() else "cpu"
    state.tokenizer = AutoTokenizer.from_pretrained(MODEL_NAME)
    state.model = AutoModelForSeq2SeqLM.from_pretrained(MODEL_NAME).to(state.device)
    state.model.eval()
    yield
    state.tokenizer = None
    state.model = None


app = FastAPI(title="Spanish Summarizer", lifespan=lifespan)


@app.get("/health")
def health():
    return {"status": "ok", "model_loaded": state.model is not None}


@app.post("/summarize", response_model=SummarizeResponse)
def summarize(payload: SummarizeRequest):
    if not payload.text.strip():
        raise HTTPException(status_code=400, detail="text cannot be empty")

    inputs = state.tokenizer(
        payload.text,
        return_tensors="pt",
        truncation=True,
        max_length=payload.max_input_tokens,
    ).to(state.device)

    with torch.inference_mode():
        output_ids = state.model.generate(
            **inputs,
            max_new_tokens=payload.max_new_tokens,
            min_new_tokens=payload.min_new_tokens,
            num_beams=payload.num_beams,
            length_penalty=payload.length_penalty,
            no_repeat_ngram_size=payload.no_repeat_ngram_size,
            early_stopping=True,
        )

    summary = state.tokenizer.decode(output_ids[0], skip_special_tokens=True).strip()

    return SummarizeResponse(
        model=MODEL_NAME,
        summary=summary,
        input_chars=len(payload.text),
        input_tokens=int(inputs["input_ids"].shape[1]),
    )
