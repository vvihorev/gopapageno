package xpath

type Logger interface {
	Printf(string, ...interface{})
}

type nopLogger struct{}

func newNopLogger() *nopLogger {
	return &nopLogger{}
}

func (*nopLogger) Printf(format string, v ...interface{}) {
	//noop
}
