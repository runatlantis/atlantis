package teo

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type teo struct {
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

// New returns a new instance of translator for the 'teo' locale
func New() locales.Translator {
	return &teo{
		locale:                 "teo",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "USh", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: ")",
		monthsAbbreviated:      []string{"", "Rar", "Muk", "Kwa", "Dun", "Mar", "Mod", "Jol", "Ped", "Sok", "Tib", "Lab", "Poo"},
		monthsNarrow:           []string{"", "R", "M", "K", "D", "M", "M", "J", "P", "S", "T", "L", "P"},
		monthsWide:             []string{"", "Orara", "Omuk", "Okwamg’", "Odung’el", "Omaruk", "Omodok’king’ol", "Ojola", "Opedel", "Osokosokoma", "Otibar", "Olabor", "Opoo"},
		daysAbbreviated:        []string{"Jum", "Bar", "Aar", "Uni", "Ung", "Kan", "Sab"},
		daysNarrow:             []string{"J", "B", "A", "U", "U", "K", "S"},
		daysWide:               []string{"Nakaejuma", "Nakaebarasa", "Nakaare", "Nakauni", "Nakaung’on", "Nakakany", "Nakasabiti"},
		periodsAbbreviated:     []string{"Taparachu", "Ebongi"},
		periodsWide:            []string{"Taparachu", "Ebongi"},
		erasAbbreviated:        []string{"KK", "BK"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"Kabla ya Christo", "Baada ya Christo"},
		timezones:              map[string]string{"HNCU": "HNCU", "WARST": "WARST", "HAT": "HAT", "MST": "MST", "CHAST": "CHAST", "WAST": "WAST", "AKST": "AKST", "MESZ": "MESZ", "LHST": "LHST", "VET": "VET", "OESZ": "OESZ", "AST": "AST", "WESZ": "WESZ", "JDT": "JDT", "NZST": "NZST", "SRT": "SRT", "CAT": "CAT", "EAT": "EAT", "CST": "CST", "UYST": "UYST", "GYT": "GYT", "ChST": "ChST", "AWDT": "AWDT", "HNEG": "HNEG", "MDT": "MDT", "CLT": "CLT", "TMST": "TMST", "HKT": "HKT", "HKST": "HKST", "NZDT": "NZDT", "SGT": "SGT", "COT": "COT", "UYT": "UYT", "GMT": "GMT", "MYT": "MYT", "ACWST": "ACWST", "EST": "EST", "HEPM": "HEPM", "HENOMX": "HENOMX", "BOT": "BOT", "ACWDT": "ACWDT", "WART": "WART", "CDT": "CDT", "AEST": "AEST", "WEZ": "WEZ", "ECT": "ECT", "HEEG": "HEEG", "HEOG": "HEOG", "WITA": "WITA", "PST": "PST", "BT": "BT", "∅∅∅": "∅∅∅", "CHADT": "CHADT", "AWST": "AWST", "JST": "JST", "EDT": "EDT", "AEDT": "AEDT", "SAST": "SAST", "HNPM": "HNPM", "TMT": "TMT", "ART": "ART", "COST": "COST", "LHDT": "LHDT", "HNNOMX": "HNNOMX", "ADT": "ADT", "WAT": "WAT", "ACST": "ACST", "MEZ": "MEZ", "PDT": "PDT", "GFT": "GFT", "ACDT": "ACDT", "HAST": "HAST", "HADT": "HADT", "WIB": "WIB", "HEPMX": "HEPMX", "HNT": "HNT", "WIT": "WIT", "OEZ": "OEZ", "HECU": "HECU", "HNPMX": "HNPMX", "AKDT": "AKDT", "HNOG": "HNOG", "IST": "IST", "CLST": "CLST", "ARST": "ARST"},
	}
}

// Locale returns the current translators string locale
func (teo *teo) Locale() string {
	return teo.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'teo'
func (teo *teo) PluralsCardinal() []locales.PluralRule {
	return teo.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'teo'
func (teo *teo) PluralsOrdinal() []locales.PluralRule {
	return teo.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'teo'
func (teo *teo) PluralsRange() []locales.PluralRule {
	return teo.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'teo'
func (teo *teo) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'teo'
func (teo *teo) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'teo'
func (teo *teo) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (teo *teo) MonthAbbreviated(month time.Month) string {
	return teo.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (teo *teo) MonthsAbbreviated() []string {
	return teo.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (teo *teo) MonthNarrow(month time.Month) string {
	return teo.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (teo *teo) MonthsNarrow() []string {
	return teo.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (teo *teo) MonthWide(month time.Month) string {
	return teo.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (teo *teo) MonthsWide() []string {
	return teo.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (teo *teo) WeekdayAbbreviated(weekday time.Weekday) string {
	return teo.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (teo *teo) WeekdaysAbbreviated() []string {
	return teo.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (teo *teo) WeekdayNarrow(weekday time.Weekday) string {
	return teo.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (teo *teo) WeekdaysNarrow() []string {
	return teo.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (teo *teo) WeekdayShort(weekday time.Weekday) string {
	return teo.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (teo *teo) WeekdaysShort() []string {
	return teo.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (teo *teo) WeekdayWide(weekday time.Weekday) string {
	return teo.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (teo *teo) WeekdaysWide() []string {
	return teo.daysWide
}

// Decimal returns the decimal point of number
func (teo *teo) Decimal() string {
	return teo.decimal
}

// Group returns the group of number
func (teo *teo) Group() string {
	return teo.group
}

// Group returns the minus sign of number
func (teo *teo) Minus() string {
	return teo.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'teo' and handles both Whole and Real numbers based on 'v'
func (teo *teo) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'teo' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (teo *teo) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'teo'
func (teo *teo) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := teo.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, teo.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, teo.group[0])
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
		b = append(b, teo.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, teo.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'teo'
// in accounting notation.
func (teo *teo) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := teo.currencies[currency]
	l := len(s) + len(symbol) + 2
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, teo.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, teo.group[0])
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

		b = append(b, teo.currencyNegativePrefix[0])

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
			b = append(b, teo.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, teo.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'teo'
func (teo *teo) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2f}...)

	if t.Month() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Month()), 10)

	b = append(b, []byte{0x2f}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'teo'
func (teo *teo) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, teo.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'teo'
func (teo *teo) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, teo.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'teo'
func (teo *teo) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, teo.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, teo.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'teo'
func (teo *teo) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, teo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'teo'
func (teo *teo) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, teo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, teo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'teo'
func (teo *teo) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, teo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, teo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'teo'
func (teo *teo) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, teo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, teo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := teo.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
