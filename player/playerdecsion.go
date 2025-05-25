package player

var (
	BLIND = DecisionEnum{"BLIND", "盲注"}
	FOLD  = DecisionEnum{"FOLD", "弃牌"}
	CHECK = DecisionEnum{"CHECK", "过牌"}
	CALL  = DecisionEnum{"CALL", "跟注"}
	RAISE = DecisionEnum{"RAISE", "加注"}
	ALLIN = DecisionEnum{"ALLIN", "全押"}
	NULL  = DecisionEnum{"NULL", "空"}
)

type PlayerDecision struct {
	DecisionEnum DecisionEnum `json:"decisionEnum"`
	Value        int          `json:"value"`
}
