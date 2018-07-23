package smn

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type smn struct {
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

// New returns a new instance of translator for the 'smn' locale
func New() locales.Translator {
	return &smn{
		locale:                 "smn",
		pluralsCardinal:        []locales.PluralRule{2, 3, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ",",
		group:                  " ",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ".",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		percentSuffix:          " ",
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "uđiv", "kuovâ", "njuhčâ", "cuáŋui", "vyesi", "kesi", "syeini", "porge", "čohčâ", "roovvâd", "skammâ", "juovlâ"},
		monthsNarrow:           []string{"", "U", "K", "NJ", "C", "V", "K", "S", "P", "Č", "R", "S", "J"},
		monthsWide:             []string{"", "uđđâivemáánu", "kuovâmáánu", "njuhčâmáánu", "cuáŋuimáánu", "vyesimáánu", "kesimáánu", "syeinimáánu", "porgemáánu", "čohčâmáánu", "roovvâdmáánu", "skammâmáánu", "juovlâmáánu"},
		daysAbbreviated:        []string{"pas", "vuo", "maj", "kos", "tuo", "vás", "láv"},
		daysNarrow:             []string{"p", "V", "M", "K", "T", "V", "L"},
		daysShort:              []string{"pa", "vu", "ma", "ko", "tu", "vá", "lá"},
		daysWide:               []string{"pasepeeivi", "vuossaargâ", "majebaargâ", "koskoho", "tuorâstuv", "vástuppeeivi", "lávurduv"},
		periodsAbbreviated:     []string{"ip.", "ep."},
		periodsNarrow:          []string{"ip.", "ep."},
		periodsWide:            []string{"ip.", "ep."},
		erasAbbreviated:        []string{"oKr.", "mKr."},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"Ovdil Kristus šoddâm", "maŋa Kristus šoddâm"},
		timezones:              map[string]string{"ADT": "ADT", "SGT": "SGT", "VET": "VET", "CHADT": "CHADT", "PDT": "PDT", "SAST": "SAST", "∅∅∅": "∅∅∅", "CLST": "CLST", "ART": "ART", "NZDT": "NZDT", "MEZ": "MEZ", "TMST": "TMST", "UYST": "UYST", "AEST": "AEST", "WESZ": "WESZ", "MYT": "MYT", "HNOG": "HNOG", "WARST": "WARST", "GMT": "GMT", "WITA": "WITA", "HNCU": "HNCU", "HNPMX": "HNPMX", "AEDT": "AEDT", "AKST": "AKST", "HEEG": "HEEG", "HNT": "HNT", "WAST": "WAST", "JST": "JST", "BOT": "BOT", "HAST": "HAST", "COST": "COST", "CHAST": "CHAST", "ARST": "ARST", "GYT": "GYT", "AWDT": "AWDT", "HKT": "HKT", "LHDT": "LHDT", "TMT": "TMT", "CDT": "CDT", "WIB": "WIB", "ACWDT": "ACWDT", "HENOMX": "HENOMX", "WIT": "WIT", "PST": "PST", "AST": "AST", "NZST": "NZST", "ACWST": "ACWST", "CLT": "CLT", "UYT": "UYT", "HNPM": "HNPM", "HNNOMX": "HNNOMX", "HADT": "HADT", "HECU": "HECU", "WAT": "WAT", "ECT": "ECT", "EST": "EST", "HNEG": "HNEG", "CST": "CST", "AWST": "AWST", "BT": "BT", "LHST": "LHST", "WART": "WART", "OESZ": "OESZ", "JDT": "JDT", "EDT": "EDT", "ACDT": "ACDT", "CAT": "CAT", "MDT": "MDT", "SRT": "SRT", "COT": "COT", "WEZ": "WEZ", "AKDT": "AKDT", "ACST": "ACST", "MST": "MST", "GFT": "GFT", "MESZ": "MESZ", "HKST": "HKST", "HEPM": "HEPM", "EAT": "EAT", "OEZ": "OEZ", "HEPMX": "HEPMX", "HEOG": "HEOG", "IST": "IST", "HAT": "HAT", "ChST": "ChST"},
	}
}

// Locale returns the current translators string locale
func (smn *smn) Locale() string {
	return smn.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'smn'
func (smn *smn) PluralsCardinal() []locales.PluralRule {
	return smn.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'smn'
func (smn *smn) PluralsOrdinal() []locales.PluralRule {
	return smn.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'smn'
func (smn *smn) PluralsRange() []locales.PluralRule {
	return smn.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'smn'
func (smn *smn) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	} else if n == 2 {
		return locales.PluralRuleTwo
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'smn'
func (smn *smn) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'smn'
func (smn *smn) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (smn *smn) MonthAbbreviated(month time.Month) string {
	return smn.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (smn *smn) MonthsAbbreviated() []string {
	return smn.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (smn *smn) MonthNarrow(month time.Month) string {
	return smn.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (smn *smn) MonthsNarrow() []string {
	return smn.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (smn *smn) MonthWide(month time.Month) string {
	return smn.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (smn *smn) MonthsWide() []string {
	return smn.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (smn *smn) WeekdayAbbreviated(weekday time.Weekday) string {
	return smn.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (smn *smn) WeekdaysAbbreviated() []string {
	return smn.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (smn *smn) WeekdayNarrow(weekday time.Weekday) string {
	return smn.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (smn *smn) WeekdaysNarrow() []string {
	return smn.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (smn *smn) WeekdayShort(weekday time.Weekday) string {
	return smn.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (smn *smn) WeekdaysShort() []string {
	return smn.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (smn *smn) WeekdayWide(weekday time.Weekday) string {
	return smn.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (smn *smn) WeekdaysWide() []string {
	return smn.daysWide
}

// Decimal returns the decimal point of number
func (smn *smn) Decimal() string {
	return smn.decimal
}

// Group returns the group of number
func (smn *smn) Group() string {
	return smn.group
}

// Group returns the minus sign of number
func (smn *smn) Minus() string {
	return smn.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'smn' and handles both Whole and Real numbers based on 'v'
func (smn *smn) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, smn.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(smn.group) - 1; j >= 0; j-- {
					b = append(b, smn.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, smn.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'smn' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (smn *smn) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 5
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, smn.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, smn.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, smn.percentSuffix...)

	b = append(b, smn.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'smn'
func (smn *smn) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := smn.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, smn.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(smn.group) - 1; j >= 0; j-- {
					b = append(b, smn.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, smn.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, smn.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, smn.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'smn'
// in accounting notation.
func (smn *smn) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := smn.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, smn.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(smn.group) - 1; j >= 0; j-- {
					b = append(b, smn.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, smn.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, smn.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, smn.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, smn.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'smn'
func (smn *smn) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e}...)
	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'smn'
func (smn *smn) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, smn.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'smn'
func (smn *smn) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, smn.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'smn'
func (smn *smn) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, []byte{0x63, 0x63, 0x63, 0x63, 0x2c, 0x20}...)
	b = append(b, smn.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'smn'
func (smn *smn) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'smn'
func (smn *smn) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'smn'
func (smn *smn) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'smn'
func (smn *smn) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := smn.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
