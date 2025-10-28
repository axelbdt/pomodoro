# Pomodoro Timer

Single-binary CLI pomodoro timer with GTK tray icon for Linux.

## Files Created

```
pomodoro/
├── main.go              # Entry point and mode routing
├── protocol.go          # Command/response protocol and state constants
├── state.go             # Timer state structure (in-memory only)
├── config.go            # TOML configuration loading
├── daemon.go            # Daemon with timer logic and socket server
├── client.go            # Client commands and daemon auto-start
├── notify.go            # Desktop notifications and sound playback
├── activity.go          # Activity detection via xprintidle
├── tray.go              # GTK status icon with tooltip and menu
├── sounds/
│   ├── work.wav         # Work start sound (800Hz, 0.3s)
│   └── break.wav        # Break start sound (600Hz, 0.5s)
├── go.mod               # Go module definition
└── README.md            # This file
```

## Build Instructions

### Prerequisites

Install system dependencies:

```bash
sudo apt install golang libgtk-3-dev libayatana-appindicator3-dev xprintidle pulseaudio-utils
```

### Build

```bash
cd pomodoro
go mod tidy
go build -o pomodoro .
```

### Install

```bash
sudo install -m 755 pomodoro /usr/local/bin/
```

## Configuration

Create config file:

```bash
mkdir -p ~/.config/pomodoro
cat > ~/.config/pomodoro/config.toml <<EOF
work_minutes = 25
short_break_minutes = 5
long_break_minutes = 20
work_sessions_per_cycle = 3
sound_work_start = ""
sound_break_start = ""
EOF
```

## Usage

### Commands

```bash
pomodoro              # Toggle: start/pause/resume (default)
pomodoro status       # Show current state
pomodoro skip         # Skip to next phase
pomodoro reset        # Stop timer and reset to idle
pomodoro stop         # Kill daemon process
pomodoro daemon       # Run daemon (auto-started by client)
pomodoro tray         # Launch GTK tray icon
```

### Tray Icon

Launch tray icon (recommended):

```bash
pomodoro tray &
```

The tray icon will:
- Auto-start daemon if not running
- Update tooltip every second with current status
- Left-click to toggle pause/resume
- Right-click for menu (skip, reset, stop daemon, quit)

### Autostart

Create autostart entry:

```bash
mkdir -p ~/.config/autostart
cat > ~/.config/autostart/pomodoro.desktop <<EOF
[Desktop Entry]
Type=Application
Name=Pomodoro Timer
Exec=/usr/local/bin/pomodoro tray
X-GNOME-Autostart-enabled=true
EOF
```

## Timer Cycle

- 3 work sessions of 25 minutes each
- 2 short breaks of 5 minutes (between work sessions)
- 1 long break of 20 minutes (after 3rd work session)

Flow: Work(1) → Short Break → Work(2) → Short Break → Work(3) → Long Break → [repeat]

### Phase Behavior

- **Work completes**: Notification + sound → break starts automatically
- **Break completes**: Notification + sound → wait for activity → auto-start work
- **Activity detection**: Monitors mouse/keyboard via xprintidle

## Testing

### Basic Test

```bash
# Terminal 1: Start daemon manually
./pomodoro daemon

# Terminal 2: Test commands
./pomodoro status       # Should show IDLE
./pomodoro              # Start timer
./pomodoro status       # Should show WORK with countdown
./pomodoro              # Pause
./pomodoro              # Resume
./pomodoro skip         # Jump to next phase
./pomodoro reset        # Back to IDLE
./pomodoro stop         # Stop daemon
```

### Kill Test

```bash
./pomodoro              # Start timer
./pomodoro status       # Note current state
pkill -9 pomodoro       # Kill daemon
./pomodoro status       # Should start fresh in IDLE (no state recovery)
```

### Tray Test

```bash
./pomodoro tray
# - Verify tooltip updates every second
# - Test left-click toggle
# - Test right-click menu actions
```

## Implementation Notes

### State Management

- All state kept in memory only
- No persistence across daemon restarts
- Daemon kill/crash always resets to IDLE

### Socket Communication

- Path: `/run/user/$UID/pomodoro.sock` (fallback: `/tmp/pomodoro-$UID.sock`)
- Protocol: Line-based text, newline-terminated
- Commands: TOGGLE, STATUS, SKIP, RESET, STOP

### Activity Detection

- Uses `xprintidle` to monitor X11 idle time
- Polls every 500ms during WAITING_WORK state
- Detects activity when idle time decreases or < 1000ms
- Falls back to manual start if xprintidle unavailable

### Notifications

- Uses `notify-send` for desktop notifications
- Sounds play via `paplay` (PulseAudio) or `aplay` (ALSA)
- Embedded WAV files extracted to `/tmp/pomodoro-*-$UID.wav`

## Architecture

- **Daemon**: Single-threaded event loop with timer goroutine
- **Client**: Connects to daemon socket, sends command, prints response
- **Tray**: GTK event loop, polls daemon every second for status updates
- **Concurrency**: Mutex protects state from timer/activity/command goroutines

## Dependencies

- Go modules:
  - `github.com/gotk3/gotk3` - GTK3 bindings
  - `github.com/BurntSushi/toml` - TOML config parser

- System packages:
  - `libgtk-3-dev` - GTK development files
  - `xprintidle` - X11 idle time monitor
  - `pulseaudio-utils` or `alsa-utils` - Sound playback
