# 🎙️ VoiceType

> Hold a key, speak, and your words appear instantly—no typing, no waiting.

VoiceType is a **Linux-native, ultra-fast speech-to-text typing replacement app** designed for **normal users on old laptops**. It runs silently in the background and allows users to **hold a global keyboard shortcut, speak, and instantly paste text** into any active application.

## ✨ Features

- **Global Hotkey**: Press and hold F6 (or custom key) to record
- **Instant Transcription**: Uses Groq's Whisper API for <1 second latency
- **Privacy-First**: No audio or text storage, everything happens in memory
- **Cross-Desktop**: Works on X11 and Wayland (GNOME, KDE, XFCE)
- **Low Resource**: Optimized for old laptops with 4GB RAM
- **Auto-Paste**: Automatically pastes transcribed text where your cursor is

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- Linux (X11 or Wayland)
- Microphone access
- Groq API key ([Get one here](https://console.groq.com/))

### Installation

1. **Clone and build:**
```bash
git clone https://github.com/yourusername/VoiceType.git
cd VoiceType
make build
```

2. **Set up your API key:**
```bash
export GROQ_API_KEY="your_api_key_here"
```

3. **Run:**
```bash
./VoiceType
```

### From Source

```bash
go build -o VoiceType ./cmd/voicetype/
```

## 📖 Usage

1. **Start VoiceType:**
   ```bash
   ./VoiceType
   ```

2. **Hold the hotkey** (default: F6) to start recording

3. **Release** to stop recording and transcribe

4. **Your text appears** instantly where your cursor is focused

### Command Line Options

```bash
./VoiceType --help
Usage: VoiceType [options]

Options:
  -hotkey string    Hotkey to trigger recording (default: "F6")
  -device string    Audio device to use (auto-detect if empty)
  -no-notify        Disable notifications
  -v                Enable verbose logging
  -version          Show version information
  -help             Show help information
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GROQ_API_KEY` | Your Groq API key | Required |
| `VOICE_TYPE_HOTKEY` | Custom hotkey | F6 |
| `VOICE_TYPE_AUDIO_DEVICE` | Audio device | Auto-detect |
| `VOICE_TYPE_MODEL` | Whisper model | whisper-large-v3 |
| `VOICE_TYPE_NOTIFICATIONS` | Enable notifications | 1 |
| `VOICE_TYPE_VERBOSE` | Enable verbose logging | 0 |

## 🏗️ Architecture

```
[Global Hotkey Listener]
        ↓
[Audio Capture (RAM)]
        ↓
[Temporary Audio Buffer]
        ↓
[Groq Whisper API]
        ↓
[Text Extraction]
        ↓
[Clipboard + Auto Paste]
        ↓
[Cleanup & Delete Audio]
```

## 📁 Project Structure

```
VoiceType/
├── cmd/voicetype/          # Main application entry point
├── internal/
│   ├── api/               # Groq API client
│   ├── audio/             # Audio capture (ALSA/PulseAudio)
│   ├── clipboard/         # Clipboard and paste operations
│   ├── hotkey/            # Global hotkey listener
│   ├── notify/            # Notification system
│   └── ui/                # Recording indicator UI
├── pkg/
│   ├── config/            # Configuration management
│   ├── errors/            # Error handling utilities
│   └── wav/               # WAV file encoding
└── Makefile               # Build automation
```

## 🧪 Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test ./pkg/wav/ -v
```

## 🔧 Building

```bash
# Build the binary
make build

# Build with debug symbols
make build-debug

# Clean build artifacts
make clean

# Create release binary
make release
```

## 🖥️ Desktop Integration

### Autostart (Optional)

Add to your desktop environment's autostart:
```bash
cp VoiceType.desktop ~/.config/autostart/
```

## 🔒 Security & Privacy

- **API Key**: Only read from environment variable, never stored
- **Audio**: Recorded directly to RAM, never saved to disk
- **Text**: Transcriptions not stored, no clipboard history
- **No Telemetry**: Zero analytics or usage tracking
- **Memory**: Audio data deleted immediately after transcription

## 📋 Requirements

- **Hardware**: Microphone, 4GB RAM (optimized)
- **Software**: Linux (kernel 5.0+), Go 1.21+
- **Audio**: ALSA or PulseAudio
- **Display**: X11 or Wayland

### Recommended Tools

For full functionality, install these optional tools:
```bash
# Clipboard and paste
sudo apt-get install xclip xsel xdotool

# Notifications
sudo apt-get install libnotify-bin

# Recording indicator
sudo apt-get install yad
```

## 🐛 Troubleshooting

### No sound detected
- Check microphone permissions
- Verify audio device: `arecord -l`
- Set device manually: `./VoiceType -device hw:0`

### Paste not working
- Install xdotool: `sudo apt-get install xdotool`
- Check display environment: `echo $DISPLAY`

### API errors
- Verify API key: `echo $GROQ_API_KEY`
- Check network connection
- Check Groq API status

### Performance issues
- Run with verbose: `./VoiceType -v`
- Close unnecessary applications
- Use SSD for swap if available

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit changes: `git commit -m 'Add amazing feature'`
4. Push to branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Groq](https://groq.com/) for the fast Whisper API
- [Whisper](https://openai.com/research/whisper) by OpenAI
- [Go](https://go.dev/) programming language
- Linux community for audio and display technologies

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/yourusername/VoiceType/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/VoiceType/discussions)

---

**Made with ❤️ for Linux users everywhere**
