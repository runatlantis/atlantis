package cs_CZ

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type cs_CZ struct {
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

// New returns a new instance of translator for the 'cs_CZ' locale
func New() locales.Translator {
	return &cs_CZ{
		locale:                 "cs_CZ",
		pluralsCardinal:        []locales.PluralRule{2, 4, 5, 6},
		pluralsOrdinal:         []locales.PluralRule{6},
		pluralsRange:           []locales.PluralRule{2, 4, 5, 6},
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
		monthsAbbreviated:      []string{"", "led", "úno", "bře", "dub", "kvě", "čvn", "čvc", "srp", "zář", "říj", "lis", "pro"},
		monthsNarrow:           []string{"", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"},
		monthsWide:             []string{"", "ledna", "února", "března", "dubna", "května", "června", "července", "srpna", "září", "října", "listopadu", "prosince"},
		daysAbbreviated:        []string{"ne", "po", "út", "st", "čt", "pá", "so"},
		daysNarrow:             []string{"N", "P", "Ú", "S", "Č", "P", "S"},
		daysShort:              []string{"ne", "po", "út", "st", "čt", "pá", "so"},
		daysWide:               []string{"neděle", "pondělí", "úterý", "středa", "čtvrtek", "pátek", "sobota"},
		periodsAbbreviated:     []string{"dop.", "odp."},
		periodsNarrow:          []string{"dop.", "odp."},
		periodsWide:            []string{"dop.", "odp."},
		erasAbbreviated:        []string{"př. n. l.", "n. l."},
		erasNarrow:             []string{"př.n.l.", "n.l."},
		erasWide:               []string{"před naším letopočtem", "našeho letopočtu"},
		timezones:              map[string]string{"∅∅∅": "Azorský letní čas", "TMT": "Turkmenský standardní čas", "MYT": "Malajský čas", "BT": "Bhútánský čas", "COST": "Kolumbijský letní čas", "GYT": "Guyanský čas", "ACST": "Středoaustralský standardní čas", "EDT": "Severoamerický východní letní čas", "MESZ": "Středoevropský letní čas", "VET": "Venezuelský čas", "WIT": "Východoindonéský čas", "ARST": "Argentinský letní čas", "CHAST": "Chathamský standardní čas", "AWST": "Západoaustralský standardní čas", "AST": "Atlantický standardní čas", "HNEG": "Východogrónský standardní čas", "WITA": "Středoindonéský čas", "CLT": "Chilský standardní čas", "ChST": "Chamorrský čas", "HECU": "Kubánský letní čas", "NZST": "Novozélandský standardní čas", "WAT": "Západoafrický standardní čas", "WAST": "Západoafrický letní čas", "HAST": "Havajsko-aleutský standardní čas", "OEZ": "Východoevropský standardní čas", "GMT": "Greenwichský střední čas", "HEPMX": "Mexický pacifický letní čas", "CST": "Severoamerický centrální standardní čas", "AEDT": "Východoaustralský letní čas", "ART": "Argentinský standardní čas", "ADT": "Atlantický letní čas", "SGT": "Singapurský čas", "OESZ": "Východoevropský letní čas", "HNCU": "Kubánský standardní čas", "WEZ": "Západoevropský standardní čas", "ECT": "Ekvádorský čas", "AKST": "Aljašský standardní čas", "ACWST": "Středozápadní australský standardní čas", "HEOG": "Západogrónský letní čas", "UYST": "Uruguayský letní čas", "SAST": "Jihoafrický čas", "AKDT": "Aljašský letní čas", "IST": "Indický čas", "SRT": "Surinamský čas", "PST": "Severoamerický pacifický standardní čas", "AWDT": "Západoaustralský letní čas", "BOT": "Bolivijský čas", "JDT": "Japonský letní čas", "ACDT": "Středoaustralský letní čas", "ACWDT": "Středozápadní australský letní čas", "HNOG": "Západogrónský standardní čas", "HNT": "Newfoundlandský standardní čas", "PDT": "Severoamerický pacifický letní čas", "HNPMX": "Mexický pacifický standardní čas", "HEPM": "Pierre-miquelonský letní čas", "EAT": "Východoafrický čas", "CLST": "Chilský letní čas", "COT": "Kolumbijský standardní čas", "AEST": "Východoaustralský standardní čas", "JST": "Japonský standardní čas", "HKST": "Hongkongský letní čas", "MEZ": "Středoevropský standardní čas", "WARST": "Západoargentinský letní čas", "HADT": "Havajsko-aleutský letní čas", "GFT": "Francouzskoguyanský čas", "WART": "Západoargentinský standardní čas", "CDT": "Severoamerický centrální letní čas", "WESZ": "Západoevropský letní čas", "LHST": "Standardní čas ostrova lorda Howa", "HNPM": "Pierre-miquelonský standardní čas", "HAT": "Newfoundlandský letní čas", "TMST": "Turkmenský letní čas", "CAT": "Středoafrický čas", "HEEG": "Východogrónský letní čas", "LHDT": "Letní čas ostrova lorda Howa", "MDT": "Macajský letní čas", "WIB": "Západoindonéský čas", "EST": "Severoamerický východní standardní čas", "HNNOMX": "Severozápadní mexický standardní čas", "MST": "Macajský standardní čas", "NZDT": "Novozélandský letní čas", "HKT": "Hongkongský standardní čas", "HENOMX": "Severozápadní mexický letní čas", "UYT": "Uruguayský standardní čas", "CHADT": "Chathamský letní čas"},
	}
}

// Locale returns the current translators string locale
func (cs *cs_CZ) Locale() string {
	return cs.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'cs_CZ'
func (cs *cs_CZ) PluralsCardinal() []locales.PluralRule {
	return cs.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'cs_CZ'
func (cs *cs_CZ) PluralsOrdinal() []locales.PluralRule {
	return cs.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'cs_CZ'
func (cs *cs_CZ) PluralsRange() []locales.PluralRule {
	return cs.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'cs_CZ'
func (cs *cs_CZ) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)

	if i == 1 && v == 0 {
		return locales.PluralRuleOne
	} else if i >= 2 && i <= 4 && v == 0 {
		return locales.PluralRuleFew
	} else if v != 0 {
		return locales.PluralRuleMany
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'cs_CZ'
func (cs *cs_CZ) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'cs_CZ'
func (cs *cs_CZ) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := cs.CardinalPluralRule(num1, v1)
	end := cs.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleMany {
		return locales.PluralRuleMany
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleFew && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleFew && end == locales.PluralRuleMany {
		return locales.PluralRuleMany
	} else if start == locales.PluralRuleFew && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleMany && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	} else if start == locales.PluralRuleMany && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleMany && end == locales.PluralRuleMany {
		return locales.PluralRuleMany
	} else if start == locales.PluralRuleMany && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleMany {
		return locales.PluralRuleMany
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (cs *cs_CZ) MonthAbbreviated(month time.Month) string {
	return cs.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (cs *cs_CZ) MonthsAbbreviated() []string {
	return cs.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (cs *cs_CZ) MonthNarrow(month time.Month) string {
	return cs.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (cs *cs_CZ) MonthsNarrow() []string {
	return cs.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (cs *cs_CZ) MonthWide(month time.Month) string {
	return cs.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (cs *cs_CZ) MonthsWide() []string {
	return cs.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (cs *cs_CZ) WeekdayAbbreviated(weekday time.Weekday) string {
	return cs.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (cs *cs_CZ) WeekdaysAbbreviated() []string {
	return cs.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (cs *cs_CZ) WeekdayNarrow(weekday time.Weekday) string {
	return cs.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (cs *cs_CZ) WeekdaysNarrow() []string {
	return cs.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (cs *cs_CZ) WeekdayShort(weekday time.Weekday) string {
	return cs.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (cs *cs_CZ) WeekdaysShort() []string {
	return cs.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (cs *cs_CZ) WeekdayWide(weekday time.Weekday) string {
	return cs.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (cs *cs_CZ) WeekdaysWide() []string {
	return cs.daysWide
}

// Decimal returns the decimal point of number
func (cs *cs_CZ) Decimal() string {
	return cs.decimal
}

// Group returns the group of number
func (cs *cs_CZ) Group() string {
	return cs.group
}

// Group returns the minus sign of number
func (cs *cs_CZ) Minus() string {
	return cs.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'cs_CZ' and handles both Whole and Real numbers based on 'v'
func (cs *cs_CZ) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, cs.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(cs.group) - 1; j >= 0; j-- {
					b = append(b, cs.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, cs.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'cs_CZ' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (cs *cs_CZ) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 5
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, cs.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, cs.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, cs.percentSuffix...)

	b = append(b, cs.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'cs_CZ'
func (cs *cs_CZ) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := cs.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, cs.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(cs.group) - 1; j >= 0; j-- {
					b = append(b, cs.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, cs.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, cs.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, cs.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'cs_CZ'
// in accounting notation.
func (cs *cs_CZ) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := cs.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, cs.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(cs.group) - 1; j >= 0; j-- {
					b = append(b, cs.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, cs.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, cs.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, cs.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, cs.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'cs_CZ'
func (cs *cs_CZ) FmtDateShort(t time.Time) string {

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

	if t.Year() > 9 {
		b = append(b, strconv.Itoa(t.Year())[2:]...)
	} else {
		b = append(b, strconv.Itoa(t.Year())[1:]...)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'cs_CZ'
func (cs *cs_CZ) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0x2e, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'cs_CZ'
func (cs *cs_CZ) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, cs.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'cs_CZ'
func (cs *cs_CZ) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, cs.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, cs.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'cs_CZ'
func (cs *cs_CZ) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, cs.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'cs_CZ'
func (cs *cs_CZ) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, cs.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, cs.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'cs_CZ'
func (cs *cs_CZ) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, cs.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, cs.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'cs_CZ'
func (cs *cs_CZ) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, cs.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, cs.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := cs.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
