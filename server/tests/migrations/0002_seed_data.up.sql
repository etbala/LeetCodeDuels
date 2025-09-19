-- users
INSERT INTO github_oauth_users (
  github_id, username, lc_username, access_token,
  created_at, updated_at, rating
) VALUES
  (12345, 'alice', 'alice_lc', 'dummy-token', now(), now(), 1000),
  (67890, 'bob', 'bob_lc', 'dummy-token', now(), now(), 1000),
  (87902, 'charlie', 'charlie_lc', 'dummy-token', now(), now(), 1000),
  (20579, 'david', 'david_lc', 'dummy-token', now(), now(), 1000),
  (25074, 'emily', 'emily_lc', 'dummy-token', now(), now(), 1000),
  (43567, 'fiona', 'fiona_lc', 'dummy-token', now(), now(), 1000),
  (56563, 'gavin', 'gavin_lc', 'dummy-token', now(), now(), 1000),
  (81970, 'henry', 'henry_lc', 'dummy-token', now(), now(), 1000),
  (92349, 'isabel', 'isabel_lc', 'dummy-token', now(), now(), 1000),
  (31657, 'juliet', 'juliet_lc', 'dummy-token', now(), now(), 1000),
  (10987, 'katya', 'katya_lc', 'dummy-token', now(), now(), 1000),
  (26354, 'lisa', 'lisa_lc', 'dummy-token', now(), now(), 1000),
  (51796, 'matt', 'matt_lc', 'dummy-token', now(), now(), 1000),
  (73443, 'nancy', 'nancy_lc', 'dummy-token', now(), now(), 1000),
  (43298, 'owen', 'owen_lc', 'dummy-token', now(), now(), 1000),
  (97862, 'philip', 'philip_lc', 'dummy-token', now(), now(), 1000),
  (70763, 'quincy', 'quincy_lc', 'dummy-token', now(), now(), 1000),
  (82352, 'rachel', 'rachel_lc', 'dummy-token', now(), now(), 1000),
  (77356, 'samuel', 'samuel_lc', 'dummy-token', now(), now(), 1000),
  (32189, 'tom', 'tom_lc', 'dummy-token', now(), now(), 1000),
  (62307, 'uri', 'uri_lc', 'dummy-token', now(), now(), 1000),
  (52340, 'victor', 'victor_lc', 'dummy-token', now(), now(), 1000),
  (49876, 'willow', 'willow_lc', 'dummy-token', now(), now(), 1000),
  (53468, 'xavier', 'xavier_lc', 'dummy-token', now(), now(), 1000),
  (41529, 'yash', 'yash_lc', 'dummy-token', now(), now(), 1000),
  (61539, 'zoe', 'zoe_lc', 'dummy-token', now(), now(), 1000)
;

-- problems
INSERT INTO problems (id, problem_num, name, slug, difficulty) VALUES
  (1, 1, 'Two Sum', 'two-sum', 'Easy'),
  (2, 2, 'Add Two Numbers', 'add-two-numbers', 'Medium'),
  (4, 4, 'Median of Two Sorted Arrays', 'median-of-two-sorted-arrays', 'Hard'),
  (5, 5, 'Longest Palindromic Substring', 'longest-palindromic-substring', 'Medium'),
  (7, 7, 'Reverse Integer', 'reverse-integer', 'Medium')
;

-- tags
INSERT INTO tags (id, name) VALUES
  (1, 'array'),
  (2, 'string'),
  (3, 'hash-table'), 
  (5, 'dynamic-programming'),
  (6, 'math'),
  (7, 'binary-search'),
  (8, 'two-pointers'),
  (9, 'linked-list'),
  (10, 'divide-and-conquer'),
  (11, 'recursion')
;

-- problem_tags
INSERT INTO problem_tags (problem_id, tag_id) VALUES
  (1, 1),
  (1, 3),
  (2, 6),
  (2, 9),
  (2, 11),
  (4, 1),
  (4, 7),
  (4, 10),
  (5, 2),
  (5, 5),
  (5, 8),
  (7, 6)
;

-- matches
INSERT INTO matches (
  id, problem_id, is_rated, status, winner_id, start_time, end_time
) VALUES 
  ('00000000-0000-0000-0000-000000000000', 1, false, 'Won',       12345, now(), now()),
  ('00000000-0000-0000-0000-000000000001', 2, false, 'Won',       87902, now(), now()),
  ('00000000-0000-0000-0000-000000000002', 4, false, 'Won',       12345, now(), now()),
  ('00000000-0000-0000-0000-000000000003', 5, false, 'Canceled',  12345, now(), now()),
  ('00000000-0000-0000-0000-000000000004', 7, false, 'Reverted',  12345, now(), now())
;

-- match_players
INSERT INTO match_players (match_id, player_id) VALUES
  ('00000000-0000-0000-0000-000000000000', 12345),
  ('00000000-0000-0000-0000-000000000000', 67890),
  ('00000000-0000-0000-0000-000000000001', 12345),
  ('00000000-0000-0000-0000-000000000001', 87902),
  ('00000000-0000-0000-0000-000000000002', 12345),
  ('00000000-0000-0000-0000-000000000002', 20579),
  ('00000000-0000-0000-0000-000000000003', 67890),
  ('00000000-0000-0000-0000-000000000003', 87902),
  ('00000000-0000-0000-0000-000000000004', 10987),
  ('00000000-0000-0000-0000-000000000004', 52340)
;

-- submissions
INSERT INTO submissions (
  match_id, submission_id, player_id, passed_test_cases, 
  total_test_cases, status, runtime, memory, lang, submitted_at
) VALUES
  ('00000000-0000-0000-0000-000000000000', 0, 12345, 10, 10, 'Accepted', 0, 5000, 'cpp', now()),
  ('00000000-0000-0000-0000-000000000001', 0, 87902, 10, 10, 'Compile Error', 0, 5000, 'java', now()),
  ('00000000-0000-0000-0000-000000000001', 1, 12345, 10, 10, 'Runtime Error', 0, 5000, 'python3', now()),
  ('00000000-0000-0000-0000-000000000001', 2, 87902, 10, 10, 'Accepted', 0, 5000, 'rust', now()),
  ('00000000-0000-0000-0000-000000000002', 0, 20579, 10, 10, 'Memory Limit Exceeded', 0, 5000, 'go', now()),
  ('00000000-0000-0000-0000-000000000002', 1, 12345, 10, 10, 'Time Limit Exceeded', 0, 5000, 'csharp', now()),
  ('00000000-0000-0000-0000-000000000002', 2, 12345, 10, 10, 'Wrong Answer', 0, 5000, 'javascript', now()),
  ('00000000-0000-0000-0000-000000000002', 3, 20579, 10, 10, 'Wrong Answer', 0, 5000, 'typescript', now()),
  ('00000000-0000-0000-0000-000000000002', 4, 12345, 10, 10, 'Accepted', 0, 5000, 'python', now()),
  ('00000000-0000-0000-0000-000000000004', 0, 52340, 10, 10, 'Accepted', 0, 5000, 'c', now())
;
