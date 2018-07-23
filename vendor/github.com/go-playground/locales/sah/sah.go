package sah

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type sah struct {
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

// New returns a new instance of translator for the 'sah' locale
func New() locales.Translator {
	return &sah{
		locale:                 "sah",
		pluralsCardinal:        []locales.PluralRule{6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ",",
		group:                  " ",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "₽", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "Тохс", "Олун", "Клн", "Мсу", "Ыам", "Бэс", "Отй", "Атр", "Блҕ", "Алт", "Сэт", "Ахс"},
		monthsNarrow:           []string{"", "Т", "О", "К", "М", "Ы", "Б", "О", "А", "Б", "А", "С", "А"},
		monthsWide:             []string{"", "Тохсунньу", "Олунньу", "Кулун тутар", "Муус устар", "Ыам ыйын", "Бэс ыйын", "От ыйын", "Атырдьых ыйын", "Балаҕан ыйын", "Алтынньы", "Сэтинньи", "ахсынньы"},
		daysAbbreviated:        []string{"бс", "бн", "оп", "сэ", "чп", "бэ", "сб"},
		daysNarrow:             []string{"Б", "Б", "О", "С", "Ч", "Б", "С"},
		daysShort:              []string{"бс", "бн", "оп", "сэ", "чп", "бэ", "сб"},
		daysWide:               []string{"баскыһыанньа", "бэнидиэнньик", "оптуорунньук", "сэрэдэ", "чэппиэр", "Бээтиҥсэ", "субуота"},
		periodsAbbreviated:     []string{"ЭИ", "ЭК"},
		periodsNarrow:          []string{"ЭИ", "ЭК"},
		periodsWide:            []string{"ЭИ", "ЭК"},
		erasAbbreviated:        []string{"б. э. и.", "б. э"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"б. э. и.", "б. э"},
		timezones:              map[string]string{"AKDT": "AKDT", "HKST": "HKST", "WART": "WART", "UYST": "UYST", "AWDT": "Арҕаа Австралия сайыҥҥы кэмэ", "ACWST": "ACWST", "HNEG": "HNEG", "HNPMX": "HNPMX", "WAST": "WAST", "HNPM": "HNPM", "HADT": "HADT", "PDT": "PDT", "SAST": "SAST", "GFT": "GFT", "HEEG": "HEEG", "HEOG": "HEOG", "LHDT": "LHDT", "HEPM": "HEPM", "HNCU": "HNCU", "ARST": "ARST", "HECU": "HECU", "MYT": "MYT", "HNT": "HNT", "WIT": "WIT", "NZDT": "Саҥа Сэйлэнд сайыҥҥы кэмэ", "ACDT": "Киин Австралия сайыҥҥы кэмэ", "IST": "Ииндийэ сүрүн кэмэ", "OEZ": "OEZ", "CDT": "CDT", "AEDT": "Илин Австралия сайыҥҥы кэмэ", "BT": "BT", "JDT": "Дьоппуон сайыҥҥы кэмэ", "LHST": "LHST", "TMST": "TMST", "HEPMX": "HEPMX", "MDT": "MDT", "EST": "EST", "HKT": "HKT", "VET": "VET", "WITA": "WITA", "CAT": "CAT", "COT": "COT", "WEZ": "WEZ", "JST": "Дьоппуон сүрүн кэмэ", "AKST": "AKST", "SRT": "SRT", "CHADT": "CHADT", "AWST": "Арҕаа Австралия сүрүн кэмэ", "OESZ": "OESZ", "AST": "AST", "AEST": "Илин Австралия сүрүн кэмэ", "BOT": "BOT", "ECT": "ECT", "HNOG": "HNOG", "MESZ": "MESZ", "TMT": "TMT", "HAST": "HAST", "ART": "ART", "ChST": "ChST", "CST": "CST", "UYT": "UYT", "MST": "MST", "SGT": "SGT", "ACST": "Киин Австралия сүрүн кэмэ", "ACWDT": "ACWDT", "WARST": "WARST", "HAT": "HAT", "EAT": "EAT", "PST": "PST", "ADT": "ADT", "WESZ": "WESZ", "∅∅∅": "∅∅∅", "HNNOMX": "HNNOMX", "CLST": "CLST", "WAT": "WAT", "WIB": "WIB", "EDT": "EDT", "CLT": "CLT", "GMT": "GMT", "GYT": "GYT", "CHAST": "CHAST", "NZST": "Саҥа Сэйлэнд сүрүн кэмэ", "MEZ": "MEZ", "HENOMX": "HENOMX", "COST": "COST"},
	}
}

// Locale returns the current translators string locale
func (sah *sah) Locale() string {
	return sah.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'sah'
func (sah *sah) PluralsCardinal() []locales.PluralRule {
	return sah.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'sah'
func (sah *sah) PluralsOrdinal() []locales.PluralRule {
	return sah.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'sah'
func (sah *sah) PluralsRange() []locales.PluralRule {
	return sah.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'sah'
func (sah *sah) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'sah'
func (sah *sah) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'sah'
func (sah *sah) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (sah *sah) MonthAbbreviated(month time.Month) string {
	return sah.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (sah *sah) MonthsAbbreviated() []string {
	return sah.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (sah *sah) MonthNarrow(month time.Month) string {
	return sah.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (sah *sah) MonthsNarrow() []string {
	return sah.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (sah *sah) MonthWide(month time.Month) string {
	return sah.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (sah *sah) MonthsWide() []string {
	return sah.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (sah *sah) WeekdayAbbreviated(weekday time.Weekday) string {
	return sah.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (sah *sah) WeekdaysAbbreviated() []string {
	return sah.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (sah *sah) WeekdayNarrow(weekday time.Weekday) string {
	return sah.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (sah *sah) WeekdaysNarrow() []string {
	return sah.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (sah *sah) WeekdayShort(weekday time.Weekday) string {
	return sah.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (sah *sah) WeekdaysShort() []string {
	return sah.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (sah *sah) WeekdayWide(weekday time.Weekday) string {
	return sah.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (sah *sah) WeekdaysWide() []string {
	return sah.daysWide
}

// Decimal returns the decimal point of number
func (sah *sah) Decimal() string {
	return sah.decimal
}

// Group returns the group of number
func (sah *sah) Group() string {
	return sah.group
}

// Group returns the minus sign of number
func (sah *sah) Minus() string {
	return sah.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'sah' and handles both Whole and Real numbers based on 'v'
func (sah *sah) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sah.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(sah.group) - 1; j >= 0; j-- {
					b = append(b, sah.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, sah.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'sah' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (sah *sah) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sah.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, sah.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, sah.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'sah'
func (sah *sah) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := sah.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sah.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(sah.group) - 1; j >= 0; j-- {
					b = append(b, sah.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, sah.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, sah.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, sah.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'sah'
// in accounting notation.
func (sah *sah) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := sah.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sah.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(sah.group) - 1; j >= 0; j-- {
					b = append(b, sah.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, sah.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, sah.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, sah.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, sah.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'sah'
func (sah *sah) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 9 {
		b = append(b, strconv.Itoa(t.Year())[2:]...)
	} else {
		b = append(b, strconv.Itoa(t.Year())[1:]...)
	}

	b = append(b, []byte{0x2f}...)
	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0x2f}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'sah'
func (sah *sah) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, sah.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'sah'
func (sah *sah) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, sah.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'sah'
func (sah *sah) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20, 0xd1, 0x81, 0xd1, 0x8b, 0xd0, 0xbb}...)
	b = append(b, []byte{0x20}...)
	b = append(b, sah.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20, 0xd0, 0xba, 0xd2, 0xaf, 0xd0, 0xbd, 0xd1, 0x8d}...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, sah.daysWide[t.Weekday()]...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'sah'
func (sah *sah) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sah.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'sah'
func (sah *sah) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sah.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, sah.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'sah'
func (sah *sah) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sah.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, sah.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'sah'
func (sah *sah) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sah.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, sah.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := sah.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
