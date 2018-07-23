package ce_RU

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type ce_RU struct {
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

// New returns a new instance of translator for the 'ce_RU' locale
func New() locales.Translator {
	return &ce_RU{
		locale:                 "ce_RU",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         []locales.PluralRule{6},
		pluralsRange:           nil,
		decimal:                ".",
		group:                  ",",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
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
		timezones:              map[string]string{"WAST": "Малхбузен Африка, аьхкенан хан", "BOT": "Боливи", "MESZ": "Юккъера Европа, аьхкенан хан", "EAT": "Малхбален Африка", "ChST": "Чаморро", "PST": "Тийна океанан стандартан хан", "AEST": "Малхбален Австрали, стандартан хан", "WEZ": "Малхбузен Европа, стандартан хан", "ECT": "Эквадор", "ACWDT": "Юккъера Австрали, малхбузен аьхкенан хан", "HNEG": "Малхбален Гренланди, стандартан хан", "TMT": "Туркменин стандартан хан", "ART": "Аргентина, стандартан хан", "HECU": "Куба, аьхкенан хан", "WART": "Малхбузен Аргентина, стандартан хан", "COST": "Колумби, аьхкенан хан", "MDT": "Лаьмнийн аьхкенан хан (АЦШ)", "SAST": "Къилба Африка", "HKST": "Гонконг, аьхкенан хан", "HNPM": "Сен-Пьер а, Микелон а, стандартан хан", "WAT": "Малхбузен Африка, стандартан хан", "JST": "Япони, стандартан хан", "AKST": "Аляска, стандартан хан", "ACDT": "Юккъера Австрали, аьхкенан хан", "IST": "Инди", "AST": "Атлантикан стандартан хан", "ADT": "Атлантикан аьхкенан хан", "NZST": "Керла Зеланди, стандартан хан", "MYT": "Малайзи", "HAT": "Ньюфаундленд, аьхкенан хан", "VET": "Венесуэла", "COT": "Колумби, стандартан хан", "MST": "Лаьмнийн стандартан хан (АЦШ)", "BT": "Бутан", "∅∅∅": "Амазонка, аьхкенан хан", "EDT": "Малхбален Америка, аьхкенан хан", "HNCU": "Куба, стандартан хан", "CDT": "Юккъера Америка, аьхкенан хан", "AKDT": "Аляска, аьхкенан хан", "ACST": "Юккъера Австрали, стандартан хан", "HEPM": "Сен-Пьер а, Микелон а, аьхкенан хан", "CLT": "Чили, стандартан хан", "HAST": "Гавайн-алеутийн стандартан хан", "CHADT": "Чатем, аьхкенан хан", "GFT": "Французийн Гвиана", "SGT": "Сингапур", "HEEG": "Малхбален Гренланди, аьхкенан хан", "HEOG": "Малхбузен Гренланди, аьхкенан хан", "MEZ": "Юккъера Европа, стандартан хан", "AWDT": "Малхбузен Австрали, аьхкенан хан", "WESZ": "Малхбузен Европа, аьхкенан хан", "WIB": "Малхбузен Индонези", "HNNOMX": "Къилбаседа Американ Мексикан стандартан хан", "EST": "Малхбален Америка, стандартан хан", "HENOMX": "Къилбаседа Американ Мексикан аьхкенан хан", "SRT": "Суринам", "GMT": "Гринвичица юкъара хан", "CST": "Юккъера Америка, стандартан хан", "PDT": "Тийна океанан аьхкенан хан", "HNPMX": "Тийна океанан Мексикан стандартан хан", "WARST": "Малхбузен Аргентина, аьхкенан хан", "HADT": "Гавайн-алеутийн аьхкенан хан", "ARST": "Аргентина, аьхкенан хан", "CHAST": "Чатем, стандартан хан", "HNOG": "Малхбузен Гренланди, стандартан хан", "HKT": "Гонконг, стандартан хан", "WITA": "Юккъера Индонези", "GYT": "Гайана", "UYST": "Уругвай, аьхкенан хан", "HEPMX": "Тийна океанан Мексикан аьхкенан хан", "AWST": "Малхбузен Австрали, стандартан хан", "AEDT": "Малхбален Австрали, аьхкенан хан", "NZDT": "Керла Зеланди, аьхкенан хан", "ACWST": "Юккъера Австрали, малхбузен стандартан хан", "CLST": "Чили, аьхкенан хан", "WIT": "Малхбален Индонези", "UYT": "Уругвай, стандартан хан", "HNT": "Ньюфаундленд, стандартан хан", "TMST": "Туркменин аьхкенан хан", "OEZ": "Малхбален Европа, стандартан хан", "OESZ": "Малхбален Европа, аьхкенан хан", "LHDT": "Лорд-Хау, аьхкенан хан", "CAT": "Юккъера Африка", "JDT": "Япони, аьхкенан хан", "LHST": "Лорд-Хау, стандартан хан"},
	}
}

// Locale returns the current translators string locale
func (ce *ce_RU) Locale() string {
	return ce.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'ce_RU'
func (ce *ce_RU) PluralsCardinal() []locales.PluralRule {
	return ce.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'ce_RU'
func (ce *ce_RU) PluralsOrdinal() []locales.PluralRule {
	return ce.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'ce_RU'
func (ce *ce_RU) PluralsRange() []locales.PluralRule {
	return ce.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'ce_RU'
func (ce *ce_RU) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'ce_RU'
func (ce *ce_RU) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'ce_RU'
func (ce *ce_RU) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (ce *ce_RU) MonthAbbreviated(month time.Month) string {
	return ce.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (ce *ce_RU) MonthsAbbreviated() []string {
	return ce.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (ce *ce_RU) MonthNarrow(month time.Month) string {
	return ce.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (ce *ce_RU) MonthsNarrow() []string {
	return ce.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (ce *ce_RU) MonthWide(month time.Month) string {
	return ce.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (ce *ce_RU) MonthsWide() []string {
	return ce.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (ce *ce_RU) WeekdayAbbreviated(weekday time.Weekday) string {
	return ce.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (ce *ce_RU) WeekdaysAbbreviated() []string {
	return ce.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (ce *ce_RU) WeekdayNarrow(weekday time.Weekday) string {
	return ce.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (ce *ce_RU) WeekdaysNarrow() []string {
	return ce.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (ce *ce_RU) WeekdayShort(weekday time.Weekday) string {
	return ce.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (ce *ce_RU) WeekdaysShort() []string {
	return ce.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (ce *ce_RU) WeekdayWide(weekday time.Weekday) string {
	return ce.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (ce *ce_RU) WeekdaysWide() []string {
	return ce.daysWide
}

// Decimal returns the decimal point of number
func (ce *ce_RU) Decimal() string {
	return ce.decimal
}

// Group returns the group of number
func (ce *ce_RU) Group() string {
	return ce.group
}

// Group returns the minus sign of number
func (ce *ce_RU) Minus() string {
	return ce.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'ce_RU' and handles both Whole and Real numbers based on 'v'
func (ce *ce_RU) FmtNumber(num float64, v uint64) string {

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

// FmtPercent returns 'num' with digits/precision of 'v' for 'ce_RU' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (ce *ce_RU) FmtPercent(num float64, v uint64) string {
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

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'ce_RU'
func (ce *ce_RU) FmtCurrency(num float64, v uint64, currency currency.Type) string {

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

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'ce_RU'
// in accounting notation.
func (ce *ce_RU) FmtAccounting(num float64, v uint64, currency currency.Type) string {

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

// FmtDateShort returns the short date representation of 't' for 'ce_RU'
func (ce *ce_RU) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'ce_RU'
func (ce *ce_RU) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'ce_RU'
func (ce *ce_RU) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'ce_RU'
func (ce *ce_RU) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'ce_RU'
func (ce *ce_RU) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'ce_RU'
func (ce *ce_RU) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'ce_RU'
func (ce *ce_RU) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'ce_RU'
func (ce *ce_RU) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}
