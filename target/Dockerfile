FROM golang:1.23-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o target ./main.go

FROM alpine:3.20 AS llama-builder
WORKDIR /src
RUN apk add --no-cache git build-base cmake linux-headers
RUN git clone --depth 1 https://github.com/ggerganov/llama.cpp.git .
RUN cmake -S . -B build -DLLAMA_BUILD_SERVER=ON
RUN cmake --build build --target llama-server -j 1

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates libstdc++ libgomp
COPY --from=go-builder /app/target ./target
COPY --from=llama-builder /src/build/bin /usr/local/lib/llama
COPY --from=llama-builder /src/build/bin/llama-server /usr/local/bin/llama-server
ENV LD_LIBRARY_PATH=/usr/local/lib/llama
ENV LLAMA_SERVER_BIN=/usr/local/bin/llama-server
CMD ["./target"]