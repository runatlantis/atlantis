package metrics

import "github.com/uber-go/tally"

func InitCounter(scope tally.Scope, name string) {
	s := scope.Counter(name)
	s.Inc(0)
}
