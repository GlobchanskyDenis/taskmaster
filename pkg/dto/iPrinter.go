package dto

type IPrinter interface {
	Printf(string, ...interface{})
}
