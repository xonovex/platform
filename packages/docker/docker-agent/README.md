# Docker Agent

Docker Compose setup for running AI coding agents with custom provider support via [CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI).

## Usage

```bash
docker compose -f packages/docker/docker-agent/compose.yaml run --rm ai-agent
```

## Services

| Service                  | Provider      | Description                |
| ------------------------ | ------------- | -------------------------- |
| `ai-agent`               | Default       | Pass-through Anthropic API |
| `ai-agent-glm`           | GLM           | Zhipu AI GLM-4 models      |
| `ai-agent-gemini`        | Gemini        | Google Gemini 3.x models   |
| `ai-agent-gemini-claude` | Gemini-Claude | Hybrid thinking models     |
| `ai-agent-gpt5-codex`    | GPT-5 Codex   | OpenAI models              |

## Environment Variables

| Variable               | Description                                     |
| ---------------------- | ----------------------------------------------- |
| `ANTHROPIC_AUTH_TOKEN` | Anthropic API token (for default provider)      |
| `ZAI_AUTH_TOKEN`       | Z.AI API token (for GLM provider)               |
| `CLI_PROXY_API_KEY`    | CLI Proxy API key (for Gemini/GPT providers)    |
| `AGENT_WORK_DIR`       | Working directory to mount (defaults to `$PWD`) |
