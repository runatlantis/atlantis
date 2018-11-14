package az_Latn

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type az_Latn struct {
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
	currencyPositivePrefix string
	currencyNegativePrefix string
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

// New returns a new instance of translator for the 'az_Latn' locale
func New() locales.Translator {
	return &az_Latn{
		locale:                 "az_Latn",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         []locales.PluralRule{2, 4, 5, 6},
		pluralsRange:           []locales.PluralRule{2, 6},
		decimal:                ",",
		group:                  ".",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositivePrefix: " ",
		currencyNegativePrefix: " ",
		monthsAbbreviated:      []string{"", "yan", "fev", "mar", "apr", "may", "iyn", "iyl", "avq", "sen", "okt", "noy", "dek"},
		monthsNarrow:           []string{"", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"},
		monthsWide:             []string{"", "yanvar", "fevral", "mart", "aprel", "may", "iyun", "iyul", "avqust", "sentyabr", "oktyabr", "noyabr", "dekabr"},
		daysAbbreviated:        []string{"B.", "B.E.", "Ç.A.", "Ç.", "C.A.", "C.", "Ş."},
		daysNarrow:             []string{"7", "1", "2", "3", "4", "5", "6"},
		daysShort:              []string{"B.", "B.E.", "Ç.A.", "Ç.", "C.A.", "C.", "Ş."},
		daysWide:               []string{"bazar", "bazar ertəsi", "çərşənbə axşamı", "çərşənbə", "cümə axşamı", "cümə", "şənbə"},
		periodsAbbreviated:     []string{"AM", "PM"},
		periodsNarrow:          []string{"a", "p"},
		periodsWide:            []string{"AM", "PM"},
		erasAbbreviated:        []string{"e.ə.", "y.e."},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"eramızdan əvvəl", "yeni era"},
		timezones:              map[string]string{"WITA": "Mərkəzi İndoneziya Vaxtı", "TMST": "Türkmənistan Yay Vaxtı", "GFT": "Fransız Qvianası Vaxtı", "AKDT": "Alyaska Yay Vaxtı", "MDT": "MDT", "CLT": "Çili Standart Vaxtı", "OEZ": "Şərqi Avropa Standart Vaxtı", "BOT": "Boliviya Vaxtı", "ACDT": "Mərkəzi Avstraliya Yay Vaxtı", "AEST": "Şərqi Avstraliya Standart Vaxtı", "WAT": "Qərbi Afrika Standart Vaxtı", "HKT": "Honq Konq Standart Vaxtı", "PST": "Şimali Amerika Sakit Okean Standart Vaxtı", "GYT": "Qayana Vaxtı", "HEOG": "Qərbi Qrenlandiya Yay Vaxtı", "∅∅∅": "Azor Yay Vaxtı", "ChST": "Çamorro Vaxtı", "AWDT": "Qərbi Avstraliya Yay Vaxtı", "NZDT": "Yeni Zelandiya Yay Vaxtı", "ACWDT": "Mərkəzi Qərbi Avstraliya Yay Vaxtı", "CDT": "Şimali Mərkəzi Amerika Yay Vaxtı", "PDT": "Şimali Amerika Sakit Okean Yay Vaxtı", "AWST": "Qərbi Avstraliya Standart Vaxtı", "WIB": "Qərbi İndoneziya Vaxtı", "SRT": "Surinam Vaxtı", "HEEG": "Şərqi Qrenlandiya Yay Vaxtı", "HAT": "Nyufaundlend Yay Vaxtı", "WIT": "Şərqi İndoneziya Vaxtı", "CST": "Şimali Mərkəzi Amerika Standart Vaxtı", "HEPMX": "Meksika Sakit Okean Yay Vaxtı", "BT": "Butan Vaxtı", "JDT": "Yaponiya Yay Vaxtı", "ECT": "Ekvador Vaxtı", "MESZ": "Mərkəzi Avropa Yay Vaxtı", "LHDT": "Lord Hau Yay vaxtı", "HNNOMX": "Şimal-Qərbi Meksika Standart Vaxtı", "OESZ": "Şərqi Avropa Yay Vaxtı", "UYST": "Uruqvay Yay Vaxtı", "HNCU": "Kuba Standart Vaxtı", "HECU": "Kuba Yay Vaxtı", "HNEG": "Şərqi Qrenlandiya Standart Vaxtı", "HNPMX": "Meksika Sakit Okean Standart Vaxtı", "IST": "Hindistan Vaxtı", "HNT": "Nyufaundlend Standart Vaxtı", "HNPM": "Müqəddəs Pyer və Mikelon Standart Vaxtı", "HENOMX": "Şimal-Qərbi Meksika Yay Vaxtı", "COT": "Kolumbiya Standart Vaxtı", "COST": "Kolumbiya Yay Vaxtı", "NZST": "Yeni Zelandiya Standart Vaxtı", "ACWST": "Mərkəzi Qərbi Avstraliya Standart Vaxtı", "MYT": "Malayziya Vaxtı", "MST": "MST", "ART": "Argentina Standart Vaxtı", "ARST": "Argentina Yay Vaxtı", "AKST": "Alyaska Standart Vaxtı", "EAT": "Şərqi Afrika Vaxtı", "CLST": "Çili Yay Vaxtı", "HAST": "Havay-Aleut Standart Vaxtı", "SAST": "Cənubi Afrika Vaxtı", "HEPM": "Müqəddəs Pyer və Mikelon Yay Vaxtı", "WARST": "Qərbi Argentina Yay Vaxtı", "HADT": "Havay-Aleut Yay Vaxtı", "WESZ": "Qərbi Avropa Yay Vaxtı", "HKST": "Honq Konq Yay Vaxtı", "WART": "Qərbi Argentina Standart Vaxtı", "GMT": "Qrinviç Orta Vaxtı", "JST": "Yaponiya Standart Vaxtı", "EDT": "Şimali Şərqi Amerika Yay Vaxtı", "EST": "Şimali Şərqi Amerika Standart Vaxtı", "VET": "Venesuela Vaxtı", "CAT": "Mərkəzi Afrika Vaxtı", "UYT": "Uruqvay Standart Vaxtı", "CHAST": "Çatham Standart Vaxtı", "ADT": "Atlantik Yay Vaxtı", "WAST": "Qərbi Afrika Yay Vaxtı", "SGT": "Sinqapur Vaxtı", "WEZ": "Qərbi Avropa Standart Vaxtı", "LHST": "Lord Hau Standart Vaxtı", "HNOG": "Qərbi Qrenlandiya Standart Vaxtı", "MEZ": "Mərkəzi Avropa Standart Vaxtı", "TMT": "Türkmənistan Standart Vaxtı", "CHADT": "Çatham Yay Vaxtı", "AST": "Atlantik Standart Vaxt", "AEDT": "Şərqi Avstraliya Yay Vaxtı", "ACST": "Mərkəzi Avstraliya Standart Vaxtı"},
	}
}

// Locale returns the current translators string locale
func (az *az_Latn) Locale() string {
	return az.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'az_Latn'
func (az *az_Latn) PluralsCardinal() []locales.PluralRule {
	return az.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'az_Latn'
func (az *az_Latn) PluralsOrdinal() []locales.PluralRule {
	return az.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'az_Latn'
func (az *az_Latn) PluralsRange() []locales.PluralRule {
	return az.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'az_Latn'
func (az *az_Latn) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'az_Latn'
func (az *az_Latn) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)
	iMod100 := i % 100
	iMod1000 := i % 1000
	iMod10 := i % 10

	if (iMod10 == 1 || iMod10 == 2 || iMod10 == 5 || iMod10 == 7 || iMod10 == 8) || (iMod100 == 20 || iMod100 == 50 || iMod100 == 70 || iMod100 == 80) {
		return locales.PluralRuleOne
	} else if (iMod10 == 3 || iMod10 == 4) || (iMod1000 == 100 || iMod1000 == 200 || iMod1000 == 300 || iMod1000 == 400 || iMod1000 == 500 || iMod1000 == 600 || iMod1000 == 700 || iMod1000 == 800 || iMod1000 == 900) {
		return locales.PluralRuleFew
	} else if (i == 0) || (iMod10 == 6) || (iMod100 == 40 || iMod100 == 60 || iMod100 == 90) {
		return locales.PluralRuleMany
	}

	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'az_Latn'
func (az *az_Latn) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := az.CardinalPluralRule(num1, v1)
	end := az.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (az *az_Latn) MonthAbbreviated(month time.Month) string {
	return az.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (az *az_Latn) MonthsAbbreviated() []string {
	return az.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (az *az_Latn) MonthNarrow(month time.Month) string {
	return az.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (az *az_Latn) MonthsNarrow() []string {
	return az.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (az *az_Latn) MonthWide(month time.Month) string {
	return az.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (az *az_Latn) MonthsWide() []string {
	return az.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (az *az_Latn) WeekdayAbbreviated(weekday time.Weekday) string {
	return az.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (az *az_Latn) WeekdaysAbbreviated() []string {
	return az.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (az *az_Latn) WeekdayNarrow(weekday time.Weekday) string {
	return az.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (az *az_Latn) WeekdaysNarrow() []string {
	return az.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (az *az_Latn) WeekdayShort(weekday time.Weekday) string {
	return az.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (az *az_Latn) WeekdaysShort() []string {
	return az.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (az *az_Latn) WeekdayWide(weekday time.Weekday) string {
	return az.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (az *az_Latn) WeekdaysWide() []string {
	return az.daysWide
}

// Decimal returns the decimal point of number
func (az *az_Latn) Decimal() string {
	return az.decimal
}

// Group returns the group of number
func (az *az_Latn) Group() string {
	return az.group
}

// Group returns the minus sign of number
func (az *az_Latn) Minus() string {
	return az.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'az_Latn' and handles both Whole and Real numbers based on 'v'
func (az *az_Latn) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, az.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, az.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, az.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'az_Latn' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (az *az_Latn) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, az.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, az.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, az.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'az_Latn'
func (az *az_Latn) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := az.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, az.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, az.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	for j := len(symbol) - 1; j >= 0; j-- {
		b = append(b, symbol[j])
	}

	for j := len(az.currencyPositivePrefix) - 1; j >= 0; j-- {
		b = append(b, az.currencyPositivePrefix[j])
	}

	if num < 0 {
		b = append(b, az.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, az.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'az_Latn'
// in accounting notation.
func (az *az_Latn) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := az.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, az.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, az.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(az.currencyNegativePrefix) - 1; j >= 0; j-- {
			b = append(b, az.currencyNegativePrefix[j])
		}

		b = append(b, az.minus[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(az.currencyPositivePrefix) - 1; j >= 0; j-- {
			b = append(b, az.currencyPositivePrefix[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, az.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'az_Latn'
func (az *az_Latn) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'az_Latn'
func (az *az_Latn) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, az.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'az_Latn'
func (az *az_Latn) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, az.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'az_Latn'
func (az *az_Latn) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, az.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, az.daysWide[t.Weekday()]...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'az_Latn'
func (az *az_Latn) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, az.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'az_Latn'
func (az *az_Latn) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, az.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, az.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'az_Latn'
func (az *az_Latn) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, az.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, az.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'az_Latn'
func (az *az_Latn) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, az.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, az.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := az.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
