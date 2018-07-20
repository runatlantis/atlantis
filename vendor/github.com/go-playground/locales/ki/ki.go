package ki

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type ki struct {
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

// New returns a new instance of translator for the 'ki' locale
func New() locales.Translator {
	return &ki{
		locale:                 "ki",
		pluralsCardinal:        nil,
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "Ksh", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: ")",
		monthsAbbreviated:      []string{"", "JEN", "WKR", "WGT", "WKN", "WTN", "WTD", "WMJ", "WNN", "WKD", "WIK", "WMW", "DIT"},
		monthsNarrow:           []string{"", "J", "K", "G", "K", "G", "G", "M", "K", "K", "I", "I", "D"},
		monthsWide:             []string{"", "Njenuarĩ", "Mwere wa kerĩ", "Mwere wa gatatũ", "Mwere wa kana", "Mwere wa gatano", "Mwere wa gatandatũ", "Mwere wa mũgwanja", "Mwere wa kanana", "Mwere wa kenda", "Mwere wa ikũmi", "Mwere wa ikũmi na ũmwe", "Ndithemba"},
		daysAbbreviated:        []string{"KMA", "NTT", "NMN", "NMT", "ART", "NMA", "NMM"},
		daysNarrow:             []string{"K", "N", "N", "N", "A", "N", "N"},
		daysWide:               []string{"Kiumia", "Njumatatũ", "Njumaine", "Njumatana", "Aramithi", "Njumaa", "Njumamothi"},
		periodsAbbreviated:     []string{"Kiroko", "Hwaĩ-inĩ"},
		periodsWide:            []string{"Kiroko", "Hwaĩ-inĩ"},
		erasAbbreviated:        []string{"MK", "TK"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"Mbere ya Kristo", "Thutha wa Kristo"},
		timezones:              map[string]string{"HKST": "HKST", "ChST": "ChST", "PST": "PST", "AEST": "AEST", "SGT": "SGT", "HEEG": "HEEG", "HEPM": "HEPM", "CAT": "CAT", "CHADT": "CHADT", "BOT": "BOT", "HNEG": "HNEG", "ACST": "ACST", "GFT": "GFT", "WIT": "WIT", "OESZ": "OESZ", "ARST": "ARST", "GMT": "GMT", "CST": "CST", "LHST": "LHST", "WART": "WART", "WESZ": "WESZ", "NZST": "NZST", "NZDT": "NZDT", "JST": "JST", "AKST": "AKST", "CDT": "CDT", "AWST": "AWST", "MDT": "MDT", "WITA": "WITA", "JDT": "JDT", "HADT": "HADT", "COT": "COT", "HNPMX": "HNPMX", "SAST": "SAST", "MYT": "MYT", "EST": "EST", "MEZ": "MEZ", "HNT": "HNT", "OEZ": "OEZ", "UYST": "UYST", "CHAST": "CHAST", "PDT": "PDT", "WIB": "WIB", "HEOG": "HEOG", "HKT": "HKT", "HENOMX": "HENOMX", "TMT": "TMT", "∅∅∅": "∅∅∅", "AKDT": "AKDT", "HNOG": "HNOG", "AWDT": "AWDT", "HEPMX": "HEPMX", "MST": "MST", "LHDT": "LHDT", "WARST": "WARST", "HAT": "HAT", "SRT": "SRT", "AEDT": "AEDT", "WEZ": "WEZ", "ACWST": "ACWST", "IST": "IST", "HAST": "HAST", "ART": "ART", "UYT": "UYT", "HECU": "HECU", "ACDT": "ACDT", "MESZ": "MESZ", "ACWDT": "ACWDT", "HNPM": "HNPM", "TMST": "TMST", "COST": "COST", "HNCU": "HNCU", "BT": "BT", "ECT": "ECT", "CLT": "CLT", "ADT": "ADT", "WAT": "WAT", "HNNOMX": "HNNOMX", "VET": "VET", "EAT": "EAT", "CLST": "CLST", "GYT": "GYT", "AST": "AST", "WAST": "WAST", "EDT": "EDT"},
	}
}

// Locale returns the current translators string locale
func (ki *ki) Locale() string {
	return ki.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'ki'
func (ki *ki) PluralsCardinal() []locales.PluralRule {
	return ki.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'ki'
func (ki *ki) PluralsOrdinal() []locales.PluralRule {
	return ki.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'ki'
func (ki *ki) PluralsRange() []locales.PluralRule {
	return ki.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'ki'
func (ki *ki) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'ki'
func (ki *ki) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'ki'
func (ki *ki) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (ki *ki) MonthAbbreviated(month time.Month) string {
	return ki.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (ki *ki) MonthsAbbreviated() []string {
	return ki.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (ki *ki) MonthNarrow(month time.Month) string {
	return ki.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (ki *ki) MonthsNarrow() []string {
	return ki.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (ki *ki) MonthWide(month time.Month) string {
	return ki.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (ki *ki) MonthsWide() []string {
	return ki.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (ki *ki) WeekdayAbbreviated(weekday time.Weekday) string {
	return ki.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (ki *ki) WeekdaysAbbreviated() []string {
	return ki.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (ki *ki) WeekdayNarrow(weekday time.Weekday) string {
	return ki.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (ki *ki) WeekdaysNarrow() []string {
	return ki.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (ki *ki) WeekdayShort(weekday time.Weekday) string {
	return ki.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (ki *ki) WeekdaysShort() []string {
	return ki.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (ki *ki) WeekdayWide(weekday time.Weekday) string {
	return ki.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (ki *ki) WeekdaysWide() []string {
	return ki.daysWide
}

// Decimal returns the decimal point of number
func (ki *ki) Decimal() string {
	return ki.decimal
}

// Group returns the group of number
func (ki *ki) Group() string {
	return ki.group
}

// Group returns the minus sign of number
func (ki *ki) Minus() string {
	return ki.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'ki' and handles both Whole and Real numbers based on 'v'
func (ki *ki) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'ki' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (ki *ki) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'ki'
func (ki *ki) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ki.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ki.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ki.group[0])
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
		b = append(b, ki.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ki.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'ki'
// in accounting notation.
func (ki *ki) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ki.currencies[currency]
	l := len(s) + len(symbol) + 2
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ki.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ki.group[0])
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

		b = append(b, ki.currencyNegativePrefix[0])

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
			b = append(b, ki.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, ki.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'ki'
func (ki *ki) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'ki'
func (ki *ki) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ki.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'ki'
func (ki *ki) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ki.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'ki'
func (ki *ki) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, ki.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ki.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'ki'
func (ki *ki) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ki.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'ki'
func (ki *ki) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ki.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ki.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'ki'
func (ki *ki) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ki.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ki.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'ki'
func (ki *ki) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ki.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ki.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := ki.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
