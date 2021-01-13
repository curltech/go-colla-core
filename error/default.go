package error

const (
	Error_NotFound    = "NotFound"
	Error_NotExist    = "NotExist"
	Error_NoValue     = "NoValue"
	Error_WrongValue  = "WrongValue"
	Error_NoId        = "NoId"
	Error_WrongId     = "WrongId"
	Error_WrongSpecId = "WrongSpecId"
	Error_NoSpecId    = "NoSpecId"
	Error_LoadFail    = "LoadFail"
	Error_NoKind      = "NoKind"
	Error_NoAlias     = "NoAlias"
	Error_RepeatKind  = "RepeatKind"
	Error_RemoveFail  = "RemoveFail"
)

//
//func IrisError(ctx iris.Context,statusCode int, msg string) {
//	ctx.StopWithStatus(statusCode)
//	ctx.JSON(msg)
//}
