package sk

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type sk struct {
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
	currencyNegativePrefix string
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

// New returns a new instance of translator for the 'sk' locale
func New() locales.Translator {
	return &sk{
		locale:                 "sk",
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
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "NIS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "EC$", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		percentSuffix:          " ",
		currencyPositiveSuffix: " ",
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: " )",
		monthsAbbreviated:      []string{"", "jan", "feb", "mar", "apr", "máj", "jún", "júl", "aug", "sep", "okt", "nov", "dec"},
		monthsNarrow:           []string{"", "j", "f", "m", "a", "m", "j", "j", "a", "s", "o", "n", "d"},
		monthsWide:             []string{"", "januára", "februára", "marca", "apríla", "mája", "júna", "júla", "augusta", "septembra", "októbra", "novembra", "decembra"},
		daysAbbreviated:        []string{"ne", "po", "ut", "st", "št", "pi", "so"},
		daysNarrow:             []string{"n", "p", "u", "s", "š", "p", "s"},
		daysShort:              []string{"ne", "po", "ut", "st", "št", "pi", "so"},
		daysWide:               []string{"nedeľa", "pondelok", "utorok", "streda", "štvrtok", "piatok", "sobota"},
		periodsAbbreviated:     []string{"AM", "PM"},
		periodsNarrow:          []string{"AM", "PM"},
		periodsWide:            []string{"AM", "PM"},
		erasAbbreviated:        []string{"pred Kr.", "po Kr."},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"pred Kristom", "po Kristovi"},
		timezones:              map[string]string{"WIT": "východoindonézsky čas", "UYST": "uruguajský letný čas", "ECT": "ekvádorský čas", "EDT": "severoamerický východný letný čas", "ACWST": "stredozápadný austrálsky štandardný čas", "IST": "indický čas", "WART": "západoargentínsky štandardný čas", "WARST": "západoargentínsky letný čas", "CST": "severoamerický centrálny štandardný čas", "AST": "atlantický štandardný čas", "GMT": "greenwichský čas", "AWDT": "západoaustrálsky letný čas", "ACWDT": "stredozápadný austrálsky letný čas", "HEEG": "východogrónsky letný čas", "HNT": "newfoundlandský štandardný čas", "MST": "macajský štandardný čas", "SRT": "surinamský čas", "TMT": "turkménsky štandardný čas", "WEZ": "západoeurópsky štandardný čas", "AKDT": "aljašský letný čas", "EST": "severoamerický východný štandardný čas", "HKT": "hongkonský štandardný čas", "BOT": "bolívijský čas", "WAT": "západoafrický štandardný čas", "JST": "japonský štandardný čas", "HEOG": "západogrónsky letný čas", "HNPM": "pierre-miquelonský štandardný čas", "MDT": "macajský letný čas", "CLT": "čilský štandardný čas", "TMST": "turkménsky letný čas", "ChST": "chamorrský čas", "AWST": "západoaustrálsky štandardný čas", "ADT": "atlantický letný čas", "HNNOMX": "severozápadný mexický štandardný čas", "HADT": "havajsko-aleutský letný čas", "WAST": "západoafrický letný čas", "COT": "kolumbijský štandardný čas", "CHAST": "chathamský štandardný čas", "ACDT": "stredoaustrálsky letný čas", "HNOG": "západogrónsky štandardný čas", "WITA": "stredoindonézsky čas", "CLST": "čilský letný čas", "OESZ": "východoeurópsky letný čas", "HAST": "havajsko-aleutský štandardný čas", "WESZ": "západoeurópsky letný čas", "WIB": "západoindonézsky čas", "AKST": "aljašský štandardný čas", "∅∅∅": "azorský letný čas", "HAT": "newfoundlandský letný čas", "OEZ": "východoeurópsky štandardný čas", "PST": "severoamerický tichomorský štandardný čas", "HNPMX": "mexický tichomorský štandardný čas", "ARST": "argentínsky letný čas", "HENOMX": "severozápadný mexický letný čas", "CAT": "stredoafrický čas", "PDT": "severoamerický tichomorský letný čas", "JDT": "japonský letný čas", "MEZ": "stredoeurópsky štandardný čas", "LHDT": "letný čas ostrova lorda Howa", "COST": "kolumbijský letný čas", "GYT": "guyanský čas", "CHADT": "chathamský letný čas", "BT": "bhutánsky čas", "NZDT": "novozélandský letný čas", "ACST": "stredoaustrálsky štandardný čas", "HKST": "hongkonský letný čas", "HEPM": "pierre-miquelonský letný čas", "EAT": "východoafrický čas", "AEST": "východoaustrálsky štandardný čas", "AEDT": "východoaustrálsky letný čas", "HEPMX": "mexický tichomorský letný čas", "SAST": "juhoafrický čas", "SGT": "singapurský štandardný čas", "HNEG": "východogrónsky štandardný čas", "MESZ": "stredoeurópsky letný čas", "LHST": "štandardný čas ostrova lorda Howa", "UYT": "uruguajský štandardný čas", "HECU": "kubánsky letný čas", "NZST": "novozélandský štandardný čas", "VET": "venezuelský čas", "ART": "argentínsky štandardný čas", "HNCU": "kubánsky štandardný čas", "CDT": "severoamerický centrálny letný čas", "MYT": "malajzijský čas", "GFT": "francúzskoguyanský čas"},
	}
}

// Locale returns the current translators string locale
func (sk *sk) Locale() string {
	return sk.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'sk'
func (sk *sk) PluralsCardinal() []locales.PluralRule {
	return sk.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'sk'
func (sk *sk) PluralsOrdinal() []locales.PluralRule {
	return sk.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'sk'
func (sk *sk) PluralsRange() []locales.PluralRule {
	return sk.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'sk'
func (sk *sk) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

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

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'sk'
func (sk *sk) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'sk'
func (sk *sk) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := sk.CardinalPluralRule(num1, v1)
	end := sk.CardinalPluralRule(num2, v2)

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
func (sk *sk) MonthAbbreviated(month time.Month) string {
	return sk.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (sk *sk) MonthsAbbreviated() []string {
	return sk.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (sk *sk) MonthNarrow(month time.Month) string {
	return sk.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (sk *sk) MonthsNarrow() []string {
	return sk.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (sk *sk) MonthWide(month time.Month) string {
	return sk.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (sk *sk) MonthsWide() []string {
	return sk.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (sk *sk) WeekdayAbbreviated(weekday time.Weekday) string {
	return sk.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (sk *sk) WeekdaysAbbreviated() []string {
	return sk.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (sk *sk) WeekdayNarrow(weekday time.Weekday) string {
	return sk.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (sk *sk) WeekdaysNarrow() []string {
	return sk.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (sk *sk) WeekdayShort(weekday time.Weekday) string {
	return sk.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (sk *sk) WeekdaysShort() []string {
	return sk.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (sk *sk) WeekdayWide(weekday time.Weekday) string {
	return sk.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (sk *sk) WeekdaysWide() []string {
	return sk.daysWide
}

// Decimal returns the decimal point of number
func (sk *sk) Decimal() string {
	return sk.decimal
}

// Group returns the group of number
func (sk *sk) Group() string {
	return sk.group
}

// Group returns the minus sign of number
func (sk *sk) Minus() string {
	return sk.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'sk' and handles both Whole and Real numbers based on 'v'
func (sk *sk) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sk.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(sk.group) - 1; j >= 0; j-- {
					b = append(b, sk.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, sk.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'sk' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (sk *sk) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 5
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sk.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, sk.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, sk.percentSuffix...)

	b = append(b, sk.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'sk'
func (sk *sk) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := sk.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sk.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(sk.group) - 1; j >= 0; j-- {
					b = append(b, sk.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, sk.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, sk.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, sk.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'sk'
// in accounting notation.
func (sk *sk) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := sk.currencies[currency]
	l := len(s) + len(symbol) + 6 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sk.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(sk.group) - 1; j >= 0; j-- {
					b = append(b, sk.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, sk.currencyNegativePrefix[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, sk.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, sk.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, sk.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'sk'
func (sk *sk) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'sk'
func (sk *sk) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'sk'
func (sk *sk) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, sk.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'sk'
func (sk *sk) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, sk.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, sk.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'sk'
func (sk *sk) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sk.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'sk'
func (sk *sk) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sk.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, sk.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'sk'
func (sk *sk) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sk.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, sk.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'sk'
func (sk *sk) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sk.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, sk.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := sk.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
