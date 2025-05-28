package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"leetcodeduels/models"
	"leetcodeduels/services"
	"leetcodeduels/store"
	"log"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v8"
)

const (
	RedisBroadcastChannel = "pubsubChan"
	userChannelPrefix     = "user:"
)

var ConnManager *connManager

type directMessage struct {
	userID  int64
	payload []byte
}

type connManager struct {
	redisClient *redis.Client
	pubsub      *redis.PubSub

	register   chan *Client
	unregister chan *Client

	// all connected clients on this node
	clients map[*Client]bool
	// connections grouped by userID
	userClients map[int64]map[*Client]bool

	// cluster-wide broadcast
	broadcast chan []byte
	// local direct queue
	direct chan directMessage

	ctx    context.Context
	cancel context.CancelFunc
}

func userChannel(userID int64) string {
	return fmt.Sprintf("%s%d", userChannelPrefix, userID)
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
	ps := client.Subscribe(context.Background(), RedisBroadcastChannel)
	cm := &connManager{
		redisClient: client,
		pubsub:      ps,
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		clients:     make(map[*Client]bool),
		userClients: make(map[int64]map[*Client]bool),
		broadcast:   make(chan []byte, 256),
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

			// per-user registration
			uc, exists := cm.userClients[c.userID]
			if !exists {
				uc = make(map[*Client]bool)
				cm.userClients[c.userID] = uc
				// first conn for this user → subscribe
				err := cm.pubsub.Subscribe(context.Background(), userChannel(c.userID))
				if err != nil {
					log.Printf("redis subscribe error user %d: %v", c.userID, err)
				}
			}
			uc[c] = true

		case c := <-cm.unregister:
			if _, ok := cm.clients[c]; !ok {
				continue
			}
			delete(cm.clients, c)

			if uc, exists := cm.userClients[c.userID]; exists {
				delete(uc, c)
				if len(uc) == 0 {
					// last conn for this user → unsubscribe
					err := cm.pubsub.Unsubscribe(context.Background(), userChannel(c.userID))
					if err != nil {
						log.Printf("redis unsubscribe error user %d: %v", c.userID, err)
					}
					delete(cm.userClients, c.userID)
				}
			}
			close(c.send)

		case msg := <-cm.broadcast:
			for c := range cm.clients {
				select {
				case c.send <- msg:
				default:
					close(c.send)
					delete(cm.clients, c)
				}
			}

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
			switch msg.Channel {
			case RedisBroadcastChannel:
				cm.broadcast <- []byte(msg.Payload)
			default:
				if strings.HasPrefix(msg.Channel, userChannelPrefix) {
					idStr := strings.TrimPrefix(msg.Channel, userChannelPrefix)
					if uid, err := strconv.ParseInt(idStr, 10, 64); err == nil {
						cm.direct <- directMessage{userID: uid, payload: []byte(msg.Payload)}
					}
				}
			}
		}
	}
}

func (h *connManager) SendToUser(userID int64, payload []byte) error {
	return h.redisClient.Publish(context.Background(), userChannel(userID), payload).Err()
}

func (h *connManager) HandleClientMessage(c *Client, env *Message) error {
	switch env.Type {
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
	// stop run() and redisListener()
	cm.cancel()

	for c := range cm.clients {
		cm.unregister <- c
	}

	if err := cm.pubsub.Close(); err != nil {
		log.Printf("error closing pubsub: %v", err)
	}

	return cm.redisClient.Close()
}

func (c *connManager) handleSendInvitation(
	userID int64, p SendInvitationPayload,
) error {
	if len(c.userClients[p.InviteeID]) == 0 {
		b, _ := json.Marshal(Message{Type: ServerMsgUserOffline})
		err := ConnManager.SendToUser(userID, b)
		if err != nil {
			return err
		}
		return nil
	}

	success, err := services.InviteManager.CreateInvite(userID, p.InviteeID, p.MatchDetails)
	if err != nil {
		return err
	}
	if success == false {
		// Invite already exists from this user
		// TODO: Either replace the existing invite or ignore this invite
		return nil
	}

	request := InvitationRequestPayload{InviterID: userID, MatchDetails: p.MatchDetails}
	payload, _ := json.Marshal(request)

	msg := Message{Type: ServerMsgInvitationRequest, Payload: payload}
	b, _ := json.Marshal(msg)
	err = ConnManager.SendToUser(p.InviteeID, b)
	if err != nil {
		return err
	}

	return nil
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

	problem, err := store.DataStore.GetRandomProblemDuel(invite.MatchDetails.Tags, invite.MatchDetails.Difficulties)
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
	if success == false {
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
	if success == false {
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
