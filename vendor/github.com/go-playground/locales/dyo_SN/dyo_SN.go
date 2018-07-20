package dyo_SN

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type dyo_SN struct {
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

// New returns a new instance of translator for the 'dyo_SN' locale
func New() locales.Translator {
	return &dyo_SN{
		locale:                 "dyo_SN",
		pluralsCardinal:        nil,
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ",",
		group:                  " ",
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "Sa", "Fe", "Ma", "Ab", "Me", "Su", "Sú", "Ut", "Se", "Ok", "No", "De"},
		monthsNarrow:           []string{"", "S", "F", "M", "A", "M", "S", "S", "U", "S", "O", "N", "D"},
		monthsWide:             []string{"", "Sanvie", "Fébirie", "Mars", "Aburil", "Mee", "Sueŋ", "Súuyee", "Ut", "Settembar", "Oktobar", "Novembar", "Disambar"},
		daysAbbreviated:        []string{"Dim", "Ten", "Tal", "Ala", "Ara", "Arj", "Sib"},
		daysNarrow:             []string{"D", "T", "T", "A", "A", "A", "S"},
		daysWide:               []string{"Dimas", "Teneŋ", "Talata", "Alarbay", "Aramisay", "Arjuma", "Sibiti"},
		erasAbbreviated:        []string{"ArY", "AtY"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"Ariŋuu Yeesu", "Atooŋe Yeesu"},
		timezones:              map[string]string{"AWST": "AWST", "BOT": "BOT", "HNT": "HNT", "PDT": "PDT", "HAT": "HAT", "OEZ": "OEZ", "GMT": "GMT", "HEPM": "HEPM", "AWDT": "AWDT", "AEST": "AEST", "WEZ": "WEZ", "JDT": "JDT", "ACDT": "ACDT", "MEZ": "MEZ", "HNPM": "HNPM", "TMST": "TMST", "HECU": "HECU", "WIB": "WIB", "HENOMX": "HENOMX", "GYT": "GYT", "HNCU": "HNCU", "MYT": "MYT", "AKDT": "AKDT", "IST": "IST", "MST": "MST", "SGT": "SGT", "ECT": "ECT", "ACWDT": "ACWDT", "MESZ": "MESZ", "UYST": "UYST", "BT": "BT", "CLT": "CLT", "HNPMX": "HNPMX", "SAST": "SAST", "EDT": "EDT", "HKT": "HKT", "EAT": "EAT", "HAST": "HAST", "ChST": "ChST", "CHADT": "CHADT", "AST": "AST", "WAST": "WAST", "WESZ": "WESZ", "WART": "WART", "TMT": "TMT", "CAT": "CAT", "CHAST": "CHAST", "MDT": "MDT", "WAT": "WAT", "GFT": "GFT", "AKST": "AKST", "HNEG": "HNEG", "SRT": "SRT", "CLST": "CLST", "EST": "EST", "HNOG": "HNOG", "∅∅∅": "∅∅∅", "CST": "CST", "PST": "PST", "ADT": "ADT", "NZDT": "NZDT", "ACWST": "ACWST", "HEEG": "HEEG", "LHDT": "LHDT", "ART": "ART", "COST": "COST", "NZST": "NZST", "JST": "JST", "HKST": "HKST", "LHST": "LHST", "WARST": "WARST", "HNNOMX": "HNNOMX", "WIT": "WIT", "VET": "VET", "AEDT": "AEDT", "ACST": "ACST", "HEOG": "HEOG", "HADT": "HADT", "COT": "COT", "UYT": "UYT", "CDT": "CDT", "HEPMX": "HEPMX", "WITA": "WITA", "OESZ": "OESZ", "ARST": "ARST"},
	}
}

// Locale returns the current translators string locale
func (dyo *dyo_SN) Locale() string {
	return dyo.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'dyo_SN'
func (dyo *dyo_SN) PluralsCardinal() []locales.PluralRule {
	return dyo.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'dyo_SN'
func (dyo *dyo_SN) PluralsOrdinal() []locales.PluralRule {
	return dyo.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'dyo_SN'
func (dyo *dyo_SN) PluralsRange() []locales.PluralRule {
	return dyo.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'dyo_SN'
func (dyo *dyo_SN) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'dyo_SN'
func (dyo *dyo_SN) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'dyo_SN'
func (dyo *dyo_SN) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (dyo *dyo_SN) MonthAbbreviated(month time.Month) string {
	return dyo.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (dyo *dyo_SN) MonthsAbbreviated() []string {
	return dyo.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (dyo *dyo_SN) MonthNarrow(month time.Month) string {
	return dyo.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (dyo *dyo_SN) MonthsNarrow() []string {
	return dyo.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (dyo *dyo_SN) MonthWide(month time.Month) string {
	return dyo.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (dyo *dyo_SN) MonthsWide() []string {
	return dyo.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (dyo *dyo_SN) WeekdayAbbreviated(weekday time.Weekday) string {
	return dyo.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (dyo *dyo_SN) WeekdaysAbbreviated() []string {
	return dyo.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (dyo *dyo_SN) WeekdayNarrow(weekday time.Weekday) string {
	return dyo.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (dyo *dyo_SN) WeekdaysNarrow() []string {
	return dyo.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (dyo *dyo_SN) WeekdayShort(weekday time.Weekday) string {
	return dyo.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (dyo *dyo_SN) WeekdaysShort() []string {
	return dyo.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (dyo *dyo_SN) WeekdayWide(weekday time.Weekday) string {
	return dyo.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (dyo *dyo_SN) WeekdaysWide() []string {
	return dyo.daysWide
}

// Decimal returns the decimal point of number
func (dyo *dyo_SN) Decimal() string {
	return dyo.decimal
}

// Group returns the group of number
func (dyo *dyo_SN) Group() string {
	return dyo.group
}

// Group returns the minus sign of number
func (dyo *dyo_SN) Minus() string {
	return dyo.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'dyo_SN' and handles both Whole and Real numbers based on 'v'
func (dyo *dyo_SN) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 1 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, dyo.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(dyo.group) - 1; j >= 0; j-- {
					b = append(b, dyo.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, dyo.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'dyo_SN' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (dyo *dyo_SN) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 1
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, dyo.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, dyo.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, dyo.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'dyo_SN'
func (dyo *dyo_SN) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := dyo.currencies[currency]
	l := len(s) + len(symbol) + 3 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, dyo.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(dyo.group) - 1; j >= 0; j-- {
					b = append(b, dyo.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, dyo.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, dyo.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, dyo.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'dyo_SN'
// in accounting notation.
func (dyo *dyo_SN) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := dyo.currencies[currency]
	l := len(s) + len(symbol) + 3 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, dyo.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(dyo.group) - 1; j >= 0; j-- {
					b = append(b, dyo.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, dyo.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, dyo.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, dyo.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, dyo.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'dyo_SN'
func (dyo *dyo_SN) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'dyo_SN'
func (dyo *dyo_SN) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, dyo.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'dyo_SN'
func (dyo *dyo_SN) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, dyo.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'dyo_SN'
func (dyo *dyo_SN) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, dyo.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, dyo.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'dyo_SN'
func (dyo *dyo_SN) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, dyo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'dyo_SN'
func (dyo *dyo_SN) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, dyo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, dyo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'dyo_SN'
func (dyo *dyo_SN) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, dyo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, dyo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'dyo_SN'
func (dyo *dyo_SN) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, dyo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, dyo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := dyo.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
