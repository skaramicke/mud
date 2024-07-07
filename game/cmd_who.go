package game

import (
	"fmt"
	"strings"
)

func handleListUsersInRoom(_ *Game, session *Session, _ string, help bool) []OutputEvent {
	if help {
		return []OutputEvent{
			{
				SessionID: session.ID,
				Message:   "List users in the room.\nUsage: /who",
			},
		}
	}

	userList := []string{}
	for _, s := range session.Room.Sessions {
		userList = append(userList, s.Name)
	}
	return []OutputEvent{{
		SessionID: session.ID,
		Message:   fmt.Sprintf("Users in this room: %s", strings.Join(userList, ", ")),
	}}
}
