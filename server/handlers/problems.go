package handlers

import (
	"encoding/json"
	"fmt"
	"leetcodeduels/models"
	"leetcodeduels/store"
	"net/http"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func RandomProblem(w http.ResponseWriter, r *http.Request) {
	l := log.Ctx(r.Context())

	queryParams := r.URL.Query()
	rawTags := queryParams["tag[]"]
	rawDifficulties := queryParams["difficulty[]"]

	l.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Strs("tag[]", rawTags).Strs("difficulty[]", rawDifficulties)
	})
	l.Info().Msg("Received request for RandomProblem")

	tagIDs, err := parseTags(rawTags)
	if err != nil {
		l.Warn().Err(err).Msg("Invalid tag(s) provided")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	difficulties, err := parseDifficulties(rawDifficulties)
	if err != nil {
		l.Warn().Err(err).Msg("Invalid difficulty(s) provided")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	problem, err := store.DataStore.GetRandomProblemByTagsAndDifficulties(tagIDs, difficulties)
	if err != nil {
		l.Error().Err(err).Msg("Could not retrieve problem")
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	if problem == nil {
		l.Warn().Msg("No problem was found matching specifications")
		http.Error(w, "No problem found matching specifications.", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(problem)
}

func AllTags(w http.ResponseWriter, r *http.Request) {
	l := log.Ctx(r.Context())

	l.Info().Msg("Received request for AllTags")

	tags, err := store.DataStore.GetAllTags()
	if err != nil {
		l.Error().Err(err).Msg("Could not retrieve tags")
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tags)
}

func parseTags(tagStrings []string) ([]int, error) {
	var tagIDs []int
	for _, tagStr := range tagStrings {
		id, err := strconv.Atoi(tagStr)
		if err != nil {
			return nil, fmt.Errorf("invalid tag ID: '%s' is not a number", tagStr)
		}
		tagIDs = append(tagIDs, id)
	}
	return tagIDs, nil
}

func parseDifficulties(difficultyStrings []string) ([]models.Difficulty, error) {
	var difficulties []models.Difficulty
	for _, diffStr := range difficultyStrings {
		difficulty, err := models.ParseDifficulty(diffStr)
		if err != nil {
			return nil, fmt.Errorf("invalid difficulty value: '%s'", diffStr)
		}
		difficulties = append(difficulties, difficulty)
	}
	return difficulties, nil
}
