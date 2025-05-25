package game

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"math"
	"math/rand"
	"pokerservice/errno"
	"pokerservice/game/pokercard"
	"pokerservice/player"
	"sort"
	"time"
)

type Banker struct {
	// 游戏上下文
	GameContext *GameContext
}

type ViewInfo struct {
	// 公共可见
	GameContext *GameContext `json:"gameContext"`

	// 仅自己可见
	HoleCards []pokercard.PokerCard `json:"holeCards"`

	// 玩家序列（每位玩家不同，以自己为起始）
	Players []player.Player `json:"players"`

	// 可选决策，为空代表未到决策时机
	OptionalDecision []player.PlayerDecision `json:"optionalDecision"`

	// 摊牌结算
	Result bool `json:"result"`
}

type LiabilitiesInfo struct {
	NeedLiability bool `json:"needLiability"`
}

func (b *Banker) InitNewPlay(players []*player.Player, initChips int, mode int) {
	b.GameContext = &GameContext{}
	b.initPokerCards()
	for i, curPlayer := range players {
		curPlayer.Chips = initChips
		curPlayer.TotalLiabilities = initChips
		curPlayer.NextPlayer = players[(i+1)%len(players)]
	}
	b.GameContext.PlayerCount = len(players)
	b.GameContext.SmallBlind = players[0]
	b.GameContext.BigBlind = players[1]
}

func (b *Banker) initPokerCards() {
	pokers := []pokercard.PokerCard{}
	for i := 0; i < 4; i++ {
		for j := 2; j < 15; j++ {
			card := pokercard.PokerCard{pokercard.GetColorEnumByCode(i), pokercard.GetValueEnumByCode(j)}
			pokers = append(pokers, card)
		}
	}
	shuffle(pokers)
	b.GameContext.PokerCards = pokers
}

func shuffle(pokers []pokercard.PokerCard) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(pokers), func(i, j int) {
		pokers[i], pokers[j] = pokers[j], pokers[i]
	})
}

func (b *Banker) StartPlayRound() {
	b.initContext()
	// 洗牌
	shuffle(b.GameContext.PokerCards)

	// 发牌
	curPlayer := b.GameContext.SmallBlind
	pokerIndex := 0
	for i := 0; i < 2; i++ {
		for j := 0; j < b.GameContext.PlayerCount; j++ {
			curPlayer.HoleCards = append(curPlayer.HoleCards, b.GameContext.PokerCards[pokerIndex])
			curPlayer = curPlayer.NextPlayer
			pokerIndex++
		}
	}
	// 第一轮下注
	for !b.checkBetOver(0) {
		// 更新状态
		b.sendUpdateInfo(curPlayer)
		decision, err := curPlayer.Decide()
		if err != nil {
			fmt.Printf("game error = %+v", err)
			return
		}
		b.convertDecision(curPlayer, decision)
		b.doPlayerDecision(curPlayer, *decision, 0)
		curPlayer = curPlayer.NextPlayer
	}
	curPlayer = b.GameContext.SmallBlind
	// 发3张公共牌
	b.GameContext.set3CommunityCards(b.GameContext.PokerCards[pokerIndex:])
	b.initBet(1)
	pokerIndex += 3

	// 第二轮下注
	for !b.checkBetOver(1) {
		// 更新状态
		b.sendUpdateInfo(curPlayer)
		decision, err := curPlayer.Decide()
		if err != nil {
			fmt.Printf("game error = %+v", err)
			return
		}
		b.convertDecision(curPlayer, decision)
		b.doPlayerDecision(curPlayer, *decision, 1)
		curPlayer = curPlayer.NextPlayer
	}
	curPlayer = b.GameContext.SmallBlind
	// 转牌
	b.GameContext.CommunityCards = append(b.GameContext.CommunityCards, b.GameContext.PokerCards[pokerIndex])
	b.initBet(2)
	pokerIndex++

	// 第三轮下注
	for !b.checkBetOver(2) {
		// 更新状态
		b.sendUpdateInfo(curPlayer)
		decision, err := curPlayer.Decide()
		if err != nil {
			fmt.Printf("game error = %+v", err)
			return
		}
		b.convertDecision(curPlayer, decision)
		b.doPlayerDecision(curPlayer, *decision, 2)
		curPlayer = curPlayer.NextPlayer
	}
	curPlayer = b.GameContext.SmallBlind
	// 河牌
	b.GameContext.CommunityCards = append(b.GameContext.CommunityCards, b.GameContext.PokerCards[pokerIndex])
	b.initBet(3)
	pokerIndex++

	// 第四轮下注
	for !b.checkBetOver(3) {
		// 更新状态
		b.sendUpdateInfo(curPlayer)
		decision, err := curPlayer.Decide()
		if err != nil {
			fmt.Printf("game error = %+v", err)
			return
		}
		b.convertDecision(curPlayer, decision)
		b.doPlayerDecision(curPlayer, *decision, 3)
		curPlayer = curPlayer.NextPlayer
	}
	curPlayer = b.GameContext.SmallBlind
	// 摊牌结算
	b.getWinner()
	b.sendUpdateInfo(nil)
	for i := 0; i < b.GameContext.PlayerCount; i++ {
		go b.getLiabilities(curPlayer)
		curPlayer = curPlayer.NextPlayer
	}
}

func (b *Banker) sendUpdateInfo(curDecidePlayer *player.Player) {
	curPlayer := b.GameContext.SmallBlind
	result := false
	if curDecidePlayer == nil {
		result = true
	}
	for i := 0; i < b.GameContext.PlayerCount; i++ {
		viewInfo := ViewInfo{}
		viewInfo.GameContext = b.GameContext
		viewInfo.Result = result
		if curDecidePlayer != nil && curPlayer == curDecidePlayer {
			// 构造决策列表
			decisionList := b.makeDecisionList(curPlayer)
			viewInfo.OptionalDecision = decisionList
		}
		players := []player.Player{}
		for j := 0; j < b.GameContext.PlayerCount; j++ {
			players = append(players, *curPlayer)
			curPlayer = curPlayer.NextPlayer
		}
		viewInfo.Players = players
		viewInfo.HoleCards = curPlayer.HoleCards
		payload := errno.Payload{2000, "", errno.UPDATE_GAME_CONTEXT, viewInfo}
		message, _ := json.Marshal(payload)
		go sendMessage(curPlayer.Conn, message)
		curPlayer = curPlayer.NextPlayer
	}
}

func sendMessage(conn *websocket.Conn, message []byte) {
	conn.WriteMessage(websocket.TextMessage, message)
}

func (b *Banker) checkBetOver(betRound int) bool {
	curPlayer := b.GameContext.SmallBlind
	for i := 0; i < b.GameContext.PlayerCount; i++ {
		if curPlayer.GetLastDecision() == nil {
			fmt.Println("GetLastDecision is nil")
			return false
		}
		if curPlayer.GetLastDecision().DecisionEnum == player.ALLIN || curPlayer.GetLastDecision().DecisionEnum == player.FOLD {
			curPlayer = curPlayer.NextPlayer
			continue
		}
		if len(curPlayer.DecisionList) < betRound+1 {
			fmt.Printf("betRound = %+v, curPlayer = %+v", betRound, curPlayer.Name)
			return false
		}
		if curPlayer.CurBet < b.GameContext.MaxBet {
			fmt.Println("curPlayer.CurBet < b.GameContext.MaxBet")
			return false
		}
		curPlayer = curPlayer.NextPlayer
	}
	return true
}

func (b *Banker) makeDecisionList(curPlayer *player.Player) (decisionList []player.PlayerDecision) {
	// 如果最近一次决策是弃牌或者allin，接下来就不用参与决策了
	if curPlayer.GetLastDecision() != nil && (curPlayer.GetLastDecision().DecisionEnum == player.FOLD || curPlayer.GetLastDecision().DecisionEnum == player.ALLIN) {
		return []player.PlayerDecision{{player.NULL, 0}}
	}
	decisionList = []player.PlayerDecision{}
	allinDecision := player.PlayerDecision{player.ALLIN, curPlayer.Chips}
	defer func() {
		decisionList = append(decisionList, allinDecision)
	}()
	// 没有任何人下注，小盲位轮次
	if curPlayer == b.GameContext.SmallBlind && curPlayer.Bet == 0 {
		smallBlindDecision := player.PlayerDecision{player.BLIND, 1}
		decisionList = append(decisionList, smallBlindDecision)
		return decisionList
	}
	// 大盲位且自己没下过注
	if curPlayer == b.GameContext.BigBlind && curPlayer.Bet == 0 {
		bigBlindDecision := player.PlayerDecision{player.BLIND, int(math.Min(float64(b.GameContext.MaxBet*2), float64(curPlayer.Chips)))}
		decisionList = append(decisionList, bigBlindDecision)
		return decisionList
	}
	// 除大小盲位，其他人均可弃牌
	foldDecision := player.PlayerDecision{player.FOLD, 0}
	decisionList = append(decisionList, foldDecision)
	// 当前下注轮无人下注，可以check
	if b.GameContext.MaxBet == 0 {
		checkDecision := player.PlayerDecision{player.CHECK, 0}
		decisionList = append(decisionList, checkDecision)
	}
	// 可以跟注（筹码数额超过跟注数额）
	if curPlayer.Chips >= b.GameContext.MaxBet-curPlayer.CurBet && b.GameContext.MaxBet > 0 {
		callDecision := player.PlayerDecision{player.CALL, b.GameContext.MaxBet - curPlayer.CurBet}
		decisionList = append(decisionList, callDecision)
	}
	// 可以加注
	if curPlayer.Chips > b.GameContext.MaxBet-curPlayer.CurBet {
		raiseDecision := player.PlayerDecision{player.RAISE, b.GameContext.MaxBet - curPlayer.CurBet + 1}
		decisionList = append(decisionList, raiseDecision)
	}
	return decisionList
}

func (b *Banker) doPlayerDecision(curPlayer *player.Player, decision player.PlayerDecision, betRound int) {
	if decision.DecisionEnum == player.NULL {
		return
	}
	if curPlayer.Chips < decision.Value {
		panic("下注金额错误")
	}
	switch decision.DecisionEnum {
	case player.BLIND:
		if curPlayer.CurBet != 0 {
			panic("盲注错误")
		}
		break
	case player.CALL:
		if curPlayer.CurBet+decision.Value != b.GameContext.MaxBet {
			panic("跟注金额错误")
		}
		break
	case player.RAISE:
		if curPlayer.CurBet+decision.Value < b.GameContext.MaxBet {
			panic("加注金额错误")
		}
	case player.ALLIN:
		if curPlayer.Chips != decision.Value {
			panic("allin金额错误")
		}
	case player.CHECK:
	case player.FOLD:
		break
	default:
		panic("决策类型错误")
	}
	b.GameContext.MaxBet = int(math.Max(float64(curPlayer.CurBet+decision.Value), float64(b.GameContext.MaxBet)))
	b.GameContext.Pot += decision.Value
	curPlayer.CurBet += decision.Value
	curPlayer.Bet += decision.Value
	curPlayer.Chips -= decision.Value
	if len(curPlayer.DecisionList) < betRound+1 {
		curPlayer.DecisionList = append(curPlayer.DecisionList, []player.PlayerDecision{})
	}
	curPlayer.DecisionList[betRound] = append(curPlayer.DecisionList[betRound], decision)
}

func (b *Banker) convertDecision(curPlayer *player.Player, decision *player.PlayerDecision) {
	if decision.Value+curPlayer.CurBet == b.GameContext.MaxBet && b.GameContext.MaxBet > 0 {
		decision.DecisionEnum = player.CALL
	}
	if decision.Value == curPlayer.Chips {
		decision.DecisionEnum = player.ALLIN
	}
}

func (b *Banker) initBet(betRound int) {
	curPlayer := b.GameContext.SmallBlind
	for i := 0; i < b.GameContext.PlayerCount; i++ {
		curPlayer.CurBet = 0
		curPlayer = curPlayer.NextPlayer
	}
	b.GameContext.BetRound = betRound
	b.GameContext.MaxBet = 0
}

func (b *Banker) getWinner() {
	curPlayer := b.GameContext.SmallBlind
	strengthMap := map[int][]*player.Player{}
	unFoldPlayerCount := 0
	for i := 0; i < b.GameContext.PlayerCount; i++ {
		fmt.Printf("curPlayer is %s, decisionList is %+v, lastDecision is %+v\n", curPlayer.Name, curPlayer.DecisionList, curPlayer.GetLastDecision())
		if curPlayer.GetLastDecision().DecisionEnum == player.FOLD {
			curPlayer = curPlayer.NextPlayer
			continue
		}
		unFoldPlayerCount++
		// 计算和自己相关的奖池大小
		playerPoint := curPlayer
		pot := 0
		for j := 0; j < b.GameContext.PlayerCount; j++ {
			pot += int(math.Min(float64(playerPoint.Bet), float64(curPlayer.Bet)))
			playerPoint = playerPoint.NextPlayer
		}
		curPlayer.Pot = pot
		_, maxStrength := b.getBestHandStrength(curPlayer.HoleCards)
		curPlayer.CardStrength = maxStrength
		if _, exist := strengthMap[maxStrength]; exist {
			strengthMap[maxStrength] = append(strengthMap[maxStrength], curPlayer)
		} else {
			strengthMap[maxStrength] = []*player.Player{curPlayer}
		}
		curPlayer = curPlayer.NextPlayer
	}
	fmt.Printf("unFoldPlayerCount = %+v\n", unFoldPlayerCount)
	if unFoldPlayerCount > 1 { // 如果最后只剩下一个玩家，不需要摊牌
		for i := 0; i < b.GameContext.PlayerCount; i++ {
			if curPlayer.GetLastDecision().DecisionEnum != player.FOLD {
				curPlayer.ViewHoleCards = curPlayer.HoleCards
				fmt.Printf("curPlayer = %s, ViewHoleCards = %+v\n", curPlayer.Name, curPlayer.ViewHoleCards)
			}
			curPlayer = curPlayer.NextPlayer
		}
	}
	// 分奖池
	assignedPot := 0
	strengthMapKey := []int{}
	for key := range strengthMap {
		strengthMapKey = append(strengthMapKey, key)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(strengthMapKey)))
	for _, strength := range strengthMapKey {
		winnerList := strengthMap[strength]
		sort.Slice(winnerList, func(i, j int) bool {
			return winnerList[i].Bet <= winnerList[j].Bet
		})
		fmt.Printf("strength = %+v, winnerList = %+v", strength, winnerList)
		unAssignedCount := len(winnerList)
		for _, winner := range winnerList {
			if winner.Pot <= assignedPot {
				continue
			}
			bonus := (winner.Pot - assignedPot) / unAssignedCount
			winner.Chips += bonus
			winner.Bonus = bonus
			assignedPot += bonus
			unAssignedCount -= 1
		}
		if b.GameContext.Pot < assignedPot {
			fmt.Println("分配奖池有误")
			return
		}
		if b.GameContext.Pot == assignedPot {
			fmt.Println("奖池分配完毕")
			break
		}
	}
}

func (b *Banker) getBestHandStrength(holeCards []pokercard.PokerCard) ([]pokercard.PokerCard, int) {
	allCards := []pokercard.PokerCard{}
	allCards = append(append(allCards, holeCards...), b.GameContext.CommunityCards...)
	combinations := generateCombinations(allCards)
	maxStrengthCombination := []pokercard.PokerCard{}
	maxStrength := 0
	for _, combination := range combinations {
		strength := evaluateStrength(combination)
		if strength > maxStrength {
			maxStrength = strength
			maxStrengthCombination = combination
		}
	}
	return maxStrengthCombination, maxStrength
}

func generateCombinations(cards []pokercard.PokerCard) [][]pokercard.PokerCard {
	var result [][]pokercard.PokerCard
	n := len(cards)
	k := 5

	// 边界条件：如果牌数不足5张直接返回空
	if n < k {
		return result
	}

	// 回溯法生成组合
	var backtrack func(start int, path []int)
	backtrack = func(start int, path []int) {
		if len(path) == k {
			combination := make([]pokercard.PokerCard, k)
			for i, idx := range path {
				combination[i] = cards[idx]
			}
			result = append(result, combination)
			return
		}

		// 剩余需要选取的牌数
		remaining := k - len(path)
		// 遍历可选范围（n-remaining保证剩余足够的位置）
		for i := start; i <= n-remaining; i++ {
			backtrack(i+1, append(path, i))
		}
	}

	backtrack(0, []int{})
	return result
}

func evaluateStrength(cards []pokercard.PokerCard) int {
	sort.Slice(cards, func(i, j int) bool {
		return cards[i].Value.Code < cards[j].Value.Code
	})
	baseStrength, mainRank, secondaryRank, thirdRank, fourthRank, fifthRank := 0, 0, 0, 0, 0, 0
	maxFrequency, maxFrequencyValue, secFrequency, secFrequencyValue := getFrequencyAndValue(cards)
	flush := isFlush(cards)
	straight := isStraight(cards)
	if flush && straight {
		baseStrength = 9
		mainRank = getStraightRank(cards)
	} else if maxFrequency == 4 {
		baseStrength = 8
		mainRank = maxFrequencyValue
		secondaryRank = secFrequencyValue
	} else if maxFrequency == 3 && secFrequency == 2 {
		baseStrength = 7
		mainRank = maxFrequencyValue
		secondaryRank = secFrequencyValue
	} else if flush {
		baseStrength = 6
		mainRank = cards[0].Value.Code
		secondaryRank = cards[1].Value.Code
		thirdRank = cards[2].Value.Code
		fourthRank = cards[3].Value.Code
		fifthRank = cards[4].Value.Code
	} else if straight {
		baseStrength = 5
		mainRank = getStraightRank(cards)
	} else if maxFrequency == 3 && secFrequency == 1 {
		baseStrength = 4
		mainRank = maxFrequencyValue
		secondaryRank = secFrequencyValue
		thirdRank = cards[4].Value.Code
	} else if maxFrequency == 2 && secFrequency == 2 {
		baseStrength = 3
		mainRank = maxFrequencyValue
		secondaryRank = secFrequencyValue
		for _, card := range cards {
			if card.Value.Code != maxFrequencyValue && card.Value.Code != secFrequencyValue {
				thirdRank = card.Value.Code
				break
			}
		}
	} else if maxFrequency == 2 && secFrequency == 1 {
		baseStrength = 2
		mainRank = maxFrequencyValue
		secondaryRank = secFrequencyValue
		if cards[0].Value.Code == maxFrequencyValue {
			thirdRank = cards[3].Value.Code
			fourthRank = cards[2].Value.Code
		} else if cards[1].Value.Code == maxFrequencyValue {
			thirdRank = cards[3].Value.Code
			fourthRank = cards[0].Value.Code
		} else {
			thirdRank = cards[1].Value.Code
			fourthRank = cards[0].Value.Code
		}
	} else {
		baseStrength = 1
		mainRank = cards[4].Value.Code
		secondaryRank = cards[3].Value.Code
		thirdRank = cards[2].Value.Code
		fourthRank = cards[1].Value.Code
		fifthRank = cards[0].Value.Code
	}
	return baseStrength*759375 + mainRank*50625 + secondaryRank*3375 + thirdRank*225 + fourthRank*15 + fifthRank
}

func isFlush(cards []pokercard.PokerCard) bool {
	return pokercard.IsSameColor(cards)
}

func isStraight(cards []pokercard.PokerCard) bool {
	valueSet := map[int]struct{}{}
	values := []int{}
	for _, card := range cards {
		if _, exist := valueSet[card.Value.Code]; !exist {
			valueSet[card.Value.Code] = struct{}{}
			values = append(values, card.Value.Code)
		}
	}
	sort.Ints(values)
	if len(values) == 5 && values[4]-values[0] == 4 {
		return true
	}
	return len(values) == 5 && values[3] == 5 && values[4] == 14
}

func getStraightRank(cards []pokercard.PokerCard) int {
	if cards[4].Value.Code == 14 {
		return 1
	}
	return cards[0].Value.Code
}

func getFrequencyAndValue(cards []pokercard.PokerCard) (int, int, int, int) {
	// 1. 统计每个牌值的出现次数
	freq := make(map[int]int)
	for _, card := range cards {
		code := card.Value.Code
		freq[code]++
	}

	// 2. 找出最高频次及对应的最大牌值
	maxFreq := 0
	maxCode := 0
	secFreq := 0
	secCode := 0
	for code, count := range freq {
		// 优先比较频次，频次相同则取更大的code
		if count > maxFreq || (count == maxFreq && code > maxCode) {
			maxFreq = count
			maxCode = code
		}
	}

	for code, count := range freq {
		// 优先比较频次，频次相同则取更大的code
		if (count > secFreq && count < maxFreq) || (count == maxFreq && code < maxCode) {
			secFreq = count
		}
	}

	if secFreq < maxFreq {
		for code, count := range freq {
			if count == secFreq && code > secCode {
				secCode = code
			}
		}
	} else if secFreq == maxFreq {
		for code, count := range freq {
			if count == secFreq && code > secCode && code < maxCode {
				secCode = code
			}
		}
	}

	return maxFreq, maxCode, secFreq, secCode
}

func (b *Banker) initContext() {
	curPlayer := b.GameContext.SmallBlind
	for i := 0; i < b.GameContext.PlayerCount; i++ {
		curPlayer.HoleCards = []pokercard.PokerCard{}
		curPlayer.ViewHoleCards = []pokercard.PokerCard{}
		curPlayer.DecisionList = [][]player.PlayerDecision{}
		curPlayer.CurBet = 0
		curPlayer.Bet = 0
		curPlayer.Pot = 0
		curPlayer.Bonus = 0
		curPlayer.CardStrength = 0
		curPlayer.TotalAssets = curPlayer.Chips - curPlayer.TotalLiabilities
		curPlayer = curPlayer.NextPlayer
	}
	b.GameContext.Pot = 0
	b.GameContext.MaxBet = 0
	b.GameContext.BigBlind = b.GameContext.SmallBlind.NextPlayer.NextPlayer
	b.GameContext.SmallBlind = b.GameContext.SmallBlind.NextPlayer
	b.GameContext.CommunityCards = []pokercard.PokerCard{}
}

func (b *Banker) getLiabilities(player *player.Player) {
	ws := player.Conn
	//读取ws中的数据
	_, messageByte, err := ws.ReadMessage()
	if err != nil {
		fmt.Printf("read message err is +%v\n", err)
		return
	}
	fmt.Printf("message is %s\n", messageByte)
	liabilitiesInfo := &LiabilitiesInfo{}
	err = json.Unmarshal(messageByte, liabilitiesInfo)
	if err != nil {
		fmt.Printf("unmarshal err %+v\n", err)
		return
	}
	if liabilitiesInfo.NeedLiability && player.Chips == 0 {
		player.TotalLiabilities += 50
		player.Chips = 50
	}
	return
}
