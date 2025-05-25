package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"pokerservice/config"
	"pokerservice/handler"
	"pokerservice/middleware"
	"pokerservice/room"
	"time"
)

func main() {
	conf := config.GetConfig()
	roomMap := map[string]*room.Room{}
	pokerHandler := &handler.MyHandler{conf, roomMap}
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.GET("/create_room", middleware.Response(pokerHandler.CreateNewRoom))
	r.GET("/join_room", middleware.Response(pokerHandler.JoinRoom))
	r.GET("/ready", middleware.Response(pokerHandler.ReadyForGame))
	r.GET("/start", middleware.Response(pokerHandler.StartNewGame))
	r.GET("/start_new_round", middleware.Response(pokerHandler.StartNewRound))
	addr := ":" + conf.ServicePort
	if err := r.Run(addr); err != nil {
		fmt.Printf("Run server error: %v", err)
	}
}
