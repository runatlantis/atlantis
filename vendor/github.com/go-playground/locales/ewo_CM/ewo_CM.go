package ewo_CM

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type ewo_CM struct {
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

// New returns a new instance of translator for the 'ewo_CM' locale
func New() locales.Translator {
	return &ewo_CM{
		locale:                 "ewo_CM",
		pluralsCardinal:        nil,
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ",",
		group:                  " ",
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "ngo", "ngb", "ngl", "ngn", "ngt", "ngs", "ngz", "ngm", "nge", "nga", "ngad", "ngab"},
		monthsNarrow:           []string{"", "o", "b", "l", "n", "t", "s", "z", "m", "e", "a", "d", "b"},
		monthsWide:             []string{"", "ngɔn osú", "ngɔn bɛ̌", "ngɔn lála", "ngɔn nyina", "ngɔn tána", "ngɔn saməna", "ngɔn zamgbála", "ngɔn mwom", "ngɔn ebulú", "ngɔn awóm", "ngɔn awóm ai dziá", "ngɔn awóm ai bɛ̌"},
		daysAbbreviated:        []string{"sɔ́n", "mɔ́n", "smb", "sml", "smn", "fúl", "sér"},
		daysNarrow:             []string{"s", "m", "s", "s", "s", "f", "s"},
		daysWide:               []string{"sɔ́ndɔ", "mɔ́ndi", "sɔ́ndɔ məlú mə́bɛ̌", "sɔ́ndɔ məlú mə́lɛ́", "sɔ́ndɔ məlú mə́nyi", "fúladé", "séradé"},
		periodsAbbreviated:     []string{"kíkíríg", "ngəgógəle"},
		periodsWide:            []string{"kíkíríg", "ngəgógəle"},
		erasAbbreviated:        []string{"oyk", "ayk"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"osúsúa Yésus kiri", "ámvus Yésus Kirís"},
		timezones:              map[string]string{"MESZ": "MESZ", "TMT": "TMT", "TMST": "TMST", "UYT": "UYT", "AEDT": "AEDT", "JST": "JST", "WART": "WART", "VET": "VET", "ChST": "ChST", "AWDT": "AWDT", "AST": "AST", "ACDT": "ACDT", "HNNOMX": "HNNOMX", "HENOMX": "HENOMX", "MDT": "MDT", "WEZ": "WEZ", "MYT": "MYT", "EDT": "EDT", "HEPM": "HEPM", "CAT": "CAT", "CLST": "CLST", "HADT": "HADT", "GMT": "GMT", "SAST": "SAST", "ECT": "ECT", "HNOG": "HNOG", "IST": "IST", "OESZ": "OESZ", "AEST": "AEST", "AKDT": "AKDT", "SGT": "SGT", "ART": "ART", "CST": "CST", "WAST": "WAST", "JDT": "JDT", "EST": "EST", "ACST": "ACST", "HKST": "HKST", "MST": "MST", "OEZ": "OEZ", "PST": "PST", "PDT": "PDT", "∅∅∅": "∅∅∅", "HNCU": "HNCU", "HNPMX": "HNPMX", "HEPMX": "HEPMX", "MEZ": "MEZ", "HNT": "HNT", "EAT": "EAT", "GYT": "GYT", "CHAST": "CHAST", "WAT": "WAT", "NZDT": "NZDT", "LHDT": "LHDT", "WITA": "WITA", "COST": "COST", "ACWDT": "ACWDT", "COT": "COT", "UYST": "UYST", "HECU": "HECU", "HKT": "HKT", "ARST": "ARST", "AWST": "AWST", "ADT": "ADT", "WESZ": "WESZ", "GFT": "GFT", "NZST": "NZST", "HEOG": "HEOG", "HEEG": "HEEG", "SRT": "SRT", "CLT": "CLT", "WIT": "WIT", "WIB": "WIB", "BOT": "BOT", "ACWST": "ACWST", "HAST": "HAST", "AKST": "AKST", "LHST": "LHST", "WARST": "WARST", "HAT": "HAT", "HNPM": "HNPM", "CHADT": "CHADT", "CDT": "CDT", "BT": "BT", "HNEG": "HNEG"},
	}
}

// Locale returns the current translators string locale
func (ewo *ewo_CM) Locale() string {
	return ewo.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'ewo_CM'
func (ewo *ewo_CM) PluralsCardinal() []locales.PluralRule {
	return ewo.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'ewo_CM'
func (ewo *ewo_CM) PluralsOrdinal() []locales.PluralRule {
	return ewo.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'ewo_CM'
func (ewo *ewo_CM) PluralsRange() []locales.PluralRule {
	return ewo.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'ewo_CM'
func (ewo *ewo_CM) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'ewo_CM'
func (ewo *ewo_CM) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'ewo_CM'
func (ewo *ewo_CM) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (ewo *ewo_CM) MonthAbbreviated(month time.Month) string {
	return ewo.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (ewo *ewo_CM) MonthsAbbreviated() []string {
	return ewo.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (ewo *ewo_CM) MonthNarrow(month time.Month) string {
	return ewo.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (ewo *ewo_CM) MonthsNarrow() []string {
	return ewo.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (ewo *ewo_CM) MonthWide(month time.Month) string {
	return ewo.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (ewo *ewo_CM) MonthsWide() []string {
	return ewo.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (ewo *ewo_CM) WeekdayAbbreviated(weekday time.Weekday) string {
	return ewo.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (ewo *ewo_CM) WeekdaysAbbreviated() []string {
	return ewo.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (ewo *ewo_CM) WeekdayNarrow(weekday time.Weekday) string {
	return ewo.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (ewo *ewo_CM) WeekdaysNarrow() []string {
	return ewo.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (ewo *ewo_CM) WeekdayShort(weekday time.Weekday) string {
	return ewo.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (ewo *ewo_CM) WeekdaysShort() []string {
	return ewo.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (ewo *ewo_CM) WeekdayWide(weekday time.Weekday) string {
	return ewo.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (ewo *ewo_CM) WeekdaysWide() []string {
	return ewo.daysWide
}

// Decimal returns the decimal point of number
func (ewo *ewo_CM) Decimal() string {
	return ewo.decimal
}

// Group returns the group of number
func (ewo *ewo_CM) Group() string {
	return ewo.group
}

// Group returns the minus sign of number
func (ewo *ewo_CM) Minus() string {
	return ewo.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'ewo_CM' and handles both Whole and Real numbers based on 'v'
func (ewo *ewo_CM) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 1 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ewo.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(ewo.group) - 1; j >= 0; j-- {
					b = append(b, ewo.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ewo.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'ewo_CM' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (ewo *ewo_CM) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 1
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ewo.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ewo.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, ewo.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'ewo_CM'
func (ewo *ewo_CM) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ewo.currencies[currency]
	l := len(s) + len(symbol) + 3 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ewo.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(ewo.group) - 1; j >= 0; j-- {
					b = append(b, ewo.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ewo.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ewo.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, ewo.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'ewo_CM'
// in accounting notation.
func (ewo *ewo_CM) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ewo.currencies[currency]
	l := len(s) + len(symbol) + 3 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ewo.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(ewo.group) - 1; j >= 0; j-- {
					b = append(b, ewo.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, ewo.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ewo.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, ewo.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, ewo.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'ewo_CM'
func (ewo *ewo_CM) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2f}...)
	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0x2f}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'ewo_CM'
func (ewo *ewo_CM) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ewo.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'ewo_CM'
func (ewo *ewo_CM) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ewo.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'ewo_CM'
func (ewo *ewo_CM) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, ewo.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ewo.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'ewo_CM'
func (ewo *ewo_CM) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ewo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'ewo_CM'
func (ewo *ewo_CM) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ewo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ewo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'ewo_CM'
func (ewo *ewo_CM) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ewo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ewo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'ewo_CM'
func (ewo *ewo_CM) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ewo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ewo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := ewo.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
