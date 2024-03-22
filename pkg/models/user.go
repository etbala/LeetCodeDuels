package models

type User struct {
	UserID   int
	Username string
	Rating   int
	Friends  []string
}

func NewUser(userID int, username string, rating int, friends []string) *User {
	return &User{
		UserID:   userID,
		Username: username,
		Rating:   rating,
		Friends:  friends,
	}
}
