package interfaces

/*
This package defines global interfaces, which can be used to pass information
between sections of code that use marginally different models.
*/

type Player interface {
	GetID() int64
	GetUsername() string
	GetRating() int
}
