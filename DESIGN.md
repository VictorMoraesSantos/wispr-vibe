# Wispr Vibe — Design System

## Color Strategy
**Restrained**: warm-tinted neutrals + one violet accent at ≤10% surface area.

## Palette

### Neutrals (warm-tinted toward violet)
| Token | RGB | Usage |
|-------|-----|-------|
| bg-base | 20, 20, 25 | Window background |
| bg-surface | 26, 26, 32 | Raised panels, inputs |
| bg-elevated | 34, 34, 40 | Buttons, cards |
| border-subtle | 38, 36, 44 | Separators |
| border-default | 50, 48, 60 | Input borders |
| text-primary | 222, 220, 228 | Body text |
| text-secondary | 140, 138, 152 | Hints, captions |
| text-disabled | 82, 80, 92 | Disabled state |

### Semantic Colors
| Token | RGB | Usage |
|-------|-----|-------|
| accent | 129, 110, 240 | Primary action, processing state |
| success | 74, 210, 135 | Ready, done states |
| error | 240, 75, 75 | Recording, errors |
| focus | 129, 110, 240 @ 70A | Focus rings |

## Typography
- System font stack (Segoe UI on Windows)
- Body: 13px, regular weight
- Heading: 17px, bold
- Sub-heading: 14px, bold
- Monospace for timer, hotkey badge

## Spacing
- Inner padding: 12px
- Outer padding: 6px
- Section gaps: 8-12px (varied for rhythm)

## Toast Overlay (Win32 native)
- Dimensions: 240 x 40px
- Corner radius: 20px (pill shape)
- Background: 18, 18, 23
- Border: 44, 42, 54
- Text: Segoe UI, 13px, weight 500
- State indicator: 8px circle (red = recording, violet = processing)
- Opacity: 235/255
- Position: bottom-center, 60px from screen edge

## States
The app communicates via a colored dot:
- **Green** (mint): ready, success
- **Red** (warm): recording, error
- **Violet**: processing/transcribing

## Principles
1. No emoji in UI text (use color and shape for state)
2. Monospace for data values (timer, hotkey)
3. Low-importance styling for secondary information
4. Form widgets for label-aligned settings
5. Consistent warm-violet tint in all neutrals
