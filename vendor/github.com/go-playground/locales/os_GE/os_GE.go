package os_GE

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type os_GE struct {
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

// New returns a new instance of translator for the 'os_GE' locale
func New() locales.Translator {
	return &os_GE{
		locale:                 "os_GE",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ",",
		group:                  " ",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositivePrefix: " ",
		currencyNegativePrefix: " ",
		monthsAbbreviated:      []string{"", "янв.", "фев.", "мар.", "апр.", "майы", "июны", "июлы", "авг.", "сен.", "окт.", "ноя.", "дек."},
		monthsNarrow:           []string{"", "Я", "Ф", "М", "А", "М", "И", "И", "А", "С", "О", "Н", "Д"},
		monthsWide:             []string{"", "январы", "февралы", "мартъийы", "апрелы", "майы", "июны", "июлы", "августы", "сентябры", "октябры", "ноябры", "декабры"},
		daysAbbreviated:        []string{"хцб", "крс", "дцг", "ӕрт", "цпр", "мрб", "сбт"},
		daysNarrow:             []string{"Х", "К", "Д", "Ӕ", "Ц", "М", "С"},
		daysWide:               []string{"хуыцаубон", "къуырисӕр", "дыццӕг", "ӕртыццӕг", "цыппӕрӕм", "майрӕмбон", "сабат"},
		periodsAbbreviated:     []string{"AM", "PM"},
		periodsWide:            []string{"ӕмбисбоны размӕ", "ӕмбисбоны фӕстӕ"},
		erasAbbreviated:        []string{"н.д.а.", "н.д."},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"", ""},
		timezones:              map[string]string{"SRT": "SRT", "JDT": "JDT", "ACST": "ACST", "HKT": "HKT", "LHDT": "LHDT", "HAT": "HAT", "CHAST": "CHAST", "GFT": "GFT", "WARST": "WARST", "HNNOMX": "HNNOMX", "HNT": "HNT", "HENOMX": "HENOMX", "∅∅∅": "∅∅∅", "CDT": "CDT", "HNOG": "HNOG", "HEOG": "HEOG", "CLT": "CLT", "GYT": "GYT", "AST": "AST", "SAST": "SAST", "WAT": "WAT", "WESZ": "Ныгъуылӕн Европӕйаг сӕрдыгон рӕстӕг", "MESZ": "Астӕуккаг Европӕйаг сӕрдыгон рӕстӕг", "MDT": "MDT", "CLST": "CLST", "OESZ": "Скӕсӕн Европӕйаг сӕрдыгон рӕстӕг", "ACWDT": "ACWDT", "MEZ": "Астӕуккаг Европӕйаг стандартон рӕстӕг", "VET": "VET", "MYT": "MYT", "HECU": "HECU", "CST": "CST", "PDT": "PDT", "ADT": "ADT", "WIB": "WIB", "WEZ": "Ныгъуылӕн Европӕйаг стандартон рӕстӕг", "MST": "MST", "COT": "COT", "HNCU": "HNCU", "AKST": "AKST", "AKDT": "AKDT", "HNPM": "HNPM", "HEPM": "HEPM", "TMST": "TMST", "CAT": "CAT", "HAST": "HAST", "AWST": "AWST", "ECT": "ECT", "HKST": "HKST", "UYST": "UYST", "PST": "PST", "WAST": "WAST", "ACDT": "ACDT", "IST": "IST", "AEDT": "AEDT", "NZST": "NZST", "SGT": "SGT", "EST": "EST", "WART": "WART", "WITA": "WITA", "EAT": "EAT", "ART": "ART", "AWDT": "AWDT", "BT": "BT", "HNEG": "HNEG", "ACWST": "ACWST", "WIT": "WIT", "HADT": "HADT", "ChST": "ChST", "HEEG": "HEEG", "TMT": "TMT", "GMT": "Гринвичы рӕстӕмбис рӕстӕг", "HNPMX": "HNPMX", "HEPMX": "HEPMX", "NZDT": "NZDT", "EDT": "EDT", "ARST": "ARST", "UYT": "UYT", "AEST": "AEST", "COST": "COST", "OEZ": "Скӕсӕн Европӕйаг стандартон рӕстӕг", "CHADT": "CHADT", "BOT": "BOT", "JST": "JST", "LHST": "LHST"},
	}
}

// Locale returns the current translators string locale
func (os *os_GE) Locale() string {
	return os.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'os_GE'
func (os *os_GE) PluralsCardinal() []locales.PluralRule {
	return os.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'os_GE'
func (os *os_GE) PluralsOrdinal() []locales.PluralRule {
	return os.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'os_GE'
func (os *os_GE) PluralsRange() []locales.PluralRule {
	return os.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'os_GE'
func (os *os_GE) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'os_GE'
func (os *os_GE) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'os_GE'
func (os *os_GE) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (os *os_GE) MonthAbbreviated(month time.Month) string {
	return os.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (os *os_GE) MonthsAbbreviated() []string {
	return os.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (os *os_GE) MonthNarrow(month time.Month) string {
	return os.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (os *os_GE) MonthsNarrow() []string {
	return os.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (os *os_GE) MonthWide(month time.Month) string {
	return os.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (os *os_GE) MonthsWide() []string {
	return os.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (os *os_GE) WeekdayAbbreviated(weekday time.Weekday) string {
	return os.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (os *os_GE) WeekdaysAbbreviated() []string {
	return os.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (os *os_GE) WeekdayNarrow(weekday time.Weekday) string {
	return os.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (os *os_GE) WeekdaysNarrow() []string {
	return os.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (os *os_GE) WeekdayShort(weekday time.Weekday) string {
	return os.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (os *os_GE) WeekdaysShort() []string {
	return os.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (os *os_GE) WeekdayWide(weekday time.Weekday) string {
	return os.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (os *os_GE) WeekdaysWide() []string {
	return os.daysWide
}

// Decimal returns the decimal point of number
func (os *os_GE) Decimal() string {
	return os.decimal
}

// Group returns the group of number
func (os *os_GE) Group() string {
	return os.group
}

// Group returns the minus sign of number
func (os *os_GE) Minus() string {
	return os.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'os_GE' and handles both Whole and Real numbers based on 'v'
func (os *os_GE) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, os.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(os.group) - 1; j >= 0; j-- {
					b = append(b, os.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, os.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'os_GE' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (os *os_GE) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, os.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, os.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, os.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'os_GE'
func (os *os_GE) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := os.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, os.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(os.group) - 1; j >= 0; j-- {
					b = append(b, os.group[j])
				}
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

	for j := len(os.currencyPositivePrefix) - 1; j >= 0; j-- {
		b = append(b, os.currencyPositivePrefix[j])
	}

	if num < 0 {
		b = append(b, os.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, os.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'os_GE'
// in accounting notation.
func (os *os_GE) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := os.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, os.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(os.group) - 1; j >= 0; j-- {
					b = append(b, os.group[j])
				}
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

		for j := len(os.currencyNegativePrefix) - 1; j >= 0; j-- {
			b = append(b, os.currencyNegativePrefix[j])
		}

		b = append(b, os.minus[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(os.currencyPositivePrefix) - 1; j >= 0; j-- {
			b = append(b, os.currencyPositivePrefix[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, os.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'os_GE'
func (os *os_GE) FmtDateShort(t time.Time) string {

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

	if t.Year() > 9 {
		b = append(b, strconv.Itoa(t.Year())[2:]...)
	} else {
		b = append(b, strconv.Itoa(t.Year())[1:]...)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'os_GE'
func (os *os_GE) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, os.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20, 0xd0, 0xb0, 0xd0, 0xb7}...)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'os_GE'
func (os *os_GE) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, os.monthsWide[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20, 0xd0, 0xb0, 0xd0, 0xb7}...)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'os_GE'
func (os *os_GE) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, os.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, os.monthsWide[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20, 0xd0, 0xb0, 0xd0, 0xb7}...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'os_GE'
func (os *os_GE) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, os.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'os_GE'
func (os *os_GE) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, os.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, os.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'os_GE'
func (os *os_GE) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, os.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, os.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'os_GE'
func (os *os_GE) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, os.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, os.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := os.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
