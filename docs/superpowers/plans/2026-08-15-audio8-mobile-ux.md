# Audio8 Mobile UX Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild the Android Flutter presentation as a consumer-first mobile TTS app with Speak, Voices, and Settings destinations while preserving the existing Audio8 native engine and startup-crash regression guard.

**Architecture:** Keep method-channel/native inference behavior unchanged and refactor only the Flutter presentation/state coordination. The root app owns bridge/status/action state; focused widgets render the three destinations and bottom sheets. Tests drive the required user-visible states before production UI changes.

**Tech Stack:** Flutter 3.44.8, Dart 3.12.x, Material 3, Android/Gradle/Kotlin native bridge already present, GitHub Actions Android build.

## Global Constraints
- Android-only application.
- Existing Audio8 native method-channel contract remains unchanged.
- No cloud synthesis, account system, analytics, or new backend.
- Three top-level destinations only: Speak, Voices, Settings.
- Existing APK DEX regression guard for `com.mrpro.audio8offline.MainActivity` must remain mandatory.
- Release package version becomes `0.2.0+3`.

---

### Task 1: Lock the new navigation and first-run UX with widget tests

**Files:**
- Modify: `test/smoke_test.dart`

**Interfaces:**
- Consumes: method channel `audio8_offline/methods`, especially `status`.
- Produces: test expectations for three-destination navigation, embedded model setup, and disabled generation before setup.

- [ ] **Step 1: Replace the first-launch test with expectations for Speak / Voices / Settings, `Download model`, `What should I say?`, and no Model/About navigation destinations.**
- [ ] **Step 2: Add a ready-state test expecting `Offline ready`, a compact selected-voice control, and enabled `Generate speech`.**
- [ ] **Step 3: Add Voices and Settings tests for `Create voice`, `Model & storage`, `Privacy`, and `About Audio8`.**
- [ ] **Step 4: Run tests in GitHub Actions before UI implementation and confirm they fail because the existing four-tab UI does not meet the new expectations.**

### Task 2: Implement the mobile-first app shell and visual system

**Files:**
- Modify: `lib/app.dart`

**Interfaces:**
- Consumes: current app state (`status`, `activity`, selected voice, busy/recording flags).
- Produces: Material 3 app shell with Speak / Voices / Settings `NavigationBar`, reusable page header, cards, section labels, and responsive padding.

- [ ] **Step 1: Replace the four-destination shell with three destinations.**
- [ ] **Step 2: Move Model and About information into Settings.**
- [ ] **Step 3: Add consistent rounded surfaces, 48-56 dp touch targets, safe-area spacing, and narrow-width-safe layouts.**
- [ ] **Step 4: Run widget tests and fix only shell-related failures.**

### Task 3: Rebuild Speak around type -> voice -> generate -> listen/share

**Files:**
- Modify: `lib/app.dart`

**Interfaces:**
- Consumes: `bridge.downloadCore`, `bridge.synthesize`, `bridge.cancel`, `bridge.play`, `bridge.stop`, `bridge.share` and status/activity data.
- Produces: embedded setup card, bottom-sheet voice picker, large composer, generation progress/cancel state, and generated-audio result card.

- [ ] **Step 1: Build a first-run setup surface with one prominent `Download model` action and plain-language offline/storage copy.**
- [ ] **Step 2: Replace the dropdown with a tappable voice card that opens a modal bottom sheet.**
- [ ] **Step 3: Make the composer the main visual surface with a 5000-character limit and clear action.**
- [ ] **Step 4: Keep Generate as the strongest full-width action and swap it to Cancel while busy.**
- [ ] **Step 5: Show Play/Stop/Share in a generated-audio result card only when audio exists.**
- [ ] **Step 6: Run ready-state and first-run widget tests.**

### Task 4: Turn voice creation into a guided mobile flow

**Files:**
- Modify: `lib/app.dart`
- Modify: `test/smoke_test.dart`

**Interfaces:**
- Consumes: `bridge.downloadRegistration`, `bridge.startRecording`, `bridge.register`, `bridge.deleteVoice`.
- Produces: voice list cards, `Create voice` modal flow, consent, recording state, transcript/name entry, delete confirmation, and selection behavior.

- [ ] **Step 1: Add a failing widget test for the `Create voice` entry point and registration-pack setup state.**
- [ ] **Step 2: Implement saved voice cards with selected-state feedback and a protected delete action.**
- [ ] **Step 3: Implement a modal creation flow that clearly separates consent/instructions, recording, and save fields while using the existing native calls.**
- [ ] **Step 4: Run Voices tests and existing smoke tests.**

### Task 5: Build Settings model/storage and About sections

**Files:**
- Modify: `lib/app.dart`
- Modify: `test/smoke_test.dart`

**Interfaces:**
- Consumes: model install booleans, model fingerprint, download/remove methods, activity progress.
- Produces: `Model & storage`, `Privacy`, supported languages, `About Audio8`, and destructive removal confirmations.

- [ ] **Step 1: Add a failing test that Settings contains model and About content previously split across tabs.**
- [ ] **Step 2: Implement readable install/download/remove model cards and inline progress.**
- [ ] **Step 3: Implement privacy/offline and About sections with the fingerprint selectable.**
- [ ] **Step 4: Run Settings tests.**

### Task 6: Version, analyze, build, and verify the APK

**Files:**
- Modify: `pubspec.yaml`
- Modify: `.github/workflows/android.yml` if packaging labels need updating.
- Keep: `tools/verify_apk.py`

**Interfaces:**
- Produces: release APK and source ZIP for version `0.2.0+3`.

- [ ] **Step 1: Set version to `0.2.0+3`.**
- [ ] **Step 2: Run `flutter analyze` and all widget tests in GitHub Actions.**
- [ ] **Step 3: Run Kotlin/native verification and the release APK build.**
- [ ] **Step 4: Run `tools/verify_apk.py` against the built APK and require PASS.**
- [ ] **Step 5: Download the CI artifact and independently verify SHA-256, non-zero APK size, and `MainActivity` presence in DEX before delivery.**
