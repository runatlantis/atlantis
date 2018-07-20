package smn_FI

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type smn_FI struct {
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

// New returns a new instance of translator for the 'smn_FI' locale
func New() locales.Translator {
	return &smn_FI{
		locale:                 "smn_FI",
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
		timezones:              map[string]string{"SGT": "SGT", "ACWST": "ACWST", "WART": "WART", "WITA": "WITA", "ART": "ART", "GMT": "GMT", "AWDT": "AWDT", "HKT": "HKT", "OESZ": "OESZ", "UYST": "UYST", "PDT": "PDT", "CDT": "CDT", "HEOG": "HEOG", "IST": "IST", "LHST": "LHST", "HAT": "HAT", "CAT": "CAT", "CST": "CST", "HAST": "HAST", "AKDT": "AKDT", "HNNOMX": "HNNOMX", "TMST": "TMST", "HNOG": "HNOG", "WARST": "WARST", "NZST": "NZST", "WAT": "WAT", "LHDT": "LHDT", "HEPM": "HEPM", "HENOMX": "HENOMX", "EAT": "EAT", "GYT": "GYT", "HNCU": "HNCU", "MYT": "MYT", "BT": "BT", "AKST": "AKST", "ACWDT": "ACWDT", "HNEG": "HNEG", "CLT": "CLT", "CLST": "CLST", "AEDT": "AEDT", "ECT": "ECT", "EST": "EST", "EDT": "EDT", "SRT": "SRT", "OEZ": "OEZ", "HEPMX": "HEPMX", "SAST": "SAST", "GFT": "GFT", "HKST": "HKST", "HNT": "HNT", "ARST": "ARST", "COST": "COST", "ChST": "ChST", "PST": "PST", "JST": "JST", "MEZ": "MEZ", "HECU": "HECU", "JDT": "JDT", "AST": "AST", "AEST": "AEST", "WEZ": "WEZ", "TMT": "TMT", "HNPMX": "HNPMX", "CHADT": "CHADT", "WAST": "WAST", "BOT": "BOT", "HEEG": "HEEG", "MESZ": "MESZ", "VET": "VET", "∅∅∅": "∅∅∅", "NZDT": "NZDT", "ACDT": "ACDT", "WIT": "WIT", "AWST": "AWST", "MST": "MST", "MDT": "MDT", "ADT": "ADT", "WESZ": "WESZ", "WIB": "WIB", "ACST": "ACST", "HNPM": "HNPM", "CHAST": "CHAST", "HADT": "HADT", "UYT": "UYT", "COT": "COT"},
	}
}

// Locale returns the current translators string locale
func (smn *smn_FI) Locale() string {
	return smn.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'smn_FI'
func (smn *smn_FI) PluralsCardinal() []locales.PluralRule {
	return smn.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'smn_FI'
func (smn *smn_FI) PluralsOrdinal() []locales.PluralRule {
	return smn.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'smn_FI'
func (smn *smn_FI) PluralsRange() []locales.PluralRule {
	return smn.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'smn_FI'
func (smn *smn_FI) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	} else if n == 2 {
		return locales.PluralRuleTwo
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'smn_FI'
func (smn *smn_FI) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'smn_FI'
func (smn *smn_FI) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (smn *smn_FI) MonthAbbreviated(month time.Month) string {
	return smn.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (smn *smn_FI) MonthsAbbreviated() []string {
	return smn.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (smn *smn_FI) MonthNarrow(month time.Month) string {
	return smn.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (smn *smn_FI) MonthsNarrow() []string {
	return smn.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (smn *smn_FI) MonthWide(month time.Month) string {
	return smn.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (smn *smn_FI) MonthsWide() []string {
	return smn.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (smn *smn_FI) WeekdayAbbreviated(weekday time.Weekday) string {
	return smn.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (smn *smn_FI) WeekdaysAbbreviated() []string {
	return smn.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (smn *smn_FI) WeekdayNarrow(weekday time.Weekday) string {
	return smn.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (smn *smn_FI) WeekdaysNarrow() []string {
	return smn.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (smn *smn_FI) WeekdayShort(weekday time.Weekday) string {
	return smn.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (smn *smn_FI) WeekdaysShort() []string {
	return smn.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (smn *smn_FI) WeekdayWide(weekday time.Weekday) string {
	return smn.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (smn *smn_FI) WeekdaysWide() []string {
	return smn.daysWide
}

// Decimal returns the decimal point of number
func (smn *smn_FI) Decimal() string {
	return smn.decimal
}

// Group returns the group of number
func (smn *smn_FI) Group() string {
	return smn.group
}

// Group returns the minus sign of number
func (smn *smn_FI) Minus() string {
	return smn.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'smn_FI' and handles both Whole and Real numbers based on 'v'
func (smn *smn_FI) FmtNumber(num float64, v uint64) string {

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

// FmtPercent returns 'num' with digits/precision of 'v' for 'smn_FI' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (smn *smn_FI) FmtPercent(num float64, v uint64) string {
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

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'smn_FI'
func (smn *smn_FI) FmtCurrency(num float64, v uint64, currency currency.Type) string {

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

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'smn_FI'
// in accounting notation.
func (smn *smn_FI) FmtAccounting(num float64, v uint64, currency currency.Type) string {

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

// FmtDateShort returns the short date representation of 't' for 'smn_FI'
func (smn *smn_FI) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'smn_FI'
func (smn *smn_FI) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'smn_FI'
func (smn *smn_FI) FmtDateLong(t time.Time) string {

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

// FmtDateFull returns the full date representation of 't' for 'smn_FI'
func (smn *smn_FI) FmtDateFull(t time.Time) string {

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

// FmtTimeShort returns the short time representation of 't' for 'smn_FI'
func (smn *smn_FI) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'smn_FI'
func (smn *smn_FI) FmtTimeMedium(t time.Time) string {

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

// FmtTimeLong returns the long time representation of 't' for 'smn_FI'
func (smn *smn_FI) FmtTimeLong(t time.Time) string {

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

// FmtTimeFull returns the full time representation of 't' for 'smn_FI'
func (smn *smn_FI) FmtTimeFull(t time.Time) string {

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
