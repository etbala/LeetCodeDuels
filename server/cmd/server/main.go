



/*

	Redis for Session Information:

		Set TTL to expire after 24 hrs

		Key: "match:<match_id>"
		Value: {
			"player_one": 1,
			"player_two": 2,
			"problem_id": 42,
			"player_one_submissions": [],
			"player_two_submissions": [],
		}

		Key: "ws_connections:<match_id>"
		Value: [
			"conn1", "conn2"
		]

		Key: "user_session:<user_id>"
		Value: {
			"match_id": "12345"
		}

	Postgres for long term storage
	- Accounts
	- Problem Metadata
	- Match History (Aborted/Completed Matches)

*/

/* Testing


Unit Tests

    Focus on testing functions and methods for correctness.
    Example:
        Authentication:
            Test JWT creation and validation.
            Mock GitHub OAuth responses.
        Redis/Database Operations:
            Mock Redis/Postgres to ensure queries behave correctly.

Integration Tests

    Test API endpoints with a mocked database and Redis.
    Use httptest to simulate HTTP requests and responses.
    Example:
        Match Creation:
            Simulate a user creating an invitation.
            Ensure the match is stored in Redis and correctly linked to the user.

WebSocket Testing

    Simulate WebSocket connections and message exchanges.
    Example:
        Test that two users connected to the same match receive real-time updates when one submits code.
    Use libraries like nhooyr/websocket or native WebSocket clients for testing.

End-to-End Tests

    Simulate real user workflows from authentication to match completion.
    Example:
        User logs in via GitHub OAuth.
        User invites another user to a duel.
        Match progresses with real-time updates.
        Match completes and results are persisted in the database.

*/
