package integrationtest

import (
	"strings"
	"testing"
)

func TestWhoCommand(t *testing.T) {
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

	// Alice uses the /who command
	sendCommand(t, aliceConn, "/who")

	// Check Alice's response
	aliceResponse := readResponses(t, aliceConn, 1)[0]
	if !strings.HasPrefix(aliceResponse, "Users in this room:") {
			t.Errorf("Unexpected response format: %s", aliceResponse)
	}

	// Extract the list of users
	userList := strings.TrimPrefix(aliceResponse, "Users in this room: ")
	users := strings.Split(userList, ", ")

	// Check that we have the correct number of users
	if len(users) != 3 {
			t.Errorf("Expected 3 users, got %d", len(users))
	}

	// Check that all expected users are in the list
	expectedUsers := map[string]bool{
			"Alice":   false,
			"Bob":     false,
			"Charlie": false,
	}

	for _, user := range users {
			if _, exists := expectedUsers[user]; exists {
					expectedUsers[user] = true
			} else {
					t.Errorf("Unexpected user in list: %s", user)
			}
	}

	// Check that all expected users were found
	for user, found := range expectedUsers {
			if !found {
					t.Errorf("Expected user %s was not in the list", user)
			}
	}
}