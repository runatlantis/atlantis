package tg

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type tg struct {
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

// New returns a new instance of translator for the 'tg' locale
func New() locales.Translator {
	return &tg{
		locale:                 "tg",
		pluralsCardinal:        nil,
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ",",
		group:                  " ",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JP¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "сом.", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "Янв", "Фев", "Мар", "Апр", "Май", "Июн", "Июл", "Авг", "Сен", "Окт", "Ноя", "Дек"},
		monthsNarrow:           []string{"", "Я", "Ф", "М", "А", "М", "И", "И", "А", "С", "О", "Н", "Д"},
		monthsWide:             []string{"", "Январ", "Феврал", "Март", "Апрел", "Май", "Июн", "Июл", "Август", "Сентябр", "Октябр", "Ноябр", "Декабр"},
		daysAbbreviated:        []string{"Яшб", "Дшб", "Сшб", "Чшб", "Пшб", "Ҷмъ", "Шнб"},
		daysNarrow:             []string{"Я", "Д", "С", "Ч", "П", "Ҷ", "Ш"},
		daysShort:              []string{"Яшб", "Дшб", "Сшб", "Чшб", "Пшб", "Ҷмъ", "Шнб"},
		daysWide:               []string{"Якшанбе", "Душанбе", "Сешанбе", "Чоршанбе", "Панҷшанбе", "Ҷумъа", "Шанбе"},
		periodsAbbreviated:     []string{"пе. чо.", "па. чо."},
		periodsWide:            []string{"пе. чо.", "па. чо."},
		erasAbbreviated:        []string{"ПеМ", "ПаМ"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"Пеш аз милод", "ПаМ"},
		timezones:              map[string]string{"MEZ": "Вақти стандартии аврупоии марказӣ", "VET": "VET", "MST": "MST", "COST": "COST", "ChST": "ChST", "MDT": "MDT", "ARST": "ARST", "ACST": "ACST", "HAST": "HAST", "CST": "Вақти стандартии марказӣ", "PDT": "Вақти рӯзонаи Уқёнуси Ором", "HNPMX": "HNPMX", "ADT": "Вақти рӯзонаи атлантикӣ", "WAST": "WAST", "WIB": "WIB", "AKST": "AKST", "UYT": "UYT", "UYST": "UYST", "HNCU": "HNCU", "AST": "Вақти стандартии атлантикӣ", "OESZ": "Вақти тобистонаи аврупоии шарқӣ", "WEZ": "Вақти стандартии аврупоии ғарбӣ", "NZDT": "NZDT", "MYT": "MYT", "SGT": "SGT", "EDT": "Вақти рӯзонаи шарқӣ", "HNT": "HNT", "OEZ": "Вақти стандартии аврупоии шарқӣ", "BT": "BT", "ACWDT": "ACWDT", "LHST": "LHST", "WART": "WART", "HNPM": "HNPM", "WITA": "WITA", "WESZ": "Вақти тобистонаи аврупоии ғарбӣ", "NZST": "NZST", "IST": "IST", "COT": "COT", "SAST": "SAST", "ACWST": "ACWST", "MESZ": "Вақти тобистонаи аврупоии марказӣ", "ART": "ART", "GYT": "GYT", "AWST": "AWST", "JST": "JST", "HAT": "HAT", "CHADT": "CHADT", "HEPMX": "HEPMX", "HADT": "HADT", "BOT": "BOT", "ACDT": "ACDT", "HNEG": "HNEG", "HKT": "HKT", "WARST": "WARST", "HEPM": "HEPM", "CLT": "CLT", "GMT": "Ба вақти Гринвич", "∅∅∅": "∅∅∅", "WAT": "WAT", "TMST": "TMST", "PST": "Вақти стандартии Уқёнуси Ором", "GFT": "GFT", "JDT": "JDT", "HNNOMX": "HNNOMX", "HNOG": "HNOG", "HEOG": "HEOG", "CHAST": "CHAST", "WIT": "WIT", "ECT": "ECT", "EST": "Вақти стандартии шарқӣ", "HEEG": "HEEG", "HKST": "HKST", "HENOMX": "HENOMX", "TMT": "TMT", "CAT": "CAT", "CDT": "Вақти рӯзонаи марказӣ", "SRT": "SRT", "EAT": "EAT", "HECU": "HECU", "AWDT": "AWDT", "AEDT": "AEDT", "AKDT": "AKDT", "LHDT": "LHDT", "CLST": "CLST", "AEST": "AEST"},
	}
}

// Locale returns the current translators string locale
func (tg *tg) Locale() string {
	return tg.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'tg'
func (tg *tg) PluralsCardinal() []locales.PluralRule {
	return tg.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'tg'
func (tg *tg) PluralsOrdinal() []locales.PluralRule {
	return tg.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'tg'
func (tg *tg) PluralsRange() []locales.PluralRule {
	return tg.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'tg'
func (tg *tg) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'tg'
func (tg *tg) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'tg'
func (tg *tg) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (tg *tg) MonthAbbreviated(month time.Month) string {
	return tg.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (tg *tg) MonthsAbbreviated() []string {
	return tg.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (tg *tg) MonthNarrow(month time.Month) string {
	return tg.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (tg *tg) MonthsNarrow() []string {
	return tg.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (tg *tg) MonthWide(month time.Month) string {
	return tg.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (tg *tg) MonthsWide() []string {
	return tg.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (tg *tg) WeekdayAbbreviated(weekday time.Weekday) string {
	return tg.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (tg *tg) WeekdaysAbbreviated() []string {
	return tg.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (tg *tg) WeekdayNarrow(weekday time.Weekday) string {
	return tg.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (tg *tg) WeekdaysNarrow() []string {
	return tg.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (tg *tg) WeekdayShort(weekday time.Weekday) string {
	return tg.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (tg *tg) WeekdaysShort() []string {
	return tg.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (tg *tg) WeekdayWide(weekday time.Weekday) string {
	return tg.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (tg *tg) WeekdaysWide() []string {
	return tg.daysWide
}

// Decimal returns the decimal point of number
func (tg *tg) Decimal() string {
	return tg.decimal
}

// Group returns the group of number
func (tg *tg) Group() string {
	return tg.group
}

// Group returns the minus sign of number
func (tg *tg) Minus() string {
	return tg.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'tg' and handles both Whole and Real numbers based on 'v'
func (tg *tg) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, tg.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(tg.group) - 1; j >= 0; j-- {
					b = append(b, tg.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, tg.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'tg' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (tg *tg) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, tg.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, tg.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, tg.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'tg'
func (tg *tg) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := tg.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, tg.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(tg.group) - 1; j >= 0; j-- {
					b = append(b, tg.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, tg.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, tg.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, tg.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'tg'
// in accounting notation.
func (tg *tg) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := tg.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, tg.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(tg.group) - 1; j >= 0; j-- {
					b = append(b, tg.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, tg.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, tg.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, tg.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, tg.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'tg'
func (tg *tg) FmtDateShort(t time.Time) string {

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

	if t.Year() > 9 {
		b = append(b, strconv.Itoa(t.Year())[2:]...)
	} else {
		b = append(b, strconv.Itoa(t.Year())[1:]...)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'tg'
func (tg *tg) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, tg.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'tg'
func (tg *tg) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, tg.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'tg'
func (tg *tg) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, tg.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, tg.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'tg'
func (tg *tg) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, tg.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'tg'
func (tg *tg) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, tg.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, tg.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'tg'
func (tg *tg) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, tg.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, tg.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'tg'
func (tg *tg) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, tg.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, tg.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := tg.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
