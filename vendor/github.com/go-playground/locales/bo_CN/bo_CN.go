package bo_CN

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type bo_CN struct {
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

// New returns a new instance of translator for the 'bo_CN' locale
func New() locales.Translator {
	return &bo_CN{
		locale:                 "bo_CN",
		pluralsCardinal:        []locales.PluralRule{6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ".",
		group:                  ",",
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositivePrefix: " ",
		currencyNegativePrefix: " ",
		monthsAbbreviated:      []string{"", "ཟླ་༡", "ཟླ་༢", "ཟླ་༣", "ཟླ་༤", "ཟླ་༥", "ཟླ་༦", "ཟླ་༧", "ཟླ་༨", "ཟླ་༩", "ཟླ་༡༠", "ཟླ་༡༡", "ཟླ་༡༢"},
		monthsWide:             []string{"", "ཟླ་བ་དང་པོ", "ཟླ་བ་གཉིས་པ", "ཟླ་བ་གསུམ་པ", "ཟླ་བ་བཞི་པ", "ཟླ་བ་ལྔ་པ", "ཟླ་བ་དྲུག་པ", "ཟླ་བ་བདུན་པ", "ཟླ་བ་བརྒྱད་པ", "ཟླ་བ་དགུ་པ", "ཟླ་བ་བཅུ་པ", "ཟླ་བ་བཅུ་གཅིག་པ", "ཟླ་བ་བཅུ་གཉིས་པ"},
		daysAbbreviated:        []string{"ཉི་མ་", "ཟླ་བ་", "མིག་དམར་", "ལྷག་པ་", "ཕུར་བུ་", "པ་སངས་", "སྤེན་པ་"},
		daysNarrow:             []string{"ཉི", "ཟླ", "མིག", "ལྷག", "ཕུར", "སངས", "སྤེན"},
		daysWide:               []string{"གཟའ་ཉི་མ་", "གཟའ་ཟླ་བ་", "གཟའ་མིག་དམར་", "གཟའ་ལྷག་པ་", "གཟའ་ཕུར་བུ་", "གཟའ་པ་སངས་", "གཟའ་སྤེན་པ་"},
		periodsAbbreviated:     []string{"སྔ་དྲོ་", "ཕྱི་དྲོ་"},
		periodsWide:            []string{"སྔ་དྲོ་", "ཕྱི་དྲོ་"},
		erasAbbreviated:        []string{"སྤྱི་ལོ་སྔོན་", "སྤྱི་ལོ་"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"", ""},
		timezones:              map[string]string{"WESZ": "WESZ", "SRT": "SRT", "OEZ": "OEZ", "HECU": "HECU", "ACST": "ACST", "CAT": "CAT", "GMT": "GMT", "CHAST": "CHAST", "HNPMX": "HNPMX", "AEDT": "AEDT", "EST": "EST", "HKST": "HKST", "WART": "WART", "WITA": "WITA", "CLST": "CLST", "AST": "AST", "WEZ": "WEZ", "WIT": "WIT", "MDT": "MDT", "HNEG": "HNEG", "CDT": "CDT", "ECT": "ECT", "HEOG": "HEOG", "HNT": "HNT", "VET": "VET", "HNNOMX": "HNNOMX", "TMT": "TMT", "MST": "MST", "ADT": "ADT", "WAT": "WAT", "COST": "COST", "ACWST": "ACWST", "HEEG": "HEEG", "WARST": "WARST", "HEPM": "HEPM", "CLT": "CLT", "JST": "JST", "ACWDT": "ACWDT", "MESZ": "MESZ", "HAST": "HAST", "HADT": "HADT", "UYT": "UYT", "WAST": "WAST", "BT": "BT", "NZST": "NZST", "ARST": "ARST", "COT": "COT", "PDT": "PDT", "HNOG": "HNOG", "MEZ": "MEZ", "LHST": "LHST", "HENOMX": "HENOMX", "GYT": "GYT", "LHDT": "LHDT", "HNPM": "HNPM", "TMST": "TMST", "EAT": "EAT", "CHADT": "CHADT", "HNCU": "HNCU", "PST": "PST", "BOT": "BOT", "GFT": "GFT", "AKST": "AKST", "IST": "IST", "OESZ": "OESZ", "UYST": "UYST", "AEST": "AEST", "WIB": "WIB", "MYT": "MYT", "NZDT": "NZDT", "EDT": "EDT", "HAT": "HAT", "HEPMX": "HEPMX", "SAST": "SAST", "ACDT": "ACDT", "ART": "ART", "ChST": "ChST", "AWST": "AWST", "AWDT": "AWDT", "CST": "CST", "∅∅∅": "∅∅∅", "JDT": "JDT", "AKDT": "AKDT", "SGT": "SGT", "HKT": "HKT"},
	}
}

// Locale returns the current translators string locale
func (bo *bo_CN) Locale() string {
	return bo.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'bo_CN'
func (bo *bo_CN) PluralsCardinal() []locales.PluralRule {
	return bo.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'bo_CN'
func (bo *bo_CN) PluralsOrdinal() []locales.PluralRule {
	return bo.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'bo_CN'
func (bo *bo_CN) PluralsRange() []locales.PluralRule {
	return bo.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'bo_CN'
func (bo *bo_CN) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'bo_CN'
func (bo *bo_CN) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'bo_CN'
func (bo *bo_CN) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (bo *bo_CN) MonthAbbreviated(month time.Month) string {
	return bo.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (bo *bo_CN) MonthsAbbreviated() []string {
	return bo.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (bo *bo_CN) MonthNarrow(month time.Month) string {
	return bo.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (bo *bo_CN) MonthsNarrow() []string {
	return nil
}

// MonthWide returns the locales wide month given the 'month' provided
func (bo *bo_CN) MonthWide(month time.Month) string {
	return bo.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (bo *bo_CN) MonthsWide() []string {
	return bo.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (bo *bo_CN) WeekdayAbbreviated(weekday time.Weekday) string {
	return bo.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (bo *bo_CN) WeekdaysAbbreviated() []string {
	return bo.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (bo *bo_CN) WeekdayNarrow(weekday time.Weekday) string {
	return bo.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (bo *bo_CN) WeekdaysNarrow() []string {
	return bo.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (bo *bo_CN) WeekdayShort(weekday time.Weekday) string {
	return bo.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (bo *bo_CN) WeekdaysShort() []string {
	return bo.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (bo *bo_CN) WeekdayWide(weekday time.Weekday) string {
	return bo.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (bo *bo_CN) WeekdaysWide() []string {
	return bo.daysWide
}

// Decimal returns the decimal point of number
func (bo *bo_CN) Decimal() string {
	return bo.decimal
}

// Group returns the group of number
func (bo *bo_CN) Group() string {
	return bo.group
}

// Group returns the minus sign of number
func (bo *bo_CN) Minus() string {
	return bo.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'bo_CN' and handles both Whole and Real numbers based on 'v'
func (bo *bo_CN) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 1 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, bo.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, bo.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, bo.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'bo_CN' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (bo *bo_CN) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 1
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, bo.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, bo.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, bo.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'bo_CN'
func (bo *bo_CN) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := bo.currencies[currency]
	l := len(s) + len(symbol) + 3 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, bo.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, bo.group[0])
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

	for j := len(bo.currencyPositivePrefix) - 1; j >= 0; j-- {
		b = append(b, bo.currencyPositivePrefix[j])
	}

	if num < 0 {
		b = append(b, bo.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, bo.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'bo_CN'
// in accounting notation.
func (bo *bo_CN) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := bo.currencies[currency]
	l := len(s) + len(symbol) + 3 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, bo.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, bo.group[0])
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

		for j := len(bo.currencyNegativePrefix) - 1; j >= 0; j-- {
			b = append(b, bo.currencyNegativePrefix[j])
		}

		b = append(b, bo.minus[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(bo.currencyPositivePrefix) - 1; j >= 0; j-- {
			b = append(b, bo.currencyPositivePrefix[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, bo.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'bo_CN'
func (bo *bo_CN) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'bo_CN'
func (bo *bo_CN) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20, 0xe0, 0xbd, 0xa3, 0xe0, 0xbd, 0xbc, 0xe0, 0xbd, 0xa0, 0xe0, 0xbd, 0xb2, 0xe0, 0xbc, 0x8b}...)
	b = append(b, bo.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0xe0, 0xbd, 0x9a, 0xe0, 0xbd, 0xba, 0xe0, 0xbd, 0xa6, 0xe0, 0xbc, 0x8b}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'bo_CN'
func (bo *bo_CN) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, []byte{0xe0, 0xbd, 0xa6, 0xe0, 0xbe, 0xa4, 0xe0, 0xbe, 0xb1, 0xe0, 0xbd, 0xb2, 0xe0, 0xbc, 0x8b, 0xe0, 0xbd, 0xa3, 0xe0, 0xbd, 0xbc, 0xe0, 0xbc, 0x8b}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, bo.monthsWide[t.Month()]...)
	b = append(b, []byte{0xe0, 0xbd, 0xa0, 0xe0, 0xbd, 0xb2, 0xe0, 0xbc, 0x8b, 0xe0, 0xbd, 0x9a, 0xe0, 0xbd, 0xba, 0xe0, 0xbd, 0xa6, 0xe0, 0xbc, 0x8b}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'bo_CN'
func (bo *bo_CN) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, bo.monthsWide[t.Month()]...)
	b = append(b, []byte{0xe0, 0xbd, 0xa0, 0xe0, 0xbd, 0xb2, 0xe0, 0xbc, 0x8b, 0xe0, 0xbd, 0x9a, 0xe0, 0xbd, 0xba, 0xe0, 0xbd, 0xa6, 0xe0, 0xbc, 0x8b}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, bo.daysWide[t.Weekday()]...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'bo_CN'
func (bo *bo_CN) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, bo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, bo.periodsAbbreviated[0]...)
	} else {
		b = append(b, bo.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'bo_CN'
func (bo *bo_CN) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, bo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, bo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, bo.periodsAbbreviated[0]...)
	} else {
		b = append(b, bo.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'bo_CN'
func (bo *bo_CN) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, bo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, bo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, bo.periodsAbbreviated[0]...)
	} else {
		b = append(b, bo.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'bo_CN'
func (bo *bo_CN) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, bo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, bo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, bo.periodsAbbreviated[0]...)
	} else {
		b = append(b, bo.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := bo.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
