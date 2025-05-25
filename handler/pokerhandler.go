package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"pokerservice/config"
	"pokerservice/errno"
	"pokerservice/player"
	"pokerservice/room"
)

type MyHandler struct {
	Conf    *config.Config
	RoomMap map[string]*room.Room
}

type CreateRoomRequest struct {
	UserName       string
	MaxPlayerCount int
}

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

//func (h *MyHandler)StartChat(c *gin.Context, ctx context.Context) {
//	//c.String(http.StatusOK, "ok")
//	fmt.Printf("%s is connected!\n")
//	userName := c.get("userName")
//	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
//	if err != nil {
//		return
//	}
//
//	h.ConnCahe[userName] = ws
//
//	defer func() {
//		h.ConnCahe[userName] = nil
//		fmt.Printf("%s is disconnected\n", userName)
//		ws.Close()
//	}()
//
//	for {
//		//读取ws中的数据
//		mt, messageByte, err := ws.ReadMessage()
//		if err != nil {
//			break
//		}
//		fmt.Printf("message is %s", messageByte)
//		newMessage := &message.Message{}
//		err = json.Unmarshal(messageByte, newMessage)
//		if err != nil {
//			fmt.Printf("unmarshal err", err)
//		} else {
//			fmt.Printf("from %s message is %v\n", account, newMessage)
//		}
//
//		if h.ConnCahe[newMessage.To] != nil {
//			writeConn := h.ConnCahe[newMessage.To]
//			go message.EmitMessage(mt, newMessage, writeConn)
//		}
//
//	}
//}

func (h *MyHandler) CreateNewRoom(c *gin.Context, ctx context.Context) errno.Payload {
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("upgrade fail")
		return errno.InternalError(5000, "upgrade fail")
	}
	userName := c.Query("user")
	curRoom := room.InitNewRoom(userName, ws)
	h.RoomMap[curRoom.RoomId] = curRoom
	// 启动协程处理消息
	//go h.handleWebSocketConnection(room.RoomId, userName.(string))
	message := [][]byte{}
	data, _ := json.Marshal(errno.Payload{2000, "message", errno.ROOM_CREATED, curRoom})
	message = append(message, data)
	fmt.Printf("curRoom is %+v", curRoom)
	go h.sendMessage(curRoom.RoomId, userName, message)
	return errno.OK(nil)
}

func (h *MyHandler) JoinRoom(c *gin.Context, ctx context.Context) errno.Payload {
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("upgrade fail")
		return errno.Payload{5000, "upgrade fail", "", nil}
	}
	roomId := c.Query("room_id")
	userName := c.Query("user")
	curRoom := h.RoomMap[roomId]
	curRoom.Players[userName] = &player.Player{}
	curRoom.Players[userName].Name = userName
	curRoom.Players[userName].Conn = ws
	curRoom.Players[userName].Status = player.INIT
	message := [][]byte{}
	joinRoomResp, _ := json.Marshal(errno.Payload{2000, "", errno.ROOM_JOINED, curRoom})
	updatePlayerMessage, _ := json.Marshal(errno.Payload{2000, "", errno.UPDATE_PLAYER_INFO, curRoom})
	message = append(message, joinRoomResp, updatePlayerMessage)
	for _, player := range curRoom.Players {
		go h.sendMessage(curRoom.RoomId, player.Name, message)
	}
	return errno.OK(nil)
}

func (h *MyHandler) ReadyForGame(c *gin.Context, ctx context.Context) errno.Payload {
	roomId := c.Query("room_id")
	userName := c.Query("user")
	curRoom := h.RoomMap[roomId]
	curRoom.Players[userName].Status = player.READY
	message := [][]byte{}
	updatePlayerMessage, _ := json.Marshal(errno.Payload{2000, "", errno.UPDATE_PLAYER_INFO, curRoom})
	message = append(message, updatePlayerMessage)
	for _, player := range curRoom.Players {
		go h.sendMessage(curRoom.RoomId, player.Name, message)
	}
	return errno.OK(nil)
}

func (h *MyHandler) StartNewGame(c *gin.Context, ctx context.Context) errno.Payload {
	roomId := c.Query("room_id")
	for playerId, curPlayer := range h.RoomMap[roomId].Players {
		// 如果有人没有准备就踢走
		if curPlayer.Status != player.READY {
			// 1. 从房间玩家列表中移除
			err := curPlayer.Conn.Close()
			if err != nil {
				fmt.Printf("close connect error")
			}
			delete(h.RoomMap[roomId].Players, playerId)
		}
	}
	playerList := []*player.Player{}
	for _, curPlayer := range h.RoomMap[roomId].Players {
		playerList = append(playerList, curPlayer)
	}
	h.RoomMap[roomId].Banker.InitNewPlay(playerList, 50, 0)
	go h.RoomMap[roomId].Banker.StartPlayRound()
	return errno.Payload{2000, "", "", nil}
}

func (h *MyHandler) StartNewRound(c *gin.Context, ctx context.Context) errno.Payload {
	roomId := c.Query("room_id")
	curPlayer := h.RoomMap[roomId].Banker.GameContext.SmallBlind
	deleteCount := 0
	for i := 0; i < h.RoomMap[roomId].Banker.GameContext.PlayerCount; i++ {
		// 如果有人没筹码了就踢走
		if curPlayer.NextPlayer.Chips == 0 {
			deleteCount++
			// 1. 从房间玩家列表中移除
			err := curPlayer.NextPlayer.Conn.Close()
			if err != nil {
				fmt.Printf("close connect error")
			}
			delete(h.RoomMap[roomId].Players, curPlayer.NextPlayer.Name)
			if curPlayer.NextPlayer == h.RoomMap[roomId].Banker.GameContext.SmallBlind {
				h.RoomMap[roomId].Banker.GameContext.SmallBlind = curPlayer.NextPlayer.NextPlayer
			}
			curPlayer.NextPlayer = curPlayer.NextPlayer.NextPlayer
		}
	}
	h.RoomMap[roomId].Banker.GameContext.PlayerCount -= deleteCount
	go h.RoomMap[roomId].Banker.StartPlayRound()
	return errno.Payload{2000, "", "", nil}
}

func (h *MyHandler) handleWebSocketConnection(roomId string, userName string) {

}

func (h *MyHandler) sendMessage(roomId, userName string, messages [][]byte) {
	curRoom := h.RoomMap[roomId]
	for _, message := range messages {
		err := curRoom.Players[userName].Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			fmt.Printf("send message error, user is %s", err)
		}
	}
}
