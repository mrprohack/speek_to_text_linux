# VoiceType - Native Linux GUI Plan

## Current State
- ✅ CLI app working (VoiceType binary)
- ✅ Web UI working (browser-based, not wanted)
- ❌ No native Linux GUI indicator

## What You Want
- **Native Linux GUI** (not web browser)
- **Background app** (no main window)
- **Small popup indicator** when recording
- **Animated recording feedback**

## Solution: Fyne GUI Framework

Fyne is:
- Pure Go (no system dependencies)
- Native look on Linux (GTK-based)
- Lightweight (~10-15MB binary increase)
- Easy to use

## Implementation Plan

### Phase 1: Add Fyne Dependency
```bash
go get fyne.io/fyne/v2@latest
```

### Phase 2: Create Native GUI App
- System tray icon (microphone)
- Small recording popup window with:
  - Animated sound waves
  - Pulsing circle
  - "Recording..." text
- Auto-hide after transcription
- Error notifications

### Phase 3: Replace Web UI
New structure:
```
cmd/voicetype-gui/
├── main.go          # Fyne GUI app
├── indicator.go     # Recording popup
└── tray.go          # System tray
```

### Phase 4: Build
```bash
make build-gui  # Builds with Fyne
```

## GUI States
```
[Idle]         → System tray icon only
[Recording]    → Small popup with animated waves
[Processing]   → Spinning indicator
[Complete]     → Brief checkmark, then close
[Error]        → Notification popup
```

## Files to Create/Modify
1. `cmd/voicetype-gui/main.go` - Fyne app
2. `cmd/voicetype-gui/indicator.go` - Recording popup
3. `cmd/voicetype-gui/tray.go` - System tray
4. Update `Makefile`

## Estimated Time
- 2-3 hours for complete implementation
- Binary size: ~15-20MB (with Fyne)

## Commands to Run
```bash
# Add Fyne dependency
go get fyne.io/fyne/v2@latest

# Create new GUI
go run cmd/voicetype-gui/main.go
```

---

**Ready to proceed? Say "NEXT" and I'll build the native GUI.**
