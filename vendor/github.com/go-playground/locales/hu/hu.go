package hu

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type hu struct {
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

// New returns a new instance of translator for the 'hu' locale
func New() locales.Translator {
	return &hu{
		locale:                 "hu",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         []locales.PluralRule{2, 6},
		pluralsRange:           []locales.PluralRule{2, 6},
		decimal:                ",",
		group:                  " ",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "Ft", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "jan.", "febr.", "márc.", "ápr.", "máj.", "jún.", "júl.", "aug.", "szept.", "okt.", "nov.", "dec."},
		monthsNarrow:           []string{"", "J", "F", "M", "Á", "M", "J", "J", "A", "Sz", "O", "N", "D"},
		monthsWide:             []string{"", "január", "február", "március", "április", "május", "június", "július", "augusztus", "szeptember", "október", "november", "december"},
		daysAbbreviated:        []string{"V", "H", "K", "Sze", "Cs", "P", "Szo"},
		daysNarrow:             []string{"V", "H", "K", "Sz", "Cs", "P", "Sz"},
		daysShort:              []string{"V", "H", "K", "Sze", "Cs", "P", "Szo"},
		daysWide:               []string{"vasárnap", "hétfő", "kedd", "szerda", "csütörtök", "péntek", "szombat"},
		periodsAbbreviated:     []string{"de.", "du."},
		periodsNarrow:          []string{"de.", "du."},
		periodsWide:            []string{"de.", "du."},
		erasAbbreviated:        []string{"i. e.", "i. sz."},
		erasNarrow:             []string{"ie.", "isz."},
		erasWide:               []string{"Krisztus előtt", "időszámításunk szerint"},
		timezones:              map[string]string{"HKST": "hongkongi nyári idő", "WARST": "nyugat-argentínai nyári idő", "COST": "kolumbiai nyári idő", "OEZ": "kelet-európai téli idő", "HNPMX": "mexikói csendes-óceáni zónaidő", "CDT": "középső államokbeli nyári idő", "AWST": "nyugat-ausztráliai téli idő", "ECT": "ecuadori téli idő", "MEZ": "közép-európai téli idő", "∅∅∅": "azori nyári idő", "WESZ": "nyugat-európai nyári idő", "GFT": "francia-guyanai idő", "ACST": "közép-ausztráliai téli idő", "ACWST": "közép-nyugat-ausztráliai téli idő", "WITA": "közép-indonéziai idő", "HEPM": "Saint-Pierre és Miquelon-i nyári idő", "CLST": "chilei nyári idő", "PDT": "csendes-óceáni nyári idő", "BT": "butáni idő", "NZDT": "új-zélandi nyári idő", "HEEG": "kelet-grönlandi nyári idő", "HNT": "új-fundlandi zónaidő", "HAT": "új-fundlandi nyári idő", "TMT": "türkmenisztáni téli idő", "ChST": "chamorrói téli idő", "SAST": "dél-afrikai téli idő", "WAST": "nyugat-afrikai nyári idő", "BOT": "bolíviai téli idő", "AKST": "alaszkai zónaidő", "EDT": "keleti államokbeli nyári idő", "ACWDT": "közép-nyugat-ausztráliai nyári idő", "MESZ": "közép-európai nyári idő", "LHST": "Lord Howe-szigeti téli idő", "LHDT": "Lord Howe-szigeti nyári idő", "HADT": "hawaii-aleuti nyári idő", "EST": "keleti államokbeli zónaidő", "HENOMX": "északnyugat-mexikói nyári idő", "WIT": "kelet-indonéziai idő", "OESZ": "kelet-európai nyári idő", "HKT": "hongkongi téli idő", "WART": "nyugat-argentínai téli idő", "MST": "Macaui zónaidő", "HEPMX": "mexikói csendes-óceáni nyári idő", "AST": "atlanti-óceáni zónaidő", "AEDT": "kelet-ausztráliai nyári idő", "TMST": "türkmenisztáni nyári idő", "CAT": "közép-afrikai téli idő", "EAT": "kelet-afrikai téli idő", "NZST": "új-zélandi téli idő", "MYT": "malajziai idő", "HNEG": "kelet-grönlandi téli idő", "GMT": "greenwichi középidő, téli idő", "HNOG": "nyugat-grönlandi téli idő", "COT": "kolumbiai téli idő", "PST": "csendes-óceáni zónaidő", "ART": "argentínai téli idő", "CHAST": "chathami téli idő", "CHADT": "chathami nyári idő", "JDT": "japán nyári idő", "HEOG": "nyugat-grönlandi nyári idő", "CLT": "chilei téli idő", "UYT": "uruguayi téli idő", "HNCU": "kubai téli idő", "SGT": "szingapúri téli idő", "ACDT": "közép-ausztráliai nyári idő", "HNNOMX": "északnyugat-mexikói zónaidő", "GYT": "guyanai téli idő", "CST": "középső államokbeli zónaidő", "WAT": "nyugat-afrikai téli idő", "IST": "indiai téli idő", "VET": "venezuelai idő", "SRT": "szurinámi idő", "HAST": "hawaii-aleuti téli idő", "HECU": "kubai nyári idő", "AWDT": "nyugat-ausztráliai nyári idő", "WIB": "nyugat-indonéziai téli idő", "WEZ": "nyugat-európai téli idő", "JST": "japán téli idő", "AKDT": "alaszkai nyári idő", "HNPM": "Saint-Pierre és Miquelon-i zónaidő", "MDT": "Macaui nyári idő", "ARST": "argentínai nyári idő", "UYST": "uruguayi nyári idő", "ADT": "atlanti-óceáni nyári idő", "AEST": "kelet-ausztráliai téli idő"},
	}
}

// Locale returns the current translators string locale
func (hu *hu) Locale() string {
	return hu.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'hu'
func (hu *hu) PluralsCardinal() []locales.PluralRule {
	return hu.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'hu'
func (hu *hu) PluralsOrdinal() []locales.PluralRule {
	return hu.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'hu'
func (hu *hu) PluralsRange() []locales.PluralRule {
	return hu.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'hu'
func (hu *hu) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'hu'
func (hu *hu) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 || n == 5 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'hu'
func (hu *hu) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := hu.CardinalPluralRule(num1, v1)
	end := hu.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (hu *hu) MonthAbbreviated(month time.Month) string {
	return hu.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (hu *hu) MonthsAbbreviated() []string {
	return hu.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (hu *hu) MonthNarrow(month time.Month) string {
	return hu.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (hu *hu) MonthsNarrow() []string {
	return hu.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (hu *hu) MonthWide(month time.Month) string {
	return hu.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (hu *hu) MonthsWide() []string {
	return hu.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (hu *hu) WeekdayAbbreviated(weekday time.Weekday) string {
	return hu.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (hu *hu) WeekdaysAbbreviated() []string {
	return hu.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (hu *hu) WeekdayNarrow(weekday time.Weekday) string {
	return hu.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (hu *hu) WeekdaysNarrow() []string {
	return hu.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (hu *hu) WeekdayShort(weekday time.Weekday) string {
	return hu.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (hu *hu) WeekdaysShort() []string {
	return hu.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (hu *hu) WeekdayWide(weekday time.Weekday) string {
	return hu.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (hu *hu) WeekdaysWide() []string {
	return hu.daysWide
}

// Decimal returns the decimal point of number
func (hu *hu) Decimal() string {
	return hu.decimal
}

// Group returns the group of number
func (hu *hu) Group() string {
	return hu.group
}

// Group returns the minus sign of number
func (hu *hu) Minus() string {
	return hu.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'hu' and handles both Whole and Real numbers based on 'v'
func (hu *hu) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, hu.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(hu.group) - 1; j >= 0; j-- {
					b = append(b, hu.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, hu.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'hu' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (hu *hu) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, hu.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, hu.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, hu.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'hu'
func (hu *hu) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := hu.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, hu.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(hu.group) - 1; j >= 0; j-- {
					b = append(b, hu.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, hu.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, hu.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, hu.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'hu'
// in accounting notation.
func (hu *hu) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := hu.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, hu.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(hu.group) - 1; j >= 0; j-- {
					b = append(b, hu.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, hu.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, hu.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, hu.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, hu.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'hu'
func (hu *hu) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2e, 0x20}...)

	if t.Month() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Month()), 10)

	b = append(b, []byte{0x2e, 0x20}...)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e}...)

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'hu'
func (hu *hu) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, hu.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e}...)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'hu'
func (hu *hu) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, hu.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e}...)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'hu'
func (hu *hu) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, hu.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x2c, 0x20}...)
	b = append(b, hu.daysWide[t.Weekday()]...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'hu'
func (hu *hu) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, hu.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'hu'
func (hu *hu) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, hu.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, hu.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'hu'
func (hu *hu) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, hu.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, hu.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'hu'
func (hu *hu) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, hu.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, hu.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := hu.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
