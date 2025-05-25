package message

import (
	"fmt"
	"github.com/gorilla/websocket"
)

type Message struct {
	From string `json:"from"`
	To string	`json:"to"`
	Content string	`json:"content"`
}

func EmitMessage(messageType int, newMessage *Message, conn *websocket.Conn) {
	messageByte := []byte(newMessage.Content)
	err := conn.WriteMessage(messageType, messageByte)
	if err != nil {
		fmt.Printf("emit message err is %s", err)
	}
}
