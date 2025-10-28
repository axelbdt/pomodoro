package main

import (
	"fmt"
	"strings"
)

// Phase constants
const (
	IDLE         = "IDLE"
	WORK         = "WORK"
	SHORT_BREAK  = "SHORT_BREAK"
	LONG_BREAK   = "LONG_BREAK"
	WAITING_WORK = "WAITING_WORK"
)

// Command constants
const (
	CMD_TOGGLE = "TOGGLE"
	CMD_STATUS = "STATUS"
	CMD_SKIP   = "SKIP"
	CMD_RESET  = "RESET"
	CMD_STOP   = "STOP"
)

// parseCommand extracts the command from a raw line
func parseCommand(line string) string {
	return strings.TrimSpace(strings.ToUpper(line))
}

// formatStatusResponse creates a STATUS response
// Format: [PHASE] [SECONDS] [N/M] [STATUS]
func formatStatusResponse(phase string, seconds int, completed int, total int, paused bool, waitingForActivity bool) string {
	status := "running"
	if paused {
		status = "paused"
	} else if waitingForActivity {
		status = "waiting"
	} else if phase == IDLE {
		status = "stopped"
	}

	return fmt.Sprintf("%s %d %d/%d %s", phase, seconds, completed, total, status)
}

// formatOKResponse creates an OK response with optional message
func formatOKResponse(message string) string {
	if message == "" {
		return "OK"
	}
	return fmt.Sprintf("OK %s", message)
}

// formatErrorResponse creates an error response
func formatErrorResponse(message string) string {
	return fmt.Sprintf("ERR %s", message)
}
