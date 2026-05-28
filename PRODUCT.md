# Wispr Vibe — Product Context

## Product Purpose
Desktop speech-to-text tool for developers. Press a hotkey, speak, text appears at your cursor. Zero workflow interruption.

## Users
- Software developers during "vibe coding" sessions
- Power users who never want to leave their keyboard/IDE
- Portuguese and English speakers
- Windows 10/11 primary platform

## Brand & Tone
- **Register**: product (design serves the product)
- **Personality**: Fast, invisible, confident. Like a great tool that disappears when you use it.
- **Tone**: Developer-friendly, minimal, no-nonsense. Not corporate, not playful. Calm and competent.
- **Anti-references**: Electron bloatware, flashy gamer aesthetics, corporate enterprise dashboards

## Design Principles
1. **Invisible by default** — the app stays out of your way. System tray, no focus stealing.
2. **Glanceable** — recording state must be obvious in <100ms
3. **Minimal chrome** — every pixel earns its place
4. **Dark-native** — developers work in dark environments. Respect that.
5. **Fast feedback** — state changes feel instant

## Key Interactions
1. Hotkey press → recording starts → toast overlay appears
2. Hotkey press again → transcribing → text appears at cursor
3. GUI is secondary; most interaction is hotkey-driven from other apps
4. Settings changes are infrequent

## Technical Constraints
- Fyne v2.7.4 for GUI (limited theming compared to web)
- Native Win32 for toast overlay (direct GDI painting)
- Must not steal focus from active application
- Must work with any app (IDE, browser, terminal)
