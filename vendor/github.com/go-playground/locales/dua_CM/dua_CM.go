package dua_CM

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type dua_CM struct {
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

// New returns a new instance of translator for the 'dua_CM' locale
func New() locales.Translator {
	return &dua_CM{
		locale:                 "dua_CM",
		pluralsCardinal:        nil,
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ",",
		group:                  " ",
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		percentSuffix:          " ",
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "di", "ŋgɔn", "sɔŋ", "diɓ", "emi", "esɔ", "mad", "diŋ", "nyɛt", "may", "tin", "elá"},
		monthsNarrow:           []string{"", "d", "ŋ", "s", "d", "e", "e", "m", "d", "n", "m", "t", "e"},
		monthsWide:             []string{"", "dimɔ́di", "ŋgɔndɛ", "sɔŋɛ", "diɓáɓá", "emiasele", "esɔpɛsɔpɛ", "madiɓɛ́díɓɛ́", "diŋgindi", "nyɛtɛki", "mayésɛ́", "tiníní", "eláŋgɛ́"},
		daysAbbreviated:        []string{"ét", "mɔ́s", "kwa", "muk", "ŋgi", "ɗón", "esa"},
		daysNarrow:             []string{"e", "m", "k", "m", "ŋ", "ɗ", "e"},
		daysWide:               []string{"éti", "mɔ́sú", "kwasú", "mukɔ́sú", "ŋgisú", "ɗónɛsú", "esaɓasú"},
		periodsAbbreviated:     []string{"idiɓa", "ebyámu"},
		periodsWide:            []string{"idiɓa", "ebyámu"},
		erasAbbreviated:        []string{"ɓ.Ys", "mb.Ys"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"ɓoso ɓwá yáɓe lá", "mbúsa kwédi a Yés"},
		timezones:              map[string]string{"MST": "MST", "EAT": "EAT", "EST": "EST", "HEOG": "HEOG", "WARST": "WARST", "IST": "IST", "MEZ": "MEZ", "HENOMX": "HENOMX", "CLT": "CLT", "ART": "ART", "CDT": "CDT", "HEPMX": "HEPMX", "HEEG": "HEEG", "MDT": "MDT", "ARST": "ARST", "BT": "BT", "SGT": "SGT", "ACDT": "ACDT", "HNT": "HNT", "CHAST": "CHAST", "AWDT": "AWDT", "ADT": "ADT", "WAT": "WAT", "JDT": "JDT", "WART": "WART", "HADT": "HADT", "AWST": "AWST", "HKST": "HKST", "CLST": "CLST", "∅∅∅": "∅∅∅", "GYT": "GYT", "BOT": "BOT", "AKDT": "AKDT", "HNPM": "HNPM", "CST": "CST", "AKST": "AKST", "ACWST": "ACWST", "ACWDT": "ACWDT", "LHDT": "LHDT", "HAT": "HAT", "CAT": "CAT", "OEZ": "OEZ", "CHADT": "CHADT", "WITA": "WITA", "ACST": "ACST", "HNEG": "HNEG", "SRT": "SRT", "WIT": "WIT", "TMST": "TMST", "HAST": "HAST", "UYST": "UYST", "ECT": "ECT", "HNNOMX": "HNNOMX", "COT": "COT", "UYT": "UYT", "WEZ": "WEZ", "GFT": "GFT", "JST": "JST", "HNCU": "HNCU", "WAST": "WAST", "NZDT": "NZDT", "HNOG": "HNOG", "VET": "VET", "HNPMX": "HNPMX", "MYT": "MYT", "HEPM": "HEPM", "TMT": "TMT", "COST": "COST", "PDT": "PDT", "AEST": "AEST", "AEDT": "AEDT", "EDT": "EDT", "LHST": "LHST", "GMT": "GMT", "ChST": "ChST", "AST": "AST", "SAST": "SAST", "WESZ": "WESZ", "WIB": "WIB", "OESZ": "OESZ", "HECU": "HECU", "PST": "PST", "NZST": "NZST", "MESZ": "MESZ", "HKT": "HKT"},
	}
}

// Locale returns the current translators string locale
func (dua *dua_CM) Locale() string {
	return dua.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'dua_CM'
func (dua *dua_CM) PluralsCardinal() []locales.PluralRule {
	return dua.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'dua_CM'
func (dua *dua_CM) PluralsOrdinal() []locales.PluralRule {
	return dua.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'dua_CM'
func (dua *dua_CM) PluralsRange() []locales.PluralRule {
	return dua.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'dua_CM'
func (dua *dua_CM) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'dua_CM'
func (dua *dua_CM) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'dua_CM'
func (dua *dua_CM) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (dua *dua_CM) MonthAbbreviated(month time.Month) string {
	return dua.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (dua *dua_CM) MonthsAbbreviated() []string {
	return dua.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (dua *dua_CM) MonthNarrow(month time.Month) string {
	return dua.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (dua *dua_CM) MonthsNarrow() []string {
	return dua.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (dua *dua_CM) MonthWide(month time.Month) string {
	return dua.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (dua *dua_CM) MonthsWide() []string {
	return dua.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (dua *dua_CM) WeekdayAbbreviated(weekday time.Weekday) string {
	return dua.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (dua *dua_CM) WeekdaysAbbreviated() []string {
	return dua.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (dua *dua_CM) WeekdayNarrow(weekday time.Weekday) string {
	return dua.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (dua *dua_CM) WeekdaysNarrow() []string {
	return dua.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (dua *dua_CM) WeekdayShort(weekday time.Weekday) string {
	return dua.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (dua *dua_CM) WeekdaysShort() []string {
	return dua.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (dua *dua_CM) WeekdayWide(weekday time.Weekday) string {
	return dua.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (dua *dua_CM) WeekdaysWide() []string {
	return dua.daysWide
}

// Decimal returns the decimal point of number
func (dua *dua_CM) Decimal() string {
	return dua.decimal
}

// Group returns the group of number
func (dua *dua_CM) Group() string {
	return dua.group
}

// Group returns the minus sign of number
func (dua *dua_CM) Minus() string {
	return dua.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'dua_CM' and handles both Whole and Real numbers based on 'v'
func (dua *dua_CM) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 1 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, dua.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(dua.group) - 1; j >= 0; j-- {
					b = append(b, dua.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, dua.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'dua_CM' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (dua *dua_CM) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, dua.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, dua.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, dua.percentSuffix...)

	b = append(b, dua.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'dua_CM'
func (dua *dua_CM) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := dua.currencies[currency]
	l := len(s) + len(symbol) + 3 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, dua.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(dua.group) - 1; j >= 0; j-- {
					b = append(b, dua.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, dua.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, dua.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, dua.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'dua_CM'
// in accounting notation.
func (dua *dua_CM) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := dua.currencies[currency]
	l := len(s) + len(symbol) + 3 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, dua.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(dua.group) - 1; j >= 0; j-- {
					b = append(b, dua.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, dua.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, dua.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, dua.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, dua.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'dua_CM'
func (dua *dua_CM) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2f}...)
	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0x2f}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'dua_CM'
func (dua *dua_CM) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, dua.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'dua_CM'
func (dua *dua_CM) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, dua.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'dua_CM'
func (dua *dua_CM) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, dua.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, dua.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'dua_CM'
func (dua *dua_CM) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, dua.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'dua_CM'
func (dua *dua_CM) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, dua.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, dua.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'dua_CM'
func (dua *dua_CM) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, dua.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, dua.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'dua_CM'
func (dua *dua_CM) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, dua.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, dua.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := dua.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
