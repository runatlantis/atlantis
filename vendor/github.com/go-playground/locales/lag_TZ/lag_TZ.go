package lag_TZ

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type lag_TZ struct {
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
	currencyPositivePrefix string
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

// New returns a new instance of translator for the 'lag_TZ' locale
func New() locales.Translator {
	return &lag_TZ{
		locale:                 "lag_TZ",
		pluralsCardinal:        []locales.PluralRule{1, 2, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositivePrefix: " ",
		currencyPositiveSuffix: "K",
		currencyNegativePrefix: " ",
		currencyNegativeSuffix: "K",
		monthsAbbreviated:      []string{"", "Fúngatɨ", "Naanɨ", "Keenda", "Ikúmi", "Inyambala", "Idwaata", "Mʉʉnchɨ", "Vɨɨrɨ", "Saatʉ", "Inyi", "Saano", "Sasatʉ"},
		monthsNarrow:           []string{"", "F", "N", "K", "I", "I", "I", "M", "V", "S", "I", "S", "S"},
		monthsWide:             []string{"", "Kʉfúngatɨ", "Kʉnaanɨ", "Kʉkeenda", "Kwiikumi", "Kwiinyambála", "Kwiidwaata", "Kʉmʉʉnchɨ", "Kʉvɨɨrɨ", "Kʉsaatʉ", "Kwiinyi", "Kʉsaano", "Kʉsasatʉ"},
		daysAbbreviated:        []string{"Píili", "Táatu", "Íne", "Táano", "Alh", "Ijm", "Móosi"},
		daysNarrow:             []string{"P", "T", "E", "O", "A", "I", "M"},
		daysWide:               []string{"Jumapíiri", "Jumatátu", "Jumaíne", "Jumatáano", "Alamíisi", "Ijumáa", "Jumamóosi"},
		periodsAbbreviated:     []string{"TOO", "MUU"},
		periodsWide:            []string{"TOO", "MUU"},
		erasAbbreviated:        []string{"KSA", "KA"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"Kɨrɨsitʉ sɨ anavyaal", "Kɨrɨsitʉ akavyaalwe"},
		timezones:              map[string]string{"ECT": "ECT", "HNEG": "HNEG", "MEZ": "MEZ", "OEZ": "OEZ", "ART": "ART", "WESZ": "WESZ", "HAT": "HAT", "COST": "COST", "HEPMX": "HEPMX", "TMT": "TMT", "WAT": "WAT", "NZST": "NZST", "SGT": "SGT", "ACDT": "ACDT", "ACWST": "ACWST", "HEPM": "HEPM", "WIT": "WIT", "AST": "AST", "WAST": "WAST", "GFT": "GFT", "AKDT": "AKDT", "HEOG": "HEOG", "WART": "WART", "ADT": "ADT", "AEDT": "AEDT", "HNOG": "HNOG", "CAT": "CAT", "UYT": "UYT", "CDT": "CDT", "EDT": "EDT", "IST": "IST", "HNNOMX": "HNNOMX", "OESZ": "OESZ", "AWST": "AWST", "AWDT": "AWDT", "VET": "VET", "CLST": "CLST", "COT": "COT", "HADT": "HADT", "GMT": "GMT", "CST": "CST", "AEST": "AEST", "WIB": "WIB", "JST": "JST", "GYT": "GYT", "HNCU": "HNCU", "HEEG": "HEEG", "HNPMX": "HNPMX", "PDT": "PDT", "HNT": "HNT", "HENOMX": "HENOMX", "∅∅∅": "∅∅∅", "ACST": "ACST", "LHDT": "LHDT", "HNPM": "HNPM", "TMST": "TMST", "UYST": "UYST", "MST": "MST", "BT": "BT", "NZDT": "NZDT", "EST": "EST", "MESZ": "MESZ", "WARST": "WARST", "CLT": "CLT", "BOT": "BOT", "PST": "PST", "SAST": "SAST", "MYT": "MYT", "HECU": "HECU", "AKST": "AKST", "HKT": "HKT", "HKST": "HKST", "WITA": "WITA", "SRT": "SRT", "ARST": "ARST", "CHAST": "CHAST", "ChST": "ChST", "MDT": "MDT", "WEZ": "WEZ", "JDT": "JDT", "ACWDT": "ACWDT", "LHST": "LHST", "EAT": "EAT", "HAST": "HAST", "CHADT": "CHADT"},
	}
}

// Locale returns the current translators string locale
func (lag *lag_TZ) Locale() string {
	return lag.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'lag_TZ'
func (lag *lag_TZ) PluralsCardinal() []locales.PluralRule {
	return lag.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'lag_TZ'
func (lag *lag_TZ) PluralsOrdinal() []locales.PluralRule {
	return lag.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'lag_TZ'
func (lag *lag_TZ) PluralsRange() []locales.PluralRule {
	return lag.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'lag_TZ'
func (lag *lag_TZ) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)

	if n == 0 {
		return locales.PluralRuleZero
	} else if (i == 0 || i == 1) && n != 0 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'lag_TZ'
func (lag *lag_TZ) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'lag_TZ'
func (lag *lag_TZ) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (lag *lag_TZ) MonthAbbreviated(month time.Month) string {
	return lag.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (lag *lag_TZ) MonthsAbbreviated() []string {
	return lag.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (lag *lag_TZ) MonthNarrow(month time.Month) string {
	return lag.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (lag *lag_TZ) MonthsNarrow() []string {
	return lag.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (lag *lag_TZ) MonthWide(month time.Month) string {
	return lag.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (lag *lag_TZ) MonthsWide() []string {
	return lag.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (lag *lag_TZ) WeekdayAbbreviated(weekday time.Weekday) string {
	return lag.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (lag *lag_TZ) WeekdaysAbbreviated() []string {
	return lag.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (lag *lag_TZ) WeekdayNarrow(weekday time.Weekday) string {
	return lag.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (lag *lag_TZ) WeekdaysNarrow() []string {
	return lag.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (lag *lag_TZ) WeekdayShort(weekday time.Weekday) string {
	return lag.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (lag *lag_TZ) WeekdaysShort() []string {
	return lag.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (lag *lag_TZ) WeekdayWide(weekday time.Weekday) string {
	return lag.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (lag *lag_TZ) WeekdaysWide() []string {
	return lag.daysWide
}

// Decimal returns the decimal point of number
func (lag *lag_TZ) Decimal() string {
	return lag.decimal
}

// Group returns the group of number
func (lag *lag_TZ) Group() string {
	return lag.group
}

// Group returns the minus sign of number
func (lag *lag_TZ) Minus() string {
	return lag.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'lag_TZ' and handles both Whole and Real numbers based on 'v'
func (lag *lag_TZ) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'lag_TZ' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (lag *lag_TZ) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'lag_TZ'
func (lag *lag_TZ) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := lag.currencies[currency]
	l := len(s) + len(symbol) + 3

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, lag.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	for j := len(symbol) - 1; j >= 0; j-- {
		b = append(b, symbol[j])
	}

	for j := len(lag.currencyPositivePrefix) - 1; j >= 0; j-- {
		b = append(b, lag.currencyPositivePrefix[j])
	}

	if num < 0 {
		b = append(b, lag.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, lag.currencyPositiveSuffix...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'lag_TZ'
// in accounting notation.
func (lag *lag_TZ) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := lag.currencies[currency]
	l := len(s) + len(symbol) + 3

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, lag.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(lag.currencyNegativePrefix) - 1; j >= 0; j-- {
			b = append(b, lag.currencyNegativePrefix[j])
		}

		b = append(b, lag.minus[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(lag.currencyPositivePrefix) - 1; j >= 0; j-- {
			b = append(b, lag.currencyPositivePrefix[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if num < 0 {
		b = append(b, lag.currencyNegativeSuffix...)
	} else {

		b = append(b, lag.currencyPositiveSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'lag_TZ'
func (lag *lag_TZ) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'lag_TZ'
func (lag *lag_TZ) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, lag.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'lag_TZ'
func (lag *lag_TZ) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, lag.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'lag_TZ'
func (lag *lag_TZ) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, lag.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, lag.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'lag_TZ'
func (lag *lag_TZ) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, lag.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'lag_TZ'
func (lag *lag_TZ) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, lag.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, lag.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'lag_TZ'
func (lag *lag_TZ) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, lag.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, lag.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'lag_TZ'
func (lag *lag_TZ) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, lag.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, lag.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := lag.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
