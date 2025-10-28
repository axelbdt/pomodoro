package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"
)

//go:embed sounds/work.wav
var workSoundData []byte

//go:embed sounds/break.wav
var breakSoundData []byte

func sendNotification(message string) {
	cmd := exec.Command("notify-send", "-a", "Pomodoro", "-i", "appointment-soon", "Pomodoro Timer", message)
	err := cmd.Run()
	if err != nil {
		log.Printf("Notification failed: %v", err)
	}
}

func playSound(soundPath string) {
	// Try paplay first (PulseAudio)
	cmd := exec.Command("paplay", soundPath)
	err := cmd.Run()
	if err == nil {
		return
	}

	// Fallback to aplay (ALSA)
	cmd = exec.Command("aplay", "-q", soundPath)
	err = cmd.Run()
	if err != nil {
		log.Printf("Sound playback failed: %v", err)
	}
}

func extractEmbeddedSounds() (workPath, breakPath string) {
	uid := os.Getuid()
	workPath = fmt.Sprintf("/tmp/pomodoro-work-%d.wav", uid)
	breakPath = fmt.Sprintf("/tmp/pomodoro-break-%d.wav", uid)

	// Extract work sound if not exists
	if _, err := os.Stat(workPath); os.IsNotExist(err) {
		if err := os.WriteFile(workPath, workSoundData, 0644); err != nil {
			log.Printf("Failed to extract work sound: %v", err)
		}
	}

	// Extract break sound if not exists
	if _, err := os.Stat(breakPath); os.IsNotExist(err) {
		if err := os.WriteFile(breakPath, breakSoundData, 0644); err != nil {
			log.Printf("Failed to extract break sound: %v", err)
		}
	}

	return workPath, breakPath
}
