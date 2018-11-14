package bas_CM

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type bas_CM struct {
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

// New returns a new instance of translator for the 'bas_CM' locale
func New() locales.Translator {
	return &bas_CM{
		locale:                 "bas_CM",
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
		monthsAbbreviated:      []string{"", "kɔn", "mac", "mat", "mto", "mpu", "hil", "nje", "hik", "dip", "bio", "may", "liɓ"},
		monthsNarrow:           []string{"", "k", "m", "m", "m", "m", "h", "n", "h", "d", "b", "m", "l"},
		monthsWide:             []string{"", "Kɔndɔŋ", "Màcɛ̂l", "Màtùmb", "Màtop", "M̀puyɛ", "Hìlòndɛ̀", "Njèbà", "Hìkaŋ", "Dìpɔ̀s", "Bìòôm", "Màyɛsèp", "Lìbuy li ńyèe"},
		daysAbbreviated:        []string{"nɔy", "nja", "uum", "ŋge", "mbɔ", "kɔɔ", "jon"},
		daysNarrow:             []string{"n", "n", "u", "ŋ", "m", "k", "j"},
		daysWide:               []string{"ŋgwà nɔ̂y", "ŋgwà njaŋgumba", "ŋgwà ûm", "ŋgwà ŋgê", "ŋgwà mbɔk", "ŋgwà kɔɔ", "ŋgwà jôn"},
		periodsAbbreviated:     []string{"I bikɛ̂glà", "I ɓugajɔp"},
		periodsWide:            []string{"I bikɛ̂glà", "I ɓugajɔp"},
		erasAbbreviated:        []string{"b.Y.K", "m.Y.K"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"bisū bi Yesù Krǐstò", "i mbūs Yesù Krǐstò"},
		timezones:              map[string]string{"WITA": "WITA", "HAT": "HAT", "TMT": "TMT", "EAT": "EAT", "UYST": "UYST", "HECU": "HECU", "MYT": "MYT", "AKDT": "AKDT", "EST": "EST", "HNNOMX": "HNNOMX", "CLST": "CLST", "CST": "CST", "AEST": "AEST", "HNCU": "HNCU", "WESZ": "WESZ", "ACWST": "ACWST", "ACWDT": "ACWDT", "HEOG": "HEOG", "HKT": "HKT", "TMST": "TMST", "CHADT": "CHADT", "JST": "JST", "OEZ": "OEZ", "SGT": "SGT", "HNEG": "HNEG", "LHDT": "LHDT", "VET": "VET", "MDT": "MDT", "ARST": "ARST", "BOT": "BOT", "HENOMX": "HENOMX", "COT": "COT", "UYT": "UYT", "PDT": "PDT", "WEZ": "WEZ", "SAST": "SAST", "IST": "IST", "PST": "PST", "AKST": "AKST", "COST": "COST", "AWDT": "AWDT", "GFT": "GFT", "CDT": "CDT", "AST": "AST", "NZST": "NZST", "ACDT": "ACDT", "ART": "ART", "HAST": "HAST", "AEDT": "AEDT", "HNPMX": "HNPMX", "NZDT": "NZDT", "ECT": "ECT", "ACST": "ACST", "MEZ": "MEZ", "WART": "WART", "OESZ": "OESZ", "ChST": "ChST", "HNOG": "HNOG", "HKST": "HKST", "MST": "MST", "CHAST": "CHAST", "WAST": "WAST", "WAT": "WAT", "WARST": "WARST", "HEPM": "HEPM", "CLT": "CLT", "HADT": "HADT", "GYT": "GYT", "WIB": "WIB", "MESZ": "MESZ", "HNT": "HNT", "SRT": "SRT", "HEPMX": "HEPMX", "HNPM": "HNPM", "CAT": "CAT", "ADT": "ADT", "JDT": "JDT", "BT": "BT", "AWST": "AWST", "HEEG": "HEEG", "EDT": "EDT", "∅∅∅": "∅∅∅", "LHST": "LHST", "WIT": "WIT", "GMT": "GMT"},
	}
}

// Locale returns the current translators string locale
func (bas *bas_CM) Locale() string {
	return bas.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'bas_CM'
func (bas *bas_CM) PluralsCardinal() []locales.PluralRule {
	return bas.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'bas_CM'
func (bas *bas_CM) PluralsOrdinal() []locales.PluralRule {
	return bas.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'bas_CM'
func (bas *bas_CM) PluralsRange() []locales.PluralRule {
	return bas.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'bas_CM'
func (bas *bas_CM) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'bas_CM'
func (bas *bas_CM) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'bas_CM'
func (bas *bas_CM) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (bas *bas_CM) MonthAbbreviated(month time.Month) string {
	return bas.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (bas *bas_CM) MonthsAbbreviated() []string {
	return bas.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (bas *bas_CM) MonthNarrow(month time.Month) string {
	return bas.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (bas *bas_CM) MonthsNarrow() []string {
	return bas.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (bas *bas_CM) MonthWide(month time.Month) string {
	return bas.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (bas *bas_CM) MonthsWide() []string {
	return bas.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (bas *bas_CM) WeekdayAbbreviated(weekday time.Weekday) string {
	return bas.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (bas *bas_CM) WeekdaysAbbreviated() []string {
	return bas.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (bas *bas_CM) WeekdayNarrow(weekday time.Weekday) string {
	return bas.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (bas *bas_CM) WeekdaysNarrow() []string {
	return bas.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (bas *bas_CM) WeekdayShort(weekday time.Weekday) string {
	return bas.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (bas *bas_CM) WeekdaysShort() []string {
	return bas.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (bas *bas_CM) WeekdayWide(weekday time.Weekday) string {
	return bas.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (bas *bas_CM) WeekdaysWide() []string {
	return bas.daysWide
}

// Decimal returns the decimal point of number
func (bas *bas_CM) Decimal() string {
	return bas.decimal
}

// Group returns the group of number
func (bas *bas_CM) Group() string {
	return bas.group
}

// Group returns the minus sign of number
func (bas *bas_CM) Minus() string {
	return bas.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'bas_CM' and handles both Whole and Real numbers based on 'v'
func (bas *bas_CM) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 1 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, bas.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(bas.group) - 1; j >= 0; j-- {
					b = append(b, bas.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, bas.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'bas_CM' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (bas *bas_CM) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, bas.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, bas.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, bas.percentSuffix...)

	b = append(b, bas.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'bas_CM'
func (bas *bas_CM) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := bas.currencies[currency]
	l := len(s) + len(symbol) + 3 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, bas.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(bas.group) - 1; j >= 0; j-- {
					b = append(b, bas.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, bas.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, bas.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, bas.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'bas_CM'
// in accounting notation.
func (bas *bas_CM) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := bas.currencies[currency]
	l := len(s) + len(symbol) + 3 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, bas.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(bas.group) - 1; j >= 0; j-- {
					b = append(b, bas.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, bas.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, bas.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, bas.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, bas.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'bas_CM'
func (bas *bas_CM) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'bas_CM'
func (bas *bas_CM) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, bas.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'bas_CM'
func (bas *bas_CM) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, bas.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'bas_CM'
func (bas *bas_CM) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, bas.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, bas.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'bas_CM'
func (bas *bas_CM) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, bas.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'bas_CM'
func (bas *bas_CM) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, bas.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, bas.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'bas_CM'
func (bas *bas_CM) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, bas.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, bas.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'bas_CM'
func (bas *bas_CM) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, bas.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, bas.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := bas.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
