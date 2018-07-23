package kam_KE

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type kam_KE struct {
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

// New returns a new instance of translator for the 'kam_KE' locale
func New() locales.Translator {
	return &kam_KE{
		locale:                 "kam_KE",
		pluralsCardinal:        nil,
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: ")",
		monthsAbbreviated:      []string{"", "Mbe", "Kel", "Ktũ", "Kan", "Ktn", "Tha", "Moo", "Nya", "Knd", "Ĩku", "Ĩkm", "Ĩkl"},
		monthsNarrow:           []string{"", "M", "K", "K", "K", "K", "T", "M", "N", "K", "Ĩ", "Ĩ", "Ĩ"},
		monthsWide:             []string{"", "Mwai wa mbee", "Mwai wa kelĩ", "Mwai wa katatũ", "Mwai wa kana", "Mwai wa katano", "Mwai wa thanthatũ", "Mwai wa muonza", "Mwai wa nyaanya", "Mwai wa kenda", "Mwai wa ĩkumi", "Mwai wa ĩkumi na ĩmwe", "Mwai wa ĩkumi na ilĩ"},
		daysAbbreviated:        []string{"Wky", "Wkw", "Wkl", "Wtũ", "Wkn", "Wtn", "Wth"},
		daysNarrow:             []string{"Y", "W", "E", "A", "A", "A", "A"},
		daysWide:               []string{"Wa kyumwa", "Wa kwambĩlĩlya", "Wa kelĩ", "Wa katatũ", "Wa kana", "Wa katano", "Wa thanthatũ"},
		periodsAbbreviated:     []string{"Ĩyakwakya", "Ĩyawĩoo"},
		periodsWide:            []string{"Ĩyakwakya", "Ĩyawĩoo"},
		erasAbbreviated:        []string{"MY", "IY"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"Mbee wa Yesũ", "Ĩtina wa Yesũ"},
		timezones:              map[string]string{"HNNOMX": "HNNOMX", "MST": "MST", "CAT": "CAT", "ACWDT": "ACWDT", "HKST": "HKST", "∅∅∅": "∅∅∅", "LHST": "LHST", "COST": "COST", "GYT": "GYT", "AEDT": "AEDT", "SGT": "SGT", "TMT": "TMT", "HAST": "HAST", "ARST": "ARST", "PST": "PST", "LHDT": "LHDT", "WART": "WART", "WITA": "WITA", "ART": "ART", "AWDT": "AWDT", "HEPMX": "HEPMX", "AST": "AST", "HEEG": "HEEG", "CLT": "CLT", "UYT": "UYT", "CHADT": "CHADT", "AWST": "AWST", "ECT": "ECT", "EDT": "EDT", "HNOG": "HNOG", "CDT": "CDT", "OESZ": "OESZ", "ChST": "ChST", "WESZ": "WESZ", "BT": "BT", "AKST": "AKST", "WARST": "WARST", "HAT": "HAT", "SRT": "SRT", "GFT": "GFT", "BOT": "BOT", "TMST": "TMST", "HADT": "HADT", "HECU": "HECU", "MYT": "MYT", "PDT": "PDT", "SAST": "SAST", "MDT": "MDT", "COT": "COT", "CHAST": "CHAST", "CST": "CST", "VET": "VET", "OEZ": "OEZ", "JDT": "JDT", "NZST": "NZST", "ACST": "ACST", "HNPM": "HNPM", "HEPM": "HEPM", "WAST": "WAST", "EST": "EST", "WIT": "WIT", "ADT": "ADT", "AEST": "AEST", "NZDT": "NZDT", "HNEG": "HNEG", "IST": "IST", "HENOMX": "HENOMX", "EAT": "EAT", "AKDT": "AKDT", "ACDT": "ACDT", "ACWST": "ACWST", "UYST": "UYST", "HNCU": "HNCU", "WEZ": "WEZ", "HEOG": "HEOG", "MESZ": "MESZ", "HKT": "HKT", "CLST": "CLST", "HNT": "HNT", "GMT": "GMT", "WAT": "WAT", "WIB": "WIB", "MEZ": "MEZ", "HNPMX": "HNPMX", "JST": "JST"},
	}
}

// Locale returns the current translators string locale
func (kam *kam_KE) Locale() string {
	return kam.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'kam_KE'
func (kam *kam_KE) PluralsCardinal() []locales.PluralRule {
	return kam.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'kam_KE'
func (kam *kam_KE) PluralsOrdinal() []locales.PluralRule {
	return kam.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'kam_KE'
func (kam *kam_KE) PluralsRange() []locales.PluralRule {
	return kam.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'kam_KE'
func (kam *kam_KE) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'kam_KE'
func (kam *kam_KE) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'kam_KE'
func (kam *kam_KE) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (kam *kam_KE) MonthAbbreviated(month time.Month) string {
	return kam.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (kam *kam_KE) MonthsAbbreviated() []string {
	return kam.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (kam *kam_KE) MonthNarrow(month time.Month) string {
	return kam.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (kam *kam_KE) MonthsNarrow() []string {
	return kam.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (kam *kam_KE) MonthWide(month time.Month) string {
	return kam.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (kam *kam_KE) MonthsWide() []string {
	return kam.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (kam *kam_KE) WeekdayAbbreviated(weekday time.Weekday) string {
	return kam.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (kam *kam_KE) WeekdaysAbbreviated() []string {
	return kam.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (kam *kam_KE) WeekdayNarrow(weekday time.Weekday) string {
	return kam.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (kam *kam_KE) WeekdaysNarrow() []string {
	return kam.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (kam *kam_KE) WeekdayShort(weekday time.Weekday) string {
	return kam.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (kam *kam_KE) WeekdaysShort() []string {
	return kam.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (kam *kam_KE) WeekdayWide(weekday time.Weekday) string {
	return kam.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (kam *kam_KE) WeekdaysWide() []string {
	return kam.daysWide
}

// Decimal returns the decimal point of number
func (kam *kam_KE) Decimal() string {
	return kam.decimal
}

// Group returns the group of number
func (kam *kam_KE) Group() string {
	return kam.group
}

// Group returns the minus sign of number
func (kam *kam_KE) Minus() string {
	return kam.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'kam_KE' and handles both Whole and Real numbers based on 'v'
func (kam *kam_KE) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'kam_KE' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (kam *kam_KE) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'kam_KE'
func (kam *kam_KE) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := kam.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, kam.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, kam.group[0])
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
		b = append(b, kam.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, kam.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'kam_KE'
// in accounting notation.
func (kam *kam_KE) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := kam.currencies[currency]
	l := len(s) + len(symbol) + 2
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, kam.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, kam.group[0])
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

		b = append(b, kam.currencyNegativePrefix[0])

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
			b = append(b, kam.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, kam.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'kam_KE'
func (kam *kam_KE) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'kam_KE'
func (kam *kam_KE) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, kam.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'kam_KE'
func (kam *kam_KE) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, kam.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'kam_KE'
func (kam *kam_KE) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, kam.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, kam.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'kam_KE'
func (kam *kam_KE) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, kam.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'kam_KE'
func (kam *kam_KE) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, kam.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, kam.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'kam_KE'
func (kam *kam_KE) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, kam.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, kam.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'kam_KE'
func (kam *kam_KE) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, kam.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, kam.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := kam.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
