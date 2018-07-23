package hu_HU

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type hu_HU struct {
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

// New returns a new instance of translator for the 'hu_HU' locale
func New() locales.Translator {
	return &hu_HU{
		locale:                 "hu_HU",
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
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
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
		timezones:              map[string]string{"LHDT": "Lord Howe-szigeti nyári idő", "WITA": "közép-indonéziai idő", "TMST": "türkmenisztáni nyári idő", "HAST": "hawaii-aleuti téli idő", "CHAST": "chathami téli idő", "CDT": "középső államokbeli nyári idő", "AEDT": "kelet-ausztráliai nyári idő", "IST": "indiai téli idő", "HECU": "kubai nyári idő", "AWDT": "nyugat-ausztráliai nyári idő", "AKST": "alaszkai zónaidő", "HNT": "új-fundlandi zónaidő", "WARST": "nyugat-argentínai nyári idő", "HADT": "hawaii-aleuti nyári idő", "COT": "kolumbiai téli idő", "UYST": "uruguayi nyári idő", "JST": "japán téli idő", "BT": "butáni idő", "EST": "keleti államokbeli zónaidő", "ACDT": "közép-ausztráliai nyári idő", "WART": "nyugat-argentínai téli idő", "GMT": "greenwichi középidő, téli idő", "UYT": "uruguayi téli idő", "SAST": "dél-afrikai téli idő", "JDT": "japán nyári idő", "BOT": "bolíviai téli idő", "EDT": "keleti államokbeli nyári idő", "HEPM": "Saint-Pierre és Miquelon-i nyári idő", "HNNOMX": "északnyugat-mexikói zónaidő", "SRT": "szurinámi idő", "ChST": "chamorrói téli idő", "PDT": "csendes-óceáni nyári idő", "AST": "atlanti-óceáni zónaidő", "MESZ": "közép-európai nyári idő", "HNPM": "Saint-Pierre és Miquelon-i zónaidő", "WIT": "kelet-indonéziai idő", "HNCU": "kubai téli idő", "HNOG": "nyugat-grönlandi téli idő", "PST": "csendes-óceáni zónaidő", "HNPMX": "mexikói csendes-óceáni zónaidő", "AKDT": "alaszkai nyári idő", "HEOG": "nyugat-grönlandi nyári idő", "HKST": "hongkongi nyári idő", "VET": "venezuelai idő", "EAT": "kelet-afrikai téli idő", "CLST": "chilei nyári idő", "ARST": "argentínai nyári idő", "ADT": "atlanti-óceáni nyári idő", "WIB": "nyugat-indonéziai téli idő", "NZDT": "új-zélandi nyári idő", "COST": "kolumbiai nyári idő", "∅∅∅": "amazóniai nyári idő", "WEZ": "nyugat-európai téli idő", "LHST": "Lord Howe-szigeti téli idő", "ART": "argentínai téli idő", "HEPMX": "mexikói csendes-óceáni nyári idő", "AEST": "kelet-ausztráliai téli idő", "WAST": "nyugat-afrikai nyári idő", "HEEG": "kelet-grönlandi nyári idő", "NZST": "új-zélandi téli idő", "GFT": "francia-guyanai idő", "ACST": "közép-ausztráliai téli idő", "MEZ": "közép-európai téli idő", "HAT": "új-fundlandi nyári idő", "WAT": "nyugat-afrikai téli idő", "ECT": "ecuadori téli idő", "CAT": "közép-afrikai téli idő", "CLT": "chilei téli idő", "TMT": "türkmenisztáni téli idő", "GYT": "guyanai téli idő", "CHADT": "chathami nyári idő", "CST": "középső államokbeli zónaidő", "OESZ": "kelet-európai nyári idő", "WESZ": "nyugat-európai nyári idő", "ACWDT": "közép-nyugat-ausztráliai nyári idő", "HNEG": "kelet-grönlandi téli idő", "MDT": "hegyvidéki nyári idő", "HENOMX": "északnyugat-mexikói nyári idő", "HKT": "hongkongi téli idő", "OEZ": "kelet-európai téli idő", "AWST": "nyugat-ausztráliai téli idő", "MST": "hegyvidéki zónaidő", "MYT": "malajziai idő", "SGT": "szingapúri téli idő", "ACWST": "közép-nyugat-ausztráliai téli idő"},
	}
}

// Locale returns the current translators string locale
func (hu *hu_HU) Locale() string {
	return hu.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'hu_HU'
func (hu *hu_HU) PluralsCardinal() []locales.PluralRule {
	return hu.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'hu_HU'
func (hu *hu_HU) PluralsOrdinal() []locales.PluralRule {
	return hu.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'hu_HU'
func (hu *hu_HU) PluralsRange() []locales.PluralRule {
	return hu.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'hu_HU'
func (hu *hu_HU) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'hu_HU'
func (hu *hu_HU) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 || n == 5 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'hu_HU'
func (hu *hu_HU) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

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
func (hu *hu_HU) MonthAbbreviated(month time.Month) string {
	return hu.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (hu *hu_HU) MonthsAbbreviated() []string {
	return hu.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (hu *hu_HU) MonthNarrow(month time.Month) string {
	return hu.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (hu *hu_HU) MonthsNarrow() []string {
	return hu.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (hu *hu_HU) MonthWide(month time.Month) string {
	return hu.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (hu *hu_HU) MonthsWide() []string {
	return hu.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (hu *hu_HU) WeekdayAbbreviated(weekday time.Weekday) string {
	return hu.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (hu *hu_HU) WeekdaysAbbreviated() []string {
	return hu.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (hu *hu_HU) WeekdayNarrow(weekday time.Weekday) string {
	return hu.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (hu *hu_HU) WeekdaysNarrow() []string {
	return hu.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (hu *hu_HU) WeekdayShort(weekday time.Weekday) string {
	return hu.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (hu *hu_HU) WeekdaysShort() []string {
	return hu.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (hu *hu_HU) WeekdayWide(weekday time.Weekday) string {
	return hu.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (hu *hu_HU) WeekdaysWide() []string {
	return hu.daysWide
}

// Decimal returns the decimal point of number
func (hu *hu_HU) Decimal() string {
	return hu.decimal
}

// Group returns the group of number
func (hu *hu_HU) Group() string {
	return hu.group
}

// Group returns the minus sign of number
func (hu *hu_HU) Minus() string {
	return hu.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'hu_HU' and handles both Whole and Real numbers based on 'v'
func (hu *hu_HU) FmtNumber(num float64, v uint64) string {

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

// FmtPercent returns 'num' with digits/precision of 'v' for 'hu_HU' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (hu *hu_HU) FmtPercent(num float64, v uint64) string {
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

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'hu_HU'
func (hu *hu_HU) FmtCurrency(num float64, v uint64, currency currency.Type) string {

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

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'hu_HU'
// in accounting notation.
func (hu *hu_HU) FmtAccounting(num float64, v uint64, currency currency.Type) string {

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

// FmtDateShort returns the short date representation of 't' for 'hu_HU'
func (hu *hu_HU) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'hu_HU'
func (hu *hu_HU) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'hu_HU'
func (hu *hu_HU) FmtDateLong(t time.Time) string {

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

// FmtDateFull returns the full date representation of 't' for 'hu_HU'
func (hu *hu_HU) FmtDateFull(t time.Time) string {

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

// FmtTimeShort returns the short time representation of 't' for 'hu_HU'
func (hu *hu_HU) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, hu.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'hu_HU'
func (hu *hu_HU) FmtTimeMedium(t time.Time) string {

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

// FmtTimeLong returns the long time representation of 't' for 'hu_HU'
func (hu *hu_HU) FmtTimeLong(t time.Time) string {

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

// FmtTimeFull returns the full time representation of 't' for 'hu_HU'
func (hu *hu_HU) FmtTimeFull(t time.Time) string {

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
