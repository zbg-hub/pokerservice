package pokercard

type PokerCard struct {
	Color ColorEnum `json:"color"`
	Value ValueEnum `json:"value"`
}

func IsSameColor(cards []PokerCard) bool {
	for i := 1; i < len(cards); i++ {
		if cards[i].Color != cards[i-1].Color {
			return false
		}
	}
	return true
}

func (p *PokerCard) String() string {
	return p.Color.Desc + p.Value.Desc
}

func PrintPokerCards(cards []PokerCard) string {
	str := ""
	for _, card := range cards {
		str += card.String() + " "
	}
	return str
}
