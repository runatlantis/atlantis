package ru_BY

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type ru_BY struct {
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

// New returns a new instance of translator for the 'ru_BY' locale
func New() locales.Translator {
	return &ru_BY{
		locale:                 "ru_BY",
		pluralsCardinal:        []locales.PluralRule{2, 4, 5, 6},
		pluralsOrdinal:         []locales.PluralRule{6},
		pluralsRange:           []locales.PluralRule{2, 4, 5, 6},
		decimal:                ",",
		group:                  " ",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "Br", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "янв.", "февр.", "мар.", "апр.", "мая", "июн.", "июл.", "авг.", "сент.", "окт.", "нояб.", "дек."},
		monthsNarrow:           []string{"", "Я", "Ф", "М", "А", "М", "И", "И", "А", "С", "О", "Н", "Д"},
		monthsWide:             []string{"", "января", "февраля", "марта", "апреля", "мая", "июня", "июля", "августа", "сентября", "октября", "ноября", "декабря"},
		daysAbbreviated:        []string{"вс", "пн", "вт", "ср", "чт", "пт", "сб"},
		daysNarrow:             []string{"вс", "пн", "вт", "ср", "чт", "пт", "сб"},
		daysShort:              []string{"вс", "пн", "вт", "ср", "чт", "пт", "сб"},
		daysWide:               []string{"воскресенье", "понедельник", "вторник", "среда", "четверг", "пятница", "суббота"},
		periodsAbbreviated:     []string{"AM", "PM"},
		periodsNarrow:          []string{"AM", "PM"},
		periodsWide:            []string{"AM", "PM"},
		erasAbbreviated:        []string{"до н. э.", "н. э."},
		erasNarrow:             []string{"до н.э.", "н.э."},
		erasWide:               []string{"до Рождества Христова", "от Рождества Христова"},
		timezones:              map[string]string{"MESZ": "Центральная Европа, летнее время", "COT": "Колумбия, стандартное время", "SAST": "Южная Африка", "NZST": "Новая Зеландия, стандартное время", "JDT": "Япония, летнее время", "HKST": "Гонконг, летнее время", "HAT": "Ньюфаундленд, летнее время", "HEPM": "Сен-Пьер и Микелон, летнее время", "OESZ": "Восточная Европа, летнее время", "ARST": "Аргентина, летнее время", "GMT": "Среднее время по Гринвичу", "CHAST": "Чатем, стандартное время", "ACWST": "Центральная Австралия, западное стандартное время", "LHDT": "Лорд-Хау, летнее время", "EAT": "Восточная Африка", "CLST": "Чили, летнее время", "HAST": "Гавайско-алеутское стандартное время", "MDT": "Летнее горное время (Северная Америка)", "ACST": "Центральная Австралия, стандартное время", "OEZ": "Восточная Европа, стандартное время", "COST": "Колумбия, летнее время", "CHADT": "Чатем, летнее время", "BOT": "Боливия", "ECT": "Эквадор", "ACWDT": "Центральная Австралия, западное летнее время", "WART": "Западная Аргентина, стандартное время", "HNNOMX": "Северо-западное мексиканское стандартное время", "PDT": "Тихоокеанское летнее время", "WAT": "Западная Африка, стандартное время", "HEEG": "Восточная Гренландия, летнее время", "MEZ": "Центральная Европа, стандартное время", "JST": "Япония, стандартное время", "SGT": "Сингапур", "ACDT": "Центральная Австралия, летнее время", "VET": "Венесуэла", "CAT": "Центральная Африка", "HADT": "Гавайско-алеутское летнее время", "ART": "Аргентина, стандартное время", "HECU": "Куба, летнее время", "∅∅∅": "Бразилия, летнее время", "WESZ": "Западная Европа, летнее время", "AKDT": "Аляска, летнее время", "HKT": "Гонконг, стандартное время", "GYT": "Гайана", "HNCU": "Куба, стандартное время", "ADT": "Атлантическое летнее время", "GFT": "Французская Гвиана", "LHST": "Лорд-Хау, стандартное время", "TMT": "Туркмения, стандартное время", "AWST": "Западная Австралия, стандартное время", "HNPMX": "Тихоокеанское мексиканское стандартное время", "HEPMX": "Тихоокеанское мексиканское летнее время", "BT": "Бутан", "AKST": "Аляска, стандартное время", "IST": "Индия", "MST": "Стандартное горное время (Северная Америка)", "CLT": "Чили, стандартное время", "WIT": "Восточная Индонезия", "ChST": "Чаморро", "CST": "Центральная Америка, стандартное время", "AWDT": "Западная Австралия, летнее время", "WARST": "Западная Аргентина, летнее время", "WITA": "Центральная Индонезия", "HENOMX": "Северо-западное мексиканское летнее время", "TMST": "Туркмения, летнее время", "UYST": "Уругвай, летнее время", "CDT": "Центральная Америка, летнее время", "WEZ": "Западная Европа, стандартное время", "WIB": "Западная Индонезия", "EDT": "Восточная Америка, летнее время", "HNEG": "Восточная Гренландия, стандарное время", "HNOG": "Западная Гренландия, стандартное время", "HEOG": "Западная Гренландия, летнее время", "HNT": "Ньюфаундленд, стандартное время", "SRT": "Суринам", "PST": "Тихоокеанское стандартное время", "AEDT": "Восточная Австралия, летнее время", "MYT": "Малайзия", "AEST": "Восточная Австралия, стандартное время", "EST": "Восточная Америка, стандартное время", "HNPM": "Сен-Пьер и Микелон, стандартное время", "UYT": "Уругвай, стандартное время", "AST": "Атлантическое стандартное время", "WAST": "Западная Африка, летнее время", "NZDT": "Новая Зеландия, летнее время"},
	}
}

// Locale returns the current translators string locale
func (ru *ru_BY) Locale() string {
	return ru.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'ru_BY'
func (ru *ru_BY) PluralsCardinal() []locales.PluralRule {
	return ru.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'ru_BY'
func (ru *ru_BY) PluralsOrdinal() []locales.PluralRule {
	return ru.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'ru_BY'
func (ru *ru_BY) PluralsRange() []locales.PluralRule {
	return ru.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'ru_BY'
func (ru *ru_BY) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)
	iMod10 := i % 10
	iMod100 := i % 100

	if v == 0 && iMod10 == 1 && iMod100 != 11 {
		return locales.PluralRuleOne
	} else if v == 0 && iMod10 >= 2 && iMod10 <= 4 && (iMod100 < 12 || iMod100 > 14) {
		return locales.PluralRuleFew
	} else if (v == 0 && iMod10 == 0) || (v == 0 && iMod10 >= 5 && iMod10 <= 9) || (v == 0 && iMod100 >= 11 && iMod100 <= 14) {
		return locales.PluralRuleMany
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'ru_BY'
func (ru *ru_BY) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'ru_BY'
func (ru *ru_BY) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := ru.CardinalPluralRule(num1, v1)
	end := ru.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleMany {
		return locales.PluralRuleMany
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleFew && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	} else if start == locales.PluralRuleFew && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleFew && end == locales.PluralRuleMany {
		return locales.PluralRuleMany
	} else if start == locales.PluralRuleFew && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleMany && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	} else if start == locales.PluralRuleMany && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleMany && end == locales.PluralRuleMany {
		return locales.PluralRuleMany
	} else if start == locales.PluralRuleMany && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleMany {
		return locales.PluralRuleMany
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (ru *ru_BY) MonthAbbreviated(month time.Month) string {
	return ru.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (ru *ru_BY) MonthsAbbreviated() []string {
	return ru.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (ru *ru_BY) MonthNarrow(month time.Month) string {
	return ru.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (ru *ru_BY) MonthsNarrow() []string {
	return ru.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (ru *ru_BY) MonthWide(month time.Month) string {
	return ru.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (ru *ru_BY) MonthsWide() []string {
	return ru.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (ru *ru_BY) WeekdayAbbreviated(weekday time.Weekday) string {
	return ru.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (ru *ru_BY) WeekdaysAbbreviated() []string {
	return ru.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (ru *ru_BY) WeekdayNarrow(weekday time.Weekday) string {
	return ru.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (ru *ru_BY) WeekdaysNarrow() []string {
	return ru.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (ru *ru_BY) WeekdayShort(weekday time.Weekday) string {
	return ru.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (ru *ru_BY) WeekdaysShort() []string {
	return ru.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (ru *ru_BY) WeekdayWide(weekday time.Weekday) string {
	return ru.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (ru *ru_BY) WeekdaysWide() []string {
	return ru.daysWide
}

// Decimal returns the decimal point of number
func (ru *ru_BY) Decimal() string {
	return ru.decimal
}

// Group returns the group of number
func (ru *ru_BY) Group() string {
	return ru.group
}

// Group returns the minus sign of number
func (ru *ru_BY) Minus() string {
	return ru.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'ru_BY' and handles both Whole and Real numbers based on 'v'
func (ru *ru_BY) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ru.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(ru.group) - 1; j >= 0; j-- {
					b = append(b, ru.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ru.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'ru_BY' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (ru *ru_BY) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ru.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ru.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, ru.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'ru_BY'
func (ru *ru_BY) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ru.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ru.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(ru.group) - 1; j >= 0; j-- {
					b = append(b, ru.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ru.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ru.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, ru.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'ru_BY'
// in accounting notation.
func (ru *ru_BY) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ru.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ru.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(ru.group) - 1; j >= 0; j-- {
					b = append(b, ru.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, ru.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ru.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, ru.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, ru.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'ru_BY'
func (ru *ru_BY) FmtDateShort(t time.Time) string {

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

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'ru_BY'
func (ru *ru_BY) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ru.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20, 0xd0, 0xb3}...)
	b = append(b, []byte{0x2e}...)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'ru_BY'
func (ru *ru_BY) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ru.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20, 0xd0, 0xb3}...)
	b = append(b, []byte{0x2e}...)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'ru_BY'
func (ru *ru_BY) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, ru.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ru.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20, 0xd0, 0xb3}...)
	b = append(b, []byte{0x2e}...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'ru_BY'
func (ru *ru_BY) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ru.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'ru_BY'
func (ru *ru_BY) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ru.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ru.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'ru_BY'
func (ru *ru_BY) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ru.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ru.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'ru_BY'
func (ru *ru_BY) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ru.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ru.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := ru.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
