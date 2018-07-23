package saq_KE

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type saq_KE struct {
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

// New returns a new instance of translator for the 'saq_KE' locale
func New() locales.Translator {
	return &saq_KE{
		locale:                 "saq_KE",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: ")",
		monthsAbbreviated:      []string{"", "Obo", "Waa", "Oku", "Ong", "Ime", "Ile", "Sap", "Isi", "Saa", "Tom", "Tob", "Tow"},
		monthsNarrow:           []string{"", "O", "W", "O", "O", "I", "I", "S", "I", "S", "T", "T", "T"},
		monthsWide:             []string{"", "Lapa le obo", "Lapa le waare", "Lapa le okuni", "Lapa le ong’wan", "Lapa le imet", "Lapa le ile", "Lapa le sapa", "Lapa le isiet", "Lapa le saal", "Lapa le tomon", "Lapa le tomon obo", "Lapa le tomon waare"},
		daysAbbreviated:        []string{"Are", "Kun", "Ong", "Ine", "Ile", "Sap", "Kwe"},
		daysNarrow:             []string{"A", "K", "O", "I", "I", "S", "K"},
		daysWide:               []string{"Mderot ee are", "Mderot ee kuni", "Mderot ee ong’wan", "Mderot ee inet", "Mderot ee ile", "Mderot ee sapa", "Mderot ee kwe"},
		periodsAbbreviated:     []string{"Tesiran", "Teipa"},
		periodsWide:            []string{"Tesiran", "Teipa"},
		erasAbbreviated:        []string{"KK", "BK"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"Kabla ya Christo", "Baada ya Christo"},
		timezones:              map[string]string{"COST": "COST", "HEPMX": "HEPMX", "AEDT": "AEDT", "NZST": "NZST", "HNPM": "HNPM", "BOT": "BOT", "CLT": "CLT", "HKST": "HKST", "AWDT": "AWDT", "HNPMX": "HNPMX", "AEST": "AEST", "AKST": "AKST", "WESZ": "WESZ", "GFT": "GFT", "UYST": "UYST", "ChST": "ChST", "ECT": "ECT", "LHST": "LHST", "WART": "WART", "WARST": "WARST", "SAST": "SAST", "WIB": "WIB", "JDT": "JDT", "BT": "BT", "GMT": "GMT", "CHAST": "CHAST", "ARST": "ARST", "CHADT": "CHADT", "HECU": "HECU", "SGT": "SGT", "TMST": "TMST", "EAT": "EAT", "HAST": "HAST", "LHDT": "LHDT", "PDT": "PDT", "MDT": "MDT", "EDT": "EDT", "IST": "IST", "HEEG": "HEEG", "MEZ": "MEZ", "HENOMX": "HENOMX", "CST": "CST", "WEZ": "WEZ", "HNOG": "HNOG", "ACWDT": "ACWDT", "VET": "VET", "CLST": "CLST", "PST": "PST", "WAST": "WAST", "JST": "JST", "NZDT": "NZDT", "HNEG": "HNEG", "HEPM": "HEPM", "ART": "ART", "OESZ": "OESZ", "AKDT": "AKDT", "HEOG": "HEOG", "HNNOMX": "HNNOMX", "SRT": "SRT", "WITA": "WITA", "WIT": "WIT", "COT": "COT", "GYT": "GYT", "CDT": "CDT", "ACST": "ACST", "HKT": "HKT", "HNT": "HNT", "CAT": "CAT", "ADT": "ADT", "MST": "MST", "MYT": "MYT", "HAT": "HAT", "ACDT": "ACDT", "OEZ": "OEZ", "HADT": "HADT", "UYT": "UYT", "AWST": "AWST", "∅∅∅": "∅∅∅", "WAT": "WAT", "EST": "EST", "MESZ": "MESZ", "AST": "AST", "ACWST": "ACWST", "TMT": "TMT", "HNCU": "HNCU"},
	}
}

// Locale returns the current translators string locale
func (saq *saq_KE) Locale() string {
	return saq.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'saq_KE'
func (saq *saq_KE) PluralsCardinal() []locales.PluralRule {
	return saq.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'saq_KE'
func (saq *saq_KE) PluralsOrdinal() []locales.PluralRule {
	return saq.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'saq_KE'
func (saq *saq_KE) PluralsRange() []locales.PluralRule {
	return saq.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'saq_KE'
func (saq *saq_KE) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'saq_KE'
func (saq *saq_KE) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'saq_KE'
func (saq *saq_KE) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (saq *saq_KE) MonthAbbreviated(month time.Month) string {
	return saq.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (saq *saq_KE) MonthsAbbreviated() []string {
	return saq.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (saq *saq_KE) MonthNarrow(month time.Month) string {
	return saq.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (saq *saq_KE) MonthsNarrow() []string {
	return saq.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (saq *saq_KE) MonthWide(month time.Month) string {
	return saq.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (saq *saq_KE) MonthsWide() []string {
	return saq.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (saq *saq_KE) WeekdayAbbreviated(weekday time.Weekday) string {
	return saq.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (saq *saq_KE) WeekdaysAbbreviated() []string {
	return saq.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (saq *saq_KE) WeekdayNarrow(weekday time.Weekday) string {
	return saq.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (saq *saq_KE) WeekdaysNarrow() []string {
	return saq.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (saq *saq_KE) WeekdayShort(weekday time.Weekday) string {
	return saq.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (saq *saq_KE) WeekdaysShort() []string {
	return saq.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (saq *saq_KE) WeekdayWide(weekday time.Weekday) string {
	return saq.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (saq *saq_KE) WeekdaysWide() []string {
	return saq.daysWide
}

// Decimal returns the decimal point of number
func (saq *saq_KE) Decimal() string {
	return saq.decimal
}

// Group returns the group of number
func (saq *saq_KE) Group() string {
	return saq.group
}

// Group returns the minus sign of number
func (saq *saq_KE) Minus() string {
	return saq.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'saq_KE' and handles both Whole and Real numbers based on 'v'
func (saq *saq_KE) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'saq_KE' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (saq *saq_KE) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'saq_KE'
func (saq *saq_KE) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := saq.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, saq.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, saq.group[0])
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
		b = append(b, saq.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, saq.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'saq_KE'
// in accounting notation.
func (saq *saq_KE) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := saq.currencies[currency]
	l := len(s) + len(symbol) + 2
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, saq.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, saq.group[0])
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

		b = append(b, saq.currencyNegativePrefix[0])

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
			b = append(b, saq.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, saq.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'saq_KE'
func (saq *saq_KE) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'saq_KE'
func (saq *saq_KE) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, saq.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'saq_KE'
func (saq *saq_KE) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, saq.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'saq_KE'
func (saq *saq_KE) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, saq.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, saq.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'saq_KE'
func (saq *saq_KE) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, saq.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'saq_KE'
func (saq *saq_KE) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, saq.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, saq.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'saq_KE'
func (saq *saq_KE) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, saq.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, saq.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'saq_KE'
func (saq *saq_KE) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, saq.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, saq.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := saq.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
