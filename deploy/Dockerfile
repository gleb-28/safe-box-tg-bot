FROM golang:1.25 AS build
WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o meme-bot ./cmd/bot

FROM debian:bookworm-slim AS run
RUN apt update && apt install -y \
    curl \
    sqlite3 \
    ca-certificates \
 && rm -rf /var/lib/apt/lists/*
WORKDIR /app
RUN mkdir -p /app/data /app/.cache
COPY --from=build /build/safe-box-bot .
RUN chmod -R 777 /app
ENV HOME=/app
EXPOSE 8081
ENTRYPOINT ["./safe-box-bot"]