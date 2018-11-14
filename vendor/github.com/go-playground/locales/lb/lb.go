package lb

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type lb struct {
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

// New returns a new instance of translator for the 'lb' locale
func New() locales.Translator {
	return &lb{
		locale:                 "lb",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ",",
		group:                  ".",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "öS", "AU$", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "₪", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "₩", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZ$", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "฿", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "NT$", "TZS", "UAH", "UAK", "UGS", "UGX", "$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "₫", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "EC$", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		percentSuffix:          " ",
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "Jan.", "Feb.", "Mäe.", "Abr.", "Mee", "Juni", "Juli", "Aug.", "Sep.", "Okt.", "Nov.", "Dez."},
		monthsNarrow:           []string{"", "J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"},
		monthsWide:             []string{"", "Januar", "Februar", "Mäerz", "Abrëll", "Mee", "Juni", "Juli", "August", "September", "Oktober", "November", "Dezember"},
		daysAbbreviated:        []string{"Son.", "Méi.", "Dën.", "Mët.", "Don.", "Fre.", "Sam."},
		daysNarrow:             []string{"S", "M", "D", "M", "D", "F", "S"},
		daysShort:              []string{"So.", "Mé.", "Dë.", "Më.", "Do.", "Fr.", "Sa."},
		daysWide:               []string{"Sonndeg", "Méindeg", "Dënschdeg", "Mëttwoch", "Donneschdeg", "Freideg", "Samschdeg"},
		periodsAbbreviated:     []string{"moies", "nomëttes"},
		periodsNarrow:          []string{"mo.", "nomë."},
		periodsWide:            []string{"moies", "nomëttes"},
		erasAbbreviated:        []string{"v. Chr.", "n. Chr."},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"v. Chr.", "n. Chr."},
		timezones:              map[string]string{"GYT": "Guyana-Zäit", "ACWST": "Zentral-/Westaustralesch Normalzäit", "TMT": "Turkmenistan-Normalzäit", "UYT": "Uruguyanesch Normalzäit", "HNCU": "Kubanesch Normalzäit", "HECU": "Kubanesch Summerzäit", "SAST": "Südafrikanesch Zäit", "ART": "Argentinesch Normalzäit", "CHAST": "Chatham-Normalzäit", "WIB": "Westindonesesch Zäit", "WESZ": "Westeuropäesch Summerzäit", "ECT": "Ecuadorianesch Zäit", "HEOG": "Westgrönland-Summerzäit", "HNPM": "Saint-Pierre-a-Miquelon-Normalzäit", "HNPMX": "Mexikanesch Pazifik-Normalzäit", "WAST": "Westafrikanesch Summerzäit", "∅∅∅": "Acre-Summerzäit", "BT": "Bhutan-Zäit", "BOT": "Bolivianesch Zäit", "EDT": "Nordamerikanesch Ostküsten-Summerzäit", "ACWDT": "Zentral-/Westaustralesch Summerzäit", "HEEG": "Ostgrönland-Summerzäit", "IST": "Indesch Zäit", "MDT": "Rocky-Mountain-Summerzäit", "ARST": "Argentinesch Summerzäit", "PST": "Nordamerikanesch Westküsten-Normalzäit", "HENOMX": "Nordwest-Mexiko-Summerzäit", "HNOG": "Westgrönland-Normalzäit", "ACST": "Zentralaustralesch Normalzäit", "ACDT": "Zentralaustralesch Summerzäit", "MEZ": "Mëtteleuropäesch Normalzäit", "MESZ": "Mëtteleuropäesch Summerzäit", "VET": "Venezuela-Zäit", "EAT": "Ostafrikanesch Zäit", "GFT": "Franséisch-Guayane-Zäit", "AWST": "Westaustralesch Normalzäit", "ChST": "Chamorro-Zäit", "HEPMX": "Mexikanesch Pazifik-Summerzäit", "HNNOMX": "Nordwest-Mexiko-Normalzäit", "HKT": "Hong-Kong-Normalzäit", "WIT": "Ostindonesesch Zäit", "WAT": "Westafrikanesch Normalzäit", "SRT": "Suriname-Zäit", "AKST": "Alaska-Normalzäit", "TMST": "Turkmenistan-Summerzäit", "CAT": "Zentralafrikanesch Zäit", "OESZ": "Osteuropäesch Summerzäit", "COST": "Kolumbianesch Summerzäit", "NZDT": "Neiséiland-Summerzäit", "JST": "Japanesch Normalzäit", "AKDT": "Alaska-Summerzäit", "HNT": "Neifundland-Normalzäit", "CST": "Nordamerikanesch Inland-Normalzäit", "AWDT": "Westaustralesch Summerzäit", "AEST": "Ostaustralesch Normalzäit", "HEPM": "Saint-Pierre-a-Miquelon-Summerzäit", "OEZ": "Osteuropäesch Normalzäit", "CHADT": "Chatham-Summerzäit", "CDT": "Nordamerikanesch Inland-Summerzäit", "PDT": "Nordamerikanesch Westküsten-Summerzäit", "AEDT": "Ostaustralesch Summerzäit", "ADT": "Atlantik-Summerzäit", "JDT": "Japanesch Summerzäit", "NZST": "Neiséiland-Normalzäit", "MYT": "Malaysesch Zäit", "HNEG": "Ostgrönland-Normalzäit", "LHDT": "Lord-Howe-Summerzäit", "AST": "Atlantik-Normalzäit", "SGT": "Singapur-Standardzäit", "LHST": "Lord-Howe-Normalzäit", "WART": "Westargentinesch Normalzäit", "WARST": "Westargentinesch Summerzäit", "HAST": "Hawaii-Aleuten-Normalzäit", "GMT": "Mëttler Greenwich-Zäit", "WEZ": "Westeuropäesch Normalzäit", "WITA": "Zentralindonesesch Zäit", "HADT": "Hawaii-Aleuten-Summerzäit", "MST": "Rocky-Mountain-Normalzäit", "HKST": "Hong-Kong-Summerzäit", "HAT": "Neifundland-Summerzäit", "CLT": "Chilenesch Normalzäit", "CLST": "Chilenesch Summerzäit", "COT": "Kolumbianesch Normalzäit", "UYST": "Uruguayanesch Summerzäit", "EST": "Nordamerikanesch Ostküsten-Normalzäit"},
	}
}

// Locale returns the current translators string locale
func (lb *lb) Locale() string {
	return lb.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'lb'
func (lb *lb) PluralsCardinal() []locales.PluralRule {
	return lb.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'lb'
func (lb *lb) PluralsOrdinal() []locales.PluralRule {
	return lb.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'lb'
func (lb *lb) PluralsRange() []locales.PluralRule {
	return lb.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'lb'
func (lb *lb) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'lb'
func (lb *lb) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'lb'
func (lb *lb) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (lb *lb) MonthAbbreviated(month time.Month) string {
	return lb.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (lb *lb) MonthsAbbreviated() []string {
	return lb.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (lb *lb) MonthNarrow(month time.Month) string {
	return lb.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (lb *lb) MonthsNarrow() []string {
	return lb.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (lb *lb) MonthWide(month time.Month) string {
	return lb.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (lb *lb) MonthsWide() []string {
	return lb.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (lb *lb) WeekdayAbbreviated(weekday time.Weekday) string {
	return lb.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (lb *lb) WeekdaysAbbreviated() []string {
	return lb.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (lb *lb) WeekdayNarrow(weekday time.Weekday) string {
	return lb.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (lb *lb) WeekdaysNarrow() []string {
	return lb.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (lb *lb) WeekdayShort(weekday time.Weekday) string {
	return lb.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (lb *lb) WeekdaysShort() []string {
	return lb.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (lb *lb) WeekdayWide(weekday time.Weekday) string {
	return lb.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (lb *lb) WeekdaysWide() []string {
	return lb.daysWide
}

// Decimal returns the decimal point of number
func (lb *lb) Decimal() string {
	return lb.decimal
}

// Group returns the group of number
func (lb *lb) Group() string {
	return lb.group
}

// Group returns the minus sign of number
func (lb *lb) Minus() string {
	return lb.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'lb' and handles both Whole and Real numbers based on 'v'
func (lb *lb) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, lb.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, lb.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, lb.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'lb' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (lb *lb) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 5
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, lb.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, lb.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, lb.percentSuffix...)

	b = append(b, lb.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'lb'
func (lb *lb) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := lb.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, lb.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, lb.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, lb.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, lb.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, lb.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'lb'
// in accounting notation.
func (lb *lb) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := lb.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, lb.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, lb.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, lb.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, lb.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, lb.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, lb.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'lb'
func (lb *lb) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'lb'
func (lb *lb) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, lb.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'lb'
func (lb *lb) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, lb.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'lb'
func (lb *lb) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, lb.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, lb.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'lb'
func (lb *lb) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, lb.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'lb'
func (lb *lb) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, lb.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, lb.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'lb'
func (lb *lb) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, lb.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, lb.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'lb'
func (lb *lb) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, lb.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, lb.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := lb.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
