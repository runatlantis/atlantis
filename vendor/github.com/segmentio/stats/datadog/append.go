package datadog

import (
	"strconv"
	"strings"

	"github.com/segmentio/stats"
)

func appendMetric(b []byte, m Metric) []byte {
	if len(m.Namespace) != 0 {
		b = append(b, m.Namespace...)
		b = append(b, '.')
	}

	b = append(b, m.Name...)
	b = append(b, ':')
	b = strconv.AppendFloat(b, m.Value, 'g', -1, 64)
	b = append(b, '|')
	b = append(b, m.Type...)

	if m.Rate != 0 && m.Rate != 1 {
		b = append(b, '|', '@')
		b = strconv.AppendFloat(b, m.Rate, 'g', -1, 64)
	}

	if n := len(m.Tags); n != 0 {
		b = append(b, '|', '#')
		b = appendTags(b, m.Tags)
	}

	return append(b, '\n')
}

func appendEvent(b []byte, e Event) []byte {
	b = append(b, '_', 'e', '{')
	b = strconv.AppendInt(b, int64(len(e.Title)), 10)
	b = append(b, ',')
	b = strconv.AppendInt(b, int64(len(e.Text)), 10)
	b = append(b, '}', ':')
	b = append(b, e.Title...)
	b = append(b, '|')

	b = append(b, strings.Replace(e.Text, "\n", "\\n", -1)...)

	if e.Priority != EventPriorityNormal {
		b = append(b, '|', 'p', ':')
		b = append(b, e.Priority...)
	}

	if e.AlertType != EventAlertTypeInfo {
		b = append(b, '|', 't', ':')
		b = append(b, e.AlertType...)
	}

	if e.Ts != int64(0) {
		b = append(b, '|', 'd', ':')
		b = strconv.AppendInt(b, e.Ts, 10)
	}

	if len(e.Host) > 0 {
		b = append(b, '|', 'h', ':')
		b = append(b, e.Host...)
	}

	if len(e.AggregationKey) > 0 {
		b = append(b, '|', 'k', ':')
		b = append(b, e.AggregationKey...)
	}

	if len(e.SourceTypeName) > 0 {
		b = append(b, '|', 's', ':')
		b = append(b, e.SourceTypeName...)
	}

	if n := len(e.Tags); n != 0 {
		b = append(b, '|', '#')
		b = appendTags(b, e.Tags)
	}

	return append(b, '\n')
}

func appendTags(b []byte, tags []stats.Tag) []byte {
	for i, t := range tags {
		if i != 0 {
			b = append(b, ',')
		}

		b = append(b, t.Name...)
		b = append(b, ':')
		b = append(b, t.Value...)
	}
	return b
}
