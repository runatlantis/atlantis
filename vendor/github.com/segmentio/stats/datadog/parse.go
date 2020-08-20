package datadog

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/segmentio/stats"
)

// Adapted from https://github.com/DataDog/datadog-agent/blob/6789e98a1e41e98700fa1783df62238bb23cb454/pkg/dogstatsd/parser.go#L141
func parseEvent(s string) (e Event, err error) {
	var next = strings.TrimSpace(s)
	var header string
	var rawTitleLen string
	var rawTextLen string
	var titleLen int64
	var textLen int64

	header, next = nextToken(next, ':')
	if len(header) < 7 {
		err = fmt.Errorf("datadog: %#v has a malformed event header", s)
		return
	}

	header = header[3 : len(header)-1] // Strip off '_e{' and '}'

	rawTitleLen, rawTextLen = split(header, ',')

	titleLen, err = strconv.ParseInt(rawTitleLen, 10, 64)
	if err != nil {
		err = fmt.Errorf("datadog: %#v has a malformed title length", s)
		return
	}

	textLen, err = strconv.ParseInt(rawTextLen, 10, 64)
	if err != nil {
		err = fmt.Errorf("datadog: %#v has a malformed text length", s)
		return
	}

	rawTitle := next[:titleLen]
	rawText := next[titleLen+1 : titleLen+1+textLen]
	next = next[titleLen+1+textLen:]

	if len(rawTitle) == 0 {
		err = fmt.Errorf("datadog: %#v has a malformed title", s)
		return
	}

	if len(rawText) == 0 {
		err = fmt.Errorf("datadog: %#v has malformed text", s)
		return
	}

	e = Event{
		Priority:  EventPriorityNormal,
		AlertType: EventAlertTypeInfo,
		Title:     rawTitle,
		Text:      strings.Replace(rawText, "\\n", "\n", -1),
	}

	var tags string

	// metadata
	if len(next) > 1 {
		rawMetadataFields := strings.Split(next[1:], "|")
		for i := range rawMetadataFields {
			switch rawMetadataFields[i][0] {
			case 'd':
				var ts int64
				ts, err = strconv.ParseInt(rawMetadataFields[i][2:], 10, 64)
				if err != nil {
					err = fmt.Errorf("datadog: %#v has a malformed timestamp", s)
					return
				}
				e.Ts = ts
			case 'p':
				e.Priority = EventPriority(rawMetadataFields[i][2:])
			case 'h':
				e.Host = rawMetadataFields[i][2:]
			case 't':
				e.AlertType = EventAlertType(rawMetadataFields[i][2:])
			case 'k':
				e.AggregationKey = rawMetadataFields[i][2:]
			case 's':
				e.SourceTypeName = rawMetadataFields[i][2:]
			case '#':
				tags = rawMetadataFields[i][1:]
			default:
				err = fmt.Errorf("datadog: %#v has unexpected metadata field", s)
				return
			}
		}
	}

	if len(tags) != 0 {
		e.Tags = make([]stats.Tag, 0, count(tags, ',')+1)

		for len(tags) != 0 {
			var tag string

			if tag, tags = nextToken(tags, ','); len(tag) != 0 {
				name, value := split(tag, ':')
				e.Tags = append(e.Tags, stats.T(name, value))
			}
		}
	}

	return
}
func parseMetric(s string) (m Metric, err error) {
	var next = strings.TrimSpace(s)
	var name string
	var val string
	var typ string
	var rate string
	var tags string

	val, next = nextToken(next, '|')
	typ, next = nextToken(next, '|')
	rate, tags = nextToken(next, '|')
	name, val = split(val, ':')

	if len(name) == 0 {
		err = fmt.Errorf("datadog: %#v is missing a metric name", s)
		return
	}

	if len(val) == 0 {
		err = fmt.Errorf("datadog: %#v is missing a metric value", s)
		return
	}

	if len(typ) == 0 {
		err = fmt.Errorf("datadog: %#v is missing a metric type", s)
		return
	}

	if len(rate) != 0 {
		switch rate[0] {
		case '#': // no sample rate, just tags
			rate, tags = "", rate
		case '@':
			rate = rate[1:]
		default:
			err = fmt.Errorf("datadog: %#v has a malformed sample rate", s)
			return
		}
	}

	if len(tags) != 0 {
		switch tags[0] {
		case '#':
			tags = tags[1:]
		default:
			err = fmt.Errorf("datadog: %#v has malformed tags", s)
			return
		}
	}

	var value float64
	var sampleRate float64

	if value, err = strconv.ParseFloat(val, 64); err != nil {
		err = fmt.Errorf("datadog: %#v has a malformed value", s)
		return
	}

	if len(rate) != 0 {
		if sampleRate, err = strconv.ParseFloat(rate, 64); err != nil {
			err = fmt.Errorf("datadog: %#v has a malformed sample rate", s)
			return
		}
	}

	if sampleRate == 0 {
		sampleRate = 1
	}

	m = Metric{
		Type:  MetricType(typ),
		Name:  name,
		Value: value,
		Rate:  sampleRate,
	}

	if len(tags) != 0 {
		m.Tags = make([]stats.Tag, 0, count(tags, ',')+1)

		for len(tags) != 0 {
			var tag string

			if tag, tags = nextToken(tags, ','); len(tag) != 0 {
				name, value := split(tag, ':')
				m.Tags = append(m.Tags, stats.T(name, value))
			}
		}
	}

	return
}

func nextToken(s string, b byte) (token string, next string) {
	if off := strings.IndexByte(s, b); off >= 0 {
		token, next = s[:off], s[off+1:]
	} else {
		token = s
	}
	return
}

func split(s string, b byte) (head string, tail string) {
	if off := strings.LastIndexByte(s, b); off >= 0 {
		head, tail = s[:off], s[off+1:]
	} else {
		head = s
	}
	return
}

func count(s string, b byte) (n int) {
	for {
		if off := strings.IndexByte(s, b); off < 0 {
			break
		} else {
			n++
			s = s[off+1:]
		}
	}
	return
}
