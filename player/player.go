package player

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"pokerservice/game/pokercard"
)

var (
	// 状态
	INIT  = "INIT"
	READY = "READY"
)

type Player struct {
	// 玩家名称
	Name string `json:"name"`

	// 每位玩家的长连接
	Conn *websocket.Conn `json:"-"`

	// 状态
	Status string `json:"status"`

	// 筹码
	Chips int `json:"chips"`

	// 下一个玩家
	NextPlayer *Player `json:"-"`

	// 与自己相关的奖池
	Pot int `json:"pot"`

	// 下注决策列表：下注轮：该轮决策列表
	DecisionList [][]PlayerDecision `json:"decisionList"`

	// 当前下注轮下注
	CurBet int `json:"curBet"`

	// 本轮游戏下注
	Bet int `json:"bet"`

	// 结算奖励
	Bonus int `json:"bonus"`

	// 手牌
	HoleCards []pokercard.PokerCard `json:"-"`

	// 摊牌后手牌
	ViewHoleCards []pokercard.PokerCard `json:"viewHoleCards"`

	// 总资产
	TotalAssets int `json:"totalAssets"`

	// 总负债
	TotalLiabilities int `json:"totalLiabilities"`

	// 牌力值
	CardStrength int `json:"cardStrength"`
}

func (p *Player) Decide() (*PlayerDecision, error) {
	ws := p.Conn
	//读取ws中的数据
	_, messageByte, err := ws.ReadMessage()
	if err != nil {
		fmt.Printf("read message err is +%v", err)
		return nil, err
	}
	fmt.Printf("message is %s\n", messageByte)
	decision := &PlayerDecision{}
	err = json.Unmarshal(messageByte, decision)
	if err != nil {
		fmt.Printf("unmarshal err", err)
		return nil, err
	}
	return decision, nil
}

func (p *Player) GetLastDecision() *PlayerDecision {
	if len(p.DecisionList) == 0 {
		return nil
	}
	return &p.DecisionList[len(p.DecisionList)-1][len(p.DecisionList[len(p.DecisionList)-1])-1]
}
