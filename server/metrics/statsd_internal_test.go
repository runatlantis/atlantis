package metrics

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally/v4"
)

type mockReporter struct {
	tally.StatsReporter

	actualName string
}

func (m *mockReporter) ReportCounter(name string, tags map[string]string, value int64) {
	m.actualName = name
}

func TestPointTagReporter(t *testing.T) {
	tests := []struct {
		name string
		tags map[string]string
	}{
		{
			name: "foo.Bar.baz",
			tags: map[string]string{"bar": "baz", "hello": "world"},
		},
		{
			name: "foo",
			tags: map[string]string{"bar": "baz.bar.Bug", "hello": "world-ok"},
		},
	}
	for idx, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			t.Parallel()

			m := &mockReporter{}
			p := customTagReporter{
				StatsReporter: m,
				separator:     ",",
			}
			p.ReportCounter(tt.name, tt.tags, 10)

			actualTags := map[string]string{}
			for _, v := range strings.Split(m.actualName, ",")[1:] {
				ss := strings.Split(v, "=")
				actualTags[ss[0]] = ss[1]
			}

			sanitizedTags := map[string]string{}
			for k, v := range tt.tags {
				sanitizedTags[k] = replaceChars(v)
			}

			assert.Equal(t, sanitizedTags, actualTags)
			assert.True(t, strings.HasPrefix(m.actualName, tt.name))
		})
	}
}
