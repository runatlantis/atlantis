package tk

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type tk struct {
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

// New returns a new instance of translator for the 'tk' locale
func New() locales.Translator {
	return &tk{
		locale:                 "tk",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         []locales.PluralRule{4, 6},
		pluralsRange:           []locales.PluralRule{2, 6},
		decimal:                ",",
		group:                  " ",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "A$", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "₪", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JP¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "₩", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZ$", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "NT$", "TZS", "UAH", "UAK", "UGS", "UGX", "US$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "₫", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "EC$", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		percentSuffix:          " ",
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "ýan", "few", "mart", "apr", "maý", "iýun", "iýul", "awg", "sen", "okt", "noý", "dek"},
		monthsNarrow:           []string{"", "Ý", "F", "M", "A", "M", "I", "I", "A", "S", "O", "N", "D"},
		monthsWide:             []string{"", "ýanwar", "fewral", "mart", "aprel", "maý", "iýun", "iýul", "awgust", "sentýabr", "oktýabr", "noýabr", "dekabr"},
		daysAbbreviated:        []string{"ýek", "duş", "siş", "çar", "pen", "ann", "şen"},
		daysNarrow:             []string{"Ý", "D", "S", "Ç", "P", "A", "Ş"},
		daysShort:              []string{"ýb", "db", "sb", "çb", "pb", "an", "şb"},
		daysWide:               []string{"ýekşenbe", "duşenbe", "sişenbe", "çarşenbe", "penşenbe", "anna", "şenbe"},
		periodsAbbreviated:     []string{"go.öň", "go.soň"},
		periodsNarrow:          []string{"öň", "soň"},
		periodsWide:            []string{"günortadan öň", "günortadan soň"},
		erasAbbreviated:        []string{"B.e.öň", "B.e."},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"Isadan öň", "Isadan soň"},
		timezones:              map[string]string{"TMT": "Türkmenistan standart wagty", "HADT": "Gawaý-Aleut tomusky wagty", "WIB": "Günbatar Indoneziýa wagty", "AKST": "Alýaska standart wagty", "MESZ": "Merkezi Ýewropa tomusky wagty", "CLT": "Çili standart wagty", "WIT": "Gündogar Indoneziýa wagty", "GFT": "Fransuz Gwianasy wagty", "HKST": "Gonkong tomusky wagty", "LHDT": "Lord-Hau tomusky wagty", "UYT": "Urugwaý standart wagty", "ChST": "Çamorro wagty", "∅∅∅": "Braziliýa tomusky wagty", "HNCU": "Kuba standart wagty", "CST": "Merkezi Amerika standart wagty", "ACST": "Merkezi Awstraliýa standart wagty", "SRT": "Surinam wagty", "HAT": "Nýufaundlend tomusky wagty", "HENOMX": "Demirgazyk-günbatar Meksika tomusky wagty", "TMST": "Türkmenistan tomusky wagty", "OESZ": "Gündogar Ýewropa tomusky wagty", "SAST": "Günorta Afrika standart wagty", "ECT": "Ekwador wagty", "ACWST": "Merkezi Awstraliýa günbatar standart wagty", "CLST": "Çili tomusky wagty", "COST": "Kolumbiýa tomusky wagty", "WESZ": "Günbatar Ýewropa tomusky wagty", "WAT": "Günbatar Afrika standart wagty", "JDT": "Ýaponiýa tomusky wagty", "AWST": "Günbatar Awstraliýa standart wagty", "AWDT": "Günbatar Awstraliýa tomusky wagty", "MDT": "Demirgazyk Amerika dag tomusky wagty", "AST": "Atlantik standart wagty", "AEST": "Gündogar Awstraliýa standart wagty", "ACWDT": "Merkezi Awstraliýa günbatar tomusky wagty", "EAT": "Gündogar Afrika wagty", "HNNOMX": "Demirgazyk-günbatar Meksika standart wagty", "ART": "Argentina standart wagty", "MST": "Demirgazyk Amerika dag standart wagty", "BT": "Butan wagty", "HNEG": "Gündogar Grenlandiýa standart wagty", "HNOG": "Günbatar Grenlandiýa standart wagty", "HNT": "Nýufaundlend standart wagty", "LHST": "Lord-Hau standart wagty", "NZDT": "Täze Zelandiýa tomusky wagty", "EDT": "Demirgazyk Amerika gündogar tomusky wagty", "CHAST": "Çatem standart wagty", "CHADT": "Çatem tomusky wagty", "WAST": "Günbatar Afrika tomusky wagty", "MYT": "Malaýziýa wagty", "JST": "Ýaponiýa standart wagty", "HECU": "Kuba tomusky wagty", "CDT": "Merkezi Amerika tomusky wagty", "BOT": "Boliwiýa wagty", "HEPM": "Sen-Pýer we Mikelon tomusky wagty", "ARST": "Argentina tomusky wagty", "WEZ": "Günbatar Ýewropa standart wagty", "AKDT": "Alýaska tomusky wagty", "HEOG": "Günbatar Grenlandiýa tomusky wagty", "VET": "Wenesuela wagty", "COT": "Kolumbiýa standart wagty", "GMT": "Grinwiç boýunça orta wagt", "PDT": "Demirgazyk Amerika Ýuwaş umman tomusky wagty", "NZST": "Täze Zelandiýa standart wagty", "EST": "Demirgazyk Amerika gündogar standart wagty", "HNPM": "Sen-Pýer we Mikelon standart wagty", "OEZ": "Gündogar Ýewropa standart wagty", "SGT": "Singapur wagty", "CAT": "Merkezi Afrika wagty", "GYT": "Gaýana wagty", "ADT": "Atlantik tomusky wagty", "MEZ": "Merkezi Ýewropa standart wagty", "WART": "Günbatar Argentina standart wagty", "WARST": "Günbatar Argentina tomusky wagty", "UYST": "Urugwaý tomusky wagty", "HEPMX": "Meksikan Ýuwaş umman tomusky wagty", "PST": "Demirgazyk Amerika Ýuwaş umman standart wagty", "HEEG": "Gündogar Grenlandiýa tomusky wagty", "ACDT": "Merkezi Awstraliýa tomusky wagty", "IST": "Hindistan standart wagty", "HNPMX": "Meksikan Ýuwaş umman standart wagty", "AEDT": "Gündogar Awstraliýa tomusky wagty", "HKT": "Gonkong standart wagty", "WITA": "Merkezi Indoneziýa wagty", "HAST": "Gawaý-Aleut standart wagty"},
	}
}

// Locale returns the current translators string locale
func (tk *tk) Locale() string {
	return tk.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'tk'
func (tk *tk) PluralsCardinal() []locales.PluralRule {
	return tk.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'tk'
func (tk *tk) PluralsOrdinal() []locales.PluralRule {
	return tk.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'tk'
func (tk *tk) PluralsRange() []locales.PluralRule {
	return tk.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'tk'
func (tk *tk) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'tk'
func (tk *tk) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	nMod10 := math.Mod(n, 10)

	if (nMod10 == 6 || nMod10 == 9) || (n == 10) {
		return locales.PluralRuleFew
	}

	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'tk'
func (tk *tk) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := tk.CardinalPluralRule(num1, v1)
	end := tk.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (tk *tk) MonthAbbreviated(month time.Month) string {
	return tk.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (tk *tk) MonthsAbbreviated() []string {
	return tk.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (tk *tk) MonthNarrow(month time.Month) string {
	return tk.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (tk *tk) MonthsNarrow() []string {
	return tk.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (tk *tk) MonthWide(month time.Month) string {
	return tk.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (tk *tk) MonthsWide() []string {
	return tk.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (tk *tk) WeekdayAbbreviated(weekday time.Weekday) string {
	return tk.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (tk *tk) WeekdaysAbbreviated() []string {
	return tk.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (tk *tk) WeekdayNarrow(weekday time.Weekday) string {
	return tk.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (tk *tk) WeekdaysNarrow() []string {
	return tk.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (tk *tk) WeekdayShort(weekday time.Weekday) string {
	return tk.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (tk *tk) WeekdaysShort() []string {
	return tk.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (tk *tk) WeekdayWide(weekday time.Weekday) string {
	return tk.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (tk *tk) WeekdaysWide() []string {
	return tk.daysWide
}

// Decimal returns the decimal point of number
func (tk *tk) Decimal() string {
	return tk.decimal
}

// Group returns the group of number
func (tk *tk) Group() string {
	return tk.group
}

// Group returns the minus sign of number
func (tk *tk) Minus() string {
	return tk.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'tk' and handles both Whole and Real numbers based on 'v'
func (tk *tk) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, tk.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(tk.group) - 1; j >= 0; j-- {
					b = append(b, tk.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, tk.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'tk' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (tk *tk) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 5
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, tk.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, tk.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, tk.percentSuffix...)

	b = append(b, tk.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'tk'
func (tk *tk) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := tk.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, tk.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(tk.group) - 1; j >= 0; j-- {
					b = append(b, tk.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, tk.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, tk.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, tk.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'tk'
// in accounting notation.
func (tk *tk) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := tk.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, tk.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(tk.group) - 1; j >= 0; j-- {
					b = append(b, tk.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, tk.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, tk.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, tk.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, tk.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'tk'
func (tk *tk) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'tk'
func (tk *tk) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, tk.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'tk'
func (tk *tk) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, tk.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'tk'
func (tk *tk) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, tk.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, tk.daysWide[t.Weekday()]...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'tk'
func (tk *tk) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, tk.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'tk'
func (tk *tk) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, tk.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, tk.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'tk'
func (tk *tk) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, tk.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, tk.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'tk'
func (tk *tk) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, tk.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, tk.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := tk.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
