# Audio8 Offline Mobile UX Redesign

## Goal
Make Audio8 Offline feel like a simple consumer text-to-speech app on Android: type text, choose a voice, generate, listen, and share without exposing model internals in the primary workflow.

## Product structure
The app has three top-level destinations: **Speak**, **Voices**, and **Settings**.

### Speak
- Opens first and owns the primary workflow.
- Header identifies Audio8 Offline and shows a compact offline/setup status.
- Voice selection is a compact card; tapping it opens a bottom sheet.
- The text composer is the visual focus and supports long text, clear, paste-friendly keyboard behavior, and a character count.
- A large reachable primary action generates speech; while busy it changes to a cancel action and exposes progress.
- Generated audio appears as a result card with play/pause affordance, status, replay/play, and share actions supported by the current native bridge.
- First launch embeds model setup on this page rather than sending users to a separate model tab.

### Voices
- Saved voices are presented as simple local voice cards.
- Creating a voice becomes a guided mobile flow: permission/consent -> record -> enter exact transcript and name -> save.
- Recording state is visually obvious with a prominent microphone/stop control and concise guidance.
- If the registration pack is missing, the screen explains why it is required and offers a direct download action.
- Selecting a saved voice immediately makes it the active voice and returns to Speak.
- Delete is a secondary destructive action and requires confirmation.

### Settings
Settings replaces the old Model and About tabs. It contains:
- Model & storage state for Audio8 core and the optional registration pack.
- Download/progress/error states using plain-language copy.
- Model removal controls separated from normal usage.
- Privacy/offline information.
- Supported language information.
- Model fingerprint and app details.

## Visual system
- Material 3 foundation with a dark-first scheme consistent with the existing app.
- One accent color for primary actions/progress; status never relies on color alone.
- 16-20 dp horizontal page margins, 12-18 dp vertical rhythm, and rounded 16-24 dp surfaces.
- Primary touch targets are at least 48 dp, with main actions approximately 56 dp high.
- Strong hierarchy: page title -> section title -> body/help text.
- Bottom sheets are used for small decisions such as voice selection; top-level navigation remains only three destinations.
- Safe-area and keyboard insets must keep the primary action reachable on narrow Android phones.

## Behavior and states
- Core model states: not installed, downloading, ready, error/retry.
- Voice registration pack states: not installed, downloading, ready, error/retry.
- Generation states: idle, generating/progress, ready, error/cancel.
- Errors should stay close to the action through inline status or a concise snackbar; the whole UI must not become unusable after a recoverable failure.
- Existing Audio8 native methods remain the source of truth. No cloud service, account, analytics, or new network backend is introduced.

## Architecture
Keep the native Audio8 engine and method-channel contract unchanged. Refactor the Flutter presentation into focused widgets/helpers so the redesign is easier to test and maintain, while keeping state coordination in the app-level state object for this small application.

## Accessibility and responsive requirements
- Test common Android widths from roughly 320 dp through 430 dp.
- No horizontal overflow at 320 dp.
- Text scaling should preserve control readability.
- Icons with non-obvious meaning get tooltips/labels.
- Destructive model/voice deletion requires confirmation.

## Testing and regression gates
- Widget test first-run state: setup card, three-destination navigation, and disabled generation until requirements are met.
- Widget test ready state: selected voice, enabled Generate action, player/result actions after synthesis state.
- Widget test Voices state: create-voice entry point and saved voice selection.
- Widget test Settings state: model/storage sections exist and technical About content has moved here.
- `flutter analyze` and all widget tests must pass in CI.
- Android release APK build must pass.
- Existing APK-level regression verifier must still confirm `com.mrpro.audio8offline.MainActivity` exists in `classes*.dex`.

## Out of scope
- Changing the Audio8 ONNX inference implementation.
- Adding cloud synthesis, accounts, analytics, or subscriptions.
- Building a full persistent generation history database.
- Replacing the current native audio playback/share implementation.
