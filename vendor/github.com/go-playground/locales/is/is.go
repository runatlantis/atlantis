package is

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type is struct {
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

// New returns a new instance of translator for the 'is' locale
func New() locales.Translator {
	return &is{
		locale:                 "is",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         []locales.PluralRule{6},
		pluralsRange:           []locales.PluralRule{2, 6},
		decimal:                ",",
		group:                  ".",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "₪", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JP¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "EC$", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "jan.", "feb.", "mar.", "apr.", "maí", "jún.", "júl.", "ágú.", "sep.", "okt.", "nóv.", "des."},
		monthsNarrow:           []string{"", "J", "F", "M", "A", "M", "J", "J", "Á", "S", "O", "N", "D"},
		monthsWide:             []string{"", "janúar", "febrúar", "mars", "apríl", "maí", "júní", "júlí", "ágúst", "september", "október", "nóvember", "desember"},
		daysAbbreviated:        []string{"sun.", "mán.", "þri.", "mið.", "fim.", "fös.", "lau."},
		daysNarrow:             []string{"S", "M", "Þ", "M", "F", "F", "L"},
		daysShort:              []string{"su.", "má.", "þr.", "mi.", "fi.", "fö.", "la."},
		daysWide:               []string{"sunnudagur", "mánudagur", "þriðjudagur", "miðvikudagur", "fimmtudagur", "föstudagur", "laugardagur"},
		periodsAbbreviated:     []string{"f.h.", "e.h."},
		periodsNarrow:          []string{"f.", "e."},
		periodsWide:            []string{"f.h.", "e.h."},
		erasAbbreviated:        []string{"f.Kr.", "e.Kr."},
		erasNarrow:             []string{"f.k.", "e.k."},
		erasWide:               []string{"fyrir Krist", "eftir Krist"},
		timezones:              map[string]string{"HEOG": "Sumartími á Vestur-Grænlandi", "ACDT": "Sumartími í Mið-Ástralíu", "MST": "MST", "MESZ": "Sumartími í Mið-Evrópu", "LHST": "Staðaltími á Lord Howe-eyju", "CST": "Staðaltími í miðhluta Bandaríkjanna og Kanada", "SAST": "Suður-Afríkutími", "HNEG": "Staðaltími á Austur-Grænlandi", "HEEG": "Sumartími á Austur-Grænlandi", "ACWST": "Staðaltími í miðvesturhluta Ástralíu", "HNPM": "Staðaltími á Sankti Pierre og Miquelon", "OEZ": "Staðaltími í Austur-Evrópu", "OESZ": "Sumartími í Austur-Evrópu", "CHAST": "Staðaltími í Chatham", "ADT": "Sumartími á Atlantshafssvæðinu", "AKDT": "Sumartími í Alaska", "ACWDT": "Sumartími í miðvesturhluta Ástralíu", "HKT": "Staðaltími í Hong Kong", "HAT": "Sumartími á Nýfundnalandi", "PST": "Staðaltími á Kyrrahafssvæðinu", "UYT": "Staðaltími í Úrúgvæ", "WIB": "Vestur-Indónesíutími", "JDT": "Sumartími í Japan", "EDT": "Sumartími í austurhluta Bandaríkjanna og Kanada", "∅∅∅": "Sumartími á Asóreyjum", "TMT": "Staðaltími í Túrkmenistan", "JST": "Staðaltími í Japan", "EAT": "Austur-Afríkutími", "AEST": "Staðaltími í Austur-Ástralíu", "BOT": "Bólivíutími", "WART": "Staðaltími í Vestur-Argentínu", "VET": "Venesúelatími", "HNNOMX": "Staðaltími í Norðvestur-Mexíkó", "WIT": "Austur-Indónesíutími", "COT": "Staðaltími í Kólumbíu", "SGT": "Singapúrtími", "EST": "Staðaltími í austurhluta Bandaríkjanna og Kanada", "LHDT": "Sumartími á Lord Howe-eyju", "HECU": "Sumartími á Kúbu", "HAST": "Staðaltími á Havaí og Aleúta", "GYT": "Gvæjanatími", "UYST": "Sumartími í Úrúgvæ", "GMT": "Greenwich-staðaltími", "CHADT": "Sumartími í Chatham", "NZDT": "Sumartími á Nýja-Sjálandi", "HEPM": "Sumartími á Sankti Pierre og Miquelon", "HENOMX": "Sumartími í Norðvestur-Mexíkó", "CLT": "Staðaltími í Síle", "HNPMX": "Staðaltími í Mexíkó á Kyrrahafssvæðinu", "WESZ": "Sumartími í Vestur-Evrópu", "MYT": "Malasíutími", "CDT": "Sumartími í miðhluta Bandaríkjanna og Kanada", "PDT": "Sumartími á Kyrrahafssvæðinu", "AWST": "Staðaltími í Vestur-Ástralíu", "COST": "Sumartími í Kólumbíu", "ChST": "Chamorro-staðaltími", "HNCU": "Staðaltími á Kúbu", "WAST": "Sumartími í Vestur-Afríku", "WEZ": "Staðaltími í Vestur-Evrópu", "MEZ": "Staðaltími í Mið-Evrópu", "IST": "Indlandstími", "WARST": "Sumartími í Vestur-Argentínu", "HNT": "Staðaltími á Nýfundnalandi", "SRT": "Súrinamtími", "AST": "Staðaltími á Atlantshafssvæðinu", "GFT": "Tími í Frönsku Gvæjana", "ECT": "Ekvadortími", "HADT": "Sumartími á Havaí og Aleúta", "AWDT": "Sumartími í Vestur-Ástralíu", "HKST": "Sumartími í Hong Kong", "CAT": "Mið-Afríkutími", "ARST": "Sumartími í Argentínu", "WAT": "Staðaltími í Vestur-Afríku", "NZST": "Staðaltími á Nýja-Sjálandi", "AKST": "Staðaltími í Alaska", "HNOG": "Staðaltími á Vestur-Grænlandi", "ACST": "Staðaltími í Mið-Ástralíu", "CLST": "Sumartími í Síle", "ART": "Staðaltími í Argentínu", "HEPMX": "Sumartími í Mexíkó á Kyrrahafssvæðinu", "AEDT": "Sumartími í Austur-Ástralíu", "BT": "Bútantími", "WITA": "Mið-Indónesíutími", "MDT": "MDT", "TMST": "Sumartími í Túrkmenistan"},
	}
}

// Locale returns the current translators string locale
func (is *is) Locale() string {
	return is.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'is'
func (is *is) PluralsCardinal() []locales.PluralRule {
	return is.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'is'
func (is *is) PluralsOrdinal() []locales.PluralRule {
	return is.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'is'
func (is *is) PluralsRange() []locales.PluralRule {
	return is.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'is'
func (is *is) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)
	t := locales.T(n, v)
	iMod10 := i % 10
	iMod100 := i % 100

	if (t == 0 && iMod10 == 1 && iMod100 != 11) || (t != 0) {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'is'
func (is *is) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'is'
func (is *is) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := is.CardinalPluralRule(num1, v1)
	end := is.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (is *is) MonthAbbreviated(month time.Month) string {
	return is.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (is *is) MonthsAbbreviated() []string {
	return is.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (is *is) MonthNarrow(month time.Month) string {
	return is.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (is *is) MonthsNarrow() []string {
	return is.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (is *is) MonthWide(month time.Month) string {
	return is.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (is *is) MonthsWide() []string {
	return is.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (is *is) WeekdayAbbreviated(weekday time.Weekday) string {
	return is.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (is *is) WeekdaysAbbreviated() []string {
	return is.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (is *is) WeekdayNarrow(weekday time.Weekday) string {
	return is.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (is *is) WeekdaysNarrow() []string {
	return is.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (is *is) WeekdayShort(weekday time.Weekday) string {
	return is.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (is *is) WeekdaysShort() []string {
	return is.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (is *is) WeekdayWide(weekday time.Weekday) string {
	return is.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (is *is) WeekdaysWide() []string {
	return is.daysWide
}

// Decimal returns the decimal point of number
func (is *is) Decimal() string {
	return is.decimal
}

// Group returns the group of number
func (is *is) Group() string {
	return is.group
}

// Group returns the minus sign of number
func (is *is) Minus() string {
	return is.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'is' and handles both Whole and Real numbers based on 'v'
func (is *is) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, is.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, is.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, is.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'is' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (is *is) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, is.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, is.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, is.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'is'
func (is *is) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := is.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, is.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, is.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, is.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, is.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, is.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'is'
// in accounting notation.
func (is *is) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := is.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, is.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, is.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, is.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, is.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, is.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, is.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'is'
func (is *is) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'is'
func (is *is) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, is.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'is'
func (is *is) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, is.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'is'
func (is *is) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, is.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, is.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'is'
func (is *is) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, is.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'is'
func (is *is) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, is.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, is.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'is'
func (is *is) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, is.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, is.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'is'
func (is *is) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, is.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, is.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := is.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
