package integrationtest

import (
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
)

func TestBasicConnectionAndNaming(t *testing.T) {
	startServer(t)
	defer stopServer()

	conn1 := connectTelnet(t)
	defer conn1.Close()
	conn2 := connectTelnet(t)
	defer conn2.Close()

	// Read welcome message and name prompt for conn1
	responses1 := readResponses(t, conn1, 2)
	if responses1[0] != "Welcome to the MUD server!" {
		t.Errorf("Unexpected welcome message: %s. Expected 'Welcome to the MUD server!'", responses1[0])
	}
	if responses1[1] != "Who are you?" {
		t.Errorf("Unexpected prompt: %s. Expected 'Who are you?'", responses1[1])
	}

	// Read welcome message and name prompt for conn2
	responses2 := readResponses(t, conn2, 2)
	if responses2[0] != "Welcome to the MUD server!" {
		t.Errorf("Unexpected welcome message: %s, expected 'Welcome to the MUD server!'", responses2[0])
	}
	if responses2[1] != "Who are you?" {
		t.Errorf("Unexpected prompt: %s, expected 'Who are you?'", responses2[1])
	}

	// Set names and verify welcome messages
	sendCommand(t, conn1, "Alice")
	responses1 = readResponses(t, conn1, 1)
	if responses1[0] != "Welcome, Alice!" {
		t.Errorf("Unexpected welcome message on Alice's side: %s. Expected 'Welcome, Alice!'", responses1[0])
	}

	responses2 = readResponses(t, conn2, 1)
	if responses2[0] != "Alice has joined the room." {
		t.Errorf("Unexpected join message: %s, expected 'Alice has joined the room.'", responses2[0])
	}

	sendCommand(t, conn2, "Bob")
	responses2 = readResponses(t, conn2, 1)
	if responses2[0] != "Welcome, Bob!" {
		t.Errorf("Unexpected welcome message: %s, expected 'Welcome, Bob!'", responses2[0])
	}

	responses1 = readResponses(t, conn1, 1)
	if responses1[0] != "Bob has joined the room." {
		t.Errorf("Unexpected join message: %s, expected 'Bob has joined the room.''", responses1[0])
	}
}

func TestManyJoins(t *testing.T) {
	startServer(t)
	defer stopServer()

	conns := make([]net.Conn, 10)

	aliceConn := connectTelnet(t)
	defer aliceConn.Close()

	// Connect all other clients
	for i := 0; i < 10; i++ {
		conns[i] = connectTelnet(t)
		defer conns[i].Close()
	}

	// Set Alice's name
	sendCommand(t, aliceConn, "Alice")
	readResponses(t, aliceConn, 2) // Read Alice's welcome messages

	// Set names for other clients and have them join
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("test%d", i+1)
		sendCommand(t, conns[i], name)
	}

	// Read join messages for Alice
	joinMessages := readResponses(t, aliceConn, 10)

	// Check if Alice saw all 10 joins
	seenJoins := make(map[string]bool)
	for _, msg := range joinMessages {
		parts := strings.Split(msg, " ")
		if len(parts) >= 4 && parts[1] == "has" && parts[2] == "joined" {
			seenJoins[parts[0]] = true
		} else {
			t.Errorf("Unexpected message format: %s", msg)
		}
	}

	// Verify all 10 test users were seen joining
	for i := 1; i <= 10; i++ {
		name := fmt.Sprintf("test%d", i)
		if !seenJoins[name] {
			t.Errorf("Alice did not see %s join the room", name)
		}
	}

	if len(seenJoins) > 10 {
		t.Errorf("Alice saw more join messages than expected: %d", len(seenJoins))
	}
}

func TestMain(m *testing.M) {
	// Run before any test
	stopServer() // Ensure no leftover server is running

	// Run tests
	code := m.Run()

	// Run after all tests
	stopServer()

	os.Exit(code)
}
