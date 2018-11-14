package yo_NG

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type yo_NG struct {
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

// New returns a new instance of translator for the 'yo_NG' locale
func New() locales.Translator {
	return &yo_NG{
		locale:                 "yo_NG",
		pluralsCardinal:        []locales.PluralRule{6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ".",
		group:                  ",",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: ")",
		monthsAbbreviated:      []string{"", "Ṣẹ́rẹ́", "Èrèlè", "Ẹrẹ̀nà", "Ìgbé", "Ẹ̀bibi", "Òkúdu", "Agẹmọ", "Ògún", "Owewe", "Ọ̀wàrà", "Bélú", "Ọ̀pẹ̀"},
		monthsNarrow:           []string{"", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"},
		monthsWide:             []string{"", "Oṣù Ṣẹ́rẹ́", "Oṣù Èrèlè", "Oṣù Ẹrẹ̀nà", "Oṣù Ìgbé", "Oṣù Ẹ̀bibi", "Oṣù Òkúdu", "Oṣù Agẹmọ", "Oṣù Ògún", "Oṣù Owewe", "Oṣù Ọ̀wàrà", "Oṣù Bélú", "Oṣù Ọ̀pẹ̀"},
		daysAbbreviated:        []string{"Àìkú", "Ajé", "Ìsẹ́gun", "Ọjọ́rú", "Ọjọ́bọ", "Ẹtì", "Àbámẹ́ta"},
		daysShort:              []string{"Àìkú", "Ajé", "Ìsẹ́gun", "Ọjọ́rú", "Ọjọ́bọ", "Ẹtì", "Àbámẹ́ta"},
		daysWide:               []string{"Ọjọ́ Àìkú", "Ọjọ́ Ajé", "Ọjọ́ Ìsẹ́gun", "Ọjọ́rú", "Ọjọ́bọ", "Ọjọ́ Ẹtì", "Ọjọ́ Àbámẹ́ta"},
		periodsAbbreviated:     []string{"Àárọ̀", "Ọ̀sán"},
		periodsNarrow:          []string{"Àárọ̀", "Ọ̀sán"},
		periodsWide:            []string{"Àárọ̀", "Ọ̀sán"},
		erasAbbreviated:        []string{"", ""},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"Saju Kristi", "Lehin Kristi"},
		timezones:              map[string]string{"JST": "JST", "EST": "EST", "IST": "IST", "HAT": "HAT", "UYST": "UYST", "GYT": "GYT", "AEDT": "AEDT", "NZST": "NZST", "WART": "WART", "EAT": "EAT", "CHADT": "CHADT", "HNCU": "HNCU", "JDT": "JDT", "ChST": "ChST", "SGT": "SGT", "ACWDT": "ACWDT", "MDT": "MDT", "SRT": "SRT", "WIB": "WIB", "BOT": "BOT", "HADT": "HADT", "SAST": "SAST", "MEZ": "MEZ", "HKT": "HKT", "LHST": "LHST", "HNT": "HNT", "OESZ": "OESZ", "NZDT": "NZDT", "MESZ": "MESZ", "HEPMX": "HEPMX", "CST": "CST", "PDT": "PDT", "∅∅∅": "∅∅∅", "OEZ": "OEZ", "GMT": "GMT", "HNPMX": "HNPMX", "HKST": "HKST", "ARST": "ARST", "HAST": "HAST", "BT": "BT", "AKST": "AKST", "COT": "COT", "CHAST": "CHAST", "CDT": "CDT", "WEZ": "WEZ", "ADT": "ADT", "MYT": "MYT", "HNOG": "HNOG", "EDT": "EDT", "HNPM": "HNPM", "HENOMX": "HENOMX", "WIT": "WIT", "COST": "COST", "HEEG": "HEEG", "HEOG": "HEOG", "VET": "VET", "CLST": "CLST", "CAT": "CAT", "AEST": "AEST", "WARST": "WARST", "MST": "MST", "TMT": "TMT", "WESZ": "WESZ", "ACST": "ACST", "LHDT": "LHDT", "HECU": "HECU", "AST": "AST", "GFT": "GFT", "HNEG": "HNEG", "ACWST": "ACWST", "HEPM": "HEPM", "WITA": "WITA", "HNNOMX": "HNNOMX", "UYT": "UYT", "ACDT": "ACDT", "CLT": "CLT", "AWDT": "AWDT", "WAT": "WAT", "ECT": "ECT", "WAST": "WAST", "AKDT": "AKDT", "TMST": "TMST", "ART": "ART", "AWST": "AWST", "PST": "PST"},
	}
}

// Locale returns the current translators string locale
func (yo *yo_NG) Locale() string {
	return yo.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'yo_NG'
func (yo *yo_NG) PluralsCardinal() []locales.PluralRule {
	return yo.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'yo_NG'
func (yo *yo_NG) PluralsOrdinal() []locales.PluralRule {
	return yo.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'yo_NG'
func (yo *yo_NG) PluralsRange() []locales.PluralRule {
	return yo.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'yo_NG'
func (yo *yo_NG) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'yo_NG'
func (yo *yo_NG) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'yo_NG'
func (yo *yo_NG) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (yo *yo_NG) MonthAbbreviated(month time.Month) string {
	return yo.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (yo *yo_NG) MonthsAbbreviated() []string {
	return yo.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (yo *yo_NG) MonthNarrow(month time.Month) string {
	return yo.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (yo *yo_NG) MonthsNarrow() []string {
	return yo.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (yo *yo_NG) MonthWide(month time.Month) string {
	return yo.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (yo *yo_NG) MonthsWide() []string {
	return yo.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (yo *yo_NG) WeekdayAbbreviated(weekday time.Weekday) string {
	return yo.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (yo *yo_NG) WeekdaysAbbreviated() []string {
	return yo.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (yo *yo_NG) WeekdayNarrow(weekday time.Weekday) string {
	return yo.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (yo *yo_NG) WeekdaysNarrow() []string {
	return yo.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (yo *yo_NG) WeekdayShort(weekday time.Weekday) string {
	return yo.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (yo *yo_NG) WeekdaysShort() []string {
	return yo.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (yo *yo_NG) WeekdayWide(weekday time.Weekday) string {
	return yo.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (yo *yo_NG) WeekdaysWide() []string {
	return yo.daysWide
}

// Decimal returns the decimal point of number
func (yo *yo_NG) Decimal() string {
	return yo.decimal
}

// Group returns the group of number
func (yo *yo_NG) Group() string {
	return yo.group
}

// Group returns the minus sign of number
func (yo *yo_NG) Minus() string {
	return yo.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'yo_NG' and handles both Whole and Real numbers based on 'v'
func (yo *yo_NG) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, yo.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, yo.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, yo.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'yo_NG' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (yo *yo_NG) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, yo.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, yo.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, yo.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'yo_NG'
func (yo *yo_NG) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := yo.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, yo.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, yo.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	for j := len(symbol) - 1; j >= 0; j-- {
		b = append(b, symbol[j])
	}

	if num < 0 {
		b = append(b, yo.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, yo.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'yo_NG'
// in accounting notation.
func (yo *yo_NG) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := yo.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, yo.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, yo.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		b = append(b, yo.currencyNegativePrefix[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, yo.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, yo.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'yo_NG'
func (yo *yo_NG) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2f}...)

	if t.Month() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Month()), 10)

	b = append(b, []byte{0x2f}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'yo_NG'
func (yo *yo_NG) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, yo.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'yo_NG'
func (yo *yo_NG) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, yo.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'yo_NG'
func (yo *yo_NG) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, yo.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, yo.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'yo_NG'
func (yo *yo_NG) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, yo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'yo_NG'
func (yo *yo_NG) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, yo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, yo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'yo_NG'
func (yo *yo_NG) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, yo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, yo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'yo_NG'
func (yo *yo_NG) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, yo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, yo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := yo.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
