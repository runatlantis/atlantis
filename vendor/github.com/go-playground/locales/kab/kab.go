package kab

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type kab struct {
	locale             string
	pluralsCardinal    []locales.PluralRule
	pluralsOrdinal     []locales.PluralRule
	pluralsRange       []locales.PluralRule
	decimal            string
	group              string
	minus              string
	percent            string
	perMille           string
	timeSeparator      string
	inifinity          string
	currencies         []string // idx = enum of currency code
	monthsAbbreviated  []string
	monthsNarrow       []string
	monthsWide         []string
	daysAbbreviated    []string
	daysNarrow         []string
	daysShort          []string
	daysWide           []string
	periodsAbbreviated []string
	periodsNarrow      []string
	periodsShort       []string
	periodsWide        []string
	erasAbbreviated    []string
	erasNarrow         []string
	erasWide           []string
	timezones          map[string]string
}

// New returns a new instance of translator for the 'kab' locale
func New() locales.Translator {
	return &kab{
		locale:             "kab",
		pluralsCardinal:    []locales.PluralRule{2, 6},
		pluralsOrdinal:     nil,
		pluralsRange:       nil,
		decimal:            ",",
		group:              " ",
		minus:              "-",
		percent:            "%",
		perMille:           "‰",
		timeSeparator:      ":",
		inifinity:          "∞",
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "$AR", "ATS", "$AU", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "FB", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "$BM", "$BN", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "$BS", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "$BZ", "$CA", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "$CL", "CNH", "CNX", "CNY", "$CO", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "£CY", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DA", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "$FJ", "£FK", "F", "£GB", "GEK", "GEL", "GHC", "GHS", "£GI", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "£IE", "£IL", "ILR", "₪", "₹", "IQD", "IRR", "ISJ", "ISK", "₤IT", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "₩", "KWD", "KYD", "KZT", "LAK", "£LB", "LKR", "LRD", "lLS", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "fMA", "MCF", "MDC", "MDL", "MGA", "Fmg", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "£MT", "MUR", "MVP", "MVR", "MWK", "$MX", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "$NA", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "$NZ", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "$RH", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "$SB", "SCR", "SDD", "SDG", "SDP", "SEK", "$SG", "SHP", "SIT", "SKK", "SLL", "SOS", "$SR", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "$TT", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "$US", "USN", "USS", "UYI", "UYP", "$UY", "UZS", "VEB", "VEF", "₫", "VNN", "VUV", "WS$", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "DTS", "XEU", "XFO", "XFU", "CFA", "XPD", "FCFP", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		monthsAbbreviated:  []string{"", "Yen", "Fur", "Meɣ", "Yeb", "May", "Yun", "Yul", "Ɣuc", "Cte", "Tub", "Nun", "Duǧ"},
		monthsNarrow:       []string{"", "Y", "F", "Ɣ", "B", "M", "N", "L", "C", "T", "R", "W", "D"},
		monthsWide:         []string{"", "Yennayer", "Fuṛar", "Meɣres", "Yebrir", "Mayyu", "Yunyu", "Yulyu", "Ɣuct", "Ctembeṛ", "Tubeṛ", "Nunembeṛ", "Duǧembeṛ"},
		daysAbbreviated:    []string{"Yan", "San", "Kraḍ", "Kuẓ", "Sam", "Sḍis", "Say"},
		daysNarrow:         []string{"C", "R", "A", "H", "M", "S", "D"},
		daysShort:          []string{"Cr", "Ri", "Ra", "Hd", "Mh", "Sm", "Sd"},
		daysWide:           []string{"Yanass", "Sanass", "Kraḍass", "Kuẓass", "Samass", "Sḍisass", "Sayass"},
		periodsAbbreviated: []string{"n tufat", "n tmeddit"},
		periodsNarrow:      []string{"f", "m"},
		periodsWide:        []string{"n tufat", "n tmeddit"},
		erasAbbreviated:    []string{"snd. T.Ɛ", "sld. T.Ɛ"},
		erasNarrow:         []string{"", ""},
		erasWide:           []string{"send talalit n Ɛisa", "seld talalit n Ɛisa"},
		timezones:          map[string]string{"CHAST": "Akud Amagnu n Catham", "UYT": "Akud amagnu n Urugway", "JDT": "Akud n Unebdu n Japun", "ACST": "Akud Amagnu n Ustralya Talemmast", "UYST": "Akud n Unebdu n Urugway", "AWST": "Akud Amagnu n Ustṛalya n Umalu", "WIB": "Akud n Umalu n Indunisya", "AKST": "Akud Amagnu n Alaska", "ACDT": "Akud n Unebdu n Ustralya Talemmast", "MEZ": "Akud amagnu n Turuft Talemmast", "OEZ": "Akud amagnu n Turuft n Usammar", "OESZ": "Akud n unebdu n Turuft n Usammar", "AST": "Akud Amagnu Aṭlasan", "WAST": "Akud n unebdu n tefriqt n umalu", "AKDT": "Akud n Unebdu n Alaska", "WART": "Akud Amagnu n Arjuntin n Usammar", "ARST": "Akud n Unebdu n Arjuntin", "PDT": "Akud Amelwi n Unebdu n Marikan n Ugafa", "NZST": "Akud Amagnu n Ziland Tamaynut", "ACWST": "Akud Amagnu n Tlemmast n Umalu n Ustṛalya", "HEOG": "Akud n Unebdu n Grinland n Umalu", "VET": "Akud n Vinizwila", "CLT": "Akud Amagnu n Cili", "GYT": "Akud n Gwiyan", "ADT": "Akud Aṭlasan n Unebdu", "MST": "Akud Amagnu n Idurar n Marikan", "BT": "Akud n Butan", "HKST": "Akud n Unebdu n Hung Kung", "∅∅∅": "Akud n Unebdu n Bṛazilya", "CDT": "Akud n Unebdu n Tlemmast n Marikan", "MDT": "Akud n Unebdu n Idurar n Marikan", "HNEG": "Akud Amagnu n Grinland n Usammar", "HAT": "Akud n Unebdu n Wakal Amaynut", "CAT": "Akud n tefriqt talemmast", "WEZ": "Akud amagnu n turuft n umalu", "HNOG": "Akud Amagnu n Grinland n Umalu", "MESZ": "Akud n unebdu n Turuft Talemmast", "HKT": "Akud Amagnu n Hung Kung", "ChST": "Akud Amagnu n Camuṛṛu", "PST": "Akud Amelwi Amagnu n Marikan n Ugafa", "HEPMX": "Akud Amelwi n Unebdu n Miksik", "SAST": "Akud amagnu n tefriqt n unẓul", "EDT": "Akud n Unebdu n Usammar Agafa n Marikan", "ACWDT": "Akud n Unebdu n Tlemmast n Umalu n Ustṛalya", "WITA": "Akud n Tlemmast n Indunisya", "WIT": "Akud n Usammar n Indunisya", "HAST": "Akud Amagnu n Haway-Aliwsyan", "HADT": "Akud n Unebu n Haway-Aliwsyan", "SGT": "Akud Amagnu n Sangapur", "HNT": "Akud Amagnu n Wakal Amaynut", "HNNOMX": "Akud Amagnu n Ugafa Amalu n Miksik", "HENOMX": "Akud n Unebdu n Ugafa Amalu n Miksik", "COT": "Akud Amagnu n Kulumbya", "GMT": "Akud alemmas n Greenwich", "WESZ": "Akud n unebdu turuft n umalu", "NZDT": "Akud n Unebdu Ziland Tamaynut", "BOT": "Akud n Bulivi", "GFT": "Akud n Gwiyan Tafṛansist", "LHDT": "Akud n Unebdu n Lord Howe", "CLST": "Akud n Unebdu n Cili", "HNCU": "Akud Amagnu n Kuba", "CST": "Akud Amagnu n Tlemmast n Marikan", "JST": "Akud Amagnu n Japun", "AEDT": "Akud n Unebdu n Ustṛalya n Usammar", "HNPM": "Akud Amagnu n San Pyir & Miklun", "SRT": "Akud n Surinam", "EAT": "Akud n tefriqt n usammar", "TMT": "Akud Amagnu n Ṭurkmanistan", "ECT": "Akud n Ikwaṭur", "TMST": "Akud n Unebdu n Ṭurkmanistan", "ART": "Akud Amagnu n Arjuntin", "COST": "Akud n Unebdu n Kulumbya", "CHADT": "Akud n Unebdu Catham", "AWDT": "Akud n Unebdu Ustṛalya n Umalu", "AEST": "Akud Amagnu n Ustṛalya n Usammar", "WAT": "Akud amagnu n tefriqt n umalu", "MYT": "Akud n Malizya", "EST": "Akud Amagnu n Usammar Agafa n Marikan", "HEEG": "Akud n Unebdu n Grinland n Usammar", "IST": "Akud Amagnu n Hend", "HEPM": "Akud n Unebdu n San Pyir & Miklun", "HECU": "Akud n Unebdu n Kuba", "HNPMX": "Akud amagnu Amelwi n Miksik", "LHST": "Akud Amagnu n Lord Howe", "WARST": "Akud n Unebdu n Arjuntin n Usammar"},
	}
}

// Locale returns the current translators string locale
func (kab *kab) Locale() string {
	return kab.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'kab'
func (kab *kab) PluralsCardinal() []locales.PluralRule {
	return kab.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'kab'
func (kab *kab) PluralsOrdinal() []locales.PluralRule {
	return kab.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'kab'
func (kab *kab) PluralsRange() []locales.PluralRule {
	return kab.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'kab'
func (kab *kab) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)

	if i == 0 || i == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'kab'
func (kab *kab) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'kab'
func (kab *kab) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (kab *kab) MonthAbbreviated(month time.Month) string {
	return kab.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (kab *kab) MonthsAbbreviated() []string {
	return kab.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (kab *kab) MonthNarrow(month time.Month) string {
	return kab.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (kab *kab) MonthsNarrow() []string {
	return kab.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (kab *kab) MonthWide(month time.Month) string {
	return kab.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (kab *kab) MonthsWide() []string {
	return kab.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (kab *kab) WeekdayAbbreviated(weekday time.Weekday) string {
	return kab.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (kab *kab) WeekdaysAbbreviated() []string {
	return kab.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (kab *kab) WeekdayNarrow(weekday time.Weekday) string {
	return kab.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (kab *kab) WeekdaysNarrow() []string {
	return kab.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (kab *kab) WeekdayShort(weekday time.Weekday) string {
	return kab.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (kab *kab) WeekdaysShort() []string {
	return kab.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (kab *kab) WeekdayWide(weekday time.Weekday) string {
	return kab.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (kab *kab) WeekdaysWide() []string {
	return kab.daysWide
}

// Decimal returns the decimal point of number
func (kab *kab) Decimal() string {
	return kab.decimal
}

// Group returns the group of number
func (kab *kab) Group() string {
	return kab.group
}

// Group returns the minus sign of number
func (kab *kab) Minus() string {
	return kab.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'kab' and handles both Whole and Real numbers based on 'v'
func (kab *kab) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, kab.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(kab.group) - 1; j >= 0; j-- {
					b = append(b, kab.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, kab.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'kab' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (kab *kab) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, kab.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, kab.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, kab.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'kab'
func (kab *kab) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := kab.currencies[currency]
	l := len(s) + len(symbol) + 2 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, kab.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(kab.group) - 1; j >= 0; j-- {
					b = append(b, kab.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, kab.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, kab.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'kab'
// in accounting notation.
func (kab *kab) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := kab.currencies[currency]
	l := len(s) + len(symbol) + 2 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, kab.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(kab.group) - 1; j >= 0; j-- {
					b = append(b, kab.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, kab.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, kab.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, symbol...)
	} else {

		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'kab'
func (kab *kab) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'kab'
func (kab *kab) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, kab.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'kab'
func (kab *kab) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, kab.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'kab'
func (kab *kab) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, kab.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, kab.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'kab'
func (kab *kab) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, kab.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, kab.periodsAbbreviated[0]...)
	} else {
		b = append(b, kab.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'kab'
func (kab *kab) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, kab.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, kab.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, kab.periodsAbbreviated[0]...)
	} else {
		b = append(b, kab.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'kab'
func (kab *kab) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, kab.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, kab.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, kab.periodsAbbreviated[0]...)
	} else {
		b = append(b, kab.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'kab'
func (kab *kab) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, kab.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, kab.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, kab.periodsAbbreviated[0]...)
	} else {
		b = append(b, kab.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := kab.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
