package fil_PH

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type fil_PH struct {
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

// New returns a new instance of translator for the 'fil_PH' locale
func New() locales.Translator {
	return &fil_PH{
		locale:                 "fil_PH",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         []locales.PluralRule{2, 6},
		pluralsRange:           []locales.PluralRule{2, 6},
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
		monthsAbbreviated:      []string{"", "Ene", "Peb", "Mar", "Abr", "May", "Hun", "Hul", "Ago", "Set", "Okt", "Nob", "Dis"},
		monthsNarrow:           []string{"", "Ene", "Peb", "Mar", "Abr", "May", "Hun", "Hul", "Ago", "Set", "Okt", "Nob", "Dis"},
		monthsWide:             []string{"", "Enero", "Pebrero", "Marso", "Abril", "Mayo", "Hunyo", "Hulyo", "Agosto", "Setyembre", "Oktubre", "Nobyembre", "Disyembre"},
		daysAbbreviated:        []string{"Lin", "Lun", "Mar", "Miy", "Huw", "Biy", "Sab"},
		daysNarrow:             []string{"Lin", "Lun", "Mar", "Miy", "Huw", "Biy", "Sab"},
		daysShort:              []string{"Li", "Lu", "Ma", "Mi", "Hu", "Bi", "Sa"},
		daysWide:               []string{"Linggo", "Lunes", "Martes", "Miyerkules", "Huwebes", "Biyernes", "Sabado"},
		periodsAbbreviated:     []string{"AM", "PM"},
		periodsNarrow:          []string{"am", "pm"},
		periodsWide:            []string{"AM", "PM"},
		erasAbbreviated:        []string{"BC", "AD"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"Before Christ", "Anno Domini"},
		timezones:              map[string]string{"EAT": "Oras sa Silangang Africa", "TMST": "Oras sa Tag-init ng Turkmenistan", "COT": "Standard na Oras sa Colombia", "BT": "Oras sa Bhutan", "SGT": "Standard na Oras sa Singapore", "ECT": "Oras sa Ecuador", "ACDT": "Daylight Time sa Gitnang Australya", "WART": "Standard na Oras sa Kanlurang Argentina", "GFT": "Oras sa French Guiana", "ACWST": "Standard Time ng Gitnang Kanluran ng Australya", "ACWDT": "Daylight Time sa Gitnang Kanlurang Australya", "HEEG": "Oras sa Tag-init ng Silangang Greenland", "CST": "Sentral na Karaniwang Oras", "CDT": "Sentral na Daylight Time", "AWST": "Standard na Oras sa Kanlurang Australya", "ADT": "Daylight Time sa Atlantiko", "VET": "Oras sa Venezuela", "CHADT": "Daylight Time sa Chatham", "AST": "Standard na Oras sa Atlantiko", "MDT": "Daylight Time sa Bundok", "HNT": "Standard na Oras sa Newfoundland", "HECU": "Daylight Time sa Cuba", "∅∅∅": "Oras sa Tag-init ng Brasilia", "EST": "Eastern na Standard na Oras", "CLT": "Standard na Oras sa Chile", "ARST": "Oras sa Tag-init ng Argentina", "AKST": "Standard na Oras sa Alaska", "MESZ": "Oras sa Tag-init ng Gitnang Europe", "TMT": "Standard na Oras sa Turkmenistan", "UYST": "Oras sa Tag-init ng Uruguay", "PST": "Standard na Oras sa Pasipiko", "MST": "Standard na Oras sa Bundok", "WAT": "Standard na Oras sa Kanlurang Africa", "NZDT": "Daylight Time sa New Zealand", "OESZ": "Oras sa Tag-init ng Silangang Europe", "LHST": "Standard na Oras sa Lord Howe", "HEPM": "Daylight Time sa Saint Pierre & Miquelon", "SRT": "Oras sa Suriname", "PDT": "Daylight Time sa Pasipiko", "AEST": "Standard na Oras sa Silangang Australya", "WIB": "Oras sa Kanlurang Indonesia", "JST": "Standard na Oras sa Japan", "MYT": "Oras sa Malaysia", "HAST": "Standard na Oras sa Hawaii-Aleutian", "CHAST": "Standard na Oras sa Chatham", "AKDT": "Daylight Time sa Alaska", "EDT": "Eastern Daylight Time", "HNEG": "Standard na Oras sa Silangang Greenland", "HKT": "Standard na Oras sa Hong Kong", "MEZ": "Standard na Oras sa Gitnang Europe", "COST": "Oras sa Tag-init ng Colombia", "HNCU": "Standard na Oras sa Cuba", "HNPMX": "Standard na Oras sa Pasipiko ng Mexico", "HEPMX": "Daylight Time sa Pasipiko ng Mexico", "WESZ": "Oras sa Tag-init ng Kanlurang Europe", "JDT": "Daylight Time sa Japan", "ChST": "Standard na Oras sa Chamorro", "AEDT": "Daylight Time sa Silangang Australya", "HENOMX": "Daylight Time sa Hilagang-kanlurang Mexico", "CAT": "Oras sa Gitnang Africa", "CLST": "Oras sa Tag-init ng Chile", "ART": "Standard na Oras sa Argentina", "GYT": "Oras sa Guyana", "SAST": "Oras sa Timog Africa", "WEZ": "Standard na Oras sa Kanlurang Europe", "HNPM": "Standard na Oras sa Saint Pierre & Miquelon", "HNNOMX": "Standard na Oras sa Hilagang-kanlurang Mexico", "UYT": "Standard na Oras sa Uruguay", "ACST": "Standard na Oras sa Gitnang Australya", "HEOG": "Oras sa Tag-init ng Kanlurang Greenland", "HKST": "Oras sa Tag-init ng Hong Kong", "OEZ": "Standard na Oras sa Silangang Europe", "BOT": "Oras sa Bolivia", "WARST": "Oras sa Tag-init ng Kanlurang Argentina", "HADT": "Oras sa Tag-init ng Hawaii-Aleutian", "AWDT": "Daylight Time sa Kanlurang Australya", "LHDT": "Daylight Time sa Lorde Howe", "HAT": "Daylight Time sa Newfoundland", "WIT": "Oras sa Silangang Indonesia", "WITA": "Oras sa Gitnang Indonesia", "GMT": "Greenwich Mean Time", "WAST": "Oras sa Tag-init ng Kanlurang Africa", "NZST": "Standard na Oras sa New Zealand", "HNOG": "Standard na Oras sa Kanlurang Greenland", "IST": "Standard na Oras sa Bhutan"},
	}
}

// Locale returns the current translators string locale
func (fil *fil_PH) Locale() string {
	return fil.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'fil_PH'
func (fil *fil_PH) PluralsCardinal() []locales.PluralRule {
	return fil.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'fil_PH'
func (fil *fil_PH) PluralsOrdinal() []locales.PluralRule {
	return fil.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'fil_PH'
func (fil *fil_PH) PluralsRange() []locales.PluralRule {
	return fil.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'fil_PH'
func (fil *fil_PH) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)
	f := locales.F(n, v)
	iMod10 := i % 10
	fMod10 := f % 10

	if (v == 0 && (i == 1 || i == 2 || i == 3)) || (v == 0 && (iMod10 != 4 && iMod10 != 6 && iMod10 != 9)) || (v != 0 && (fMod10 != 4 && fMod10 != 6 && fMod10 != 9)) {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'fil_PH'
func (fil *fil_PH) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'fil_PH'
func (fil *fil_PH) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := fil.CardinalPluralRule(num1, v1)
	end := fil.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (fil *fil_PH) MonthAbbreviated(month time.Month) string {
	return fil.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (fil *fil_PH) MonthsAbbreviated() []string {
	return fil.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (fil *fil_PH) MonthNarrow(month time.Month) string {
	return fil.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (fil *fil_PH) MonthsNarrow() []string {
	return fil.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (fil *fil_PH) MonthWide(month time.Month) string {
	return fil.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (fil *fil_PH) MonthsWide() []string {
	return fil.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (fil *fil_PH) WeekdayAbbreviated(weekday time.Weekday) string {
	return fil.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (fil *fil_PH) WeekdaysAbbreviated() []string {
	return fil.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (fil *fil_PH) WeekdayNarrow(weekday time.Weekday) string {
	return fil.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (fil *fil_PH) WeekdaysNarrow() []string {
	return fil.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (fil *fil_PH) WeekdayShort(weekday time.Weekday) string {
	return fil.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (fil *fil_PH) WeekdaysShort() []string {
	return fil.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (fil *fil_PH) WeekdayWide(weekday time.Weekday) string {
	return fil.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (fil *fil_PH) WeekdaysWide() []string {
	return fil.daysWide
}

// Decimal returns the decimal point of number
func (fil *fil_PH) Decimal() string {
	return fil.decimal
}

// Group returns the group of number
func (fil *fil_PH) Group() string {
	return fil.group
}

// Group returns the minus sign of number
func (fil *fil_PH) Minus() string {
	return fil.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'fil_PH' and handles both Whole and Real numbers based on 'v'
func (fil *fil_PH) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, fil.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, fil.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, fil.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'fil_PH' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (fil *fil_PH) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, fil.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, fil.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, fil.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'fil_PH'
func (fil *fil_PH) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := fil.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, fil.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, fil.group[0])
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
		b = append(b, fil.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, fil.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'fil_PH'
// in accounting notation.
func (fil *fil_PH) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := fil.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, fil.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, fil.group[0])
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

		b = append(b, fil.currencyNegativePrefix[0])

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
			b = append(b, fil.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, fil.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'fil_PH'
func (fil *fil_PH) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'fil_PH'
func (fil *fil_PH) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, fil.monthsAbbreviated[t.Month()]...)
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

// FmtDateLong returns the long date representation of 't' for 'fil_PH'
func (fil *fil_PH) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, fil.monthsWide[t.Month()]...)
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

// FmtDateFull returns the full date representation of 't' for 'fil_PH'
func (fil *fil_PH) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, fil.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, fil.monthsWide[t.Month()]...)
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

// FmtTimeShort returns the short time representation of 't' for 'fil_PH'
func (fil *fil_PH) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, fil.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, fil.periodsAbbreviated[0]...)
	} else {
		b = append(b, fil.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'fil_PH'
func (fil *fil_PH) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, fil.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, fil.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, fil.periodsAbbreviated[0]...)
	} else {
		b = append(b, fil.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'fil_PH'
func (fil *fil_PH) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, fil.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, fil.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, fil.periodsAbbreviated[0]...)
	} else {
		b = append(b, fil.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'fil_PH'
func (fil *fil_PH) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, fil.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, fil.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, fil.periodsAbbreviated[0]...)
	} else {
		b = append(b, fil.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := fil.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
