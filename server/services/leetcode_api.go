package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const LC_API_URL = "https://leetcode.com/graphql"

type graphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type AcceptedSubmission struct {
	SubmissionID int64
	Timestamp    time.Time
	TitleSlug    string
}

type apiAcSubmission struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	TitleSlug string `json:"titleSlug"`
}

type apiAcSubmissionListData struct {
	Submissions []apiAcSubmission `json:"recentAcSubmissionList"`
}

type graphQLAcSubmissionListResponse struct {
	Data apiAcSubmissionListData `json:"data"`
}

// Fetches the most recent accepted submission for a given user
func GetLastAcceptedSubmission(leetcodeUsername string) (*AcceptedSubmission, error) {
	if leetcodeUsername == "" {
		return nil, fmt.Errorf("leetcodeUsername cannot be empty")
	}

	query := `
        query recentAcSubmissionList($username: String!, $limit: Int!) {
            recentAcSubmissionList(username: $username, limit: $limit) {
                id
                timestamp
                titleSlug
            }
        }`

	requestPayload := graphQLRequest{
		Query: query,
		Variables: map[string]interface{}{
			"username": leetcodeUsername,
			"limit":    1,
		},
	}

	var apiResp graphQLAcSubmissionListResponse
	err := makeGraphQLRequest(requestPayload, &apiResp)
	if err != nil {
		return nil, fmt.Errorf("error making public API request: %w", err)
	}

	if len(apiResp.Data.Submissions) == 0 {
		return nil, nil
	}

	mostRecentSub := apiResp.Data.Submissions[0]

	submissionID, err := strconv.ParseInt(mostRecentSub.ID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing submission ID '%s': %w", mostRecentSub.ID, err)
	}

	timestamp, err := strconv.ParseInt(mostRecentSub.Timestamp, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing timestamp '%s': %w", mostRecentSub.Timestamp, err)
	}

	result := &AcceptedSubmission{
		SubmissionID: submissionID,
		Timestamp:    time.Unix(timestamp, 0),
		TitleSlug:    mostRecentSub.TitleSlug,
	}

	return result, nil
}

func makeGraphQLRequest(payload interface{}, responseTarget interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshalling graphql request: %w", err)
	}

	req, err := http.NewRequest("POST", LC_API_URL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("error creating http request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36")
	req.Header.Set("Referer", "https://leetcode.com/")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned non-200 status: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(responseTarget); err != nil {
		return fmt.Errorf("error decoding response: %w", err)
	}

	return nil
}
