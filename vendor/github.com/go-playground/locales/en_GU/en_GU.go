package en_GU

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type en_GU struct {
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

// New returns a new instance of translator for the 'en_GU' locale
func New() locales.Translator {
	return &en_GU{
		locale:                 "en_GU",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         []locales.PluralRule{2, 3, 4, 6},
		pluralsRange:           []locales.PluralRule{6},
		decimal:                ".",
		group:                  ",",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: ")",
		monthsAbbreviated:      []string{"", "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"},
		monthsNarrow:           []string{"", "J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"},
		monthsWide:             []string{"", "January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"},
		daysAbbreviated:        []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"},
		daysNarrow:             []string{"S", "M", "T", "W", "T", "F", "S"},
		daysShort:              []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"},
		daysWide:               []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"},
		periodsAbbreviated:     []string{"am", "pm"},
		periodsNarrow:          []string{"a", "p"},
		periodsWide:            []string{"am", "pm"},
		erasAbbreviated:        []string{"BC", "AD"},
		erasNarrow:             []string{"B", "A"},
		erasWide:               []string{"Before Christ", "Anno Domini"},
		timezones:              map[string]string{"HEOG": "West Greenland Summer Time", "LHDT": "Lord Howe Daylight Time", "CAT": "Central Africa Time", "CLT": "Chile Standard Time", "∅∅∅": "Amazon Summer Time", "HNCU": "Cuba Standard Time", "ADT": "Atlantic Daylight Time", "ACWST": "Australian Central Western Standard Time", "HNEG": "East Greenland Standard Time", "EAT": "East Africa Time", "AEDT": "Australian Eastern Daylight Time", "WAT": "West Africa Standard Time", "WIB": "Western Indonesia Time", "ACST": "Australian Central Standard Time", "OEZ": "Eastern European Standard Time", "COST": "Colombia Summer Time", "HENOMX": "Northwest Mexico Daylight Time", "NZST": "New Zealand Standard Time", "BOT": "Bolivia Time", "HEEG": "East Greenland Summer Time", "WITA": "Central Indonesia Time", "CST": "Central Standard Time", "AKST": "Alaska Standard Time", "HAT": "Newfoundland Daylight Time", "WARST": "Western Argentina Summer Time", "HNT": "Newfoundland Standard Time", "TMT": "Turkmenistan Standard Time", "COT": "Colombia Standard Time", "SAST": "South Africa Standard Time", "WESZ": "Western European Summer Time", "ECT": "Ecuador Time", "TMST": "Turkmenistan Summer Time", "GYT": "Guyana Time", "JDT": "Japan Daylight Time", "ACDT": "Australian Central Daylight Time", "HNNOMX": "Northwest Mexico Standard Time", "AWST": "Australian Western Standard Time", "HNPMX": "Mexican Pacific Standard Time", "CLST": "Chile Summer Time", "IST": "India Standard Time", "VET": "Venezuela Time", "ART": "Argentina Standard Time", "UYT": "Uruguay Standard Time", "AWDT": "Australian Western Daylight Time", "NZDT": "New Zealand Daylight Time", "GFT": "French Guiana Time", "MESZ": "Central European Summer Time", "HKT": "Hong Kong Standard Time", "HADT": "Hawaii-Aleutian Daylight Time", "CHAST": "Chatham Standard Time", "CHADT": "Chatham Daylight Time", "MDT": "Mountain Daylight Time", "MYT": "Malaysia Time", "JST": "Japan Standard Time", "HNOG": "West Greenland Standard Time", "HNPM": "St. Pierre & Miquelon Standard Time", "HAST": "Hawaii-Aleutian Standard Time", "ChST": "Chamorro Standard Time", "HEPMX": "Mexican Pacific Daylight Time", "AEST": "Australian Eastern Standard Time", "ARST": "Argentina Summer Time", "GMT": "Greenwich Mean Time", "MST": "Mountain Standard Time", "LHST": "Lord Howe Standard Time", "BT": "Bhutan Time", "HKST": "Hong Kong Summer Time", "SRT": "Suriname Time", "OESZ": "Eastern European Summer Time", "UYST": "Uruguay Summer Time", "HECU": "Cuba Daylight Time", "WAST": "West Africa Summer Time", "PDT": "Pacific Daylight Time", "MEZ": "Central European Standard Time", "WART": "Western Argentina Standard Time", "WIT": "Eastern Indonesia Time", "CDT": "Central Daylight Time", "PST": "Pacific Standard Time", "AKDT": "Alaska Daylight Time", "EDT": "Eastern Daylight Time", "ACWDT": "Australian Central Western Daylight Time", "HEPM": "St. Pierre & Miquelon Daylight Time", "AST": "Atlantic Standard Time", "WEZ": "Western European Standard Time", "SGT": "Singapore Standard Time", "EST": "Eastern Standard Time"},
	}
}

// Locale returns the current translators string locale
func (en *en_GU) Locale() string {
	return en.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'en_GU'
func (en *en_GU) PluralsCardinal() []locales.PluralRule {
	return en.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'en_GU'
func (en *en_GU) PluralsOrdinal() []locales.PluralRule {
	return en.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'en_GU'
func (en *en_GU) PluralsRange() []locales.PluralRule {
	return en.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'en_GU'
func (en *en_GU) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)

	if i == 1 && v == 0 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'en_GU'
func (en *en_GU) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	nMod10 := math.Mod(n, 10)
	nMod100 := math.Mod(n, 100)

	if nMod10 == 1 && nMod100 != 11 {
		return locales.PluralRuleOne
	} else if nMod10 == 2 && nMod100 != 12 {
		return locales.PluralRuleTwo
	} else if nMod10 == 3 && nMod100 != 13 {
		return locales.PluralRuleFew
	}

	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'en_GU'
func (en *en_GU) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (en *en_GU) MonthAbbreviated(month time.Month) string {
	return en.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (en *en_GU) MonthsAbbreviated() []string {
	return en.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (en *en_GU) MonthNarrow(month time.Month) string {
	return en.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (en *en_GU) MonthsNarrow() []string {
	return en.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (en *en_GU) MonthWide(month time.Month) string {
	return en.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (en *en_GU) MonthsWide() []string {
	return en.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (en *en_GU) WeekdayAbbreviated(weekday time.Weekday) string {
	return en.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (en *en_GU) WeekdaysAbbreviated() []string {
	return en.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (en *en_GU) WeekdayNarrow(weekday time.Weekday) string {
	return en.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (en *en_GU) WeekdaysNarrow() []string {
	return en.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (en *en_GU) WeekdayShort(weekday time.Weekday) string {
	return en.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (en *en_GU) WeekdaysShort() []string {
	return en.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (en *en_GU) WeekdayWide(weekday time.Weekday) string {
	return en.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (en *en_GU) WeekdaysWide() []string {
	return en.daysWide
}

// Decimal returns the decimal point of number
func (en *en_GU) Decimal() string {
	return en.decimal
}

// Group returns the group of number
func (en *en_GU) Group() string {
	return en.group
}

// Group returns the minus sign of number
func (en *en_GU) Minus() string {
	return en.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'en_GU' and handles both Whole and Real numbers based on 'v'
func (en *en_GU) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, en.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, en.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, en.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'en_GU' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (en *en_GU) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, en.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, en.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, en.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'en_GU'
func (en *en_GU) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := en.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, en.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, en.group[0])
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

	if num < 0 {
		b = append(b, en.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, en.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'en_GU'
// in accounting notation.
func (en *en_GU) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := en.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, en.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, en.group[0])
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

		b = append(b, en.currencyNegativePrefix[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, en.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, en.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'en_GU'
func (en *en_GU) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0x2f}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2f}...)

	if t.Year() > 9 {
		b = append(b, strconv.Itoa(t.Year())[2:]...)
	} else {
		b = append(b, strconv.Itoa(t.Year())[1:]...)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'en_GU'
func (en *en_GU) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, en.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'en_GU'
func (en *en_GU) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, en.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'en_GU'
func (en *en_GU) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, en.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, en.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'en_GU'
func (en *en_GU) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, en.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, en.periodsAbbreviated[0]...)
	} else {
		b = append(b, en.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'en_GU'
func (en *en_GU) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, en.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, en.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, en.periodsAbbreviated[0]...)
	} else {
		b = append(b, en.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'en_GU'
func (en *en_GU) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, en.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, en.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, en.periodsAbbreviated[0]...)
	} else {
		b = append(b, en.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'en_GU'
func (en *en_GU) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, en.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, en.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, en.periodsAbbreviated[0]...)
	} else {
		b = append(b, en.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := en.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
