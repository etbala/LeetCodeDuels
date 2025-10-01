import http from 'k6/http';
import ws from 'k6/ws';
import { check, group, sleep } from 'k6';
import { randomString, randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';
import crypto from 'k6/crypto'
import encoding from 'k6/encoding';

// --------------------------------------------------------------------------------
// ---                           CONFIGURATION                                  ---
// --------------------------------------------------------------------------------
// Use environment variables to configure the test, e.g.:
// k6 run -e BASE_URL=http://localhost:8080 -e JWT_SECRET=your_secret stress-test.js

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const WS_URL = BASE_URL.replace('http', 'ws');
const JWT_SECRET = __ENV.JWT_SECRET || 'testjwtsecret';

// Static user IDs for cross-endpoint checks
const STATIC_USER_ID_1 = 9001;
const STATIC_USER_ID_2 = 9002;

// --------------------------------------------------------------------------------
// ---                              K6 OPTIONS                                  ---
// --------------------------------------------------------------------------------

export const options = {
    scenarios: {
        // Scenario for hammering the REST API endpoints
        rest_api: {
            executor: 'constant-vus',
            exec: 'restApiScenario',
            vus: 10,
            duration: '1m',
            tags: { test_type: 'rest_api' },
        },
        // Scenario for simulating the full WebSocket game flow
        websocket_flow: {
            executor: 'per-vu-iterations',
            exec: 'websocketGameScenario',
            vus: 5,
            iterations: 10,
            maxDuration: '2m',
            tags: { test_type: 'websocket_game_flow' },
        },
    },
    thresholds: {
        'http_req_failed': ['rate<0.01'], // http errors should be less than 1%
        'http_req_duration': ['p(95)<500'], // 95% of requests should be below 500ms
        'ws_msgs_sent': ['count>0'],
        'ws_msgs_received': ['count>0'],
        'checks': ['rate>0.98'], // Over 98% of checks should pass
    },
};

// --------------------------------------------------------------------------------
// ---                           HELPER FUNCTIONS                               ---
// --------------------------------------------------------------------------------

/**
 * Simple base64url encoding
 */
function base64urlEncode(str) {
    return encoding.b64encode(str, 'rawstd')
        .replace(/\+/g, '-')
        .replace(/\//g, '_')
        .replace(/=/g, '');
}

/**
 * Generates an HS256 JWT for a given user ID.
 * @param {number} userID - The user ID to include in the token.
 * @returns {string} The generated JWT.
 */
function generateJWT(userID) {
    const header = {
        alg: 'HS256',
        typ: 'JWT'
    };
    
    const payload = {
        userid: userID,
        exp: Math.floor(Date.now() / 1000) + (60 * 60),
        iat: Math.floor(Date.now() / 1000)
    };
    
    const encodedHeader = base64urlEncode(JSON.stringify(header));
    const encodedPayload = base64urlEncode(JSON.stringify(payload));
    const message = `${encodedHeader}.${encodedPayload}`;
    
    const signature = crypto.hmac('sha256', JWT_SECRET, message, 'base64rawurl');
    
    return `${message}.${signature}`;
}

/**
 * Creates a WebSocket submission message object.
 * @param {string} status - e.g., "Accepted", "Wrong Answer"
 * @param {number} passed - Number of test cases passed.
 * @param {number} total - Total number of test cases.
 * @returns {object} The complete WebSocket message for a submission.
 */
function createSubmission(status, passed, total) {
    return {
        type: 'submission',
        payload: {
            status: status,
            passedTestCases: passed,
            totalTestCases: total,
            runtime: randomIntBetween(20, 200),
            memory: randomIntBetween(15000, 30000),
            language: ['python3', 'go', 'javascript', 'java'][randomIntBetween(0, 3)],
            time: new Date().toISOString(),
        }
    };
}


// --------------------------------------------------------------------------------
// ---                       SCENARIO 1: REST API TESTS                         ---
// --------------------------------------------------------------------------------

export function restApiScenario() {
    const user1Token = generateJWT(STATIC_USER_ID_1);
    const authHeaders = { 
        headers: { 
            'Authorization': `Bearer ${user1Token}`,
            'Content-Type': 'application/json'
        } 
    };

    group('Public Endpoints', () => {
        const healthRes = http.get(`${BASE_URL}/api/v1/health`);
        check(healthRes, { 'GET /api/v1/health': (r) => r.status === 200 });

        const tagsRes = http.get(`${BASE_URL}/api/v1/problems/tags`);
        check(tagsRes, { 'GET /api/v1/problems/tags': (r) => r.status === 200 });

        const randomProblemRes = http.get(`${BASE_URL}/api/v1/problems/random?difficulty[]=Easy&difficulty[]=Medium`);
        check(randomProblemRes, { 'GET /api/v1/problems/random': (r) => r.status === 200 });
    });

    sleep(1);

    group('Authenticated User/Account Endpoints', () => {
        // Test /users/me endpoints
        const meRes = http.get(`${BASE_URL}/api/v1/users/me`, authHeaders);
        check(meRes, { 'GET /users/me': (r) => r.status === 200 });

        // Update user with correct field names from UpdateUserRequest
        const patchPayload = JSON.stringify({ 
            username: `k6-user-${randomString(6)}`,
            lc_username: `leetcode-${randomString(4)}`
        });
        const patchRes = http.patch(`${BASE_URL}/api/v1/users/me`, patchPayload, authHeaders);
        check(patchRes, { 'PATCH /users/me': (r) => r.status === 200 || r.status === 204 });

        const notificationsRes = http.get(`${BASE_URL}/api/v1/users/me/notifications`, authHeaders);
        check(notificationsRes, { 'GET /users/me/notifications': (r) => r.status === 200 });
        
        // Test user search and profile lookup
        const searchRes = http.get(`${BASE_URL}/api/v1/users?username=user&limit=5`, authHeaders);
        check(searchRes, { 'GET /users?username=...': (r) => r.status === 200 });

        const profileRes = http.get(`${BASE_URL}/api/v1/users/${STATIC_USER_ID_2}`, authHeaders);
        check(profileRes, { 'GET /users/{id}': (r) => r.status === 200 });

        const statusRes = http.get(`${BASE_URL}/api/v1/users/${STATIC_USER_ID_2}/status`, authHeaders);
        check(statusRes, { 'GET /users/{id}/status': (r) => r.status === 200 });

        const matchesRes = http.get(`${BASE_URL}/api/v1/users/${STATIC_USER_ID_1}/matches?page=1&limit=10`, authHeaders);
        check(matchesRes, { 'GET /users/{id}/matches': (r) => r.status === 200 });
    });

    sleep(1);

    group('Queue Endpoint', () => {
        // Test queue size endpoint
        const queueSizeRes = http.get(`${BASE_URL}/api/v1/queue/size`, authHeaders);
        check(queueSizeRes, { 'GET /queue/size': (r) => r.status === 200 });
    });

    sleep(1);
}

// --------------------------------------------------------------------------------
// ---                  SCENARIO 2: WEBSOCKET GAME FLOW TESTS                   ---
// --------------------------------------------------------------------------------

export function websocketGameScenario() {
    // Each VU iteration uses unique player IDs to avoid conflicts
    const inviterId = 10000 + (__VU * 10) + __ITER;
    const inviteeId = 20000 + (__VU * 10) + __ITER;

    const inviterToken = generateJWT(inviterId);
    const inviteeToken = generateJWT(inviteeId);

    const inviterAuthParams = { headers: { 'Authorization': `Bearer ${inviterToken}` } };
    const inviteeAuthParams = { headers: { 'Authorization': `Bearer ${inviteeToken}` } };

    let sessionID = null; // Will be captured from the 'start_game' message

    const res = ws.connect(`${WS_URL}/ws`, inviterAuthParams, function (inviterSocket) {
        inviterSocket.on('open', function open() {
            // Inviter connected, now connect the invitee
            ws.connect(`${WS_URL}/ws`, inviteeAuthParams, function (inviteeSocket) {
                // Invitee connected, now both players are online.
                
                // --- Set up Invitee's behavior ---
                inviteeSocket.on('message', function (data) {
                    const msg = JSON.parse(data);
                    switch (msg.type) {
                        case 'invitation_request':
                            check(msg, { '[Invitee] receives invitation': m => m.payload && m.payload.from_user });
                            // Accept the invitation
                            inviteeSocket.send(JSON.stringify({
                                type: 'accept_invitation',
                                payload: { inviterID: inviterId }
                            }));
                            break;
                        case 'start_game':
                            check(msg, { '[Invitee] game starts': m => m.payload && m.payload.sessionID });
                            sessionID = msg.payload.sessionID;
                            // Send a failing submission after a random delay
                            setTimeout(() => {
                                inviteeSocket.send(JSON.stringify(createSubmission('Wrong Answer', 5, 20)));
                            }, randomIntBetween(500, 2000));
                            break;
                        case 'game_over':
                            check(msg, { '[Invitee] receives game over': m => m.payload && m.payload.winnerID });
                            inviteeSocket.close();
                            break;
                    }
                });

                // --- Set up Inviter's behavior ---
                inviterSocket.on('message', function (data) {
                    const msg = JSON.parse(data);
                     switch (msg.type) {
                        case 'start_game':
                            check(msg, { '[Inviter] game starts': m => m.payload && m.payload.sessionID });
                            sessionID = msg.payload.sessionID; // Capture the session ID!
                            break;
                        case 'opponent_submission':
                            check(msg, { '[Inviter] receives opponent submission': m => m.payload && m.payload.playerID });
                             // Send a winning submission after a random delay
                            setTimeout(() => {
                                inviterSocket.send(JSON.stringify(createSubmission('Accepted', 20, 20)));
                            }, randomIntBetween(500, 1500));
                            break;
                        case 'game_over':
                            check(msg, { '[Inviter] receives game over': m => m.payload && m.payload.winnerID });
                            inviterSocket.close();
                            break;
                     }
                });
                
                // Add handlers for close/error events for robustness
                inviteeSocket.on('close', () => console.log(`Invitee ${inviteeId} disconnected.`));
                inviteeSocket.on('error', (e) => console.error(`Invitee ${inviteeId} error: ${e.error()}`));
                inviterSocket.on('error', (e) => console.error(`Inviter ${inviterId} error: ${e.error()}`));

                // Set up heartbeats to keep connections alive
                const inviterHeartbeat = setInterval(() => {
                    if (inviterSocket.readyState === 1) {
                        inviterSocket.send(JSON.stringify({ type: 'heartbeat' }));
                    }
                }, 30000);
                
                const inviteeHeartbeat = setInterval(() => {
                    if (inviteeSocket.readyState === 1) {
                        inviteeSocket.send(JSON.stringify({ type: 'heartbeat' }));
                    }
                }, 30000);

                // Clean up intervals on close
                inviterSocket.on('close', () => clearInterval(inviterHeartbeat));
                inviteeSocket.on('close', () => clearInterval(inviteeHeartbeat));

                // --- Kick off the flow by sending the invitation ---
                inviterSocket.send(JSON.stringify({
                    type: 'send_invitation',
                    payload: {
                        inviteeID: inviteeId,
                        matchDetails: { 
                            isRated: true,
                            difficulties: ['Easy', 'Medium'], 
                            tags: [1, 2] 
                        }
                    }
                }));
            });
        });

        // --- Post-Game REST Checks ---
        inviterSocket.on('close', function () {
            console.log(`Inviter ${inviterId} disconnected. Game session ${sessionID} concluded.`);
            
            if (sessionID) {
                group('Post-Game Match API Checks', () => {
                    const matchAuthHeaders = { 
                        headers: { 
                            'Authorization': `Bearer ${inviterToken}`,
                            'Content-Type': 'application/json'
                        } 
                    };
                    
                    const resMatch = http.get(`${BASE_URL}/api/v1/matches/${sessionID}`, matchAuthHeaders);
                    check(resMatch, {
                        'GET /matches/{id} for completed game': (r) => r.status === 200,
                    });

                    const resSubmissions = http.get(`${BASE_URL}/api/v1/matches/${sessionID}/submissions`, matchAuthHeaders);
                    check(resSubmissions, {
                        'GET /matches/{id}/submissions for completed game': (r) => r.status === 200,
                    });
                });
            } else {
                 console.error(`[VU ${__VU}] Failed to capture sessionID. Skipping post-game checks.`);
            }
        });
    });

    check(res, { '[WS Flow] Connections established': (r) => r && r.status === 101 });
}