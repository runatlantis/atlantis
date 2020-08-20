package datadog

import (
	"math"
	"strconv"

	"github.com/segmentio/stats"
)

// AppendMeasure is a formatting routine to append the dogstatsd protocol
// representation of a measure to a memory buffer.
func AppendMeasure(b []byte, m stats.Measure) []byte {
	return AppendMeasureFiltered(b, m, nil)
}

// AppendMeasureFiltered is a formatting routine to append the dogstatsd protocol
// representation of a measure to a memory buffer. Tags listed in the filters map
// are removed. (some tags may not be suitable for submission to DataDog)
func AppendMeasureFiltered(b []byte, m stats.Measure, filters map[string]struct{}) []byte {
	for _, field := range m.Fields {
		b = append(b, m.Name...)
		if len(field.Name) != 0 {
			b = append(b, '.')
			b = append(b, field.Name...)
		}
		b = append(b, ':')

		switch v := field.Value; v.Type() {
		case stats.Bool:
			if v.Bool() {
				b = append(b, '1')
			} else {
				b = append(b, '0')
			}
		case stats.Int:
			b = strconv.AppendInt(b, v.Int(), 10)
		case stats.Uint:
			b = strconv.AppendUint(b, v.Uint(), 10)
		case stats.Float:
			b = strconv.AppendFloat(b, normalizeFloat(v.Float()), 'g', -1, 64)
		case stats.Duration:
			b = strconv.AppendFloat(b, v.Duration().Seconds(), 'g', -1, 64)
		default:
			b = append(b, '0')
		}

		switch field.Type() {
		case stats.Counter:
			b = append(b, '|', 'c')
		case stats.Gauge:
			b = append(b, '|', 'g')
		default:
			b = append(b, '|', 'h')
		}

		if n := len(m.Tags); n != 0 {
			b = append(b, '|', '#')

			for i, t := range m.Tags {
				if _, ok := filters[t.Name]; !ok {
					if i != 0 {
						b = append(b, ',')
					}
					b = append(b, t.Name...)
					b = append(b, ':')
					b = append(b, t.Value...)
				}
			}
		}

		b = append(b, '\n')
	}

	return b
}

func normalizeFloat(f float64) float64 {
	switch {
	case math.IsNaN(f):
		return 0.0
	case math.IsInf(f, +1):
		return +math.MaxFloat64
	case math.IsInf(f, -1):
		return -math.MaxFloat64
	default:
		return f
	}
}
