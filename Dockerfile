FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server.go

FROM alpine:latest
RUN adduser -D nonroot
WORKDIR /home/nonroot/
COPY --from=builder /server .
COPY --from=builder /app/prompts ./prompts
RUN chown -R nonroot:nonroot .
USER nonroot

EXPOSE 8080
CMD ["./server"]