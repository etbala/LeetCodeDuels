package handlers

import (
	"leetcodeduels/models"
	"leetcodeduels/services"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func CanSendInvite(w http.ResponseWriter, r *http.Request) {
	l := log.Ctx(r.Context())

	claims, err := services.GetClaimsFromRequest(r)
	if err != nil {
		l.Warn().Msg("Attempted to call CanSendInvite without valid claims")
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	l.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Int64("user_id", claims.UserID)
	})
	l.Info().Msg("Received request for CanSendInvite")

	invite, err := services.InviteManager.InviteDetails(claims.UserID)
	if err != nil {
		l.Error().Err(err).Msg("Failed to check invite status")
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	canSend := invite == nil
	writeSuccess(w, canSend)
}

func SentInvites(w http.ResponseWriter, r *http.Request) {
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
	l.Info().Msg("Received request for SentInvites")

	invite, err := services.InviteManager.InviteDetails(claims.UserID)
	if err != nil {
		l.Error().Err(err).Msg("Failed to get sent invites")
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if invite == nil {
		writeSuccess(w, []models.Invite{})
		return
	}

	var invites []models.Invite
	invites = append(invites, *invite)
	writeSuccess(w, invites)
}

func ReceivedInvites(w http.ResponseWriter, r *http.Request) {
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
	l.Info().Msg("Received request for ReceivedInvites")

	invites, err := services.InviteManager.GetPendingInvites(claims.UserID)
	if err != nil {
		l.Error().Err(err).Msg("Failed to get received invites")
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	writeSuccess(w, invites)
}
