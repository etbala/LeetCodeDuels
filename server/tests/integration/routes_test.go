package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"leetcodeduels/auth"
	"leetcodeduels/handlers"
	"leetcodeduels/models"
	"net/http"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
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

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/user/%d", ts.URL, 12345), nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := ts.Client().Do(req)
	assert.NoError(t, err)
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

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

	req2, err := http.NewRequest("GET", fmt.Sprintf("%s/user/%d", ts.URL, 61539), nil)
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
	msg := handlers.RenameRequest{NewUsername: newName}
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
	req2, err := http.NewRequest("GET", fmt.Sprintf("%s/user/%d", ts.URL, 41529), nil)
	assert.NoError(t, err)
	req2.Header.Set("Authorization", "Bearer "+token)

	res2, err := ts.Client().Do(req2)
	assert.NoError(t, err)
	defer res2.Body.Close()
	assert.Equal(t, http.StatusOK, res2.StatusCode)

	var user models.User
	err = json.NewDecoder(res2.Body).Decode(&user)
	assert.NoError(t, err)
	assert.Equal(t, newName, user.Username)
}

func TestLCRenameUser(t *testing.T) {
	token, err := auth.GenerateJWT(53468) // Xavier
	assert.NoError(t, err)

	newLCName := "xavier2_lc"
	msg := handlers.RenameRequest{NewUsername: newLCName}
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
	req2, err := http.NewRequest("GET", fmt.Sprintf("%s/user/%d", ts.URL, 53468), nil)
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

}

func TestMatchHistory(t *testing.T) {

}

func TestGetMatch(t *testing.T) {

}

func TestGetMatchSubmissions(t *testing.T) {

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

func TestWSUpgrader(t *testing.T) {
	token, err := auth.GenerateJWT(12345) // Alice
	assert.NoError(t, err)

	// http://127.0.0.1:12345 -> ws://127.0.0.1:12345/ws
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"

	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)

	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, header)
	assert.NoError(t, err, "should upgrade to WebSocket without error")
	defer conn.Close()
	assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)

	// perform a simple ping-pong to verify channel works
	err = conn.WriteMessage(websocket.TextMessage, []byte("ping"))
	assert.NoError(t, err)
}
