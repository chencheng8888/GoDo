package response

const (
	SuccessCode = 200
	SuccessMsg  = "success"
)

const (
	InvalidRequestCode = 4000 + iota
	FileNotUploadedCode
	FileSaveFailedCode
	DeleteTaskFailedCode
	AuthorizationHeaderRequiredCode
	BearerRequiredCode
	InvalidTokenCode
	LoginFailedCode
	SignTokenFailed
)

const (
	InvalidRequestMsg              = "invalid request"
	FileNotUploadedMsg             = "file not uploaded"
	FileSaveFailedMsg              = "file save failed"
	DeleteTaskFailedMsg            = "delete task failed"
	AuthorizationHeaderRequiredMsg = "Authorization header required"
	BearerRequiredMsg              = "Authorization header format must be Bearer <token>"
	InvalidTokenMsg                = "Invalid or expired token"
	LoginFailedMsg                 = "login failed"
	SignTokenMsg                   = "sign token failed"
)
