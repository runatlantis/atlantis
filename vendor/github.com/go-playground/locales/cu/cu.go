package cu

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type cu struct {
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

// New returns a new instance of translator for the 'cu' locale
func New() locales.Translator {
	return &cu{
		locale:                 "cu",
		pluralsCardinal:        nil,
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ",",
		group:                  " ",
		minus:                  "-",
		percent:                "%",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JP¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "₸", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "₽", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "₴", "UAK", "UGS", "UGX", "$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		percentSuffix:          " ",
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "і҆аⷩ҇", "феⷡ҇", "маⷬ҇", "а҆пⷬ҇", "маꙵ", "і҆ꙋⷩ҇", "і҆ꙋⷧ҇", "а҆́ѵⷢ҇", "сеⷫ҇", "ѻ҆кⷮ", "ноеⷨ", "деⷦ҇"},
		monthsNarrow:           []string{"", "І҆", "Ф", "М", "А҆", "М", "І҆", "І҆", "А҆", "С", "Ѻ҆", "Н", "Д"},
		monthsWide:             []string{"", "і҆аннꙋа́рїа", "феврꙋа́рїа", "ма́рта", "а҆прі́ллїа", "ма́їа", "і҆ꙋ́нїа", "і҆ꙋ́лїа", "а҆́ѵгꙋста", "септе́мврїа", "ѻ҆ктѡ́врїа", "ное́мврїа", "деке́мврїа"},
		daysAbbreviated:        []string{"ндⷧ҇ѧ", "пнⷣе", "втоⷬ҇", "срⷣе", "чеⷦ҇", "пѧⷦ҇", "сꙋⷠ҇"},
		daysNarrow:             []string{"Н", "П", "В", "С", "Ч", "П", "С"},
		daysShort:              []string{"ндⷧ҇ѧ", "пнⷣе", "втоⷬ҇", "срⷣе", "чеⷦ҇", "пѧⷦ҇", "сꙋⷠ҇"},
		daysWide:               []string{"недѣ́лѧ", "понедѣ́льникъ", "вто́рникъ", "среда̀", "четверто́къ", "пѧто́къ", "сꙋббѡ́та"},
		periodsAbbreviated:     []string{"ДП", "ПП"},
		periodsNarrow:          []string{"ДП", "ПП"},
		periodsWide:            []string{"ДП", "ПП"},
		erasAbbreviated:        []string{"пре́дъ р.\u00a0х.", "ѿ р. х."},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"пре́дъ р.\u00a0х.", "по р.\u00a0х."},
		timezones:              map[string]string{"NZDT": "NZDT", "WART": "WART", "MDT": "MDT", "SRT": "SRT", "GMT": "сре́днее вре́мѧ по грі́нꙋичꙋ", "HEPMX": "HEPMX", "AWDT": "AWDT", "JDT": "JDT", "WARST": "WARST", "HNPMX": "HNPMX", "CDT": "среднеамерїка́нское лѣ́тнее вре́мѧ", "BT": "BT", "AKDT": "AKDT", "ACWST": "ACWST", "HEPM": "HEPM", "AKST": "AKST", "HEOG": "HEOG", "CLT": "CLT", "UYT": "UYT", "AWST": "AWST", "ADT": "а҆тланті́ческое лѣ́тнее вре́мѧ", "AEST": "AEST", "WIB": "WIB", "IST": "IST", "HNT": "HNT", "AEDT": "AEDT", "HEEG": "HEEG", "EAT": "EAT", "HNOG": "HNOG", "MST": "MST", "UYST": "UYST", "ECT": "ECT", "ACST": "ACST", "ACDT": "ACDT", "ACWDT": "ACWDT", "WEZ": "западноєѵрѡпе́йское зи́мнее вре́мѧ", "EST": "восточноамерїка́нское зи́мнее вре́мѧ", "WITA": "WITA", "HNPM": "HNPM", "WIT": "WIT", "TMST": "TMST", "CLST": "CLST", "COT": "COT", "HKT": "HKT", "OEZ": "восточноєѵрѡпе́йское зи́мнее вре́мѧ", "AST": "а҆тланті́ческое зи́мнее вре́мѧ", "WAT": "WAT", "MYT": "MYT", "HENOMX": "HENOMX", "CAT": "CAT", "OESZ": "восточноєѵрѡпе́йское лѣ́тнее вре́мѧ", "HNCU": "HNCU", "BOT": "BOT", "EDT": "восточноамерїка́нское лѣ́тнее вре́мѧ", "LHDT": "LHDT", "HNEG": "HNEG", "LHST": "LHST", "COST": "COST", "CHADT": "CHADT", "CST": "среднеамерїка́нское зи́мнее вре́мѧ", "PST": "тихоѻкеа́нское зи́мнее вре́мѧ", "GFT": "GFT", "NZST": "NZST", "∅∅∅": "∅∅∅", "HADT": "HADT", "JST": "JST", "TMT": "TMT", "HECU": "HECU", "HKST": "HKST", "MESZ": "среднеєѵрѡпе́йское лѣ́тнее вре́мѧ", "HAT": "HAT", "HAST": "HAST", "ART": "ART", "ARST": "ARST", "GYT": "GYT", "WESZ": "западноєѵрѡпе́йское лѣ́тнее вре́мѧ", "SGT": "SGT", "VET": "VET", "HNNOMX": "HNNOMX", "ChST": "ChST", "CHAST": "CHAST", "PDT": "тихоѻкеа́нское лѣ́тнее вре́мѧ", "SAST": "SAST", "WAST": "WAST", "MEZ": "среднеєѵрѡпе́йское зи́мнее вре́мѧ"},
	}
}

// Locale returns the current translators string locale
func (cu *cu) Locale() string {
	return cu.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'cu'
func (cu *cu) PluralsCardinal() []locales.PluralRule {
	return cu.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'cu'
func (cu *cu) PluralsOrdinal() []locales.PluralRule {
	return cu.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'cu'
func (cu *cu) PluralsRange() []locales.PluralRule {
	return cu.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'cu'
func (cu *cu) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'cu'
func (cu *cu) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'cu'
func (cu *cu) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (cu *cu) MonthAbbreviated(month time.Month) string {
	return cu.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (cu *cu) MonthsAbbreviated() []string {
	return cu.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (cu *cu) MonthNarrow(month time.Month) string {
	return cu.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (cu *cu) MonthsNarrow() []string {
	return cu.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (cu *cu) MonthWide(month time.Month) string {
	return cu.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (cu *cu) MonthsWide() []string {
	return cu.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (cu *cu) WeekdayAbbreviated(weekday time.Weekday) string {
	return cu.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (cu *cu) WeekdaysAbbreviated() []string {
	return cu.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (cu *cu) WeekdayNarrow(weekday time.Weekday) string {
	return cu.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (cu *cu) WeekdaysNarrow() []string {
	return cu.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (cu *cu) WeekdayShort(weekday time.Weekday) string {
	return cu.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (cu *cu) WeekdaysShort() []string {
	return cu.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (cu *cu) WeekdayWide(weekday time.Weekday) string {
	return cu.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (cu *cu) WeekdaysWide() []string {
	return cu.daysWide
}

// Decimal returns the decimal point of number
func (cu *cu) Decimal() string {
	return cu.decimal
}

// Group returns the group of number
func (cu *cu) Group() string {
	return cu.group
}

// Group returns the minus sign of number
func (cu *cu) Minus() string {
	return cu.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'cu' and handles both Whole and Real numbers based on 'v'
func (cu *cu) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, cu.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(cu.group) - 1; j >= 0; j-- {
					b = append(b, cu.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, cu.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'cu' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (cu *cu) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 5
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, cu.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, cu.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, cu.percentSuffix...)

	b = append(b, cu.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'cu'
func (cu *cu) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := cu.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, cu.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(cu.group) - 1; j >= 0; j-- {
					b = append(b, cu.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, cu.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, cu.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, cu.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'cu'
// in accounting notation.
func (cu *cu) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := cu.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, cu.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(cu.group) - 1; j >= 0; j-- {
					b = append(b, cu.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, cu.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, cu.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, cu.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, cu.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'cu'
func (cu *cu) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2e}...)

	if t.Month() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Month()), 10)

	b = append(b, []byte{0x2e}...)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'cu'
func (cu *cu) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, cu.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'cu'
func (cu *cu) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, cu.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'cu'
func (cu *cu) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, cu.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, cu.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20, 0xd0, 0xbb}...)
	b = append(b, []byte{0x2e, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2e}...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'cu'
func (cu *cu) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, cu.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'cu'
func (cu *cu) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, cu.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, cu.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'cu'
func (cu *cu) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, cu.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, cu.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'cu'
func (cu *cu) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, cu.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, cu.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := cu.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
