FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/service ./services/policy-engine

FROM alpine:3.19
RUN apk --no-cache add ca-certificates && \
    addgroup -S appgroup && adduser -S appuser -G appgroup
COPY --from=builder /bin/service /bin/service
USER appuser
EXPOSE 8082
CMD ["/bin/service"]
