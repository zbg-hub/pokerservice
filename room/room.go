package room

import (
	"github.com/gorilla/websocket"
	"pokerservice/game"
	"pokerservice/player"
)

type Room struct {
	RoomId  string                    `json:"roomId"`
	Players map[string]*player.Player `json:"players"`
	Banker  game.Banker               `json:"-"`
}

func InitNewRoom(playerName string, ws *websocket.Conn) *Room {
	room := &Room{"123", map[string]*player.Player{}, game.Banker{}}
	room.Players[playerName] = &player.Player{}
	room.Players[playerName].Name = playerName
	room.Players[playerName].Conn = ws
	room.Players[playerName].Status = player.INIT
	return room
}
