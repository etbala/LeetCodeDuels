package database

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
