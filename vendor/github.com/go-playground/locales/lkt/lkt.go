package lkt

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type lkt struct {
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

// New returns a new instance of translator for the 'lkt' locale
func New() locales.Translator {
	return &lkt{
		locale:                 "lkt",
		pluralsCardinal:        []locales.PluralRule{6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ".",
		group:                  ",",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositivePrefix: " ",
		currencyPositiveSuffix: "K",
		currencyNegativePrefix: " ",
		currencyNegativeSuffix: "K",
		monthsWide:             []string{"", "Wiótheȟika Wí", "Thiyóȟeyuŋka Wí", "Ištáwičhayazaŋ Wí", "Pȟežítȟo Wí", "Čhaŋwápetȟo Wí", "Wípazukȟa-wašté Wí", "Čhaŋpȟásapa Wí", "Wasútȟuŋ Wí", "Čhaŋwápeǧi Wí", "Čhaŋwápe-kasná Wí", "Waníyetu Wí", "Tȟahékapšuŋ Wí"},
		daysNarrow:             []string{"A", "W", "N", "Y", "T", "Z", "O"},
		daysWide:               []string{"Aŋpétuwakȟaŋ", "Aŋpétuwaŋži", "Aŋpétunuŋpa", "Aŋpétuyamni", "Aŋpétutopa", "Aŋpétuzaptaŋ", "Owáŋgyužažapi"},
		timezones:              map[string]string{"AKST": "AKST", "ACWDT": "ACWDT", "MEZ": "MEZ", "HECU": "HECU", "CST": "CST", "BT": "BT", "BOT": "BOT", "MESZ": "MESZ", "OESZ": "OESZ", "HNPMX": "HNPMX", "GFT": "GFT", "HEPM": "HEPM", "CLT": "CLT", "GYT": "GYT", "TMT": "TMT", "TMST": "TMST", "ARST": "ARST", "GMT": "GMT", "WAST": "WAST", "EST": "EST", "HNPM": "HNPM", "HNNOMX": "HNNOMX", "CHAST": "CHAST", "CHADT": "CHADT", "AWDT": "AWDT", "AST": "AST", "NZDT": "NZDT", "EDT": "EDT", "UYT": "UYT", "PST": "PST", "WART": "WART", "VET": "VET", "AEST": "AEST", "MYT": "MYT", "ECT": "ECT", "HEEG": "HEEG", "LHDT": "LHDT", "WARST": "WARST", "SRT": "SRT", "OEZ": "OEZ", "COT": "COT", "CDT": "CDT", "JST": "JST", "ACST": "ACST", "HNEG": "HNEG", "ChST": "ChST", "SGT": "SGT", "MST": "MST", "HNCU": "HNCU", "AKDT": "AKDT", "AEDT": "AEDT", "HAST": "HAST", "∅∅∅": "∅∅∅", "JDT": "JDT", "ACDT": "ACDT", "HNT": "HNT", "HAT": "HAT", "HENOMX": "HENOMX", "CAT": "CAT", "UYST": "UYST", "SAST": "SAST", "WESZ": "WESZ", "HKT": "HKT", "LHST": "LHST", "ADT": "ADT", "IST": "IST", "WEZ": "WEZ", "WIB": "WIB", "NZST": "NZST", "HEOG": "HEOG", "COST": "COST", "AWST": "AWST", "HEPMX": "HEPMX", "HNOG": "HNOG", "WITA": "WITA", "WIT": "WIT", "HADT": "HADT", "EAT": "EAT", "CLST": "CLST", "PDT": "PDT", "ART": "ART", "WAT": "WAT", "ACWST": "ACWST", "HKST": "HKST", "MDT": "MDT"},
	}
}

// Locale returns the current translators string locale
func (lkt *lkt) Locale() string {
	return lkt.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'lkt'
func (lkt *lkt) PluralsCardinal() []locales.PluralRule {
	return lkt.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'lkt'
func (lkt *lkt) PluralsOrdinal() []locales.PluralRule {
	return lkt.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'lkt'
func (lkt *lkt) PluralsRange() []locales.PluralRule {
	return lkt.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'lkt'
func (lkt *lkt) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'lkt'
func (lkt *lkt) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'lkt'
func (lkt *lkt) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (lkt *lkt) MonthAbbreviated(month time.Month) string {
	return lkt.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (lkt *lkt) MonthsAbbreviated() []string {
	return nil
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (lkt *lkt) MonthNarrow(month time.Month) string {
	return lkt.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (lkt *lkt) MonthsNarrow() []string {
	return nil
}

// MonthWide returns the locales wide month given the 'month' provided
func (lkt *lkt) MonthWide(month time.Month) string {
	return lkt.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (lkt *lkt) MonthsWide() []string {
	return lkt.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (lkt *lkt) WeekdayAbbreviated(weekday time.Weekday) string {
	return lkt.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (lkt *lkt) WeekdaysAbbreviated() []string {
	return lkt.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (lkt *lkt) WeekdayNarrow(weekday time.Weekday) string {
	return lkt.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (lkt *lkt) WeekdaysNarrow() []string {
	return lkt.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (lkt *lkt) WeekdayShort(weekday time.Weekday) string {
	return lkt.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (lkt *lkt) WeekdaysShort() []string {
	return lkt.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (lkt *lkt) WeekdayWide(weekday time.Weekday) string {
	return lkt.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (lkt *lkt) WeekdaysWide() []string {
	return lkt.daysWide
}

// Decimal returns the decimal point of number
func (lkt *lkt) Decimal() string {
	return lkt.decimal
}

// Group returns the group of number
func (lkt *lkt) Group() string {
	return lkt.group
}

// Group returns the minus sign of number
func (lkt *lkt) Minus() string {
	return lkt.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'lkt' and handles both Whole and Real numbers based on 'v'
func (lkt *lkt) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'lkt' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (lkt *lkt) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'lkt'
func (lkt *lkt) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := lkt.currencies[currency]
	l := len(s) + len(symbol) + 5

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, lkt.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	for j := len(symbol) - 1; j >= 0; j-- {
		b = append(b, symbol[j])
	}

	for j := len(lkt.currencyPositivePrefix) - 1; j >= 0; j-- {
		b = append(b, lkt.currencyPositivePrefix[j])
	}

	if num < 0 {
		b = append(b, lkt.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, lkt.currencyPositiveSuffix...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'lkt'
// in accounting notation.
func (lkt *lkt) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := lkt.currencies[currency]
	l := len(s) + len(symbol) + 5

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, lkt.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(lkt.currencyNegativePrefix) - 1; j >= 0; j-- {
			b = append(b, lkt.currencyNegativePrefix[j])
		}

		b = append(b, lkt.minus[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(lkt.currencyPositivePrefix) - 1; j >= 0; j-- {
			b = append(b, lkt.currencyPositivePrefix[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if num < 0 {
		b = append(b, lkt.currencyNegativeSuffix...)
	} else {

		b = append(b, lkt.currencyPositiveSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'lkt'
func (lkt *lkt) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0x2f}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2f}...)

	if t.Year() > 9 {
		b = append(b, strconv.Itoa(t.Year())[2:]...)
	} else {
		b = append(b, strconv.Itoa(t.Year())[1:]...)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'lkt'
func (lkt *lkt) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, lkt.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'lkt'
func (lkt *lkt) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, lkt.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'lkt'
func (lkt *lkt) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, lkt.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, lkt.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'lkt'
func (lkt *lkt) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, lkt.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, lkt.periodsAbbreviated[0]...)
	} else {
		b = append(b, lkt.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'lkt'
func (lkt *lkt) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, lkt.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, lkt.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, lkt.periodsAbbreviated[0]...)
	} else {
		b = append(b, lkt.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'lkt'
func (lkt *lkt) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, lkt.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, lkt.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, lkt.periodsAbbreviated[0]...)
	} else {
		b = append(b, lkt.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'lkt'
func (lkt *lkt) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, lkt.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, lkt.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, lkt.periodsAbbreviated[0]...)
	} else {
		b = append(b, lkt.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := lkt.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
