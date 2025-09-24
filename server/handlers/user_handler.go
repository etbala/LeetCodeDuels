package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"leetcodeduels/auth"
	"leetcodeduels/services"
	"leetcodeduels/store"
	"leetcodeduels/util"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func GetProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	profile, err := store.DataStore.GetUserProfile(userID)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if profile == nil {
		http.Error(w, "User Not Found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

func UserInGame(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	inGame, err := services.GameManager.IsPlayerInGame(userID)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inGame)
}

func MyProfile(w http.ResponseWriter, r *http.Request) {
	tokenString, err := auth.ExtractTokenString(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	claims, err := auth.ValidateJWT(tokenString)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	profile, err := store.DataStore.GetUserProfile(claims.UserID)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	tokenString, err := auth.ExtractTokenString(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	claims, err := auth.ValidateJWT(tokenString)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = store.DataStore.DeleteUser(claims.UserID)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
}

func GenerateUniqueDiscriminator(username string) (string, error) {
	maxRetries := 10
	for range maxRetries {
		num := util.RandInt(1, 9999)
		discriminator := fmt.Sprintf("%04d", num)

		exists, err := store.DataStore.DiscriminatorExists(username, discriminator)
		if err != nil {
			// Database error
			return "", err
		}

		if !exists {
			// Found a unique one
			return discriminator, nil
		}
	}

	// todo: figure out what to do if we can't find a unique one after several tries
	return "", errors.New("could not generate a unique discriminator for provided username")
}

func RenameUser(w http.ResponseWriter, r *http.Request) {
	var req RenameRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tokenString, err := auth.ExtractTokenString(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	claims, err := auth.ValidateJWT(tokenString)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	discrinimator, err := GenerateUniqueDiscriminator(req.NewUsername)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	err = store.DataStore.UpdateUsernameDiscriminator(claims.UserID, req.NewUsername, discrinimator)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func RenameLCUser(w http.ResponseWriter, r *http.Request) {
	var req RenameRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tokenString, err := auth.ExtractTokenString(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	claims, err := auth.ValidateJWT(tokenString)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = store.DataStore.UpdateLCUsername(claims.UserID, req.NewUsername)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
}
