# 🎙️ SpeekToText Linux

**Linux Native • Ultra-Fast • Minimalist One-Shot Speech-to-Text**

SpeekToText Linux is a lightweight, Linux-native speech-to-text typing replacement. It allows you to speak and instantly type text into any active application (browser, terminal, editor, or chat) bypassing the clipboard.

![VoiceType Pill Design](https://github.com/mrprohack/speek_to_text_linux/raw/master/PLAN_NATIVE_GUI.md) *(UI inspired by modern voice assistants)*

## ✨ Features

- **One-Shot Mode**: Launches, records immediately, and quits automatically after typing. Zero background overhead.
- **Direct Typing**: Transcribes your voice and "types" it directly at your cursor. **Bypasses the clipboard** so your copied links/passwords are safe.
- **Ultra-Fast**: Perceived latency of <1 second using Groq's Whisper API.
- **Minimalist Pill UI**: A beautiful, floating pill that shows real-time wave animations and "Listening..." / "Typing..." status.
- **Smart Toggle**: Rerunning the app while it's recording acts as a "Stop" command.
- **Wayland Native**: Optimized for modern Linux distros (Ubuntu, Fedora) using `wtype` and `wl-copy`.

## 🚀 Getting Started

### Prerequisites

- **OS**: Linux (X11 or Wayland)
- **Dependencies**: `xdotool`, `xinput`, `wtype` (Wayland)

  ```bash
  # Debian/Ubuntu
  sudo apt install xdotool xinput wtype
  ```

### Installation

1. **Clone the repository**:

   ```bash
   git clone https://github.com/mrprohack/speek_to_text_linux.git
   cd speek_to_text_linux
   ```

2. **Setup your API Key**:
   Get a free API key from [Groq Console](https://console.groq.com/).

   ```bash
   export GROQ_API_KEY="your_api_key_here"
   ```

3. **Build**:

   ```bash
   make build-gui
   ```

## ⌨️ Usage

### The "One-Shot" Flow

1. **Launch**: Run `./VoiceType-gui`. It starts recording **immediately**.
2. **Stop**: Press `Ctrl + Space` (if configured) or simply **launch the app again**.
3. **Typing**: The app transcribes and types the text at your cursor.
4. **Auto-Exit**: The app closes itself 1 second after finishing.

## 🎛️ Customizing Keymap (Wayland Recommended)

For a 100% reliable experience on modern Linux (Wayland):

1. Go to **Settings** -> **Keyboard** -> **Custom Shortcuts**.
2. Add a new shortcut:
   - **Name**: `SpeekToText`
   - **Command**: `/path/to/your/VoiceType-gui --toggle`
   - **Shortcut**: `Ctrl + Space`
3. Now, pressing `Ctrl + Space` once starts recording, and pressing it again stops and types!

## 🛠️ Build Commands

```bash
make build-gui   # Build the modern GUI version
make build       # Build the CLI version
make clean       # Remove build artifacts
```

## 🛡️ License

Distributed under the MIT License. See `LICENSE` for more information.
