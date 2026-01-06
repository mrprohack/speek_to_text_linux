# 🎙️ VoiceType – Linux Speech-to-Text App

## Todo List

> Hold a key, speak, and your words appear instantly—no typing, no waiting.

---

## 🚀 Phase 1: Project Setup

- [ ] **Initialize Go project**
  - [ ] Create main module with `go mod init VoiceType`
  - [ ] Set up project structure (cmd/, internal/, pkg/)
  - [ ] Add initial dependencies (alsa, pulseaudio, x11, wayland)
  - [ ] Configure build tags for X11/Wayland support

- [ ] **Configure build system**
  - [ ] Create Makefile with build targets
  - [ ] Set up GoReleaser config (future AppImage/.deb)
  - [ ] Configure CGO for audio library bindings
  - [ ] Add goreleaser.yml for CI/CD

**Priority: High**

---

## 🎯 Phase 2: Core Infrastructure

### 2.1 Global Hotkey Listener
- [ ] **Implement X11 hotkey listener**
  - [ ] Use `github.com/BurntSushi/xgb` for X11 grab
  - [ ] Capture key events system-wide
  - [ ] Support configurable hotkey (default: F6 or similar)
  - [ ] Add Wayland support via `wlroots`
  - [ ] Create hotkey manager abstraction layer

- [ ] **Implement key event handlers**
  - [ ] KeyDown: Start audio recording
  - [ ] KeyUp: Stop recording and trigger transcription
  - [ ] Add debounce/throttle handling
  - [ ] Implement state machine (idle → recording → processing)

**Priority: High**

### 2.2 Audio Capture System
- [ ] **Implement ALSA audio capture**
  - [ ] Open default microphone device
  - [ ] Configure sample rate (16kHz for Whisper)
  - [ ] Set up ring buffer for real-time recording
  - [ ] Handle device enumeration and selection

- [ ] **Implement PulseAudio support**
  - [ ] Create PulseAudio monitor capture
  - [ ] Fallback to ALSA if Pulse unavailable
  - [ ] Auto-detect best audio source

- [ ] **Memory-optimized audio buffer**
  - [ ] Allocate fixed-size circular buffer in RAM
  - [ ] Auto-resize based on recording duration
  - [ ] Zero-copy audio data handling
  - [ ] Immediate cleanup after transcription

**Priority: High**

### 2.3 Audio Encoding
- [ ] **WAV file generation**
  - [ ] Convert raw audio to WAV format in memory
  - [ ] Write WAV headers dynamically
  - [ ] Optimize for minimal memory footprint
  - [ ] Use streaming for large recordings

**Priority: Medium**

---

## 🗣️ Phase 3: Speech-to-Text Engine

### 3.1 Groq API Integration
- [ ] **Implement API client**
  - [ ] Create HTTP client with connection pooling
  - [ ] Handle authentication via GROQ_API_KEY env var
  - [ ] Implement retry logic (3 retries with backoff)
  - [ ] Set timeout to 5 seconds max

- [ ] **API request handling**
  - [ ] POST multipart form with audio file
  - [ ] Model: whisper-large-v3
  - [ ] Temperature: 0
  - [ ] Response format: verbose_json

- [ ] **Response parsing**
  - [ ] Extract text from verbose_json response
  - [ ] Handle API errors gracefully
  - [ ] Parse error codes and messages

**Priority: High**

### 3.2 Error Handling
- [ ] **Network error handling**
  - [ ] Detect connection failures
  - [ ] Show notification on failure
  - [ ] Auto-retry once on transient errors

- [ ] **API error handling**
  - [ ] Handle 401 (invalid API key)
  - [ ] Handle 429 (rate limit) with backoff
  - [ ] Handle 500+ server errors
  - [ ] Show human-readable error messages

**Priority: High**

---

## 📋 Phase 4: Output & Input

### 4.1 Text Processing
- [ ] **Text extraction**
  - [ ] Parse JSON response from Whisper API
  - [ ] Extract clean text content
  - [ ] Remove timestamps and metadata
  - [ ] Basic text sanitization (trim whitespace)

**Priority: High**

### 4.2 Clipboard & Auto-Paste
- [ ] **Clipboard operations**
  - [ ] Copy text to system clipboard
  - [ ] Support X11 and Wayland clipboards
  - [ ] Handle Unicode/text encoding

- [ ] **Auto-paste simulation**
  - [ ] Simulate Ctrl+V key sequence
  - [ ] X11: Use XTest extension
  - [ ] Wayland: Use wl_keyboard API
  - [ ] Fallback: xdotool for compatibility

- [ ] **Timing optimization**
  - [ ] Paste immediately after clipboard set
  - [ ] Minimize delay between API response and paste
  - [ ] Target: <100ms from text ready to pasted

**Priority: High**

---

## 🎨 Phase 5: UI/UX

### 5.1 Recording Indicator
- [ ] **Visual feedback system**
  - [ ] Create minimal overlay window
  - [ ] Animated recording circle/waveform
  - [ ] Use neutral colors (blue/white)
  - [ ] Position: corner of screen

- [ ] **Window management**
  - [ ] X11: Override redirect window
  - [ ] Wayland: layer-shell protocol
  - [ ] Always on top, no focus
  - [ ] Click-through enabled

- [ ] **Animation system**
  - [ ] Smooth pulsing animation
  - [ ] Low CPU usage (requestAnimationFrame)
  - [ ] Stop immediately on recording end

**Priority: High**

### 5.2 Error Notifications
- [ ] **Notification system**
  - [ ] Use libnotify (via D-Bus)
  - [ ] Show toast notifications
  - [ ] Auto-dismiss after 3 seconds
  - [ ] Critical errors persist until clicked

- [ ] **Error types and messages**
  - [ ] Network failure: "Check your connection"
  - [ ] API error: "Transcription failed"
  - [ ] Missing API key: "Set GROQ_API_KEY environment variable"
  - [ ] Microphone access denied: "Check permissions"

**Priority: Medium**

### 5.3 System Tray (Optional)
- [ ] **Status indicator**
  - [ ] Show app is running
  - [ ] Right-click menu (Quit only)
  - [ ] Minimize to tray on close

**Priority: Low**

---

## ⚡ Phase 6: Performance Optimization

### 6.1 Latency Optimization
- [ ] **API latency reduction**
  - [ ] Pre-establish HTTP/2 connection
  - [ ] Parallelize clipboard set + paste
  - [ ] Zero-copy audio buffer processing

- [ ] **Startup performance**
  - [ ] Lazy-load audio drivers
  - [ ] Pre-init hotkey system on app start
  - [ ] Target: <200ms startup time

- [ ] **Memory optimization**
  - [ ] Reuse audio buffer between recordings
  - [ ] Stream audio directly to API (no temp file)
  - [ ] Target: <50MB RAM usage

**Priority: High**

### 6.2 CPU Optimization
- [ ] **Background processing**
  - [ ] Non-blocking API calls (goroutines)
  - [ ] Audio capture in separate goroutine
  - [ ] Minimize main thread work

- [ ] **Animation efficiency**
  - [ ] Use GPU rendering
  - [ ] Throttle frame rate to 30fps
  - [ ] Stop animation loop when hidden

**Priority: Medium**

---

## 🔒 Phase 7: Security & Privacy

### 7.1 API Key Management
- [ ] **Environment variable handling**
  - [ ] Read GROQ_API_KEY from env only
  - [ ] Never write to config file
  - [ ] Clear from memory after use
  - [ ] Warn if not set on startup

- [ ] **Runtime security**
  - [ ] No logging of API key access
  - [ ] Memory locking to prevent swapping
  - [ ] Secure cleanup on exit

**Priority: High**

### 7.2 Privacy Implementation
- [ ] **Audio data protection**
  - [ ] RAM-only audio storage
  - [ ] Zero disk writes for audio
  - [ ] Immediate deletion after transcription
  - [ ] No caching of audio data

- [ ] **Text privacy**
  - [ ] No storage of transcriptions
  - [ ] No clipboard history
  - [ ] No telemetry or analytics
  - [ ] No network calls except API

**Priority: High**

---

## 🧪 Phase 8: Testing

### 8.1 Unit Testing
- [ ] **Core components**
  - [ ] Hotkey listener tests
  - [ ] Audio buffer tests
  - [ ] WAV encoding tests
  - [ ] Text extraction tests

- [ ] **API integration**
  - [ ] Mock API responses
  - [ ] Error handling tests
  - [ ] Timeout handling tests

**Priority: Medium**

### 8.2 Integration Testing
- [ ] **End-to-end tests**
  - [ ] Full recording → API → paste flow
  - [ ] Hotkey trigger verification
  - [ ] Error recovery testing

- [ ] **Compatibility testing**
  - [ ] X11 and Wayland environments
  - [ ] GNOME, KDE, XFCE
  - [ ] ALSA and PulseAudio

**Priority: Medium**

### 8.3 Performance Testing
- [ ] **Latency benchmarks**
  - [ ] Measure hotkey-to-recording-start time
  - [ ] Measure recording-end-to-API-call time
  - [ ] Measure API-response-to-paste time
  - [ ] Target: <1s total perceived latency

- [ ] **Resource usage**
  - [ ] Memory footprint tests
  - [ ] CPU usage during idle
  - [ ] CPU usage during recording

**Priority: Medium**

---

## 📚 Phase 9: Documentation

### 9.1 README
- [ ] **Project overview**
  - [ ] One-line description
  - [ ] Core features list
  - [ ] Target users and use cases

- [ ] **Installation guide**
  - [ ] Prerequisites (Go, libraries)
  - [ ] Build instructions
  - [ ] Running the app
  - [ ] Environment setup (API key)

- [ ] **Usage guide**
  - [ ] Hotkey configuration
  - [ ] Recording tips
  - [ ] Troubleshooting

**Priority: High**

### 9.2 Technical Documentation
- [ ] **Architecture docs**
  - [ ] System diagram
  - [ ] Component descriptions
  - [ ] Data flow explanation

- [ ] **API documentation**
  - [ ] Groq API usage
  - [ ] Response format
  - [ ] Error codes

- [ ] **Development guide**
  - [ ] Code style guidelines
  - [ ] Testing instructions
  - [ ] Contributing guidelines

**Priority: Medium**

### 9.3 Security Documentation
- [ ] **Privacy policy**
  - [ ] Data handling explanation
  - [ ] API key security
  - [ ] No telemetry statement

**Priority: Medium**

---

## 📦 Phase 10: Packaging & Distribution

### 10.1 Binary Build
- [ ] **Static binary compilation**
  - [ ] Build with CGO disabled where possible
  - [ ] Include all dependencies
  - [ ] Test on multiple distributions

- [ ] **Multi-platform support**
  - [ ] Build for amd64
  - [ ] Build for arm64
  - [ ] CI/CD pipeline setup

**Priority: High**

### 10.2 Package Creation
- [ ] **AppImage creation**
  - [ ] Configure AppImage build
  - [ ] Include runtime dependencies
  - [ ] Test on multiple distros

- [ ] **Debian package (.deb)**
  - [ ] Create debian/ directory
  - [ ] Set up dependencies
  - [ ] Post-install scripts

**Priority: Low (Future)**

---

## ✅ Phase 11: Finalization

- [ ] **Code review**
  - [ ] Self-review all changes
  - [ ] Check for security issues
  - [ ] Verify performance requirements

- [ ] **Final testing**
  - [ ] Manual testing on target systems
  - [ ] Test on old hardware (4GB RAM)
  - [ ] Stress testing

- [ ] **Release preparation**
  - [ ] Version numbering (v1.0.0)
  - [ ] Create release notes
  - [ ] Tag release in git

**Priority: High**

---

## 📋 Quick Reference

### Current Phase
**Phase 1: Project Setup**

### Completed Tasks
_None yet_

### Next Tasks
1. Initialize Go project structure
2. Set up build system with Makefile
3. Configure CGO for audio libraries

### Dependencies
- Go 1.21+
- ALSA development libraries (`libasound2-dev`)
- PulseAudio development libraries (`libpulse-dev`)
- X11 development libraries (`libx11-dev`, `libxtst-dev`)
- Wayland development libraries (`libwlroots-dev`, `wayland-protocols`)

---

## 🎯 Success Criteria

- [ ] User can speak and see text pasted instantly
- [ ] App feels faster than typing
- [ ] Works reliably on old hardware (4GB RAM)
- [ ] No privacy concerns (no data storage)
- [ ] Simple and stable operation
- [ ] Perceived latency: <1 second

---

*Last updated: January 2026*
*Version: 1.0.0 (Planning)*
