# Pomodoro Timer - Single Process Specification

## Architecture

Single Go binary. No daemon, no CLI, no socket IPC. One process creates systray icon and manages all timer logic internally.

## Core Flow

```
User runs: ./pomodoro
→ Single process starts
→ Creates systray icon
→ Runs timer loop in goroutine
→ User controls via systray menu only
→ Process exits when user clicks Quit
```

## Timer Cycle

**3 work sessions per cycle:**
- Work (25min) → Short Break (5min) → Work (25min) → Short Break (5min) → Work (25min) → Long Break (20min) → [repeat]

**Phase transitions:**
- Work completes → Break starts immediately (notification + sound)
- Break completes → Enter WAITING state (notification + sound)
- User activity detected → Next work starts automatically
- Manual start available via menu during WAITING

## State Machine

```
IDLE (0/3)
  ↓ [Start clicked]
WORK (1/3) - 25:00
  ↓ [timer expires]
SHORT_BREAK (1/3) - 5:00
  ↓ [timer expires]
WAITING_WORK (1/3)
  ↓ [activity detected OR Start clicked]
WORK (2/3) - 25:00
  ↓ [timer expires]
SHORT_BREAK (2/3) - 5:00
  ↓ [timer expires]
WAITING_WORK (2/3)
  ↓ [activity detected OR Start clicked]
WORK (3/3) - 25:00
  ↓ [timer expires]
LONG_BREAK (3/3) - 20:00
  ↓ [timer expires]
WAITING_WORK (0/3)  [counter resets]
  ↓ [activity detected OR Start clicked]
WORK (1/3) - 25:00  [cycle repeats]
```

**Pause behavior:**
- Pause available during WORK and BREAK phases only
- Paused timer shows ⏸️ prefix
- Not available during WAITING or IDLE

## Systray UI

### Icon States

4 PNG icons (16x16 or 24x24):
- `idle.png` - Gray circle - shown when IDLE or PAUSED
- `work.png` - Red tomato - shown during WORK
- `break.png` - Green coffee cup - shown during SHORT_BREAK and LONG_BREAK
- `waiting.png` - Yellow hourglass - shown during WAITING_WORK

Icon updates every second based on current phase.

### Title (Panel Text)

**No hover tooltip works reliably.** Title is the only visible text in system panel.

Format patterns:
```
IDLE:           "Pomodoro"
WORK:           "🍅 25:00 [1/3]"
SHORT_BREAK:    "☕ 5:00 [1/3]"
LONG_BREAK:     "🌴 20:00 [3/3]"
WAITING:        "⏳ Ready [1/3]"
PAUSED:         "⏸️ 25:00 [1/3]"
```

Format: `[emoji] [MM:SS] [completed/total]`
- Emoji indicates phase type
- MM:SS shows time remaining (omitted for WAITING/IDLE)
- [N/M] shows session progress
- Paused adds ⏸️ prefix

Updates every second during active timer.

### Context Menu (Right-Click)

Menu structure (static, always present):

```
Start/Resume
Pause
Skip Phase
Reset
─────────────
Quit
```

**Menu item states:**

| State | Start/Resume | Pause | Skip Phase | Reset |
|-------|-------------|-------|------------|-------|
| IDLE | Enabled | Disabled | Disabled | Disabled |
| WORK running | Disabled | Enabled | Enabled | Enabled |
| WORK paused | Enabled | Disabled | Enabled | Enabled |
| BREAK running | Disabled | Enabled | Enabled | Enabled |
| BREAK paused | Enabled | Disabled | Enabled | Enabled |
| WAITING | Enabled | Disabled | Enabled | Enabled |

**Actions:**

- **Start/Resume**: 
  - IDLE → Start first work (1/3)
  - WAITING → Start next work
  - PAUSED → Resume timer
  
- **Pause**: 
  - WORK/BREAK running → Pause timer
  
- **Skip Phase**:
  - WORK → Skip to break (short or long based on session count)
  - BREAK → Skip to WAITING
  - WAITING → Start next work immediately
  
- **Reset**:
  - Any state → IDLE (0/3), clear all progress
  
- **Quit**:
  - Always enabled
  - Exit process (no state persistence)

### Click Behavior

**Left click on icon:** Toggle Start/Pause
- IDLE → Start first work
- WAITING → Start next work
- Running → Pause
- Paused → Resume

**Right click:** Show menu

**Middle click:** None (not reliable)

## Notifications

Use `notify-send` for desktop notifications:

```bash
notify-send -a "Pomodoro" -i "appointment-soon" "Pomodoro Timer" "<message>"
```

**Notification triggers:**

| Event | Message |
|-------|---------|
| Work complete (→ short break) | "Work complete! Time for a short break." |
| Work complete (→ long break) | "Work complete! Time for a long break." |
| Break complete (→ waiting) | "Break over! Move your mouse when ready." |
| Timer paused | "Timer paused" |
| Timer resumed | "Timer resumed" |

## Sound

Embedded WAV files using `go:embed`:

**work.wav** (break → work transition):
- 800Hz sine wave
- 0.3 seconds duration
- 16-bit PCM, 44.1kHz sample rate

**break.wav** (work → break transition):
- 600Hz sine wave  
- 0.5 seconds duration
- 16-bit PCM, 44.1kHz sample rate

Playback priority:
1. Try `paplay <file>` (PulseAudio)
2. Fallback `aplay -q <file>` (ALSA)
3. Silent fail if both unavailable

Extract embedded files to `/tmp/pomodoro-work-<uid>.wav` and `/tmp/pomodoro-break-<uid>.wav` on startup, reuse if exists.

## Activity Detection

Use `xprintidle` to detect user activity during WAITING state.

**Logic:**
- Poll every 500ms while in WAITING state
- Track previous idle time
- Activity detected if: `currentIdle < previousIdle` OR `currentIdle < 1000ms`
- On activity: automatically transition to next work phase
- If xprintidle unavailable: disable auto-start, require manual Start click

## Configuration

TOML file at `~/.config/pomodoro/config.toml`:

```toml
work_minutes = 25
short_break_minutes = 5
long_break_minutes = 20
work_sessions_per_cycle = 3
```

Load on startup. Missing file or invalid values → use defaults above.

No hot-reload: changes require restart.

## File Structure

```
pomodoro/
├── main.go              # Entry point, systray setup
├── timer.go             # Timer state machine and logic
├── config.go            # TOML config loading
├── notify.go            # Notifications and sound
├── activity.go          # xprintidle monitoring
├── ui.go                # Systray menu and updates
├── sounds/
│   ├── work.wav
│   └── break.wav
├── icons/
│   ├── idle.png
│   ├── work.png
│   ├── break.png
│   └── waiting.png
├── go.mod
└── README.md
```

## Dependencies

```go
require (
    github.com/getlantern/systray v1.2.2
    github.com/BurntSushi/toml v1.5.0
)
```

System packages:
- `libayatana-appindicator3-dev` (Ubuntu/Debian)
- `xprintidle` (optional, for activity detection)
- `pulseaudio-utils` or `alsa-utils` (for sound)

## Build

```bash
go build -o pomodoro .
```

Single binary output. No installation required - run directly.

## Usage

```bash
# Run (stays in foreground)
./pomodoro

# Run in background
./pomodoro &

# Autostart: create ~/.config/autostart/pomodoro.desktop
[Desktop Entry]
Type=Application
Name=Pomodoro Timer
Exec=/path/to/pomodoro
X-GNOME-Autostart-enabled=true
```

## Edge Cases

**Process killed:** All state lost, no recovery. User restarts from IDLE.

**Activity detection fails:** User must manually click Start in WAITING state.

**Sound playback fails:** Silent, continue timer normally.

**Notification fails:** Silent, continue timer normally.

**Multiple instances:** Each creates own systray icon. No lock file. User's problem.

## Non-Features

- No CLI interface
- No daemon/client architecture
- No state persistence
- No pause during WAITING/IDLE
- No custom notification sounds via config (embedded only)
- No statistics or history tracking
- No system tray tooltip (unreliable, title only)
