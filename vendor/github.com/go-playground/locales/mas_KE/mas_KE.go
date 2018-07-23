package mas_KE

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type mas_KE struct {
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

// New returns a new instance of translator for the 'mas_KE' locale
func New() locales.Translator {
	return &mas_KE{
		locale:                 "mas_KE",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: ")",
		monthsAbbreviated:      []string{"", "Dal", "Ará", "Ɔɛn", "Doy", "Lép", "Rok", "Sás", "Bɔ́r", "Kús", "Gís", "Shʉ́", "Ntʉ́"},
		monthsWide:             []string{"", "Oladalʉ́", "Arát", "Ɔɛnɨ́ɔɨŋɔk", "Olodoyíóríê inkókúâ", "Oloilépūnyīē inkókúâ", "Kújúɔrɔk", "Mórusásin", "Ɔlɔ́ɨ́bɔ́rárɛ", "Kúshîn", "Olgísan", "Pʉshʉ́ka", "Ntʉ́ŋʉ́s"},
		daysAbbreviated:        []string{"Jpi", "Jtt", "Jnn", "Jtn", "Alh", "Iju", "Jmo"},
		daysNarrow:             []string{"2", "3", "4", "5", "6", "7", "1"},
		daysWide:               []string{"Jumapílí", "Jumatátu", "Jumane", "Jumatánɔ", "Alaámisi", "Jumáa", "Jumamósi"},
		periodsAbbreviated:     []string{"Ɛnkakɛnyá", "Ɛndámâ"},
		periodsWide:            []string{"Ɛnkakɛnyá", "Ɛndámâ"},
		erasAbbreviated:        []string{"MY", "EY"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"Meínō Yɛ́sʉ", "Eínō Yɛ́sʉ"},
		timezones:              map[string]string{"HNT": "HNT", "HAST": "HAST", "HNPMX": "HNPMX", "LHST": "LHST", "WART": "WART", "IST": "IST", "HNNOMX": "HNNOMX", "COT": "COT", "PDT": "PDT", "AST": "AST", "GFT": "GFT", "UYT": "UYT", "∅∅∅": "∅∅∅", "HNOG": "HNOG", "HKT": "HKT", "ACDT": "ACDT", "HAT": "HAT", "WESZ": "WESZ", "JST": "JST", "EST": "EST", "WEZ": "WEZ", "JDT": "JDT", "WITA": "WITA", "CLT": "CLT", "AEDT": "AEDT", "WIB": "WIB", "WAT": "WAT", "HEEG": "HEEG", "HKST": "HKST", "VET": "VET", "EAT": "EAT", "HEPMX": "HEPMX", "PST": "PST", "AKST": "AKST", "HEOG": "HEOG", "TMT": "TMT", "TMST": "TMST", "CHAST": "CHAST", "CST": "CST", "MDT": "MDT", "CDT": "CDT", "MYT": "MYT", "GMT": "GMT", "BOT": "BOT", "NZST": "NZST", "HNPM": "HNPM", "ARST": "ARST", "ADT": "ADT", "MST": "MST", "WAST": "WAST", "HNEG": "HNEG", "ACWDT": "ACWDT", "WIT": "WIT", "OESZ": "OESZ", "ART": "ART", "UYST": "UYST", "HNCU": "HNCU", "HECU": "HECU", "LHDT": "LHDT", "AEST": "AEST", "OEZ": "OEZ", "ACST": "ACST", "HENOMX": "HENOMX", "SRT": "SRT", "CAT": "CAT", "HADT": "HADT", "BT": "BT", "ECT": "ECT", "EDT": "EDT", "GYT": "GYT", "COST": "COST", "ChST": "ChST", "AWST": "AWST", "AWDT": "AWDT", "MEZ": "MEZ", "SGT": "SGT", "WARST": "WARST", "SAST": "SAST", "NZDT": "NZDT", "AKDT": "AKDT", "HEPM": "HEPM", "CLST": "CLST", "CHADT": "CHADT", "ACWST": "ACWST", "MESZ": "MESZ"},
	}
}

// Locale returns the current translators string locale
func (mas *mas_KE) Locale() string {
	return mas.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'mas_KE'
func (mas *mas_KE) PluralsCardinal() []locales.PluralRule {
	return mas.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'mas_KE'
func (mas *mas_KE) PluralsOrdinal() []locales.PluralRule {
	return mas.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'mas_KE'
func (mas *mas_KE) PluralsRange() []locales.PluralRule {
	return mas.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'mas_KE'
func (mas *mas_KE) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'mas_KE'
func (mas *mas_KE) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'mas_KE'
func (mas *mas_KE) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (mas *mas_KE) MonthAbbreviated(month time.Month) string {
	return mas.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (mas *mas_KE) MonthsAbbreviated() []string {
	return mas.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (mas *mas_KE) MonthNarrow(month time.Month) string {
	return mas.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (mas *mas_KE) MonthsNarrow() []string {
	return nil
}

// MonthWide returns the locales wide month given the 'month' provided
func (mas *mas_KE) MonthWide(month time.Month) string {
	return mas.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (mas *mas_KE) MonthsWide() []string {
	return mas.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (mas *mas_KE) WeekdayAbbreviated(weekday time.Weekday) string {
	return mas.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (mas *mas_KE) WeekdaysAbbreviated() []string {
	return mas.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (mas *mas_KE) WeekdayNarrow(weekday time.Weekday) string {
	return mas.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (mas *mas_KE) WeekdaysNarrow() []string {
	return mas.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (mas *mas_KE) WeekdayShort(weekday time.Weekday) string {
	return mas.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (mas *mas_KE) WeekdaysShort() []string {
	return mas.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (mas *mas_KE) WeekdayWide(weekday time.Weekday) string {
	return mas.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (mas *mas_KE) WeekdaysWide() []string {
	return mas.daysWide
}

// Decimal returns the decimal point of number
func (mas *mas_KE) Decimal() string {
	return mas.decimal
}

// Group returns the group of number
func (mas *mas_KE) Group() string {
	return mas.group
}

// Group returns the minus sign of number
func (mas *mas_KE) Minus() string {
	return mas.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'mas_KE' and handles both Whole and Real numbers based on 'v'
func (mas *mas_KE) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'mas_KE' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (mas *mas_KE) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'mas_KE'
func (mas *mas_KE) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := mas.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, mas.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, mas.group[0])
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
		b = append(b, mas.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, mas.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'mas_KE'
// in accounting notation.
func (mas *mas_KE) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := mas.currencies[currency]
	l := len(s) + len(symbol) + 2
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, mas.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, mas.group[0])
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

		b = append(b, mas.currencyNegativePrefix[0])

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
			b = append(b, mas.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, mas.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'mas_KE'
func (mas *mas_KE) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'mas_KE'
func (mas *mas_KE) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, mas.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'mas_KE'
func (mas *mas_KE) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, mas.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'mas_KE'
func (mas *mas_KE) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, mas.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, mas.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'mas_KE'
func (mas *mas_KE) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, mas.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'mas_KE'
func (mas *mas_KE) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, mas.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, mas.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'mas_KE'
func (mas *mas_KE) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, mas.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, mas.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'mas_KE'
func (mas *mas_KE) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, mas.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, mas.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := mas.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
