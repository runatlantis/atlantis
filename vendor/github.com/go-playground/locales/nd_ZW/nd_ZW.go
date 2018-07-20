package nd_ZW

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type nd_ZW struct {
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

// New returns a new instance of translator for the 'nd_ZW' locale
func New() locales.Translator {
	return &nd_ZW{
		locale:                 "nd_ZW",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: ")",
		monthsAbbreviated:      []string{"", "Zib", "Nhlo", "Mbi", "Mab", "Nkw", "Nhla", "Ntu", "Ncw", "Mpan", "Mfu", "Lwe", "Mpal"},
		monthsNarrow:           []string{"", "Z", "N", "M", "M", "N", "N", "N", "N", "M", "M", "L", "M"},
		monthsWide:             []string{"", "Zibandlela", "Nhlolanja", "Mbimbitho", "Mabasa", "Nkwenkwezi", "Nhlangula", "Ntulikazi", "Ncwabakazi", "Mpandula", "Mfumfu", "Lwezi", "Mpalakazi"},
		daysAbbreviated:        []string{"Son", "Mvu", "Sib", "Sit", "Sin", "Sih", "Mgq"},
		daysNarrow:             []string{"S", "M", "S", "S", "S", "S", "M"},
		daysWide:               []string{"Sonto", "Mvulo", "Sibili", "Sithathu", "Sine", "Sihlanu", "Mgqibelo"},
		erasAbbreviated:        []string{"BC", "AD"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"UKristo angakabuyi", "Ukristo ebuyile"},
		timezones:              map[string]string{"ART": "ART", "PST": "PST", "AST": "AST", "WIB": "WIB", "HNEG": "HNEG", "MST": "MST", "UYST": "UYST", "HNCU": "HNCU", "HAT": "HAT", "OESZ": "OESZ", "COT": "COT", "GMT": "GMT", "AEST": "AEST", "MYT": "MYT", "ACWST": "ACWST", "CAT": "CAT", "HNPM": "HNPM", "HENOMX": "HENOMX", "TMT": "TMT", "COST": "COST", "CST": "CST", "ACDT": "ACDT", "WARST": "WARST", "MDT": "MDT", "BT": "BT", "NZST": "NZST", "HKST": "HKST", "∅∅∅": "∅∅∅", "CLT": "CLT", "ChST": "ChST", "ECT": "ECT", "HNOG": "HNOG", "MEZ": "MEZ", "VET": "VET", "SAST": "SAST", "AKDT": "AKDT", "HAST": "HAST", "SRT": "SRT", "AWST": "AWST", "ACST": "ACST", "WART": "WART", "LHDT": "LHDT", "WIT": "WIT", "CDT": "CDT", "AWDT": "AWDT", "ADT": "ADT", "NZDT": "NZDT", "IST": "IST", "UYT": "UYT", "AEDT": "AEDT", "EDT": "EDT", "HEOG": "HEOG", "HEPM": "HEPM", "WESZ": "WESZ", "BOT": "BOT", "JST": "JST", "AKST": "AKST", "MESZ": "MESZ", "WAT": "WAT", "EAT": "EAT", "CLST": "CLST", "HNPMX": "HNPMX", "HEPMX": "HEPMX", "GFT": "GFT", "HEEG": "HEEG", "HKT": "HKT", "HNT": "HNT", "OEZ": "OEZ", "HECU": "HECU", "ACWDT": "ACWDT", "TMST": "TMST", "HNNOMX": "HNNOMX", "HADT": "HADT", "ARST": "ARST", "GYT": "GYT", "CHADT": "CHADT", "WAST": "WAST", "SGT": "SGT", "LHST": "LHST", "EST": "EST", "WEZ": "WEZ", "JDT": "JDT", "CHAST": "CHAST", "PDT": "PDT", "WITA": "WITA"},
	}
}

// Locale returns the current translators string locale
func (nd *nd_ZW) Locale() string {
	return nd.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'nd_ZW'
func (nd *nd_ZW) PluralsCardinal() []locales.PluralRule {
	return nd.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'nd_ZW'
func (nd *nd_ZW) PluralsOrdinal() []locales.PluralRule {
	return nd.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'nd_ZW'
func (nd *nd_ZW) PluralsRange() []locales.PluralRule {
	return nd.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'nd_ZW'
func (nd *nd_ZW) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'nd_ZW'
func (nd *nd_ZW) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'nd_ZW'
func (nd *nd_ZW) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (nd *nd_ZW) MonthAbbreviated(month time.Month) string {
	return nd.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (nd *nd_ZW) MonthsAbbreviated() []string {
	return nd.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (nd *nd_ZW) MonthNarrow(month time.Month) string {
	return nd.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (nd *nd_ZW) MonthsNarrow() []string {
	return nd.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (nd *nd_ZW) MonthWide(month time.Month) string {
	return nd.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (nd *nd_ZW) MonthsWide() []string {
	return nd.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (nd *nd_ZW) WeekdayAbbreviated(weekday time.Weekday) string {
	return nd.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (nd *nd_ZW) WeekdaysAbbreviated() []string {
	return nd.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (nd *nd_ZW) WeekdayNarrow(weekday time.Weekday) string {
	return nd.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (nd *nd_ZW) WeekdaysNarrow() []string {
	return nd.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (nd *nd_ZW) WeekdayShort(weekday time.Weekday) string {
	return nd.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (nd *nd_ZW) WeekdaysShort() []string {
	return nd.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (nd *nd_ZW) WeekdayWide(weekday time.Weekday) string {
	return nd.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (nd *nd_ZW) WeekdaysWide() []string {
	return nd.daysWide
}

// Decimal returns the decimal point of number
func (nd *nd_ZW) Decimal() string {
	return nd.decimal
}

// Group returns the group of number
func (nd *nd_ZW) Group() string {
	return nd.group
}

// Group returns the minus sign of number
func (nd *nd_ZW) Minus() string {
	return nd.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'nd_ZW' and handles both Whole and Real numbers based on 'v'
func (nd *nd_ZW) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'nd_ZW' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (nd *nd_ZW) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'nd_ZW'
func (nd *nd_ZW) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := nd.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, nd.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, nd.group[0])
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
		b = append(b, nd.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, nd.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'nd_ZW'
// in accounting notation.
func (nd *nd_ZW) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := nd.currencies[currency]
	l := len(s) + len(symbol) + 2
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, nd.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, nd.group[0])
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

		b = append(b, nd.currencyNegativePrefix[0])

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
			b = append(b, nd.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, nd.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'nd_ZW'
func (nd *nd_ZW) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'nd_ZW'
func (nd *nd_ZW) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, nd.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'nd_ZW'
func (nd *nd_ZW) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, nd.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'nd_ZW'
func (nd *nd_ZW) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, nd.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, nd.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'nd_ZW'
func (nd *nd_ZW) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, nd.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'nd_ZW'
func (nd *nd_ZW) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, nd.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, nd.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'nd_ZW'
func (nd *nd_ZW) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, nd.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, nd.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'nd_ZW'
func (nd *nd_ZW) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, nd.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, nd.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := nd.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
