package main

import (
	"log"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func monitorActivity() {
	if !commandExists("xprintidle") {
		log.Println("xprintidle not found, activity detection disabled")
		return
	}

	var prevIdle int64 = math.MaxInt64

	for {
		select {
		case <-activityStopChan:
			return
		default:
			// Check if still in WAITING_WORK state
			state.Lock()
			if !state.WaitingForActivity {
				state.Unlock()
				return
			}
			state.Unlock()

			// Get current idle time
			cmd := exec.Command("xprintidle")
			output, err := cmd.Output()
			if err != nil {
				time.Sleep(500 * time.Millisecond)
				continue
			}

			currentIdle, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
			if err != nil {
				time.Sleep(500 * time.Millisecond)
				continue
			}

			// Activity detected: idle time decreased or very low
			if currentIdle < prevIdle || currentIdle < 1000 {
				onActivityDetected()
				return
			}

			prevIdle = currentIdle
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func onActivityDetected() {
	state.Lock()
	defer state.Unlock()

	if state.WaitingForActivity {
		startNextWorkPhase()
	}
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
