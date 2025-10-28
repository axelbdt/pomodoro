package main

import (
	"fmt"
	"os"
)

func main() {
	// Default command is toggle
	cmd := "toggle"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "daemon":
		runDaemon()
	case "tray":
		runTray()
	case "toggle":
		runClient(CMD_TOGGLE)
	case "status":
		runClient(CMD_STATUS)
	case "skip":
		runClient(CMD_SKIP)
	case "reset":
		runClient(CMD_RESET)
	case "stop":
		runClient(CMD_STOP)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		fmt.Fprintf(os.Stderr, "Usage: pomodoro [daemon|tray|toggle|status|skip|reset|stop]\n")
		os.Exit(1)
	}
}
