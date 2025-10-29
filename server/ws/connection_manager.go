package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"leetcodeduels/config"
	"leetcodeduels/models"
	"leetcodeduels/services"
	"leetcodeduels/store"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	userLocationPrefix  = "user_location:"
	serverChannelPrefix = "server:"
	wsTicketPrefix      = "ws_ticket:"
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

	log *zerolog.Logger
}

func userLocationKey(userID int64) string {
	return fmt.Sprintf("%s%d", userLocationPrefix, userID)
}

func serverChannel(serverID string) string {
	return fmt.Sprintf("%s%s", serverChannelPrefix, serverID)
}

func wsTicketKey(ticket string) string {
	return fmt.Sprintf("%s%s", wsTicketPrefix, ticket)
}

func newConnManager(redisURL string) (*connManager, error) {
	logger := log.With().Str("component", "conn_manager").Logger()

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
	log.Info().Str("server_id", serverUUID).Msg("Starting new connection manager")

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
		log:         &logger,
	}

	go cm.run()
	go cm.redisListener()

	cm.log.Info().
		Str("server_id", serverUUID).
		Str("redis_channel", serverChannel(serverUUID)).
		Msg("Connection manager initialized successfully")

	return cm, nil
}

func InitConnManager(redisURL string) error {
	var err error
	ConnManager, err = newConnManager(redisURL)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize connection manager")
		return err
	}
	log.Info().Msg("Global connection manager initialized")
	return nil
}

func (cm *connManager) run() {
	for {
		select {
		case <-cm.ctx.Done():
			cm.log.Info().Msg("Connection manager context cancelled, stopping run loop")
			return

		case c := <-cm.register:
			cm.handleClientRegister(c)

		case c := <-cm.unregister:
			cm.handleClientUnregister(c)

		case dm := <-cm.direct:
			cm.handleDirectMessage(dm)
		}
	}
}

func (cm *connManager) handleClientRegister(c *Client) {
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
		cm.log.Error().
			Err(err).
			Int64("user_id", c.userID).
			Msg("Failed to set user location in Redis")
	}
	cm.log.Info().Int64("user_id", c.userID).Msg("Client registered")
}

func (cm *connManager) handleClientUnregister(c *Client) {
	if _, ok := cm.clients[c]; !ok {
		cm.log.Warn().
			Int64("user_id", c.userID).
			Msg("Attempted to unregister non-existent client")
		return
	}
	delete(cm.clients, c)

	if uc, exists := cm.userClients[c.userID]; exists {
		delete(uc, c)
		if len(uc) == 0 {
			cm.cleanupUserLocation(c.userID)
		}
	}
	close(c.send)
}

func (cm *connManager) cleanupUserLocation(userID int64) {
	delete(cm.userClients, userID)
	err := cm.redisClient.Del(context.Background(), userLocationKey(userID)).Err()
	if err != nil {
		cm.log.Error().
			Err(err).
			Int64("user_id", userID).
			Msg("Failed to delete user location from Redis")
	} else {
		cm.log.Info().
			Int64("user_id", userID).
			Str("server_id", cm.serverID).
			Msg("User completely disconnected from server")
	}
}

func (cm *connManager) handleDirectMessage(dm directMessage) {
	conns, ok := cm.userClients[dm.userID]
	if !ok {
		cm.log.Warn().
			Int64("user_id", dm.userID).
			Msg("No local connections found for user")
		return
	}

	delivered := 0
	failed := 0

	for c := range conns {
		select {
		case c.send <- dm.payload:
			delivered++
		default:
			cm.log.Warn().
				Int64("user_id", dm.userID).
				Msg("Client send buffer full, closing connection")
			close(c.send)
			delete(conns, c)
			failed++
		}
	}

	cm.log.Debug().
		Int64("user_id", dm.userID).
		Int("delivered", delivered).
		Int("failed", failed).
		Int("payload_size", len(dm.payload)).
		Msg("Direct message delivery completed")
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
				cm.log.Error().Err(err).Str("raw_payload", msg.Payload).Msg("Could not unmarshal Redis pubsub message")
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
		cm.log.Error().Err(err).Int64("user_id", userID).Msg("Failed to check user online status")
		return false, fmt.Errorf("could not check user online status for user %d: %w", userID, err)
	}
	return exists == 1, nil
}

func (cm *connManager) SendToUser(userID int64, payload []byte) error {
	serverID, err := cm.redisClient.Get(context.Background(), userLocationKey(userID)).Result()
	if err == redis.Nil {
		cm.log.Warn().Int64("user_id", userID).Msg("User is offline, message not sent")
		return nil
	}
	if err != nil {
		cm.log.Error().Err(err).Int64("user_id", userID).Msg("Failed to get user location")
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
		cm.log.Error().Err(err).Int64("user_id", userID).Msg("Failed to marshal pubsub message")
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
			h.log.Error().Err(err).Int64("user_id", c.userID).Str("message_type", string(env.Type)).Msg("Invalid payload")
			return fmt.Errorf("invalid payload for %s: %w", env.Type, err)
		}
		return h.handleSendInvitation(c.userID, p)

	case ClientMsgAcceptInvitation:
		var p AcceptInvitationPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			h.log.Error().Err(err).Int64("user_id", c.userID).Str("message_type", string(env.Type)).Msg("Invalid payload")
			return fmt.Errorf("invalid payload for %s: %w", env.Type, err)
		}
		return h.handleAcceptInvitation(c.userID, p)

	case ClientMsgCancelInvitation:
		return h.handleCancelInvitation(c.userID)

	case ClientMsgDeclineInvitation:
		var p DeclineInvitationPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			h.log.Error().Err(err).Int64("user_id", c.userID).Str("message_type", string(env.Type)).Msg("Invalid payload")
			return fmt.Errorf("invalid payload for %s: %w", env.Type, err)
		}
		return h.handleDeclineInvitation(p)

	case ClientMsgEnterQueue:
		var p EnterQueuePayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			h.log.Error().Err(err).Int64("user_id", c.userID).Str("message_type", string(env.Type)).Msg("Invalid payload")
			return fmt.Errorf("invalid payload for %s: %w", env.Type, err)
		}
		return h.handleEnterQueue(c.userID, p)

	case ClientMsgLeaveQueue:
		return h.handleLeaveQueue(c.userID)

	case ClientMsgSubmission:
		var p SubmissionPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			h.log.Error().Err(err).Int64("user_id", c.userID).Str("message_type", string(env.Type)).Msg("Invalid payload")
			return fmt.Errorf("invalid payload for %s: %w", env.Type, err)
		}
		return h.handleSubmission(c.userID, p)

	case ClientMsgForfeit:
		return h.handleForfeit(c.userID)

	default:
		h.log.Warn().Int64("user_id", c.userID).Str("message_type", string(env.Type)).Msg("Unknown message type received")
		c.sendError("unknown_type", "message type not recognized")
		return nil
	}
}

func (cm *connManager) Close() error {
	cm.log.Info().Int("client_count", len(cm.clients)).Msg("Closing connection manager")

	for c := range cm.clients {
		err := cm.redisClient.Del(context.Background(), userLocationKey(c.userID)).Err()
		if err != nil {
			cm.log.Error().Err(err).Int64("user_id", c.userID).Msg("Failed to delete user location during shutdown")
		}
		cm.unregister <- c
	}

	// stop run() and redisListener()
	cm.cancel()

	if err := cm.pubsub.Close(); err != nil {
		cm.log.Error().Err(err).Msg("Error closing Redis pubsub")
	}

	err := cm.redisClient.Close()
	if err != nil {
		cm.log.Error().Err(err).Msg("Error closing Redis client")
	}

	cm.log.Info().Msg("Connection manager closed successfully")
	return err
}

func (c *connManager) handleSendInvitation(
	userID int64, p SendInvitationPayload,
) error {
	c.log.Info().
		Int64("inviter_id", userID).
		Int64("invitee_id", p.InviteeID).
		Msg("Processing invitation request")

	// todo: check if inviter and invitee are the same

	isOnline := c.redisClient.Exists(context.Background(), userLocationKey(p.InviteeID)).Val() == 1
	if !isOnline {
		b, _ := json.Marshal(Message{Type: ServerMsgUserOffline})
		c.direct <- directMessage{userID: userID, payload: b}
		return nil
	}

	// todo: check if user is in-game already.

	success, err := services.InviteManager.CreateInvite(userID, p.InviteeID, p.MatchDetails)
	if err != nil {
		c.log.Error().Err(err).Int64("inviter_id", userID).Int64("invitee_id", p.InviteeID).Msg("Failed to create invite")
		return err
	}
	if !success {
		// Invite already exists from this user
		// TODO: Either replace the existing invite or ignore this invite
		c.log.Warn().Int64("inviter_id", userID).Int64("invitee_id", p.InviteeID).Msg("Invite already exists from this user")
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
	c.log.Info().
		Int64("accepter_id", userID).
		Int64("inviter_id", p.InviterID).
		Msg("Processing invitation acceptance")

	invite, err := services.InviteManager.InviteDetails(p.InviterID)
	if err != nil {
		c.log.Error().Err(err).Int64("accepter_id", userID).Int64("inviter_id", p.InviterID).Msg("Failed to get invite details")
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
		c.log.Error().Err(err).Int64("inviter_id", p.InviterID).Msg("Failed to remove invite")
		return err
	}
	if !removed {
		b, _ := json.Marshal(Message{Type: ServerMsgInviteDoesNotExist})
		ConnManager.SendToUser(userID, b)
		return nil
	}

	// todo: check if user is already in game

	problem, err := store.DataStore.GetRandomProblemByTagsAndDifficulties(invite.MatchDetails.Tags, invite.MatchDetails.Difficulties)
	if err != nil {
		c.log.Error().Err(err).Msg("Failed to get random problem")
		return err
	}
	if problem == nil {
		c.log.Warn().Msg("No problem found matching preferences")
		return fmt.Errorf("no problem found matching preferences")
	}

	// start the session
	sessionID, err := services.GameManager.StartGame(
		[]int64{p.InviterID, userID},
		*problem,
	)
	if err != nil {
		c.log.Error().Err(err).Ints64("players", []int64{p.InviterID, userID}).Msg("Failed to start game")
		return err
	}

	c.log.Info().
		Str("session_id", sessionID).
		Ints64("players", []int64{p.InviterID, userID}).
		Str("problem_slug", problem.Slug).
		Msg("Game started successfully")

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
		c.log.Error().Err(err).Int64("user_id", userID).Str("session_id", sessionID).Msg("Failed to notify accepter of game start")
		return err
	}

	// notify inviter
	startPayload.OpponentID = userID
	b, _ = json.Marshal(Message{Type: ServerMsgStartGame, Payload: MarshalPayload(startPayload)})
	err = ConnManager.SendToUser(p.InviterID, b)
	if err != nil {
		c.log.Error().Err(err).Int64("user_id", p.InviterID).Str("session_id", sessionID).Msg("Failed to notify inviter of game start")
		return err
	}

	return nil
}

func (c *connManager) handleDeclineInvitation(p DeclineInvitationPayload) error {
	c.log.Info().
		Int64("inviter_id", p.InviterID).
		Msg("Processing invitation decline")

	invite, err := services.InviteManager.InviteDetails(p.InviterID)
	if err != nil {
		c.log.Error().Err(err).Int64("inviter_id", p.InviterID).Msg("Failed to get invite details")
		return err
	}
	if invite == nil {
		c.log.Warn().Int64("inviter_id", p.InviterID).Msg("No invite to decline")
		return nil
	}

	success, err := services.InviteManager.RemoveInvite(p.InviterID)
	if err != nil {
		c.log.Error().Err(err).Int64("inviter_id", p.InviterID).Msg("Failed to remove invite")
		return err
	}
	if !success {
		c.log.Warn().Int64("inviter_id", p.InviterID).Msg("No invite to decline")
		return nil
	}

	b, _ := json.Marshal(Message{Type: ServerMsgInvitationDeclined})
	err = ConnManager.SendToUser(p.InviterID, b)
	if err != nil {
		c.log.Error().Err(err).Int64("inviter_id", p.InviterID).Msg("Failed to notify inviter of decline")
		return err
	}

	return nil
}

func (c *connManager) handleCancelInvitation(userID int64) error {
	c.log.Info().
		Int64("inviter_id", userID).
		Msg("Processing invitation cancellation")

	invite, err := services.InviteManager.InviteDetails(userID)
	if err != nil {
		c.log.Error().Err(err).Int64("inviter_id", userID).Msg("Failed to get invite details")
		return err
	}
	if invite == nil {
		c.log.Warn().Int64("inviter_id", userID).Msg("No invite to cancel")

		b, _ := json.Marshal(Message{Type: ServerMsgInviteDoesNotExist})
		err = ConnManager.SendToUser(userID, b)
		if err != nil {
			c.log.Error().Err(err).Int64("user_id", userID).Msg("Failed to send invite does not exist message")
			return err
		}
		return nil
	}

	success, err := services.InviteManager.RemoveInvite(userID)
	if err != nil {
		c.log.Error().Err(err).Int64("inviter_id", userID).Msg("Failed to remove invite")
		return err
	}
	if !success {
		// this shouldn’t really happen, but guard anyway
		c.log.Warn().Int64("inviter_id", userID).Msg("Could not cancel invite - already removed")

		b, _ := json.Marshal(Message{Type: ServerMsgInviteDoesNotExist})
		err = ConnManager.SendToUser(invite.InviteeID, b)
		if err != nil {
			c.log.Error().Err(err).Int64("invitee_id", invite.InviteeID).Msg("Failed to send invite does not exist message")
			return err
		}
		return nil
	}

	payload := InvitationCanceledPayload{InviterID: userID}
	b, _ := json.Marshal(Message{Type: ServerMsgInvitationCanceled, Payload: MarshalPayload(payload)})
	err = ConnManager.SendToUser(invite.InviteeID, b)
	if err != nil {
		c.log.Error().Err(err).Int64("invitee_id", invite.InviteeID).Msg("Failed to notify invitee of cancellation")
		return err
	}

	return nil
}

func (c *connManager) handleEnterQueue(userID int64, p EnterQueuePayload) error {
	c.log.Warn().Int64("user_id", userID).Msg("Queue entry not implemented")
	return fmt.Errorf("unimplemented")
}

func (c *connManager) handleLeaveQueue(userID int64) error {
	c.log.Warn().Int64("user_id", userID).Msg("Queue leave not implemented")
	return fmt.Errorf("unimplemented")
}

func (c *connManager) handleSubmission(userID int64, p SubmissionPayload) error {
	c.log.Info().
		Int64("user_id", userID).
		Str("status", string(p.Status)).
		Msg("Processing submission")

	sessionID, err := services.GameManager.GetSessionIDByPlayer(userID)
	if err != nil {
		c.log.Error().Err(err).Int64("user_id", userID).Msg("Failed to get session ID")
		return err
	}
	if sessionID == "" {
		// User is not in-game
		c.log.Warn().Int64("user_id", userID).Msg("User is not in-game")
		return err
	}

	session, err := services.GameManager.GetGame(sessionID)
	if err != nil {
		c.log.Error().Err(err).Str("session_id", sessionID).Msg("Failed to get game session")
		return err
	}
	if session == nil {
		// this shouldn’t really happen, but guard anyway
		c.log.Error().Str("session_id", sessionID).Msg("Game session not found")
		return err
	}

	// get leetcode username associated with userID
	lcUsername, err := store.DataStore.GetLCUsername(userID)
	if err != nil {
		c.log.Error().Err(err).Int64("user_id", userID).Msg("Failed to get user data")
		return err
	}
	if lcUsername == "" {
		c.log.Error().Int64("user_id", userID).Msg("No LeetCode username associated with user")
		return fmt.Errorf("no LeetCode username associated with user, cannot validate submission")
	}

	cfg := config.GetConfig()
	if cfg.SUBMISSION_VALIDATION && p.Status == models.Accepted {
		lastSubmission, err := services.GetLastAcceptedSubmission(lcUsername)
		if err != nil {
			c.log.Error().Err(err).Int64("user_id", userID).Msg("Failed to get last accepted submission")
			return err
		}
		if lastSubmission == nil {
			c.log.Warn().Int64("user_id", userID).Msg("No accepted submissions found for user")
			return fmt.Errorf("no accepted submissions found for user, cannot validate submission")
		}
		if lastSubmission.SubmissionID != p.ID {
			c.log.Warn().Int64("user_id", userID).Msg("Submission ID does not match last accepted submission")
			return fmt.Errorf("submission ID does not match last accepted submission, cannot validate submission")
		}

		if lastSubmission.TitleSlug != session.Problem.Slug {
			c.log.Warn().
				Int64("user_id", userID).
				Str("expected_slug", session.Problem.Slug).
				Str("actual_slug", lastSubmission.TitleSlug).
				Msg("Submission problem slug does not match game problem")
			return fmt.Errorf("submission problem slug does not match game problem, cannot validate submission")
		}

		p.Time = lastSubmission.Timestamp
	}

	// todo: need to check if opponent made a submission first before declaring victor

	submissionID := p.ID
	submission := models.PlayerSubmission{
		ID:              submissionID,
		PlayerID:        userID,
		PassedTestCases: p.PassedTestCases,
		TotalTestCases:  p.TotalTestCases,
		Status:          p.Status,
		Runtime:         p.Runtime,
		Memory:          p.Memory,
		Lang:            p.Language,
		Time:            p.Time,
	}

	err = services.GameManager.AddSubmission(sessionID, submission)
	if err != nil {
		c.log.Error().Err(err).Int64("user_id", userID).Msg("Failed to add submission")
		return err
	}

	opponentID, err := services.GameManager.GetOpponent(sessionID, userID)
	if err != nil {
		c.log.Error().Err(err).Int64("user_id", userID).Msg("Failed to get opponent")
		return err
	}

	if p.Status == models.Accepted {
		session, err = services.GameManager.CompleteGame(sessionID, userID)
		if err != nil {
			c.log.Error().Err(err).Int64("user_id", userID).Msg("Failed to complete game")
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
			c.log.Error().Err(err).Int64("user_id", userID).Msg("Failed to send game over message to user")
			return err
		}
		err = ConnManager.SendToUser(opponentID, b)
		if err != nil {
			c.log.Error().Err(err).Int64("user_id", opponentID).Msg("Failed to send game over message to opponent")
			return err
		}

		err = store.DataStore.StoreMatch(session)
		if err != nil {
			c.log.Error().Err(err).Msg("Failed to store match data")
			return err
		}
		return nil
	}

	reply := OpponentSubmissionPayload{
		ID:       p.ID,
		PlayerID: userID,
		Status:   p.Status,
		Language: p.Language,
		Time:     p.Time,
	}

	payload, _ := json.Marshal(reply)
	msg := Message{Type: ServerMsgOpponentSubmission, Payload: payload}
	b, _ := json.Marshal(msg)
	err = ConnManager.SendToUser(opponentID, b)
	if err != nil {
		c.log.Error().Err(err).Int64("user_id", opponentID).Msg("Failed to send game over message to opponent")
		return err
	}

	return nil
}

func (cm *connManager) handleForfeit(userID int64) error {
	cm.log.Info().Int64("user_id", userID).Msg("Processing forfeit request")

	sessionID, err := services.GameManager.GetSessionIDByPlayer(userID)
	if err != nil {
		cm.log.Error().Err(err).Int64("user_id", userID).Msg("Failed to get session ID for forfeiting user")
		return err
	}
	if sessionID == "" {
		cm.log.Warn().Int64("user_id", userID).Msg("User attempted to forfeit but is not in a game")
		return nil
	}

	opponentID, err := services.GameManager.GetOpponent(sessionID, userID)
	if err != nil {
		cm.log.Error().Err(err).Int64("user_id", userID).Str("session_id", sessionID).Msg("Failed to get opponent for forfeit")
		return err
	}
	if opponentID == 0 {
		cm.log.Error().Int64("user_id", userID).Str("session_id", sessionID).Msg("Could not find opponent in session")
		return fmt.Errorf("opponent not found in session %s", sessionID)
	}

	completedSession, err := services.GameManager.CompleteGame(sessionID, opponentID)
	if err != nil {
		cm.log.Error().Err(err).Int64("user_id", userID).Str("session_id", sessionID).Msg("Failed to complete game after forfeit")
		return err
	}
	if completedSession == nil {
		cm.log.Error().Str("session_id", sessionID).Msg("Game session not found by CompleteGame")
		return fmt.Errorf("session %s not found", sessionID)
	}

	duration := completedSession.EndTime.Sub(completedSession.StartTime)
	durationSecs := int64(duration.Seconds())

	reply := GameOverPayload{
		WinnerID:  opponentID,
		SessionID: sessionID,
		Duration:  durationSecs,
	}
	payload, _ := json.Marshal(reply)
	msg := Message{Type: ServerMsgGameOver, Payload: payload}
	b, _ := json.Marshal(msg)

	err = ConnManager.SendToUser(userID, b)
	if err != nil {
		cm.log.Error().Err(err).Int64("user_id", userID).Msg("Failed to send game over message to forfeiting user")
		// Continue to notify opponent
	}

	err = ConnManager.SendToUser(opponentID, b)
	if err != nil {
		cm.log.Error().Err(err).Int64("user_id", opponentID).Msg("Failed to send game over message to opponent (winner)")
		// Continue to storage
	}

	cm.log.Info().Str("session_id", sessionID).Int64("winner_id", opponentID).Int64("loser_id", userID).Msg("Game ended due to forfeit")

	err = store.DataStore.StoreMatch(completedSession)
	if err != nil {
		cm.log.Error().Err(err).Msg("Failed to store match data after forfeit")
		return err
	}
	return nil
}

func (cm *connManager) StoreTicket(ctx context.Context, ticket string, userID int64, ttl time.Duration) error {
	return cm.redisClient.Set(ctx, wsTicketKey(ticket), userID, ttl).Err()
}

func (cm *connManager) ValidateTicket(ctx context.Context, ticket string) (int64, error) {
	key := wsTicketKey(ticket)
	pipe := cm.redisClient.TxPipeline()
	getResult := pipe.Get(ctx, key)
	pipe.Del(ctx, key)

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return 0, fmt.Errorf("redis transaction failed: %w", err)
	}

	userIDStr, err := getResult.Result()
	if err == redis.Nil {
		return 0, fmt.Errorf("ticket not found or already used")
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get ticket from redis: %w", err)
	}

	return strconv.ParseInt(userIDStr, 10, 64)
}
