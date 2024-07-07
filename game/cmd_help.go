package game

import (
	"fmt"
	"log"
	"strings"
)

func handleHelp(g *Game, session *Session, params string, help bool) []OutputEvent {

	if help {
		return []OutputEvent{{
			SessionID: session.ID,
			Message:   "List available commands.\nUsage: /help [<command>]",
		}}
	}
	if len(params) > 0 {
		parts := strings.Split(params, " ")
		if len(parts) > 0 {
			cmd := parts[0]
			log.Printf("User %s issued help command for %s. '%s': %+v, %d", session.Name, cmd, params, parts, len(parts))

			if command, exists := g.commands[cmd]; exists {
				return command(g, session, params, true)
			} else {
				return []OutputEvent{{
					SessionID: session.ID,
					Message:   fmt.Sprintf("Unknown command: %s", cmd),
				}}
			}
		}
	}
	commandNames := []string{}
	for name := range g.commands {
		commandNames = append(commandNames, fmt.Sprintf("/%s", name))
	}

	return []OutputEvent{{
		SessionID: session.ID,
		Message:   fmt.Sprintf("Available commands: %s", strings.Join(commandNames, ", ")),
	}}
}
