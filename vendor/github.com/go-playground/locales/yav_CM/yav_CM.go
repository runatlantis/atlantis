package yav_CM

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type yav_CM struct {
	locale                 string
	pluralsCardinal        []locales.PluralRule
	pluralsOrdinal         []locales.PluralRule
	pluralsRange           []locales.PluralRule
	decimal                string
	group                  string
	minus                  string
	percent                string
	perMille               string
	timeSeparator          string
	inifinity              string
	currencies             []string // idx = enum of currency code
	currencyPositiveSuffix string
	currencyNegativePrefix string
	currencyNegativeSuffix string
	monthsAbbreviated      []string
	monthsNarrow           []string
	monthsWide             []string
	daysAbbreviated        []string
	daysNarrow             []string
	daysShort              []string
	daysWide               []string
	periodsAbbreviated     []string
	periodsNarrow          []string
	periodsShort           []string
	periodsWide            []string
	erasAbbreviated        []string
	erasNarrow             []string
	erasWide               []string
	timezones              map[string]string
}

// New returns a new instance of translator for the 'yav_CM' locale
func New() locales.Translator {
	return &yav_CM{
		locale:                 "yav_CM",
		pluralsCardinal:        nil,
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ",",
		group:                  " ",
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositiveSuffix: " ",
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: " )",
		monthsAbbreviated:      []string{"", "o.1", "o.2", "o.3", "o.4", "o.5", "o.6", "o.7", "o.8", "o.9", "o.10", "o.11", "o.12"},
		monthsWide:             []string{"", "pikítíkítie, oólí ú kutúan", "siɛyɛ́, oóli ú kándíɛ", "ɔnsúmbɔl, oóli ú kátátúɛ", "mesiŋ, oóli ú kénie", "ensil, oóli ú kátánuɛ", "ɔsɔn", "efute", "pisuyú", "imɛŋ i puɔs", "imɛŋ i putúk,oóli ú kátíɛ", "makandikɛ", "pilɔndɔ́"},
		daysAbbreviated:        []string{"sd", "md", "mw", "et", "kl", "fl", "ss"},
		daysNarrow:             []string{"s", "m", "m", "e", "k", "f", "s"},
		daysWide:               []string{"sɔ́ndiɛ", "móndie", "muányáŋmóndie", "metúkpíápɛ", "kúpélimetúkpiapɛ", "feléte", "séselé"},
		periodsAbbreviated:     []string{"kiɛmɛ́ɛm", "kisɛ́ndɛ"},
		periodsWide:            []string{"kiɛmɛ́ɛm", "kisɛ́ndɛ"},
		erasAbbreviated:        []string{"k.Y.", "+J.C."},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"katikupíen Yésuse", "ékélémkúnupíén n"},
		timezones:              map[string]string{"GYT": "GYT", "UYT": "UYT", "HECU": "HECU", "PDT": "PDT", "ADT": "ADT", "GFT": "GFT", "HEOG": "HEOG", "HENOMX": "HENOMX", "SRT": "SRT", "MST": "MST", "ACWST": "ACWST", "TMST": "TMST", "ARST": "ARST", "ACDT": "ACDT", "HNEG": "HNEG", "HAT": "HAT", "AWDT": "AWDT", "MDT": "MDT", "ACST": "ACST", "WART": "WART", "CAT": "CAT", "WIT": "WIT", "HADT": "HADT", "GMT": "GMT", "WESZ": "WESZ", "EDT": "EDT", "HNNOMX": "HNNOMX", "EAT": "EAT", "BT": "BT", "NZST": "NZST", "HKT": "HKT", "LHDT": "LHDT", "TMT": "TMT", "AEST": "AEST", "WIB": "WIB", "AKDT": "AKDT", "AST": "AST", "AEDT": "AEDT", "NZDT": "NZDT", "MYT": "MYT", "COT": "COT", "HEPMX": "HEPMX", "WAT": "WAT", "WAST": "WAST", "ECT": "ECT", "LHST": "LHST", "CLT": "CLT", "∅∅∅": "∅∅∅", "CHAST": "CHAST", "HNCU": "HNCU", "PST": "PST", "WEZ": "WEZ", "VET": "VET", "OESZ": "OESZ", "JDT": "JDT", "SGT": "SGT", "WITA": "WITA", "OEZ": "OEZ", "HAST": "HAST", "ChST": "ChST", "AWST": "AWST", "HEEG": "HEEG", "MEZ": "MEZ", "MESZ": "MESZ", "IST": "IST", "HEPM": "HEPM", "SAST": "SAST", "JST": "JST", "EST": "EST", "HNPM": "HNPM", "ART": "ART", "HNPMX": "HNPMX", "BOT": "BOT", "CLST": "CLST", "UYST": "UYST", "CST": "CST", "HNT": "HNT", "COST": "COST", "CHADT": "CHADT", "CDT": "CDT", "AKST": "AKST", "ACWDT": "ACWDT", "HNOG": "HNOG", "HKST": "HKST", "WARST": "WARST"},
	}
}

// Locale returns the current translators string locale
func (yav *yav_CM) Locale() string {
	return yav.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'yav_CM'
func (yav *yav_CM) PluralsCardinal() []locales.PluralRule {
	return yav.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'yav_CM'
func (yav *yav_CM) PluralsOrdinal() []locales.PluralRule {
	return yav.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'yav_CM'
func (yav *yav_CM) PluralsRange() []locales.PluralRule {
	return yav.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'yav_CM'
func (yav *yav_CM) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'yav_CM'
func (yav *yav_CM) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'yav_CM'
func (yav *yav_CM) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (yav *yav_CM) MonthAbbreviated(month time.Month) string {
	return yav.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (yav *yav_CM) MonthsAbbreviated() []string {
	return yav.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (yav *yav_CM) MonthNarrow(month time.Month) string {
	return yav.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (yav *yav_CM) MonthsNarrow() []string {
	return nil
}

// MonthWide returns the locales wide month given the 'month' provided
func (yav *yav_CM) MonthWide(month time.Month) string {
	return yav.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (yav *yav_CM) MonthsWide() []string {
	return yav.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (yav *yav_CM) WeekdayAbbreviated(weekday time.Weekday) string {
	return yav.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (yav *yav_CM) WeekdaysAbbreviated() []string {
	return yav.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (yav *yav_CM) WeekdayNarrow(weekday time.Weekday) string {
	return yav.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (yav *yav_CM) WeekdaysNarrow() []string {
	return yav.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (yav *yav_CM) WeekdayShort(weekday time.Weekday) string {
	return yav.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (yav *yav_CM) WeekdaysShort() []string {
	return yav.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (yav *yav_CM) WeekdayWide(weekday time.Weekday) string {
	return yav.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (yav *yav_CM) WeekdaysWide() []string {
	return yav.daysWide
}

// Decimal returns the decimal point of number
func (yav *yav_CM) Decimal() string {
	return yav.decimal
}

// Group returns the group of number
func (yav *yav_CM) Group() string {
	return yav.group
}

// Group returns the minus sign of number
func (yav *yav_CM) Minus() string {
	return yav.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'yav_CM' and handles both Whole and Real numbers based on 'v'
func (yav *yav_CM) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 1 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, yav.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(yav.group) - 1; j >= 0; j-- {
					b = append(b, yav.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, yav.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'yav_CM' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (yav *yav_CM) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 1
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, yav.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, yav.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, yav.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'yav_CM'
func (yav *yav_CM) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := yav.currencies[currency]
	l := len(s) + len(symbol) + 3 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, yav.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(yav.group) - 1; j >= 0; j-- {
					b = append(b, yav.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, yav.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, yav.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, yav.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'yav_CM'
// in accounting notation.
func (yav *yav_CM) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := yav.currencies[currency]
	l := len(s) + len(symbol) + 5 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, yav.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(yav.group) - 1; j >= 0; j-- {
					b = append(b, yav.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, yav.currencyNegativePrefix[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, yav.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, yav.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, yav.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'yav_CM'
func (yav *yav_CM) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'yav_CM'
func (yav *yav_CM) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, yav.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'yav_CM'
func (yav *yav_CM) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, yav.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'yav_CM'
func (yav *yav_CM) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, yav.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, yav.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'yav_CM'
func (yav *yav_CM) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, yav.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'yav_CM'
func (yav *yav_CM) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, yav.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, yav.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'yav_CM'
func (yav *yav_CM) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, yav.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, yav.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'yav_CM'
func (yav *yav_CM) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, yav.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, yav.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := yav.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
