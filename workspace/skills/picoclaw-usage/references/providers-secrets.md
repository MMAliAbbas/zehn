# Providers, Models, And Secrets

## Model Configuration

PicoClaw uses `model_list`. Prefer explicit `provider` plus native `model`:

```json
{
  "model_list": [
    {
      "model_name": "primary",
      "provider": "openai",
      "model": "gpt-5.4"
    },
    {
      "model_name": "fallback",
      "provider": "anthropic",
      "model": "claude-sonnet-4.6"
    }
  ],
  "agents": {
    "defaults": {
      "model_name": "primary"
    }
  }
}
```

Legacy `provider/model` form still works when `provider` is omitted, for example `"model": "openai/gpt-5.4"`.

## Common Providers

Supported providers include OpenAI, Anthropic, Gemini, OpenRouter, DeepSeek, Qwen, Zhipu/GLM, Groq, Mistral, Cerebras, Moonshot/Kimi, Volcengine, NVIDIA, Ollama, LM Studio, vLLM, LiteLLM, Azure, GitHub Copilot, Antigravity, and Bedrock with build tags.

Use local providers for cost or privacy, but do not assume local quality is enough for business-critical engineering work on 16 GB RAM.

## Secrets

Store secrets in `.security.yml`, not `config.json`. Values map directly by structure.

```yaml
model_list:
  primary:
    api_keys:
      - "sk-..."
  fallback:
    api_keys:
      - "sk-ant-..."
channels:
  telegram:
    token: "123:abc"
  discord:
    token: "discord-token"
web:
  brave:
    api_keys:
      - "BSA..."
skills:
  github:
    token: "ghp_..."
```

Set permissions:

```bash
chmod 600 ~/.picoclaw/.security.yml
```

`config.json` and `.security.yml` are written with restrictive permissions by PicoClaw. API keys should use `api_keys` arrays for models, even for one key.

## Voice

Voice transcription can use `voice.model_name` pointing at a multimodal model. If not configured, Groq Whisper remains a fallback when Groq credentials are present.

