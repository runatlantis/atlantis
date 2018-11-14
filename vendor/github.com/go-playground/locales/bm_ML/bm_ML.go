package bm_ML

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type bm_ML struct {
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
	currencyNegativePrefix string
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

// New returns a new instance of translator for the 'bm_ML' locale
func New() locales.Translator {
	return &bm_ML{
		locale:                 "bm_ML",
		pluralsCardinal:        []locales.PluralRule{6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: ")",
		monthsAbbreviated:      []string{"", "zan", "feb", "mar", "awi", "mɛ", "zuw", "zul", "uti", "sɛt", "ɔku", "now", "des"},
		monthsNarrow:           []string{"", "Z", "F", "M", "A", "M", "Z", "Z", "U", "S", "Ɔ", "N", "D"},
		monthsWide:             []string{"", "zanwuye", "feburuye", "marisi", "awirili", "mɛ", "zuwɛn", "zuluye", "uti", "sɛtanburu", "ɔkutɔburu", "nowanburu", "desanburu"},
		daysAbbreviated:        []string{"kar", "ntɛ", "tar", "ara", "ala", "jum", "sib"},
		daysNarrow:             []string{"K", "N", "T", "A", "A", "J", "S"},
		daysWide:               []string{"kari", "ntɛnɛ", "tarata", "araba", "alamisa", "juma", "sibiri"},
		erasAbbreviated:        []string{"J.-C. ɲɛ", "ni J.-C."},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"jezu krisiti ɲɛ", "jezu krisiti minkɛ"},
		timezones:              map[string]string{"CHAST": "CHAST", "WAST": "WAST", "JDT": "JDT", "SGT": "SGT", "HKT": "HKT", "COT": "COT", "GYT": "GYT", "AKDT": "AKDT", "WART": "WART", "AWDT": "AWDT", "AST": "AST", "WESZ": "WESZ", "UYST": "UYST", "HEPMX": "HEPMX", "MYT": "MYT", "NZST": "NZST", "ACWST": "ACWST", "COST": "COST", "HAST": "HAST", "WIT": "WIT", "AWST": "AWST", "CDT": "CDT", "BOT": "BOT", "HEEG": "HEEG", "EST": "EST", "EDT": "EDT", "MST": "MST", "GFT": "GFT", "ChST": "ChST", "BT": "BT", "ECT": "ECT", "HNPMX": "HNPMX", "AEST": "AEST", "NZDT": "NZDT", "AKST": "AKST", "HKST": "HKST", "CAT": "CAT", "SAST": "SAST", "CLT": "CLT", "HNCU": "HNCU", "ADT": "ADT", "HNOG": "HNOG", "WARST": "WARST", "GMT": "GMT", "CHADT": "CHADT", "HECU": "HECU", "∅∅∅": "∅∅∅", "UYT": "UYT", "CST": "CST", "PST": "PST", "AEDT": "AEDT", "MEZ": "MEZ", "SRT": "SRT", "ARST": "ARST", "HADT": "HADT", "LHST": "LHST", "MDT": "MDT", "EAT": "EAT", "OEZ": "OEZ", "WIB": "WIB", "TMST": "TMST", "ART": "ART", "OESZ": "OESZ", "JST": "JST", "ACWDT": "ACWDT", "MESZ": "MESZ", "HNNOMX": "HNNOMX", "HENOMX": "HENOMX", "WEZ": "WEZ", "WAT": "WAT", "LHDT": "LHDT", "HNPM": "HNPM", "HNT": "HNT", "HAT": "HAT", "CLST": "CLST", "TMT": "TMT", "PDT": "PDT", "ACST": "ACST", "IST": "IST", "HEPM": "HEPM", "WITA": "WITA", "VET": "VET", "HNEG": "HNEG", "HEOG": "HEOG", "ACDT": "ACDT"},
	}
}

// Locale returns the current translators string locale
func (bm *bm_ML) Locale() string {
	return bm.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'bm_ML'
func (bm *bm_ML) PluralsCardinal() []locales.PluralRule {
	return bm.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'bm_ML'
func (bm *bm_ML) PluralsOrdinal() []locales.PluralRule {
	return bm.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'bm_ML'
func (bm *bm_ML) PluralsRange() []locales.PluralRule {
	return bm.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'bm_ML'
func (bm *bm_ML) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'bm_ML'
func (bm *bm_ML) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'bm_ML'
func (bm *bm_ML) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (bm *bm_ML) MonthAbbreviated(month time.Month) string {
	return bm.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (bm *bm_ML) MonthsAbbreviated() []string {
	return bm.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (bm *bm_ML) MonthNarrow(month time.Month) string {
	return bm.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (bm *bm_ML) MonthsNarrow() []string {
	return bm.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (bm *bm_ML) MonthWide(month time.Month) string {
	return bm.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (bm *bm_ML) MonthsWide() []string {
	return bm.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (bm *bm_ML) WeekdayAbbreviated(weekday time.Weekday) string {
	return bm.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (bm *bm_ML) WeekdaysAbbreviated() []string {
	return bm.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (bm *bm_ML) WeekdayNarrow(weekday time.Weekday) string {
	return bm.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (bm *bm_ML) WeekdaysNarrow() []string {
	return bm.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (bm *bm_ML) WeekdayShort(weekday time.Weekday) string {
	return bm.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (bm *bm_ML) WeekdaysShort() []string {
	return bm.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (bm *bm_ML) WeekdayWide(weekday time.Weekday) string {
	return bm.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (bm *bm_ML) WeekdaysWide() []string {
	return bm.daysWide
}

// Decimal returns the decimal point of number
func (bm *bm_ML) Decimal() string {
	return bm.decimal
}

// Group returns the group of number
func (bm *bm_ML) Group() string {
	return bm.group
}

// Group returns the minus sign of number
func (bm *bm_ML) Minus() string {
	return bm.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'bm_ML' and handles both Whole and Real numbers based on 'v'
func (bm *bm_ML) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'bm_ML' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (bm *bm_ML) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'bm_ML'
func (bm *bm_ML) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := bm.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, bm.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, bm.group[0])
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

	if num < 0 {
		b = append(b, bm.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, bm.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'bm_ML'
// in accounting notation.
func (bm *bm_ML) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := bm.currencies[currency]
	l := len(s) + len(symbol) + 2
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, bm.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, bm.group[0])
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

		b = append(b, bm.currencyNegativePrefix[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, bm.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, bm.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'bm_ML'
func (bm *bm_ML) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2f}...)
	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0x2f}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'bm_ML'
func (bm *bm_ML) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, bm.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'bm_ML'
func (bm *bm_ML) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, bm.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'bm_ML'
func (bm *bm_ML) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, bm.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, bm.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'bm_ML'
func (bm *bm_ML) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, bm.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'bm_ML'
func (bm *bm_ML) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, bm.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, bm.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'bm_ML'
func (bm *bm_ML) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, bm.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, bm.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'bm_ML'
func (bm *bm_ML) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, bm.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, bm.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := bm.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
