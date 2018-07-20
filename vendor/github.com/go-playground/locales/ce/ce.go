package ce

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type ce struct {
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

// New returns a new instance of translator for the 'ce' locale
func New() locales.Translator {
	return &ce{
		locale:                 "ce",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         []locales.PluralRule{6},
		pluralsRange:           nil,
		decimal:                ".",
		group:                  ",",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "A$", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "₪", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JP¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "₩", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZ$", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "₽", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "NT$", "TZS", "UAH", "UAK", "UGS", "UGX", "US$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "₫", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "EC$", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		percentSuffix:          " ",
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "янв", "фев", "мар", "апр", "май", "июн", "июл", "авг", "сен", "окт", "ноя", "дек"},
		monthsNarrow:           []string{"", "Я", "Ф", "М", "А", "М", "И", "И", "А", "С", "О", "Н", "Д"},
		monthsWide:             []string{"", "январь", "февраль", "март", "апрель", "май", "июнь", "июль", "август", "сентябрь", "октябрь", "ноябрь", "декабрь"},
		daysAbbreviated:        []string{"кӀи", "ор", "ши", "кха", "еа", "пӀе", "шуо"},
		daysNarrow:             []string{"кӀи", "ор", "ши", "кха", "еа", "пӀе", "шуо"},
		daysShort:              []string{"кӀи", "ор", "ши", "кха", "еа", "пӀе", "шуо"},
		daysWide:               []string{"кӀира", "оршот", "шинара", "кхаара", "еара", "пӀераска", "шуот"},
		erasAbbreviated:        []string{"в. э. тӀ. я", "в. э"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"Ӏийса пайхамар вина де кхачале", "Ӏийса пайхамар вина дийнахь дуьйна"},
		timezones:              map[string]string{"UYST": "Уругвай, аьхкенан хан", "AEST": "Малхбален Австрали, стандартан хан", "JDT": "Япони, аьхкенан хан", "CAT": "Юккъера Африка", "TMST": "Туркменин аьхкенан хан", "AWST": "Малхбузен Австрали, стандартан хан", "HEOG": "Малхбузен Гренланди, аьхкенан хан", "MDT": "MDT", "TMT": "Туркменин стандартан хан", "HADT": "Гавайн-алеутийн аьхкенан хан", "AWDT": "Малхбузен Австрали, аьхкенан хан", "AKDT": "Аляска, аьхкенан хан", "VET": "Венесуэла", "CHADT": "Чатем, аьхкенан хан", "AST": "Атлантикан стандартан хан", "BT": "Бутан", "MYT": "Малайзи", "HENOMX": "Къилбаседа Американ Мексикан аьхкенан хан", "GYT": "Гайана", "CDT": "Юккъера Америка, аьхкенан хан", "NZDT": "Керла Зеланди, аьхкенан хан", "IST": "Инди", "HEPM": "Сен-Пьер а, Микелон а, аьхкенан хан", "GMT": "Гринвичица юкъара хан", "GFT": "Французийн Гвиана", "ACWDT": "Юккъера Австрали, малхбузен аьхкенан хан", "WARST": "Малхбузен Аргентина, аьхкенан хан", "HNT": "Ньюфаундленд, стандартан хан", "HAT": "Ньюфаундленд, аьхкенан хан", "COT": "Колумби, стандартан хан", "WAT": "Малхбузен Африка, стандартан хан", "WEZ": "Малхбузен Европа, стандартан хан", "MESZ": "Юккъера Европа, аьхкенан хан", "HKT": "Гонконг, стандартан хан", "HKST": "Гонконг, аьхкенан хан", "EAT": "Малхбален Африка", "PDT": "Тийна океанан аьхкенан хан", "SGT": "Сингапур", "ACWST": "Юккъера Австрали, малхбузен стандартан хан", "MEZ": "Юккъера Европа, стандартан хан", "ACST": "Юккъера Австрали, стандартан хан", "LHDT": "Лорд-Хау, аьхкенан хан", "OEZ": "Малхбален Европа, стандартан хан", "HECU": "Куба, аьхкенан хан", "HEPMX": "Тийна океанан Мексикан аьхкенан хан", "WAST": "Малхбузен Африка, аьхкенан хан", "JST": "Япони, стандартан хан", "CLST": "Чили, аьхкенан хан", "ChST": "Чаморро", "NZST": "Керла Зеланди, стандартан хан", "AKST": "Аляска, стандартан хан", "HEEG": "Малхбален Гренланди, аьхкенан хан", "WITA": "Юккъера Индонези", "OESZ": "Малхбален Европа, аьхкенан хан", "ART": "Аргентина, стандартан хан", "COST": "Колумби, аьхкенан хан", "ACDT": "Юккъера Австрали, аьхкенан хан", "AEDT": "Малхбален Австрали, аьхкенан хан", "HNEG": "Малхбален Гренланди, стандартан хан", "LHST": "Лорд-Хау, стандартан хан", "CLT": "Чили, стандартан хан", "WIT": "Малхбален Индонези", "HAST": "Гавайн-алеутийн стандартан хан", "∅∅∅": "Амазонка, аьхкенан хан", "CST": "Юккъера Америка, стандартан хан", "HNCU": "Куба, стандартан хан", "HNPMX": "Тийна океанан Мексикан стандартан хан", "WIB": "Малхбузен Индонези", "EST": "Малхбален Америка, стандартан хан", "CHAST": "Чатем, стандартан хан", "BOT": "Боливи", "EDT": "Малхбален Америка, аьхкенан хан", "HNPM": "Сен-Пьер а, Микелон а, стандартан хан", "HNNOMX": "Къилбаседа Американ Мексикан стандартан хан", "MST": "MST", "ARST": "Аргентина, аьхкенан хан", "UYT": "Уругвай, стандартан хан", "HNOG": "Малхбузен Гренланди, стандартан хан", "WART": "Малхбузен Аргентина, стандартан хан", "ECT": "Эквадор", "SRT": "Суринам", "PST": "Тийна океанан стандартан хан", "ADT": "Атлантикан аьхкенан хан", "SAST": "Къилба Африка", "WESZ": "Малхбузен Европа, аьхкенан хан"},
	}
}

// Locale returns the current translators string locale
func (ce *ce) Locale() string {
	return ce.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'ce'
func (ce *ce) PluralsCardinal() []locales.PluralRule {
	return ce.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'ce'
func (ce *ce) PluralsOrdinal() []locales.PluralRule {
	return ce.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'ce'
func (ce *ce) PluralsRange() []locales.PluralRule {
	return ce.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'ce'
func (ce *ce) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'ce'
func (ce *ce) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'ce'
func (ce *ce) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (ce *ce) MonthAbbreviated(month time.Month) string {
	return ce.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (ce *ce) MonthsAbbreviated() []string {
	return ce.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (ce *ce) MonthNarrow(month time.Month) string {
	return ce.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (ce *ce) MonthsNarrow() []string {
	return ce.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (ce *ce) MonthWide(month time.Month) string {
	return ce.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (ce *ce) MonthsWide() []string {
	return ce.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (ce *ce) WeekdayAbbreviated(weekday time.Weekday) string {
	return ce.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (ce *ce) WeekdaysAbbreviated() []string {
	return ce.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (ce *ce) WeekdayNarrow(weekday time.Weekday) string {
	return ce.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (ce *ce) WeekdaysNarrow() []string {
	return ce.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (ce *ce) WeekdayShort(weekday time.Weekday) string {
	return ce.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (ce *ce) WeekdaysShort() []string {
	return ce.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (ce *ce) WeekdayWide(weekday time.Weekday) string {
	return ce.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (ce *ce) WeekdaysWide() []string {
	return ce.daysWide
}

// Decimal returns the decimal point of number
func (ce *ce) Decimal() string {
	return ce.decimal
}

// Group returns the group of number
func (ce *ce) Group() string {
	return ce.group
}

// Group returns the minus sign of number
func (ce *ce) Minus() string {
	return ce.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'ce' and handles both Whole and Real numbers based on 'v'
func (ce *ce) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ce.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ce.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ce.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'ce' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (ce *ce) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 5
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ce.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ce.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, ce.percentSuffix...)

	b = append(b, ce.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'ce'
func (ce *ce) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ce.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ce.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ce.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ce.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ce.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, ce.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'ce'
// in accounting notation.
func (ce *ce) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ce.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ce.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ce.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, ce.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ce.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, ce.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, ce.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'ce'
func (ce *ce) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'ce'
func (ce *ce) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'ce'
func (ce *ce) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'ce'
func (ce *ce) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'ce'
func (ce *ce) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'ce'
func (ce *ce) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'ce'
func (ce *ce) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'ce'
func (ce *ce) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}
