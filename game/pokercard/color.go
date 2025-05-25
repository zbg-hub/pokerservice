package pokercard

type ColorEnum struct {
	Code int    `json:"code"`
	Desc string `json:"desc"`
}

var (
	HONGTAO  = ColorEnum{0, "♥️"}
	MEIHUA   = ColorEnum{1, "♣️"}
	FANGKUAI = ColorEnum{2, "♦️"}
	HEITAO   = ColorEnum{3, "♠️"}

	codeMap = map[int]ColorEnum{
		0: HONGTAO,
		1: MEIHUA,
		2: FANGKUAI,
		3: HEITAO,
	}
)

func GetColorEnumByCode(code int) ColorEnum {
	if color, exists := codeMap[code]; exists {
		return color
	}
	return ColorEnum{}
}
