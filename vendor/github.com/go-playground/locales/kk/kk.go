package kk

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type kk struct {
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

// New returns a new instance of translator for the 'kk' locale
func New() locales.Translator {
	return &kk{
		locale:                 "kk",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         []locales.PluralRule{5, 6},
		pluralsRange:           []locales.PluralRule{2, 6},
		decimal:                ",",
		group:                  " ",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "A$", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "₪", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JP¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "₩", "KWD", "KYD", "₸", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZ$", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "₽", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "฿", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "NT$", "TZS", "UAH", "UAK", "UGS", "UGX", "$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "₫", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "EC$", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "қаң.", "ақп.", "нау.", "сәу.", "мам.", "мау.", "шіл.", "там.", "қыр.", "қаз.", "қар.", "жел."},
		monthsNarrow:           []string{"", "Қ", "А", "Н", "С", "М", "М", "Ш", "Т", "Қ", "Қ", "Қ", "Ж"},
		monthsWide:             []string{"", "қаңтар", "ақпан", "наурыз", "сәуір", "мамыр", "маусым", "шілде", "тамыз", "қыркүйек", "қазан", "қараша", "желтоқсан"},
		daysAbbreviated:        []string{"Жс", "Дс", "Сс", "Ср", "Бс", "Жм", "Сб"},
		daysNarrow:             []string{"Ж", "Д", "С", "С", "Б", "Ж", "С"},
		daysShort:              []string{"Жс", "Дс", "Сс", "Ср", "Бс", "Жм", "Сб"},
		daysWide:               []string{"жексенбі", "дүйсенбі", "сейсенбі", "сәрсенбі", "бейсенбі", "жұма", "сенбі"},
		periodsAbbreviated:     []string{"AM", "PM"},
		periodsNarrow:          []string{"AM", "PM"},
		periodsWide:            []string{"AM", "PM"},
		erasAbbreviated:        []string{"б.з.д.", "б.з."},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"", ""},
		timezones:              map[string]string{"ART": "Аргентина стандартты уақыты", "COST": "Колумбия жазғы уақыты", "GYT": "Гайана уақыты", "CST": "Солтүстік Америка стандартты орталық уақыты", "PDT": "Солтүстік Америка жазғы Тынық мұхиты уақыты", "∅∅∅": "Азор аралдары жазғы уақыты", "HNT": "Ньюфаундленд стандартты уақыты", "VET": "Венесуэла уақыты", "AWDT": "Австралия жазғы батыс уақыты", "BOT": "Боливия уақыты", "HADT": "Гавай және Алеут аралдары жазғы уақыты", "HNPMX": "Мексика стандартты Тынық мұхит уақыты", "NZST": "Жаңа Зеландия стандартты уақыты", "NZDT": "Жаңа Зеландия жазғы уақыты", "ACST": "Австралия стандартты орталық уақыты", "WARST": "Батыс Аргентина жазғы уақыты", "HEPM": "Сен-Пьер және Микелон жазғы уақыты", "OESZ": "Шығыс Еуропа жазғы уақыты", "WESZ": "Батыс Еуропа жазғы уақыты", "LHST": "Лорд-Хау стандартты уақыты", "WART": "Батыс Аргентина стандартты уақыты", "ARST": "Аргентина жазғы уақыты", "GFT": "Француз Гвианасы уақыты", "HAT": "Ньюфаундленд жазғы уақыты", "TMT": "Түрікменстан стандартты уақыты", "ChST": "Чаморро стандартты уақыты", "AKST": "Аляска стандартты уақыты", "LHDT": "Лорд-Хау жазғы уақыты", "HAST": "Гавай және Алеут аралдары стандартты уақыты", "UYST": "Уругвай жазғы уақыты", "CLST": "Чили жазғы уақыты", "EST": "Солтүстік Америка стандартты шығыс уақыты", "MESZ": "Орталық Еуропа жазғы уақыты", "TMST": "Түрікменстан жазғы уақыты", "EAT": "Шығыс Африка уақыты", "AWST": "Австралия стандартты батыс уақыты", "AST": "Атлантика стандартты уақыты", "WIB": "Батыс Индонезия уақыты", "HEEG": "Шығыс Гренландия жазғы уақыты", "CLT": "Чили стандартты уақыты", "UYT": "Уругвай стандартты уақыты", "CDT": "Солтүстік Америка жазғы орталық уақыты", "MDT": "MDT", "ACWST": "Австралия стандартты орталық-батыс уақыты", "IST": "Үндістан стандартты уақыты", "WIT": "Шығыс Индонезия уақыты", "AKDT": "Аляска жазғы уақыты", "SRT": "Суринам уақыты", "OEZ": "Шығыс Еуропа стандартты уақыты", "GMT": "Гринвич уақыты", "AEDT": "Австралия жазғы шығыс уақыты", "HKST": "Гонконг жазғы уақыты", "HENOMX": "Солтүстік-батыс Мексика жазғы уақыты", "MST": "MST", "CHAST": "Чатем стандартты уақыты", "WAST": "Батыс Африка жазғы уақыты", "MYT": "Малайзия уақыты", "SGT": "Сингапур стандартты уақыты", "ACWDT": "Австралия жазғы орталық-батыс уақыты", "HKT": "Гонконг стандартты уақыты", "HNNOMX": "Солтүстік-батыс Мексика стандартты уақыты", "CAT": "Орталық Африка уақыты", "HNEG": "Шығыс Гренландия стандартты уақыты", "JDT": "Жапония жазғы уақыты", "HEOG": "Батыс Гренландия жазғы уақыты", "EDT": "Солтүстік Америка жазғы шығыс уақыты", "MEZ": "Орталық Еуропа стандартты уақыты", "COT": "Колумбия стандартты уақыты", "ADT": "Атлантика жазғы уақыты", "HNPM": "Сен-Пьер және Микелон стандартты уақыты", "SAST": "Оңтүстік Африка уақыты", "ECT": "Эквадор уақыты", "HNOG": "Батыс Гренландия стандартты уақыты", "AEST": "Австралия стандартты шығыс уақыты", "WAT": "Батыс Африка стандартты уақыты", "JST": "Жапония стандартты уақыты", "WITA": "Орталық Индонезия уақыты", "HNCU": "Куба стандартты уақыты", "HECU": "Куба жазғы уақыты", "WEZ": "Батыс Еуропа стандартты уақыты", "BT": "Бутан уақыты", "ACDT": "Австралия жазғы орталық уақыты", "CHADT": "Чатем жазғы уақыты", "PST": "Солтүстік Америка стандартты Тынық мұхиты уақыты", "HEPMX": "Мексика жазғы Тынық мұхит уақыты"},
	}
}

// Locale returns the current translators string locale
func (kk *kk) Locale() string {
	return kk.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'kk'
func (kk *kk) PluralsCardinal() []locales.PluralRule {
	return kk.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'kk'
func (kk *kk) PluralsOrdinal() []locales.PluralRule {
	return kk.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'kk'
func (kk *kk) PluralsRange() []locales.PluralRule {
	return kk.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'kk'
func (kk *kk) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'kk'
func (kk *kk) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	nMod10 := math.Mod(n, 10)

	if (nMod10 == 6) || (nMod10 == 9) || (nMod10 == 0 && n != 0) {
		return locales.PluralRuleMany
	}

	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'kk'
func (kk *kk) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := kk.CardinalPluralRule(num1, v1)
	end := kk.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (kk *kk) MonthAbbreviated(month time.Month) string {
	return kk.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (kk *kk) MonthsAbbreviated() []string {
	return kk.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (kk *kk) MonthNarrow(month time.Month) string {
	return kk.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (kk *kk) MonthsNarrow() []string {
	return kk.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (kk *kk) MonthWide(month time.Month) string {
	return kk.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (kk *kk) MonthsWide() []string {
	return kk.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (kk *kk) WeekdayAbbreviated(weekday time.Weekday) string {
	return kk.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (kk *kk) WeekdaysAbbreviated() []string {
	return kk.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (kk *kk) WeekdayNarrow(weekday time.Weekday) string {
	return kk.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (kk *kk) WeekdaysNarrow() []string {
	return kk.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (kk *kk) WeekdayShort(weekday time.Weekday) string {
	return kk.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (kk *kk) WeekdaysShort() []string {
	return kk.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (kk *kk) WeekdayWide(weekday time.Weekday) string {
	return kk.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (kk *kk) WeekdaysWide() []string {
	return kk.daysWide
}

// Decimal returns the decimal point of number
func (kk *kk) Decimal() string {
	return kk.decimal
}

// Group returns the group of number
func (kk *kk) Group() string {
	return kk.group
}

// Group returns the minus sign of number
func (kk *kk) Minus() string {
	return kk.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'kk' and handles both Whole and Real numbers based on 'v'
func (kk *kk) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, kk.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(kk.group) - 1; j >= 0; j-- {
					b = append(b, kk.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, kk.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'kk' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (kk *kk) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, kk.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, kk.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, kk.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'kk'
func (kk *kk) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := kk.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, kk.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(kk.group) - 1; j >= 0; j-- {
					b = append(b, kk.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, kk.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, kk.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, kk.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'kk'
// in accounting notation.
func (kk *kk) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := kk.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, kk.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(kk.group) - 1; j >= 0; j-- {
					b = append(b, kk.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, kk.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, kk.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, kk.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, kk.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'kk'
func (kk *kk) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'kk'
func (kk *kk) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20, 0xd0, 0xb6}...)
	b = append(b, []byte{0x2e, 0x20}...)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, kk.monthsAbbreviated[t.Month()]...)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'kk'
func (kk *kk) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20, 0xd0, 0xb6}...)
	b = append(b, []byte{0x2e, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, kk.monthsWide[t.Month()]...)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'kk'
func (kk *kk) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20, 0xd0, 0xb6}...)
	b = append(b, []byte{0x2e, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, kk.monthsWide[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, kk.daysWide[t.Weekday()]...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'kk'
func (kk *kk) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, kk.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'kk'
func (kk *kk) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, kk.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, kk.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'kk'
func (kk *kk) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, kk.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, kk.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'kk'
func (kk *kk) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, kk.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, kk.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := kk.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
