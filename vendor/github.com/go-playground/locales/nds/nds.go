package nds

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type nds struct {
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

// New returns a new instance of translator for the 'nds' locale
func New() locales.Translator {
	return &nds{
		locale:                 "nds",
		pluralsCardinal:        nil,
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ",",
		group:                  ".",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AU$", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JP¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "₩", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "฿", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "NT$", "TZS", "UAH", "UAK", "UGS", "UGX", "$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		percentSuffix:          " ",
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "Jan.", "Feb.", "März", "Apr.", "Mai", "Juni", "Juli", "Aug.", "Sep.", "Okt.", "Nov.", "Dez."},
		monthsNarrow:           []string{"", "J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"},
		monthsWide:             []string{"", "Januaar", "Februaar", "März", "April", "Mai", "Juni", "Juli", "August", "September", "Oktover", "November", "Dezember"},
		daysAbbreviated:        []string{"Sü.", "Ma.", "Di.", "Mi.", "Du.", "Fr.", "Sa."},
		daysNarrow:             []string{"S", "M", "D", "M", "D", "F", "S"},
		daysWide:               []string{"Sünndag", "Maandag", "Dingsdag", "Middeweken", "Dunnersdag", "Freedag", "Sünnavend"},
		periodsAbbreviated:     []string{"vm", "nm"},
		periodsWide:            []string{"vm", "nm"},
		erasAbbreviated:        []string{"v.Chr.", "n.Chr."},
		erasNarrow:             []string{"vC", "nC"},
		erasWide:               []string{"vör Christus", "na Christus"},
		timezones:              map[string]string{"HNEG": "HNEG", "EAT": "Oostafrikaansch Tiet", "AWST": "Westaustraalsch Standardtiet", "NZST": "NZST", "ACST": "Zentraalaustraalsch Standardtiet", "ACWDT": "Westzentraalaustraalsch Summertiet", "HEEG": "HEEG", "HNPM": "HNPM", "ARST": "ARST", "CHAST": "CHAST", "AST": "Noordamerikaansch Atlantik-Standardtiet", "EDT": "Noordamerikaansch oosten Summertiet", "HNCU": "HNCU", "AEDT": "Oostaustraalsch Summertiet", "ACWST": "Westzentraalaustraalsch Standardtiet", "HNNOMX": "HNNOMX", "WIB": "Westindoneesch Tiet", "SGT": "SGT", "HKST": "HKST", "VET": "VET", "CLST": "CLST", "HADT": "HADT", "CST": "Noordamerikaansch zentraal Standardtiet", "WAT": "Westafrikaansch Standardtiet", "GMT": "Gröönwisch-Welttiet", "WAST": "Westafrikaansch Summertiet", "MEZ": "Zentraaleuropääsch Standardtiet", "WART": "WART", "CLT": "CLT", "TMT": "TMT", "OESZ": "Oosteuropääsch Summertiet", "COT": "COT", "JDT": "Japaansch Summertiet", "EST": "Noordamerikaansch oosten Standardtiet", "WARST": "WARST", "MDT": "MDT", "WEZ": "Westeuropääsch Standardtiet", "MYT": "MYT", "ECT": "ECT", "HKT": "HKT", "LHDT": "LHDT", "HAST": "HAST", "PST": "Noordamerikaansch Pazifik-Standardtiet", "WESZ": "Westeuropääsch Summertiet", "NZDT": "NZDT", "PDT": "Noordamerikaansch Pazifik-Summertiet", "AEST": "Oostaustraalsch Standardtiet", "SAST": "Söödafrikaansch Tiet", "JST": "Japaansch Standardtiet", "MST": "MST", "WIT": "Oostindoneesch Tiet", "TMST": "TMST", "UYT": "UYT", "AKDT": "AKDT", "HENOMX": "HENOMX", "ACDT": "Zentraalaustraalsch Summertiet", "HEPM": "HEPM", "∅∅∅": "∅∅∅", "CDT": "Noordamerikaansch zentraal Summertiet", "HEPMX": "HEPMX", "GFT": "GFT", "AWDT": "Westaustraalsch Summertiet", "BT": "BT", "AKST": "AKST", "LHST": "LHST", "BOT": "BOT", "HNOG": "HNOG", "HAT": "HAT", "SRT": "SRT", "OEZ": "Oosteuropääsch Standardtiet", "CHADT": "CHADT", "ADT": "Noordamerikaansch Atlantik-Summertiet", "HNPMX": "HNPMX", "HEOG": "HEOG", "MESZ": "Zentraaleuropääsch Summertiet", "WITA": "Indoneesch Zentraaltiet", "ART": "ART", "COST": "COST", "UYST": "UYST", "HECU": "HECU", "CAT": "Zentraalafrikaansch Tiet", "IST": "Indien-Tiet", "GYT": "GYT", "ChST": "ChST", "HNT": "HNT"},
	}
}

// Locale returns the current translators string locale
func (nds *nds) Locale() string {
	return nds.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'nds'
func (nds *nds) PluralsCardinal() []locales.PluralRule {
	return nds.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'nds'
func (nds *nds) PluralsOrdinal() []locales.PluralRule {
	return nds.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'nds'
func (nds *nds) PluralsRange() []locales.PluralRule {
	return nds.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'nds'
func (nds *nds) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'nds'
func (nds *nds) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'nds'
func (nds *nds) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (nds *nds) MonthAbbreviated(month time.Month) string {
	return nds.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (nds *nds) MonthsAbbreviated() []string {
	return nds.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (nds *nds) MonthNarrow(month time.Month) string {
	return nds.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (nds *nds) MonthsNarrow() []string {
	return nds.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (nds *nds) MonthWide(month time.Month) string {
	return nds.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (nds *nds) MonthsWide() []string {
	return nds.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (nds *nds) WeekdayAbbreviated(weekday time.Weekday) string {
	return nds.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (nds *nds) WeekdaysAbbreviated() []string {
	return nds.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (nds *nds) WeekdayNarrow(weekday time.Weekday) string {
	return nds.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (nds *nds) WeekdaysNarrow() []string {
	return nds.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (nds *nds) WeekdayShort(weekday time.Weekday) string {
	return nds.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (nds *nds) WeekdaysShort() []string {
	return nds.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (nds *nds) WeekdayWide(weekday time.Weekday) string {
	return nds.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (nds *nds) WeekdaysWide() []string {
	return nds.daysWide
}

// Decimal returns the decimal point of number
func (nds *nds) Decimal() string {
	return nds.decimal
}

// Group returns the group of number
func (nds *nds) Group() string {
	return nds.group
}

// Group returns the minus sign of number
func (nds *nds) Minus() string {
	return nds.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'nds' and handles both Whole and Real numbers based on 'v'
func (nds *nds) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, nds.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, nds.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, nds.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'nds' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (nds *nds) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 5
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, nds.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, nds.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, nds.percentSuffix...)

	b = append(b, nds.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'nds'
func (nds *nds) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := nds.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, nds.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, nds.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, nds.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, nds.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, nds.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'nds'
// in accounting notation.
func (nds *nds) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := nds.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, nds.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, nds.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, nds.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, nds.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, nds.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, nds.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'nds'
func (nds *nds) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

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

// FmtDateMedium returns the medium date representation of 't' for 'nds'
func (nds *nds) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, nds.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'nds'
func (nds *nds) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, nds.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'nds'
func (nds *nds) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, nds.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20, 0x64, 0x65}...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, nds.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'nds'
func (nds *nds) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, []byte{0x4b, 0x6c}...)
	b = append(b, []byte{0x2e, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'nds'
func (nds *nds) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, []byte{0x4b, 0x6c, 0x6f, 0x63, 0x6b}...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, nds.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'nds'
func (nds *nds) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, []byte{0x4b, 0x6c, 0x6f, 0x63, 0x6b}...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, nds.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20, 0x28}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	b = append(b, []byte{0x29}...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'nds'
func (nds *nds) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, []byte{0x4b, 0x6c, 0x6f, 0x63, 0x6b}...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, nds.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20, 0x28}...)

	tz, _ := t.Zone()

	if btz, ok := nds.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	b = append(b, []byte{0x29}...)

	return string(b)
}
