package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"leetcodeduels/auth"
	"leetcodeduels/models"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealth(t *testing.T) {
	res, err := http.Get(ts.URL + "/health")
	assert.NoError(t, err)
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestGetProfile(t *testing.T) {
	token, err := auth.GenerateJWT(12345) // Alice
	assert.NoError(t, err)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/user/profile/%d", ts.URL, 12345), nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := ts.Client().Do(req)
	assert.NoError(t, err)
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected 200 OK, got %d\nbody: %s", res.StatusCode, string(body))
	}

	var user models.User
	err = json.NewDecoder(res.Body).Decode(&user)
	assert.NoError(t, err)
	assert.Equal(t, int64(12345), user.ID)
	assert.Equal(t, "alice", user.Username)
	assert.Equal(t, "alice_lc", user.LeetCodeUsername)
	assert.Equal(t, 1000, user.Rating)
}

func TestMyProfile(t *testing.T) {
	token, err := auth.GenerateJWT(12345) // Alice
	assert.NoError(t, err)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/user/me", ts.URL), nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := ts.Client().Do(req)
	assert.NoError(t, err)
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected 200 OK, got %d\nbody: %s", res.StatusCode, string(body))
	}

	var user models.User
	err = json.NewDecoder(res.Body).Decode(&user)
	assert.NoError(t, err)
	assert.Equal(t, int64(12345), user.ID)
	assert.Equal(t, "alice", user.Username)
	assert.Equal(t, "alice_lc", user.LeetCodeUsername)
	assert.Equal(t, 1000, user.Rating)
}

func TestDeleteProfile(t *testing.T) {
	token, err := auth.GenerateJWT(61539) // Zoe
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/user/me/delete", ts.URL), nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := ts.Client().Do(req)
	assert.NoError(t, err)
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Now check profile was deleted (using valid user)
	token, err = auth.GenerateJWT(12345) // Alice
	assert.NoError(t, err)

	req2, err := http.NewRequest("GET", fmt.Sprintf("%s/user/profile/%d", ts.URL, 61539), nil)
	assert.NoError(t, err)
	req2.Header.Set("Authorization", "Bearer "+token)

	res2, err := ts.Client().Do(req2)
	assert.NoError(t, err)
	defer res2.Body.Close()
	assert.Equal(t, http.StatusNotFound, res2.StatusCode)
}

func TestRenameUser(t *testing.T) {
	token, err := auth.GenerateJWT(41529) // Yash
	assert.NoError(t, err)

	newName := "yash2"
	msg := models.RenameRequest{NewUsername: newName}
	payload, _ := json.Marshal(msg)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/user/me/rename", ts.URL), bytes.NewReader(payload))
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	res, err := ts.Client().Do(req)
	assert.NoError(t, err)
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Now check profile was renamed
	req2, err := http.NewRequest("GET", fmt.Sprintf("%s/user/profile/%d", ts.URL, 41529), nil)
	assert.NoError(t, err)
	req2.Header.Set("Authorization", "Bearer "+token)

	res2, err := ts.Client().Do(req2)
	assert.NoError(t, err)
	defer res2.Body.Close()
	assert.Equal(t, http.StatusOK, res2.StatusCode)

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected 200 OK, got %d\nbody: %s", res.StatusCode, string(body))
	}

	var user models.User
	err = json.NewDecoder(res2.Body).Decode(&user)
	assert.NoError(t, err)
	assert.Equal(t, newName, user.Username)
}

func TestLCRenameUser(t *testing.T) {
	token, err := auth.GenerateJWT(53468) // Xavier
	assert.NoError(t, err)

	newLCName := "xavier2_lc"
	msg := models.RenameRequest{NewUsername: newLCName}
	payload, _ := json.Marshal(msg)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/user/me/lcrename", ts.URL), bytes.NewReader(payload))
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	res, err := ts.Client().Do(req)
	assert.NoError(t, err)
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Now check profile was renamed
	req2, err := http.NewRequest("GET", fmt.Sprintf("%s/user/profile/%d", ts.URL, 53468), nil)
	assert.NoError(t, err)
	req2.Header.Set("Authorization", "Bearer "+token)

	res2, err := ts.Client().Do(req2)
	assert.NoError(t, err)
	defer res2.Body.Close()
	assert.Equal(t, http.StatusOK, res2.StatusCode)

	var user models.User
	err = json.NewDecoder(res2.Body).Decode(&user)
	assert.NoError(t, err)
	assert.Equal(t, newLCName, user.LeetCodeUsername)
}

func TestUserNotInGame(t *testing.T) {
	token, err := auth.GenerateJWT(12345) // Alice
	assert.NoError(t, err)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/user/in-game/%d", ts.URL, 12345), nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := ts.Client().Do(req)
	assert.NoError(t, err)
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected 200 OK, got %d\nbody: %s", res.StatusCode, string(body))
	}

	var inGame bool
	err = json.NewDecoder(res.Body).Decode(&inGame)
	assert.NoError(t, err)
	assert.Equal(t, false, inGame)
}

func TestMatchHistory(t *testing.T) {
	token, err := auth.GenerateJWT(12345) // Alice
	assert.NoError(t, err)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/game/history/%d", ts.URL, 12345), nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := ts.Client().Do(req)
	assert.NoError(t, err)
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected 200 OK, got %d\nbody: %s", res.StatusCode, string(body))
	}

	var matches []models.Session
	err = json.NewDecoder(res.Body).Decode(&matches)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(matches))

	assert.Equal(t, "00000000-0000-0000-0000-000000000000", matches[0].ID)
	assert.Equal(t, models.MatchWon, matches[0].Status)
	assert.Equal(t, false, matches[0].IsRated)
	assert.Equal(t, 1, matches[0].Problem.ID)
	assert.Equal(t, 2, len(matches[0].Players))
	assert.Equal(t, int64(12345), matches[0].Winner)

	assert.Equal(t, "00000000-0000-0000-0000-000000000001", matches[1].ID)
	assert.Equal(t, models.MatchWon, matches[1].Status)
	assert.Equal(t, false, matches[1].IsRated)
	assert.Equal(t, 2, matches[1].Problem.ID)
	assert.Equal(t, 2, len(matches[1].Players))
	assert.Equal(t, int64(87902), matches[1].Winner)

	assert.Equal(t, "00000000-0000-0000-0000-000000000002", matches[2].ID)
	assert.Equal(t, models.MatchWon, matches[2].Status)
	assert.Equal(t, false, matches[2].IsRated)
	assert.Equal(t, 4, matches[2].Problem.ID)
	assert.Equal(t, 2, len(matches[2].Players))
	assert.Equal(t, int64(12345), matches[2].Winner)
}

func TestGetMatch(t *testing.T) {
	token, err := auth.GenerateJWT(12345) // Alice
	assert.NoError(t, err)

	matchID := "00000000-0000-0000-0000-000000000000"
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/game/details/%s", ts.URL, matchID), nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := ts.Client().Do(req)
	assert.NoError(t, err)
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected 200 OK, got %d\nbody: %s", res.StatusCode, string(body))
	}

	var match models.Session
	err = json.NewDecoder(res.Body).Decode(&match)
	assert.NoError(t, err)
	assert.Equal(t, matchID, match.ID)
	assert.Equal(t, models.MatchWon, match.Status)
	assert.Equal(t, false, match.IsRated)
	assert.Equal(t, 1, match.Problem.ID)
	assert.Equal(t, "Two Sum", match.Problem.Name)
	assert.Equal(t, "two-sum", match.Problem.Slug)
	assert.Equal(t, models.Easy, match.Problem.Difficulty)
	assert.Equal(t, 2, len(match.Players))
	assert.Equal(t, 1, len(match.Submissions))
	assert.Equal(t, int64(12345), match.Winner)
}

func TestGetMatchSubmissions(t *testing.T) {
	token, err := auth.GenerateJWT(12345) // Alice
	assert.NoError(t, err)

	matchID := "00000000-0000-0000-0000-000000000000"
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/game/submissions/%s", ts.URL, matchID), nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := ts.Client().Do(req)
	assert.NoError(t, err)
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected 200 OK, got %d\nbody: %s", res.StatusCode, string(body))
	}

	var submissions []models.PlayerSubmission
	err = json.NewDecoder(res.Body).Decode(&submissions)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(submissions))
	assert.Equal(t, 0, submissions[0].ID)
	assert.Equal(t, int64(12345), submissions[0].PlayerID)
	assert.Equal(t, models.Accepted, submissions[0].Status)
	assert.Equal(t, models.Cpp, submissions[0].Lang)
}

func TestAllTags(t *testing.T) {
	res, err := http.Get(ts.URL + "/problems/tags")
	assert.NoError(t, err)
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var tags []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	err = json.NewDecoder(res.Body).Decode(&tags)
	assert.NoError(t, err)
	// we seeded 10 tags
	assert.True(t, len(tags) >= 10, fmt.Sprintf("expected at least 10 tags, got %d", len(tags)))

	// ensure a known tag is present
	found := false
	for _, tag := range tags {
		if tag.Name == "array" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected tag 'array' not found")
}
