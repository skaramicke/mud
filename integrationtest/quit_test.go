package integrationtest

import (
	"bufio"
	"strings"
	"testing"
	"time"
)

func TestQuitCommand(t *testing.T) {
	startServer(t)
	defer stopServer()

	// Connect two users
	aliceConn := connectTelnet(t)
	defer aliceConn.Close()
	bobConn := connectTelnet(t)
	defer bobConn.Close()

	// Set names for users
	sendCommand(t, aliceConn, "Alice")
	sendCommand(t, bobConn, "Bob")

	// Clear initial messages
	readResponses(t, aliceConn, 3) // Welcome + Who are you? + 1 join message
	readResponses(t, bobConn, 2)   // Welcome + Who are you?

	// Alice quits
	sendCommand(t, aliceConn, "/quit")

	// Check Alice's goodbye message
	aliceResponse := readResponses(t, aliceConn, 1)[0]
	expectedGoodbye := "Goodbye!"
	if aliceResponse != expectedGoodbye {
		t.Errorf("Unexpected goodbye message for Alice: got %s, want %s", aliceResponse, expectedGoodbye)
	}

	// Check that Bob receives notification of Alice leaving
	bobResponse := readResponses(t, bobConn, 1)[0]
	expectedNotification := "Alice has left the room."
	if bobResponse != expectedNotification {
		t.Errorf("Unexpected notification for Bob: got %s, want %s", bobResponse, expectedNotification)
	}

	// Verify that Alice's connection is closed
	aliceConn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	_, err := bufio.NewReader(aliceConn).ReadString('\n')
	if err == nil {
		t.Errorf("Expected Alice's connection to be closed, but it's still open")
	} else if !strings.Contains(err.Error(), "EOF") && !strings.Contains(err.Error(), "use of closed network connection") {
		t.Errorf("Unexpected error when checking Alice's connection: %v", err)
	}

	// Verify that Bob can still use commands
	sendCommand(t, bobConn, "/who")
	bobResponse = readResponses(t, bobConn, 1)[0]
	if !strings.Contains(bobResponse, "Users in this room: Bob") {
		t.Errorf("Unexpected /who response for Bob after Alice quit: %s", bobResponse)
	}
}