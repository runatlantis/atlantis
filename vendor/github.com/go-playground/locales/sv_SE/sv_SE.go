package sv_SE

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type sv_SE struct {
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

// New returns a new instance of translator for the 'sv_SE' locale
func New() locales.Translator {
	return &sv_SE{
		locale:                 "sv_SE",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         []locales.PluralRule{2, 6},
		pluralsRange:           []locales.PluralRule{6},
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
		monthsAbbreviated:      []string{"", "jan.", "feb.", "mars", "apr.", "maj", "juni", "juli", "aug.", "sep.", "okt.", "nov.", "dec."},
		monthsNarrow:           []string{"", "J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"},
		monthsWide:             []string{"", "januari", "februari", "mars", "april", "maj", "juni", "juli", "augusti", "september", "oktober", "november", "december"},
		daysAbbreviated:        []string{"sön", "mån", "tis", "ons", "tors", "fre", "lör"},
		daysNarrow:             []string{"S", "M", "T", "O", "T", "F", "L"},
		daysShort:              []string{"sö", "må", "ti", "on", "to", "fr", "lö"},
		daysWide:               []string{"söndag", "måndag", "tisdag", "onsdag", "torsdag", "fredag", "lördag"},
		periodsAbbreviated:     []string{"fm", "em"},
		periodsNarrow:          []string{"fm", "em"},
		periodsWide:            []string{"fm", "em"},
		erasAbbreviated:        []string{"f.Kr.", "e.Kr."},
		erasNarrow:             []string{"f.Kr.", "e.Kr."},
		erasWide:               []string{"före Kristus", "efter Kristus"},
		timezones:              map[string]string{"HNCU": "kubansk normaltid", "HNPMX": "mexikansk stillahavstid, normaltid", "HNNOMX": "nordvästmexikansk normaltid", "HENOMX": "nordvästmexikansk sommartid", "TMST": "turkmensk sommartid", "OEZ": "östeuropeisk normaltid", "UYT": "uruguayansk normaltid", "UYST": "uruguayansk sommartid", "AEST": "östaustralisk normaltid", "WAST": "västafrikansk sommartid", "JDT": "japansk sommartid", "HNEG": "östgrönländsk normaltid", "MESZ": "centraleuropeisk sommartid", "ECT": "ecuadoriansk tid", "HNT": "Newfoundland, normaltid", "MST": "Macaunormaltid", "ART": "östargentinsk normaltid", "COT": "colombiansk normaltid", "OESZ": "östeuropeisk sommartid", "AEDT": "östaustralisk sommartid", "WEZ": "västeuropeisk normaltid", "HEPM": "S:t Pierre och Miquelon, sommartid", "EAT": "östafrikansk tid", "GMT": "Greenwichtid", "HEPMX": "mexikansk stillahavstid, sommartid", "NZDT": "nyzeeländsk sommartid", "EDT": "östnordamerikansk sommartid", "WART": "västargentinsk normaltid", "MDT": "Macausommartid", "ARST": "östargentinsk sommartid", "AST": "nordamerikansk atlantnormaltid", "WIB": "västindonesisk tid", "HNOG": "västgrönländsk normaltid", "WARST": "västargentinsk sommartid", "SRT": "Surinamtid", "WIT": "östindonesisk tid", "HECU": "kubansk sommartid", "CHAST": "Chatham, normaltid", "CHADT": "Chatham, sommartid", "GYT": "Guyanatid", "CST": "centralnordamerikansk normaltid", "ACWDT": "västcentralaustralisk sommartid", "WITA": "centralindonesisk tid", "AWST": "västaustralisk normaltid", "BT": "bhutansk tid", "ACDT": "centralaustralisk sommartid", "LHDT": "Lord Howe, sommartid", "TMT": "turkmensk normaltid", "BOT": "boliviansk tid", "SGT": "Singaporetid", "HEEG": "östgrönländsk sommartid", "AWDT": "västaustralisk sommartid", "VET": "venezuelansk tid", "ACST": "centralaustralisk normaltid", "HKST": "Hongkong, sommartid", "MEZ": "centraleuropeisk normaltid", "HAST": "Honolulu, normaltid", "ADT": "nordamerikansk atlantsommartid", "GFT": "Franska Guyanatid", "AKDT": "Alaska, sommartid", "HEOG": "västgrönländsk sommartid", "LHST": "Lord Howe, normaltid", "COST": "colombiansk sommartid", "PDT": "västnordamerikansk sommartid", "WESZ": "västeuropeisk sommartid", "AKST": "Alaska, normaltid", "ACWST": "västcentralaustralisk normaltid", "HNPM": "S:t Pierre och Miquelon, normaltid", "NZST": "nyzeeländsk normaltid", "JST": "japansk normaltid", "CLT": "chilensk normaltid", "CLST": "chilensk sommartid", "HADT": "Honolulu, sommartid", "∅∅∅": "Amazonas, sommartid", "ChST": "Chamorrotid", "SAST": "sydafrikansk tid", "EST": "östnordamerikansk normaltid", "PST": "västnordamerikansk normaltid", "MYT": "malaysisk tid", "HAT": "Newfoundland, sommartid", "CAT": "centralafrikansk tid", "CDT": "centralnordamerikansk sommartid", "WAT": "västafrikansk normaltid", "HKT": "Hongkong, normaltid", "IST": "indisk tid"},
	}
}

// Locale returns the current translators string locale
func (sv *sv_SE) Locale() string {
	return sv.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'sv_SE'
func (sv *sv_SE) PluralsCardinal() []locales.PluralRule {
	return sv.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'sv_SE'
func (sv *sv_SE) PluralsOrdinal() []locales.PluralRule {
	return sv.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'sv_SE'
func (sv *sv_SE) PluralsRange() []locales.PluralRule {
	return sv.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'sv_SE'
func (sv *sv_SE) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)

	if i == 1 && v == 0 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'sv_SE'
func (sv *sv_SE) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	nMod10 := math.Mod(n, 10)
	nMod100 := math.Mod(n, 100)

	if (nMod10 == 1 || nMod10 == 2) && (nMod100 != 11 && nMod100 != 12) {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'sv_SE'
func (sv *sv_SE) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (sv *sv_SE) MonthAbbreviated(month time.Month) string {
	return sv.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (sv *sv_SE) MonthsAbbreviated() []string {
	return sv.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (sv *sv_SE) MonthNarrow(month time.Month) string {
	return sv.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (sv *sv_SE) MonthsNarrow() []string {
	return sv.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (sv *sv_SE) MonthWide(month time.Month) string {
	return sv.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (sv *sv_SE) MonthsWide() []string {
	return sv.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (sv *sv_SE) WeekdayAbbreviated(weekday time.Weekday) string {
	return sv.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (sv *sv_SE) WeekdaysAbbreviated() []string {
	return sv.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (sv *sv_SE) WeekdayNarrow(weekday time.Weekday) string {
	return sv.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (sv *sv_SE) WeekdaysNarrow() []string {
	return sv.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (sv *sv_SE) WeekdayShort(weekday time.Weekday) string {
	return sv.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (sv *sv_SE) WeekdaysShort() []string {
	return sv.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (sv *sv_SE) WeekdayWide(weekday time.Weekday) string {
	return sv.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (sv *sv_SE) WeekdaysWide() []string {
	return sv.daysWide
}

// Decimal returns the decimal point of number
func (sv *sv_SE) Decimal() string {
	return sv.decimal
}

// Group returns the group of number
func (sv *sv_SE) Group() string {
	return sv.group
}

// Group returns the minus sign of number
func (sv *sv_SE) Minus() string {
	return sv.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'sv_SE' and handles both Whole and Real numbers based on 'v'
func (sv *sv_SE) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sv.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(sv.group) - 1; j >= 0; j-- {
					b = append(b, sv.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(sv.minus) - 1; j >= 0; j-- {
			b = append(b, sv.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'sv_SE' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (sv *sv_SE) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 7
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sv.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(sv.minus) - 1; j >= 0; j-- {
			b = append(b, sv.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, sv.percentSuffix...)

	b = append(b, sv.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'sv_SE'
func (sv *sv_SE) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := sv.currencies[currency]
	l := len(s) + len(symbol) + 6 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sv.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(sv.group) - 1; j >= 0; j-- {
					b = append(b, sv.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(sv.minus) - 1; j >= 0; j-- {
			b = append(b, sv.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, sv.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, sv.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'sv_SE'
// in accounting notation.
func (sv *sv_SE) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := sv.currencies[currency]
	l := len(s) + len(symbol) + 6 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sv.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(sv.group) - 1; j >= 0; j-- {
					b = append(b, sv.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		for j := len(sv.minus) - 1; j >= 0; j-- {
			b = append(b, sv.minus[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, sv.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, sv.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, sv.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'sv_SE'
func (sv *sv_SE) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2d}...)

	if t.Month() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Month()), 10)

	b = append(b, []byte{0x2d}...)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'sv_SE'
func (sv *sv_SE) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, sv.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'sv_SE'
func (sv *sv_SE) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, sv.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'sv_SE'
func (sv *sv_SE) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, sv.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, sv.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'sv_SE'
func (sv *sv_SE) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sv.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'sv_SE'
func (sv *sv_SE) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sv.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, sv.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'sv_SE'
func (sv *sv_SE) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sv.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, sv.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'sv_SE'
func (sv *sv_SE) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, []byte{0x6b, 0x6c}...)
	b = append(b, []byte{0x2e, 0x20}...)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sv.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, sv.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := sv.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
