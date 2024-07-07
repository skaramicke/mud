package integrationtest

import (
	"strings"
	"testing"
)

func TestHelpCommand(t *testing.T) {
	startServer(t)
	defer stopServer()

	// Connect a user
	conn := connectTelnet(t)
	defer conn.Close()

	// Set name for user
	sendCommand(t, conn, "Alice")
	readResponses(t, conn, 2) // Welcome + Who are you?

	// Test general help command
	sendCommand(t, conn, "/help")
	response := strings.Join(readResponses(t, conn, 1), "\n")

	// Check that the response includes "Available commands:"
	if !strings.Contains(response, "Available commands:") {
		t.Errorf("Help response doesn't contain 'Available commands:': %s", response)
	}

	// Check that all expected commands are listed
	expectedCommands := []string{"/whisper", "/who", "/help", "/quit"}
	for _, cmd := range expectedCommands {
		if !strings.Contains(response, cmd) {
			t.Errorf("Help response doesn't contain expected command '%s': %s", cmd, response)
		}
	}

	// Test help for specific commands
	specificCommands := []string{"whisper", "who", "help", "quit"}
	for _, cmd := range specificCommands {
		sendCommand(t, conn, "/help "+cmd)
		response = strings.Join(readResponses(t, conn, 2), "\n")

		// Check that the response includes "Usage:" for each command
		if !strings.Contains(response, "Usage:") {
			t.Errorf("Help response for '%s' doesn't contain 'Usage:': %s", cmd, response)
		}

		// Check that the response includes the command name
		if !strings.Contains(response, cmd) {
			t.Errorf("Help response for '%s' doesn't contain the command name: %s", cmd, response)
		}
	}

	// Test help for non-existent command
	sendCommand(t, conn, "/help nonexistentcommand")
	response = readResponses(t, conn, 1)[0]
	expectedErrorResponse := "Unknown command: nonexistentcommand"
	if response != expectedErrorResponse {
		t.Errorf("Unexpected error response for non-existent command: got %s, want %s", response, expectedErrorResponse)
	}
}