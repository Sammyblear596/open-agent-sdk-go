# Open Agent SDK (Go)

A lightweight, open-source Go SDK for building AI agents. Run the full agent loop in-process — no CLI or subprocess required. Deploy anywhere: cloud, serverless, Docker, CI/CD.

Also available in [TypeScript](https://github.com/codeany-ai/open-agent-sdk-typescript).

## Features

- **Agent Loop** — Streaming agentic loop with tool execution, multi-turn conversations, and cost tracking
- **Built-in Tools** — Bash, Read, Write, Edit, Glob, Grep, WebFetch, WebSearch, Agent (subagents), AskUser, TaskTools, ToolSearch
- **MCP Support** — Connect to MCP servers via stdio, HTTP, and SSE transports
- **Permission System** — Configurable tool approval with allow/deny rules and filesystem path validation
- **Hook System** — Pre/post tool-use hooks, post-sampling hooks, structured output enforcement
- **Extended Thinking** — Support for extended thinking with budget tokens
- **Cost Tracking** — Per-model token usage, API/tool duration, code change stats
- **Custom Tools** — Implement the `Tool` interface to add your own tools

## Quick Start

```bash
go get github.com/codeany-ai/open-agent-sdk-go
```

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/codeany-ai/open-agent-sdk-go/agent"
    "github.com/codeany-ai/open-agent-sdk-go/types"
)

func main() {
    a := agent.New(agent.Options{
        Model:  "sonnet-4-6",
        APIKey: os.Getenv("CODEANY_API_KEY"),
    })
    defer a.Close()

    ctx := context.Background()

    // Streaming
    events, errs := a.Query(ctx, "What files are in this directory?")
    for event := range events {
        if event.Type == types.MessageTypeAssistant && event.Message != nil {
            fmt.Print(types.ExtractText(event.Message))
        }
    }
    if err := <-errs; err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    }

    // Or use the blocking API
    result, _ := a.Prompt(ctx, "Count lines in go.mod")
    fmt.Println(result.Text)
}
```

## Examples

| #   | Example                                                   | Description                                      |
| --- | --------------------------------------------------------- | ------------------------------------------------ |
| 01  | [Simple Query](examples/01-simple-query/)                 | Streaming query with tool calls                  |
| 02  | [Multi-Tool](examples/02-multi-tool/)                     | Glob + Bash multi-tool orchestration             |
| 03  | [Multi-Turn](examples/03-multi-turn/)                     | Multi-turn conversation with session persistence |
| 04  | [Prompt API](examples/04-prompt-api/)                     | Blocking `Prompt()` for one-shot queries         |
| 05  | [Custom System Prompt](examples/05-custom-system-prompt/) | Custom system prompt for code review             |
| 06  | [MCP Server](examples/06-mcp-server/)                     | MCP server integration (stdio transport)         |
| 07  | [Custom Tools](examples/07-custom-tools/)                 | Define and use custom tools                      |
| 08  | [One-shot Query](examples/08-official-api-compat/)        | Quick one-shot agent query                       |
| 09  | [Subagents](examples/09-subagents/)                       | Specialized subagent with restricted tools       |
| 10  | [Permissions](examples/10-permissions/)                   | Read-only agent with AllowedTools                |
| 11  | [Web Chat](examples/web/)                                 | Web-based chat UI with streaming                 |

Run any example:

```bash
export CODEANY_BASE_URL=https://openrouter.ai/api
export CODEANY_API_KEY=your-api-key
export CODEANY_MODEL=anthropic/claude-sonnet-4
go run ./examples/01-simple-query/
```

For the web chat UI:

```bash
go run ./examples/web/
# Open http://localhost:8082
```

## Custom Tools

Implement the `types.Tool` interface:

```go
type MyTool struct{}

func (t *MyTool) Name() string                                    { return "MyTool" }
func (t *MyTool) Description() string                             { return "Does something useful" }
func (t *MyTool) InputSchema() types.ToolInputSchema              { return types.ToolInputSchema{...} }
func (t *MyTool) IsConcurrencySafe(map[string]interface{}) bool   { return true }
func (t *MyTool) IsReadOnly(map[string]interface{}) bool          { return true }
func (t *MyTool) Call(ctx context.Context, input map[string]interface{}, tCtx *types.ToolUseContext) (*types.ToolResult, error) {
    // Your logic here
    return &types.ToolResult{
        Content: []types.ContentBlock{{Type: types.ContentBlockText, Text: "result"}},
    }, nil
}

// Use it
a := agent.New(agent.Options{
    CustomTools: []types.Tool{&MyTool{}},
})
```

## MCP Servers

```go
a := agent.New(agent.Options{
    MCPServers: map[string]types.MCPServerConfig{
        "filesystem": {
            Type:    types.MCPTransportStdio,
            Command: "npx",
            Args:    []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
        },
    },
})
a.Init(ctx) // Connects to MCP servers
```

## Architecture

```
open-agent-sdk-go/
├── agent/              # Agent loop, query engine, options
├── api/                # Messages API client (streaming + non-streaming)
├── types/              # Core types: Message, Tool, ContentBlock, MCP
├── tools/              # Built-in tool implementations + registry + executor
│   └── diff/           # Unified diff generation
├── mcp/                # MCP client (stdio, HTTP, SSE) + resources + reconnection
├── permissions/        # Permission rules, filesystem validation
├── hooks/              # Pre/post tool-use hooks
├── costtracker/        # Token usage and cost tracking
├── context/            # System/user context injection (git status, CODEANY.md)
├── history/            # Conversation history persistence (JSONL)
└── examples/           # 11 runnable examples
```

## Configuration

Environment variables:

| Variable                     | Description                                  |
| ---------------------------- | -------------------------------------------- |
| `CODEANY_API_KEY`            | API key (required)                           |
| `CODEANY_MODEL`              | Default model (default: `sonnet-4-6`)        |
| `CODEANY_BASE_URL`           | API base URL override                        |
| `CODEANY_CUSTOM_HEADERS`     | Custom headers (comma-separated `key:value`) |
| `API_TIMEOUT_MS`             | API request timeout in ms                    |
| `HTTPS_PROXY` / `HTTP_PROXY` | Proxy URL                                    |

## Links

- Website: [codeany.ai](https://codeany.ai)
- TypeScript SDK: [github.com/codeany-ai/open-agent-sdk-typescript](https://github.com/codeany-ai/open-agent-sdk-typescript)
- Issues: [github.com/codeany-ai/open-agent-sdk-go/issues](https://github.com/codeany-ai/open-agent-sdk-go/issues)

## License

MIT — see [LICENSE](LICENSE)
