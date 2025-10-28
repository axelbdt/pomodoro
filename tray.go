package main

import (
	_ "embed"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/getlantern/systray"
)

//go:embed icons/work.png
var workIcon []byte

//go:embed icons/break.png
var breakIcon []byte

//go:embed icons/waiting.png
var waitingIcon []byte

//go:embed icons/idle.png
var idleIcon []byte

var lastPhase string

func runTray() {
	systray.Run(onReady, onExit)
}

func onReady() {
	// Set initial icon
	systray.SetIcon(idleIcon)
	systray.SetTitle("Pomodoro")
	systray.SetTooltip("Pomodoro Timer - Starting...")

	// Ensure daemon is running
	ensureDaemon()

	// Menu items
	mToggle := systray.AddMenuItem("Start/Pause", "Toggle timer")
	mSkip := systray.AddMenuItem("Skip Phase", "Skip to next phase")
	mReset := systray.AddMenuItem("Reset Timer", "Reset to idle")
	systray.AddSeparator()
	mStopDaemon := systray.AddMenuItem("Stop Daemon", "Stop daemon process")
	mQuit := systray.AddMenuItem("Quit Tray", "Exit tray application")

	// Start update loop
	go updateLoop()

	// Handle menu clicks
	go func() {
		for {
			select {
			case <-mToggle.ClickedCh:
				sendCommand(CMD_TOGGLE)
			case <-mSkip.ClickedCh:
				sendCommand(CMD_SKIP)
			case <-mReset.ClickedCh:
				sendCommand(CMD_RESET)
			case <-mStopDaemon.ClickedCh:
				sendCommand(CMD_STOP)
				time.Sleep(200 * time.Millisecond)
				systray.Quit()
			case <-mQuit.ClickedCh:
				systray.Quit()
			}
		}
	}()
}

func onExit() {
	// Cleanup on exit
}

func ensureDaemon() {
	socketPath := getSocketPath()
	if !socketExists(socketPath) {
		if err := startDaemon(); err != nil {
			log.Printf("Failed to start daemon: %v", err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func updateLoop() {
	// Give daemon time to start
	time.Sleep(500 * time.Millisecond)
	
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		response, err := sendCommand(CMD_STATUS)
		if err != nil {
			systray.SetTitle("Disconnected")
			systray.SetTooltip("Disconnected - click Start/Pause to begin")
			if lastPhase != "disconnected" {
				systray.SetIcon(idleIcon)
				lastPhase = "disconnected"
			}
			continue
		}

		// Parse: "WORK 1234 2/3 running"
		parts := strings.Split(response, " ")
		if len(parts) < 4 {
			log.Printf("Invalid status format: %s", response)
			systray.SetTitle("Error")
			systray.SetTooltip("Error parsing status")
			continue
		}

		phase := parts[0]
		seconds, _ := strconv.Atoi(parts[1])
		cycle := parts[2]
		status := parts[3]

		// Update icon if phase changed
		updateIcon(phase, status)

		// Format for display
		mins := seconds / 60
		secs := seconds % 60
		timeStr := fmt.Sprintf("%02d:%02d", mins, secs)
		
		// Set title (shows in panel)
		var title string
		switch phase {
		case WORK:
			title = fmt.Sprintf("[Work] %s [%s]", timeStr, cycle)
		case SHORT_BREAK:
			title = fmt.Sprintf("[Break] %s [%s]", timeStr, cycle)
		case LONG_BREAK:
			title = fmt.Sprintf("[Long Break] %s [%s]", timeStr, cycle)
		case WAITING_WORK:
			title = fmt.Sprintf("[Waiting] [%s]", cycle)
		case IDLE:
			title = "[Idle]  Idle"
		default:
			title = "Pomodoro"
		}
		
		if status == "paused" {
			title = "⏸️  " + title
		}
		
		systray.SetTitle(title)
		
		// Set full tooltip
		tooltip := formatTooltip(phase, seconds, cycle, status)
		systray.SetTooltip(tooltip)
	}
}

func updateIcon(phase string, status string) {
	// Determine which icon to show
	var iconKey string
	
	if status == "paused" {
		iconKey = "idle"
	} else {
		switch phase {
		case WORK:
			iconKey = "work"
		case SHORT_BREAK, LONG_BREAK:
			iconKey = "break"
		case WAITING_WORK:
			iconKey = "waiting"
		case IDLE:
			iconKey = "idle"
		default:
			iconKey = "idle"
		}
	}

	// Only update if changed
	if lastPhase != iconKey {
		switch iconKey {
		case "work":
			systray.SetIcon(workIcon)
		case "break":
			systray.SetIcon(breakIcon)
		case "waiting":
			systray.SetIcon(waitingIcon)
		case "idle":
			systray.SetIcon(idleIcon)
		}
		lastPhase = iconKey
	}
}

func formatTooltip(phase string, seconds int, cycle string, status string) string {
	mins := seconds / 60
	secs := seconds % 60
	timeStr := fmt.Sprintf("%02d:%02d", mins, secs)

	var result string

	switch phase {
	case WORK:
		result = fmt.Sprintf("🍅 Work: %s [%s]", timeStr, cycle)
	case SHORT_BREAK:
		result = fmt.Sprintf("☕ Short Break: %s [%s]", timeStr, cycle)
	case LONG_BREAK:
		result = fmt.Sprintf("🌴 Long Break: %s [%s]", timeStr, cycle)
	case WAITING_WORK:
		result = fmt.Sprintf("⏳ Waiting for activity... [%s]", cycle)
	case IDLE:
		result = "⏸️  Idle - Click to start"
	default:
		result = "Unknown"
	}

	if status == "paused" {
		result = "⏸️  Paused - " + result
	}

	return result
}
