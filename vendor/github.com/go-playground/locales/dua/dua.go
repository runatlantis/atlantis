package dua

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type dua struct {
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

// New returns a new instance of translator for the 'dua' locale
func New() locales.Translator {
	return &dua{
		locale:                 "dua",
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
		timezones:              map[string]string{"WITA": "WITA", "HEPM": "HEPM", "SAST": "SAST", "WAST": "WAST", "EDT": "EDT", "MESZ": "MESZ", "UYT": "UYT", "AST": "AST", "BOT": "BOT", "MEZ": "MEZ", "COT": "COT", "PST": "PST", "IST": "IST", "CLST": "CLST", "HKT": "HKT", "LHST": "LHST", "VET": "VET", "HENOMX": "HENOMX", "MST": "MST", "MDT": "MDT", "BT": "BT", "HEOG": "HEOG", "HEPMX": "HEPMX", "AEDT": "AEDT", "HECU": "HECU", "WESZ": "WESZ", "ACWST": "ACWST", "TMST": "TMST", "COST": "COST", "OEZ": "OEZ", "GMT": "GMT", "ChST": "ChST", "JDT": "JDT", "ACDT": "ACDT", "HKST": "HKST", "WARST": "WARST", "HNT": "HNT", "HAT": "HAT", "PDT": "PDT", "AWST": "AWST", "NZST": "NZST", "HNOG": "HNOG", "AEST": "AEST", "HADT": "HADT", "UYST": "UYST", "HEEG": "HEEG", "HNPM": "HNPM", "ART": "ART", "TMT": "TMT", "EAT": "EAT", "WART": "WART", "CDT": "CDT", "CLT": "CLT", "GYT": "GYT", "CHAST": "CHAST", "HNCU": "HNCU", "AWDT": "AWDT", "NZDT": "NZDT", "HNEG": "HNEG", "JST": "JST", "∅∅∅": "∅∅∅", "EST": "EST", "ACWDT": "ACWDT", "HNNOMX": "HNNOMX", "CAT": "CAT", "HAST": "HAST", "CHADT": "CHADT", "SGT": "SGT", "ECT": "ECT", "HNPMX": "HNPMX", "WEZ": "WEZ", "ACST": "ACST", "WIT": "WIT", "WAT": "WAT", "MYT": "MYT", "AKST": "AKST", "GFT": "GFT", "AKDT": "AKDT", "OESZ": "OESZ", "ARST": "ARST", "CST": "CST", "ADT": "ADT", "WIB": "WIB", "LHDT": "LHDT", "SRT": "SRT"},
	}
}

// Locale returns the current translators string locale
func (dua *dua) Locale() string {
	return dua.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'dua'
func (dua *dua) PluralsCardinal() []locales.PluralRule {
	return dua.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'dua'
func (dua *dua) PluralsOrdinal() []locales.PluralRule {
	return dua.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'dua'
func (dua *dua) PluralsRange() []locales.PluralRule {
	return dua.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'dua'
func (dua *dua) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'dua'
func (dua *dua) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'dua'
func (dua *dua) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (dua *dua) MonthAbbreviated(month time.Month) string {
	return dua.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (dua *dua) MonthsAbbreviated() []string {
	return dua.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (dua *dua) MonthNarrow(month time.Month) string {
	return dua.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (dua *dua) MonthsNarrow() []string {
	return dua.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (dua *dua) MonthWide(month time.Month) string {
	return dua.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (dua *dua) MonthsWide() []string {
	return dua.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (dua *dua) WeekdayAbbreviated(weekday time.Weekday) string {
	return dua.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (dua *dua) WeekdaysAbbreviated() []string {
	return dua.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (dua *dua) WeekdayNarrow(weekday time.Weekday) string {
	return dua.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (dua *dua) WeekdaysNarrow() []string {
	return dua.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (dua *dua) WeekdayShort(weekday time.Weekday) string {
	return dua.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (dua *dua) WeekdaysShort() []string {
	return dua.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (dua *dua) WeekdayWide(weekday time.Weekday) string {
	return dua.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (dua *dua) WeekdaysWide() []string {
	return dua.daysWide
}

// Decimal returns the decimal point of number
func (dua *dua) Decimal() string {
	return dua.decimal
}

// Group returns the group of number
func (dua *dua) Group() string {
	return dua.group
}

// Group returns the minus sign of number
func (dua *dua) Minus() string {
	return dua.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'dua' and handles both Whole and Real numbers based on 'v'
func (dua *dua) FmtNumber(num float64, v uint64) string {

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

// FmtPercent returns 'num' with digits/precision of 'v' for 'dua' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (dua *dua) FmtPercent(num float64, v uint64) string {
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

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'dua'
func (dua *dua) FmtCurrency(num float64, v uint64, currency currency.Type) string {

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

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'dua'
// in accounting notation.
func (dua *dua) FmtAccounting(num float64, v uint64, currency currency.Type) string {

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

// FmtDateShort returns the short date representation of 't' for 'dua'
func (dua *dua) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'dua'
func (dua *dua) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'dua'
func (dua *dua) FmtDateLong(t time.Time) string {

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

// FmtDateFull returns the full date representation of 't' for 'dua'
func (dua *dua) FmtDateFull(t time.Time) string {

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

// FmtTimeShort returns the short time representation of 't' for 'dua'
func (dua *dua) FmtTimeShort(t time.Time) string {

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

// FmtTimeMedium returns the medium time representation of 't' for 'dua'
func (dua *dua) FmtTimeMedium(t time.Time) string {

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

// FmtTimeLong returns the long time representation of 't' for 'dua'
func (dua *dua) FmtTimeLong(t time.Time) string {

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

// FmtTimeFull returns the full time representation of 't' for 'dua'
func (dua *dua) FmtTimeFull(t time.Time) string {

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
