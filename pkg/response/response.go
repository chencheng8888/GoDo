package response

const (
	SuccessCode = 200
	SuccessMsg  = "success"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func Success(data interface{}) *Response {
	return &Response{
		Code: SuccessCode,
		Msg:  SuccessMsg,
		Data: data,
	}
}

func Error(code int, msg string) *Response {
	return &Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	}
}
