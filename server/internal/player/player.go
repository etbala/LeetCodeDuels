package player

/*
This packages defines a global interface for Players, which can be used to
pass players between segments of code that use marginally different Player
models (e.g. Game Session Handler & Matchmaking Service).
*/

type PlayerInterface interface {
	GetID() string
	GetUsername() string
	GetRating() int
}
