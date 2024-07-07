package game

type command func(*Game, *Session, string, bool) []OutputEvent