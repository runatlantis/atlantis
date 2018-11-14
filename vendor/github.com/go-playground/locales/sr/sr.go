package sr

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type sr struct {
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

// New returns a new instance of translator for the 'sr' locale
func New() locales.Translator {
	return &sr{
		locale:                 "sr",
		pluralsCardinal:        []locales.PluralRule{2, 4, 6},
		pluralsOrdinal:         []locales.PluralRule{6},
		pluralsRange:           []locales.PluralRule{2, 4, 6},
		decimal:                ",",
		group:                  ".",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "КМ", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "₪", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "NT$", "TZS", "UAH", "UAK", "UGS", "UGX", "US$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "EC$", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositiveSuffix: " ",
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: " )",
		monthsAbbreviated:      []string{"", "јан", "феб", "мар", "апр", "мај", "јун", "јул", "авг", "сеп", "окт", "нов", "дец"},
		monthsNarrow:           []string{"", "ј", "ф", "м", "а", "м", "ј", "ј", "а", "с", "о", "н", "д"},
		monthsWide:             []string{"", "јануар", "фебруар", "март", "април", "мај", "јун", "јул", "август", "септембар", "октобар", "новембар", "децембар"},
		daysAbbreviated:        []string{"нед", "пон", "уто", "сре", "чет", "пет", "суб"},
		daysNarrow:             []string{"н", "п", "у", "с", "ч", "п", "с"},
		daysShort:              []string{"не", "по", "ут", "ср", "че", "пе", "су"},
		daysWide:               []string{"недеља", "понедељак", "уторак", "среда", "четвртак", "петак", "субота"},
		periodsAbbreviated:     []string{"пре подне", "по подне"},
		periodsNarrow:          []string{"a", "p"},
		periodsWide:            []string{"пре подне", "по подне"},
		erasAbbreviated:        []string{"п. н. е.", "н. е."},
		erasNarrow:             []string{"п.н.е.", "н.е."},
		erasWide:               []string{"пре нове ере", "нове ере"},
		timezones:              map[string]string{"AKDT": "Аљаска, летње време", "VET": "Венецуела време", "WIT": "Источно-индонезијско време", "COT": "Колумбија, стандардно време", "MDT": "Северноамеричко планинско летње време", "WESZ": "Западноевропско летње време", "ART": "Аргентина, стандардно време", "ChST": "Чаморо време", "CHADT": "Чатам, летње време", "HNCU": "Куба, стандардно време", "WAT": "Западно-афричко стандардно време", "SGT": "Сингапур, стандардно време", "ECT": "Еквадор време", "∅∅∅": "Бразилија, летње време", "SAST": "Јужно-афричко време", "MESZ": "Средњеевропско летње време", "WART": "Западна Аргентина, стандардно време", "JST": "Јапанско стандардно време", "COST": "Колумбија, летње време", "UYST": "Уругвај, летње време", "HECU": "Куба, летње време", "CST": "Северноамеричко централно стандардно време", "GFT": "Француска Гвајана време", "ACWDT": "Аустралијско централно западно летње време", "MEZ": "Средњеевропско стандардно време", "HEEG": "Источни Гренланд, летње време", "IST": "Индијско стандардно време", "AWST": "Аустралијско западно стандардно време", "HEPMX": "Мексички Пацифик, летње време", "ADT": "Атлантско летње време", "WAST": "Западно-афричко летње време", "EST": "Северноамеричко источно стандардно време", "HNEG": "Источни Гренланд, стандардно време", "LHST": "Лорд Хов, стандардно време", "HNPM": "Сен Пјер и Микелон, стандардно време", "HNNOMX": "Северозападни Мексико, стандардно време", "PDT": "Северноамеричко пацифичко летње време", "NZST": "Нови Зеланд, стандардно време", "NZDT": "Нови Зеланд, летње време", "HNOG": "Западни Гренланд, стандардно време", "HKT": "Хонг Конг, стандардно време", "GMT": "Средње време по Гриничу", "AEDT": "Аустралијско источно летње време", "JDT": "Јапанско летње време", "HEPM": "Сен Пјер и Микелон, летње време", "AKST": "Аљаска, стандардно време", "CLT": "Чиле, стандардно време", "TMT": "Туркменистан, стандардно време", "MST": "Северноамеричко планинско стандардно време", "ACST": "Аустралијско централно стандардно време", "HENOMX": "Северозападни Мексико, летње време", "OESZ": "Источноевропско летње време", "ARST": "Аргентина, летње време", "AST": "Атлантско стандардно време", "BT": "Бутан време", "HKST": "Хонг Конг, летње време", "HAT": "Њуфаундленд, летње време", "EAT": "Источно-афричко време", "OEZ": "Источноевропско стандардно време", "HAST": "Хавајско-алеутско стандардно време", "CHAST": "Чатам, стандардно време", "CDT": "Северноамеричко централно летње време", "HNPMX": "Мексички Пацифик, стандардно време", "WEZ": "Западноевропско стандардно време", "EDT": "Северноамеричко источно летње време", "CLST": "Чиле, летње време", "WARST": "Западна Аргентина, летње време", "WITA": "Централно-индонезијско време", "PST": "Северноамеричко пацифичко стандардно време", "AWDT": "Аустралијско западно летње време", "WIB": "Западно-индонезијско време", "ACDT": "Аустралијско централно летње време", "HEOG": "Западни Гренланд, летње време", "LHDT": "Лорд Хов, летње време", "BOT": "Боливија време", "ACWST": "Аустралијско централно западно стандардно време", "SRT": "Суринам време", "CAT": "Централно-афричко време", "TMST": "Туркменистан, летње време", "HADT": "Хавајско-алеутско летње време", "AEST": "Аустралијско источно стандардно време", "MYT": "Малезија време", "HNT": "Њуфаундленд, стандардно време", "GYT": "Гвајана време", "UYT": "Уругвај, стандардно време"},
	}
}

// Locale returns the current translators string locale
func (sr *sr) Locale() string {
	return sr.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'sr'
func (sr *sr) PluralsCardinal() []locales.PluralRule {
	return sr.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'sr'
func (sr *sr) PluralsOrdinal() []locales.PluralRule {
	return sr.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'sr'
func (sr *sr) PluralsRange() []locales.PluralRule {
	return sr.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'sr'
func (sr *sr) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)
	f := locales.F(n, v)
	iMod100 := i % 100
	iMod10 := i % 10
	fMod10 := f % 10
	fMod100 := f % 100

	if (v == 0 && iMod10 == 1 && iMod100 != 11) || (fMod10 == 1 && fMod100 != 11) {
		return locales.PluralRuleOne
	} else if (v == 0 && iMod10 >= 2 && iMod10 <= 4 && (iMod100 < 12 || iMod100 > 14)) || (fMod10 >= 2 && fMod10 <= 4 && (fMod100 < 12 || fMod100 > 14)) {
		return locales.PluralRuleFew
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'sr'
func (sr *sr) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'sr'
func (sr *sr) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := sr.CardinalPluralRule(num1, v1)
	end := sr.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleFew && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	} else if start == locales.PluralRuleFew && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleFew && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (sr *sr) MonthAbbreviated(month time.Month) string {
	return sr.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (sr *sr) MonthsAbbreviated() []string {
	return sr.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (sr *sr) MonthNarrow(month time.Month) string {
	return sr.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (sr *sr) MonthsNarrow() []string {
	return sr.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (sr *sr) MonthWide(month time.Month) string {
	return sr.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (sr *sr) MonthsWide() []string {
	return sr.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (sr *sr) WeekdayAbbreviated(weekday time.Weekday) string {
	return sr.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (sr *sr) WeekdaysAbbreviated() []string {
	return sr.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (sr *sr) WeekdayNarrow(weekday time.Weekday) string {
	return sr.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (sr *sr) WeekdaysNarrow() []string {
	return sr.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (sr *sr) WeekdayShort(weekday time.Weekday) string {
	return sr.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (sr *sr) WeekdaysShort() []string {
	return sr.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (sr *sr) WeekdayWide(weekday time.Weekday) string {
	return sr.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (sr *sr) WeekdaysWide() []string {
	return sr.daysWide
}

// Decimal returns the decimal point of number
func (sr *sr) Decimal() string {
	return sr.decimal
}

// Group returns the group of number
func (sr *sr) Group() string {
	return sr.group
}

// Group returns the minus sign of number
func (sr *sr) Minus() string {
	return sr.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'sr' and handles both Whole and Real numbers based on 'v'
func (sr *sr) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sr.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, sr.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, sr.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'sr' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (sr *sr) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sr.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, sr.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, sr.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'sr'
func (sr *sr) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := sr.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sr.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, sr.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, sr.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, sr.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, sr.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'sr'
// in accounting notation.
func (sr *sr) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := sr.currencies[currency]
	l := len(s) + len(symbol) + 6 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sr.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, sr.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, sr.currencyNegativePrefix[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, sr.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, sr.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, sr.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'sr'
func (sr *sr) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e}...)
	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Year() > 9 {
		b = append(b, strconv.Itoa(t.Year())[2:]...)
	} else {
		b = append(b, strconv.Itoa(t.Year())[1:]...)
	}

	b = append(b, []byte{0x2e}...)

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'sr'
func (sr *sr) FmtDateMedium(t time.Time) string {

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

	b = append(b, []byte{0x2e}...)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'sr'
func (sr *sr) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, sr.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2e}...)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'sr'
func (sr *sr) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, sr.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, sr.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2e}...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'sr'
func (sr *sr) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sr.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'sr'
func (sr *sr) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sr.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, sr.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'sr'
func (sr *sr) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sr.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, sr.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'sr'
func (sr *sr) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sr.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, sr.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := sr.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
