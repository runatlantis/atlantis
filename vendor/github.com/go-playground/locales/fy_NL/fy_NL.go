package fy_NL

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type fy_NL struct {
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

// New returns a new instance of translator for the 'fy_NL' locale
func New() locales.Translator {
	return &fy_NL{
		locale:                 "fy_NL",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         []locales.PluralRule{6},
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
		currencyNegativePrefix: "( ",
		currencyNegativeSuffix: ")",
		monthsAbbreviated:      []string{"", "Jan", "Feb", "Mrt", "Apr", "Mai", "Jun", "Jul", "Aug", "Sep", "Okt", "Nov", "Des"},
		monthsNarrow:           []string{"", "J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"},
		monthsWide:             []string{"", "Jannewaris", "Febrewaris", "Maart", "April", "Maaie", "Juny", "July", "Augustus", "Septimber", "Oktober", "Novimber", "Desimber"},
		daysAbbreviated:        []string{"si", "mo", "ti", "wo", "to", "fr", "so"},
		daysNarrow:             []string{"S", "M", "T", "W", "T", "F", "S"},
		daysShort:              []string{"si", "mo", "ti", "wo", "to", "fr", "so"},
		daysWide:               []string{"snein", "moandei", "tiisdei", "woansdei", "tongersdei", "freed", "sneon"},
		periodsAbbreviated:     []string{"AM", "PM"},
		periodsNarrow:          []string{"AM", "PM"},
		periodsWide:            []string{"AM", "PM"},
		erasAbbreviated:        []string{"f.Kr.", "n.Kr."},
		erasNarrow:             []string{"f.K.", "n.K."},
		erasWide:               []string{"Foar Kristus", "nei Kristus"},
		timezones:              map[string]string{"EST": "Eastern-standerttiid", "CLT": "Sileenske standerttiid", "HNCU": "Kubaanske standerttiid", "OEZ": "East-Europeeske standerttiid", "WESZ": "West-Europeeske simmertiid", "SRT": "Surinaamske tiid", "COT": "Kolombiaanske standerttiid", "OESZ": "East-Europeeske simmertiid", "AEDT": "East-Australyske simmertiid", "HNOG": "West-Groenlânske standerttiid", "WARST": "West-Argentynske simmertiid", "HAT": "Newfoundlânske-simmertiid", "HENOMX": "HENOMX", "GMT": "Greenwich Mean Time", "AWDT": "West-Australyske simmertiid", "HEEG": "East-Groenlânske simmertiid", "WART": "West-Argentynske standerttiid", "HNPM": "Saint Pierre en Miquelon-standerttiid", "ART": "Argentynske standerttiid", "JDT": "Japanske simmertiid", "HNEG": "East-Groenlânske standerttiid", "IST": "Yndiaaske tiid", "ChST": "Chamorro-tiid", "AST": "Atlantic-standerttiid", "HADT": "Hawaii-Aleoetyske simmertiid", "WAST": "West-Afrikaanske simmertiid", "AKST": "Alaska-standerttiid", "ACDT": "Midden-Australyske simmertiid", "HKT": "Hongkongse standerttiid", "HNT": "Newfoundlânske-standerttiid", "CLST": "Sileenske simmertiid", "WEZ": "West-Europeeske standerttiid", "GFT": "Frâns-Guyaanske tiid", "BT": "Bhutaanske tiid", "PST": "Pasifik-standerttiid", "WAT": "West-Afrikaanske standerttiid", "MYT": "Maleisyske tiid", "HEPMX": "HEPMX", "AEST": "East-Australyske standerttiid", "UYST": "Uruguayaanske simmertiid", "JST": "Japanske standerttiid", "NZDT": "Nij-Seelânske simmertiid", "MEZ": "Midden-Europeeske standerttiid", "MDT": "Macause simmertiid", "GYT": "Guyaanske tiid", "PDT": "Pasifik-simmertiid", "NZST": "Nij-Seelânske standerttiid", "EDT": "Eastern-simmertiid", "LHDT": "Lord Howe-eilânske simmertiid", "CHADT": "Chatham simmertiid", "HNPMX": "HNPMX", "BOT": "Boliviaanske tiid", "SGT": "Singaporese standerttiid", "ACWDT": "Midden-Australyske westelijke simmertiid", "EAT": "East-Afrikaanske tiid", "WIB": "West-Yndonezyske tiid", "COST": "Kolombiaanske simmertiid", "ARST": "Argentynske simmertiid", "AWST": "West-Australyske standerttiid", "ECT": "Ecuadoraanske tiid", "MESZ": "Midden-Europeeske simmertiid", "HKST": "Hongkongse simmertiid", "HNNOMX": "HNNOMX", "TMST": "Turkmeense simmertiid", "TMT": "Turkmeense standerttiid", "CHAST": "Chatham standerttiid", "HECU": "Kubaanske simmertiid", "ADT": "Atlantic-simmertiid", "AKDT": "Alaska-simmertiid", "ACST": "Midden-Australyske standerttiid", "WITA": "Sintraal-Yndonezyske tiid", "MST": "Macause standerttiid", "LHST": "Lord Howe-eilânske standerttiid", "CST": "Central-standerttiid", "ACWST": "Midden-Australyske westelijke standerttiid", "∅∅∅": "Amazone-simmertiid", "UYT": "Uruguayaanske standerttiid", "CAT": "Sintraal-Afrikaanske tiid", "CDT": "Central-simmertiid", "HEPM": "Saint Pierre en Miquelon-simmertiid", "WIT": "East-Yndonezyske tiid", "SAST": "Sûd-Afrikaanske tiid", "HEOG": "West-Groenlânske simmertiid", "VET": "Fenezolaanske tiid", "HAST": "Hawaii-Aleoetyske standerttiid"},
	}
}

// Locale returns the current translators string locale
func (fy *fy_NL) Locale() string {
	return fy.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'fy_NL'
func (fy *fy_NL) PluralsCardinal() []locales.PluralRule {
	return fy.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'fy_NL'
func (fy *fy_NL) PluralsOrdinal() []locales.PluralRule {
	return fy.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'fy_NL'
func (fy *fy_NL) PluralsRange() []locales.PluralRule {
	return fy.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'fy_NL'
func (fy *fy_NL) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)

	if i == 1 && v == 0 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'fy_NL'
func (fy *fy_NL) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'fy_NL'
func (fy *fy_NL) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (fy *fy_NL) MonthAbbreviated(month time.Month) string {
	return fy.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (fy *fy_NL) MonthsAbbreviated() []string {
	return fy.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (fy *fy_NL) MonthNarrow(month time.Month) string {
	return fy.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (fy *fy_NL) MonthsNarrow() []string {
	return fy.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (fy *fy_NL) MonthWide(month time.Month) string {
	return fy.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (fy *fy_NL) MonthsWide() []string {
	return fy.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (fy *fy_NL) WeekdayAbbreviated(weekday time.Weekday) string {
	return fy.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (fy *fy_NL) WeekdaysAbbreviated() []string {
	return fy.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (fy *fy_NL) WeekdayNarrow(weekday time.Weekday) string {
	return fy.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (fy *fy_NL) WeekdaysNarrow() []string {
	return fy.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (fy *fy_NL) WeekdayShort(weekday time.Weekday) string {
	return fy.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (fy *fy_NL) WeekdaysShort() []string {
	return fy.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (fy *fy_NL) WeekdayWide(weekday time.Weekday) string {
	return fy.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (fy *fy_NL) WeekdaysWide() []string {
	return fy.daysWide
}

// Decimal returns the decimal point of number
func (fy *fy_NL) Decimal() string {
	return fy.decimal
}

// Group returns the group of number
func (fy *fy_NL) Group() string {
	return fy.group
}

// Group returns the minus sign of number
func (fy *fy_NL) Minus() string {
	return fy.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'fy_NL' and handles both Whole and Real numbers based on 'v'
func (fy *fy_NL) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, fy.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, fy.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, fy.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'fy_NL' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (fy *fy_NL) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, fy.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, fy.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, fy.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'fy_NL'
func (fy *fy_NL) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := fy.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, fy.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, fy.group[0])
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

	for j := len(fy.currencyPositivePrefix) - 1; j >= 0; j-- {
		b = append(b, fy.currencyPositivePrefix[j])
	}

	if num < 0 {
		b = append(b, fy.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, fy.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'fy_NL'
// in accounting notation.
func (fy *fy_NL) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := fy.currencies[currency]
	l := len(s) + len(symbol) + 6 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, fy.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, fy.group[0])
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

		for j := len(fy.currencyNegativePrefix) - 1; j >= 0; j-- {
			b = append(b, fy.currencyNegativePrefix[j])
		}

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(fy.currencyPositivePrefix) - 1; j >= 0; j-- {
			b = append(b, fy.currencyPositivePrefix[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, fy.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, fy.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'fy_NL'
func (fy *fy_NL) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2d}...)

	if t.Month() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Month()), 10)

	b = append(b, []byte{0x2d}...)

	if t.Year() > 9 {
		b = append(b, strconv.Itoa(t.Year())[2:]...)
	} else {
		b = append(b, strconv.Itoa(t.Year())[1:]...)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'fy_NL'
func (fy *fy_NL) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, fy.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'fy_NL'
func (fy *fy_NL) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, fy.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'fy_NL'
func (fy *fy_NL) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, fy.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, fy.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'fy_NL'
func (fy *fy_NL) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, fy.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'fy_NL'
func (fy *fy_NL) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, fy.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, fy.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'fy_NL'
func (fy *fy_NL) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, fy.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, fy.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'fy_NL'
func (fy *fy_NL) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, fy.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, fy.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := fy.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
