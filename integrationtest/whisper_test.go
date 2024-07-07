package integrationtest

import (
	"bufio"
	"strings"
	"testing"
	"time"
)

func TestWhisper(t *testing.T) {
	startServer(t)
	defer stopServer()

	// Connect three users
	aliceConn := connectTelnet(t)
	defer aliceConn.Close()
	bobConn := connectTelnet(t)
	defer bobConn.Close()
	charlieConn := connectTelnet(t)
	defer charlieConn.Close()

	// Set names for users
	sendCommand(t, aliceConn, "Alice")
	sendCommand(t, bobConn, "Bob")
	sendCommand(t, charlieConn, "Charlie")

	// Clear initial messages
	readResponses(t, aliceConn, 4)  // Welcome + Who are you? + 2 join messages
	readResponses(t, bobConn, 3)    // Welcome + Who are you? + 1 join message
	readResponses(t, charlieConn, 2) // Welcome + Who are you?

	// Alice whispers to Bob
	sendCommand(t, aliceConn, "/whisper Bob Hello, this is a secret message")

	// Check Alice's confirmation
	aliceResponse := readResponses(t, aliceConn, 1)[0]
	expectedAliceResponse := "You whispered to Bob: Hello, this is a secret message"
	if aliceResponse != expectedAliceResponse {
			t.Errorf("Unexpected response for Alice: got %s, want %s", aliceResponse, expectedAliceResponse)
	}

	// Check Bob's received whisper
	bobResponse := readResponses(t, bobConn, 1)[0]
	expectedBobResponse := "Alice whispers: Hello, this is a secret message"
	if bobResponse != expectedBobResponse {
			t.Errorf("Unexpected response for Bob: got %s, want %s", bobResponse, expectedBobResponse)
	}

	// Verify Charlie didn't receive the whisper
	charlieConn.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
	reader := bufio.NewReader(charlieConn)
	_, err := reader.ReadString('\n')
	if err == nil {
			t.Errorf("Charlie unexpectedly received a message")
	} else if !strings.Contains(err.Error(), "timeout") {
			t.Errorf("Unexpected error when checking Charlie's connection: %v", err)
	}

	// Test whispering to non-existent user
	sendCommand(t, aliceConn, "/whisper NonExistentUser This should fail")
	aliceErrorResponse := readResponses(t, aliceConn, 1)[0]
	expectedErrorResponse := "User 'NonExistentUser' not found."
	if aliceErrorResponse != expectedErrorResponse {
			t.Errorf("Unexpected error response: got %s, want %s", aliceErrorResponse, expectedErrorResponse)
	}
}