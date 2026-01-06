# AGENTS.md - VoiceType Development Guide

Quick reference for agentic coding agents working on VoiceType.

## Build Commands
```bash
make build          # Standard build
make build-debug    # With debug symbols
make build-all      # All platforms (amd64, arm64)
make clean          # Clean artifacts
```

## Test Commands
```bash
make test           # All tests
make test-coverage  # With coverage
go test ./pkg/wav/ -v              # Single package
go test ./pkg/config/ -v           # Config tests
go test ./pkg/errors/ -v           # Error tests
go test ./pkg/wav/ -v -run TestEncode  # Single function
make verify         # Build + tests
```

## Code Quality
```bash
make fmt            # Format code
make fmt-check      # Check formatting
make vet            # Vet code
make lint           # Run linter
```

## Run the App
```bash
export GROQ_API_KEY="your_key"
./VoiceType -hotkey=F12 -v -no-notify -device=hw:0
```

## Code Style

### Project Structure
- **cmd/**: Main entry point
- **internal/**: Private code (api, audio, clipboard, hotkey, notify, ui)
- **pkg/**: Public libraries (config, errors, wav)

### Imports
```go
import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "speek_to_text_linux/pkg/errors"
)
// Group: std lib → internal
```

### Naming
- **Packages**: lowercase (api, wav, config)
- **Types**: PascalCase (Client, Response, System)
- **Variables**: camelCase (httpClient, audioBuffer)
- **Constants**: camelCase, UPPER_Snake_Case for exported
- **Functions**: camelCase, PascalCase for exported
- **Interfaces**: -er suffix (Reader, Writer)

### Error Handling
- Use `*errors.Error` from `pkg/errors`
- Pass `*errors.Handler` to components
- Log: `errHandler.Error("operation failed: %v", err)`
- Wrap: `return errors.Wrap(err, ErrorTypeAPI, "msg")`
- Check: `IsType(err, ErrorTypeAPI)`
- Never suppress errors with `_`

### Types
```go
// Unexported fields with public getters
type Client struct {
    apiKey     string
    httpClient *http.Client
}

// Config with JSON tags
type Config struct {
    GROQ_API_KEY string `json:"groq_api_key"`
    Hotkey       string `json:"hotkey"`
}

func NewClient(apiKey string, errHandler *errors.Handler) *Client
```

### Context
- Pass `context.Context` as first parameter
- Use `context.Background()` for top-level
- Check cancellation in long-running ops

### Concurrency
- Goroutines for I/O (API, audio)
- `sync.WaitGroup` for sync
- `sync.Mutex` for shared state
- Channels or mutex for data passing
- Avoid goroutine leaks

### Configuration
- Environment variables for secrets (GROQ_API_KEY)
- `pkg/config.Load()` for config
- Environment variable overrides
- Default values for optional config

### Documentation
- Package doc on first line
- Public function docs with params/returns
- Complex logic needs inline comments
- Update TODOs: "TODO: [desc] - [issue]"

### Testing
- Test files: `*_test.go` in same package
- Test functions: `TestFunctionName`
- Table-driven tests for multiple scenarios
- Mock external deps
- Aim for >80% coverage on pkg

### Security
- Never log secrets/API keys
- Environment variables for sensitive data
- Clear sensitive data from memory
- Validate external inputs
- No telemetry/analytics

### Performance
- Reuse buffers
- Stream data (don't load entirely)
- Profile before optimizing
- Target: <50MB RAM, <1s latency

## Important Notes
- **API Key**: Requires `GROQ_API_KEY` env var
- **Environment**: Needs X11/Wayland display
- **Platform**: Linux-only (X11/Wayland)
- **Privacy**: Zero storage, RAM-only processing

## File Paths
- Main: `cmd/voicetype/main.go`
- Tests: `pkg/*/*_test.go`
- Build: `Makefile`
- Docs: `README.md`, `VOICE_TYPE_LINUX_APP.md`
