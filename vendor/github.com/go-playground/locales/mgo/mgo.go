package mgo

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type mgo struct {
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

// New returns a new instance of translator for the 'mgo' locale
func New() locales.Translator {
	return &mgo{
		locale:                 "mgo",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ".",
		group:                  ",",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositivePrefix: " ",
		currencyNegativePrefix: " ",
		monthsAbbreviated:      []string{"", "mbegtug", "imeg àbùbì", "imeg mbəŋchubi", "iməg ngwə̀t", "iməg fog", "iməg ichiibɔd", "iməg àdùmbə̀ŋ", "iməg ichika", "iməg kud", "iməg tèsiʼe", "iməg zò", "iməg krizmed"},
		monthsNarrow:           []string{"", "M1", "A2", "M3", "N4", "F5", "I6", "A7", "I8", "K9", "10", "11", "12"},
		monthsWide:             []string{"", "iməg mbegtug", "imeg àbùbì", "imeg mbəŋchubi", "iməg ngwə̀t", "iməg fog", "iməg ichiibɔd", "iməg àdùmbə̀ŋ", "iməg ichika", "iməg kud", "iməg tèsiʼe", "iməg zò", "iməg krizmed"},
		daysAbbreviated:        []string{"Aneg 1", "Aneg 2", "Aneg 3", "Aneg 4", "Aneg 5", "Aneg 6", "Aneg 7"},
		daysNarrow:             []string{"A1", "A2", "A3", "A4", "A5", "A6", "A7"},
		daysShort:              []string{"1", "2", "3", "4", "5", "6", "7"},
		daysWide:               []string{"Aneg 1", "Aneg 2", "Aneg 3", "Aneg 4", "Aneg 5", "Aneg 6", "Aneg 7"},
		periodsAbbreviated:     []string{"AM", "PM"},
		periodsWide:            []string{"AM", "PM"},
		erasAbbreviated:        []string{"BCE", "CE"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"", ""},
		timezones:              map[string]string{"SGT": "SGT", "EST": "EST", "ACDT": "ACDT", "IST": "IST", "WARST": "WARST", "CLT": "CLT", "HNOG": "HNOG", "HENOMX": "HENOMX", "OEZ": "OEZ", "HEPMX": "HEPMX", "PDT": "PDT", "AST": "AST", "BOT": "BOT", "AKST": "AKST", "LHST": "LHST", "VET": "VET", "TMST": "TMST", "CHAST": "CHAST", "ADT": "ADT", "MYT": "MYT", "JST": "JST", "∅∅∅": "∅∅∅", "COST": "COST", "ACWST": "ACWST", "HEPM": "HEPM", "HNNOMX": "HNNOMX", "AWST": "AWST", "AWDT": "AWDT", "SAST": "SAST", "JDT": "JDT", "ECT": "ECT", "EDT": "EDT", "CHADT": "CHADT", "NZDT": "NZDT", "GMT": "GMT", "CST": "CST", "AEST": "AEST", "AKDT": "AKDT", "HNPM": "HNPM", "PST": "PST", "WEZ": "WEZ", "HEEG": "HEEG", "MESZ": "MESZ", "LHDT": "LHDT", "ART": "ART", "ARST": "ARST", "WESZ": "WESZ", "WAT": "WAT", "WIB": "WIB", "GFT": "GFT", "HKT": "HKT", "HNT": "HNT", "GYT": "GYT", "HNPMX": "HNPMX", "AEDT": "AEDT", "HNEG": "HNEG", "MST": "MST", "MDT": "MDT", "CAT": "CAT", "CLST": "CLST", "UYST": "UYST", "COT": "COT", "ACST": "ACST", "HEOG": "HEOG", "WART": "WART", "HAT": "HAT", "EAT": "EAT", "HAST": "HAST", "WAST": "WAST", "NZST": "NZST", "HKST": "HKST", "UYT": "UYT", "ChST": "ChST", "HECU": "HECU", "BT": "BT", "TMT": "TMT", "OESZ": "OESZ", "HADT": "HADT", "HNCU": "HNCU", "CDT": "CDT", "ACWDT": "ACWDT", "MEZ": "MEZ", "WITA": "WITA", "SRT": "SRT", "WIT": "WIT"},
	}
}

// Locale returns the current translators string locale
func (mgo *mgo) Locale() string {
	return mgo.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'mgo'
func (mgo *mgo) PluralsCardinal() []locales.PluralRule {
	return mgo.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'mgo'
func (mgo *mgo) PluralsOrdinal() []locales.PluralRule {
	return mgo.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'mgo'
func (mgo *mgo) PluralsRange() []locales.PluralRule {
	return mgo.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'mgo'
func (mgo *mgo) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'mgo'
func (mgo *mgo) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'mgo'
func (mgo *mgo) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (mgo *mgo) MonthAbbreviated(month time.Month) string {
	return mgo.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (mgo *mgo) MonthsAbbreviated() []string {
	return mgo.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (mgo *mgo) MonthNarrow(month time.Month) string {
	return mgo.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (mgo *mgo) MonthsNarrow() []string {
	return mgo.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (mgo *mgo) MonthWide(month time.Month) string {
	return mgo.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (mgo *mgo) MonthsWide() []string {
	return mgo.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (mgo *mgo) WeekdayAbbreviated(weekday time.Weekday) string {
	return mgo.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (mgo *mgo) WeekdaysAbbreviated() []string {
	return mgo.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (mgo *mgo) WeekdayNarrow(weekday time.Weekday) string {
	return mgo.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (mgo *mgo) WeekdaysNarrow() []string {
	return mgo.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (mgo *mgo) WeekdayShort(weekday time.Weekday) string {
	return mgo.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (mgo *mgo) WeekdaysShort() []string {
	return mgo.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (mgo *mgo) WeekdayWide(weekday time.Weekday) string {
	return mgo.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (mgo *mgo) WeekdaysWide() []string {
	return mgo.daysWide
}

// Decimal returns the decimal point of number
func (mgo *mgo) Decimal() string {
	return mgo.decimal
}

// Group returns the group of number
func (mgo *mgo) Group() string {
	return mgo.group
}

// Group returns the minus sign of number
func (mgo *mgo) Minus() string {
	return mgo.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'mgo' and handles both Whole and Real numbers based on 'v'
func (mgo *mgo) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, mgo.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, mgo.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, mgo.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'mgo' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (mgo *mgo) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, mgo.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, mgo.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, mgo.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'mgo'
func (mgo *mgo) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := mgo.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, mgo.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, mgo.group[0])
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

	for j := len(mgo.currencyPositivePrefix) - 1; j >= 0; j-- {
		b = append(b, mgo.currencyPositivePrefix[j])
	}

	if num < 0 {
		b = append(b, mgo.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, mgo.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'mgo'
// in accounting notation.
func (mgo *mgo) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := mgo.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, mgo.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, mgo.group[0])
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

		for j := len(mgo.currencyNegativePrefix) - 1; j >= 0; j-- {
			b = append(b, mgo.currencyNegativePrefix[j])
		}

		b = append(b, mgo.minus[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(mgo.currencyPositivePrefix) - 1; j >= 0; j-- {
			b = append(b, mgo.currencyPositivePrefix[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, mgo.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'mgo'
func (mgo *mgo) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2d}...)

	if t.Month() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Month()), 10)

	b = append(b, []byte{0x2d}...)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'mgo'
func (mgo *mgo) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, mgo.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'mgo'
func (mgo *mgo) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, mgo.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'mgo'
func (mgo *mgo) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, mgo.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, mgo.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'mgo'
func (mgo *mgo) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, mgo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'mgo'
func (mgo *mgo) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, mgo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, mgo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'mgo'
func (mgo *mgo) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, mgo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, mgo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'mgo'
func (mgo *mgo) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, mgo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, mgo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := mgo.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
