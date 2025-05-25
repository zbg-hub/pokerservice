package game

import (
	"pokerservice/game/pokercard"
	"pokerservice/player"
)

type GameContext struct {
	// 大盲位
	BigBlind *player.Player `json:"bigBlind"`

	// 小盲位
	SmallBlind *player.Player `json:"smallBlind"`

	// 底池
	Pot int `json:"pot"`

	// 下注轮数
	BetRound int `json:"betRound"`

	// 当前下注轮最大下注
	MaxBet int `json:"maxBet"`

	// 扑克
	PokerCards []pokercard.PokerCard `json:"-"`

	// 公共牌
	CommunityCards []pokercard.PokerCard `json:"communityCards"`

	// 玩家数量
	PlayerCount int `json:"playerCount"`
}

func (g *GameContext) set3CommunityCards(cards []pokercard.PokerCard) {
	g.CommunityCards = []pokercard.PokerCard{cards[0], cards[1], cards[2]}
}

func (g *GameContext) appendCommunityCard(card pokercard.PokerCard) {
	g.CommunityCards = append(g.CommunityCards, card)
}
