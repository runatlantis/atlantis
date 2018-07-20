package se_SE

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type se_SE struct {
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

// New returns a new instance of translator for the 'se_SE' locale
func New() locales.Translator {
	return &se_SE{
		locale:                 "se_SE",
		pluralsCardinal:        []locales.PluralRule{2, 3, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ",",
		group:                  " ",
		minus:                  "−",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "Nkr", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "kr", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		percentSuffix:          " ",
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "ođđj", "guov", "njuk", "cuo", "mies", "geas", "suoi", "borg", "čakč", "golg", "skáb", "juov"},
		monthsNarrow:           []string{"", "O", "G", "N", "C", "M", "G", "S", "B", "Č", "G", "S", "J"},
		monthsWide:             []string{"", "ođđajagemánnu", "guovvamánnu", "njukčamánnu", "cuoŋománnu", "miessemánnu", "geassemánnu", "suoidnemánnu", "borgemánnu", "čakčamánnu", "golggotmánnu", "skábmamánnu", "juovlamánnu"},
		daysAbbreviated:        []string{"sotn", "vuos", "maŋ", "gask", "duor", "bear", "láv"},
		daysNarrow:             []string{"S", "V", "M", "G", "D", "B", "L"},
		daysShort:              []string{"sotn", "vuos", "maŋ", "gask", "duor", "bear", "láv"},
		daysWide:               []string{"sotnabeaivi", "vuossárga", "maŋŋebárga", "gaskavahkku", "duorasdat", "bearjadat", "lávvardat"},
		periodsAbbreviated:     []string{"i.b.", "e.b."},
		periodsNarrow:          []string{"i.b.", "e.b."},
		periodsWide:            []string{"iđitbeaivet", "eahketbeaivet"},
		erasAbbreviated:        []string{"o.Kr.", "m.Kr."},
		erasNarrow:             []string{"ooá", "oá"},
		erasWide:               []string{"ovdal Kristtusa", "maŋŋel Kristtusa"},
		timezones:              map[string]string{"CHADT": "CHADT", "ACWST": "ACWST", "TMST": "TMST", "AEST": "AEST", "HEOG": "HEOG", "HENOMX": "HENOMX", "EAT": "EAT", "ART": "ART", "CHAST": "CHAST", "EST": "EST", "HNNOMX": "HNNOMX", "COST": "COST", "AWDT": "AWDT", "BT": "BT", "AKST": "AKST", "AKDT": "AKDT", "CAT": "CAT", "HNCU": "HNCU", "SAST": "SAST", "WESZ": "oarje-Eurohpá geassiáigi", "NZDT": "NZDT", "HEEG": "HEEG", "ACDT": "ACDT", "HKST": "HKST", "LHST": "LHST", "HNT": "HNT", "ARST": "ARST", "BOT": "BOT", "JST": "JST", "SRT": "SRT", "WIT": "WIT", "HAST": "HAST", "GYT": "GYT", "HECU": "HECU", "CST": "CST", "WIB": "WIB", "HNEG": "HNEG", "EDT": "EDT", "AST": "AST", "WEZ": "oarje-Eurohpá dábálašáigi", "HNOG": "HNOG", "WART": "WART", "HNPM": "HNPM", "CLST": "CLST", "UYST": "UYST", "OEZ": "nuorti-Eurohpá dábálašáigi", "HADT": "HADT", "∅∅∅": "∅∅∅", "PDT": "PDT", "AWST": "AWST", "MST": "MST", "HKT": "HKT", "IST": "IST", "HAT": "HAT", "UYT": "UYT", "WAST": "WAST", "WARST": "WARST", "WITA": "WITA", "OESZ": "nuorti-Eurohpá geassiáigi", "HEPMX": "HEPMX", "CDT": "CDT", "GFT": "GFT", "PST": "PST", "ADT": "ADT", "AEDT": "AEDT", "WAT": "WAT", "SGT": "SGT", "MEZ": "gaska-Eurohpá dábálašáigi", "VET": "VET", "HNPMX": "HNPMX", "JDT": "JDT", "NZST": "NZST", "ACST": "ACST", "ACWDT": "ACWDT", "MDT": "MDT", "MESZ": "gaska-Eurohpá geassiáigi", "COT": "COT", "ChST": "ChST", "MYT": "MYT", "ECT": "ECT", "LHDT": "LHDT", "HEPM": "HEPM", "TMT": "TMT", "CLT": "CLT", "GMT": "Greenwich gaskka áigi"},
	}
}

// Locale returns the current translators string locale
func (se *se_SE) Locale() string {
	return se.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'se_SE'
func (se *se_SE) PluralsCardinal() []locales.PluralRule {
	return se.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'se_SE'
func (se *se_SE) PluralsOrdinal() []locales.PluralRule {
	return se.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'se_SE'
func (se *se_SE) PluralsRange() []locales.PluralRule {
	return se.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'se_SE'
func (se *se_SE) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	} else if n == 2 {
		return locales.PluralRuleTwo
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'se_SE'
func (se *se_SE) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'se_SE'
func (se *se_SE) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (se *se_SE) MonthAbbreviated(month time.Month) string {
	return se.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (se *se_SE) MonthsAbbreviated() []string {
	return se.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (se *se_SE) MonthNarrow(month time.Month) string {
	return se.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (se *se_SE) MonthsNarrow() []string {
	return se.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (se *se_SE) MonthWide(month time.Month) string {
	return se.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (se *se_SE) MonthsWide() []string {
	return se.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (se *se_SE) WeekdayAbbreviated(weekday time.Weekday) string {
	return se.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (se *se_SE) WeekdaysAbbreviated() []string {
	return se.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (se *se_SE) WeekdayNarrow(weekday time.Weekday) string {
	return se.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (se *se_SE) WeekdaysNarrow() []string {
	return se.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (se *se_SE) WeekdayShort(weekday time.Weekday) string {
	return se.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (se *se_SE) WeekdaysShort() []string {
	return se.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (se *se_SE) WeekdayWide(weekday time.Weekday) string {
	return se.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (se *se_SE) WeekdaysWide() []string {
	return se.daysWide
}

// Decimal returns the decimal point of number
func (se *se_SE) Decimal() string {
	return se.decimal
}

// Group returns the group of number
func (se *se_SE) Group() string {
	return se.group
}

// Group returns the minus sign of number
func (se *se_SE) Minus() string {
	return se.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'se_SE' and handles both Whole and Real numbers based on 'v'
func (se *se_SE) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, se.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(se.group) - 1; j >= 0; j-- {
					b = append(b, se.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(se.minus) - 1; j >= 0; j-- {
			b = append(b, se.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'se_SE' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (se *se_SE) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 7
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, se.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(se.minus) - 1; j >= 0; j-- {
			b = append(b, se.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, se.percentSuffix...)

	b = append(b, se.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'se_SE'
func (se *se_SE) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := se.currencies[currency]
	l := len(s) + len(symbol) + 6 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, se.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(se.group) - 1; j >= 0; j-- {
					b = append(b, se.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(se.minus) - 1; j >= 0; j-- {
			b = append(b, se.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, se.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, se.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'se_SE'
// in accounting notation.
func (se *se_SE) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := se.currencies[currency]
	l := len(s) + len(symbol) + 6 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, se.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(se.group) - 1; j >= 0; j-- {
					b = append(b, se.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		for j := len(se.minus) - 1; j >= 0; j-- {
			b = append(b, se.minus[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, se.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, se.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, se.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'se_SE'
func (se *se_SE) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'se_SE'
func (se *se_SE) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, se.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'se_SE'
func (se *se_SE) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, se.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'se_SE'
func (se *se_SE) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, se.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, se.daysWide[t.Weekday()]...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'se_SE'
func (se *se_SE) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, se.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'se_SE'
func (se *se_SE) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, se.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, se.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'se_SE'
func (se *se_SE) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, se.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, se.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'se_SE'
func (se *se_SE) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, se.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, se.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := se.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
