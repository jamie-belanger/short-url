# Docker build file
FROM golang:latest AS builder
WORKDIR /app

# Copy in source and grab dependencies
COPY go.mod go.sum ./
RUN go mod download
COPY ./src ./src

# Build the GO app
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o short-url-app ./src

# Stage 2: Lightweight container
FROM debian:stable-slim

WORKDIR /app

# Add new user 'appuser'. App should be run without root privileges as a security measure
RUN useradd -m -d /appuser -s /bin/bash appuser
USER appuser

COPY --from=builder /app/short-url-app .

# Define environment variables with default values
ENV PORT=8080
ENV DATABASE=memory

EXPOSE $PORT

CMD ["sh", "-c", "./short-url-app -port=$PORT -database=$DATABASE"]
