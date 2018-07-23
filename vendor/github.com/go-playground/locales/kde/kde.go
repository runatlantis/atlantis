package kde

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type kde struct {
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

// New returns a new instance of translator for the 'kde' locale
func New() locales.Translator {
	return &kde{
		locale:                 "kde",
		pluralsCardinal:        []locales.PluralRule{6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TSh", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: ")",
		monthsAbbreviated:      []string{"", "Jan", "Feb", "Mac", "Apr", "Mei", "Jun", "Jul", "Ago", "Sep", "Okt", "Nov", "Des"},
		monthsNarrow:           []string{"", "J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"},
		monthsWide:             []string{"", "Mwedi Ntandi", "Mwedi wa Pili", "Mwedi wa Tatu", "Mwedi wa Nchechi", "Mwedi wa Nnyano", "Mwedi wa Nnyano na Umo", "Mwedi wa Nnyano na Mivili", "Mwedi wa Nnyano na Mitatu", "Mwedi wa Nnyano na Nchechi", "Mwedi wa Nnyano na Nnyano", "Mwedi wa Nnyano na Nnyano na U", "Mwedi wa Nnyano na Nnyano na M"},
		daysAbbreviated:        []string{"Ll2", "Ll3", "Ll4", "Ll5", "Ll6", "Ll7", "Ll1"},
		daysNarrow:             []string{"2", "3", "4", "5", "6", "7", "1"},
		daysWide:               []string{"Liduva lyapili", "Liduva lyatatu", "Liduva lyanchechi", "Liduva lyannyano", "Liduva lyannyano na linji", "Liduva lyannyano na mavili", "Liduva litandi"},
		periodsAbbreviated:     []string{"Muhi", "Chilo"},
		periodsWide:            []string{"Muhi", "Chilo"},
		erasAbbreviated:        []string{"AY", "NY"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"Akanapawa Yesu", "Nankuida Yesu"},
		timezones:              map[string]string{"CDT": "CDT", "HKT": "HKT", "CLST": "CLST", "HNNOMX": "HNNOMX", "MST": "MST", "HNCU": "HNCU", "CST": "CST", "PST": "PST", "HKST": "HKST", "VET": "VET", "WIT": "WIT", "GYT": "GYT", "CHADT": "CHADT", "AST": "AST", "AEDT": "AEDT", "WESZ": "WESZ", "HEEG": "HEEG", "HEOG": "HEOG", "LHDT": "LHDT", "AWST": "AWST", "AWDT": "AWDT", "WIB": "WIB", "AKDT": "AKDT", "∅∅∅": "∅∅∅", "IST": "IST", "NZST": "NZST", "WARST": "WARST", "CLT": "CLT", "GFT": "GFT", "AKST": "AKST", "MEZ": "MEZ", "COST": "COST", "ACWDT": "ACWDT", "HNOG": "HNOG", "GMT": "GMT", "ACWST": "ACWST", "HENOMX": "HENOMX", "COT": "COT", "HNEG": "HNEG", "ART": "ART", "OESZ": "OESZ", "HAST": "HAST", "UYST": "UYST", "CHAST": "CHAST", "WEZ": "WEZ", "BOT": "BOT", "SGT": "SGT", "MDT": "MDT", "JDT": "JDT", "SRT": "SRT", "TMT": "TMT", "HNPMX": "HNPMX", "WAT": "WAT", "MYT": "MYT", "EDT": "EDT", "WART": "WART", "EAT": "EAT", "ARST": "ARST", "AEST": "AEST", "NZDT": "NZDT", "LHST": "LHST", "HNPM": "HNPM", "HAT": "HAT", "WITA": "WITA", "UYT": "UYT", "ChST": "ChST", "PDT": "PDT", "WAST": "WAST", "ACST": "ACST", "HNT": "HNT", "HECU": "HECU", "HEPMX": "HEPMX", "ADT": "ADT", "JST": "JST", "MESZ": "MESZ", "HEPM": "HEPM", "CAT": "CAT", "BT": "BT", "EST": "EST", "ACDT": "ACDT", "OEZ": "OEZ", "HADT": "HADT", "SAST": "SAST", "ECT": "ECT", "TMST": "TMST"},
	}
}

// Locale returns the current translators string locale
func (kde *kde) Locale() string {
	return kde.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'kde'
func (kde *kde) PluralsCardinal() []locales.PluralRule {
	return kde.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'kde'
func (kde *kde) PluralsOrdinal() []locales.PluralRule {
	return kde.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'kde'
func (kde *kde) PluralsRange() []locales.PluralRule {
	return kde.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'kde'
func (kde *kde) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'kde'
func (kde *kde) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'kde'
func (kde *kde) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (kde *kde) MonthAbbreviated(month time.Month) string {
	return kde.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (kde *kde) MonthsAbbreviated() []string {
	return kde.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (kde *kde) MonthNarrow(month time.Month) string {
	return kde.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (kde *kde) MonthsNarrow() []string {
	return kde.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (kde *kde) MonthWide(month time.Month) string {
	return kde.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (kde *kde) MonthsWide() []string {
	return kde.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (kde *kde) WeekdayAbbreviated(weekday time.Weekday) string {
	return kde.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (kde *kde) WeekdaysAbbreviated() []string {
	return kde.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (kde *kde) WeekdayNarrow(weekday time.Weekday) string {
	return kde.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (kde *kde) WeekdaysNarrow() []string {
	return kde.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (kde *kde) WeekdayShort(weekday time.Weekday) string {
	return kde.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (kde *kde) WeekdaysShort() []string {
	return kde.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (kde *kde) WeekdayWide(weekday time.Weekday) string {
	return kde.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (kde *kde) WeekdaysWide() []string {
	return kde.daysWide
}

// Decimal returns the decimal point of number
func (kde *kde) Decimal() string {
	return kde.decimal
}

// Group returns the group of number
func (kde *kde) Group() string {
	return kde.group
}

// Group returns the minus sign of number
func (kde *kde) Minus() string {
	return kde.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'kde' and handles both Whole and Real numbers based on 'v'
func (kde *kde) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'kde' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (kde *kde) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'kde'
func (kde *kde) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := kde.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, kde.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, kde.group[0])
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
		b = append(b, kde.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, kde.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'kde'
// in accounting notation.
func (kde *kde) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := kde.currencies[currency]
	l := len(s) + len(symbol) + 2
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, kde.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, kde.group[0])
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

		b = append(b, kde.currencyNegativePrefix[0])

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
			b = append(b, kde.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, kde.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'kde'
func (kde *kde) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'kde'
func (kde *kde) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, kde.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'kde'
func (kde *kde) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, kde.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'kde'
func (kde *kde) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, kde.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, kde.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'kde'
func (kde *kde) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, kde.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'kde'
func (kde *kde) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, kde.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, kde.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'kde'
func (kde *kde) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, kde.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, kde.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'kde'
func (kde *kde) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, kde.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, kde.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := kde.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
