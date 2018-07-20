package nn_NO

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type nn_NO struct {
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

// New returns a new instance of translator for the 'nn_NO' locale
func New() locales.Translator {
	return &nn_NO{
		locale:                 "nn_NO",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ",",
		group:                  " ",
		minus:                  "−",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		percentSuffix:          " ",
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "jan.", "feb.", "mars", "apr.", "mai", "juni", "juli", "aug.", "sep.", "okt.", "nov.", "des."},
		monthsNarrow:           []string{"", "J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"},
		monthsWide:             []string{"", "januar", "februar", "mars", "april", "mai", "juni", "juli", "august", "september", "oktober", "november", "desember"},
		daysAbbreviated:        []string{"sø.", "må.", "ty.", "on.", "to.", "fr.", "la."},
		daysNarrow:             []string{"S", "M", "T", "O", "T", "F", "L"},
		daysShort:              []string{"sø.", "må.", "ty.", "on.", "to.", "fr.", "la."},
		daysWide:               []string{"søndag", "måndag", "tysdag", "onsdag", "torsdag", "fredag", "laurdag"},
		periodsAbbreviated:     []string{"f.m.", "e.m."},
		periodsNarrow:          []string{"f.m.", "e.m."},
		periodsWide:            []string{"formiddag", "ettermiddag"},
		erasAbbreviated:        []string{"f.Kr.", "e.Kr."},
		erasNarrow:             []string{"f.Kr.", "e.Kr."},
		erasWide:               []string{"f.Kr.", "e.Kr."},
		timezones:              map[string]string{"MESZ": "sentraleuropeisk sommartid", "HKST": "hongkongkinesisk sumartid", "COT": "kolombiansk normaltid", "HECU": "kubansk sumartid", "ECT": "ecuadoriansk tid", "EAT": "austafrikansk tid", "AEST": "austaustralsk standardtid", "BT": "bhutansk tid", "HEEG": "austgrønlandsk sumartid", "HEPM": "sumartid for Saint-Pierre-et-Miquelon", "HNNOMX": "normaltid for nordvestlege Mexico", "HNPM": "normaltid for Saint-Pierre-et-Miquelon", "WIT": "austindonesisk tid", "OEZ": "austeuropeisk standardtid", "PDT": "sumartid for den nordamerikanske stillehavskysten", "AWDT": "vestaustralsk sommartid", "HEOG": "vestgrønlandsk sumartid", "MEZ": "sentraleuropeisk standardtid", "∅∅∅": "asorisk sumartid", "NZDT": "nyzealandsk sumartid", "OESZ": "austeuropeisk sommartid", "ARST": "argentinsk sumartid", "UYST": "uruguayansk sumartid", "ADT": "sumartid for den nordamerikanske atlanterhavskysten", "JDT": "japansk sumartid", "WART": "vestargentinsk normaltid", "MST": "MST", "SRT": "surinamsk tid", "AKST": "alaskisk normaltid", "TMST": "turkmensk sumartid", "CHADT": "sumartid for Chatham", "WAT": "vestafrikansk standardtid", "JST": "japansk normaltid", "SGT": "singaporsk tid", "ACWDT": "vest-sentralaustralsk sommartid", "HNOG": "vestgrønlandsk normaltid", "VET": "venezuelansk tid", "GMT": "Greenwich middeltid", "UYT": "uruguayansk normaltid", "CDT": "sumartid for sentrale Nord-Amerika", "AWST": "vestaustralsk standardtid", "WAST": "vestafrikansk sommartid", "ACST": "sentralaustralsk standardtid", "ACWST": "vest-sentralaustralsk standardtid", "ART": "argentinsk normaltid", "AST": "normaltid for den nordamerikanske atlanterhavskysten", "SAST": "sørafrikansk tid", "WEZ": "vesteuropeisk standardtid", "NZST": "nyzealandsk normaltid", "GFT": "tidssone for Fransk Guyana", "IST": "indisk tid", "CLT": "chilensk normaltid", "HEPMX": "sumartid for den meksikanske stillehavskysten", "EST": "normaltid for den nordamerikansk austkysten", "EDT": "sumartid for den nordamerikansk austkysten", "WITA": "sentralindonesisk tid", "COST": "kolombiansk sumartid", "CST": "normaltid for sentrale Nord-Amerika", "AEDT": "austaustralsk sommartid", "ACDT": "sentralaustralsk sommartid", "HKT": "hongkongkinesisk normaltid", "LHDT": "sumartid for Lord Howe-øya", "BOT": "boliviansk tid", "AKDT": "alaskisk sumartid", "TMT": "turkmensk normaltid", "HNPMX": "normaltid for den meksikanske stillehavskysten", "MYT": "malaysisk tid", "WARST": "vestargentinsk sumartid", "CHAST": "normaltid for Chatham", "WESZ": "vesteuropeisk sommartid", "HNEG": "austgrønlandsk normaltid", "HADT": "sumartid for Hawaii og Aleutene", "HNCU": "kubansk normaltid", "HAT": "sumartid for Newfoundland", "MDT": "MDT", "CLST": "chilensk sumartid", "HENOMX": "sumartid for nordvestlege Mexico", "PST": "normaltid for den nordamerikanske stillehavskysten", "HAST": "normaltid for Hawaii og Aleutene", "GYT": "guyansk tid", "ChST": "tidssone for Chamorro", "WIB": "vestindonesisk tid", "LHST": "normaltid for Lord Howe-øya", "HNT": "normaltid for Newfoundland", "CAT": "sentralafrikansk tid"},
	}
}

// Locale returns the current translators string locale
func (nn *nn_NO) Locale() string {
	return nn.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'nn_NO'
func (nn *nn_NO) PluralsCardinal() []locales.PluralRule {
	return nn.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'nn_NO'
func (nn *nn_NO) PluralsOrdinal() []locales.PluralRule {
	return nn.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'nn_NO'
func (nn *nn_NO) PluralsRange() []locales.PluralRule {
	return nn.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'nn_NO'
func (nn *nn_NO) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'nn_NO'
func (nn *nn_NO) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'nn_NO'
func (nn *nn_NO) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (nn *nn_NO) MonthAbbreviated(month time.Month) string {
	return nn.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (nn *nn_NO) MonthsAbbreviated() []string {
	return nn.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (nn *nn_NO) MonthNarrow(month time.Month) string {
	return nn.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (nn *nn_NO) MonthsNarrow() []string {
	return nn.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (nn *nn_NO) MonthWide(month time.Month) string {
	return nn.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (nn *nn_NO) MonthsWide() []string {
	return nn.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (nn *nn_NO) WeekdayAbbreviated(weekday time.Weekday) string {
	return nn.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (nn *nn_NO) WeekdaysAbbreviated() []string {
	return nn.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (nn *nn_NO) WeekdayNarrow(weekday time.Weekday) string {
	return nn.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (nn *nn_NO) WeekdaysNarrow() []string {
	return nn.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (nn *nn_NO) WeekdayShort(weekday time.Weekday) string {
	return nn.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (nn *nn_NO) WeekdaysShort() []string {
	return nn.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (nn *nn_NO) WeekdayWide(weekday time.Weekday) string {
	return nn.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (nn *nn_NO) WeekdaysWide() []string {
	return nn.daysWide
}

// Decimal returns the decimal point of number
func (nn *nn_NO) Decimal() string {
	return nn.decimal
}

// Group returns the group of number
func (nn *nn_NO) Group() string {
	return nn.group
}

// Group returns the minus sign of number
func (nn *nn_NO) Minus() string {
	return nn.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'nn_NO' and handles both Whole and Real numbers based on 'v'
func (nn *nn_NO) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, nn.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(nn.group) - 1; j >= 0; j-- {
					b = append(b, nn.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(nn.minus) - 1; j >= 0; j-- {
			b = append(b, nn.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'nn_NO' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (nn *nn_NO) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 7
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, nn.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(nn.minus) - 1; j >= 0; j-- {
			b = append(b, nn.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, nn.percentSuffix...)

	b = append(b, nn.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'nn_NO'
func (nn *nn_NO) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := nn.currencies[currency]
	l := len(s) + len(symbol) + 6 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, nn.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(nn.group) - 1; j >= 0; j-- {
					b = append(b, nn.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(nn.minus) - 1; j >= 0; j-- {
			b = append(b, nn.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, nn.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, nn.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'nn_NO'
// in accounting notation.
func (nn *nn_NO) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := nn.currencies[currency]
	l := len(s) + len(symbol) + 6 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, nn.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(nn.group) - 1; j >= 0; j-- {
					b = append(b, nn.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		for j := len(nn.minus) - 1; j >= 0; j-- {
			b = append(b, nn.minus[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, nn.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, nn.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, nn.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'nn_NO'
func (nn *nn_NO) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Month() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Month()), 10)

	b = append(b, []byte{0x2e}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'nn_NO'
func (nn *nn_NO) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, nn.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'nn_NO'
func (nn *nn_NO) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, nn.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'nn_NO'
func (nn *nn_NO) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, nn.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, nn.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'nn_NO'
func (nn *nn_NO) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, nn.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'nn_NO'
func (nn *nn_NO) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, nn.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, nn.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'nn_NO'
func (nn *nn_NO) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, nn.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, nn.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'nn_NO'
func (nn *nn_NO) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, []byte{0x6b, 0x6c}...)
	b = append(b, []byte{0x2e, 0x20}...)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, nn.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, nn.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := nn.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
