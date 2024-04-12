package auth

import (
    "leetcodeduels/pkg/models"
    "net/http"

    "golang.org/x/crypto/bcrypt"
)

func (s *Store) CreateUser(w http.ResponseWriter, user models.User) error {
    // Hash the password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
    if err != nil {
        return err
    }

    // Insert the user into the database
    _, err = s.db.Exec("INSERT INTO users (username, password_hash, email, rating) VALUES ($1, $2, $3, $4)",
        user.Username, string(hashedPassword), user.Email, user.Rating)
    if err != nil {
        return err
    }

    // Set cookie to indicate successful sign-up
    http.SetCookie(w, &http.Cookie{
        Name:  "loggedIn",
        Value: "true",
        Path:  "/", // Cookie is valid for all paths
    })

    return nil
}

func (s *Store) AuthenticateUser(username, password string) (bool, error) {
	var hashedPassword string
	err := s.db.QueryRow("SELECT password_hash FROM users WHERE username = $1", username).Scan(&hashedPassword)
	if err != nil {
		// User not found or other error
		return false, err
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		// Password does not match
		return false, nil
	}

	// Authentication successful
	return true, nil
}
