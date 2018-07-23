package yi

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type yi struct {
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
	currencyPositiveSuffix string
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

// New returns a new instance of translator for the 'yi' locale
func New() locales.Translator {
	return &yi{
		locale:                 "yi",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ".",
		group:                  ",",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JP¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositivePrefix: " ",
		currencyPositiveSuffix: "K",
		currencyNegativePrefix: " ",
		currencyNegativeSuffix: "K",
		monthsAbbreviated:      []string{"", "יאַנואַר", "פֿעברואַר", "מערץ", "אַפּריל", "מיי", "יוני", "יולי", "אויגוסט", "סעפּטעמבער", "אקטאבער", "נאוועמבער", "דעצעמבער"},
		monthsWide:             []string{"", "יאַנואַר", "פֿעברואַר", "מערץ", "אַפּריל", "מיי", "יוני", "יולי", "אויגוסט", "סעפּטעמבער", "אקטאבער", "נאוועמבער", "דעצעמבער"},
		daysAbbreviated:        []string{"זונטיק", "מאָנטיק", "דינסטיק", "מיטוואך", "דאנערשטיק", "פֿרײַטיק", "שבת"},
		daysShort:              []string{"זונטיק", "מאָנטיק", "דינסטיק", "מיטוואך", "דאנערשטיק", "פֿרײַטיק", "שבת"},
		daysWide:               []string{"זונטיק", "מאָנטיק", "דינסטיק", "מיטוואך", "דאנערשטיק", "פֿרײַטיק", "שבת"},
		periodsAbbreviated:     []string{"פֿאַרמיטאָג", "נאָכמיטאָג"},
		periodsWide:            []string{"פֿאַרמיטאָג", "נאָכמיטאָג"},
		timezones:              map[string]string{"GYT": "GYT", "AKST": "AKST", "SGT": "SGT", "MEZ": "MEZ", "HKT": "HKT", "EAT": "EAT", "UYST": "UYST", "CHADT": "CHADT", "CDT": "CDT", "MDT": "MDT", "BOT": "BOT", "NZDT": "NZDT", "HNOG": "HNOG", "HNNOMX": "HNNOMX", "HECU": "HECU", "AEST": "AEST", "JDT": "JDT", "MYT": "MYT", "WARST": "WARST", "COST": "COST", "LHDT": "LHDT", "WITA": "WITA", "ARST": "ARST", "HNCU": "HNCU", "HNPMX": "HNPMX", "CST": "CST", "HADT": "HADT", "ACWST": "ACWST", "HENOMX": "HENOMX", "SRT": "SRT", "WIT": "WIT", "COT": "COT", "ADT": "ADT", "LHST": "LHST", "CHAST": "CHAST", "PST": "PST", "AWDT": "AWDT", "AST": "AST", "MST": "MST", "HEEG": "HEEG", "CLST": "CLST", "OESZ": "OESZ", "WESZ": "WESZ", "ACDT": "ACDT", "HNPM": "HNPM", "CLT": "CLT", "OEZ": "OEZ", "GMT": "GMT", "PDT": "PDT", "WAST": "WAST", "WEZ": "WEZ", "NZST": "NZST", "EST": "EST", "TMT": "TMT", "ChST": "ChST", "∅∅∅": "∅∅∅", "WART": "WART", "VET": "VET", "TMST": "TMST", "CAT": "CAT", "AKDT": "AKDT", "HEOG": "HEOG", "ACST": "ACST", "ART": "ART", "HEPM": "HEPM", "SAST": "SAST", "WAT": "WAT", "JST": "JST", "ECT": "ECT", "MESZ": "MESZ", "WIB": "WIB", "EDT": "EDT", "HAT": "HAT", "UYT": "UYT", "GFT": "GFT", "HNT": "HNT", "HAST": "HAST", "HEPMX": "HEPMX", "AWST": "AWST", "IST": "IST", "AEDT": "AEDT", "BT": "BT", "HNEG": "HNEG", "ACWDT": "ACWDT", "HKST": "HKST"},
	}
}

// Locale returns the current translators string locale
func (yi *yi) Locale() string {
	return yi.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'yi'
func (yi *yi) PluralsCardinal() []locales.PluralRule {
	return yi.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'yi'
func (yi *yi) PluralsOrdinal() []locales.PluralRule {
	return yi.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'yi'
func (yi *yi) PluralsRange() []locales.PluralRule {
	return yi.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'yi'
func (yi *yi) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)

	if i == 1 && v == 0 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'yi'
func (yi *yi) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'yi'
func (yi *yi) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (yi *yi) MonthAbbreviated(month time.Month) string {
	return yi.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (yi *yi) MonthsAbbreviated() []string {
	return yi.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (yi *yi) MonthNarrow(month time.Month) string {
	return yi.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (yi *yi) MonthsNarrow() []string {
	return nil
}

// MonthWide returns the locales wide month given the 'month' provided
func (yi *yi) MonthWide(month time.Month) string {
	return yi.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (yi *yi) MonthsWide() []string {
	return yi.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (yi *yi) WeekdayAbbreviated(weekday time.Weekday) string {
	return yi.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (yi *yi) WeekdaysAbbreviated() []string {
	return yi.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (yi *yi) WeekdayNarrow(weekday time.Weekday) string {
	return yi.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (yi *yi) WeekdaysNarrow() []string {
	return yi.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (yi *yi) WeekdayShort(weekday time.Weekday) string {
	return yi.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (yi *yi) WeekdaysShort() []string {
	return yi.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (yi *yi) WeekdayWide(weekday time.Weekday) string {
	return yi.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (yi *yi) WeekdaysWide() []string {
	return yi.daysWide
}

// Decimal returns the decimal point of number
func (yi *yi) Decimal() string {
	return yi.decimal
}

// Group returns the group of number
func (yi *yi) Group() string {
	return yi.group
}

// Group returns the minus sign of number
func (yi *yi) Minus() string {
	return yi.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'yi' and handles both Whole and Real numbers based on 'v'
func (yi *yi) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'yi' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (yi *yi) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'yi'
func (yi *yi) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := yi.currencies[currency]
	l := len(s) + len(symbol) + 5

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, yi.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	for j := len(symbol) - 1; j >= 0; j-- {
		b = append(b, symbol[j])
	}

	for j := len(yi.currencyPositivePrefix) - 1; j >= 0; j-- {
		b = append(b, yi.currencyPositivePrefix[j])
	}

	if num < 0 {
		b = append(b, yi.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, yi.currencyPositiveSuffix...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'yi'
// in accounting notation.
func (yi *yi) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := yi.currencies[currency]
	l := len(s) + len(symbol) + 5

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, yi.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(yi.currencyNegativePrefix) - 1; j >= 0; j-- {
			b = append(b, yi.currencyNegativePrefix[j])
		}

		b = append(b, yi.minus[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(yi.currencyPositivePrefix) - 1; j >= 0; j-- {
			b = append(b, yi.currencyPositivePrefix[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if num < 0 {
		b = append(b, yi.currencyNegativeSuffix...)
	} else {

		b = append(b, yi.currencyPositiveSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'yi'
func (yi *yi) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'yi'
func (yi *yi) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0xd7, 0x98, 0xd7, 0x9f, 0x20}...)
	b = append(b, yi.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'yi'
func (yi *yi) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0xd7, 0x98, 0xd7, 0x9f, 0x20}...)
	b = append(b, yi.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'yi'
func (yi *yi) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, yi.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0xd7, 0x98, 0xd7, 0x9f, 0x20}...)
	b = append(b, yi.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'yi'
func (yi *yi) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, yi.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'yi'
func (yi *yi) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, yi.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, yi.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'yi'
func (yi *yi) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, yi.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, yi.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'yi'
func (yi *yi) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, yi.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, yi.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := yi.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
