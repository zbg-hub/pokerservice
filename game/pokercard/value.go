package pokercard

type ValueEnum struct {
	Code int    `json:"code"`
	Desc string `json:"desc"`
}

var (
	A     = ValueEnum{14, "A"}
	K     = ValueEnum{13, "K"}
	Q     = ValueEnum{12, "Q"}
	J     = ValueEnum{11, "J"}
	TEN   = ValueEnum{10, "10"}
	NINE  = ValueEnum{9, "9"}
	EIGHT = ValueEnum{8, "8"}
	SEVEN = ValueEnum{7, "7"}
	SIX   = ValueEnum{6, "6"}
	FIVE  = ValueEnum{5, "5"}
	FOUR  = ValueEnum{4, "4"}
	THREE = ValueEnum{3, "3"}
	TWO   = ValueEnum{2, "2"}

	codeValueMap = map[int]ValueEnum{
		14: A,
		13: K,
		12: Q,
		11: J,
		10: TEN,
		9:  NINE,
		8:  EIGHT,
		7:  SEVEN,
		6:  SIX,
		5:  FIVE,
		4:  FOUR,
		3:  THREE,
		2:  TWO,
	}
)

func GetValueEnumByCode(code int) ValueEnum {
	if val, exists := codeValueMap[code]; exists {
		return val
	}
	return ValueEnum{}
}
