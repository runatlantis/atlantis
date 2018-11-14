package nd

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type nd struct {
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

// New returns a new instance of translator for the 'nd' locale
func New() locales.Translator {
	return &nd{
		locale:                 "nd",
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
		timezones:              map[string]string{"MDT": "MDT", "COST": "COST", "HEEG": "HEEG", "LHST": "LHST", "HNNOMX": "HNNOMX", "SRT": "SRT", "ADT": "ADT", "NZDT": "NZDT", "AKDT": "AKDT", "WART": "WART", "WARST": "WARST", "WITA": "WITA", "AWDT": "AWDT", "NZST": "NZST", "ACDT": "ACDT", "HNOG": "HNOG", "HNPM": "HNPM", "CHADT": "CHADT", "PST": "PST", "AEDT": "AEDT", "JDT": "JDT", "TMT": "TMT", "AKST": "AKST", "HKST": "HKST", "∅∅∅": "∅∅∅", "GMT": "GMT", "CDT": "CDT", "AEST": "AEST", "CST": "CST", "AST": "AST", "BOT": "BOT", "ACST": "ACST", "WIT": "WIT", "OEZ": "OEZ", "COT": "COT", "ChST": "ChST", "ART": "ART", "GYT": "GYT", "AWST": "AWST", "EDT": "EDT", "HAT": "HAT", "HEPM": "HEPM", "WAST": "WAST", "WEZ": "WEZ", "ECT": "ECT", "HNT": "HNT", "MST": "MST", "EAT": "EAT", "PDT": "PDT", "WESZ": "WESZ", "JST": "JST", "BT": "BT", "SGT": "SGT", "VET": "VET", "TMST": "TMST", "ARST": "ARST", "UYT": "UYT", "WIB": "WIB", "UYST": "UYST", "ACWST": "ACWST", "HNEG": "HNEG", "HEOG": "HEOG", "CLT": "CLT", "OESZ": "OESZ", "HAST": "HAST", "HADT": "HADT", "HENOMX": "HENOMX", "EST": "EST", "HKT": "HKT", "LHDT": "LHDT", "CAT": "CAT", "WAT": "WAT", "ACWDT": "ACWDT", "MEZ": "MEZ", "HNPMX": "HNPMX", "HEPMX": "HEPMX", "SAST": "SAST", "GFT": "GFT", "MYT": "MYT", "MESZ": "MESZ", "IST": "IST", "CLST": "CLST", "CHAST": "CHAST", "HNCU": "HNCU", "HECU": "HECU"},
	}
}

// Locale returns the current translators string locale
func (nd *nd) Locale() string {
	return nd.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'nd'
func (nd *nd) PluralsCardinal() []locales.PluralRule {
	return nd.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'nd'
func (nd *nd) PluralsOrdinal() []locales.PluralRule {
	return nd.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'nd'
func (nd *nd) PluralsRange() []locales.PluralRule {
	return nd.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'nd'
func (nd *nd) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'nd'
func (nd *nd) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'nd'
func (nd *nd) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (nd *nd) MonthAbbreviated(month time.Month) string {
	return nd.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (nd *nd) MonthsAbbreviated() []string {
	return nd.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (nd *nd) MonthNarrow(month time.Month) string {
	return nd.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (nd *nd) MonthsNarrow() []string {
	return nd.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (nd *nd) MonthWide(month time.Month) string {
	return nd.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (nd *nd) MonthsWide() []string {
	return nd.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (nd *nd) WeekdayAbbreviated(weekday time.Weekday) string {
	return nd.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (nd *nd) WeekdaysAbbreviated() []string {
	return nd.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (nd *nd) WeekdayNarrow(weekday time.Weekday) string {
	return nd.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (nd *nd) WeekdaysNarrow() []string {
	return nd.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (nd *nd) WeekdayShort(weekday time.Weekday) string {
	return nd.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (nd *nd) WeekdaysShort() []string {
	return nd.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (nd *nd) WeekdayWide(weekday time.Weekday) string {
	return nd.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (nd *nd) WeekdaysWide() []string {
	return nd.daysWide
}

// Decimal returns the decimal point of number
func (nd *nd) Decimal() string {
	return nd.decimal
}

// Group returns the group of number
func (nd *nd) Group() string {
	return nd.group
}

// Group returns the minus sign of number
func (nd *nd) Minus() string {
	return nd.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'nd' and handles both Whole and Real numbers based on 'v'
func (nd *nd) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'nd' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (nd *nd) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'nd'
func (nd *nd) FmtCurrency(num float64, v uint64, currency currency.Type) string {

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

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'nd'
// in accounting notation.
func (nd *nd) FmtAccounting(num float64, v uint64, currency currency.Type) string {

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

// FmtDateShort returns the short date representation of 't' for 'nd'
func (nd *nd) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'nd'
func (nd *nd) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'nd'
func (nd *nd) FmtDateLong(t time.Time) string {

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

// FmtDateFull returns the full date representation of 't' for 'nd'
func (nd *nd) FmtDateFull(t time.Time) string {

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

// FmtTimeShort returns the short time representation of 't' for 'nd'
func (nd *nd) FmtTimeShort(t time.Time) string {

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

// FmtTimeMedium returns the medium time representation of 't' for 'nd'
func (nd *nd) FmtTimeMedium(t time.Time) string {

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

// FmtTimeLong returns the long time representation of 't' for 'nd'
func (nd *nd) FmtTimeLong(t time.Time) string {

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

// FmtTimeFull returns the full time representation of 't' for 'nd'
func (nd *nd) FmtTimeFull(t time.Time) string {

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
