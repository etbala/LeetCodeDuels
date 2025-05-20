package handlers

import (
	"leetcodeduels/auth"
	"net/http"
)

func WSConnect(w http.ResponseWriter, r *http.Request) {
	tokenString, err := auth.ExtractTokenString(r)
	if err != nil {

	}

	claims, err := auth.ValidateJWT(tokenString)
	if err != nil {

	}

	userID := claims.UserID

}
