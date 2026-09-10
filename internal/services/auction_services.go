package services

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type MessageKind int

const (
	//Requests
	PlaceBid MessageKind = iota

	// Ok/Success
	SuccessfullyPlaceBid

	// Erros
	FailedToPlaceBid

	//Info
	NewBidPlaced
	AuctionFinished
)

type Message struct {
	Message string
	Amount  float64
	Kind    int
	UserId  uuid.UUID
}

type AuctionLobby struct {
	sync.Mutex
	Rooms map[uuid.UUID]*AuctionRoom
}

type AuctionRoom struct {
	Id          uuid.UUID
	Context     context.Context
	Broadcast   chan Message
	Register    chan *Client
	Unregister  chan *Client
	Clients     map[uuid.UUID]*Client
	BidsService *BidsService
}

func NewAuctionRoom(ctx context.Context, id uuid.UUID, BidsService BidsService) *AuctionRoom {
	return &AuctionRoom{
		Id:          id,
		Context:     ctx,
		Broadcast:   make(chan Message),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Clients:     make(map[uuid.UUID]*Client),
		BidsService: &BidsService,
	}
}

func (r *AuctionRoom) registerClient(c *Client) {
	slog.Info("New user conected", "Client", c)
	r.Clients[c.UserId] = c
}

func (r *AuctionRoom) unregisterClient(c *Client) {
	slog.Info("User disconected", "Client", c)
	delete(r.Clients, c.UserId)
}

func (r *AuctionRoom) broadcastMessage(m Message) {
	slog.Info("New message recieved", "RoomId", r.Id, "message", m.Message, "user_id", m.UserId)

	switch m.Kind {
	case int(PlaceBid):
		bid, err := r.BidsService.Placebid(r.Context, r.Id, m.UserId, m.Amount)
		if err != nil {
			if errors.Is(err, ErrBidIsTooLow) {
				if client, ok := r.Clients[m.UserId]; ok {
					client.Send <- Message{
						Kind:    int(FailedToPlaceBid),
						Message: ErrBidIsTooLow.Error(),
					}
				}

				return
			}
		}

		if client, ok := r.Clients[m.UserId]; ok {
			client.Send <- Message{
				Kind:    int(SuccessfullyPlaceBid),
				Message: "Your bid was successfully placed.",
			}
		}

		for id, client := range r.Clients {
			newBidMessage := Message{
				Kind:    int(NewBidPlaced),
				Message: "A new bid was placed",
				Amount:  bid.BidAmount,
			}
			if id == m.UserId {
				continue
			}
			client.Send <- newBidMessage
		}
	}
}

func (r *AuctionRoom) Run() {
	slog.Info("Auction has begun", "auctionId", r.Id)
	defer func() {
		close(r.Broadcast)
		close(r.Register)
		close(r.Unregister)
	}()

	for {
		select {
		case client := <-r.Register:
			r.registerClient(client)
		case client := <-r.Unregister:
			r.unregisterClient(client)
		case message := <-r.Broadcast:
			r.broadcastMessage(message)
		case <-r.Context.Done():
			slog.Info("Auction has ended.", "auctionID", r.Id)
			for _, client := range r.Clients {
				client.Send <- Message{
					Kind:    int(AuctionFinished),
					Message: "Auction has been finished.",
				}
			}
			return
			// send message end of auction
		}
	}
}

type Client struct {
	Room   *AuctionRoom
	Conn   *websocket.Conn
	Send   chan Message
	UserId uuid.UUID
}

func NewClient(room *AuctionRoom, conn *websocket.Conn, userId uuid.UUID) *Client {
	return &Client{
		Room:   room,
		Conn:   conn,
		Send:   make(chan Message, 512),
		UserId: userId,
	}
}
