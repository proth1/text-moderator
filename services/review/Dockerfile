FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o /bin/service ./services/review

FROM alpine:3.19
RUN apk --no-cache add ca-certificates wget
COPY --from=builder /bin/service /bin/service
EXPOSE 8083
CMD ["/bin/service"]
