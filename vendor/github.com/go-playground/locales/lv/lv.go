package lv

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type lv struct {
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

// New returns a new instance of translator for the 'lv' locale
func New() locales.Translator {
	return &lv{
		locale:                 "lv",
		pluralsCardinal:        []locales.PluralRule{1, 2, 6},
		pluralsOrdinal:         []locales.PluralRule{6},
		pluralsRange:           []locales.PluralRule{2, 6},
		decimal:                ",",
		group:                  " ",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AU$", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "₪", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "₩", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "Ls", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZ$", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "฿", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "NT$", "TZS", "UAH", "UAK", "UGS", "UGX", "$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "₫", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "EC$", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "janv.", "febr.", "marts", "apr.", "maijs", "jūn.", "jūl.", "aug.", "sept.", "okt.", "nov.", "dec."},
		monthsNarrow:           []string{"", "J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"},
		monthsWide:             []string{"", "janvāris", "februāris", "marts", "aprīlis", "maijs", "jūnijs", "jūlijs", "augusts", "septembris", "oktobris", "novembris", "decembris"},
		daysAbbreviated:        []string{"svētd.", "pirmd.", "otrd.", "trešd.", "ceturtd.", "piektd.", "sestd."},
		daysNarrow:             []string{"S", "P", "O", "T", "C", "P", "S"},
		daysShort:              []string{"Sv", "Pr", "Ot", "Tr", "Ce", "Pk", "Se"},
		daysWide:               []string{"svētdiena", "pirmdiena", "otrdiena", "trešdiena", "ceturtdiena", "piektdiena", "sestdiena"},
		periodsAbbreviated:     []string{"priekšp.", "pēcp."},
		periodsNarrow:          []string{"priekšp.", "pēcp."},
		periodsWide:            []string{"priekšpusdienā", "pēcpusdienā"},
		erasAbbreviated:        []string{"p.m.ē.", "m.ē."},
		erasNarrow:             []string{"p.m.ē.", "m.ē."},
		erasWide:               []string{"pirms mūsu ēras", "mūsu ērā"},
		timezones:              map[string]string{"SGT": "Singapūras laiks", "ACWDT": "Austrālijas centrālais rietumu vasaras laiks", "CLST": "Čīles vasaras laiks", "TMT": "Turkmenistānas ziemas laiks", "OESZ": "Austrumeiropas vasaras laiks", "HADT": "Havaju–Aleutu vasaras laiks", "HNCU": "Kubas ziemas laiks", "CST": "Centrālais ziemas laiks", "EST": "Austrumu ziemas laiks", "EDT": "Austrumu vasaras laiks", "ACDT": "Austrālijas centrālais vasaras laiks", "MEZ": "Centrāleiropas ziemas laiks", "JDT": "Japānas vasaras laiks", "HNEG": "Austrumgrenlandes ziemas laiks", "HKST": "Honkongas vasaras laiks", "WIT": "Austrumindonēzijas laiks", "CHAST": "Četemas ziemas laiks", "CHADT": "Četemas vasaras laiks", "SAST": "Dienvidāfrikas ziemas laiks", "SRT": "Surinamas laiks", "COT": "Kolumbijas ziemas laiks", "BT": "Butānas laiks", "GFT": "Francijas Gviānas laiks", "WART": "Rietumargentīnas ziemas laiks", "HNNOMX": "Ziemeļrietumu Meksikas ziemas laiks", "OEZ": "Austrumeiropas ziemas laiks", "GMT": "Griničas laiks", "AEST": "Austrālijas austrumu ziemas laiks", "HENOMX": "Ziemeļrietumu Meksikas vasaras laiks", "ART": "Argentīnas ziemas laiks", "GYT": "Gajānas laiks", "WAT": "Rietumāfrikas ziemas laiks", "MYT": "Malaizijas laiks", "HNOG": "Rietumgrenlandes ziemas laiks", "HKT": "Honkongas ziemas laiks", "HNT": "Ņūfaundlendas ziemas laiks", "HEPMX": "Meksikas Klusā okeāna piekrastes vasaras laiks", "MESZ": "Centrāleiropas vasaras laiks", "HAT": "Ņūfaundlendas vasaras laiks", "HEPM": "Senpjēras un Mikelonas vasaras laiks", "ChST": "Čamorra ziemas laiks", "CDT": "Centrālais vasaras laiks", "WAST": "Rietumāfrikas vasaras laiks", "NZST": "Jaunzēlandes ziemas laiks", "ACWST": "Austrālijas centrālais rietumu ziemas laiks", "∅∅∅": "Azoru salu vasaras laiks", "EAT": "Austrumāfrikas laiks", "AWST": "Austrālijas rietumu ziemas laiks", "HNPMX": "Meksikas Klusā okeāna piekrastes ziemas laiks", "BOT": "Bolīvijas laiks", "ACST": "Austrālijas centrālais ziemas laiks", "LHST": "Lorda Hava salas ziemas laiks", "LHDT": "Lorda Hava salas vasaras laiks", "WITA": "Centrālindonēzijas laiks", "CLT": "Čīles ziemas laiks", "NZDT": "Jaunzēlandes vasaras laiks", "JST": "Japānas ziemas laiks", "ECT": "Ekvadoras laiks", "HNPM": "Senpjēras un Mikelonas ziemas laiks", "MDT": "MDT", "HAST": "Havaju–Aleutu ziemas laiks", "PST": "Klusā okeāna ziemas laiks", "AST": "Atlantijas ziemas laiks", "AKST": "Aļaskas ziemas laiks", "IST": "Indijas ziemas laiks", "WARST": "Rietumargentīnas vasaras laiks", "VET": "Venecuēlas laiks", "UYT": "Urugvajas ziemas laiks", "AKDT": "Aļaskas vasaras laiks", "CAT": "Centrālāfrikas laiks", "TMST": "Turkmenistānas vasaras laiks", "UYST": "Urugvajas vasaras laiks", "ADT": "Atlantijas vasaras laiks", "WEZ": "Rietumeiropas ziemas laiks", "WESZ": "Rietumeiropas vasaras laiks", "WIB": "Rietumindonēzijas laiks", "AEDT": "Austrālijas austrumu vasaras laiks", "HEEG": "Austrumgrenlandes vasaras laiks", "HEOG": "Rietumgrenlandes vasaras laiks", "MST": "MST", "ARST": "Argentīnas vasaras laiks", "COST": "Kolumbijas vasaras laiks", "HECU": "Kubas vasaras laiks", "PDT": "Klusā okeāna vasaras laiks", "AWDT": "Austrālijas rietumu vasaras laiks"},
	}
}

// Locale returns the current translators string locale
func (lv *lv) Locale() string {
	return lv.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'lv'
func (lv *lv) PluralsCardinal() []locales.PluralRule {
	return lv.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'lv'
func (lv *lv) PluralsOrdinal() []locales.PluralRule {
	return lv.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'lv'
func (lv *lv) PluralsRange() []locales.PluralRule {
	return lv.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'lv'
func (lv *lv) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	f := locales.F(n, v)
	nMod10 := math.Mod(n, 10)
	nMod100 := math.Mod(n, 100)
	fMod100 := f % 100
	fMod10 := f % 10

	if (nMod10 == 0) || (nMod100 >= 11 && nMod100 <= 19) || (v == 2 && fMod100 >= 11 && fMod100 <= 19) {
		return locales.PluralRuleZero
	} else if (nMod10 == 1 && nMod100 != 11) || (v == 2 && fMod10 == 1 && fMod100 != 11) || (v != 2 && fMod10 == 1) {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'lv'
func (lv *lv) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'lv'
func (lv *lv) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := lv.CardinalPluralRule(num1, v1)
	end := lv.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleZero && end == locales.PluralRuleZero {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleZero && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	} else if start == locales.PluralRuleZero && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleZero {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleZero {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (lv *lv) MonthAbbreviated(month time.Month) string {
	return lv.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (lv *lv) MonthsAbbreviated() []string {
	return lv.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (lv *lv) MonthNarrow(month time.Month) string {
	return lv.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (lv *lv) MonthsNarrow() []string {
	return lv.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (lv *lv) MonthWide(month time.Month) string {
	return lv.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (lv *lv) MonthsWide() []string {
	return lv.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (lv *lv) WeekdayAbbreviated(weekday time.Weekday) string {
	return lv.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (lv *lv) WeekdaysAbbreviated() []string {
	return lv.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (lv *lv) WeekdayNarrow(weekday time.Weekday) string {
	return lv.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (lv *lv) WeekdaysNarrow() []string {
	return lv.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (lv *lv) WeekdayShort(weekday time.Weekday) string {
	return lv.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (lv *lv) WeekdaysShort() []string {
	return lv.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (lv *lv) WeekdayWide(weekday time.Weekday) string {
	return lv.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (lv *lv) WeekdaysWide() []string {
	return lv.daysWide
}

// Decimal returns the decimal point of number
func (lv *lv) Decimal() string {
	return lv.decimal
}

// Group returns the group of number
func (lv *lv) Group() string {
	return lv.group
}

// Group returns the minus sign of number
func (lv *lv) Minus() string {
	return lv.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'lv' and handles both Whole and Real numbers based on 'v'
func (lv *lv) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, lv.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(lv.group) - 1; j >= 0; j-- {
					b = append(b, lv.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, lv.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'lv' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (lv *lv) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, lv.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, lv.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, lv.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'lv'
func (lv *lv) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := lv.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, lv.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(lv.group) - 1; j >= 0; j-- {
					b = append(b, lv.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, lv.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, lv.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, lv.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'lv'
// in accounting notation.
func (lv *lv) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := lv.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, lv.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(lv.group) - 1; j >= 0; j-- {
					b = append(b, lv.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, lv.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, lv.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, lv.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, lv.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'lv'
func (lv *lv) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'lv'
func (lv *lv) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2e, 0x20, 0x67, 0x61, 0x64, 0x61}...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, lv.monthsAbbreviated[t.Month()]...)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'lv'
func (lv *lv) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2e, 0x20, 0x67, 0x61, 0x64, 0x61}...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, lv.monthsWide[t.Month()]...)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'lv'
func (lv *lv) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, lv.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2e, 0x20, 0x67, 0x61, 0x64, 0x61}...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, lv.monthsWide[t.Month()]...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'lv'
func (lv *lv) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, lv.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'lv'
func (lv *lv) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, lv.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, lv.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'lv'
func (lv *lv) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, lv.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, lv.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'lv'
func (lv *lv) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, lv.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, lv.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := lv.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
