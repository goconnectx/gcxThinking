# GCx Thinking

A Go-based toolkit and client libraries for AI/LLM evaluation systems.

## Project Structure

| Directory | Description |
|-----------|-------------|
| `systemone/` | Go client library for the TypeSafe SystemOne evaluation API |
| `agents.md` | Project overview and setup notes |
| `API_Reference.md` | External API reference documentation |

## SystemOne Go Client

A feature-complete Go client for the Local SystemOne (TypeSafe) evaluation API.

### Features

- **Typed Evaluations**: Supports `noul` (yes/no), `choice` (multiple choice), and `score` (rating) question types
- **Structured State**: Pass plain text or structured JSON as evaluation state
- **Type-Safe Answers**: Accessor methods (`AsNoul`, `AsChoice`, `AsScore`) with proper error handling
- **Functional Options**: Configure base URL, API key, timeout via clean option pattern
- **Error Handling**: Built-in helpers for retryable errors (rate limits, overload)
- **Docker Support**: Multi-stage Dockerfile using Go 1.26

### Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/goconnectx/gcxThinking/systemone"
)

func main() {
    client, err := systemone.NewClient(
        systemone.WithBaseURL("http://localhost:11434/v1/systemone"),
        systemone.WithAPIKey("your-api-key"),
    )
    if err != nil {
        log.Fatal(err)
    }

    resp, err := client.Evaluate(context.Background(), &systemone.EvaluateRequest{
        Model: "jev-latest",
        State: "Help! My payouts have been failing for 3 days.",
        Questions: map[string]systemone.Question{
            "is_urgent": systemone.NewNoulQuestionWithCriteria(
                "Does this convey urgency?",
                "Explicitly time-sensitive",
                "No urgency expressed",
            ),
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    urgency, _ := resp.Answers["is_urgent"].AsNoul()
    fmt.Printf("Urgency: %.2f\n", urgency)
}
```

### Docker Build

```bash
cd systemone
docker build -t systemone-go:latest .
```

See [systemone/README.md](systemone/README.md) for full documentation.

## Development

- **Language**: Go 1.26
- **Testing**: `go test ./...`
- **Formatting**: `gofmt -w .`

## License

MIT
