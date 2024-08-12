# azure2openai

A simple tool to proxy OpenAI‘s request to Azure OpenAI Service

### Configuration
```json
{
  "endpoint_format": {
    "chat_completions": "https://YOUR_NAME.openai.azure.com/openai/deployments/%s/chat/completions?api-version=2024-02-15-preview",
    "image_generations": "https://YOUR_NAME.openai.azure.com/openai/deployments/%s/images/generations?api-version=2024-02-15-preview",
    "models": "https://YOUR_NAME.openai.azure.com/openai/models?api-version=2024-06-01"
  },
  "model_map": {
    "gpt-3.5-turbo": "gpt-35-turbo"
  },
  "address":"0.0.0.0:8789"
}
```

### Usage
```bash
azure2openai -config config.json
```

### License
**azure2openai** is licensed under the MIT License.[See License](LICENSE) for details.