# VoiceType GUI - Native Linux Speech-to-Text

## Installation

### 1. Install Dependencies

```bash
# Required for hotkey detection
sudo apt install xdotool xclip

# Optional: for notifications
sudo apt install libnotify-bin zenity
```

### 2. Set API Key

```bash
export GROQ_API_KEY="your_groq_api_key"
```

### 3. Run

```bash
./VoiceType-gui
```

## Usage

1. **Hold Ctrl+Space** to start recording
2. **Release** to stop and transcribe
3. Text is automatically pasted where your cursor is

## Features

- 🎙️ Animated recording indicator popup
- 🌊 Sound wave animation while recording
- ✅ Success popup with transcribed text
- ❌ Error notifications on failures
- 🖥️ System tray icon (when supported)

## Troubleshooting

### "Warning: No hotkey tool found"

Install xdotool:
```bash
sudo apt install xdotool
```

### Hotkey not working

1. Make sure you're in an X11 or Wayland session
2. Check if xdotool is installed: `which xdotool`
3. Try a different hotkey: `./VoiceType-gui -hotkey=F6`

### No audio detected

Check microphone:
```bash
arecord -l
```

Set device manually:
```bash
./VoiceType-gui -device=hw:0
```

## Files

- `VoiceType-gui` - GUI application with animations
- `VoiceType` - CLI version (no GUI)
