package integrationtest

import (
	"bufio"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"testing"
	"time"
)

var serverCmd *exec.Cmd

func isPortInUse(port string) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("localhost", port), time.Second)
	if err != nil {
		return false
	}
	if conn != nil {
		conn.Close()
		return true
	}
	return false
}

func killProcessOnPort(port string) {
	killCmd := exec.Command("sh", "-c", fmt.Sprintf("lsof -ti :%s | xargs kill -9", port))
	killCmd.Run()
}

func startServer(t *testing.T) {
	stopServer() // Ensure any previous server is stopped

	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		if !isPortInUse("2323") {
			break
		}
		if i == maxRetries-1 {
			t.Fatal("Port 2323 is still in use after multiple attempts to free it")
		}
		time.Sleep(100 * time.Millisecond)
	}

	serverCmd = exec.Command("go", "run", "../main.go")
	err := serverCmd.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	time.Sleep(300 * time.Millisecond) // Give the server time to start
}

func stopServer() {
	if serverCmd != nil && serverCmd.Process != nil {
		serverCmd.Process.Kill()
		serverCmd.Wait()
	}

	killProcessOnPort("2323")
}

func connectTelnet(t *testing.T) net.Conn {
	conn, err := net.Dial("tcp", "localhost:2323")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	return conn
}

func sendCommand(t *testing.T, conn net.Conn, command string) {
	_, err := conn.Write([]byte(command + "\n"))
	if err != nil {
		t.Fatalf("Failed to send command: %v", err)
	}
	defer time.Sleep(time.Millisecond)
}

func readResponses(t *testing.T, conn net.Conn, expectedCount int) []string {
	var responses []string
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	reader := bufio.NewReader(conn)

	for i := 0; i < expectedCount; i++ {
		response, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("Failed to read response %d: %v", i+1, err)
		}
		responses = append(responses, strings.TrimSpace(response))
	}

	return responses
}
