package ky

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type ky struct {
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

// New returns a new instance of translator for the 'ky' locale
func New() locales.Translator {
	return &ky{
		locale:                 "ky",
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
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JP¥", "KES", "сом", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "฿", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "₫", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
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
		timezones:              map[string]string{"WEZ": "Батыш Европа кышкы убакыты", "WIB": "Батыш Индонезия убактысы", "ACST": "Австралия борбордук кышкы убакыты", "EST": "Түндүк Америка, чыгыш кышкы убактысы", "HENOMX": "Түндүк-чыгыш Мексика жайкы убактысы", "EAT": "Чыгыш Африка убактысы", "CDT": "Түндүк Америка, борбордук жайкы убакыт", "JST": "Жапон кышкы убакыты", "HEOG": "Батыш Гренландия жайкы убактысы", "LHST": "Лорд Хау кышкы убакыты", "WAT": "Батыш Африка кышкы убакыты", "WESZ": "Батыш Европа жайкы убактысы", "BOT": "Боливия убактысы", "WART": "Батыш Аргентина кышкы убактысы", "HNPM": "Сен Пьер жана Микелон кышкы убактысы", "HEPM": "Сен Пьер жана Микелон жайкы убактысы", "SRT": "Суринаме убактысы", "HNEG": "Чыгыш Гренландия кышкы убактысы", "MDT": "MDT", "CAT": "Борбордук Африка убактысы", "ARST": "Аргентина жайкы убактысы", "AEST": "Австралия чыгыш кышкы убакыты", "ChST": "Чаморро убактысы", "PDT": "Түндүк Америка, Тынч океан жайкы убактысы", "HNPMX": "Мексика, Тынч океан кышкы убактысы", "ACDT": "Австралия борбордук жайкы убактысы", "LHDT": "Лорд Хау жайкы убактысы", "HNNOMX": "Түндүк-чыгыш Мексика кышкы убактысы", "CLST": "Чили жайкы убактысы", "UYT": "Уругвай кышкы убактысы", "GYT": "Гвиана убактысы", "PST": "Түндүк Америка, Тынч океан кышкы убактысы", "WAST": "Батыш Африка жайкы убактысы", "WARST": "Батыш Аргентина жайкы убактысы", "HNT": "Нюфаундлэнд кышкы убактысы", "HAT": "Нюфаундлэнд жайкы убактысы", "CLT": "Чили кышкы убактысы", "WIT": "Чыгыш Индонезия убактысы", "HKST": "Гонконг жайкы убактысы", "MST": "MST", "ART": "Аргентина кышкы убактысы", "OESZ": "Чыгыш Европа жайкы убактысы", "UYST": "Уругвай жайкы убактысы", "SGT": "Сингапур убактысы", "EDT": "Түндүк Америка, чыгыш жайкы убактысы", "HEEG": "Чыгыш Гренландия жайкы убактысы", "HKT": "Гонконг кышкы убакыты", "OEZ": "Чыгыш Европа кышкы убакыты", "CHADT": "Чатам жайкы убактысы", "AEDT": "Австралия чыгыш жайкы убактысы", "COT": "Колумбия кышкы убактысы", "CHAST": "Чатам кышкы убакыт", "HEPMX": "Мексика, Тынч океан жайкы убактысы", "BT": "Бутан убактысы", "AKST": "Аляска кышкы убактысы", "ACWST": "Австралия борбордук батыш кышкы убакыты", "TMT": "Түркмөнстан кышкы убакыты", "TMST": "Түркмөнстан жайкы убактысы", "ADT": "Атлантика жайкы убактысы", "HECU": "Куба жайкы убактысы", "AST": "Атлантика кышкы убактысы", "NZDT": "Жаңы Зеландия жайкы убакыты", "GFT": "Француз Гвиана убактысы", "AKDT": "Аляска жайкы убактысы", "ACWDT": "Австралия борбордук чыгыш жайкы убактысы", "∅∅∅": "Азорс жайкы убактысы", "CST": "Түндүк Америка, борбордук кышкы убактысы", "SAST": "Түштүк Африка убактысы", "NZST": "Жаӊы Зеландия кышкы убакыты", "ECT": "Экуадор убактысы", "IST": "Индия убактысы", "VET": "Венесуэла убактысы", "AWST": "Австралия батыш кышкы убакыты", "MYT": "Малайзия убактысы", "WITA": "Борбордук Индонезия убактысы", "GMT": "GMT, кышкы убакыты", "HNOG": "Батыш Гренландия кышкы убактысы", "MEZ": "Борбордук Европа кышкы убакыты", "HAST": "Гавайи-Алеут кышкы убактысы", "HNCU": "Куба кышкы убактысы", "JDT": "Жапон жайкы убактысы", "COST": "Колумбия жайкы убактысы", "MESZ": "Борбордук Европа жайкы убактысы", "HADT": "Гавайи-Алеут жайкы убактысы", "AWDT": "Австралия батыш жайкы убактысы"},
	}
}

// Locale returns the current translators string locale
func (ky *ky) Locale() string {
	return ky.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'ky'
func (ky *ky) PluralsCardinal() []locales.PluralRule {
	return ky.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'ky'
func (ky *ky) PluralsOrdinal() []locales.PluralRule {
	return ky.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'ky'
func (ky *ky) PluralsRange() []locales.PluralRule {
	return ky.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'ky'
func (ky *ky) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'ky'
func (ky *ky) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'ky'
func (ky *ky) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

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
func (ky *ky) MonthAbbreviated(month time.Month) string {
	return ky.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (ky *ky) MonthsAbbreviated() []string {
	return ky.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (ky *ky) MonthNarrow(month time.Month) string {
	return ky.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (ky *ky) MonthsNarrow() []string {
	return ky.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (ky *ky) MonthWide(month time.Month) string {
	return ky.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (ky *ky) MonthsWide() []string {
	return ky.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (ky *ky) WeekdayAbbreviated(weekday time.Weekday) string {
	return ky.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (ky *ky) WeekdaysAbbreviated() []string {
	return ky.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (ky *ky) WeekdayNarrow(weekday time.Weekday) string {
	return ky.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (ky *ky) WeekdaysNarrow() []string {
	return ky.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (ky *ky) WeekdayShort(weekday time.Weekday) string {
	return ky.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (ky *ky) WeekdaysShort() []string {
	return ky.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (ky *ky) WeekdayWide(weekday time.Weekday) string {
	return ky.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (ky *ky) WeekdaysWide() []string {
	return ky.daysWide
}

// Decimal returns the decimal point of number
func (ky *ky) Decimal() string {
	return ky.decimal
}

// Group returns the group of number
func (ky *ky) Group() string {
	return ky.group
}

// Group returns the minus sign of number
func (ky *ky) Minus() string {
	return ky.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'ky' and handles both Whole and Real numbers based on 'v'
func (ky *ky) FmtNumber(num float64, v uint64) string {

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

// FmtPercent returns 'num' with digits/precision of 'v' for 'ky' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (ky *ky) FmtPercent(num float64, v uint64) string {
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

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'ky'
func (ky *ky) FmtCurrency(num float64, v uint64, currency currency.Type) string {

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

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'ky'
// in accounting notation.
func (ky *ky) FmtAccounting(num float64, v uint64, currency currency.Type) string {

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

// FmtDateShort returns the short date representation of 't' for 'ky'
func (ky *ky) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'ky'
func (ky *ky) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'ky'
func (ky *ky) FmtDateLong(t time.Time) string {

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

// FmtDateFull returns the full date representation of 't' for 'ky'
func (ky *ky) FmtDateFull(t time.Time) string {

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

// FmtTimeShort returns the short time representation of 't' for 'ky'
func (ky *ky) FmtTimeShort(t time.Time) string {

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

// FmtTimeMedium returns the medium time representation of 't' for 'ky'
func (ky *ky) FmtTimeMedium(t time.Time) string {

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

// FmtTimeLong returns the long time representation of 't' for 'ky'
func (ky *ky) FmtTimeLong(t time.Time) string {

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

// FmtTimeFull returns the full time representation of 't' for 'ky'
func (ky *ky) FmtTimeFull(t time.Time) string {

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
