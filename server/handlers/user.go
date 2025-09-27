package handlers

import (
	"encoding/json"
	"fmt"
	"leetcodeduels/models"
	"leetcodeduels/services"
	"leetcodeduels/store"
	"leetcodeduels/ws"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func SearchUsers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	username := query.Get("username")
	discriminator := query.Get("discriminator")
	limitStr := query.Get("limit")

	if username == "" {
		http.Error(w, "Username parameter is required", http.StatusBadRequest)
		return
	}

	// If username and discriminator are both provided, only return exact match
	if discriminator != "" {
		user, err := store.DataStore.GetUserProfileByUsername(username, discriminator)
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		var res []models.UserInfoResponse
		if user != nil {
			res = append(res, models.UserInfoResponse{
				ID:            user.ID,
				Username:      user.Username,
				Discriminator: user.Discriminator,
				LCUsername:    user.LeetCodeUsername,
				AvatarURL:     user.AvatarURL,
				Rating:        user.Rating,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
		return
	}

	limit := 5
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 20 {
			http.Error(w, "Invalid limit parameter. Must be between 1 and 20 (inclusive).", http.StatusBadRequest)
			return
		}
	}

	users, err := store.DataStore.SearchUsersByUsername(username, limit)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

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

func UserStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	online, err := ws.ConnManager.IsUserOnline(userID)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	inGame, err := services.GameManager.IsPlayerInGame(userID)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	var res models.UserStatusResponse = models.UserStatusResponse{
		Online: online,
		InGame: inGame,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func MyProfile(w http.ResponseWriter, r *http.Request) {
	claims, err := services.GetClaimsFromRequest(r)
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
	claims, err := services.GetClaimsFromRequest(r)
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

// UpdateUser handles partial updates to the authenticated user's profile.
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	claims, err := services.GetClaimsFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" && req.LeetCodeUsername == "" {
		http.Error(w, "No update fields provided", http.StatusBadRequest)
		return
	}

	var discriminator string
	if req.Username != "" {
		discriminator, err = services.GenerateUniqueDiscriminator(req.Username)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	err = store.DataStore.UpdateUser(claims.UserID, req.Username, discriminator, req.LeetCodeUsername)
	if err != nil {
		http.Error(w, "Unknown Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func UserMatches(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	sessions, err := store.DataStore.GetPlayerMatches(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

func MyNotifications(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not Implemented", http.StatusNotImplemented)
}
