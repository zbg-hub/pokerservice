package player

type PlayerInterface interface {
	// 决策
	Decide() PlayerDecision
}
