package response

const (
	SuccessCode = 200
	SuccessMsg  = "success"

	InvalidRequestCode = 4001
	InvalidRequestMsg  = "invalid request"

	FileNotUploadedCode = 4002
	FileNotUploadedMsg  = "file not uploaded"

	FileSaveFailedCode = 4003
	FileSaveFailedMsg  = "file save failed"

	DeleteTaskFailedCode = 4004
	DeleteTaskFailedMsg  = "delete task failed"

	LoginFailedCode = 4005
	LoginFailedMsg  = "login failed"

	UserNotFoundCode = 4006
	UserNotFoundMsg  = "user not found"

	PasswordIncorrectCode = 4007
	PasswordIncorrectMsg  = "password incorrect"

	UnauthorizedCode = 4008
	UnauthorizedMsg  = "unauthorized"

	EmailExistsCode = 4009
	EmailExistsMsg  = "email already exists"

	UserNameExistsCode = 4010
	UserNameExistsMsg  = "username already exists"

	InternalErrorCode = 5000
	InternalErrorMsg  = "internal server error"
)
