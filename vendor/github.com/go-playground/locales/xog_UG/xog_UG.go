package xog_UG

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type xog_UG struct {
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

// New returns a new instance of translator for the 'xog_UG' locale
func New() locales.Translator {
	return &xog_UG{
		locale:                 "xog_UG",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "Jan", "Feb", "Mar", "Apu", "Maa", "Juu", "Jul", "Agu", "Seb", "Oki", "Nov", "Des"},
		monthsNarrow:           []string{"", "J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"},
		monthsWide:             []string{"", "Janwaliyo", "Febwaliyo", "Marisi", "Apuli", "Maayi", "Juuni", "Julaayi", "Agusito", "Sebuttemba", "Okitobba", "Novemba", "Desemba"},
		daysAbbreviated:        []string{"Sabi", "Bala", "Kubi", "Kusa", "Kuna", "Kuta", "Muka"},
		daysNarrow:             []string{"S", "B", "B", "S", "K", "K", "M"},
		daysWide:               []string{"Sabiiti", "Balaza", "Owokubili", "Owokusatu", "Olokuna", "Olokutaanu", "Olomukaaga"},
		periodsAbbreviated:     []string{"Munkyo", "Eigulo"},
		periodsWide:            []string{"Munkyo", "Eigulo"},
		erasAbbreviated:        []string{"AZ", "AF"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"Kulisto nga azilawo", "Kulisto nga affile"},
		timezones:              map[string]string{"ART": "ART", "AEDT": "AEDT", "JST": "JST", "HKST": "HKST", "CHADT": "CHADT", "AWDT": "AWDT", "AEST": "AEST", "WAT": "WAT", "WEZ": "WEZ", "HNT": "HNT", "HNNOMX": "HNNOMX", "EAT": "EAT", "CDT": "CDT", "SGT": "SGT", "HEEG": "HEEG", "WARST": "WARST", "WITA": "WITA", "NZST": "NZST", "MYT": "MYT", "∅∅∅": "∅∅∅", "WESZ": "WESZ", "HNEG": "HNEG", "IST": "IST", "HNPM": "HNPM", "GFT": "GFT", "MESZ": "MESZ", "MDT": "MDT", "SRT": "SRT", "CAT": "CAT", "CLST": "CLST", "OESZ": "OESZ", "COST": "COST", "UYT": "UYT", "HNOG": "HNOG", "SAST": "SAST", "CHAST": "CHAST", "ACDT": "ACDT", "LHST": "LHST", "WART": "WART", "HADT": "HADT", "HECU": "HECU", "PST": "PST", "AWST": "AWST", "ACST": "ACST", "TMST": "TMST", "HNCU": "HNCU", "NZDT": "NZDT", "AKDT": "AKDT", "HEOG": "HEOG", "HENOMX": "HENOMX", "HEPMX": "HEPMX", "ADT": "ADT", "JDT": "JDT", "AKST": "AKST", "HKT": "HKT", "HEPM": "HEPM", "CST": "CST", "PDT": "PDT", "MEZ": "MEZ", "WIT": "WIT", "UYST": "UYST", "VET": "VET", "HNPMX": "HNPMX", "BT": "BT", "BOT": "BOT", "MST": "MST", "ACWST": "ACWST", "ACWDT": "ACWDT", "TMT": "TMT", "COT": "COT", "AST": "AST", "WAST": "WAST", "EST": "EST", "LHDT": "LHDT", "CLT": "CLT", "OEZ": "OEZ", "GMT": "GMT", "ECT": "ECT", "EDT": "EDT", "HAST": "HAST", "ARST": "ARST", "GYT": "GYT", "ChST": "ChST", "WIB": "WIB", "HAT": "HAT"},
	}
}

// Locale returns the current translators string locale
func (xog *xog_UG) Locale() string {
	return xog.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'xog_UG'
func (xog *xog_UG) PluralsCardinal() []locales.PluralRule {
	return xog.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'xog_UG'
func (xog *xog_UG) PluralsOrdinal() []locales.PluralRule {
	return xog.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'xog_UG'
func (xog *xog_UG) PluralsRange() []locales.PluralRule {
	return xog.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'xog_UG'
func (xog *xog_UG) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'xog_UG'
func (xog *xog_UG) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'xog_UG'
func (xog *xog_UG) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (xog *xog_UG) MonthAbbreviated(month time.Month) string {
	return xog.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (xog *xog_UG) MonthsAbbreviated() []string {
	return xog.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (xog *xog_UG) MonthNarrow(month time.Month) string {
	return xog.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (xog *xog_UG) MonthsNarrow() []string {
	return xog.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (xog *xog_UG) MonthWide(month time.Month) string {
	return xog.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (xog *xog_UG) MonthsWide() []string {
	return xog.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (xog *xog_UG) WeekdayAbbreviated(weekday time.Weekday) string {
	return xog.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (xog *xog_UG) WeekdaysAbbreviated() []string {
	return xog.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (xog *xog_UG) WeekdayNarrow(weekday time.Weekday) string {
	return xog.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (xog *xog_UG) WeekdaysNarrow() []string {
	return xog.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (xog *xog_UG) WeekdayShort(weekday time.Weekday) string {
	return xog.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (xog *xog_UG) WeekdaysShort() []string {
	return xog.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (xog *xog_UG) WeekdayWide(weekday time.Weekday) string {
	return xog.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (xog *xog_UG) WeekdaysWide() []string {
	return xog.daysWide
}

// Decimal returns the decimal point of number
func (xog *xog_UG) Decimal() string {
	return xog.decimal
}

// Group returns the group of number
func (xog *xog_UG) Group() string {
	return xog.group
}

// Group returns the minus sign of number
func (xog *xog_UG) Minus() string {
	return xog.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'xog_UG' and handles both Whole and Real numbers based on 'v'
func (xog *xog_UG) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'xog_UG' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (xog *xog_UG) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'xog_UG'
func (xog *xog_UG) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := xog.currencies[currency]
	l := len(s) + len(symbol) + 2
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, xog.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, xog.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, xog.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, xog.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, xog.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'xog_UG'
// in accounting notation.
func (xog *xog_UG) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := xog.currencies[currency]
	l := len(s) + len(symbol) + 2
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, xog.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, xog.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, xog.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, xog.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, xog.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, xog.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'xog_UG'
func (xog *xog_UG) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'xog_UG'
func (xog *xog_UG) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, xog.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'xog_UG'
func (xog *xog_UG) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, xog.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'xog_UG'
func (xog *xog_UG) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, xog.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, xog.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'xog_UG'
func (xog *xog_UG) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, xog.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'xog_UG'
func (xog *xog_UG) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, xog.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, xog.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'xog_UG'
func (xog *xog_UG) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, xog.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, xog.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'xog_UG'
func (xog *xog_UG) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, xog.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, xog.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := xog.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
