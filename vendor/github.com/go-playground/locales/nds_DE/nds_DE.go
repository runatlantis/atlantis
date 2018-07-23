package nds_DE

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type nds_DE struct {
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

// New returns a new instance of translator for the 'nds_DE' locale
func New() locales.Translator {
	return &nds_DE{
		locale:                 "nds_DE",
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
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
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
		timezones:              map[string]string{"BOT": "BOT", "CLST": "CLST", "HEPMX": "HEPMX", "NZST": "NZST", "CDT": "Noordamerikaansch zentraal Summertiet", "ADT": "Noordamerikaansch Atlantik-Summertiet", "WIB": "Westindoneesch Tiet", "HNEG": "HNEG", "HNOG": "HNOG", "TMT": "TMT", "HNPMX": "HNPMX", "JDT": "Japaansch Summertiet", "HKT": "HKT", "HAT": "HAT", "WITA": "Indoneesch Zentraaltiet", "JST": "Japaansch Standardtiet", "EST": "Noordamerikaansch oosten Standardtiet", "COT": "COT", "PDT": "Noordamerikaansch Pazifik-Summertiet", "GFT": "GFT", "MEZ": "Zentraaleuropääsch Standardtiet", "SRT": "SRT", "GYT": "GYT", "WARST": "WARST", "GMT": "Gröönwisch-Welttiet", "ChST": "ChST", "HNCU": "HNCU", "CST": "Noordamerikaansch zentraal Standardtiet", "AKDT": "AKDT", "ACWST": "Westzentraalaustraalsch Standardtiet", "HEPM": "HEPM", "HENOMX": "HENOMX", "WIT": "Oostindoneesch Tiet", "CHADT": "CHADT", "AEST": "Oostaustraalsch Standardtiet", "AEDT": "Oostaustraalsch Summertiet", "AKST": "AKST", "SGT": "SGT", "HEEG": "HEEG", "SAST": "Söödafrikaansch Tiet", "NZDT": "NZDT", "ART": "ART", "ARST": "ARST", "PST": "Noordamerikaansch Pazifik-Standardtiet", "COST": "COST", "UYST": "UYST", "BT": "BT", "HKST": "HKST", "IST": "Indien-Tiet", "CLT": "CLT", "EAT": "Oostafrikaansch Tiet", "HADT": "HADT", "WEZ": "Westeuropääsch Standardtiet", "MYT": "MYT", "EDT": "Noordamerikaansch oosten Summertiet", "∅∅∅": "∅∅∅", "LHDT": "LHDT", "HECU": "HECU", "HNPM": "HNPM", "HNNOMX": "HNNOMX", "OEZ": "Oosteuropääsch Standardtiet", "VET": "VET", "MST": "MST", "CAT": "Zentraalafrikaansch Tiet", "HAST": "HAST", "AWDT": "Westaustraalsch Summertiet", "ECT": "ECT", "MESZ": "Zentraaleuropääsch Summertiet", "HNT": "HNT", "WAST": "Westafrikaansch Summertiet", "WAT": "Westafrikaansch Standardtiet", "WESZ": "Westeuropääsch Summertiet", "HEOG": "HEOG", "WART": "WART", "TMST": "TMST", "CHAST": "CHAST", "AST": "Noordamerikaansch Atlantik-Standardtiet", "ACDT": "Zentraalaustraalsch Summertiet", "LHST": "LHST", "MDT": "MDT", "UYT": "UYT", "AWST": "Westaustraalsch Standardtiet", "ACST": "Zentraalaustraalsch Standardtiet", "ACWDT": "Westzentraalaustraalsch Summertiet", "OESZ": "Oosteuropääsch Summertiet"},
	}
}

// Locale returns the current translators string locale
func (nds *nds_DE) Locale() string {
	return nds.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'nds_DE'
func (nds *nds_DE) PluralsCardinal() []locales.PluralRule {
	return nds.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'nds_DE'
func (nds *nds_DE) PluralsOrdinal() []locales.PluralRule {
	return nds.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'nds_DE'
func (nds *nds_DE) PluralsRange() []locales.PluralRule {
	return nds.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'nds_DE'
func (nds *nds_DE) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'nds_DE'
func (nds *nds_DE) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'nds_DE'
func (nds *nds_DE) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (nds *nds_DE) MonthAbbreviated(month time.Month) string {
	return nds.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (nds *nds_DE) MonthsAbbreviated() []string {
	return nds.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (nds *nds_DE) MonthNarrow(month time.Month) string {
	return nds.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (nds *nds_DE) MonthsNarrow() []string {
	return nds.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (nds *nds_DE) MonthWide(month time.Month) string {
	return nds.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (nds *nds_DE) MonthsWide() []string {
	return nds.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (nds *nds_DE) WeekdayAbbreviated(weekday time.Weekday) string {
	return nds.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (nds *nds_DE) WeekdaysAbbreviated() []string {
	return nds.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (nds *nds_DE) WeekdayNarrow(weekday time.Weekday) string {
	return nds.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (nds *nds_DE) WeekdaysNarrow() []string {
	return nds.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (nds *nds_DE) WeekdayShort(weekday time.Weekday) string {
	return nds.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (nds *nds_DE) WeekdaysShort() []string {
	return nds.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (nds *nds_DE) WeekdayWide(weekday time.Weekday) string {
	return nds.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (nds *nds_DE) WeekdaysWide() []string {
	return nds.daysWide
}

// Decimal returns the decimal point of number
func (nds *nds_DE) Decimal() string {
	return nds.decimal
}

// Group returns the group of number
func (nds *nds_DE) Group() string {
	return nds.group
}

// Group returns the minus sign of number
func (nds *nds_DE) Minus() string {
	return nds.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'nds_DE' and handles both Whole and Real numbers based on 'v'
func (nds *nds_DE) FmtNumber(num float64, v uint64) string {

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

// FmtPercent returns 'num' with digits/precision of 'v' for 'nds_DE' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (nds *nds_DE) FmtPercent(num float64, v uint64) string {
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

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'nds_DE'
func (nds *nds_DE) FmtCurrency(num float64, v uint64, currency currency.Type) string {

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

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'nds_DE'
// in accounting notation.
func (nds *nds_DE) FmtAccounting(num float64, v uint64, currency currency.Type) string {

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

// FmtDateShort returns the short date representation of 't' for 'nds_DE'
func (nds *nds_DE) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'nds_DE'
func (nds *nds_DE) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'nds_DE'
func (nds *nds_DE) FmtDateLong(t time.Time) string {

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

// FmtDateFull returns the full date representation of 't' for 'nds_DE'
func (nds *nds_DE) FmtDateFull(t time.Time) string {

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

// FmtTimeShort returns the short time representation of 't' for 'nds_DE'
func (nds *nds_DE) FmtTimeShort(t time.Time) string {

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

// FmtTimeMedium returns the medium time representation of 't' for 'nds_DE'
func (nds *nds_DE) FmtTimeMedium(t time.Time) string {

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

// FmtTimeLong returns the long time representation of 't' for 'nds_DE'
func (nds *nds_DE) FmtTimeLong(t time.Time) string {

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

// FmtTimeFull returns the full time representation of 't' for 'nds_DE'
func (nds *nds_DE) FmtTimeFull(t time.Time) string {

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
