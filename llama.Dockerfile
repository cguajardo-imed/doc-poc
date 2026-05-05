FROM ubuntu:22.04

# download llama.cpp from https://github.com/ggml-org/llama.cpp/releases/download/b8972/llama-b8972-bin-ubuntu-x64.tar.gz

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    cmake \
    git \
    wget \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Download and extract llama.cpp binaries
RUN wget https://github.com/ggml-org/llama.cpp/releases/download/b8972/llama-b8972-bin-ubuntu-x64.tar.gz && \
    tar -xvf llama-b8972-bin-ubuntu-x64.tar.gz && \
    rm llama-b8972-bin-ubuntu-x64.tar.gz

# Copy the llama.cpp binaries to the /app directory
COPY llama-server /app/llama-server

# Download the model file Gemma-4-E2B-it-Q8_0.gguf from https://huggingface.co/ggml-org/gemma-4-E2B-it-GGUF/resolve/main/gemma-4-E2B-it-Q8_0.gguf?download=true
RUN wget https://huggingface.co/ggml-org/gemma-4-E2B-it-GGUF/resolve/main/gemma-4-E2B-it-Q8_0.gguf -O /models/gemma-4-E2B-it-Q8_0.gguf

EXPOSE 8000

CMD ["/app/llama-server", "--no-mmap", "--no-warmup", "--model", "/models/gemma-4-E2B-it-Q8_0.gguf", "--port", "8000", "--host", "0.0.0.0", "--predict", "512", "--temp", "0.5", "--top-p", "0.8", "--top-k", "20", "--presence-penalty", "1.5"]
