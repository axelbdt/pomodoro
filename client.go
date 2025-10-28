package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

func runClient(command string) {
	// Ensure daemon is running
	socketPath := getSocketPath()
	if !socketExists(socketPath) {
		if err := startDaemon(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start daemon: %v\n", err)
			os.Exit(1)
		}
	}

	// Send command
	response, err := sendCommand(command)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Print response
	fmt.Println(response)
}

func sendCommand(command string) (string, error) {
	socketPath := getSocketPath()

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return "", fmt.Errorf("failed to connect to daemon: %v", err)
	}
	defer conn.Close()

	// Send command
	fmt.Fprintf(conn, "%s\n", command)

	// Read response
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	return strings.TrimSpace(response), nil
}

func socketExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func startDaemon() error {
	// Get path to current executable
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}

	// Start daemon in background
	cmd := exec.Command(exePath, "daemon")
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon process: %v", err)
	}

	// Wait for socket to be created
	socketPath := getSocketPath()
	for i := 0; i < 20; i++ {
		if socketExists(socketPath) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("daemon failed to create socket")
}
