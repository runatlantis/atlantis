package tt_RU

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type tt_RU struct {
	locale             string
	pluralsCardinal    []locales.PluralRule
	pluralsOrdinal     []locales.PluralRule
	pluralsRange       []locales.PluralRule
	decimal            string
	group              string
	minus              string
	percent            string
	percentSuffix      string
	perMille           string
	timeSeparator      string
	inifinity          string
	currencies         []string // idx = enum of currency code
	monthsAbbreviated  []string
	monthsNarrow       []string
	monthsWide         []string
	daysAbbreviated    []string
	daysNarrow         []string
	daysShort          []string
	daysWide           []string
	periodsAbbreviated []string
	periodsNarrow      []string
	periodsShort       []string
	periodsWide        []string
	erasAbbreviated    []string
	erasNarrow         []string
	erasWide           []string
	timezones          map[string]string
}

// New returns a new instance of translator for the 'tt_RU' locale
func New() locales.Translator {
	return &tt_RU{
		locale:             "tt_RU",
		pluralsCardinal:    nil,
		pluralsOrdinal:     nil,
		pluralsRange:       nil,
		decimal:            ",",
		group:              " ",
		minus:              "-",
		percent:            "%",
		perMille:           "‰",
		timeSeparator:      ":",
		inifinity:          "∞",
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		percentSuffix:      " ",
		monthsAbbreviated:  []string{"", "гыйн.", "фев.", "мар.", "апр.", "май", "июнь", "июль", "авг.", "сент.", "окт.", "нояб.", "дек."},
		monthsNarrow:       []string{"", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"},
		monthsWide:         []string{"", "гыйнвар", "февраль", "март", "апрель", "май", "июнь", "июль", "август", "сентябрь", "октябрь", "ноябрь", "декабрь"},
		daysAbbreviated:    []string{"якш.", "дүш.", "сиш.", "чәр.", "пәнҗ.", "җом.", "шим."},
		daysNarrow:         []string{"Я", "Д", "С", "Ч", "П", "Җ", "Ш"},
		daysShort:          []string{"якш.", "дүш.", "сиш.", "чәр.", "пәнҗ.", "җом.", "шим."},
		daysWide:           []string{"якшәмбе", "дүшәмбе", "сишәмбе", "чәршәмбе", "пәнҗешәмбе", "җомга", "шимбә"},
		periodsAbbreviated: []string{"AM", "PM"},
		periodsNarrow:      []string{"AM", "PM"},
		periodsWide:        []string{"AM", "PM"},
		erasAbbreviated:    []string{"б.э.к.", "б.э."},
		erasNarrow:         []string{"", ""},
		erasWide:           []string{"безнең эрага кадәр", "безнең эра"},
		timezones:          map[string]string{"AWDT": "AWDT", "GYT": "GYT", "ChST": "ChST", "LHDT": "LHDT", "VET": "VET", "WAST": "WAST", "WESZ": "җәйге Көнбатыш Европа вакыты", "GFT": "GFT", "HKT": "HKT", "LHST": "LHST", "SGT": "SGT", "HENOMX": "HENOMX", "UYST": "UYST", "NZST": "NZST", "WITA": "WITA", "CLST": "CLST", "WIT": "WIT", "COST": "COST", "CST": "Төньяк Америка гадәти үзәк вакыты", "HNPMX": "HNPMX", "ECT": "ECT", "EDT": "Төньяк Америка җәйге көнчыгыш вакыты", "EAT": "EAT", "MDT": "Төньяк Америка җәйге тау вакыты", "JST": "JST", "IST": "IST", "HAST": "HAST", "MESZ": "җәйге Үзәк Европа вакыты", "UYT": "UYT", "ADT": "Төньяк Америка җәйге атлантик вакыты", "WAT": "WAT", "MYT": "MYT", "EST": "Төньяк Америка гадәти көнчыгыш вакыты", "HNEG": "HNEG", "PST": "Төньяк Америка гадәти Тын океан вакыты", "SAST": "SAST", "SRT": "SRT", "ARST": "ARST", "AKDT": "AKDT", "ACWDT": "ACWDT", "HEOG": "HEOG", "HNNOMX": "HNNOMX", "COT": "COT", "HNT": "HNT", "HEPM": "HEPM", "OESZ": "җәйге Көнчыгыш Европа вакыты", "HEPMX": "HEPMX", "AEST": "AEST", "BT": "BT", "AKST": "AKST", "HKST": "HKST", "HADT": "HADT", "AEDT": "AEDT", "MST": "Төньяк Америка гадәти тау вакыты", "WARST": "WARST", "CAT": "CAT", "OEZ": "гадәти Көнчыгыш Европа вакыты", "GMT": "Гринвич уртача вакыты", "HNCU": "HNCU", "PDT": "Төньяк Америка җәйге Тын океан вакыты", "AST": "Төньяк Америка гадәти атлантик вакыты", "NZDT": "NZDT", "BOT": "BOT", "AWST": "AWST", "WIB": "WIB", "JDT": "JDT", "MEZ": "гадәти Үзәк Европа вакыты", "TMT": "TMT", "TMST": "TMST", "ART": "ART", "HECU": "HECU", "CDT": "Төньяк Америка җәйге үзәк вакыты", "WEZ": "гадәти Көнбатыш Европа вакыты", "ACDT": "ACDT", "HNOG": "HNOG", "ACST": "ACST", "HAT": "HAT", "WART": "WART", "HNPM": "HNPM", "CLT": "CLT", "CHAST": "CHAST", "CHADT": "CHADT", "∅∅∅": "∅∅∅", "ACWST": "ACWST", "HEEG": "HEEG"},
	}
}

// Locale returns the current translators string locale
func (tt *tt_RU) Locale() string {
	return tt.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'tt_RU'
func (tt *tt_RU) PluralsCardinal() []locales.PluralRule {
	return tt.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'tt_RU'
func (tt *tt_RU) PluralsOrdinal() []locales.PluralRule {
	return tt.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'tt_RU'
func (tt *tt_RU) PluralsRange() []locales.PluralRule {
	return tt.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'tt_RU'
func (tt *tt_RU) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'tt_RU'
func (tt *tt_RU) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'tt_RU'
func (tt *tt_RU) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (tt *tt_RU) MonthAbbreviated(month time.Month) string {
	return tt.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (tt *tt_RU) MonthsAbbreviated() []string {
	return tt.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (tt *tt_RU) MonthNarrow(month time.Month) string {
	return tt.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (tt *tt_RU) MonthsNarrow() []string {
	return tt.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (tt *tt_RU) MonthWide(month time.Month) string {
	return tt.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (tt *tt_RU) MonthsWide() []string {
	return tt.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (tt *tt_RU) WeekdayAbbreviated(weekday time.Weekday) string {
	return tt.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (tt *tt_RU) WeekdaysAbbreviated() []string {
	return tt.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (tt *tt_RU) WeekdayNarrow(weekday time.Weekday) string {
	return tt.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (tt *tt_RU) WeekdaysNarrow() []string {
	return tt.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (tt *tt_RU) WeekdayShort(weekday time.Weekday) string {
	return tt.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (tt *tt_RU) WeekdaysShort() []string {
	return tt.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (tt *tt_RU) WeekdayWide(weekday time.Weekday) string {
	return tt.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (tt *tt_RU) WeekdaysWide() []string {
	return tt.daysWide
}

// Decimal returns the decimal point of number
func (tt *tt_RU) Decimal() string {
	return tt.decimal
}

// Group returns the group of number
func (tt *tt_RU) Group() string {
	return tt.group
}

// Group returns the minus sign of number
func (tt *tt_RU) Minus() string {
	return tt.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'tt_RU' and handles both Whole and Real numbers based on 'v'
func (tt *tt_RU) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, tt.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(tt.group) - 1; j >= 0; j-- {
					b = append(b, tt.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, tt.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'tt_RU' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (tt *tt_RU) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 5
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, tt.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, tt.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, tt.percentSuffix...)

	b = append(b, tt.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'tt_RU'
func (tt *tt_RU) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := tt.currencies[currency]
	l := len(s) + len(symbol) + 2 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, tt.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(tt.group) - 1; j >= 0; j-- {
					b = append(b, tt.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, tt.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, tt.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'tt_RU'
// in accounting notation.
func (tt *tt_RU) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := tt.currencies[currency]
	l := len(s) + len(symbol) + 2 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, tt.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(tt.group) - 1; j >= 0; j-- {
					b = append(b, tt.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, tt.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, tt.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, symbol...)
	} else {

		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'tt_RU'
func (tt *tt_RU) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Month() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Month()), 10)

	b = append(b, []byte{0x2e}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'tt_RU'
func (tt *tt_RU) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, tt.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20, 0xd0, 0xb5, 0xd0, 0xbb}...)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'tt_RU'
func (tt *tt_RU) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, tt.monthsWide[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20, 0xd0, 0xb5, 0xd0, 0xbb}...)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'tt_RU'
func (tt *tt_RU) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, tt.monthsWide[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20, 0xd0, 0xb5, 0xd0, 0xbb}...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, tt.daysWide[t.Weekday()]...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'tt_RU'
func (tt *tt_RU) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, tt.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'tt_RU'
func (tt *tt_RU) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, tt.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, tt.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'tt_RU'
func (tt *tt_RU) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, tt.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, tt.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'tt_RU'
func (tt *tt_RU) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, tt.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, tt.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := tt.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
