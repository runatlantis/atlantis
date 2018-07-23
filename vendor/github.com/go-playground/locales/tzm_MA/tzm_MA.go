package tzm_MA

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type tzm_MA struct {
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

// New returns a new instance of translator for the 'tzm_MA' locale
func New() locales.Translator {
	return &tzm_MA{
		locale:                 "tzm_MA",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ",",
		group:                  " ",
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "Yen", "Yeb", "Mar", "Ibr", "May", "Yun", "Yul", "Ɣuc", "Cut", "Kṭu", "Nwa", "Duj"},
		monthsNarrow:           []string{"", "Y", "Y", "M", "I", "M", "Y", "Y", "Ɣ", "C", "K", "N", "D"},
		monthsWide:             []string{"", "Yennayer", "Yebrayer", "Mars", "Ibrir", "Mayyu", "Yunyu", "Yulyuz", "Ɣuct", "Cutanbir", "Kṭuber", "Nwanbir", "Dujanbir"},
		daysAbbreviated:        []string{"Asa", "Ayn", "Asn", "Akr", "Akw", "Asm", "Asḍ"},
		daysNarrow:             []string{"A", "A", "A", "A", "A", "A", "A"},
		daysWide:               []string{"Asamas", "Aynas", "Asinas", "Akras", "Akwas", "Asimwas", "Asiḍyas"},
		periodsAbbreviated:     []string{"Zdat azal", "Ḍeffir aza"},
		periodsWide:            []string{"Zdat azal", "Ḍeffir aza"},
		erasAbbreviated:        []string{"ZƐ", "ḌƐ"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"Zdat Ɛisa (TAƔ)", "Ḍeffir Ɛisa (TAƔ)"},
		timezones:              map[string]string{"OESZ": "OESZ", "GYT": "GYT", "HECU": "HECU", "HEPMX": "HEPMX", "ADT": "ADT", "WARST": "WARST", "TMST": "TMST", "CAT": "CAT", "WIB": "WIB", "HEEG": "HEEG", "NZST": "NZST", "GFT": "GFT", "LHST": "LHST", "CLST": "CLST", "UYST": "UYST", "AST": "AST", "VET": "VET", "AKST": "AKST", "EDT": "EDT", "AWDT": "AWDT", "CDT": "CDT", "AEDT": "AEDT", "NZDT": "NZDT", "ACWDT": "ACWDT", "WITA": "WITA", "WIT": "WIT", "COT": "COT", "JST": "JST", "HNOG": "HNOG", "HEOG": "HEOG", "HNNOMX": "HNNOMX", "WEZ": "WEZ", "WESZ": "WESZ", "HNEG": "HNEG", "LHDT": "LHDT", "CHADT": "CHADT", "CST": "CST", "SGT": "SGT", "ECT": "ECT", "SRT": "SRT", "MDT": "MDT", "CLT": "CLT", "HNPMX": "HNPMX", "AEST": "AEST", "SAST": "SAST", "HNPM": "HNPM", "HAST": "HAST", "COST": "COST", "AKDT": "AKDT", "CHAST": "CHAST", "MYT": "MYT", "BOT": "BOT", "HNCU": "HNCU", "WAT": "WAT", "HKST": "HKST", "MESZ": "MESZ", "HAT": "HAT", "BT": "BT", "ACWST": "ACWST", "TMT": "TMT", "UYT": "UYT", "AWST": "AWST", "ChST": "ChST", "PST": "PST", "MEZ": "MEZ", "MST": "MST", "GMT": "GMT", "ART": "ART", "OEZ": "OEZ", "PDT": "PDT", "ACST": "ACST", "∅∅∅": "∅∅∅", "IST": "IST", "HEPM": "HEPM", "ARST": "ARST", "WAST": "WAST", "JDT": "JDT", "ACDT": "ACDT", "EST": "EST", "WART": "WART", "HNT": "HNT", "EAT": "EAT", "HKT": "HKT", "HENOMX": "HENOMX", "HADT": "HADT"},
	}
}

// Locale returns the current translators string locale
func (tzm *tzm_MA) Locale() string {
	return tzm.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'tzm_MA'
func (tzm *tzm_MA) PluralsCardinal() []locales.PluralRule {
	return tzm.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'tzm_MA'
func (tzm *tzm_MA) PluralsOrdinal() []locales.PluralRule {
	return tzm.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'tzm_MA'
func (tzm *tzm_MA) PluralsRange() []locales.PluralRule {
	return tzm.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'tzm_MA'
func (tzm *tzm_MA) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if (n >= 0 && n <= 1) || (n >= 11 && n <= 99) {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'tzm_MA'
func (tzm *tzm_MA) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'tzm_MA'
func (tzm *tzm_MA) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (tzm *tzm_MA) MonthAbbreviated(month time.Month) string {
	return tzm.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (tzm *tzm_MA) MonthsAbbreviated() []string {
	return tzm.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (tzm *tzm_MA) MonthNarrow(month time.Month) string {
	return tzm.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (tzm *tzm_MA) MonthsNarrow() []string {
	return tzm.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (tzm *tzm_MA) MonthWide(month time.Month) string {
	return tzm.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (tzm *tzm_MA) MonthsWide() []string {
	return tzm.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (tzm *tzm_MA) WeekdayAbbreviated(weekday time.Weekday) string {
	return tzm.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (tzm *tzm_MA) WeekdaysAbbreviated() []string {
	return tzm.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (tzm *tzm_MA) WeekdayNarrow(weekday time.Weekday) string {
	return tzm.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (tzm *tzm_MA) WeekdaysNarrow() []string {
	return tzm.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (tzm *tzm_MA) WeekdayShort(weekday time.Weekday) string {
	return tzm.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (tzm *tzm_MA) WeekdaysShort() []string {
	return tzm.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (tzm *tzm_MA) WeekdayWide(weekday time.Weekday) string {
	return tzm.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (tzm *tzm_MA) WeekdaysWide() []string {
	return tzm.daysWide
}

// Decimal returns the decimal point of number
func (tzm *tzm_MA) Decimal() string {
	return tzm.decimal
}

// Group returns the group of number
func (tzm *tzm_MA) Group() string {
	return tzm.group
}

// Group returns the minus sign of number
func (tzm *tzm_MA) Minus() string {
	return tzm.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'tzm_MA' and handles both Whole and Real numbers based on 'v'
func (tzm *tzm_MA) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'tzm_MA' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (tzm *tzm_MA) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'tzm_MA'
func (tzm *tzm_MA) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := tzm.currencies[currency]
	l := len(s) + len(symbol) + 3 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, tzm.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(tzm.group) - 1; j >= 0; j-- {
					b = append(b, tzm.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, tzm.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, tzm.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, tzm.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'tzm_MA'
// in accounting notation.
func (tzm *tzm_MA) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := tzm.currencies[currency]
	l := len(s) + len(symbol) + 3 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, tzm.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(tzm.group) - 1; j >= 0; j-- {
					b = append(b, tzm.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, tzm.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, tzm.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, tzm.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, tzm.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'tzm_MA'
func (tzm *tzm_MA) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'tzm_MA'
func (tzm *tzm_MA) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, tzm.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'tzm_MA'
func (tzm *tzm_MA) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, tzm.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'tzm_MA'
func (tzm *tzm_MA) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, tzm.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, tzm.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'tzm_MA'
func (tzm *tzm_MA) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'tzm_MA'
func (tzm *tzm_MA) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'tzm_MA'
func (tzm *tzm_MA) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'tzm_MA'
func (tzm *tzm_MA) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}
