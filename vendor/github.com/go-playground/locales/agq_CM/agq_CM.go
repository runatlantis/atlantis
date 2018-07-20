package agq_CM

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type agq_CM struct {
	locale             string
	pluralsCardinal    []locales.PluralRule
	pluralsOrdinal     []locales.PluralRule
	pluralsRange       []locales.PluralRule
	decimal            string
	group              string
	minus              string
	percent            string
	perMille           string
	timeSeparator      string
	inifinity          string
	currencies         []string // idx = enum of currency code
	monthsAbbreviated  []string
	monthsNarrow       []string
	monthsWide         []string
	daysAbbreviated    []string
	daysNarrow         []string
	daysShort          []string
	daysWide           []string
	periodsAbbreviated []string
	periodsNarrow      []string
	periodsShort       []string
	periodsWide        []string
	erasAbbreviated    []string
	erasNarrow         []string
	erasWide           []string
	timezones          map[string]string
}

// New returns a new instance of translator for the 'agq_CM' locale
func New() locales.Translator {
	return &agq_CM{
		locale:             "agq_CM",
		pluralsCardinal:    nil,
		pluralsOrdinal:     nil,
		pluralsRange:       nil,
		decimal:            ",",
		group:              " ",
		timeSeparator:      ":",
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		monthsAbbreviated:  []string{"", "nùm", "kɨz", "tɨd", "taa", "see", "nzu", "dum", "fɔe", "dzu", "lɔm", "kaa", "fwo"},
		monthsNarrow:       []string{"", "n", "k", "t", "t", "s", "z", "k", "f", "d", "l", "c", "f"},
		monthsWide:         []string{"", "ndzɔ̀ŋɔ̀nùm", "ndzɔ̀ŋɔ̀kƗ̀zùʔ", "ndzɔ̀ŋɔ̀tƗ̀dʉ̀ghà", "ndzɔ̀ŋɔ̀tǎafʉ̄ghā", "ndzɔ̀ŋèsèe", "ndzɔ̀ŋɔ̀nzùghò", "ndzɔ̀ŋɔ̀dùmlo", "ndzɔ̀ŋɔ̀kwîfɔ̀e", "ndzɔ̀ŋɔ̀tƗ̀fʉ̀ghàdzughù", "ndzɔ̀ŋɔ̀ghǔuwelɔ̀m", "ndzɔ̀ŋɔ̀chwaʔàkaa wo", "ndzɔ̀ŋèfwòo"},
		daysAbbreviated:    []string{"nts", "kpa", "ghɔ", "tɔm", "ume", "ghɨ", "dzk"},
		daysNarrow:         []string{"n", "k", "g", "t", "u", "g", "d"},
		daysWide:           []string{"tsuʔntsɨ", "tsuʔukpà", "tsuʔughɔe", "tsuʔutɔ̀mlò", "tsuʔumè", "tsuʔughɨ̂m", "tsuʔndzɨkɔʔɔ"},
		periodsAbbreviated: []string{"a.g", "a.k"},
		periodsWide:        []string{"a.g", "a.k"},
		erasAbbreviated:    []string{"SK", "BK"},
		erasNarrow:         []string{"", ""},
		erasWide:           []string{"Sěe Kɨ̀lesto", "Bǎa Kɨ̀lesto"},
		timezones:          map[string]string{"VET": "VET", "HEPM": "HEPM", "HNCU": "HNCU", "AKDT": "AKDT", "HKST": "HKST", "WITA": "WITA", "MST": "MST", "HECU": "HECU", "SGT": "SGT", "NZST": "NZST", "JST": "JST", "ACWDT": "ACWDT", "MEZ": "MEZ", "CHADT": "CHADT", "HEPMX": "HEPMX", "AEDT": "AEDT", "ACST": "ACST", "ARST": "ARST", "WIB": "WIB", "BT": "BT", "BOT": "BOT", "HNNOMX": "HNNOMX", "OEZ": "OEZ", "UYST": "UYST", "WAT": "WAT", "CST": "CST", "HNPMX": "HNPMX", "WIT": "WIT", "TMT": "TMT", "GYT": "GYT", "WESZ": "WESZ", "ChST": "ChST", "PST": "PST", "AEST": "AEST", "CDT": "CDT", "AST": "AST", "ACDT": "ACDT", "WART": "WART", "WARST": "WARST", "CLST": "CLST", "HAST": "HAST", "UYT": "UYT", "LHDT": "LHDT", "HENOMX": "HENOMX", "GMT": "GMT", "SAST": "SAST", "EDT": "EDT", "CLT": "CLT", "AWDT": "AWDT", "∅∅∅": "∅∅∅", "JDT": "JDT", "AKST": "AKST", "LHST": "LHST", "TMST": "TMST", "ART": "ART", "COT": "COT", "ECT": "ECT", "HEOG": "HEOG", "HAT": "HAT", "EAT": "EAT", "COST": "COST", "GFT": "GFT", "EST": "EST", "PDT": "PDT", "NZDT": "NZDT", "MYT": "MYT", "ADT": "ADT", "WAST": "WAST", "ACWST": "ACWST", "HEEG": "HEEG", "HNOG": "HNOG", "SRT": "SRT", "OESZ": "OESZ", "AWST": "AWST", "HKT": "HKT", "HNT": "HNT", "IST": "IST", "CAT": "CAT", "CHAST": "CHAST", "WEZ": "WEZ", "MESZ": "MESZ", "HNPM": "HNPM", "MDT": "MDT", "HADT": "HADT", "HNEG": "HNEG"},
	}
}

// Locale returns the current translators string locale
func (agq *agq_CM) Locale() string {
	return agq.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'agq_CM'
func (agq *agq_CM) PluralsCardinal() []locales.PluralRule {
	return agq.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'agq_CM'
func (agq *agq_CM) PluralsOrdinal() []locales.PluralRule {
	return agq.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'agq_CM'
func (agq *agq_CM) PluralsRange() []locales.PluralRule {
	return agq.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'agq_CM'
func (agq *agq_CM) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'agq_CM'
func (agq *agq_CM) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'agq_CM'
func (agq *agq_CM) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (agq *agq_CM) MonthAbbreviated(month time.Month) string {
	return agq.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (agq *agq_CM) MonthsAbbreviated() []string {
	return agq.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (agq *agq_CM) MonthNarrow(month time.Month) string {
	return agq.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (agq *agq_CM) MonthsNarrow() []string {
	return agq.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (agq *agq_CM) MonthWide(month time.Month) string {
	return agq.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (agq *agq_CM) MonthsWide() []string {
	return agq.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (agq *agq_CM) WeekdayAbbreviated(weekday time.Weekday) string {
	return agq.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (agq *agq_CM) WeekdaysAbbreviated() []string {
	return agq.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (agq *agq_CM) WeekdayNarrow(weekday time.Weekday) string {
	return agq.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (agq *agq_CM) WeekdaysNarrow() []string {
	return agq.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (agq *agq_CM) WeekdayShort(weekday time.Weekday) string {
	return agq.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (agq *agq_CM) WeekdaysShort() []string {
	return agq.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (agq *agq_CM) WeekdayWide(weekday time.Weekday) string {
	return agq.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (agq *agq_CM) WeekdaysWide() []string {
	return agq.daysWide
}

// Decimal returns the decimal point of number
func (agq *agq_CM) Decimal() string {
	return agq.decimal
}

// Group returns the group of number
func (agq *agq_CM) Group() string {
	return agq.group
}

// Group returns the minus sign of number
func (agq *agq_CM) Minus() string {
	return agq.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'agq_CM' and handles both Whole and Real numbers based on 'v'
func (agq *agq_CM) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 1 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, agq.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(agq.group) - 1; j >= 0; j-- {
					b = append(b, agq.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, agq.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'agq_CM' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (agq *agq_CM) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 1
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, agq.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, agq.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, agq.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'agq_CM'
func (agq *agq_CM) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := agq.currencies[currency]
	l := len(s) + len(symbol) + 1 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, agq.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(agq.group) - 1; j >= 0; j-- {
					b = append(b, agq.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, agq.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, agq.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'agq_CM'
// in accounting notation.
func (agq *agq_CM) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := agq.currencies[currency]
	l := len(s) + len(symbol) + 1 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, agq.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(agq.group) - 1; j >= 0; j-- {
					b = append(b, agq.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, agq.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, agq.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, symbol...)
	} else {

		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'agq_CM'
func (agq *agq_CM) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2f}...)
	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0x2f}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'agq_CM'
func (agq *agq_CM) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, agq.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'agq_CM'
func (agq *agq_CM) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, agq.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'agq_CM'
func (agq *agq_CM) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, agq.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, agq.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'agq_CM'
func (agq *agq_CM) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, agq.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'agq_CM'
func (agq *agq_CM) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, agq.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, agq.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'agq_CM'
func (agq *agq_CM) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, agq.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, agq.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'agq_CM'
func (agq *agq_CM) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, agq.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, agq.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := agq.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
