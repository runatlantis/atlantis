package kok

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type kok struct {
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

// New returns a new instance of translator for the 'kok' locale
func New() locales.Translator {
	return &kok{
		locale:                 "kok",
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
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JP¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "US$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositivePrefix: " ",
		currencyNegativePrefix: " ",
		monthsAbbreviated:      []string{"", "जानेवारी", "फेब्रुवारी", "मार्च", "एप्रिल", "मे", "जून", "जुलाय", "आगोस्त", "सप्टेंबर", "ऑक्टोबर", "नोव्हेंबर", "डिसेंबर"},
		monthsNarrow:           []string{"", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"},
		monthsWide:             []string{"", "जानेवारी", "फेब्रुवारी", "मार्च", "एप्रिल", "मे", "जून", "जुलाय", "आगोस्त", "सप्टेंबर", "ऑक्टोबर", "नोव्हेंबर", "डिसेंबर"},
		daysAbbreviated:        []string{"आयतार", "सोमार", "मंगळार", "बुधवार", "गुरुवार", "शुक्रार", "शेनवार"},
		daysNarrow:             []string{"आ", "सो", "मं", "बु", "गु", "शु", "शे"},
		daysShort:              []string{"आय", "सोम", "मंगळ", "बुध", "गुरु", "शुक्र", "शेन"},
		daysWide:               []string{"आयतार", "सोमार", "मंगळार", "बुधवार", "गुरुवार", "शुक्रार", "शेनवार"},
		periodsAbbreviated:     []string{"म.पू.", "म.नं."},
		periodsNarrow:          []string{"म.पू.", "म.नं."},
		periodsWide:            []string{"म.पू.", "म.नं."},
		erasAbbreviated:        []string{"क्रिस्तपूर्व", "क्रिस्तशखा"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"क्रिस्तपूर्व", "क्रिस्तशखा"},
		timezones:              map[string]string{"WIT": "WIT", "TMT": "TMT", "CST": "CST", "AEST": "AEST", "WART": "WART", "HNT": "HNT", "MDT": "MDT", "CLT": "CLT", "BOT": "BOT", "SRT": "SRT", "AST": "AST", "EDT": "EDT", "LHDT": "LHDT", "COT": "COT", "JST": "JST", "ACWDT": "ACWDT", "IST": "भारतीय समय", "CAT": "CAT", "OEZ": "OEZ", "HNEG": "HNEG", "ChST": "ChST", "CDT": "CDT", "ACDT": "ACDT", "HEOG": "HEOG", "MST": "MST", "CLST": "CLST", "HADT": "HADT", "ART": "ART", "SGT": "SGT", "GYT": "GYT", "UYT": "UYT", "HECU": "HECU", "WAT": "WAT", "MEZ": "MEZ", "COST": "COST", "WAST": "WAST", "AKDT": "AKDT", "NZDT": "NZDT", "ECT": "ECT", "AKST": "AKST", "ACWST": "ACWST", "ARST": "ARST", "UYST": "UYST", "GMT": "GMT", "ADT": "ADT", "HEEG": "HEEG", "HNOG": "HNOG", "PST": "PST", "PDT": "PDT", "WIB": "WIB", "MYT": "MYT", "WARST": "WARST", "HEPMX": "HEPMX", "WESZ": "WESZ", "EST": "EST", "AEDT": "AEDT", "GFT": "GFT", "HKT": "HKT", "HNNOMX": "HNNOMX", "OESZ": "OESZ", "HNPMX": "HNPMX", "AWST": "AWST", "AWDT": "AWDT", "WEZ": "WEZ", "BT": "BT", "HKST": "HKST", "MESZ": "MESZ", "HNPM": "HNPM", "WITA": "WITA", "ACST": "ACST", "NZST": "NZST", "LHST": "LHST", "VET": "VET", "HEPM": "HEPM", "HENOMX": "HENOMX", "EAT": "EAT", "JDT": "JDT", "SAST": "SAST", "∅∅∅": "∅∅∅", "HAT": "HAT", "HAST": "HAST", "CHAST": "CHAST", "TMST": "TMST", "HNCU": "HNCU", "CHADT": "CHADT"},
	}
}

// Locale returns the current translators string locale
func (kok *kok) Locale() string {
	return kok.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'kok'
func (kok *kok) PluralsCardinal() []locales.PluralRule {
	return kok.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'kok'
func (kok *kok) PluralsOrdinal() []locales.PluralRule {
	return kok.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'kok'
func (kok *kok) PluralsRange() []locales.PluralRule {
	return kok.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'kok'
func (kok *kok) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'kok'
func (kok *kok) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'kok'
func (kok *kok) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (kok *kok) MonthAbbreviated(month time.Month) string {
	return kok.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (kok *kok) MonthsAbbreviated() []string {
	return kok.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (kok *kok) MonthNarrow(month time.Month) string {
	return kok.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (kok *kok) MonthsNarrow() []string {
	return kok.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (kok *kok) MonthWide(month time.Month) string {
	return kok.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (kok *kok) MonthsWide() []string {
	return kok.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (kok *kok) WeekdayAbbreviated(weekday time.Weekday) string {
	return kok.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (kok *kok) WeekdaysAbbreviated() []string {
	return kok.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (kok *kok) WeekdayNarrow(weekday time.Weekday) string {
	return kok.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (kok *kok) WeekdaysNarrow() []string {
	return kok.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (kok *kok) WeekdayShort(weekday time.Weekday) string {
	return kok.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (kok *kok) WeekdaysShort() []string {
	return kok.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (kok *kok) WeekdayWide(weekday time.Weekday) string {
	return kok.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (kok *kok) WeekdaysWide() []string {
	return kok.daysWide
}

// Decimal returns the decimal point of number
func (kok *kok) Decimal() string {
	return kok.decimal
}

// Group returns the group of number
func (kok *kok) Group() string {
	return kok.group
}

// Group returns the minus sign of number
func (kok *kok) Minus() string {
	return kok.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'kok' and handles both Whole and Real numbers based on 'v'
func (kok *kok) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	inSecondary := false
	groupThreshold := 3

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, kok.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {

			if count == groupThreshold {
				b = append(b, kok.group[0])
				count = 1

				if !inSecondary {
					inSecondary = true
					groupThreshold = 2
				}
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, kok.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'kok' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (kok *kok) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, kok.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, kok.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, kok.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'kok'
func (kok *kok) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := kok.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	inSecondary := false
	groupThreshold := 3

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, kok.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {

			if count == groupThreshold {
				b = append(b, kok.group[0])
				count = 1

				if !inSecondary {
					inSecondary = true
					groupThreshold = 2
				}
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	for j := len(symbol) - 1; j >= 0; j-- {
		b = append(b, symbol[j])
	}

	for j := len(kok.currencyPositivePrefix) - 1; j >= 0; j-- {
		b = append(b, kok.currencyPositivePrefix[j])
	}

	if num < 0 {
		b = append(b, kok.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, kok.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'kok'
// in accounting notation.
func (kok *kok) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := kok.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	inSecondary := false
	groupThreshold := 3

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, kok.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {

			if count == groupThreshold {
				b = append(b, kok.group[0])
				count = 1

				if !inSecondary {
					inSecondary = true
					groupThreshold = 2
				}
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

		for j := len(kok.currencyNegativePrefix) - 1; j >= 0; j-- {
			b = append(b, kok.currencyNegativePrefix[j])
		}

		b = append(b, kok.minus[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(kok.currencyPositivePrefix) - 1; j >= 0; j-- {
			b = append(b, kok.currencyPositivePrefix[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, kok.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'kok'
func (kok *kok) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2d}...)
	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0x2d}...)

	if t.Year() > 9 {
		b = append(b, strconv.Itoa(t.Year())[2:]...)
	} else {
		b = append(b, strconv.Itoa(t.Year())[1:]...)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'kok'
func (kok *kok) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2d}...)

	if t.Month() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Month()), 10)

	b = append(b, []byte{0x2d}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'kok'
func (kok *kok) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, kok.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'kok'
func (kok *kok) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, kok.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, kok.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'kok'
func (kok *kok) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, kok.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, kok.periodsAbbreviated[0]...)
	} else {
		b = append(b, kok.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'kok'
func (kok *kok) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, kok.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, kok.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, kok.periodsAbbreviated[0]...)
	} else {
		b = append(b, kok.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'kok'
func (kok *kok) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, kok.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, kok.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, kok.periodsAbbreviated[0]...)
	} else {
		b = append(b, kok.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'kok'
func (kok *kok) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, kok.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, kok.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, kok.periodsAbbreviated[0]...)
	} else {
		b = append(b, kok.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := kok.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
