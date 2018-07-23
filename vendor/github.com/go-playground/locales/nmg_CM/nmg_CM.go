package nmg_CM

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type nmg_CM struct {
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

// New returns a new instance of translator for the 'nmg_CM' locale
func New() locales.Translator {
	return &nmg_CM{
		locale:                 "nmg_CM",
		pluralsCardinal:        nil,
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ",",
		group:                  " ",
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "ng1", "ng2", "ng3", "ng4", "ng5", "ng6", "ng7", "ng8", "ng9", "ng10", "ng11", "kris"},
		monthsWide:             []string{"", "ngwɛn matáhra", "ngwɛn ńmba", "ngwɛn ńlal", "ngwɛn ńna", "ngwɛn ńtan", "ngwɛn ńtuó", "ngwɛn hɛmbuɛrí", "ngwɛn lɔmbi", "ngwɛn rɛbvuâ", "ngwɛn wum", "ngwɛn wum navǔr", "krísimin"},
		daysAbbreviated:        []string{"sɔ́n", "mɔ́n", "smb", "sml", "smn", "mbs", "sas"},
		daysNarrow:             []string{"s", "m", "s", "s", "s", "m", "s"},
		daysWide:               []string{"sɔ́ndɔ", "mɔ́ndɔ", "sɔ́ndɔ mafú mába", "sɔ́ndɔ mafú málal", "sɔ́ndɔ mafú mána", "mabágá má sukul", "sásadi"},
		periodsAbbreviated:     []string{"maná", "kugú"},
		periodsWide:            []string{"maná", "kugú"},
		erasAbbreviated:        []string{"BL", "PB"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"Bó Lahlɛ̄", "Pfiɛ Burī"},
		timezones:              map[string]string{"COT": "COT", "AWDT": "AWDT", "HKT": "HKT", "GFT": "GFT", "SGT": "SGT", "LHDT": "LHDT", "GYT": "GYT", "PDT": "PDT", "AEST": "AEST", "WIB": "WIB", "MYT": "MYT", "ACST": "ACST", "HNEG": "HNEG", "CAT": "CAT", "EAT": "EAT", "CLT": "CLT", "BT": "BT", "ECT": "ECT", "HNPM": "HNPM", "HEPM": "HEPM", "ACDT": "ACDT", "WART": "WART", "HADT": "HADT", "HNPMX": "HNPMX", "CHADT": "CHADT", "EST": "EST", "IST": "IST", "HNT": "HNT", "ADT": "ADT", "EDT": "EDT", "AKDT": "AKDT", "LHST": "LHST", "∅∅∅": "∅∅∅", "HNNOMX": "HNNOMX", "MDT": "MDT", "ChST": "ChST", "WAT": "WAT", "ACWST": "ACWST", "ACWDT": "ACWDT", "WITA": "WITA", "HENOMX": "HENOMX", "CLST": "CLST", "OESZ": "OESZ", "HECU": "HECU", "WEZ": "WEZ", "CDT": "CDT", "AWST": "AWST", "VET": "VET", "TMST": "TMST", "OEZ": "OEZ", "ART": "ART", "COST": "COST", "UYST": "UYST", "WAST": "WAST", "JDT": "JDT", "HKST": "HKST", "MST": "MST", "WIT": "WIT", "AST": "AST", "JST": "JST", "HEOG": "HEOG", "MESZ": "MESZ", "SRT": "SRT", "TMT": "TMT", "SAST": "SAST", "BOT": "BOT", "PST": "PST", "NZST": "NZST", "NZDT": "NZDT", "HNOG": "HNOG", "MEZ": "MEZ", "WESZ": "WESZ", "AKST": "AKST", "HAST": "HAST", "ARST": "ARST", "GMT": "GMT", "HNCU": "HNCU", "HEPMX": "HEPMX", "AEDT": "AEDT", "HEEG": "HEEG", "WARST": "WARST", "HAT": "HAT", "UYT": "UYT", "CHAST": "CHAST", "CST": "CST"},
	}
}

// Locale returns the current translators string locale
func (nmg *nmg_CM) Locale() string {
	return nmg.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'nmg_CM'
func (nmg *nmg_CM) PluralsCardinal() []locales.PluralRule {
	return nmg.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'nmg_CM'
func (nmg *nmg_CM) PluralsOrdinal() []locales.PluralRule {
	return nmg.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'nmg_CM'
func (nmg *nmg_CM) PluralsRange() []locales.PluralRule {
	return nmg.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'nmg_CM'
func (nmg *nmg_CM) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'nmg_CM'
func (nmg *nmg_CM) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'nmg_CM'
func (nmg *nmg_CM) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (nmg *nmg_CM) MonthAbbreviated(month time.Month) string {
	return nmg.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (nmg *nmg_CM) MonthsAbbreviated() []string {
	return nmg.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (nmg *nmg_CM) MonthNarrow(month time.Month) string {
	return nmg.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (nmg *nmg_CM) MonthsNarrow() []string {
	return nil
}

// MonthWide returns the locales wide month given the 'month' provided
func (nmg *nmg_CM) MonthWide(month time.Month) string {
	return nmg.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (nmg *nmg_CM) MonthsWide() []string {
	return nmg.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (nmg *nmg_CM) WeekdayAbbreviated(weekday time.Weekday) string {
	return nmg.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (nmg *nmg_CM) WeekdaysAbbreviated() []string {
	return nmg.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (nmg *nmg_CM) WeekdayNarrow(weekday time.Weekday) string {
	return nmg.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (nmg *nmg_CM) WeekdaysNarrow() []string {
	return nmg.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (nmg *nmg_CM) WeekdayShort(weekday time.Weekday) string {
	return nmg.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (nmg *nmg_CM) WeekdaysShort() []string {
	return nmg.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (nmg *nmg_CM) WeekdayWide(weekday time.Weekday) string {
	return nmg.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (nmg *nmg_CM) WeekdaysWide() []string {
	return nmg.daysWide
}

// Decimal returns the decimal point of number
func (nmg *nmg_CM) Decimal() string {
	return nmg.decimal
}

// Group returns the group of number
func (nmg *nmg_CM) Group() string {
	return nmg.group
}

// Group returns the minus sign of number
func (nmg *nmg_CM) Minus() string {
	return nmg.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'nmg_CM' and handles both Whole and Real numbers based on 'v'
func (nmg *nmg_CM) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 1 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, nmg.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(nmg.group) - 1; j >= 0; j-- {
					b = append(b, nmg.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, nmg.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'nmg_CM' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (nmg *nmg_CM) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 1
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, nmg.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, nmg.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, nmg.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'nmg_CM'
func (nmg *nmg_CM) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := nmg.currencies[currency]
	l := len(s) + len(symbol) + 3 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, nmg.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(nmg.group) - 1; j >= 0; j-- {
					b = append(b, nmg.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, nmg.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, nmg.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, nmg.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'nmg_CM'
// in accounting notation.
func (nmg *nmg_CM) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := nmg.currencies[currency]
	l := len(s) + len(symbol) + 3 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, nmg.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(nmg.group) - 1; j >= 0; j-- {
					b = append(b, nmg.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, nmg.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, nmg.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, nmg.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, nmg.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'nmg_CM'
func (nmg *nmg_CM) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'nmg_CM'
func (nmg *nmg_CM) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, nmg.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'nmg_CM'
func (nmg *nmg_CM) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, nmg.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'nmg_CM'
func (nmg *nmg_CM) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, nmg.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, nmg.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'nmg_CM'
func (nmg *nmg_CM) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, nmg.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'nmg_CM'
func (nmg *nmg_CM) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, nmg.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, nmg.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'nmg_CM'
func (nmg *nmg_CM) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, nmg.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, nmg.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'nmg_CM'
func (nmg *nmg_CM) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, nmg.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, nmg.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := nmg.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
