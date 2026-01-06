# VoiceType Roadmap: UX/UI & Linux Integration

This document outlines the planned improvements for VoiceType to achieve a world-class, seamless speech-to-typing experience on Linux.

## 🎨 Phase 1: UX & Visual Refinement

- [ ] **Window Draggability**: Allow the pill to be moved by dragging it anywhere on the screen.
- [ ] **Fade Animations**: Add smooth fade-in/out transitions when the pill appears or disappears.
- [ ] **Custom Themes**: Create a "Glass" theme with background blur (using compositor features if available).
- [ ] **Position Memory**: Remember the last coordinates of the pill window across restarts.
- [ ] **Adaptive Size**: The pill slightly expands when recording and contracts when transcribing.

## ⌨️ Phase 2: Linux System Integration

- [ ] **Configurable Keymap**: Create a simple `.voicetype.conf` (TOML or JSON) to let users change the hotkey (e.g., `Super + T`).
- [ ] **Push-to-Talk Mode**: Add a mode where recording only happens while the hotkey is held down.
- [ ] **System Tray (AppIndicator)**: A small icon in the top bar to show status, quit the app, or open settings.
- [ ] **Wayland Native Support**: Research and implement `wlr-layer-shell` or `xdg-shell` to ensure transparency and "always-on-top" works perfectly on Wayland compositors (GNOME/Hyprland).

## ⚡ Phase 3: Performance & Features

- [ ] **Local Model Option**: Support for local Faster-Whisper for offline or ultra-private transcription.
- [ ] **Multi-Language Selector**: A small UI toggle to switch between Groq models or target languages.
- [ ] **Sound Effects**: Subtle "pop" or "ping" sounds when recording starts/stops (optional/toggleable).

## ✅ Completed Roadmap

- [x] Initial Fyne-based GUI
- [x] Pill-shaped minimalist design
- [x] Real-time waveform animation
- [x] Global Ctrl+Space hotkey (X11/Wayland bridge)
- [x] Automatic text pasting
