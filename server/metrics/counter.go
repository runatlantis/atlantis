package metrics

import tally "github.com/uber-go/tally/v4"

func InitCounter(scope tally.Scope, name string) {
	s := scope.Counter(name)
	s.Inc(0)
}
