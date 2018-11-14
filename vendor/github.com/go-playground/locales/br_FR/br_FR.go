package br_FR

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type br_FR struct {
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

// New returns a new instance of translator for the 'br_FR' locale
func New() locales.Translator {
	return &br_FR{
		locale:                 "br_FR",
		pluralsCardinal:        []locales.PluralRule{2, 3, 4, 5, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ",",
		group:                  " ",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		percentSuffix:          " ",
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "Gen.", "Cʼhwe.", "Meur.", "Ebr.", "Mae", "Mezh.", "Goue.", "Eost", "Gwen.", "Here", "Du", "Kzu."},
		monthsNarrow:           []string{"", "01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12"},
		monthsWide:             []string{"", "Genver", "Cʼhwevrer", "Meurzh", "Ebrel", "Mae", "Mezheven", "Gouere", "Eost", "Gwengolo", "Here", "Du", "Kerzu"},
		daysAbbreviated:        []string{"Sul", "Lun", "Meu.", "Mer.", "Yaou", "Gwe.", "Sad."},
		daysNarrow:             []string{"Su", "L", "Mz", "Mc", "Y", "G", "Sa"},
		daysShort:              []string{"Sul", "Lun", "Meu.", "Mer.", "Yaou", "Gwe.", "Sad."},
		daysWide:               []string{"Sul", "Lun", "Meurzh", "Mercʼher", "Yaou", "Gwener", "Sadorn"},
		periodsAbbreviated:     []string{"A.M.", "G.M."},
		periodsNarrow:          []string{"am", "gm"},
		periodsWide:            []string{"A.M.", "G.M."},
		erasAbbreviated:        []string{"a-raok J.K.", "goude J.K."},
		erasNarrow:             []string{"a-raok J.K.", "goude J.K."},
		erasWide:               []string{"a-raok Jezuz-Krist", "goude Jezuz-Krist"},
		timezones:              map[string]string{"CAT": "eur Kreizafrika", "COT": "eur cʼhoañv Kolombia", "EDT": "eur hañv ar Reter", "ACST": "eur cʼhoañv Kreizaostralia", "LHST": "LHST", "CHADT": "eur hañv Chatham", "PST": "PST", "MDT": "eur hañv ar Menezioù", "WART": "eur cʼhoañv Arcʼhantina ar Cʼhornôg", "PDT": "PDT", "AWDT": "eur hañv Aostralia ar Cʼhornôg", "GFT": "eur Gwiana cʼhall", "NZST": "eur cʼhoañv Zeland-Nevez", "HNOG": "eur cʼhoañv Greunland ar Cʼhornôg", "OEZ": "eur cʼhoañv Europa ar Reter", "ARST": "eur hañv Arcʼhantina", "CST": "CST", "OESZ": "eur hañv Europa ar Reter", "ART": "eur cʼhoañv Arcʼhantina", "AEDT": "eur hañv Aostralia ar Reter", "ACWST": "eur cʼhoañv Kreizaostralia ar Cʼhornôg", "SRT": "eur Surinam", "COST": "eur hañv Kolombia", "∅∅∅": "eur hañv an Amazon", "HADT": "HADT", "SAST": "eur cʼhoañv Suafrika", "EST": "eur cʼhoañv ar Reter", "VET": "eur Venezuela", "HNCU": "eur cʼhoañv Kuba", "HKST": "eur hañv Hong Kong", "HNT": "eur cʼhoañv Newfoundland", "EAT": "eur Afrika ar Reter", "CLST": "eur hañv Chile", "GMT": "Amzer keitat Greenwich (AKG)", "MST": "eur cʼhoañv ar Menezioù", "WEZ": "eur cʼhoañv Europa ar Cʼhornôg", "WESZ": "eur hañv Europa ar Cʼhornôg", "ACWDT": "eur hañv Kreizaostralia ar Cʼhornôg", "TMST": "eur hañv Turkmenistan", "GYT": "eur Guyana", "ChST": "ChST", "BT": "eur Bhoutan", "NZDT": "eur hañv Zeland-Nevez", "AKDT": "eur hañv Alaska", "WITA": "WITA", "TMT": "eur cʼhoañv Turkmenistan", "HAST": "HAST", "HECU": "eur hañv Kuba", "CDT": "CDT", "ECT": "eur Ecuador", "HNPMX": "HNPMX", "HEPMX": "HEPMX", "AST": "AST", "AEST": "eur cʼhoañv Aostralia ar Reter", "AKST": "eur cʼhoañv Alaska", "SGT": "eur cʼhoañv Singapour", "HEOG": "eur hañv Greunland ar Cʼhornôg", "ACDT": "eur hañv Kreizaostralia", "MEZ": "eur cʼhoañv Kreizeuropa", "HENOMX": "eur hañv Gwalarn Mecʼhiko", "WIT": "eur Indonezia ar Reter", "CLT": "eur cʼhoañv Chile", "UYST": "eur hañv Uruguay", "CHAST": "eur cʼhoañv Chatham", "WAST": "eur hañv Afrika ar Cʼhornôg", "HNEG": "eur cʼhoañv Greunland ar Reter", "HEEG": "eur hañv Greunland ar Reter", "HNPM": "eur cʼhoañv Sant-Pêr-ha-Mikelon", "ADT": "ADT", "JST": "eur cʼhoañv Japan", "MYT": "eur Malaysia", "MESZ": "eur hañv Kreizeuropa", "WARST": "eur hañv Arcʼhantina ar Cʼhornôg", "HNNOMX": "eur cʼhoañv Gwalarn Mecʼhiko", "WAT": "eur cʼhoañv Afrika ar Cʼhornôg", "BOT": "eur Bolivia", "JDT": "eur hañv Japan", "HKT": "eur cʼhoañv Hong Kong", "IST": "eur cʼhoañv India", "LHDT": "LHDT", "HAT": "eur hañv Newfoundland", "UYT": "eur cʼhoañv Uruguay", "HEPM": "eur hañv Sant-Pêr-ha-Mikelon", "AWST": "eur cʼhoañv Aostralia ar Cʼhornôg", "WIB": "eur Indonezia ar Cʼhornôg"},
	}
}

// Locale returns the current translators string locale
func (br *br_FR) Locale() string {
	return br.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'br_FR'
func (br *br_FR) PluralsCardinal() []locales.PluralRule {
	return br.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'br_FR'
func (br *br_FR) PluralsOrdinal() []locales.PluralRule {
	return br.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'br_FR'
func (br *br_FR) PluralsRange() []locales.PluralRule {
	return br.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'br_FR'
func (br *br_FR) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	nMod1000000 := math.Mod(n, 1000000)
	nMod10 := math.Mod(n, 10)
	nMod100 := math.Mod(n, 100)

	if nMod10 == 1 && (nMod100 != 11 && nMod100 != 71 && nMod100 != 91) {
		return locales.PluralRuleOne
	} else if nMod10 == 2 && (nMod100 != 12 && nMod100 != 72 && nMod100 != 92) {
		return locales.PluralRuleTwo
	} else if nMod10 >= 3 && nMod10 <= 4 && (nMod10 == 9) && (nMod100 < 10 || nMod100 > 19) || (nMod100 < 70 || nMod100 > 79) || (nMod100 < 90 || nMod100 > 99) {
		return locales.PluralRuleFew
	} else if n != 0 && nMod1000000 == 0 {
		return locales.PluralRuleMany
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'br_FR'
func (br *br_FR) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'br_FR'
func (br *br_FR) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (br *br_FR) MonthAbbreviated(month time.Month) string {
	return br.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (br *br_FR) MonthsAbbreviated() []string {
	return br.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (br *br_FR) MonthNarrow(month time.Month) string {
	return br.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (br *br_FR) MonthsNarrow() []string {
	return br.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (br *br_FR) MonthWide(month time.Month) string {
	return br.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (br *br_FR) MonthsWide() []string {
	return br.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (br *br_FR) WeekdayAbbreviated(weekday time.Weekday) string {
	return br.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (br *br_FR) WeekdaysAbbreviated() []string {
	return br.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (br *br_FR) WeekdayNarrow(weekday time.Weekday) string {
	return br.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (br *br_FR) WeekdaysNarrow() []string {
	return br.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (br *br_FR) WeekdayShort(weekday time.Weekday) string {
	return br.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (br *br_FR) WeekdaysShort() []string {
	return br.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (br *br_FR) WeekdayWide(weekday time.Weekday) string {
	return br.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (br *br_FR) WeekdaysWide() []string {
	return br.daysWide
}

// Decimal returns the decimal point of number
func (br *br_FR) Decimal() string {
	return br.decimal
}

// Group returns the group of number
func (br *br_FR) Group() string {
	return br.group
}

// Group returns the minus sign of number
func (br *br_FR) Minus() string {
	return br.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'br_FR' and handles both Whole and Real numbers based on 'v'
func (br *br_FR) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, br.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(br.group) - 1; j >= 0; j-- {
					b = append(b, br.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, br.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'br_FR' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (br *br_FR) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 5
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, br.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, br.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, br.percentSuffix...)

	b = append(b, br.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'br_FR'
func (br *br_FR) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := br.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, br.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(br.group) - 1; j >= 0; j-- {
					b = append(b, br.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, br.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, br.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, br.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'br_FR'
// in accounting notation.
func (br *br_FR) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := br.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, br.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(br.group) - 1; j >= 0; j-- {
					b = append(b, br.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, br.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, br.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, br.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, br.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'br_FR'
func (br *br_FR) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'br_FR'
func (br *br_FR) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, br.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'br_FR'
func (br *br_FR) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, br.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'br_FR'
func (br *br_FR) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, br.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, br.daysWide[t.Weekday()]...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'br_FR'
func (br *br_FR) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, br.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'br_FR'
func (br *br_FR) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, br.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, br.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'br_FR'
func (br *br_FR) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, br.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, br.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'br_FR'
func (br *br_FR) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, br.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, br.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := br.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
