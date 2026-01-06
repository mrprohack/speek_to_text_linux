# VoiceType - Improvement Plan

## Current Status
✅ Working core functionality  
❌ Some bugs on shutdown  
❌ Basic logging  
❌ Simple GUI  

---

## 1. Better Logging System

### Goals
- Structured JSON logs for easy parsing
- Log levels: DEBUG, INFO, WARN, ERROR
- Log to file + console
- Test mode with verbose output

### Implementation
```go
// New logging package
type Logger struct {
    level   LogLevel
    output  io.Writer
    format  string
}

func NewLogger(level LogLevel, output io.Writer) *Logger

// Methods
func (l *Logger) Debug(msg string, args ...)
func (l *Logger) Info(msg string, args ...)
func (l *Logger) Warn(msg string, args ...)
func (l *Logger) Error(msg string, args ...)

// Test command
./VoiceType --log-level=debug --log-file=voicetype.log
```

### Log Format
```
2026-01-06T13:22:53Z [INFO] audio: Audio system initialized with device: default
2026-01-06T13:22:53Z [DEBUG] hotkey: Polling for key: Ctrl+Space
2026-01-06T13:22:53Z [ERROR] api: Transcription failed: rate limit exceeded
```

---

## 2. Testing Tools

### CLI Test Mode
```bash
# Test audio input
./VoiceType --test-audio

# Test API connection
./VoiceType --test-api

# Test hotkey detection
./VoiceType --test-hotkey

# Full test suite
./VoiceType --test-all
```

### Test Audio Script
```bash
#!/bin/bash
# test_voicetype.sh

echo "=== VoiceType Test Suite ==="

# Test 1: Audio device
echo "[1/4] Testing audio device..."
arecord -d 2 -f cd -t wav /tmp/test.wav

# Test 2: API key
echo "[2/4] Testing API..."
curl -s -H "Authorization: Bearer $GROQ_API_KEY" \
  https://api.groq.com/openai/v1/models

# Test 3: Transcription
echo "[3/4] Testing transcription..."
./VoiceType --transcribe=/tmp/test.wav

# Test 4: Hotkey
echo "[4/4] Testing hotkey..."
echo "Press Ctrl+Space within 5 seconds..."
timeout 5 ./VoiceType --detect-hotkey

echo "=== Tests Complete ==="
```

---

## 3. Enhanced GUI

### Better Popup Design
```
┌─────────────────────────┐
│     🎙️ VoiceType        │
│                         │
│   █ █ █ █ █ █ █ █      │
│   Recording...          │
│                         │
│   ⏱️ 00:03              │
│                         │
│   Press Ctrl+Space      │
│   to stop               │
└─────────────────────────┘
```

### Features to Add
1. **Recording timer** - Show duration
2. **Sound level meter** - Real-time audio visualization  
3. **Transcription preview** - Show partial results while recording
4. **History** - Show last 5 transcriptions
5. **Settings button** - Configure hotkey, device, model
6. **Minimize to tray** - Keep running in background

### GUI Menu
```
File:
  - Settings...
  - Test Audio Device
  - Exit

View:
  - Show History
  - Clear History

Help:
  - About
  - Documentation
```

---

## 4. Bug Fixes

### Critical Bugs
| Bug | Fix | Priority |
|-----|-----|----------|
| Fyne thread error on shutdown | Use proper goroutine sync | HIGH |
| Warnings in Wayland | Add ydotool support | MEDIUM |
| Audio device selection | Add device picker dialog | MEDIUM |

### Bug Tracking
```
cmd/voicetype-gui/main.go:528  - Fyne.Do[AndWait] error
cmd/voicetype-gui/main.go:281  - Fyne thread error
internal/hotkey/listener.go   - Wayland polling warning
```

---

## 5. File Structure

```
speek_to_text_linux/
├── cmd/
│   ├── voicetype/main.go         # CLI version
│   └── voicetype-gui/main.go     # GUI version
├── internal/
│   ├── api/client.go             # API client
│   ├── audio/system.go           # Audio capture
│   ├── clipboard/system.go       # Clipboard
│   ├── hotkey/listener.go        # Hotkey detection
│   └── logger/                   # NEW: Logging package
├── pkg/
│   ├── config/config.go          # Config
│   └── errors/handler.go         # Errors
├── test/
│   ├── test_audio.sh             # Audio tests
│   ├── test_api.sh               # API tests
│   └── test_full.sh              # Full test suite
└── tools/
    ├── audio_check.py            # Audio diagnostic
    └── log_viewer.py             # Log viewer GUI
```

---

## 6. Implementation Tasks

### Phase 1: Logging (Day 1)
- [ ] Create `internal/logger` package
- [ ] Add log levels
- [ ] Add file output
- [ ] Add `--verbose` flag
- [ ] Add `--log-file` flag

### Phase 2: Testing (Day 2)
- [ ] Create `test/test_audio.sh`
- [ ] Create `test/test_api.sh`  
- [ ] Add `--test-audio` flag
- [ ] Add `--test-api` flag
- [ ] Add `--test-all` flag

### Phase 3: GUI Improvements (Day 3)
- [ ] Add recording timer
- [ ] Add sound level meter
- [ ] Add settings dialog
- [ ] Add history view
- [ ] Fix shutdown bugs

### Phase 4: Bug Fixes (Day 4)
- [ ] Fix Fyne thread error
- [ ] Add ydotool support for Wayland
- [ ] Add device picker
- [ ] Test on real user system

---

## 7. Commands Reference

### Current Commands
```bash
./VoiceType           # CLI version
./VoiceType-gui       # GUI version
./VoiceType --help    # Show help
./VoiceType -hotkey=F6 # Set custom hotkey
./VoiceType -device=hw:0 # Set audio device
```

### New Commands (Proposed)
```bash
./VoiceType --verbose              # Debug logs
./VoiceType --log-file=app.log     # Save logs to file
./VoiceType --test-audio           # Test microphone
./VoiceType --test-api             # Test API connection
./VoiceType --test-hotkey          # Test hotkey detection
./VoiceType --test-all             # Run all tests
./VoiceType --transcribe=file.wav  # Transcribe audio file
./VoiceType --list-devices         # List audio devices
./VoiceType --settings             # Open settings GUI
./VoiceType --history              # Show transcription history
```

---

## 8. Success Criteria

- ✅ No errors on shutdown
- ✅ Clean logs without warnings
- ✅ All tests pass
- ✅ User-friendly GUI
- ✅ Easy troubleshooting

---

## Next Steps

Say **"NEXT"** and I'll start implementing:

1. **Better logging system** first
2. **Testing tools** second  
3. **GUI improvements** third
4. **Bug fixes** last

Or tell me which priority you'd like me to start with!
