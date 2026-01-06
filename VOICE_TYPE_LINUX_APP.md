# 🎙️ VoiceType – Linux Native Speech-to-Text Typing App

## 1. App Overview
VoiceType is a **Linux-native, ultra-fast speech-to-text typing replacement app** designed for **normal users on old laptops**.  
It runs silently in the background and allows users to **hold a global keyboard shortcut, speak, and instantly paste text** into any active application.

The app is **open source**, **privacy-first**, and optimized for **<1 second perceived latency** using the **Groq Whisper API**.

---

## 2. Core Goals
- Replace keyboard typing with voice
- Work **anywhere** (browser, editor, terminal, chat apps)
- Extremely fast response
- Minimal RAM & CPU usage
- Simple, professional UI
- No audio or text storage

---

## 3. Target Users
- Normal (non-technical) Linux users
- Users with **old laptops (4GB RAM)**
- Users who want quick voice typing
- Accessibility users
- Developers (secondary)

---

## 4. Supported Platforms
- Linux (V1 only)
- X11 and Wayland
- Desktop environments: GNOME, KDE, XFCE

---

## 5. Key Features (Version 1)

### 5.1 Input
- Global keyboard shortcut (system-wide)
- **Push-to-talk**:
  - Hold key → record
  - Release key → stop & transcribe
- Microphone input only
- No system audio capture

---

### 5.2 Audio Handling
- Record audio directly into memory (RAM)
- Temporary audio buffer only
- Auto-delete audio immediately after transcription
- No audio saved to disk permanently

---

### 5.3 Speech-to-Text Engine
- API: Groq OpenAI-compatible endpoint
- Model: `whisper-large-v3`
- Language: English only
- Temperature: 0
- Response format: `verbose_json`

#### API Example:
```bash
curl https://api.groq.com/openai/v1/audio/transcriptions \
  -H "Authorization: Bearer $GROQ_API_KEY" \
  -F "model=whisper-large-v3" \
  -F "file=@audio.wav" \
  -F "temperature=0" \
  -F "response_format=verbose_json"
````

---

### 5.4 Output Behavior

* Extract plain text from `verbose_json`
* Automatically paste text into:

  * Active application
  * Current cursor position
* No confirmation popup
* No editing step (instant typing)

---

## 6. UI / UX Design

### 6.1 UI Style

* Minimal GUI
* Professional, clean look
* No heavy animations
* No distracting visuals

---

### 6.2 Recording Feedback

* Small on-screen indicator while recording:

  * Subtle animated circle or waveform
  * Neutral colors (blue/white)
* No sound effects

---

### 6.3 Error Handling

* Show **small popup notification** on:

  * Network failure
  * API error
  * Missing API key
* Errors must be human-readable
* App must recover gracefully

---

## 7. Performance Requirements

* Must run smoothly on old laptops
* Low memory footprint
* Minimal background CPU usage
* Fast startup time
* Perceived transcription time: **< 1 second**

---

## 8. App Architecture

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

---

## 9. Technology Stack

### Core Language

* **Go (Golang)**

### System Integration

* Audio: ALSA / PulseAudio
* Clipboard access
* Auto-paste simulation
* Global hotkey listener

---

## 10. App Type

* Background Linux app
* No main window
* UI appears only when:

  * Recording
  * Error occurs

---

## 11. Packaging & Distribution

* Single static binary
* No runtime dependencies
* Optional future packaging:

  * AppImage
  * `.deb`

---

## 12. Security & Privacy

### API Key

* Provided via environment variable only:

  ```bash
  export GROQ_API_KEY=your_key_here
  ```
* No config file storing keys

### Privacy

* Never store audio
* Never store transcriptions
* No telemetry
* No analytics
* No logging of user speech

---

## 13. Open Source Requirements

* Public GitHub repository
* Clear README
* Simple build instructions
* MIT or Apache-2.0 license
* Clean code structure

---

## 14. Version 1 Scope (Strict)

* Linux only
* English only
* Online only
* Push-to-talk only
* Auto-paste only
* No settings UI
* No history
* No offline mode

---

## 15. Future (Out of Scope for V1)

* Tamil + English
* Offline model
* Custom shortcuts UI
* Text preview editor
* Per-word live typing
* System audio transcription

---

## 16. Success Criteria

* User can speak and see text pasted instantly
* App feels faster than typing
* Works reliably on old hardware
* No privacy concerns
* Simple and stable

---

## 17. Project Name Ideas

* VoiceType
* SpeakPaste
* VoxKey
* Talk2Type
* QuickDict

---

## 18. One-Line Description

> “Hold a key, speak, and your words appear instantly—no typing, no waiting.”

---

## ✅ What you can do next
- Use this as:
  - AI **app-builder prompt**
  - **GitHub README**
  - **Developer spec**
  - **Open-source project foundation**

If you want, next I can:
- Generate **Go project folder structure**
- Write **initial Go code**
- Create **README.md**
- Design **hotkey + audio code**
- Optimize for **<1s latency**

Just say **NEXT** 🔥

