FROM golang:1.23 AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN make build

FROM alpine:latest
COPY --from=builder /app/build/azure2openai /main
ENTRYPOINT ["/main"]
CMD ["--config", "/config.json"]
