package it_CH

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type it_CH struct {
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

// New returns a new instance of translator for the 'it_CH' locale
func New() locales.Translator {
	return &it_CH{
		locale:                 "it_CH",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         []locales.PluralRule{5, 6},
		pluralsRange:           []locales.PluralRule{2, 6},
		decimal:                ".",
		group:                  "’",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositivePrefix: " ",
		currencyNegativePrefix: " ",
		monthsAbbreviated:      []string{"", "gen", "feb", "mar", "apr", "mag", "giu", "lug", "ago", "set", "ott", "nov", "dic"},
		monthsNarrow:           []string{"", "G", "F", "M", "A", "M", "G", "L", "A", "S", "O", "N", "D"},
		monthsWide:             []string{"", "gennaio", "febbraio", "marzo", "aprile", "maggio", "giugno", "luglio", "agosto", "settembre", "ottobre", "novembre", "dicembre"},
		daysAbbreviated:        []string{"dom", "lun", "mar", "mer", "gio", "ven", "sab"},
		daysNarrow:             []string{"D", "L", "M", "M", "G", "V", "S"},
		daysShort:              []string{"dom", "lun", "mar", "mer", "gio", "ven", "sab"},
		daysWide:               []string{"domenica", "lunedì", "martedì", "mercoledì", "giovedì", "venerdì", "sabato"},
		periodsAbbreviated:     []string{"AM", "PM"},
		periodsNarrow:          []string{"m.", "p."},
		periodsWide:            []string{"AM", "PM"},
		erasAbbreviated:        []string{"a.C.", "d.C."},
		erasNarrow:             []string{"aC", "dC"},
		erasWide:               []string{"avanti Cristo", "dopo Cristo"},
		timezones:              map[string]string{"HEPM": "Ora legale di Saint-Pierre e Miquelon", "TMT": "Ora standard del Turkmenistan", "HADT": "Ora legale delle Isole Hawaii-Aleutine", "CHADT": "Ora legale delle Chatham", "HECU": "Ora legale di Cuba", "WIB": "Ora dell’Indonesia occidentale", "HNT": "Ora standard di Terranova", "WARST": "Ora legale dell’Argentina occidentale", "AWST": "Ora standard dell’Australia occidentale", "WAST": "Ora legale dell’Africa occidentale", "HNEG": "Ora standard della Groenlandia orientale", "MESZ": "Ora legale dell’Europa centrale", "COST": "Ora legale della Colombia", "AEDT": "Ora legale dell’Australia orientale", "NZDT": "Ora legale della Nuova Zelanda", "AKST": "Ora standard dell’Alaska", "ACWDT": "Ora legale dell’Australia centroccidentale", "HAST": "Ora standard delle Isole Hawaii-Aleutine", "ARST": "Ora legale dell’Argentina", "GYT": "Ora della Guyana", "SGT": "Ora di Singapore", "VET": "Ora del Venezuela", "CLST": "Ora legale del Cile", "NZST": "Ora standard della Nuova Zelanda", "AKDT": "Ora legale dell’Alaska", "HKT": "Ora standard di Hong Kong", "WIT": "Ora dell’Indonesia orientale", "PST": "Ora standard del Pacifico USA", "MDT": "Ora legale Montagne Rocciose USA", "SAST": "Ora dell’Africa meridionale", "BT": "Ora del Bhutan", "HAT": "Ora legale di Terranova", "GMT": "Ora del meridiano di Greenwich", "HNOG": "Ora standard della Groenlandia occidentale", "IST": "Ora standard dell’India", "LHDT": "Ora legale di Lord Howe", "WART": "Ora standard dell’Argentina occidentale", "HNPM": "Ora standard di Saint-Pierre e Miquelon", "WEZ": "Ora standard dell’Europa occidentale", "BOT": "Ora della Bolivia", "HENOMX": "Ora legale del Messico nord-occidentale", "CAT": "Ora dell’Africa centrale", "CST": "Ora standard centrale USA", "MYT": "Ora della Malesia", "ART": "Ora standard dell’Argentina", "GFT": "Ora della Guiana francese", "HKST": "Ora legale di Hong Kong", "SRT": "Ora del Suriname", "UYST": "Ora legale dell’Uruguay", "CDT": "Ora legale centrale USA", "HEPMX": "Ora legale del Pacifico (Messico)", "ADT": "Ora legale dell’Atlantico", "OESZ": "Ora legale dell’Europa orientale", "COT": "Ora standard della Colombia", "CHAST": "Ora standard delle Chatham", "HNCU": "Ora standard di Cuba", "AEST": "Ora standard dell’Australia orientale", "WITA": "Ora dell’Indonesia centrale", "CLT": "Ora standard del Cile", "WAT": "Ora standard dell’Africa occidentale", "JST": "Ora standard del Giappone", "ACST": "Ora standard dell’Australia centrale", "ACDT": "Ora legale dell’Australia centrale", "ChST": "Ora di Chamorro", "WESZ": "Ora legale dell’Europa occidentale", "ECT": "Ora dell’Ecuador", "LHST": "Ora standard di Lord Howe", "EAT": "Ora dell’Africa orientale", "UYT": "Ora standard dell’Uruguay", "HNPMX": "Ora standard del Pacifico (Messico)", "EST": "Ora standard orientale USA", "HEEG": "Ora legale della Groenlandia orientale", "EDT": "Ora legale orientale USA", "MEZ": "Ora standard dell’Europa centrale", "OEZ": "Ora standard dell’Europa orientale", "∅∅∅": "Ora legale di Brasilia", "PDT": "Ora legale del Pacifico USA", "MST": "Ora standard Montagne Rocciose USA", "JDT": "Ora legale del Giappone", "HNNOMX": "Ora standard del Messico nord-occidentale", "TMST": "Ora legale del Turkmenistan", "AWDT": "Ora legale dell’Australia occidentale", "AST": "Ora standard dell’Atlantico", "ACWST": "Ora standard dell’Australia centroccidentale", "HEOG": "Ora legale della Groenlandia occidentale"},
	}
}

// Locale returns the current translators string locale
func (it *it_CH) Locale() string {
	return it.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'it_CH'
func (it *it_CH) PluralsCardinal() []locales.PluralRule {
	return it.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'it_CH'
func (it *it_CH) PluralsOrdinal() []locales.PluralRule {
	return it.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'it_CH'
func (it *it_CH) PluralsRange() []locales.PluralRule {
	return it.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'it_CH'
func (it *it_CH) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)

	if i == 1 && v == 0 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'it_CH'
func (it *it_CH) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 11 || n == 8 || n == 80 || n == 800 {
		return locales.PluralRuleMany
	}

	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'it_CH'
func (it *it_CH) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := it.CardinalPluralRule(num1, v1)
	end := it.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (it *it_CH) MonthAbbreviated(month time.Month) string {
	return it.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (it *it_CH) MonthsAbbreviated() []string {
	return it.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (it *it_CH) MonthNarrow(month time.Month) string {
	return it.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (it *it_CH) MonthsNarrow() []string {
	return it.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (it *it_CH) MonthWide(month time.Month) string {
	return it.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (it *it_CH) MonthsWide() []string {
	return it.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (it *it_CH) WeekdayAbbreviated(weekday time.Weekday) string {
	return it.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (it *it_CH) WeekdaysAbbreviated() []string {
	return it.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (it *it_CH) WeekdayNarrow(weekday time.Weekday) string {
	return it.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (it *it_CH) WeekdaysNarrow() []string {
	return it.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (it *it_CH) WeekdayShort(weekday time.Weekday) string {
	return it.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (it *it_CH) WeekdaysShort() []string {
	return it.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (it *it_CH) WeekdayWide(weekday time.Weekday) string {
	return it.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (it *it_CH) WeekdaysWide() []string {
	return it.daysWide
}

// Decimal returns the decimal point of number
func (it *it_CH) Decimal() string {
	return it.decimal
}

// Group returns the group of number
func (it *it_CH) Group() string {
	return it.group
}

// Group returns the minus sign of number
func (it *it_CH) Minus() string {
	return it.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'it_CH' and handles both Whole and Real numbers based on 'v'
func (it *it_CH) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 3*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, it.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(it.group) - 1; j >= 0; j-- {
					b = append(b, it.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, it.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'it_CH' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (it *it_CH) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, it.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, it.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, it.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'it_CH'
func (it *it_CH) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := it.currencies[currency]
	l := len(s) + len(symbol) + 4 + 3*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, it.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(it.group) - 1; j >= 0; j-- {
					b = append(b, it.group[j])
				}
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

	for j := len(it.currencyPositivePrefix) - 1; j >= 0; j-- {
		b = append(b, it.currencyPositivePrefix[j])
	}

	if num < 0 {
		b = append(b, it.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, it.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'it_CH'
// in accounting notation.
func (it *it_CH) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := it.currencies[currency]
	l := len(s) + len(symbol) + 4 + 3*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, it.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(it.group) - 1; j >= 0; j-- {
					b = append(b, it.group[j])
				}
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

		for j := len(it.currencyNegativePrefix) - 1; j >= 0; j-- {
			b = append(b, it.currencyNegativePrefix[j])
		}

		b = append(b, it.minus[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(it.currencyPositivePrefix) - 1; j >= 0; j-- {
			b = append(b, it.currencyPositivePrefix[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, it.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'it_CH'
func (it *it_CH) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'it_CH'
func (it *it_CH) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, it.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'it_CH'
func (it *it_CH) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, it.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'it_CH'
func (it *it_CH) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, it.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, it.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'it_CH'
func (it *it_CH) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, it.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'it_CH'
func (it *it_CH) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, it.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, it.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'it_CH'
func (it *it_CH) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, it.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, it.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'it_CH'
func (it *it_CH) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, it.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, it.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := it.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
