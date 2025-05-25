package game

import (
	"fmt"
	"pokerservice/game/pokercard"
	"testing"
)

func TestGetFrequencyAndValue(t *testing.T) {
	cards1 := []pokercard.PokerCard{{pokercard.HONGTAO, pokercard.FIVE}, {pokercard.FANGKUAI, pokercard.TEN}, {pokercard.HONGTAO, pokercard.TEN}, {pokercard.MEIHUA, pokercard.TEN}, {pokercard.HEITAO, pokercard.TEN}}
	cards2 := []pokercard.PokerCard{{pokercard.FANGKUAI, pokercard.TEN}, {pokercard.HONGTAO, pokercard.TEN}, {pokercard.MEIHUA, pokercard.TEN}, {pokercard.HEITAO, pokercard.TEN}, {pokercard.HONGTAO, pokercard.J}}
	cards3 := []pokercard.PokerCard{{pokercard.FANGKUAI, pokercard.TEN}, {pokercard.HONGTAO, pokercard.TEN}, {pokercard.MEIHUA, pokercard.TEN}, {pokercard.HEITAO, pokercard.J}, {pokercard.HONGTAO, pokercard.J}}
	cards4 := []pokercard.PokerCard{{pokercard.FANGKUAI, pokercard.TEN}, {pokercard.HONGTAO, pokercard.TEN}, {pokercard.MEIHUA, pokercard.J}, {pokercard.HEITAO, pokercard.J}, {pokercard.HONGTAO, pokercard.J}}
	cards5 := []pokercard.PokerCard{{pokercard.FANGKUAI, pokercard.TEN}, {pokercard.HONGTAO, pokercard.TEN}, {pokercard.MEIHUA, pokercard.TEN}, {pokercard.HEITAO, pokercard.J}, {pokercard.HONGTAO, pokercard.Q}}
	cards6 := []pokercard.PokerCard{{pokercard.FANGKUAI, pokercard.TEN}, {pokercard.HONGTAO, pokercard.TEN}, {pokercard.MEIHUA, pokercard.J}, {pokercard.HEITAO, pokercard.J}, {pokercard.HONGTAO, pokercard.Q}}
	cards7 := []pokercard.PokerCard{{pokercard.HONGTAO, pokercard.FIVE}, {pokercard.FANGKUAI, pokercard.FIVE}, {pokercard.HONGTAO, pokercard.SIX}, {pokercard.MEIHUA, pokercard.EIGHT}, {pokercard.HEITAO, pokercard.TEN}}
	cards8 := []pokercard.PokerCard{{pokercard.HONGTAO, pokercard.TWO}, {pokercard.FANGKUAI, pokercard.FOUR}, {pokercard.HONGTAO, pokercard.SIX}, {pokercard.MEIHUA, pokercard.EIGHT}, {pokercard.HEITAO, pokercard.TEN}}
	cardss := [][]pokercard.PokerCard{cards1, cards2, cards3, cards4, cards5, cards6, cards7, cards8}
	for _, cards := range cardss {
		maxFrequency, maxFrequencyValue, secFrequency, secFrequencyValue := getFrequencyAndValue(cards)
		fmt.Printf("maxFrequency = %v, maxFrequencyValue = %v, secFrequency = %v, secFrequencyValue = %v\n", maxFrequency, maxFrequencyValue, secFrequency, secFrequencyValue)
	}
}

func TestGetBestHandStrength(t *testing.T) {
	b := Banker{}
	b.GameContext = &GameContext{}
	b.GameContext.CommunityCards = []pokercard.PokerCard{{pokercard.HONGTAO, pokercard.FIVE}, {pokercard.FANGKUAI, pokercard.TWO}, {pokercard.HEITAO, pokercard.TEN}, {pokercard.FANGKUAI, pokercard.K}, {pokercard.HEITAO, pokercard.EIGHT}}
	holeCards1 := []pokercard.PokerCard{{pokercard.MEIHUA, pokercard.SEVEN}, {pokercard.HONGTAO, pokercard.Q}}
	holeCards2 := []pokercard.PokerCard{{pokercard.HONGTAO, pokercard.NINE}, {pokercard.HONGTAO, pokercard.TEN}}
	allCombinations := generateCombinations(append(append([]pokercard.PokerCard{}, holeCards1...), b.GameContext.CommunityCards...))
	for _, combination := range allCombinations {
		fmt.Printf("combination is %s, strength = %+v\n", pokercard.PrintPokerCards(combination), evaluateStrength(combination))
	}
	conbination1, strength1 := b.getBestHandStrength(holeCards1)
	conbination2, strength2 := b.getBestHandStrength(holeCards2)
	fmt.Printf("conbination1 = %+v, strength1 = %+v, conbination2 = %+v, strength2 = %+v", pokercard.PrintPokerCards(conbination1), strength1, pokercard.PrintPokerCards(conbination2), strength2)
}
