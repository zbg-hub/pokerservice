package errno

var (
	MessageOK = "ok"
	OkCode    = 2000
	// 4xx
	AccountParamErr = 10400

	// 5xx
	InternalErr        = 10500
	AccountInternalErr = 10501

	// response type
	ROOM_CREATED = "ROOM_CREATED"

	ROOM_JOINED = "ROOM_JOINED"

	UPDATE_PLAYER_INFO = "UPDATE_PLAYER_INFO"

	UPDATE_GAME_CONTEXT = "UPDATE_GAME_CONTEXT"
)

type Payload struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Type    string      `json:"type"`
	Data    interface{} `json:"data,omitempty"`
}

func OK(data interface{}) Payload {
	return Payload{
		Code:    OkCode,
		Message: MessageOK,
		Data:    data,
	}
}

func InternalError(code int, message string) Payload {
	return Payload{
		Code:    code,
		Message: message,
		Data:    nil,
	}
}
