package pegomock

type FailHandler func(message string, callerSkip ...int)

type Mock interface {
	SetFailHandler(FailHandler)
	FailHandler() FailHandler
}
type Param interface{}
type ReturnValue interface{}
type ReturnValues []ReturnValue

type Option interface{ Apply(Mock) }

type OptionFunc func(mock Mock)

func (f OptionFunc) Apply(mock Mock) { f(mock) }

func WithFailHandler(fail FailHandler) Option {
	return OptionFunc(func(mock Mock) { mock.SetFailHandler(fail) })
}
