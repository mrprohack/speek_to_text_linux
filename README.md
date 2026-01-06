# 🎙️ VoiceType
**Linux Native • Ultra-Fast • Minimalist Speech-to-Text**

VoiceType is a lightweight, Linux-native speech-to-text typing replacement. It allows you to speak and instantly paste text into any active application (browser, terminal, editor, or chat) using a global hotkey.

![VoiceType Pill Design](https://github.com/mrprohack/sst_linux/raw/master/PLAN_NATIVE_GUI.md) *(UI inspired by modern voice assistants)*

## ✨ Features
- **Ultra-Fast**: Perceived latency of <1 second using Groq's Whisper API.
- **Minimalist Pill UI**: A beautiful, borderless, floating pill that shows real-time wave animations.
- **Global Hotkey**: Press `Ctrl + Space` to start speaking, and again to stop and paste.
- **Privacy First**: Zero local storage. Your audio is processed in RAM and deleted immediately after transcription.
- **Smart Filtering**: Automatically ignores silence or empty vocalizations to keep your workspace clean.
- **Always on Top**: The UI stays floating above all windows for quick access.

## 🚀 Getting Started

### Prerequisites
- **OS**: Linux (X11 or Wayland)
- **Dependencies**: `xclip`, `xdotool`, `xinput`, `wmctrl`
  ```bash
  sudo apt install xclip xdotool xinput wmctrl
  ```

### Installation
1. **Clone the repository**:
   ```bash
   git clone https://github.com/mrprohack/sst_linux.git
   cd sst_linux
   ```
2. **Setup your API Key**:
   Get a free API key from [Groq Console](https://console.groq.com/).
   ```bash
   export GROQ_API_KEY="your_api_key_here"
   ```
3. **Build & Run**:
   ```bash
   make build-gui
   ./VoiceType-gui
   ```

## ⌨️ Usage
- **Start/Stop Recording**: Press `Ctrl + Space`
- **Transcription**: The app will automatically transcribe your voice and paste it at your current cursor position.
- **Feedback**:
  - **Pink Border**: Listening...
  - **Green Border**: Successfully pasted!
  - **Red Border**: Error (check network or API key).

## 🛠️ Build Commands
```bash
make build-gui   # Build the modern GUI version
make build       # Build the CLI version
make clean       # Remove build artifacts
```

## 🛡️ License
Distributed under the MIT License. See `LICENSE` for more information.
