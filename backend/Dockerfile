FROM golang:1.24 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /bin/heimdall ./cmd/heimdall

# Download Lumber model files and ONNX runtime
FROM debian:bookworm-slim AS models
RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*
WORKDIR /tmp
RUN mkdir -p /models/2_Dense && \
    curl -fsSL -o /models/model_quantized.onnx \
      https://huggingface.co/onnx-community/mdbr-leaf-mt-ONNX/resolve/main/onnx/model_quantized.onnx && \
    curl -fsSL -o /models/model_quantized.onnx_data \
      https://huggingface.co/onnx-community/mdbr-leaf-mt-ONNX/resolve/main/onnx/model_quantized.onnx_data && \
    curl -fsSL -o /models/vocab.txt \
      https://huggingface.co/onnx-community/mdbr-leaf-mt-ONNX/resolve/main/vocab.txt && \
    curl -fsSL -o /models/tokenizer_config.json \
      https://huggingface.co/onnx-community/mdbr-leaf-mt-ONNX/resolve/main/tokenizer_config.json && \
    curl -fsSL -o /models/2_Dense/model.safetensors \
      https://huggingface.co/MongoDB/mdbr-leaf-mt/resolve/main/2_Dense/model.safetensors && \
    curl -fsSL -o /models/2_Dense/config.json \
      https://huggingface.co/MongoDB/mdbr-leaf-mt/resolve/main/2_Dense/config.json
# ONNX Runtime shared library — Lumber expects libonnxruntime.so in the model directory
RUN curl -fSL "https://github.com/microsoft/onnxruntime/releases/download/v1.24.1/onnxruntime-linux-x64-1.24.1.tgz" \
      -o /tmp/ort.tgz && \
    tar xzf /tmp/ort.tgz -C /tmp && \
    cp /tmp/onnxruntime-linux-x64-1.24.1/lib/libonnxruntime.so.1.24.1 /models/libonnxruntime.so && \
    rm -rf /tmp/ort.tgz /tmp/onnxruntime-linux-x64-1.24.1

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=build /bin/heimdall /usr/local/bin/heimdall
COPY --from=models /models /opt/lumber/models
EXPOSE 8080
ENTRYPOINT ["heimdall"]
