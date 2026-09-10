package api

import (
	"github.com/ViniciusVaz/gobid/internal/services"
	"github.com/alexedwards/scs/v2"
	"github.com/gorilla/websocket"

	"github.com/go-chi/chi/v5"
)

type Api struct {
	Router         *chi.Mux
	UserService    services.UserService
	ProductService services.ProductService
	Sessions       *scs.SessionManager
	WsUpgrader     websocket.Upgrader
	AuctionLobby   services.AuctionLobby
	BidService     services.BidsService
}
