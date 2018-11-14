package prg

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type prg struct {
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

// New returns a new instance of translator for the 'prg' locale
func New() locales.Translator {
	return &prg{
		locale:                 "prg",
		pluralsCardinal:        []locales.PluralRule{1, 2, 6},
		pluralsOrdinal:         []locales.PluralRule{6},
		pluralsRange:           nil,
		decimal:                ",",
		group:                  " ",
		minus:                  "-",
		percent:                "%",
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "rag", "was", "pūl", "sak", "zal", "sīm", "līp", "dag", "sil", "spa", "lap", "sal"},
		monthsNarrow:           []string{"", "R", "W", "P", "S", "Z", "S", "L", "D", "S", "S", "L", "S"},
		monthsWide:             []string{"", "rags", "wassarins", "pūlis", "sakkis", "zallaws", "sīmenis", "līpa", "daggis", "sillins", "spallins", "lapkrūtis", "sallaws"},
		daysAbbreviated:        []string{"nad", "pan", "wis", "pus", "ket", "pēn", "sab"},
		daysNarrow:             []string{"N", "P", "W", "P", "K", "P", "S"},
		daysWide:               []string{"nadīli", "panadīli", "wisasīdis", "pussisawaiti", "ketwirtiks", "pēntniks", "sabattika"},
		periodsAbbreviated:     []string{"AM", "PM"},
		periodsWide:            []string{"ankstāinan", "pa pussideinan"},
		erasAbbreviated:        []string{"BC", "AD"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"", ""},
		timezones:              map[string]string{"EST": "Dēiniskas Amērikas zēimas kerdā", "EDT": "Dēiniskas Amērikas daggas kerdā", "HAT": "HAT", "CAT": "CAT", "ART": "ART", "AWDT": "AWDT", "AEDT": "AEDT", "WAT": "WAT", "SGT": "SGT", "SRT": "SRT", "LHST": "LHST", "AWST": "AWST", "BOT": "BOT", "AKST": "AKST", "ACST": "ACST", "ACDT": "ACDT", "AEST": "AEST", "HENOMX": "HENOMX", "CHAST": "CHAST", "HNPMX": "HNPMX", "HNOG": "HNOG", "HNNOMX": "HNNOMX", "CLST": "CLST", "IST": "IST", "HNPM": "HNPM", "HADT": "HADT", "PDT": "Pacīfiskas Amērikas daggas kerdā", "JDT": "JDT", "AKDT": "AKDT", "MESZ": "Centrālas Eurōpas daggas kerdā", "LHDT": "LHDT", "ECT": "ECT", "HNT": "HNT", "GYT": "GYT", "AST": "Atlāntiska zēimas kerdā", "ADT": "Atlāntiska daggas kerdā", "WEZ": "Wakkariskas Eurōpas zēimas kerdā", "SAST": "SAST", "GFT": "GFT", "UYST": "UYST", "ARST": "ARST", "COT": "COT", "OEZ": "Dēiniskas Eurōpas zēimas kerdā", "HNCU": "HNCU", "CDT": "Centrālas Amērikas daggas kerdā", "HKT": "HKT", "MEZ": "Centrālas Eurōpas zēimas kerdā", "WART": "WART", "TMT": "TMT", "UYT": "UYT", "HECU": "HECU", "∅∅∅": "∅∅∅", "HEPMX": "HEPMX", "WIB": "WIB", "VET": "VET", "CST": "Centrālas Amērikas zēimas kerdā", "MST": "Amērikas gārban zēimas kerdā", "HKST": "HKST", "WITA": "WITA", "CLT": "CLT", "WIT": "WIT", "HAST": "HAST", "NZDT": "NZDT", "HEOG": "HEOG", "WARST": "WARST", "EAT": "EAT", "WESZ": "Wakkariskas Eurōpas daggas kerdā", "MYT": "MYT", "ACWST": "ACWST", "ACWDT": "ACWDT", "HNEG": "HNEG", "CHADT": "CHADT", "WAST": "WAST", "HEPM": "HEPM", "GMT": "Greenwich kerdā", "JST": "JST", "COST": "COST", "OESZ": "Dēiniskas Eurōpas daggas kerdā", "HEEG": "HEEG", "TMST": "TMST", "ChST": "ChST", "PST": "Pacīfiskas Amērikas zēimas kerdā", "MDT": "Amērikas gārban daggas kerdā", "NZST": "NZST", "BT": "BT"},
	}
}

// Locale returns the current translators string locale
func (prg *prg) Locale() string {
	return prg.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'prg'
func (prg *prg) PluralsCardinal() []locales.PluralRule {
	return prg.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'prg'
func (prg *prg) PluralsOrdinal() []locales.PluralRule {
	return prg.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'prg'
func (prg *prg) PluralsRange() []locales.PluralRule {
	return prg.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'prg'
func (prg *prg) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	f := locales.F(n, v)
	nMod10 := math.Mod(n, 10)
	nMod100 := math.Mod(n, 100)
	fMod100 := f % 100
	fMod10 := f % 10

	if (nMod10 == 0) || (nMod100 >= 11 && nMod100 <= 19) || (v == 2 && fMod100 >= 11 && fMod100 <= 19) {
		return locales.PluralRuleZero
	} else if (nMod10 == 1 && nMod100 != 11) || (v == 2 && fMod10 == 1 && fMod100 != 11) || (v != 2 && fMod10 == 1) {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'prg'
func (prg *prg) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'prg'
func (prg *prg) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (prg *prg) MonthAbbreviated(month time.Month) string {
	return prg.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (prg *prg) MonthsAbbreviated() []string {
	return prg.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (prg *prg) MonthNarrow(month time.Month) string {
	return prg.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (prg *prg) MonthsNarrow() []string {
	return prg.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (prg *prg) MonthWide(month time.Month) string {
	return prg.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (prg *prg) MonthsWide() []string {
	return prg.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (prg *prg) WeekdayAbbreviated(weekday time.Weekday) string {
	return prg.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (prg *prg) WeekdaysAbbreviated() []string {
	return prg.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (prg *prg) WeekdayNarrow(weekday time.Weekday) string {
	return prg.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (prg *prg) WeekdaysNarrow() []string {
	return prg.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (prg *prg) WeekdayShort(weekday time.Weekday) string {
	return prg.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (prg *prg) WeekdaysShort() []string {
	return prg.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (prg *prg) WeekdayWide(weekday time.Weekday) string {
	return prg.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (prg *prg) WeekdaysWide() []string {
	return prg.daysWide
}

// Decimal returns the decimal point of number
func (prg *prg) Decimal() string {
	return prg.decimal
}

// Group returns the group of number
func (prg *prg) Group() string {
	return prg.group
}

// Group returns the minus sign of number
func (prg *prg) Minus() string {
	return prg.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'prg' and handles both Whole and Real numbers based on 'v'
func (prg *prg) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, prg.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(prg.group) - 1; j >= 0; j-- {
					b = append(b, prg.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, prg.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'prg' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (prg *prg) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, prg.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, prg.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, prg.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'prg'
func (prg *prg) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := prg.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, prg.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(prg.group) - 1; j >= 0; j-- {
					b = append(b, prg.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, prg.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, prg.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, prg.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'prg'
// in accounting notation.
func (prg *prg) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := prg.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, prg.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(prg.group) - 1; j >= 0; j-- {
					b = append(b, prg.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, prg.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, prg.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, prg.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, prg.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'prg'
func (prg *prg) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Month() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Month()), 10)

	b = append(b, []byte{0x2e}...)

	if t.Year() > 9 {
		b = append(b, strconv.Itoa(t.Year())[2:]...)
	} else {
		b = append(b, strconv.Itoa(t.Year())[1:]...)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'prg'
func (prg *prg) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Month() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Month()), 10)

	b = append(b, []byte{0x20, 0x73, 0x74}...)
	b = append(b, []byte{0x2e, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'prg'
func (prg *prg) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20, 0x6d, 0x65, 0x74, 0x74, 0x61, 0x73}...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, prg.monthsWide[t.Month()]...)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'prg'
func (prg *prg) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, prg.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20, 0x6d, 0x65, 0x74, 0x74, 0x61, 0x73}...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, prg.monthsWide[t.Month()]...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'prg'
func (prg *prg) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, prg.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'prg'
func (prg *prg) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, prg.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, prg.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'prg'
func (prg *prg) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, prg.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, prg.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'prg'
func (prg *prg) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, prg.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, prg.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := prg.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
