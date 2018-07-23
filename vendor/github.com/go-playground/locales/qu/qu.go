package qu

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type qu struct {
	locale                 string
	pluralsCardinal        []locales.PluralRule
	pluralsOrdinal         []locales.PluralRule
	pluralsRange           []locales.PluralRule
	decimal                string
	group                  string
	minus                  string
	percent                string
	percentSuffix          string
	perMille               string
	timeSeparator          string
	inifinity              string
	currencies             []string // idx = enum of currency code
	currencyPositivePrefix string
	currencyNegativePrefix string
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

// New returns a new instance of translator for the 'qu' locale
func New() locales.Translator {
	return &qu{
		locale:                 "qu",
		pluralsCardinal:        nil,
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ".",
		group:                  ",",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "S/", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		percentSuffix:          " ",
		currencyPositivePrefix: " ",
		currencyNegativePrefix: " ",
		monthsAbbreviated:      []string{"", "Ene", "Feb", "Mar", "Abr", "May", "Jun", "Jul", "Ago", "Set", "Oct", "Nov", "Dic"},
		monthsNarrow:           []string{"", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"},
		monthsWide:             []string{"", "Enero", "Febrero", "Marzo", "Abril", "Mayo", "Junio", "Julio", "Agosto", "Setiembre", "Octubre", "Noviembre", "Diciembre"},
		daysAbbreviated:        []string{"Dom", "Lun", "Mar", "Mié", "Jue", "Vie", "Sab"},
		daysNarrow:             []string{"D", "L", "M", "X", "J", "V", "S"},
		daysShort:              []string{"Dom", "Lun", "Mar", "Mié", "Jue", "Vie", "Sab"},
		daysWide:               []string{"Domingo", "Lunes", "Martes", "Miércoles", "Jueves", "Viernes", "Sábado"},
		periodsAbbreviated:     []string{"a.m.", "p.m."},
		periodsNarrow:          []string{"a.m.", "p.m."},
		periodsWide:            []string{"a.m.", "p.m."},
		erasAbbreviated:        []string{"", ""},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"", ""},
		timezones:              map[string]string{"HNOG": "HNOG", "WART": "WART", "HENOMX": "HENOMX", "SRT": "SRT", "ARST": "Argentina Summer Time", "CHADT": "CHADT", "MDT": "MDT", "MYT": "Malaysia Time", "CLST": "Chile Summer Time", "ACDT": "ACDT", "HEOG": "HEOG", "HKT": "Hong Kong Standard Time", "ART": "Argentina Standard Time", "CST": "CST", "MST": "MST", "CAT": "CAT", "EDT": "EDT", "WARST": "WARST", "OEZ": "OEZ", "SAST": "SAST", "GFT": "French Guiana Time", "EST": "EST", "AST": "AST", "HNNOMX": "HNNOMX", "HEPM": "HEPM", "GYT": "GYT", "HEPMX": "HEPMX", "BT": "Bhutan Time", "AKST": "AKST", "HKST": "Hong Kong Summer Time", "MEZ": "MEZ", "HAT": "HAT", "EAT": "EAT", "HADT": "HADT", "ChST": "ChST", "HNPMX": "HNPMX", "JDT": "Japan Daylight Time", "LHDT": "LHDT", "WEZ": "WEZ", "WIB": "WIB", "NZDT": "New Zealand Daylight Time", "ACST": "ACST", "UYST": "UYST", "GMT": "GMT", "HEEG": "East Greenland Summer Time", "HNT": "HNT", "COT": "Colombia Standard Time", "UYT": "UYT", "ADT": "ADT", "BOT": "Bolivia Time", "AEDT": "AEDT", "ECT": "Ecuador Time", "VET": "VET", "HNPM": "HNPM", "HECU": "Cuba Daylight Time", "CHAST": "CHAST", "PDT": "PDT", "AWST": "AWST", "AEST": "AEST", "WAST": "WAST", "TMT": "Turkmenistan Standard Time", "HNEG": "East Greenland Standard Time", "COST": "Colombia Summer Time", "HNCU": "Cuba Standard Time", "AWDT": "AWDT", "SGT": "Singapore Standard Time", "WAT": "WAT", "WESZ": "WESZ", "NZST": "New Zealand Standard Time", "LHST": "LHST", "CLT": "Chile Standard Time", "TMST": "Turkmenistan Summer Time", "HAST": "HAST", "JST": "Japan Standard Time", "AKDT": "AKDT", "WITA": "Central Indonesia Time", "ACWST": "ACWST", "ACWDT": "ACWDT", "MESZ": "MESZ", "IST": "India Standard Time", "∅∅∅": "∅∅∅", "OESZ": "OESZ", "PST": "PST", "CDT": "CDT", "WIT": "Eastern Indonesia Time"},
	}
}

// Locale returns the current translators string locale
func (qu *qu) Locale() string {
	return qu.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'qu'
func (qu *qu) PluralsCardinal() []locales.PluralRule {
	return qu.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'qu'
func (qu *qu) PluralsOrdinal() []locales.PluralRule {
	return qu.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'qu'
func (qu *qu) PluralsRange() []locales.PluralRule {
	return qu.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'qu'
func (qu *qu) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'qu'
func (qu *qu) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'qu'
func (qu *qu) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (qu *qu) MonthAbbreviated(month time.Month) string {
	return qu.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (qu *qu) MonthsAbbreviated() []string {
	return qu.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (qu *qu) MonthNarrow(month time.Month) string {
	return qu.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (qu *qu) MonthsNarrow() []string {
	return qu.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (qu *qu) MonthWide(month time.Month) string {
	return qu.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (qu *qu) MonthsWide() []string {
	return qu.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (qu *qu) WeekdayAbbreviated(weekday time.Weekday) string {
	return qu.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (qu *qu) WeekdaysAbbreviated() []string {
	return qu.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (qu *qu) WeekdayNarrow(weekday time.Weekday) string {
	return qu.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (qu *qu) WeekdaysNarrow() []string {
	return qu.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (qu *qu) WeekdayShort(weekday time.Weekday) string {
	return qu.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (qu *qu) WeekdaysShort() []string {
	return qu.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (qu *qu) WeekdayWide(weekday time.Weekday) string {
	return qu.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (qu *qu) WeekdaysWide() []string {
	return qu.daysWide
}

// Decimal returns the decimal point of number
func (qu *qu) Decimal() string {
	return qu.decimal
}

// Group returns the group of number
func (qu *qu) Group() string {
	return qu.group
}

// Group returns the minus sign of number
func (qu *qu) Minus() string {
	return qu.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'qu' and handles both Whole and Real numbers based on 'v'
func (qu *qu) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, qu.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, qu.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, qu.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'qu' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (qu *qu) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 5
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, qu.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, qu.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, qu.percentSuffix...)

	b = append(b, qu.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'qu'
func (qu *qu) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := qu.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, qu.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, qu.group[0])
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

	for j := len(qu.currencyPositivePrefix) - 1; j >= 0; j-- {
		b = append(b, qu.currencyPositivePrefix[j])
	}

	if num < 0 {
		b = append(b, qu.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, qu.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'qu'
// in accounting notation.
func (qu *qu) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := qu.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, qu.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, qu.group[0])
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

		for j := len(qu.currencyNegativePrefix) - 1; j >= 0; j-- {
			b = append(b, qu.currencyNegativePrefix[j])
		}

		b = append(b, qu.minus[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(qu.currencyPositivePrefix) - 1; j >= 0; j-- {
			b = append(b, qu.currencyPositivePrefix[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, qu.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'qu'
func (qu *qu) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'qu'
func (qu *qu) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, qu.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'qu'
func (qu *qu) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, qu.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'qu'
func (qu *qu) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, qu.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, qu.monthsWide[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'qu'
func (qu *qu) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, qu.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'qu'
func (qu *qu) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, qu.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, qu.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'qu'
func (qu *qu) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, qu.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, qu.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'qu'
func (qu *qu) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, qu.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, qu.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := qu.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
