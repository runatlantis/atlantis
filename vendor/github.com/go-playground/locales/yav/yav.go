package yav

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type yav struct {
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

// New returns a new instance of translator for the 'yav' locale
func New() locales.Translator {
	return &yav{
		locale:                 "yav",
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
		timezones:              map[string]string{"AWST": "AWST", "WAT": "WAT", "BT": "BT", "NZST": "NZST", "EAT": "EAT", "ARST": "ARST", "COST": "COST", "ChST": "ChST", "ECT": "ECT", "CDT": "CDT", "EST": "EST", "WIB": "WIB", "ACWST": "ACWST", "HNEG": "HNEG", "HKST": "HKST", "WITA": "WITA", "MDT": "MDT", "HAST": "HAST", "PDT": "PDT", "HADT": "HADT", "SAST": "SAST", "WAST": "WAST", "HNOG": "HNOG", "HNT": "HNT", "HNNOMX": "HNNOMX", "MST": "MST", "WIT": "WIT", "HKT": "HKT", "GFT": "GFT", "AKDT": "AKDT", "HEOG": "HEOG", "IST": "IST", "LHDT": "LHDT", "HENOMX": "HENOMX", "MYT": "MYT", "TMT": "TMT", "GMT": "GMT", "AKST": "AKST", "ACDT": "ACDT", "WART": "WART", "CLST": "CLST", "PST": "PST", "MESZ": "MESZ", "SRT": "SRT", "AWDT": "AWDT", "ACST": "ACST", "MEZ": "MEZ", "WEZ": "WEZ", "LHST": "LHST", "VET": "VET", "CHADT": "CHADT", "ADT": "ADT", "TMST": "TMST", "OEZ": "OEZ", "HNPMX": "HNPMX", "EDT": "EDT", "JST": "JST", "HEEG": "HEEG", "ART": "ART", "HECU": "HECU", "CST": "CST", "WESZ": "WESZ", "JDT": "JDT", "CAT": "CAT", "OESZ": "OESZ", "HNCU": "HNCU", "AEDT": "AEDT", "AST": "AST", "ACWDT": "ACWDT", "WARST": "WARST", "COT": "COT", "GYT": "GYT", "UYT": "UYT", "CHAST": "CHAST", "AEST": "AEST", "SGT": "SGT", "∅∅∅": "∅∅∅", "HEPM": "HEPM", "CLT": "CLT", "UYST": "UYST", "BOT": "BOT", "HAT": "HAT", "HNPM": "HNPM", "HEPMX": "HEPMX", "NZDT": "NZDT"},
	}
}

// Locale returns the current translators string locale
func (yav *yav) Locale() string {
	return yav.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'yav'
func (yav *yav) PluralsCardinal() []locales.PluralRule {
	return yav.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'yav'
func (yav *yav) PluralsOrdinal() []locales.PluralRule {
	return yav.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'yav'
func (yav *yav) PluralsRange() []locales.PluralRule {
	return yav.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'yav'
func (yav *yav) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'yav'
func (yav *yav) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'yav'
func (yav *yav) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (yav *yav) MonthAbbreviated(month time.Month) string {
	return yav.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (yav *yav) MonthsAbbreviated() []string {
	return yav.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (yav *yav) MonthNarrow(month time.Month) string {
	return yav.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (yav *yav) MonthsNarrow() []string {
	return nil
}

// MonthWide returns the locales wide month given the 'month' provided
func (yav *yav) MonthWide(month time.Month) string {
	return yav.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (yav *yav) MonthsWide() []string {
	return yav.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (yav *yav) WeekdayAbbreviated(weekday time.Weekday) string {
	return yav.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (yav *yav) WeekdaysAbbreviated() []string {
	return yav.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (yav *yav) WeekdayNarrow(weekday time.Weekday) string {
	return yav.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (yav *yav) WeekdaysNarrow() []string {
	return yav.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (yav *yav) WeekdayShort(weekday time.Weekday) string {
	return yav.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (yav *yav) WeekdaysShort() []string {
	return yav.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (yav *yav) WeekdayWide(weekday time.Weekday) string {
	return yav.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (yav *yav) WeekdaysWide() []string {
	return yav.daysWide
}

// Decimal returns the decimal point of number
func (yav *yav) Decimal() string {
	return yav.decimal
}

// Group returns the group of number
func (yav *yav) Group() string {
	return yav.group
}

// Group returns the minus sign of number
func (yav *yav) Minus() string {
	return yav.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'yav' and handles both Whole and Real numbers based on 'v'
func (yav *yav) FmtNumber(num float64, v uint64) string {

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

// FmtPercent returns 'num' with digits/precision of 'v' for 'yav' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (yav *yav) FmtPercent(num float64, v uint64) string {
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

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'yav'
func (yav *yav) FmtCurrency(num float64, v uint64, currency currency.Type) string {

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

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'yav'
// in accounting notation.
func (yav *yav) FmtAccounting(num float64, v uint64, currency currency.Type) string {

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

// FmtDateShort returns the short date representation of 't' for 'yav'
func (yav *yav) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'yav'
func (yav *yav) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'yav'
func (yav *yav) FmtDateLong(t time.Time) string {

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

// FmtDateFull returns the full date representation of 't' for 'yav'
func (yav *yav) FmtDateFull(t time.Time) string {

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

// FmtTimeShort returns the short time representation of 't' for 'yav'
func (yav *yav) FmtTimeShort(t time.Time) string {

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

// FmtTimeMedium returns the medium time representation of 't' for 'yav'
func (yav *yav) FmtTimeMedium(t time.Time) string {

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

// FmtTimeLong returns the long time representation of 't' for 'yav'
func (yav *yav) FmtTimeLong(t time.Time) string {

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

// FmtTimeFull returns the full time representation of 't' for 'yav'
func (yav *yav) FmtTimeFull(t time.Time) string {

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
