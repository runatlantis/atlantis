package ckb_IQ

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type ckb_IQ struct {
	locale                 string
	pluralsCardinal        []locales.PluralRule
	pluralsOrdinal         []locales.PluralRule
	pluralsRange           []locales.PluralRule
	decimal                string
	group                  string
	minus                  string
	percent                string
	percentSuffix          string
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

// New returns a new instance of translator for the 'ckb_IQ' locale
func New() locales.Translator {
	return &ckb_IQ{
		locale:                 "ckb_IQ",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                "٫",
		group:                  "٬",
		minus:                  "‏-",
		percent:                "٪",
		perMille:               "؉",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		percentSuffix:          " ",
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "کانوونی دووەم", "شوبات", "ئازار", "نیسان", "ئایار", "حوزەیران", "تەمووز", "ئاب", "ئەیلوول", "تشرینی یەکەم", "تشرینی دووەم", "کانونی یەکەم"},
		monthsNarrow:           []string{"", "ک", "ش", "ئ", "ن", "ئ", "ح", "ت", "ئ", "ئ", "ت", "ت", "ک"},
		monthsWide:             []string{"", "کانوونی دووەم", "شوبات", "ئازار", "نیسان", "ئایار", "حوزەیران", "تەمووز", "ئاب", "ئەیلوول", "تشرینی یەکەم", "تشرینی دووەم", "کانونی یەکەم"},
		daysAbbreviated:        []string{"یەکشەممە", "دووشەممە", "سێشەممە", "چوارشەممە", "پێنجشەممە", "ھەینی", "شەممە"},
		daysNarrow:             []string{"ی", "د", "س", "چ", "پ", "ھ", "ش"},
		daysShort:              []string{"١ش", "٢ش", "٣ش", "٤ش", "٥ش", "ھ", "ش"},
		daysWide:               []string{"یەکشەممە", "دووشەممە", "سێشەممە", "چوارشەممە", "پێنجشەممە", "ھەینی", "شەممە"},
		periodsAbbreviated:     []string{"ب.ن", "د.ن"},
		periodsNarrow:          []string{"ب.ن", "د.ن"},
		periodsWide:            []string{"ب.ن", "د.ن"},
		erasAbbreviated:        []string{"پێش زایین", "زایینی"},
		erasNarrow:             []string{"پ.ن", "ز"},
		erasWide:               []string{"پێش زایین", "زایینی"},
		timezones:              map[string]string{"ARST": "ARST", "COST": "COST", "CHAST": "CHAST", "GFT": "GFT", "HEEG": "HEEG", "IST": "IST", "UYST": "UYST", "WAST": "WAST", "BT": "BT", "NZST": "NZST", "∅∅∅": "∅∅∅", "WARST": "WARST", "HNNOMX": "HNNOMX", "CLST": "CLST", "PST": "PST", "WAT": "WAT", "LHST": "LHST", "HEPM": "HEPM", "ChST": "ChST", "HEPMX": "HEPMX", "JST": "JST", "ACWST": "ACWST", "HKST": "HKST", "TMT": "TMT", "COT": "COT", "ADT": "ADT", "BOT": "BOT", "HNEG": "HNEG", "HNOG": "HNOG", "MEZ": "MEZ", "WART": "WART", "HNPMX": "HNPMX", "AST": "AST", "WESZ": "WESZ", "HNPM": "HNPM", "TMST": "TMST", "CST": "CST", "WEZ": "WEZ", "NZDT": "NZDT", "PDT": "PDT", "EST": "EST", "LHDT": "LHDT", "HENOMX": "HENOMX", "MST": "MST", "GMT": "GMT", "CHADT": "CHADT", "CDT": "CDT", "SAST": "SAST", "JDT": "JDT", "VET": "VET", "WITA": "WITA", "OEZ": "OEZ", "WIB": "WIB", "MYT": "MYT", "EDT": "EDT", "SRT": "SRT", "EAT": "EAT", "WIT": "WIT", "UYT": "UYT", "HEOG": "HEOG", "MESZ": "MESZ", "HKT": "HKT", "HECU": "HECU", "AKST": "AKST", "SGT": "SGT", "HNT": "HNT", "CAT": "CAT", "HAST": "HAST", "ART": "ART", "ECT": "ECT", "ACWDT": "ACWDT", "MDT": "MDT", "OESZ": "OESZ", "AWST": "AWST", "AWDT": "AWDT", "AEDT": "AEDT", "AKDT": "AKDT", "ACST": "ACST", "ACDT": "ACDT", "HAT": "HAT", "HADT": "HADT", "AEST": "AEST", "CLT": "CLT", "GYT": "GYT", "HNCU": "HNCU"},
	}
}

// Locale returns the current translators string locale
func (ckb *ckb_IQ) Locale() string {
	return ckb.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'ckb_IQ'
func (ckb *ckb_IQ) PluralsCardinal() []locales.PluralRule {
	return ckb.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'ckb_IQ'
func (ckb *ckb_IQ) PluralsOrdinal() []locales.PluralRule {
	return ckb.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'ckb_IQ'
func (ckb *ckb_IQ) PluralsRange() []locales.PluralRule {
	return ckb.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'ckb_IQ'
func (ckb *ckb_IQ) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'ckb_IQ'
func (ckb *ckb_IQ) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'ckb_IQ'
func (ckb *ckb_IQ) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (ckb *ckb_IQ) MonthAbbreviated(month time.Month) string {
	return ckb.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (ckb *ckb_IQ) MonthsAbbreviated() []string {
	return ckb.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (ckb *ckb_IQ) MonthNarrow(month time.Month) string {
	return ckb.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (ckb *ckb_IQ) MonthsNarrow() []string {
	return ckb.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (ckb *ckb_IQ) MonthWide(month time.Month) string {
	return ckb.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (ckb *ckb_IQ) MonthsWide() []string {
	return ckb.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (ckb *ckb_IQ) WeekdayAbbreviated(weekday time.Weekday) string {
	return ckb.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (ckb *ckb_IQ) WeekdaysAbbreviated() []string {
	return ckb.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (ckb *ckb_IQ) WeekdayNarrow(weekday time.Weekday) string {
	return ckb.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (ckb *ckb_IQ) WeekdaysNarrow() []string {
	return ckb.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (ckb *ckb_IQ) WeekdayShort(weekday time.Weekday) string {
	return ckb.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (ckb *ckb_IQ) WeekdaysShort() []string {
	return ckb.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (ckb *ckb_IQ) WeekdayWide(weekday time.Weekday) string {
	return ckb.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (ckb *ckb_IQ) WeekdaysWide() []string {
	return ckb.daysWide
}

// Decimal returns the decimal point of number
func (ckb *ckb_IQ) Decimal() string {
	return ckb.decimal
}

// Group returns the group of number
func (ckb *ckb_IQ) Group() string {
	return ckb.group
}

// Group returns the minus sign of number
func (ckb *ckb_IQ) Minus() string {
	return ckb.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'ckb_IQ' and handles both Whole and Real numbers based on 'v'
func (ckb *ckb_IQ) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 6 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			for j := len(ckb.decimal) - 1; j >= 0; j-- {
				b = append(b, ckb.decimal[j])
			}
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(ckb.group) - 1; j >= 0; j-- {
					b = append(b, ckb.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(ckb.minus) - 1; j >= 0; j-- {
			b = append(b, ckb.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'ckb_IQ' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (ckb *ckb_IQ) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 10
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			for j := len(ckb.decimal) - 1; j >= 0; j-- {
				b = append(b, ckb.decimal[j])
			}
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(ckb.minus) - 1; j >= 0; j-- {
			b = append(b, ckb.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, ckb.percentSuffix...)

	b = append(b, ckb.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'ckb_IQ'
func (ckb *ckb_IQ) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ckb.currencies[currency]
	l := len(s) + len(symbol) + 8 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			for j := len(ckb.decimal) - 1; j >= 0; j-- {
				b = append(b, ckb.decimal[j])
			}
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(ckb.group) - 1; j >= 0; j-- {
					b = append(b, ckb.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(ckb.minus) - 1; j >= 0; j-- {
			b = append(b, ckb.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ckb.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, ckb.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'ckb_IQ'
// in accounting notation.
func (ckb *ckb_IQ) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ckb.currencies[currency]
	l := len(s) + len(symbol) + 8 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			for j := len(ckb.decimal) - 1; j >= 0; j-- {
				b = append(b, ckb.decimal[j])
			}
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(ckb.group) - 1; j >= 0; j-- {
					b = append(b, ckb.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		for j := len(ckb.minus) - 1; j >= 0; j-- {
			b = append(b, ckb.minus[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ckb.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, ckb.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, ckb.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'ckb_IQ'
func (ckb *ckb_IQ) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'ckb_IQ'
func (ckb *ckb_IQ) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, ckb.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'ckb_IQ'
func (ckb *ckb_IQ) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0xdb, 0x8c, 0x20}...)
	b = append(b, ckb.monthsWide[t.Month()]...)
	b = append(b, []byte{0xdb, 0x8c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'ckb_IQ'
func (ckb *ckb_IQ) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, ckb.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, ckb.daysWide[t.Weekday()]...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'ckb_IQ'
func (ckb *ckb_IQ) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ckb.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ckb.periodsAbbreviated[0]...)
	} else {
		b = append(b, ckb.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'ckb_IQ'
func (ckb *ckb_IQ) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ckb.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ckb.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ckb.periodsAbbreviated[0]...)
	} else {
		b = append(b, ckb.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'ckb_IQ'
func (ckb *ckb_IQ) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ckb.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ckb.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ckb.periodsAbbreviated[0]...)
	} else {
		b = append(b, ckb.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'ckb_IQ'
func (ckb *ckb_IQ) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ckb.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ckb.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ckb.periodsAbbreviated[0]...)
	} else {
		b = append(b, ckb.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := ckb.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
