FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the binary, ensuring it's statically linked
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /app/server ./cmd/server

FROM alpine:latest

# Create a non-root user to run the application
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

WORKDIR /home/appuser

# Copy only the compiled binary from the builder stage
COPY --from=builder /app/server .

# Expose the port the server listens on
EXPOSE 8080

# The command to run the application
CMD ["./server"]