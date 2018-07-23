package ckb_IR

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type ckb_IR struct {
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

// New returns a new instance of translator for the 'ckb_IR' locale
func New() locales.Translator {
	return &ckb_IR{
		locale:                 "ckb_IR",
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
		timezones:              map[string]string{"HEPM": "HEPM", "HADT": "HADT", "PDT": "PDT", "EST": "EST", "GMT": "GMT", "HNCU": "HNCU", "BT": "BT", "EDT": "EDT", "WART": "WART", "MDT": "MDT", "OEZ": "OEZ", "WAT": "WAT", "NZDT": "NZDT", "LHDT": "LHDT", "OESZ": "OESZ", "HNPMX": "HNPMX", "HAT": "HAT", "CLT": "CLT", "NZST": "NZST", "HEEG": "HEEG", "HENOMX": "HENOMX", "WEZ": "WEZ", "WESZ": "WESZ", "BOT": "BOT", "AKDT": "AKDT", "ACWST": "ACWST", "MEZ": "MEZ", "LHST": "LHST", "TMST": "TMST", "CHADT": "CHADT", "ACST": "ACST", "MESZ": "MESZ", "VET": "VET", "WIT": "WIT", "AWST": "AWST", "AST": "AST", "WIB": "WIB", "HNOG": "HNOG", "WARST": "WARST", "EAT": "EAT", "CLST": "CLST", "UYT": "UYT", "HECU": "HECU", "WITA": "WITA", "HAST": "HAST", "HEPMX": "HEPMX", "HKT": "HKT", "ARST": "ARST", "AEST": "AEST", "AEDT": "AEDT", "MYT": "MYT", "SGT": "SGT", "CAT": "CAT", "COST": "COST", "PST": "PST", "AWDT": "AWDT", "WAST": "WAST", "JDT": "JDT", "HEOG": "HEOG", "HNT": "HNT", "HNNOMX": "HNNOMX", "ART": "ART", "CHAST": "CHAST", "JST": "JST", "ADT": "ADT", "ECT": "ECT", "ACDT": "ACDT", "HKST": "HKST", "IST": "IST", "HNPM": "HNPM", "COT": "COT", "UYST": "UYST", "ACWDT": "ACWDT", "∅∅∅": "∅∅∅", "TMT": "TMT", "ChST": "ChST", "GFT": "GFT", "AKST": "AKST", "SAST": "SAST", "HNEG": "HNEG", "SRT": "SRT", "MST": "MST", "GYT": "GYT", "CST": "CST", "CDT": "CDT"},
	}
}

// Locale returns the current translators string locale
func (ckb *ckb_IR) Locale() string {
	return ckb.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'ckb_IR'
func (ckb *ckb_IR) PluralsCardinal() []locales.PluralRule {
	return ckb.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'ckb_IR'
func (ckb *ckb_IR) PluralsOrdinal() []locales.PluralRule {
	return ckb.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'ckb_IR'
func (ckb *ckb_IR) PluralsRange() []locales.PluralRule {
	return ckb.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'ckb_IR'
func (ckb *ckb_IR) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'ckb_IR'
func (ckb *ckb_IR) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'ckb_IR'
func (ckb *ckb_IR) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (ckb *ckb_IR) MonthAbbreviated(month time.Month) string {
	return ckb.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (ckb *ckb_IR) MonthsAbbreviated() []string {
	return ckb.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (ckb *ckb_IR) MonthNarrow(month time.Month) string {
	return ckb.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (ckb *ckb_IR) MonthsNarrow() []string {
	return ckb.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (ckb *ckb_IR) MonthWide(month time.Month) string {
	return ckb.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (ckb *ckb_IR) MonthsWide() []string {
	return ckb.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (ckb *ckb_IR) WeekdayAbbreviated(weekday time.Weekday) string {
	return ckb.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (ckb *ckb_IR) WeekdaysAbbreviated() []string {
	return ckb.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (ckb *ckb_IR) WeekdayNarrow(weekday time.Weekday) string {
	return ckb.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (ckb *ckb_IR) WeekdaysNarrow() []string {
	return ckb.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (ckb *ckb_IR) WeekdayShort(weekday time.Weekday) string {
	return ckb.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (ckb *ckb_IR) WeekdaysShort() []string {
	return ckb.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (ckb *ckb_IR) WeekdayWide(weekday time.Weekday) string {
	return ckb.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (ckb *ckb_IR) WeekdaysWide() []string {
	return ckb.daysWide
}

// Decimal returns the decimal point of number
func (ckb *ckb_IR) Decimal() string {
	return ckb.decimal
}

// Group returns the group of number
func (ckb *ckb_IR) Group() string {
	return ckb.group
}

// Group returns the minus sign of number
func (ckb *ckb_IR) Minus() string {
	return ckb.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'ckb_IR' and handles both Whole and Real numbers based on 'v'
func (ckb *ckb_IR) FmtNumber(num float64, v uint64) string {

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

// FmtPercent returns 'num' with digits/precision of 'v' for 'ckb_IR' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (ckb *ckb_IR) FmtPercent(num float64, v uint64) string {
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

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'ckb_IR'
func (ckb *ckb_IR) FmtCurrency(num float64, v uint64, currency currency.Type) string {

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

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'ckb_IR'
// in accounting notation.
func (ckb *ckb_IR) FmtAccounting(num float64, v uint64, currency currency.Type) string {

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

// FmtDateShort returns the short date representation of 't' for 'ckb_IR'
func (ckb *ckb_IR) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'ckb_IR'
func (ckb *ckb_IR) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'ckb_IR'
func (ckb *ckb_IR) FmtDateLong(t time.Time) string {

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

// FmtDateFull returns the full date representation of 't' for 'ckb_IR'
func (ckb *ckb_IR) FmtDateFull(t time.Time) string {

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

// FmtTimeShort returns the short time representation of 't' for 'ckb_IR'
func (ckb *ckb_IR) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ckb.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'ckb_IR'
func (ckb *ckb_IR) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
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

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'ckb_IR'
func (ckb *ckb_IR) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
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

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'ckb_IR'
func (ckb *ckb_IR) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
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

	tz, _ := t.Zone()

	if btz, ok := ckb.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
