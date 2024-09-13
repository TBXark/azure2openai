FROM alpine:latest
COPY /build/linux_x86/azure2openai /main
ENTRYPOINT ["/main"]
CMD ["--config", "/config/config.json"]