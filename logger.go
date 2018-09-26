package rtcp

type LogFunc func(format string, args ...interface{})

func (l LogFunc) Printf(format string, args ...interface{})  {
	l(format, args...)
}

type Logger interface {
	Printf(format string, args ...interface{})
}