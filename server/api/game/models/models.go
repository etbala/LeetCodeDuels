package models

import "time"

type Player struct {
	ID       int
	Username string
	RoomID   string
}

type Question struct {
	ID         int
	Title      string
	TitleSlug  string
	Difficulty string
	Tags       []string
}

type PlayerSubmission struct {
	Question Question
	Pass     bool
	Time     time.Time
}
