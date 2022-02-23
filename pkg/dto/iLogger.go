package dto

type ILogger interface {
	LogPanic(Fields map[string]interface{}, err error, msg string)
	LogFatal(Fields map[string]interface{}, err error, msg string)
	LogError(Fields map[string]interface{}, err error, msg string)
	LogWarning(Fields map[string]interface{}, err error, msg string)
	LogInfo(Fields map[string]interface{}, msg string)
	LogDebug(Fields map[string]interface{}, msg string)
}
