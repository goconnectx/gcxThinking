# SystemOne Go Client

A Go client library for the Local SystemOne (TypeSafe) evaluation API.

## Features

- Idiomatic Go with full `context.Context` support
- Functional options pattern for configuration
- Strongly typed request/response structs matching the TypeSafe API
- Type-safe accessors for Noul, Choice, and Score answer types
- Built-in error handling with retry helpers for rate limits and overload
- Support for structured `state` and `instructions`

## Installation

```bash
go get github.com/goconnectx/gcxThinking/systemone
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/goconnectx/gcxThinking/systemone"
)

func main() {
    ctx := context.Background()

    client, err := systemone.NewClient(
        systemone.WithBaseURL("http://localhost:11434/v1/systemone"),
        systemone.WithAPIKey("your-api-key"),
    )
    if err != nil {
        log.Fatal(err)
    }

    resp, err := client.Evaluate(ctx, &systemone.EvaluateRequest{
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

## Configuration Options

| Option | Description |
|--------|-------------|
| `WithBaseURL(url)` | Set the API base URL (default: `http://localhost:11434/v1/systemone`) |
| `WithAPIKey(key)` | Set the Bearer token for authentication |
| `WithTimeout(duration)` | Set the HTTP client timeout |
| `WithHTTPClient(client)` | Provide a custom `*http.Client` |
| `WithUserAgent(ua)` | Set a custom User-Agent string |

## Question Types

### Noul (yes/no)

Returns a probability from 0 (no) to 1 (yes).

```go
question := systemone.NewNoulQuestion("Does this convey urgency?")
// or with criteria:
question := systemone.NewNoulQuestionWithCriteria(
    "Does this convey urgency?",
    "Explicitly time-sensitive",
    "No urgency expressed",
)
```

```go
val, err := resp.Answers["is_urgent"].AsNoul()
// val is a float64 between 0 and 1
```

### Choice

Picks one option from a set you define. Returns the chosen option and probability distribution.

```go
question := systemone.NewChoiceQuestion(
    "Which team should handle this?",
    map[string]interface{}{
        "billing":   "Payments, invoicing, refunds",
        "technical": "Bugs, outages, integrations",
        "sales":     "Pricing, upgrades, new accounts",
    },
)
```

```go
choice, probabilities, confidence, err := resp.Answers["department"].AsChoice()
```

### Score

Rates the state along a rubric you define. Returns a probability-weighted value.

```go
question := systemone.NewScoreQuestion(
    "How frustrated is the customer?",
    []string{"Calm", "Frustrated", "Very angry"},
)
```

```go
score, legend, probabilities, confidence, err := resp.Answers["frustration"].AsScore()
```

## Structured State and Instructions

`state` and `instructions` can be plain strings or structured data:

```go
resp, err := client.Evaluate(ctx, &systemone.EvaluateRequest{
    Model: "jev-latest",
    State: map[string]interface{}{
        "customer_message": "Help! My payouts have been failing for 3 days.",
        "account_tier":   "enterprise",
    },
    Questions: map[string]systemone.Question{
        "needs_escalation": systemone.NewNoulQuestionWithCriteria(
            map[string]interface{}{
                "question": "Should this be escalated to the enterprise team?",
                "account_tier": "enterprise",
            },
            "Enterprise accounts get priority escalation",
            "Standard support is sufficient",
        ),
    },
})
```

## Error Handling

The library returns `*systemone.APIError` for non-2xx responses:

```go
resp, err := client.Evaluate(ctx, req)
if err != nil {
    if systemone.IsUnauthorized(err) {
        log.Println("Invalid API key")
    } else if systemone.IsUnprocessable(err) {
        log.Println("Request validation failed")
    } else if systemone.IsRetryable(err) {
        log.Println("Should retry with exponential backoff")
    } else {
        log.Printf("API error: %v\n", err)
    }
}
```

| Helper | Status |
|--------|--------|
| `IsUnauthorized(err)` | 401 |
| `IsUnprocessable(err)` | 422 |
| `IsRateLimited(err)` | 429 |
| `IsOverloaded(err)` | 529 |
| `IsRetryable(err)` | 429 or 529 |

## Testing

```bash
cd systemone
go test ./...
```

## Building with Docker

A Dockerfile is provided using Go 1.26:

```bash
cd systemone
docker build -t systemone-go:latest .
```

Or using Docker Compose:

```bash
cd systemone
docker-compose up --build
```

## License

MIT
