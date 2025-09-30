package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"leetcodeduels/models"
	"leetcodeduels/services"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealth(t *testing.T) {
	res, err := http.Get(ts.URL + "/api/v1/health")
	assert.NoError(t, err)
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestSearchUsers(t *testing.T) {
	token, err := services.GenerateJWT(12345) // Alice's token
	assert.NoError(t, err)

	t.Run("Search with username and discriminator - success", func(t *testing.T) {
		url := fmt.Sprintf("%s/api/v1/users?username=alice&discriminator=0001", ts.URL)
		req, err := http.NewRequest("GET", url, nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		res, err := ts.Client().Do(req)
		assert.NoError(t, err)
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)

		var users []models.UserInfoResponse
		err = json.NewDecoder(res.Body).Decode(&users)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(users), "Expected exactly one user for an exact match")
		assert.Equal(t, "alice", users[0].Username)
		assert.Equal(t, "0001", users[0].Discriminator)
	})

	t.Run("Search with username and discriminator - not found", func(t *testing.T) {
		url := fmt.Sprintf("%s/api/v1/users?username=alice&discriminator=6666", ts.URL)
		req, err := http.NewRequest("GET", url, nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		res, err := ts.Client().Do(req)
		assert.NoError(t, err)
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)

		var users []models.UserInfoResponse
		err = json.NewDecoder(res.Body).Decode(&users)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(users), "Expected an empty array for a user not found")
	})

	t.Run("Search with prefix matching multiple users", func(t *testing.T) {
		url := fmt.Sprintf("%s/api/v1/users?username=sam", ts.URL)
		req, err := http.NewRequest("GET", url, nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		res, err := ts.Client().Do(req)
		assert.NoError(t, err)
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		var users []models.UserInfoResponse
		err = json.NewDecoder(res.Body).Decode(&users)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(users), "Expected exactly two users for the prefix 'sam'")

		expectedUsernames := map[string]bool{
			"samuel":   true,
			"samantha": true,
		}
		foundUsernames := make(map[string]bool)
		for _, u := range users {
			foundUsernames[u.Username] = true
		}
		assert.Equal(t, expectedUsernames, foundUsernames, "The returned users did not match the expected set")
	})

	t.Run("Search with prefix and limit", func(t *testing.T) {
		url := fmt.Sprintf("%s/api/v1/users?username=sam&limit=1", ts.URL)
		req, err := http.NewRequest("GET", url, nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		res, err := ts.Client().Do(req)
		assert.NoError(t, err)
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		var users []models.UserInfoResponse
		err = json.NewDecoder(res.Body).Decode(&users)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(users), "Expected exactly one user when limit = 1")
	})

	t.Run("Fail with no username", func(t *testing.T) {
		url := fmt.Sprintf("%s/api/v1/users", ts.URL)
		req, err := http.NewRequest("GET", url, nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		res, err := ts.Client().Do(req)
		assert.NoError(t, err)
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("Fail with invalid limit", func(t *testing.T) {
		url := fmt.Sprintf("%s/api/v1/users?username=a&limit=99", ts.URL)
		req, err := http.NewRequest("GET", url, nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		res, err := ts.Client().Do(req)
		assert.NoError(t, err)
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})
}

func TestGetProfile(t *testing.T) {
	token, err := services.GenerateJWT(12345) // Alice
	assert.NoError(t, err)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/users/%d", ts.URL, 12345), nil)
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

	var user models.UserInfoResponse
	err = json.NewDecoder(res.Body).Decode(&user)
	assert.NoError(t, err)
	assert.Equal(t, int64(12345), user.ID)
	assert.Equal(t, "alice", user.Username)
	assert.Equal(t, "0001", user.Discriminator)
	assert.Equal(t, "alice_lc", user.LCUsername)
	assert.Equal(t, 1000, user.Rating)
}

func TestMyProfile(t *testing.T) {
	token, err := services.GenerateJWT(12345) // Alice
	assert.NoError(t, err)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/users/me", ts.URL), nil)
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

	var user models.UserInfoResponse
	err = json.NewDecoder(res.Body).Decode(&user)
	assert.NoError(t, err)
	assert.Equal(t, int64(12345), user.ID)
	assert.Equal(t, "alice", user.Username)
	assert.Equal(t, "0001", user.Discriminator)
	assert.Equal(t, "alice_lc", user.LCUsername)
	assert.Equal(t, 1000, user.Rating)
}

func TestDeleteProfile(t *testing.T) {
	token, err := services.GenerateJWT(61539) // Zoe
	assert.NoError(t, err)

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/users/me", ts.URL), nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := ts.Client().Do(req)
	assert.NoError(t, err)
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	token, err = services.GenerateJWT(12345) // Alice
	assert.NoError(t, err)

	req2, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/users/%d", ts.URL, 61539), nil)
	assert.NoError(t, err)
	req2.Header.Set("Authorization", "Bearer "+token)

	res2, err := ts.Client().Do(req2)
	assert.NoError(t, err)
	defer res2.Body.Close()
	assert.Equal(t, http.StatusNotFound, res2.StatusCode)
}

func TestUpdateUser(t *testing.T) {
	// Helper function to perform the verification GET request
	verifyUser := func(t *testing.T, userID int64, authToken string) models.UserInfoResponse {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/users/%d", ts.URL, userID), nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+authToken)

		res, err := ts.Client().Do(req)
		assert.NoError(t, err)
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)

		var user models.UserInfoResponse
		err = json.NewDecoder(res.Body).Decode(&user)
		assert.NoError(t, err)
		return user
	}

	t.Run("Update username only", func(t *testing.T) {
		userID := int64(41529) // Yash
		token, err := services.GenerateJWT(userID)
		assert.NoError(t, err)

		originalUser := verifyUser(t, userID, token)

		newUsername := "yash_the_great"
		payload, _ := json.Marshal(models.UpdateUserRequest{Username: newUsername})
		req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/api/v1/users/me", ts.URL), bytes.NewReader(payload))
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		res, err := ts.Client().Do(req)
		assert.NoError(t, err)
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)

		updatedUser := verifyUser(t, userID, token)
		assert.Equal(t, newUsername, updatedUser.Username)
		assert.NotEmpty(t, updatedUser.Discriminator)
		assert.Equal(t, originalUser.LCUsername, updatedUser.LCUsername)
	})

	t.Run("Update LeetCode username only", func(t *testing.T) {
		userID := int64(53468) // Xavier
		token, err := services.GenerateJWT(userID)
		assert.NoError(t, err)

		originalUser := verifyUser(t, userID, token)

		newLCUsername := "xavier_codes_alot"
		payload, _ := json.Marshal(models.UpdateUserRequest{LeetCodeUsername: newLCUsername})
		req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/api/v1/users/me", ts.URL), bytes.NewReader(payload))
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		res, err := ts.Client().Do(req)
		assert.NoError(t, err)
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)

		updatedUser := verifyUser(t, userID, token)
		assert.Equal(t, newLCUsername, updatedUser.LCUsername)
		assert.Equal(t, originalUser.Username, updatedUser.Username)
	})

	t.Run("Update both usernames", func(t *testing.T) {
		userID := int64(49876) // Willow
		token, err := services.GenerateJWT(userID)
		assert.NoError(t, err)

		newUsername := "new_dual_name"
		newLCUsername := "new_dual_lc_name"
		payload, _ := json.Marshal(models.UpdateUserRequest{Username: newUsername, LeetCodeUsername: newLCUsername})
		req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/api/v1/users/me", ts.URL), bytes.NewReader(payload))
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		res, err := ts.Client().Do(req)
		assert.NoError(t, err)
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)

		updatedUser := verifyUser(t, userID, token)
		assert.Equal(t, newUsername, updatedUser.Username)
		assert.Equal(t, newLCUsername, updatedUser.LCUsername)
	})

	t.Run("Not modified status with no update fields", func(t *testing.T) {
		userID := int64(41529) // Yash
		token, err := services.GenerateJWT(userID)
		assert.NoError(t, err)

		payload, _ := json.Marshal(models.UpdateUserRequest{})
		req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/api/v1/users/me", ts.URL), bytes.NewReader(payload))
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		res, err := ts.Client().Do(req)
		assert.NoError(t, err)
		defer res.Body.Close()

		assert.Equal(t, http.StatusNotModified, res.StatusCode)
	})
}

func TestUserStatusOffline(t *testing.T) {
	token, err := services.GenerateJWT(12345) // Alice
	assert.NoError(t, err)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/users/%d/status", ts.URL, 12345), nil)
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

	var status models.UserStatusResponse
	err = json.NewDecoder(res.Body).Decode(&status)
	assert.NoError(t, err)
	assert.Equal(t, false, status.InGame)
	assert.Equal(t, false, status.Online)
}

func TestMatchHistory(t *testing.T) {
	token, err := services.GenerateJWT(12345) // Alice
	assert.NoError(t, err)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/users/%d/matches", ts.URL, 12345), nil)
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
	token, err := services.GenerateJWT(12345) // Alice
	assert.NoError(t, err)

	matchID := "00000000-0000-0000-0000-000000000000"
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/matches/%s", ts.URL, matchID), nil)
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
	token, err := services.GenerateJWT(12345) // Alice
	assert.NoError(t, err)

	matchID := "00000000-0000-0000-0000-000000000000"
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/matches/%s/submissions", ts.URL, matchID), nil)
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
	res, err := http.Get(ts.URL + "/api/v1/problems/tags")
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
