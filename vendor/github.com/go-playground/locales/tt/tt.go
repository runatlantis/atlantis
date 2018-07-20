package tt

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type tt struct {
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

// New returns a new instance of translator for the 'tt' locale
func New() locales.Translator {
	return &tt{
		locale:             "tt",
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
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JP¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "₽", "р.", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
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
		timezones:          map[string]string{"COT": "COT", "CHAST": "CHAST", "WIB": "WIB", "MYT": "MYT", "HEOG": "HEOG", "HKT": "HKT", "LHDT": "LHDT", "CST": "Төньяк Америка гадәти үзәк вакыты", "AWDT": "AWDT", "ACWST": "ACWST", "MESZ": "җәйге Үзәк Европа вакыты", "WART": "WART", "∅∅∅": "∅∅∅", "PDT": "Төньяк Америка җәйге Тын океан вакыты", "HNPMX": "HNPMX", "AST": "Төньяк Америка гадәти атлантик вакыты", "AEST": "AEST", "SAST": "SAST", "WAT": "WAT", "HNPM": "HNPM", "WAST": "WAST", "JDT": "JDT", "AKST": "AKST", "CHADT": "CHADT", "AEDT": "AEDT", "HNT": "HNT", "HEPM": "HEPM", "SRT": "SRT", "EST": "Төньяк Америка гадәти көнчыгыш вакыты", "MST": "MST", "NZDT": "NZDT", "BOT": "BOT", "SGT": "SGT", "ACST": "ACST", "WARST": "WARST", "WITA": "WITA", "HNNOMX": "HNNOMX", "CAT": "CAT", "CLT": "CLT", "UYT": "UYT", "HECU": "HECU", "ADT": "Төньяк Америка җәйге атлантик вакыты", "WEZ": "гадәти Көнбатыш Европа вакыты", "GFT": "GFT", "HNEG": "HNEG", "WESZ": "җәйге Көнбатыш Европа вакыты", "AKDT": "AKDT", "EDT": "Төньяк Америка җәйге көнчыгыш вакыты", "ACWDT": "ACWDT", "TMT": "TMT", "HNCU": "HNCU", "AWST": "AWST", "ECT": "ECT", "HEEG": "HEEG", "MEZ": "гадәти Үзәк Европа вакыты", "LHST": "LHST", "OESZ": "җәйге Көнчыгыш Европа вакыты", "COST": "COST", "JST": "JST", "IST": "IST", "VET": "VET", "EAT": "EAT", "TMST": "TMST", "ART": "ART", "UYST": "UYST", "ACDT": "ACDT", "HENOMX": "HENOMX", "MDT": "MDT", "CLST": "CLST", "WIT": "WIT", "HAST": "HAST", "ARST": "ARST", "HEPMX": "HEPMX", "BT": "BT", "NZST": "NZST", "HNOG": "HNOG", "OEZ": "гадәти Көнчыгыш Европа вакыты", "HADT": "HADT", "ChST": "ChST", "PST": "Төньяк Америка гадәти Тын океан вакыты", "GMT": "Гринвич уртача вакыты", "GYT": "GYT", "CDT": "Төньяк Америка җәйге үзәк вакыты", "HKST": "HKST", "HAT": "HAT"},
	}
}

// Locale returns the current translators string locale
func (tt *tt) Locale() string {
	return tt.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'tt'
func (tt *tt) PluralsCardinal() []locales.PluralRule {
	return tt.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'tt'
func (tt *tt) PluralsOrdinal() []locales.PluralRule {
	return tt.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'tt'
func (tt *tt) PluralsRange() []locales.PluralRule {
	return tt.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'tt'
func (tt *tt) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'tt'
func (tt *tt) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'tt'
func (tt *tt) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (tt *tt) MonthAbbreviated(month time.Month) string {
	return tt.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (tt *tt) MonthsAbbreviated() []string {
	return tt.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (tt *tt) MonthNarrow(month time.Month) string {
	return tt.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (tt *tt) MonthsNarrow() []string {
	return tt.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (tt *tt) MonthWide(month time.Month) string {
	return tt.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (tt *tt) MonthsWide() []string {
	return tt.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (tt *tt) WeekdayAbbreviated(weekday time.Weekday) string {
	return tt.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (tt *tt) WeekdaysAbbreviated() []string {
	return tt.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (tt *tt) WeekdayNarrow(weekday time.Weekday) string {
	return tt.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (tt *tt) WeekdaysNarrow() []string {
	return tt.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (tt *tt) WeekdayShort(weekday time.Weekday) string {
	return tt.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (tt *tt) WeekdaysShort() []string {
	return tt.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (tt *tt) WeekdayWide(weekday time.Weekday) string {
	return tt.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (tt *tt) WeekdaysWide() []string {
	return tt.daysWide
}

// Decimal returns the decimal point of number
func (tt *tt) Decimal() string {
	return tt.decimal
}

// Group returns the group of number
func (tt *tt) Group() string {
	return tt.group
}

// Group returns the minus sign of number
func (tt *tt) Minus() string {
	return tt.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'tt' and handles both Whole and Real numbers based on 'v'
func (tt *tt) FmtNumber(num float64, v uint64) string {

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

// FmtPercent returns 'num' with digits/precision of 'v' for 'tt' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (tt *tt) FmtPercent(num float64, v uint64) string {
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

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'tt'
func (tt *tt) FmtCurrency(num float64, v uint64, currency currency.Type) string {

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

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'tt'
// in accounting notation.
func (tt *tt) FmtAccounting(num float64, v uint64, currency currency.Type) string {

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

// FmtDateShort returns the short date representation of 't' for 'tt'
func (tt *tt) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'tt'
func (tt *tt) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'tt'
func (tt *tt) FmtDateLong(t time.Time) string {

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

// FmtDateFull returns the full date representation of 't' for 'tt'
func (tt *tt) FmtDateFull(t time.Time) string {

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

// FmtTimeShort returns the short time representation of 't' for 'tt'
func (tt *tt) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, tt.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'tt'
func (tt *tt) FmtTimeMedium(t time.Time) string {

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

// FmtTimeLong returns the long time representation of 't' for 'tt'
func (tt *tt) FmtTimeLong(t time.Time) string {

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

// FmtTimeFull returns the full time representation of 't' for 'tt'
func (tt *tt) FmtTimeFull(t time.Time) string {

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
