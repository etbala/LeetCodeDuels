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

// Sets headers and writes JSON response with status code
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	if data == nil {
		w.WriteHeader(statusCode)
		return nil
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

// Writes an error response with message
func writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// Writes a success response (200 OK)
func writeSuccess(w http.ResponseWriter, data interface{}) error {
	return writeJSON(w, http.StatusOK, data)
}

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
		writeError(w, http.StatusBadRequest, "Username parameter is required")
		return
	}

	// If username and discriminator are both provided, only return exact match
	if discriminator != "" {
		user, err := store.DataStore.GetUserProfileByUsername(username, discriminator)
		if err != nil {
			l.Error().Err(err).Msg("SearchUsers Failed to get user profile by username")
			writeError(w, http.StatusInternalServerError, "Internal Error")
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
		writeSuccess(w, res)
		return
	}

	limit := 5
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 20 {
			l.Warn().Msg("SearchUsers called with invalid limit")
			writeError(w, http.StatusBadRequest, "Invalid limit parameter. Must be between 1 and 20 (inclusive).")
			return
		}
	}

	users, err := store.DataStore.SearchUsersByUsername(username, limit)
	if err != nil {
		l.Error().Err(err).Msg("Failed to search users by username")
		writeError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	writeSuccess(w, users)
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
		writeError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	profile, err := store.DataStore.GetUserProfile(userID)
	if err != nil {
		l.Error().Err(err).Msg("Error fetching user profile")
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if profile == nil {
		l.Info().Msg("User profile not found")
		writeError(w, http.StatusNotFound, "User Not Found")
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

	writeSuccess(w, res)
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
		writeError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	online, err := ws.ConnManager.IsUserOnline(userID)
	if err != nil {
		l.Error().Err(err).Msg("Error checking if user is online")
		writeError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	inGame, err := services.GameManager.IsPlayerInGame(userID)
	if err != nil {
		l.Error().Err(err).Msg("Error checking if user is in-game")
		writeError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	var res models.UserStatusResponse = models.UserStatusResponse{
		Online: online,
		InGame: inGame,
	}

	if inGame {
		sessionID, err := services.GameManager.GetSessionIDByPlayer(userID)
		if err != nil {
			l.Error().Err(err).Msg("Error getting session ID for in-game user")
			writeError(w, http.StatusInternalServerError, "Internal Error")
			return
		}

		res.GameID = sessionID
	}

	writeSuccess(w, res)
}

func MyProfile(w http.ResponseWriter, r *http.Request) {
	l := log.Ctx(r.Context())

	claims, err := services.GetClaimsFromRequest(r)
	if err != nil {
		l.Warn().Msg("Attempted to call MyProfile without valid claims")
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	l.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Int64("user_id", claims.UserID)
	})
	l.Info().Msg("Received request for MyProfile")

	profile, err := store.DataStore.GetUserProfile(claims.UserID)
	if err != nil {
		l.Error().Err(err).Msg("Error getting user profile")
		writeError(w, http.StatusInternalServerError, "Internal Error")
		return
	}
	if profile == nil {
		// Should not happen unless user was deleted before jwt expired
		l.Warn().Msg("Attempted to view profile of user that does not exist")
		writeError(w, http.StatusNotFound, "User Not Found")
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

	writeSuccess(w, res)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	l := log.Ctx(r.Context())

	claims, err := services.GetClaimsFromRequest(r)
	if err != nil {
		l.Warn().Msg("Attempted to call DeleteUser without valid claims")
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	l.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Int64("user_id", claims.UserID)
	})
	l.Info().Msg("Received request for DeleteUser")

	err = store.DataStore.DeleteUser(claims.UserID)
	if err != nil {
		l.Error().Err(err).Msg("Failed to delete user from datastore")
		writeError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	writeSuccess(w, nil)
}

// UpdateUser handles partial updates to the authenticated user's profile.
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	l := log.Ctx(r.Context())

	claims, err := services.GetClaimsFromRequest(r)
	if err != nil {
		l.Warn().Msg("Attempted to call UpdateUser without valid claims")
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	l.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Int64("user_id", claims.UserID)
	})
	l.Info().Msg("Received request for UpdateUser")

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		l.Warn().Err(err).Msg("Failed to decode update user request body")
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Username == "" && req.LeetCodeUsername == "" {
		l.Warn().Msg("Request had no update fields")
		writeError(w, http.StatusNotModified, "No update fields provided")
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
			writeError(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}
	}

	err = store.DataStore.UpdateUser(claims.UserID, req.Username, discriminator, req.LeetCodeUsername)
	if err != nil {
		l.Error().Err(err).Msg("Failed to update user in datastore")
		writeError(w, http.StatusInternalServerError, "Unknown Error")
		return
	}

	writeSuccess(w, nil)
}

func UserMatches(w http.ResponseWriter, r *http.Request) {
	l := log.Ctx(r.Context())

	vars := mux.Vars(r)
	userIDStr := vars["id"]

	query := r.URL.Query()
	pageStr := query.Get("page")
	limitStr := query.Get("limit")

	// page is optional param, defaults to 1
	// limit is optional param, defaults to 10, max 50
	page := 1
	limit := 10
	var err error
	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			l.Warn().Msg("UserMatches called with invalid page")
			writeError(w, http.StatusBadRequest, "Invalid page parameter. Must be a positive integer.")
			return
		}
	}
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 50 {
			l.Warn().Msg("UserMatches called with invalid limit")
			writeError(w, http.StatusBadRequest, "Invalid limit parameter. Must be between 1 and 50 (inclusive).")
			return
		}
	}

	l.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("user_id", userIDStr).Str("page", pageStr).Str("limit", limitStr)
	})
	l.Info().Msg("Received request for UserMatches")

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		l.Warn().Msg("Invalid user ID format in path parameter")
		writeError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	sessions, err := store.DataStore.GetPlayerMatches(userID, page, limit)
	if err != nil {
		l.Error().Err(err).Msg("Failed to get match history")
		writeError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	writeSuccess(w, sessions)
}

func MyNotifications(w http.ResponseWriter, r *http.Request) {
	l := log.Ctx(r.Context())
	l.Warn().Msg("Unimplemented Endpoint Called")
	writeError(w, http.StatusNotImplemented, "Not Implemented")
}
