package ln

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type ln struct {
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

// New returns a new instance of translator for the 'ln' locale
func New() locales.Translator {
	return &ln{
		locale:                 "ln",
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
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "Kz", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "A$", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "FC", "CHE", "Fr.", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZ$", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "T$", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "F CFP", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "yan", "fbl", "msi", "apl", "mai", "yun", "yul", "agt", "stb", "ɔtb", "nvb", "dsb"},
		monthsNarrow:           []string{"", "y", "f", "m", "a", "m", "y", "y", "a", "s", "ɔ", "n", "d"},
		monthsWide:             []string{"", "sánzá ya yambo", "sánzá ya míbalé", "sánzá ya mísáto", "sánzá ya mínei", "sánzá ya mítáno", "sánzá ya motóbá", "sánzá ya nsambo", "sánzá ya mwambe", "sánzá ya libwa", "sánzá ya zómi", "sánzá ya zómi na mɔ̌kɔ́", "sánzá ya zómi na míbalé"},
		daysAbbreviated:        []string{"eye", "ybo", "mbl", "mst", "min", "mtn", "mps"},
		daysNarrow:             []string{"e", "y", "m", "m", "m", "m", "p"},
		daysWide:               []string{"eyenga", "mokɔlɔ mwa yambo", "mokɔlɔ mwa míbalé", "mokɔlɔ mwa mísáto", "mokɔlɔ ya mínéi", "mokɔlɔ ya mítáno", "mpɔ́sɔ"},
		periodsAbbreviated:     []string{"ntɔ́ngɔ́", "mpókwa"},
		periodsWide:            []string{"ntɔ́ngɔ́", "mpókwa"},
		erasAbbreviated:        []string{"libóso ya", "nsima ya Y"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"Yambo ya Yézu Krís", "Nsima ya Yézu Krís"},
		timezones:              map[string]string{"CLST": "CLST", "UYST": "UYST", "WIB": "WIB", "HENOMX": "HENOMX", "SRT": "SRT", "GMT": "Ntángo ya Londoni", "WESZ": "WESZ", "JST": "JST", "HEEG": "HEEG", "AEDT": "AEDT", "SAST": "Ntángo ya Afríka ya Sidi", "MESZ": "MESZ", "LHDT": "LHDT", "WITA": "WITA", "CAT": "Ntángo ya Lubumbashi", "OEZ": "OEZ", "HAST": "HAST", "UYT": "UYT", "ADT": "ADT", "AKDT": "AKDT", "OESZ": "OESZ", "GYT": "GYT", "JDT": "JDT", "ECT": "ECT", "MEZ": "MEZ", "HKST": "HKST", "WART": "WART", "CLT": "CLT", "ARST": "ARST", "EST": "EST", "WARST": "WARST", "HAT": "HAT", "TMST": "TMST", "∅∅∅": "∅∅∅", "AWST": "AWST", "AWDT": "AWDT", "AEST": "AEST", "NZDT": "NZDT", "ACWST": "ACWST", "WIT": "WIT", "WAST": "WAST", "EDT": "EDT", "VET": "VET", "HADT": "HADT", "HECU": "HECU", "BT": "BT", "IST": "IST", "LHST": "LHST", "HKT": "HKT", "MST": "MST", "ART": "ART", "HNPMX": "HNPMX", "AST": "AST", "MYT": "MYT", "ACST": "ACST", "HNOG": "HNOG", "HNT": "HNT", "COST": "COST", "CHADT": "CHADT", "PST": "PST", "GFT": "GFT", "HNPM": "HNPM", "ChST": "ChST", "HNCU": "HNCU", "WEZ": "WEZ", "HNEG": "HNEG", "CST": "CST", "CDT": "CDT", "ACWDT": "ACWDT", "BOT": "BOT", "HNNOMX": "HNNOMX", "TMT": "TMT", "HEPMX": "HEPMX", "NZST": "NZST", "SGT": "SGT", "ACDT": "ACDT", "HEPM": "HEPM", "HEOG": "HEOG", "MDT": "MDT", "EAT": "Ntángo ya Afríka ya Ɛ́sita", "COT": "COT", "CHAST": "CHAST", "PDT": "PDT", "WAT": "WAT", "AKST": "AKST"},
	}
}

// Locale returns the current translators string locale
func (ln *ln) Locale() string {
	return ln.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'ln'
func (ln *ln) PluralsCardinal() []locales.PluralRule {
	return ln.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'ln'
func (ln *ln) PluralsOrdinal() []locales.PluralRule {
	return ln.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'ln'
func (ln *ln) PluralsRange() []locales.PluralRule {
	return ln.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'ln'
func (ln *ln) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n >= 0 && n <= 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'ln'
func (ln *ln) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'ln'
func (ln *ln) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (ln *ln) MonthAbbreviated(month time.Month) string {
	return ln.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (ln *ln) MonthsAbbreviated() []string {
	return ln.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (ln *ln) MonthNarrow(month time.Month) string {
	return ln.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (ln *ln) MonthsNarrow() []string {
	return ln.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (ln *ln) MonthWide(month time.Month) string {
	return ln.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (ln *ln) MonthsWide() []string {
	return ln.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (ln *ln) WeekdayAbbreviated(weekday time.Weekday) string {
	return ln.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (ln *ln) WeekdaysAbbreviated() []string {
	return ln.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (ln *ln) WeekdayNarrow(weekday time.Weekday) string {
	return ln.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (ln *ln) WeekdaysNarrow() []string {
	return ln.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (ln *ln) WeekdayShort(weekday time.Weekday) string {
	return ln.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (ln *ln) WeekdaysShort() []string {
	return ln.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (ln *ln) WeekdayWide(weekday time.Weekday) string {
	return ln.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (ln *ln) WeekdaysWide() []string {
	return ln.daysWide
}

// Decimal returns the decimal point of number
func (ln *ln) Decimal() string {
	return ln.decimal
}

// Group returns the group of number
func (ln *ln) Group() string {
	return ln.group
}

// Group returns the minus sign of number
func (ln *ln) Minus() string {
	return ln.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'ln' and handles both Whole and Real numbers based on 'v'
func (ln *ln) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ln.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ln.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ln.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'ln' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (ln *ln) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'ln'
func (ln *ln) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ln.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ln.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ln.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ln.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ln.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, ln.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'ln'
// in accounting notation.
func (ln *ln) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ln.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ln.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ln.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, ln.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ln.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, ln.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, ln.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'ln'
func (ln *ln) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'ln'
func (ln *ln) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ln.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'ln'
func (ln *ln) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ln.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'ln'
func (ln *ln) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, ln.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ln.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'ln'
func (ln *ln) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ln.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'ln'
func (ln *ln) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ln.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ln.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'ln'
func (ln *ln) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ln.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ln.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'ln'
func (ln *ln) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ln.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ln.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := ln.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
