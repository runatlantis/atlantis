package fur_IT

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type fur_IT struct {
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

// New returns a new instance of translator for the 'fur_IT' locale
func New() locales.Translator {
	return &fur_IT{
		locale:                 "fur_IT",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ",",
		group:                  ".",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositivePrefix: " ",
		currencyNegativePrefix: " ",
		monthsAbbreviated:      []string{"", "Zen", "Fev", "Mar", "Avr", "Mai", "Jug", "Lui", "Avo", "Set", "Otu", "Nov", "Dic"},
		monthsNarrow:           []string{"", "Z", "F", "M", "A", "M", "J", "L", "A", "S", "O", "N", "D"},
		monthsWide:             []string{"", "Zenâr", "Fevrâr", "Març", "Avrîl", "Mai", "Jugn", "Lui", "Avost", "Setembar", "Otubar", "Novembar", "Dicembar"},
		daysAbbreviated:        []string{"dom", "lun", "mar", "mie", "joi", "vin", "sab"},
		daysNarrow:             []string{"D", "L", "M", "M", "J", "V", "S"},
		daysWide:               []string{"domenie", "lunis", "martars", "miercus", "joibe", "vinars", "sabide"},
		periodsAbbreviated:     []string{"a.", "p."},
		periodsWide:            []string{"a.", "p."},
		erasAbbreviated:        []string{"pdC", "ddC"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"", ""},
		timezones:              map[string]string{"HNOG": "HNOG", "HEPM": "HEPM", "CLT": "CLT", "JST": "JST", "GMT": "GMT", "HECU": "HECU", "CST": "CST", "AST": "AST", "EDT": "EDT", "HNEG": "HNEG", "WITA": "WITA", "OEZ": "Ore standard de Europe orientâl", "AKST": "AKST", "HEPMX": "HEPMX", "ADT": "ADT", "WAT": "WAT", "EST": "EST", "HEEG": "HEEG", "HNPM": "HNPM", "CHAST": "CHAST", "TMT": "TMT", "ARST": "ARST", "PDT": "PDT", "BT": "BT", "HEOG": "HEOG", "IST": "IST", "LHST": "LHST", "HNNOMX": "HNNOMX", "NZDT": "NZDT", "UYST": "UYST", "CHADT": "CHADT", "HADT": "HADT", "AEDT": "AEDT", "NZST": "NZST", "BOT": "BOT", "HKT": "HKT", "HNT": "HNT", "TMST": "TMST", "CLST": "CLST", "HENOMX": "HENOMX", "SGT": "SGT", "ACDT": "ACDT", "ACWDT": "ACWDT", "MEZ": "Ore standard de Europe centrâl", "VET": "VET", "OESZ": "Ore estive de Europe orientâl", "MYT": "MYT", "GFT": "GFT", "ECT": "ECT", "MST": "MST", "MDT": "MDT", "CDT": "CDT", "WIB": "WIB", "∅∅∅": "∅∅∅", "WARST": "WARST", "PST": "PST", "AKDT": "AKDT", "WART": "WART", "WIT": "WIT", "HAST": "HAST", "AWST": "AWST", "AWDT": "AWDT", "ACWST": "ACWST", "LHDT": "LHDT", "CAT": "CAT", "EAT": "EAT", "HNPMX": "HNPMX", "GYT": "GYT", "ChST": "ChST", "JDT": "JDT", "ART": "ART", "COT": "COT", "HNCU": "HNCU", "AEST": "AEST", "COST": "COST", "SAST": "SAST", "WAST": "WAST", "WESZ": "Ore estive de Europe ocidentâl", "ACST": "ACST", "HKST": "HKST", "HAT": "HAT", "SRT": "SRT", "MESZ": "Ore estive de Europe centrâl", "UYT": "UYT", "WEZ": "Ore standard de Europe ocidentâl"},
	}
}

// Locale returns the current translators string locale
func (fur *fur_IT) Locale() string {
	return fur.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'fur_IT'
func (fur *fur_IT) PluralsCardinal() []locales.PluralRule {
	return fur.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'fur_IT'
func (fur *fur_IT) PluralsOrdinal() []locales.PluralRule {
	return fur.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'fur_IT'
func (fur *fur_IT) PluralsRange() []locales.PluralRule {
	return fur.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'fur_IT'
func (fur *fur_IT) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'fur_IT'
func (fur *fur_IT) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'fur_IT'
func (fur *fur_IT) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (fur *fur_IT) MonthAbbreviated(month time.Month) string {
	return fur.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (fur *fur_IT) MonthsAbbreviated() []string {
	return fur.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (fur *fur_IT) MonthNarrow(month time.Month) string {
	return fur.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (fur *fur_IT) MonthsNarrow() []string {
	return fur.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (fur *fur_IT) MonthWide(month time.Month) string {
	return fur.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (fur *fur_IT) MonthsWide() []string {
	return fur.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (fur *fur_IT) WeekdayAbbreviated(weekday time.Weekday) string {
	return fur.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (fur *fur_IT) WeekdaysAbbreviated() []string {
	return fur.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (fur *fur_IT) WeekdayNarrow(weekday time.Weekday) string {
	return fur.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (fur *fur_IT) WeekdaysNarrow() []string {
	return fur.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (fur *fur_IT) WeekdayShort(weekday time.Weekday) string {
	return fur.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (fur *fur_IT) WeekdaysShort() []string {
	return fur.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (fur *fur_IT) WeekdayWide(weekday time.Weekday) string {
	return fur.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (fur *fur_IT) WeekdaysWide() []string {
	return fur.daysWide
}

// Decimal returns the decimal point of number
func (fur *fur_IT) Decimal() string {
	return fur.decimal
}

// Group returns the group of number
func (fur *fur_IT) Group() string {
	return fur.group
}

// Group returns the minus sign of number
func (fur *fur_IT) Minus() string {
	return fur.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'fur_IT' and handles both Whole and Real numbers based on 'v'
func (fur *fur_IT) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, fur.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, fur.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, fur.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'fur_IT' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (fur *fur_IT) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, fur.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, fur.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, fur.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'fur_IT'
func (fur *fur_IT) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := fur.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, fur.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, fur.group[0])
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

	for j := len(fur.currencyPositivePrefix) - 1; j >= 0; j-- {
		b = append(b, fur.currencyPositivePrefix[j])
	}

	if num < 0 {
		b = append(b, fur.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, fur.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'fur_IT'
// in accounting notation.
func (fur *fur_IT) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := fur.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, fur.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, fur.group[0])
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

		for j := len(fur.currencyNegativePrefix) - 1; j >= 0; j-- {
			b = append(b, fur.currencyNegativePrefix[j])
		}

		b = append(b, fur.minus[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(fur.currencyPositivePrefix) - 1; j >= 0; j-- {
			b = append(b, fur.currencyPositivePrefix[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, fur.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'fur_IT'
func (fur *fur_IT) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'fur_IT'
func (fur *fur_IT) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'fur_IT'
func (fur *fur_IT) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20, 0x64, 0x69}...)
	b = append(b, []byte{0x20}...)
	b = append(b, fur.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20, 0x64, 0x61, 0x6c}...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'fur_IT'
func (fur *fur_IT) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, fur.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20, 0x64, 0x69}...)
	b = append(b, []byte{0x20}...)
	b = append(b, fur.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20, 0x64, 0x61, 0x6c}...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'fur_IT'
func (fur *fur_IT) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, fur.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'fur_IT'
func (fur *fur_IT) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, fur.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, fur.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'fur_IT'
func (fur *fur_IT) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, fur.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, fur.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'fur_IT'
func (fur *fur_IT) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, fur.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, fur.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := fur.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
