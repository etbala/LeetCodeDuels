import http from 'k6/http';
import ws from 'k6/ws';
import { check, group, sleep } from 'k6';
import { Counter } from 'k6/metrics';
import { randomString, randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';
import crypto from 'k6/crypto'
import encoding from 'k6/encoding';

// --------------------------------------------------------------------------------
// ---                           CONFIGURATION                                  ---
// --------------------------------------------------------------------------------
// Use environment variables to configure the test, e.g.:
// k6 run -e BASE_URL=http://localhost:8080 -e JWT_SECRET=jwtsecret stress-test.js

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const WS_URL = BASE_URL.replace('http', 'ws');
const JWT_SECRET = __ENV.JWT_SECRET || 'testjwtsecret';

const gamesCompleted = new Counter('games_completed');

const SEED_USER_IDS = [
    12345, 67890, 87902, 20579, 25074, 43567, 56563, 81970, 92349, 31657,
    10987, 26354, 51796, 73443, 43298, 97862, 70763, 82352, 77356, 12346,
    32189, 62307, 52340, 49876, 53468, 41529, 61539, 9001, 9002
];

const EXISTING_USERNAMES = ['alice', 'bob', 'charlie', 'david', 'emily', 'fiona',
     'gavin', 'henry', 'isabel', 'juliet', 'katya', 'lisa', 'matt', 'nancy', 'owen',
    'philip', 'quincy', 'rachel', 'samuel', 'samantha', 'tom', 'uri', 'victor',
    'willow', 'xavier', 'yash', 'zoe', 'k6user1', 'k6user2'];

const DIFFICULTIES = ['Easy', 'Medium', 'Hard'];
const TAG_IDS = [1, 2, 3, 5, 6, 7, 8, 9, 10, 11];

export const options = {
    scenarios: {
        // Scenario for hammering REST endpoints
        rest_api: {
            executor: 'constant-vus',
            exec: 'restApiScenario',
            vus: 29,
            duration: '1m',
            tags: { test_type: 'rest_api' },
        },
        // Scenario for continuous WebSocket games
        websocket_game_pairs: {
            executor: 'constant-vus',
            exec: 'gamePairScenario',
            vus: 10, // Must be even number
            duration: '1m',
        }
    },
    thresholds: {
        'http_req_failed': ['rate<0.01'], // http errors should be less than 1%
        'http_req_duration': ['p(95)<300'], // 95% of requests should be below 300ms
        'ws_msgs_sent': ['count>0'],
        'ws_msgs_received': ['count>0'],
        'checks': ['rate>0.98'], // Over 98% of checks should pass
    },
};

function getRandomUserForVU() {
    // Use VU number to get consistent user per VU, but random across VUs
    const index = (__VU - 1) % SEED_USER_IDS.length;
    return SEED_USER_IDS[index];
}

function getSecondRandomUser(firstUserId) {
    const availableUsers = SEED_USER_IDS.filter(id => id !== firstUserId);
    return availableUsers[randomIntBetween(0, availableUsers.length - 1)];
}

function getRandomDifficulties() {
    const numDifficulties = randomIntBetween(1, DIFFICULTIES.length);
    const shuffled = [...DIFFICULTIES].sort(() => 0.5 - Math.random());
    return shuffled.slice(0, numDifficulties);
}

function getRandomTags() {
    const numTags = randomIntBetween(1, Math.min(3, TAG_IDS.length));
    const shuffled = [...TAG_IDS].sort(() => 0.5 - Math.random());
    return shuffled.slice(0, numTags);
}

// Simple base64url encoding
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
    const languages = ['cpp', 'java', 'python3', 'go'];
    return {
        type: 'submission',
        payload: {
            submissionID: randomIntBetween(1000000, 9999999),
            status: status,
            passedTestCases: passed,
            totalTestCases: total,
            runtime: randomIntBetween(20, 200),
            runtimePercentile: Math.random(),
            memory: randomIntBetween(15000, 30000),
            memoryPercentile: Math.random(),
            language: languages[randomIntBetween(0, languages.length - 1)],
            time: new Date().toISOString(),
        }
    };
}

function connectWithTicket(userId, callback) {
    const token = generateJWT(userId);
    const ticketRes = http.post(`${BASE_URL}/api/v1/ws-ticket`, null, {
        headers: { 'Authorization': `Bearer ${token}` },
    });

    if (!check(ticketRes, { '[WS Ticket] Auth ticket request successful': (r) => r.status === 200 })) {
        console.error(`[VU ${__VU}] Failed to get ticket for user ${userId}: ${ticketRes.body}`);
        return null;
    }

    const ticket = ticketRes.json('ticket');
    if (!check(ticket, { '[WS Ticket] Ticket received': (t) => t && typeof t === 'string' })) {
        console.error(`[VU ${__VU}] Ticket was empty for user ${userId}`);
        return null;
    }

    const url = `${WS_URL}/ws?ticket=${ticket}`;
    return ws.connect(url, null, callback);
}

export function restApiScenario() {
    const currentUserId = getRandomUserForVU();
    const userToken = generateJWT(currentUserId);
    const authHeaders = { 
        headers: { 
            'Authorization': `Bearer ${userToken}`,
            'Content-Type': 'application/json'
        } 
    };

    group('Public Endpoints', () => {
        const healthRes = http.get(`${BASE_URL}/api/v1/health`);
        check(healthRes, { 'GET /api/v1/health': (r) => r.status === 200 });

        const tagsRes = http.get(`${BASE_URL}/api/v1/problems/tags`);
        check(tagsRes, { 'GET /api/v1/problems/tags': (r) => r.status === 200 });

        const randomDifficulties = getRandomDifficulties();
        const randomTags = getRandomTags();
        const difficultyParams = randomDifficulties.map(d => `difficulty[]=${d}`).join('&');
        const tagParams = randomTags.map(t => `tags[]=${t}`).join('&');
        const queryString = `${difficultyParams}&${tagParams}`;
        
        const randomProblemRes = http.get(`${BASE_URL}/api/v1/problems/random?${queryString}`);
        check(randomProblemRes, { 'GET /api/v1/problems/random': (r) => r.status === 200 });
    });

    sleep(1);

    group('Authenticated User/Account Endpoints', () => {
        const meRes = http.get(`${BASE_URL}/api/v1/users/me`, authHeaders);
        check(meRes, { 'GET /users/me': (r) => r.status === 200 });

        const patchPayload = JSON.stringify({ 
            username: `k6-user-${randomString(6)}`,
            lc_username: `leetcode-${randomString(4)}`
        });
        const patchRes = http.patch(`${BASE_URL}/api/v1/users/me`, patchPayload, authHeaders);
        check(patchRes, { 'PATCH /users/me': (r) => r.status === 200 || r.status === 204 });

        const searchUsername = EXISTING_USERNAMES[randomIntBetween(0, EXISTING_USERNAMES.length - 1)];
        const searchRes = http.get(`${BASE_URL}/api/v1/users?username=${searchUsername}&limit=5`, authHeaders);
        check(searchRes, { 'GET /users?username=...': (r) => r.status === 200 });

        const targetUserId = getSecondRandomUser(currentUserId);
        const profileRes = http.get(`${BASE_URL}/api/v1/users/${targetUserId}`, authHeaders);
        check(profileRes, { 'GET /users/{id}': (r) => r.status === 200 });

        const statusRes = http.get(`${BASE_URL}/api/v1/users/${targetUserId}/status`, authHeaders);
        check(statusRes, { 'GET /users/{id}/status': (r) => r.status === 200 });

        const matchesRes = http.get(`${BASE_URL}/api/v1/users/${currentUserId}/matches?page=1&limit=10`, authHeaders);
        check(matchesRes, { 'GET /users/{id}/matches': (r) => r.status === 200 });
    });

    sleep(1);

    group('Match History Endpoints', () => {
        const existingMatchId = '11111111-1111-1111-1111-111111111111';
        
        const matchRes = http.get(`${BASE_URL}/api/v1/matches/${existingMatchId}`, authHeaders);
        check(matchRes, { 'GET /matches/{id} (existing)': (r) => r.status === 200 });

        const submissionsRes = http.get(`${BASE_URL}/api/v1/matches/${existingMatchId}/submissions`, authHeaders);
        check(submissionsRes, { 'GET /matches/{id}/submissions (existing)': (r) => r.status === 200 });
    });

    sleep(1);
}

export function gamePairScenario() {
    const myVUID = __VU;
    const isInviter = myVUID % 2 !== 0;
    const partnerVUID = isInviter ? myVUID + 1 : myVUID - 1;

    const myUserID = SEED_USER_IDS[myVUID - 1];
    const partnerUserID = SEED_USER_IDS[partnerVUID - 1];
    const role = isInviter ? 'INVITER' : 'INVITEE';

    const res = connectWithTicket(myUserID, function (socket) {
        socket.on('open', () => {
            console.log(`[VU ${myVUID} | ${role}] WebSocket connected.`);
            // If I am the inviter, send the FIRST invitation to kick things off.
            if (isInviter) {
                sendInvitation(socket);
            }
        });

        socket.on('message', (data) => {
            const msg = JSON.parse(data);
            console.log(`[VU ${myVUID} | ${role}] Received: ${msg.type}`);

            switch (msg.type) {
                case 'invitation_request':
                    if (!isInviter) {
                        console.log(`[VU ${myVUID} | ${role}] Accepting invitation.`);
                        socket.send(JSON.stringify({
                            type: 'accept_invitation',
                            payload: { inviterID: msg.payload.inviterID }
                        }));
                    }
                    break;
                
                case 'start_game':
                    if (!isInviter) {
                        // Invitee sends a winning submission after a short delay
                        socket.setTimeout(() => {
                            console.log(`[VU ${myVUID} | ${role}] Sending submission.`);
                            socket.send(JSON.stringify(createSubmission('Accepted', 20, 20)));
                        }, randomIntBetween(20, 50));
                    }
                    break;
                
                case 'game_over':
                    gamesCompleted.add(1);
                    // LOOP: If I'm the inviter, wait and start a new game.
                    if (isInviter) {
                        socket.setTimeout(() => {
                            sendInvitation(socket);
                        }, randomIntBetween(30, 60)); // Wait 30-60ms before next game
                    }
                    break;
            }
        });

        socket.on('close', () => console.log(`[VU ${myVUID} | ${role}] Disconnected.`));
        socket.on('error', (e) => console.error(`[VU ${myVUID} | ${role}] Error: ${e.error()}`));
    });

    check(res, { [`[VU ${myVUID}] Connection successful`]: (r) => r && r.status === 101 });

    function sendInvitation(socket) {
        console.log(`[VU ${myVUID} | ${role}] Sending invitation to ${partnerUserID}.`);
        socket.send(JSON.stringify({
            type: 'send_invitation',
            payload: {
                inviteeID: partnerUserID,
                matchDetails: {
                    isRated: Math.random() > 0.5,
                    difficulties: getRandomDifficulties(),
                    tags: getRandomTags(),
                }
            }
        }));
    }

    // 3. Keep the VU running for the test duration
    // sleep(options.scenarios.websocket_game_pairs.duration);
}