package game

func handleQuit(g *Game, session *Session, params string, help bool) []OutputEvent {

	if help {
		return []OutputEvent{{
			SessionID: session.ID,
			Message:   "Quit the game.\nUsage: /quit",
		}}
	}

	return []OutputEvent{{
		SessionID: session.ID,
		Message:   "Goodbye!",
		Quit: 		true,
	}}
}