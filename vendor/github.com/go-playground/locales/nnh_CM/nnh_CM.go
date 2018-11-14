package nnh_CM

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type nnh_CM struct {
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

// New returns a new instance of translator for the 'nnh_CM' locale
func New() locales.Translator {
	return &nnh_CM{
		locale:                 "nnh_CM",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ",",
		group:                  ".",
		percent:                "%",
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositivePrefix: " ",
		currencyNegativePrefix: " ",
		monthsAbbreviated:      []string{"", "saŋ tsetsɛ̀ɛ lùm", "saŋ kàg ngwóŋ", "saŋ lepyè shúm", "saŋ cÿó", "saŋ tsɛ̀ɛ cÿó", "saŋ njÿoláʼ", "saŋ tyɛ̀b tyɛ̀b mbʉ̀ŋ", "saŋ mbʉ̀ŋ", "saŋ ngwɔ̀ʼ mbÿɛ", "saŋ tàŋa tsetsáʼ", "saŋ mejwoŋó", "saŋ lùm"},
		monthsWide:             []string{"", "saŋ tsetsɛ̀ɛ lùm", "saŋ kàg ngwóŋ", "saŋ lepyè shúm", "saŋ cÿó", "saŋ tsɛ̀ɛ cÿó", "saŋ njÿoláʼ", "saŋ tyɛ̀b tyɛ̀b mbʉ̀ŋ", "saŋ mbʉ̀ŋ", "saŋ ngwɔ̀ʼ mbÿɛ", "saŋ tàŋa tsetsáʼ", "saŋ mejwoŋó", "saŋ lùm"},
		daysAbbreviated:        []string{"lyɛʼɛ́ sẅíŋtè", "mvfò lyɛ̌ʼ", "mbɔ́ɔntè mvfò lyɛ̌ʼ", "tsètsɛ̀ɛ lyɛ̌ʼ", "mbɔ́ɔntè tsetsɛ̀ɛ lyɛ̌ʼ", "mvfò màga lyɛ̌ʼ", "màga lyɛ̌ʼ"},
		daysShort:              []string{"lyɛʼɛ́ sẅíŋtè", "mvfò lyɛ̌ʼ", "mbɔ́ɔntè mvfò lyɛ̌ʼ", "tsètsɛ̀ɛ lyɛ̌ʼ", "mbɔ́ɔntè tsetsɛ̀ɛ lyɛ̌ʼ", "mvfò màga lyɛ̌ʼ", "màga lyɛ̌ʼ"},
		daysWide:               []string{"lyɛʼɛ́ sẅíŋtè", "mvfò lyɛ̌ʼ", "mbɔ́ɔntè mvfò lyɛ̌ʼ", "tsètsɛ̀ɛ lyɛ̌ʼ", "mbɔ́ɔntè tsetsɛ̀ɛ lyɛ̌ʼ", "mvfò màga lyɛ̌ʼ", "màga lyɛ̌ʼ"},
		periodsAbbreviated:     []string{"mbaʼámbaʼ", "ncwònzém"},
		periodsWide:            []string{"mbaʼámbaʼ", "ncwònzém"},
		erasAbbreviated:        []string{"m.z.Y.", "m.g.n.Y."},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"mé zyé Yěsô", "mé gÿo ńzyé Yěsô"},
		timezones:              map[string]string{"CLST": "CLST", "CDT": "CDT", "ADT": "ADT", "JDT": "JDT", "HNNOMX": "HNNOMX", "EST": "EST", "CAT": "CAT", "CHAST": "CHAST", "CHADT": "CHADT", "AEDT": "AEDT", "ACWDT": "ACWDT", "SGT": "SGT", "HNEG": "HNEG", "HEOG": "HEOG", "CLT": "CLT", "COST": "COST", "AWST": "AWST", "WIB": "WIB", "SAST": "SAST", "SRT": "SRT", "ART": "ART", "ECT": "ECT", "HNPM": "HNPM", "ARST": "ARST", "HNPMX": "HNPMX", "PST": "PST", "MESZ": "MESZ", "VET": "VET", "NZST": "NZST", "LHST": "LHST", "MDT": "MDT", "WEZ": "WEZ", "GFT": "GFT", "HEEG": "HEEG", "IST": "IST", "UYST": "UYST", "HEPMX": "HEPMX", "HKST": "HKST", "WART": "WART", "HNT": "HNT", "∅∅∅": "∅∅∅", "OESZ": "OESZ", "UYT": "UYT", "EDT": "EDT", "ACDT": "ACDT", "ACST": "ACST", "EAT": "EAT", "HAST": "HAST", "HADT": "HADT", "AWDT": "AWDT", "AKST": "AKST", "HNOG": "HNOG", "HKT": "HKT", "WARST": "WARST", "WITA": "WITA", "OEZ": "OEZ", "GMT": "GMT", "CST": "CST", "MYT": "MYT", "AKDT": "AKDT", "HEPM": "HEPM", "HAT": "HAT", "HENOMX": "HENOMX", "WIT": "WIT", "TMT": "TMT", "AST": "AST", "WAT": "WAT", "BOT": "BOT", "NZDT": "NZDT", "LHDT": "LHDT", "ChST": "ChST", "HNCU": "HNCU", "PDT": "PDT", "WAST": "WAST", "JST": "JST", "ACWST": "ACWST", "MEZ": "MEZ", "TMST": "TMST", "COT": "COT", "GYT": "GYT", "HECU": "HECU", "BT": "BT", "MST": "MST", "AEST": "AEST", "WESZ": "WESZ"},
	}
}

// Locale returns the current translators string locale
func (nnh *nnh_CM) Locale() string {
	return nnh.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'nnh_CM'
func (nnh *nnh_CM) PluralsCardinal() []locales.PluralRule {
	return nnh.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'nnh_CM'
func (nnh *nnh_CM) PluralsOrdinal() []locales.PluralRule {
	return nnh.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'nnh_CM'
func (nnh *nnh_CM) PluralsRange() []locales.PluralRule {
	return nnh.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'nnh_CM'
func (nnh *nnh_CM) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'nnh_CM'
func (nnh *nnh_CM) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'nnh_CM'
func (nnh *nnh_CM) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (nnh *nnh_CM) MonthAbbreviated(month time.Month) string {
	return nnh.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (nnh *nnh_CM) MonthsAbbreviated() []string {
	return nnh.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (nnh *nnh_CM) MonthNarrow(month time.Month) string {
	return nnh.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (nnh *nnh_CM) MonthsNarrow() []string {
	return nil
}

// MonthWide returns the locales wide month given the 'month' provided
func (nnh *nnh_CM) MonthWide(month time.Month) string {
	return nnh.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (nnh *nnh_CM) MonthsWide() []string {
	return nnh.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (nnh *nnh_CM) WeekdayAbbreviated(weekday time.Weekday) string {
	return nnh.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (nnh *nnh_CM) WeekdaysAbbreviated() []string {
	return nnh.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (nnh *nnh_CM) WeekdayNarrow(weekday time.Weekday) string {
	return nnh.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (nnh *nnh_CM) WeekdaysNarrow() []string {
	return nnh.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (nnh *nnh_CM) WeekdayShort(weekday time.Weekday) string {
	return nnh.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (nnh *nnh_CM) WeekdaysShort() []string {
	return nnh.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (nnh *nnh_CM) WeekdayWide(weekday time.Weekday) string {
	return nnh.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (nnh *nnh_CM) WeekdaysWide() []string {
	return nnh.daysWide
}

// Decimal returns the decimal point of number
func (nnh *nnh_CM) Decimal() string {
	return nnh.decimal
}

// Group returns the group of number
func (nnh *nnh_CM) Group() string {
	return nnh.group
}

// Group returns the minus sign of number
func (nnh *nnh_CM) Minus() string {
	return nnh.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'nnh_CM' and handles both Whole and Real numbers based on 'v'
func (nnh *nnh_CM) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 1 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, nnh.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, nnh.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, nnh.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'nnh_CM' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (nnh *nnh_CM) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'nnh_CM'
func (nnh *nnh_CM) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := nnh.currencies[currency]
	l := len(s) + len(symbol) + 3 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, nnh.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, nnh.group[0])
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

	for j := len(nnh.currencyPositivePrefix) - 1; j >= 0; j-- {
		b = append(b, nnh.currencyPositivePrefix[j])
	}

	if num < 0 {
		b = append(b, nnh.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, nnh.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'nnh_CM'
// in accounting notation.
func (nnh *nnh_CM) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := nnh.currencies[currency]
	l := len(s) + len(symbol) + 3 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, nnh.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, nnh.group[0])
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

		for j := len(nnh.currencyNegativePrefix) - 1; j >= 0; j-- {
			b = append(b, nnh.currencyNegativePrefix[j])
		}

		b = append(b, nnh.minus[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(nnh.currencyPositivePrefix) - 1; j >= 0; j-- {
			b = append(b, nnh.currencyPositivePrefix[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, nnh.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'nnh_CM'
func (nnh *nnh_CM) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'nnh_CM'
func (nnh *nnh_CM) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, nnh.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'nnh_CM'
func (nnh *nnh_CM) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, []byte{0x6c, 0x79, 0xc9, 0x9b}...)
	b = append(b, []byte{0xcc, 0x8c, 0xca, 0xbc, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20, 0x6e, 0x61}...)
	b = append(b, []byte{0x20}...)
	b = append(b, nnh.monthsWide[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'nnh_CM'
func (nnh *nnh_CM) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, nnh.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20, 0x2c, 0x20, 0x6c, 0x79, 0xc9, 0x9b}...)
	b = append(b, []byte{0xcc, 0x8c, 0xca, 0xbc, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20, 0x6e, 0x61}...)
	b = append(b, []byte{0x20}...)
	b = append(b, nnh.monthsWide[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'nnh_CM'
func (nnh *nnh_CM) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'nnh_CM'
func (nnh *nnh_CM) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'nnh_CM'
func (nnh *nnh_CM) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'nnh_CM'
func (nnh *nnh_CM) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}
