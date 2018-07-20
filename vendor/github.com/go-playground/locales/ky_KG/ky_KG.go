package ky_KG

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type ky_KG struct {
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

// New returns a new instance of translator for the 'ky_KG' locale
func New() locales.Translator {
	return &ky_KG{
		locale:                 "ky_KG",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         []locales.PluralRule{6},
		pluralsRange:           []locales.PluralRule{2, 6},
		decimal:                ",",
		group:                  " ",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "янв.", "фев.", "мар.", "апр.", "май", "июн.", "июл.", "авг.", "сен.", "окт.", "ноя.", "дек."},
		monthsNarrow:           []string{"", "Я", "Ф", "М", "А", "М", "И", "И", "А", "С", "О", "Н", "Д"},
		monthsWide:             []string{"", "январь", "февраль", "март", "апрель", "май", "июнь", "июль", "август", "сентябрь", "октябрь", "ноябрь", "декабрь"},
		daysAbbreviated:        []string{"жек.", "дүй.", "шейш.", "шарш.", "бейш.", "жума", "ишм."},
		daysNarrow:             []string{"Ж", "Д", "Ш", "Ш", "Б", "Ж", "И"},
		daysShort:              []string{"жек.", "дүй.", "шейш.", "шарш.", "бейш.", "жума", "ишм."},
		daysWide:               []string{"жекшемби", "дүйшөмбү", "шейшемби", "шаршемби", "бейшемби", "жума", "ишемби"},
		periodsAbbreviated:     []string{"тң", "тк"},
		periodsNarrow:          []string{"тң", "тк"},
		periodsWide:            []string{"таңкы", "түштөн кийинки"},
		erasAbbreviated:        []string{"б.з.ч.", "б.з."},
		erasNarrow:             []string{"б.з.ч.", "б.з."},
		erasWide:               []string{"", ""},
		timezones:              map[string]string{"MEZ": "Борбордук Европа кышкы убакыты", "AWDT": "Австралия батыш жайкы убактысы", "NZDT": "Жаңы Зеландия жайкы убакыты", "ADT": "Атлантика жайкы убактысы", "MYT": "Малайзия убактысы", "HEEG": "Чыгыш Гренландия жайкы убактысы", "MESZ": "Борбордук Европа жайкы убактысы", "CLT": "Чили кышкы убактысы", "HAT": "Нюфаундлэнд жайкы убактысы", "HEPMX": "Мексика, Тынч океан жайкы убактысы", "WAST": "Батыш Африка жайкы убактысы", "BOT": "Боливия убактысы", "UYT": "Уругвай кышкы убактысы", "AST": "Атлантика кышкы убактысы", "AEDT": "Австралия чыгыш жайкы убактысы", "CDT": "Түндүк Америка, борбордук жайкы убакыт", "WEZ": "Батыш Европа кышкы убакыты", "SGT": "Сингапур убактысы", "HNT": "Нюфаундлэнд кышкы убактысы", "CHADT": "Чатам жайкы убактысы", "COST": "Колумбия жайкы убактысы", "WESZ": "Батыш Европа жайкы убактысы", "HNEG": "Чыгыш Гренландия кышкы убактысы", "HNPM": "Сен Пьер жана Микелон кышкы убактысы", "COT": "Колумбия кышкы убактысы", "AWST": "Австралия батыш кышкы убакыты", "NZST": "Жаӊы Зеландия кышкы убакыты", "JDT": "Жапон жайкы убактысы", "EST": "Түндүк Америка, чыгыш кышкы убактысы", "HENOMX": "Түндүк-чыгыш Мексика жайкы убактысы", "GYT": "Гвиана убактысы", "WARST": "Батыш Аргентина жайкы убактысы", "WITA": "Борбордук Индонезия убактысы", "TMT": "Түркмөнстан кышкы убакыты", "HNCU": "Куба кышкы убактысы", "WAT": "Батыш Африка кышкы убакыты", "JST": "Жапон кышкы убакыты", "ACWDT": "Австралия борбордук чыгыш жайкы убактысы", "IST": "Индия убактысы", "OESZ": "Чыгыш Европа жайкы убактысы", "UYST": "Уругвай жайкы убактысы", "HECU": "Куба жайкы убактысы", "PDT": "Түндүк Америка, Тынч океан жайкы убактысы", "WIB": "Батыш Индонезия убактысы", "ECT": "Экуадор убактысы", "HEOG": "Батыш Гренландия жайкы убактысы", "MST": "MST", "HEPM": "Сен Пьер жана Микелон жайкы убактысы", "ARST": "Аргентина жайкы убактысы", "CHAST": "Чатам кышкы убакыт", "SAST": "Түштүк Африка убактысы", "EDT": "Түндүк Америка, чыгыш жайкы убактысы", "ACST": "Австралия борбордук кышкы убакыты", "ACDT": "Австралия борбордук жайкы убактысы", "OEZ": "Чыгыш Европа кышкы убакыты", "AKST": "Аляска кышкы убактысы", "HNOG": "Батыш Гренландия кышкы убактысы", "VET": "Венесуэла убактысы", "TMST": "Түркмөнстан жайкы убактысы", "HAST": "Гавайи-Алеут кышкы убактысы", "HADT": "Гавайи-Алеут жайкы убактысы", "GFT": "Француз Гвиана убактысы", "HKT": "Гонконг кышкы убакыты", "LHDT": "Лорд Хау жайкы убактысы", "EAT": "Чыгыш Африка убактысы", "SRT": "Суринаме убактысы", "ART": "Аргентина кышкы убактысы", "GMT": "GMT, кышкы убакыты", "CST": "Түндүк Америка, борбордук кышкы убактысы", "BT": "Бутан убактысы", "AKDT": "Аляска жайкы убактысы", "WART": "Батыш Аргентина кышкы убактысы", "MDT": "MDT", "ChST": "Чаморро убактысы", "AEST": "Австралия чыгыш кышкы убакыты", "LHST": "Лорд Хау кышкы убакыты", "HNNOMX": "Түндүк-чыгыш Мексика кышкы убактысы", "CAT": "Борбордук Африка убактысы", "CLST": "Чили жайкы убактысы", "WIT": "Чыгыш Индонезия убактысы", "ACWST": "Австралия борбордук батыш кышкы убакыты", "HKST": "Гонконг жайкы убактысы", "∅∅∅": "Азорс жайкы убактысы", "PST": "Түндүк Америка, Тынч океан кышкы убактысы", "HNPMX": "Мексика, Тынч океан кышкы убактысы"},
	}
}

// Locale returns the current translators string locale
func (ky *ky_KG) Locale() string {
	return ky.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'ky_KG'
func (ky *ky_KG) PluralsCardinal() []locales.PluralRule {
	return ky.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'ky_KG'
func (ky *ky_KG) PluralsOrdinal() []locales.PluralRule {
	return ky.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'ky_KG'
func (ky *ky_KG) PluralsRange() []locales.PluralRule {
	return ky.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'ky_KG'
func (ky *ky_KG) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'ky_KG'
func (ky *ky_KG) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'ky_KG'
func (ky *ky_KG) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := ky.CardinalPluralRule(num1, v1)
	end := ky.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (ky *ky_KG) MonthAbbreviated(month time.Month) string {
	return ky.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (ky *ky_KG) MonthsAbbreviated() []string {
	return ky.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (ky *ky_KG) MonthNarrow(month time.Month) string {
	return ky.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (ky *ky_KG) MonthsNarrow() []string {
	return ky.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (ky *ky_KG) MonthWide(month time.Month) string {
	return ky.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (ky *ky_KG) MonthsWide() []string {
	return ky.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (ky *ky_KG) WeekdayAbbreviated(weekday time.Weekday) string {
	return ky.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (ky *ky_KG) WeekdaysAbbreviated() []string {
	return ky.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (ky *ky_KG) WeekdayNarrow(weekday time.Weekday) string {
	return ky.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (ky *ky_KG) WeekdaysNarrow() []string {
	return ky.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (ky *ky_KG) WeekdayShort(weekday time.Weekday) string {
	return ky.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (ky *ky_KG) WeekdaysShort() []string {
	return ky.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (ky *ky_KG) WeekdayWide(weekday time.Weekday) string {
	return ky.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (ky *ky_KG) WeekdaysWide() []string {
	return ky.daysWide
}

// Decimal returns the decimal point of number
func (ky *ky_KG) Decimal() string {
	return ky.decimal
}

// Group returns the group of number
func (ky *ky_KG) Group() string {
	return ky.group
}

// Group returns the minus sign of number
func (ky *ky_KG) Minus() string {
	return ky.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'ky_KG' and handles both Whole and Real numbers based on 'v'
func (ky *ky_KG) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ky.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(ky.group) - 1; j >= 0; j-- {
					b = append(b, ky.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ky.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'ky_KG' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (ky *ky_KG) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ky.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ky.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, ky.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'ky_KG'
func (ky *ky_KG) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ky.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ky.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(ky.group) - 1; j >= 0; j-- {
					b = append(b, ky.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ky.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ky.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, ky.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'ky_KG'
// in accounting notation.
func (ky *ky_KG) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ky.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ky.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(ky.group) - 1; j >= 0; j-- {
					b = append(b, ky.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, ky.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ky.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, ky.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, ky.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'ky_KG'
func (ky *ky_KG) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2f}...)
	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0x2f}...)

	if t.Year() > 9 {
		b = append(b, strconv.Itoa(t.Year())[2:]...)
	} else {
		b = append(b, strconv.Itoa(t.Year())[1:]...)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'ky_KG'
func (ky *ky_KG) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2d, 0xd0, 0xb6}...)
	b = append(b, []byte{0x2e, 0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2d}...)
	b = append(b, ky.monthsAbbreviated[t.Month()]...)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'ky_KG'
func (ky *ky_KG) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2d, 0xd0, 0xb6}...)
	b = append(b, []byte{0x2e, 0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2d}...)
	b = append(b, ky.monthsWide[t.Month()]...)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'ky_KG'
func (ky *ky_KG) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2d, 0xd0, 0xb6}...)
	b = append(b, []byte{0x2e, 0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2d}...)
	b = append(b, ky.monthsWide[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, ky.daysWide[t.Weekday()]...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'ky_KG'
func (ky *ky_KG) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ky.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'ky_KG'
func (ky *ky_KG) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ky.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ky.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'ky_KG'
func (ky *ky_KG) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ky.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ky.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'ky_KG'
func (ky *ky_KG) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ky.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ky.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := ky.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
