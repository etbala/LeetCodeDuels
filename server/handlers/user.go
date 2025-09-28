package handlers

import (
	"encoding/json"
	"leetcodeduels/models"
	"leetcodeduels/services"
	"leetcodeduels/store"
	"leetcodeduels/ws"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func SearchUsers(w http.ResponseWriter, r *http.Request) {
	l := log.Ctx(r.Context())

	query := r.URL.Query()
	username := query.Get("username")
	discriminator := query.Get("discriminator")
	limitStr := query.Get("limit")

	l.Info().
		Str("username", username).
		Str("discriminator", discriminator).
		Str("limit", limitStr).
		Msg("Received request for SearchUsers")

	if username == "" {
		l.Warn().Msg("SearchUsers called without username parameter")
		http.Error(w, "Username parameter is required", http.StatusBadRequest)
		return
	}

	// If username and discriminator are both provided, only return exact match
	if discriminator != "" {
		user, err := store.DataStore.GetUserProfileByUsername(username, discriminator)
		if err != nil {
			l.Error().Err(err).Msg("SearchUsers Failed to get user profile by username")
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
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
		return
	}

	limit := 5
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 20 {
			l.Warn().Msg("SearchUsers called with invalid limit")
			http.Error(w, "Invalid limit parameter. Must be between 1 and 20 (inclusive).", http.StatusBadRequest)
			return
		}
	}

	users, err := store.DataStore.SearchUsersByUsername(username, limit)
	if err != nil {
		l.Error().Err(err).Msg("Failed to search users by username")
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func GetProfile(w http.ResponseWriter, r *http.Request) {
	l := log.Ctx(r.Context())

	vars := mux.Vars(r)
	userIDStr := vars["id"]

	l.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("user_id", userIDStr)
	})
	l.Info().Msg("Received request for GetProfile")

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		l.Warn().Msg("Invalid user ID format in path parameter")
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	profile, err := store.DataStore.GetUserProfile(userID)
	if err != nil {
		l.Error().Err(err).Msg("Error fetching user profile")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if profile == nil {
		l.Info().Msg("User profile not found")
		http.Error(w, "User Not Found", http.StatusNotFound)
		return
	}

	var res models.UserInfoResponse = models.UserInfoResponse{
		ID:            profile.ID,
		Username:      profile.Username,
		Discriminator: profile.Discriminator,
		LCUsername:    profile.LeetCodeUsername,
		AvatarURL:     profile.AvatarURL,
		Rating:        profile.Rating,
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func UserStatus(w http.ResponseWriter, r *http.Request) {
	l := log.Ctx(r.Context())

	vars := mux.Vars(r)
	userIDStr := vars["id"]

	l.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("user_id", userIDStr)
	})
	l.Info().Msg("Received request for UserStatus")

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		l.Warn().Msg("Invalid user ID format in path parameter")
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	online, err := ws.ConnManager.IsUserOnline(userID)
	if err != nil {
		l.Error().Err(err).Msg("Error checking if user is online")
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	inGame, err := services.GameManager.IsPlayerInGame(userID)
	if err != nil {
		l.Error().Err(err).Msg("Error checking if user is in-game")
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	var res models.UserStatusResponse = models.UserStatusResponse{
		Online: online,
		InGame: inGame,
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func MyProfile(w http.ResponseWriter, r *http.Request) {
	l := log.Ctx(r.Context())

	claims, err := services.GetClaimsFromRequest(r)
	if err != nil {
		l.Warn().Msg("Attempted to call MyProfile without valid claims")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	l.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Int64("user_id", claims.UserID)
	})
	l.Info().Msg("Received request for MyProfile")

	profile, err := store.DataStore.GetUserProfile(claims.UserID)
	if err != nil {
		l.Error().Err(err).Msg("Error getting user profile")
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	if profile == nil {
		// Should not happen unless user was deleted before jwt expired
		l.Warn().Msg("Attempted to view profile of user that does not exist")
		http.Error(w, "User Not Found", http.StatusNotFound)
		return
	}

	var res models.UserInfoResponse = models.UserInfoResponse{
		ID:            profile.ID,
		Username:      profile.Username,
		Discriminator: profile.Discriminator,
		LCUsername:    profile.LeetCodeUsername,
		AvatarURL:     profile.AvatarURL,
		Rating:        profile.Rating,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	l := log.Ctx(r.Context())

	claims, err := services.GetClaimsFromRequest(r)
	if err != nil {
		l.Warn().Msg("Attempted to call DeleteUser without valid claims")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	l.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Int64("user_id", claims.UserID)
	})
	l.Info().Msg("Received request for DeleteUser")

	err = store.DataStore.DeleteUser(claims.UserID)
	if err != nil {
		l.Error().Err(err).Msg("Failed to delete user from datastore")
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// UpdateUser handles partial updates to the authenticated user's profile.
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	l := log.Ctx(r.Context())

	claims, err := services.GetClaimsFromRequest(r)
	if err != nil {
		l.Warn().Msg("Attempted to call UpdateUser without valid claims")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	l.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Int64("user_id", claims.UserID)
	})
	l.Info().Msg("Received request for UpdateUser")

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		l.Warn().Err(err).Msg("Failed to decode update user request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" && req.LeetCodeUsername == "" {
		l.Warn().Msg("Request had no update fields")
		http.Error(w, "No update fields provided", http.StatusNotModified)
		return
	}

	l.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("new_username", req.Username).Str("new_lc_username", req.LeetCodeUsername)
	})

	l.Info().Msg("User requested profile update")

	var discriminator string
	if req.Username != "" {
		discriminator, err = services.GenerateUniqueDiscriminator(req.Username)
		if err != nil {
			l.Error().Err(err).Msg("Failed to generate unique discriminator")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	err = store.DataStore.UpdateUser(claims.UserID, req.Username, discriminator, req.LeetCodeUsername)
	if err != nil {
		l.Error().Err(err).Msg("Failed to update user in datastore")
		http.Error(w, "Unknown Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func UserMatches(w http.ResponseWriter, r *http.Request) {
	l := log.Ctx(r.Context())

	vars := mux.Vars(r)
	userIDStr := vars["id"]

	l.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("user_id", userIDStr)
	})
	l.Info().Msg("Received request for UserMatches")

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		l.Warn().Msg("Invalid user ID format in path parameter")
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	sessions, err := store.DataStore.GetPlayerMatches(userID)
	if err != nil {
		l.Error().Err(err).Msg("Failed to get match history")
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

func MyNotifications(w http.ResponseWriter, r *http.Request) {
	l := log.Ctx(r.Context())
	l.Warn().Msg("Unimplemented Endpoint Called")
	http.Error(w, "Not Implemented", http.StatusNotImplemented)
}
