package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"leetcodeduels/models"
	"leetcodeduels/services"
	"leetcodeduels/store"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

const (
	userLocationPrefix  = "user_location:"
	serverChannelPrefix = "server:"
	userLocationTTL     = 60 * time.Second
)

var ConnManager *connManager

type directMessage struct {
	userID  int64
	payload []byte
}

type redisPubSubMessage struct {
	UserID  int64           `json:"user_id"`
	Payload json.RawMessage `json:"payload"`
}

type connManager struct {
	serverID    string
	redisClient *redis.Client
	pubsub      *redis.PubSub

	register   chan *Client
	unregister chan *Client

	clients     map[*Client]bool           // all connected clients on this node
	userClients map[int64]map[*Client]bool // connections grouped by userID

	// local direct queue for delivering messages to local clients
	direct chan directMessage

	ctx    context.Context
	cancel context.CancelFunc
}

func userLocationKey(userID int64) string {
	return fmt.Sprintf("%s%d", userLocationPrefix, userID)
}

func serverChannel(serverID string) string {
	return fmt.Sprintf("%s%s", serverChannelPrefix, serverID)
}

func newConnManager(redisURL string) (*connManager, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redis URL: %w", err)
	}
	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	serverUUID := uuid.New().String()
	log.Printf("Starting server with ID: %s", serverUUID)

	ps := client.Subscribe(context.Background(), serverChannel(serverUUID))

	cm := &connManager{
		serverID:    serverUUID,
		redisClient: client,
		pubsub:      ps,
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		clients:     make(map[*Client]bool),
		userClients: make(map[int64]map[*Client]bool),
		direct:      make(chan directMessage, 256),
		ctx:         ctx,
		cancel:      cancel,
	}
	go cm.run()
	go cm.redisListener()
	return cm, nil
}

func InitConnManager(redisURL string) error {
	var err error
	ConnManager, err = newConnManager(redisURL)
	if err != nil {
		return err
	}
	return nil
}

func (cm *connManager) run() {
	for {
		select {
		case <-cm.ctx.Done():
			return

		case c := <-cm.register:
			cm.clients[c] = true
			uc, exists := cm.userClients[c.userID]
			if !exists {
				uc = make(map[*Client]bool)
				cm.userClients[c.userID] = uc
			}
			uc[c] = true

			err := cm.redisClient.Set(
				context.Background(),
				userLocationKey(c.userID),
				cm.serverID,
				userLocationTTL,
			).Err()
			if err != nil {
				log.Printf("failed to set user location for user %d: %v", c.userID, err)
			}
			log.Printf("User %d registered on server %s", c.userID, cm.serverID)

		case c := <-cm.unregister:
			if _, ok := cm.clients[c]; !ok {
				continue
			}
			delete(cm.clients, c)

			if uc, exists := cm.userClients[c.userID]; exists {
				delete(uc, c)
				if len(uc) == 0 {
					delete(cm.userClients, c.userID)
					cm.redisClient.Del(context.Background(), userLocationKey(c.userID))
					log.Printf("User %d unregistered from server %s", c.userID, cm.serverID)
				}
			}
			close(c.send)

		case dm := <-cm.direct:
			if conns, ok := cm.userClients[dm.userID]; ok {
				for c := range conns {
					select {
					case c.send <- dm.payload:
					default:
						close(c.send)
						delete(conns, c)
					}
				}
			}
		}
	}
}

func (cm *connManager) redisListener() {
	ch := cm.pubsub.Channel()
	for {
		select {
		case <-cm.ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}

			var pubSubMsg redisPubSubMessage
			if err := json.Unmarshal([]byte(msg.Payload), &pubSubMsg); err != nil {
				log.Printf("could not unmarshal redis pubsub message: %v", err)
				continue
			}

			cm.direct <- directMessage{
				userID:  pubSubMsg.UserID,
				payload: pubSubMsg.Payload,
			}
		}
	}
}

func (cm *connManager) IsUserOnline(userID int64) (bool, error) {
	exists, err := cm.redisClient.Exists(context.Background(), userLocationKey(userID)).Result()
	if err != nil {
		return false, fmt.Errorf("could not check user online status for user %d: %w", userID, err)
	}
	return exists == 1, nil
}

func (cm *connManager) SendToUser(userID int64, payload []byte) error {
	serverID, err := cm.redisClient.Get(context.Background(), userLocationKey(userID)).Result()
	if err == redis.Nil {
		// User is not connected to any server instance.
		log.Printf("attempted to send message to offline user %d", userID)
		return nil
	}
	if err != nil {
		return fmt.Errorf("could not get user location for user %d: %w", userID, err)
	}

	if serverID == cm.serverID {
		cm.direct <- directMessage{userID: userID, payload: payload}
		return nil
	}

	pubSubMsg := redisPubSubMessage{
		UserID:  userID,
		Payload: payload,
	}
	b, err := json.Marshal(pubSubMsg)
	if err != nil {
		return fmt.Errorf("could not marshal pubsub message: %w", err)
	}

	return cm.redisClient.Publish(context.Background(), serverChannel(serverID), b).Err()
}

func (cm *connManager) refreshUserTTL(userID int64) error {
	return cm.redisClient.Expire(context.Background(), userLocationKey(userID), userLocationTTL).Err()
}

func (h *connManager) HandleClientMessage(c *Client, env *Message) error {
	switch env.Type {

	case ClientMsgHeartbeat:
		return h.refreshUserTTL(c.userID)

	case ClientMsgSendInvitation:
		var p SendInvitationPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return fmt.Errorf("invalid payload for %s: %w", env.Type, err)
		}
		return h.handleSendInvitation(c.userID, p)

	case ClientMsgAcceptInvitation:
		var p AcceptInvitationPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return fmt.Errorf("invalid payload for %s: %w", env.Type, err)
		}
		return h.handleAcceptInvitation(c.userID, p)

	case ClientMsgCancelInvitation:
		return h.handleCancelInvitation(c.userID)

	case ClientMsgDeclineInvitation:
		var p DeclineInvitationPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return fmt.Errorf("invalid payload for %s: %w", env.Type, err)
		}
		return h.handleDeclineInvitation(p)

	case ClientMsgEnterQueue:
		var p EnterQueuePayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return fmt.Errorf("invalid payload for %s: %w", env.Type, err)
		}
		return h.handleEnterQueue(c.userID, p)

	case ClientMsgLeaveQueue:
		return h.handleLeaveQueue(c.userID)

	case ClientMsgSubmission:
		var p SubmissionPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return fmt.Errorf("invalid payload for %s: %w", env.Type, err)
		}
		return h.handleSubmission(c.userID, p)

	default:
		c.sendError("unknown_type", "message type not recognized")
		return nil
	}
}

func (cm *connManager) Close() error {
	for c := range cm.clients {
		cm.redisClient.Del(context.Background(), userLocationKey(c.userID))
		cm.unregister <- c
	}

	// stop run() and redisListener()
	cm.cancel()

	if err := cm.pubsub.Close(); err != nil {
		log.Printf("error closing pubsub: %v", err)
	}

	return cm.redisClient.Close()
}

func (c *connManager) handleSendInvitation(
	userID int64, p SendInvitationPayload,
) error {
	isOnline := c.redisClient.Exists(context.Background(), userLocationKey(p.InviteeID)).Val() == 1
	if !isOnline {
		b, _ := json.Marshal(Message{Type: ServerMsgUserOffline})
		c.direct <- directMessage{userID: userID, payload: b}
		return nil
	}

	success, err := services.InviteManager.CreateInvite(userID, p.InviteeID, p.MatchDetails)
	if err != nil {
		return err
	}
	if !success {
		// Invite already exists from this user
		// TODO: Either replace the existing invite or ignore this invite
		return nil
	}

	request := InvitationRequestPayload{InviterID: userID, MatchDetails: p.MatchDetails}
	payload, _ := json.Marshal(request)

	msg := Message{Type: ServerMsgInvitationRequest, Payload: payload}
	b, _ := json.Marshal(msg)

	// SendToUser will now correctly route the message across servers.
	return ConnManager.SendToUser(p.InviteeID, b)
}

func (c *connManager) handleAcceptInvitation(userID int64, p AcceptInvitationPayload) error {
	invite, err := services.InviteManager.InviteDetails(p.InviterID)
	if err != nil {
		return err
	}
	if invite == nil {
		b, _ := json.Marshal(Message{Type: ServerMsgInviteDoesNotExist})
		ConnManager.SendToUser(userID, b)
		return nil
	}

	// remove the invite
	removed, err := services.InviteManager.RemoveInvite(p.InviterID)
	if err != nil {
		return err
	}
	if !removed {
		b, _ := json.Marshal(Message{Type: ServerMsgInviteDoesNotExist})
		ConnManager.SendToUser(userID, b)
		return nil
	}

	problem, err := store.DataStore.GetRandomProblemByTagsAndDifficulties(invite.MatchDetails.Tags, invite.MatchDetails.Difficulties)
	if err != nil {
		return err
	}
	if problem == nil {
		return fmt.Errorf("no problem found matching preferences")
	}

	// start the session
	sessionID, err := services.GameManager.StartGame(
		[]int64{p.InviterID, userID},
		*problem,
	)
	if err != nil {
		return err
	}

	problemURL := fmt.Sprintf("https://leetcode.com/problems/%s", problem.Slug)

	// notify accepter
	startPayload := StartGamePayload{
		SessionID:  sessionID,
		ProblemURL: problemURL,
		OpponentID: p.InviterID,
	}
	b, _ := json.Marshal(Message{Type: ServerMsgStartGame, Payload: MarshalPayload(startPayload)})
	err = ConnManager.SendToUser(userID, b)
	if err != nil {
		return err
	}

	// notify inviter
	startPayload.OpponentID = userID
	b, _ = json.Marshal(Message{Type: ServerMsgStartGame, Payload: MarshalPayload(startPayload)})
	err = ConnManager.SendToUser(p.InviterID, b)
	if err != nil {
		return err
	}

	return nil
}

func (c *connManager) handleDeclineInvitation(p DeclineInvitationPayload) error {
	invite, err := services.InviteManager.InviteDetails(p.InviterID)
	if err != nil {
		return err
	}
	if invite == nil {
		// no invite to decline
		return nil
	}

	success, err := services.InviteManager.RemoveInvite(p.InviterID)
	if err != nil {
		return err
	}
	if !success {
		// no invite to decline
		return nil
	}

	b, _ := json.Marshal(Message{Type: ServerMsgInvitationDeclined})
	err = ConnManager.SendToUser(p.InviterID, b)
	if err != nil {
		return err
	}

	return nil
}

func (c *connManager) handleCancelInvitation(userID int64) error {
	invite, err := services.InviteManager.InviteDetails(userID)
	if err != nil {
		return err
	}
	if invite == nil {
		b, _ := json.Marshal(Message{Type: ServerMsgInviteDoesNotExist})
		err = ConnManager.SendToUser(userID, b)
		if err != nil {
			return err
		}
		return nil
	}

	success, err := services.InviteManager.RemoveInvite(userID)
	if err != nil {
		return err
	}
	if !success {
		// this shouldn’t really happen, but guard anyway
		b, _ := json.Marshal(Message{Type: ServerMsgInviteDoesNotExist})
		err = ConnManager.SendToUser(invite.InviteeID, b)
		if err != nil {
			return err
		}
		return nil
	}

	payload := InvitationCanceledPayload{InviterID: userID}
	b, _ := json.Marshal(Message{Type: ServerMsgInvitationCanceled, Payload: MarshalPayload(payload)})
	err = ConnManager.SendToUser(invite.InviteeID, b)
	if err != nil {
		return err
	}

	return nil
}

func (c *connManager) handleEnterQueue(userID int64, p EnterQueuePayload) error {
	return fmt.Errorf("unimplemented")
}

func (c *connManager) handleLeaveQueue(userID int64) error {
	return fmt.Errorf("unimplemented")
}

func (c *connManager) handleSubmission(userID int64, p SubmissionPayload) error {
	sessionID, err := services.GameManager.GetSessionIDByPlayer(userID)
	if err != nil {
		return err
	}
	if sessionID == "" {
		// User is not in-game
		return err
	}

	session, err := services.GameManager.GetGame(sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		// this shouldn’t really happen, but guard anyway
		return err
	}

	submissionID := len(session.Submissions)
	submissionStatus, _ := models.ParseSubmissionStatus(p.Status)
	submissionLang, _ := models.ParseLang(p.Language)
	submission := models.PlayerSubmission{
		ID:              submissionID,
		PlayerID:        userID,
		PassedTestCases: p.PassedTestCases,
		TotalTestCases:  p.TotalTestCases,
		Status:          submissionStatus,
		Runtime:         p.Runtime,
		Memory:          p.Memory,
		Lang:            submissionLang,
		Time:            p.Time,
	}

	// TODO: Verify submission information is correct against LeetCode's API

	err = services.GameManager.AddSubmission(sessionID, submission)
	if err != nil {
		return err
	}

	opponentID, err := services.GameManager.GetOpponent(sessionID, userID)
	if err != nil {
		return err
	}

	if submissionStatus == models.Accepted {
		session, err = services.GameManager.CompleteGame(sessionID, userID)
		if err != nil {
			return err
		}

		duration := submission.Time.Sub(session.StartTime)
		durationSecs := int64(duration.Seconds())

		reply := GameOverPayload{
			WinnerID:  userID,
			SessionID: sessionID,
			Duration:  durationSecs,
		}
		payload, _ := json.Marshal(reply)
		msg := Message{Type: ServerMsgGameOver, Payload: payload}
		b, _ := json.Marshal(msg)

		err = ConnManager.SendToUser(userID, b)
		if err != nil {
			return err
		}
		err = ConnManager.SendToUser(opponentID, b)
		if err != nil {
			return err
		}

		err = store.DataStore.StoreMatch(session)
		if err != nil {
			return err
		}
		return nil
	}

	reply := OpponentSubmissionPayload{
		ID:              submissionID,
		PlayerID:        userID,
		PassedTestCases: p.PassedTestCases,
		TotalTestCases:  p.TotalTestCases,
		Status:          p.Status,
		Runtime:         p.Runtime,
		Memory:          p.Memory,
		Language:        p.Language,
		Time:            p.Time,
	}

	payload, _ := json.Marshal(reply)
	msg := Message{Type: ServerMsgOpponentSubmission, Payload: payload}
	b, _ := json.Marshal(msg)
	err = ConnManager.SendToUser(opponentID, b)
	if err != nil {
		return err
	}

	return nil
}
