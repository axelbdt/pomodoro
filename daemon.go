package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	state              *TimerState
	config             *Config
	timerStopChan      chan bool
	activityStopChan   chan bool
	workSoundPath      string
	breakSoundPath     string
)

func runDaemon() {
	log.SetPrefix("[pomodoro-daemon] ")

	// Load configuration
	config = LoadConfig()

	// Initialize state
	state = NewTimerState(config.WorkSessionsPerCycle)

	// Extract embedded sounds
	workSoundPath, breakSoundPath = extractEmbeddedSounds()

	// Get socket path
	socketPath := getSocketPath()

	// Check if socket already exists (another daemon running)
	if _, err := os.Stat(socketPath); err == nil {
		log.Fatalf("Socket already exists at %s, another daemon may be running", socketPath)
	}

	// Create Unix socket
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Failed to create socket: %v", err)
	}
	defer os.Remove(socketPath)
	defer listener.Close()

	log.Printf("Daemon started, listening on %s", socketPath)

	// Start timer goroutine
	timerStopChan = make(chan bool)
	go timerLoop()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal")
		timerStopChan <- true
		if activityStopChan != nil {
			activityStopChan <- true
		}
		listener.Close()
		os.Exit(0)
	}()

	// Accept client connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Read error: %v", err)
		return
	}

	cmd := parseCommand(line)
	response := handleCommand(cmd)

	fmt.Fprintf(conn, "%s\n", response)
}

func handleCommand(cmd string) string {
	state.Lock()
	defer state.Unlock()

	switch cmd {
	case CMD_TOGGLE:
		return handleToggle()
	case CMD_STATUS:
		return handleStatus()
	case CMD_SKIP:
		return handleSkip()
	case CMD_RESET:
		return handleReset()
	case CMD_STOP:
		go func() {
			time.Sleep(100 * time.Millisecond)
			timerStopChan <- true
			if activityStopChan != nil {
				activityStopChan <- true
			}
			os.Exit(0)
		}()
		return formatOKResponse("")
	default:
		return formatErrorResponse("Unknown command")
	}
}

func handleToggle() string {
	if state.Phase == IDLE {
		// Start first work session
		startPhase(WORK, config.WorkMinutes*60)
		return formatOKResponse("started")
	} else if state.WaitingForActivity {
		// Manual start of next work session
		startNextWorkPhase()
		return formatOKResponse("started")
	} else if state.Paused {
		// Resume
		state.Paused = false
		sendNotification("Timer Resumed")
		return formatOKResponse("running")
	} else {
		// Pause
		state.Paused = true
		sendNotification("Timer Paused")
		return formatOKResponse("paused")
	}
}

func handleStatus() string {
	return formatStatusResponse(
		state.Phase,
		state.SecondsRemaining,
		state.CompletedSessions,
		state.TotalSessions,
		state.Paused,
		state.WaitingForActivity,
	)
}

func handleSkip() string {
	if state.Phase == IDLE {
		return formatErrorResponse("No active timer")
	}

	// Determine next phase
	var nextPhase string
	switch state.Phase {
	case WORK:
		if state.CompletedSessions+1 >= state.TotalSessions {
			nextPhase = LONG_BREAK
			state.CompletedSessions++
			startPhase(LONG_BREAK, config.LongBreakMinutes*60)
		} else {
			nextPhase = SHORT_BREAK
			state.CompletedSessions++
			startPhase(SHORT_BREAK, config.ShortBreakMinutes*60)
		}
	case SHORT_BREAK, LONG_BREAK:
		if state.Phase == LONG_BREAK {
			state.CompletedSessions = 0
		}
		enterWaitingState()
		nextPhase = WAITING_WORK
	case WAITING_WORK:
		startNextWorkPhase()
		nextPhase = WORK
	}

	return formatOKResponse(nextPhase)
}

func handleReset() string {
	if activityStopChan != nil {
		activityStopChan <- true
		activityStopChan = nil
	}

	state.Phase = IDLE
	state.SecondsRemaining = 0
	state.Paused = false
	state.CompletedSessions = 0
	state.WaitingForActivity = false

	return formatOKResponse("idle")
}

func timerLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timerStopChan:
			return
		case <-ticker.C:
			state.Lock()

			if state.Paused || state.Phase == IDLE || state.WaitingForActivity {
				state.Unlock()
				continue
			}

			state.SecondsRemaining--

			if state.SecondsRemaining <= 0 {
				handlePhaseComplete()
			}

			state.Unlock()
		}
	}
}

func handlePhaseComplete() {
	// Called with state lock held

	switch state.Phase {
	case WORK:
		state.CompletedSessions++

		if state.CompletedSessions >= state.TotalSessions {
			// After 3rd work: long break
			sendNotification("Work Complete! Time for a long break.")
			playSound(breakSoundPath)
			startPhase(LONG_BREAK, config.LongBreakMinutes*60)
		} else {
			// After 1st/2nd work: short break
			sendNotification("Work Complete! Time for a short break.")
			playSound(breakSoundPath)
			startPhase(SHORT_BREAK, config.ShortBreakMinutes*60)
		}

	case SHORT_BREAK:
		sendNotification(fmt.Sprintf("Break Over! Ready for work session %d/%d?", state.CompletedSessions+1, state.TotalSessions))
		playSound(workSoundPath)
		enterWaitingState()

	case LONG_BREAK:
		state.CompletedSessions = 0
		sendNotification("Long Break Over! Starting new cycle.")
		playSound(workSoundPath)
		enterWaitingState()
	}
}

func startPhase(phase string, seconds int) {
	// Called with state lock held
	state.Phase = phase
	state.SecondsRemaining = seconds
	state.Paused = false
	state.WaitingForActivity = false

	// Stop activity monitor if running
	if activityStopChan != nil {
		activityStopChan <- true
		activityStopChan = nil
	}
}

func enterWaitingState() {
	// Called with state lock held
	state.Phase = WAITING_WORK
	state.SecondsRemaining = 0
	state.WaitingForActivity = true
	state.Paused = false

	sendNotification("Move your mouse to start next work session")

	// Start activity monitor
	activityStopChan = make(chan bool)
	go monitorActivity()
}

func startNextWorkPhase() {
	// Called with state lock held
	startPhase(WORK, config.WorkMinutes*60)
}

func getSocketPath() string {
	uid := os.Getuid()
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")

	if runtimeDir != "" {
		return fmt.Sprintf("%s/pomodoro.sock", runtimeDir)
	}

	return fmt.Sprintf("/tmp/pomodoro-%d.sock", uid)
}
