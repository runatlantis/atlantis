package tk_TM

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type tk_TM struct {
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

// New returns a new instance of translator for the 'tk_TM' locale
func New() locales.Translator {
	return &tk_TM{
		locale:                 "tk_TM",
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
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
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
		timezones:              map[string]string{"WAST": "Günbatar Afrika tomusky wagty", "ACDT": "Merkezi Awstraliýa tomusky wagty", "WARST": "Günbatar Argentina tomusky wagty", "OEZ": "Gündogar Ýewropa standart wagty", "HADT": "Gawaý-Aleut tomusky wagty", "CST": "Merkezi Amerika standart wagty", "WIB": "Günbatar Indoneziýa wagty", "SGT": "Singapur wagty", "WART": "Günbatar Argentina standart wagty", "ACWDT": "Merkezi Awstraliýa günbatar tomusky wagty", "HEEG": "Gündogar Grenlandiýa tomusky wagty", "VET": "Wenesuela wagty", "HNNOMX": "Demirgazyk-günbatar Meksika standart wagty", "TMST": "Türkmenistan tomusky wagty", "UYST": "Urugwaý tomusky wagty", "CHADT": "Çatem tomusky wagty", "HNPMX": "Meksikan Ýuwaş umman standart wagty", "SAST": "Günorta Afrika standart wagty", "∅∅∅": "∅∅∅", "ACST": "Merkezi Awstraliýa standart wagty", "LHST": "Lord-Hau standart wagty", "HENOMX": "Demirgazyk-günbatar Meksika tomusky wagty", "COT": "Kolumbiýa standart wagty", "GFT": "Fransuz Gwianasy wagty", "AKDT": "Alýaska tomusky wagty", "IST": "Hindistan standart wagty", "WITA": "Merkezi Indoneziýa wagty", "TMT": "Türkmenistan standart wagty", "ARST": "Argentina tomusky wagty", "ChST": "Çamorro wagty", "AWDT": "Günbatar Awstraliýa tomusky wagty", "AST": "Atlantik standart wagty", "ADT": "Atlantik tomusky wagty", "HAT": "Nýufaundlend tomusky wagty", "HEPM": "Sen-Pýer we Mikelon tomusky wagty", "PDT": "Demirgazyk Amerika Ýuwaş umman tomusky wagty", "JDT": "Ýaponiýa tomusky wagty", "HAST": "Gawaý-Aleut standart wagty", "WEZ": "Günbatar Ýewropa standart wagty", "NZST": "Täze Zelandiýa standart wagty", "CLT": "Çili standart wagty", "CLST": "Çili tomusky wagty", "WESZ": "Günbatar Ýewropa tomusky wagty", "ECT": "Ekwador wagty", "HNCU": "Kuba standart wagty", "HECU": "Kuba tomusky wagty", "MYT": "Malaýziýa wagty", "EAT": "Gündogar Afrika wagty", "PST": "Demirgazyk Amerika Ýuwaş umman standart wagty", "BT": "Butan wagty", "EST": "Demirgazyk Amerika gündogar standart wagty", "LHDT": "Lord-Hau tomusky wagty", "HNT": "Nýufaundlend standart wagty", "CAT": "Merkezi Afrika wagty", "ART": "Argentina standart wagty", "COST": "Kolumbiýa tomusky wagty", "HEPMX": "Meksikan Ýuwaş umman tomusky wagty", "BOT": "Boliwiýa wagty", "ACWST": "Merkezi Awstraliýa günbatar standart wagty", "MEZ": "Merkezi Ýewropa standart wagty", "MDT": "MDT", "OESZ": "Gündogar Ýewropa tomusky wagty", "NZDT": "Täze Zelandiýa tomusky wagty", "EDT": "Demirgazyk Amerika gündogar tomusky wagty", "AEST": "Gündogar Awstraliýa standart wagty", "WAT": "Günbatar Afrika standart wagty", "HNEG": "Gündogar Grenlandiýa standart wagty", "HNOG": "Günbatar Grenlandiýa standart wagty", "HKST": "Gonkong tomusky wagty", "SRT": "Surinam wagty", "GMT": "Grinwiç boýunça orta wagt", "GYT": "Gaýana wagty", "CDT": "Merkezi Amerika tomusky wagty", "AWST": "Günbatar Awstraliýa standart wagty", "JST": "Ýaponiýa standart wagty", "HEOG": "Günbatar Grenlandiýa tomusky wagty", "MESZ": "Merkezi Ýewropa tomusky wagty", "HKT": "Gonkong standart wagty", "MST": "MST", "UYT": "Urugwaý standart wagty", "AKST": "Alýaska standart wagty", "HNPM": "Sen-Pýer we Mikelon standart wagty", "WIT": "Gündogar Indoneziýa wagty", "CHAST": "Çatem standart wagty", "AEDT": "Gündogar Awstraliýa tomusky wagty"},
	}
}

// Locale returns the current translators string locale
func (tk *tk_TM) Locale() string {
	return tk.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'tk_TM'
func (tk *tk_TM) PluralsCardinal() []locales.PluralRule {
	return tk.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'tk_TM'
func (tk *tk_TM) PluralsOrdinal() []locales.PluralRule {
	return tk.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'tk_TM'
func (tk *tk_TM) PluralsRange() []locales.PluralRule {
	return tk.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'tk_TM'
func (tk *tk_TM) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'tk_TM'
func (tk *tk_TM) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	nMod10 := math.Mod(n, 10)

	if (nMod10 == 6 || nMod10 == 9) || (n == 10) {
		return locales.PluralRuleFew
	}

	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'tk_TM'
func (tk *tk_TM) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

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
func (tk *tk_TM) MonthAbbreviated(month time.Month) string {
	return tk.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (tk *tk_TM) MonthsAbbreviated() []string {
	return tk.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (tk *tk_TM) MonthNarrow(month time.Month) string {
	return tk.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (tk *tk_TM) MonthsNarrow() []string {
	return tk.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (tk *tk_TM) MonthWide(month time.Month) string {
	return tk.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (tk *tk_TM) MonthsWide() []string {
	return tk.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (tk *tk_TM) WeekdayAbbreviated(weekday time.Weekday) string {
	return tk.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (tk *tk_TM) WeekdaysAbbreviated() []string {
	return tk.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (tk *tk_TM) WeekdayNarrow(weekday time.Weekday) string {
	return tk.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (tk *tk_TM) WeekdaysNarrow() []string {
	return tk.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (tk *tk_TM) WeekdayShort(weekday time.Weekday) string {
	return tk.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (tk *tk_TM) WeekdaysShort() []string {
	return tk.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (tk *tk_TM) WeekdayWide(weekday time.Weekday) string {
	return tk.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (tk *tk_TM) WeekdaysWide() []string {
	return tk.daysWide
}

// Decimal returns the decimal point of number
func (tk *tk_TM) Decimal() string {
	return tk.decimal
}

// Group returns the group of number
func (tk *tk_TM) Group() string {
	return tk.group
}

// Group returns the minus sign of number
func (tk *tk_TM) Minus() string {
	return tk.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'tk_TM' and handles both Whole and Real numbers based on 'v'
func (tk *tk_TM) FmtNumber(num float64, v uint64) string {

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

// FmtPercent returns 'num' with digits/precision of 'v' for 'tk_TM' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (tk *tk_TM) FmtPercent(num float64, v uint64) string {
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

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'tk_TM'
func (tk *tk_TM) FmtCurrency(num float64, v uint64, currency currency.Type) string {

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

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'tk_TM'
// in accounting notation.
func (tk *tk_TM) FmtAccounting(num float64, v uint64, currency currency.Type) string {

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

// FmtDateShort returns the short date representation of 't' for 'tk_TM'
func (tk *tk_TM) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'tk_TM'
func (tk *tk_TM) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'tk_TM'
func (tk *tk_TM) FmtDateLong(t time.Time) string {

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

// FmtDateFull returns the full date representation of 't' for 'tk_TM'
func (tk *tk_TM) FmtDateFull(t time.Time) string {

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

// FmtTimeShort returns the short time representation of 't' for 'tk_TM'
func (tk *tk_TM) FmtTimeShort(t time.Time) string {

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

// FmtTimeMedium returns the medium time representation of 't' for 'tk_TM'
func (tk *tk_TM) FmtTimeMedium(t time.Time) string {

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

// FmtTimeLong returns the long time representation of 't' for 'tk_TM'
func (tk *tk_TM) FmtTimeLong(t time.Time) string {

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

// FmtTimeFull returns the full time representation of 't' for 'tk_TM'
func (tk *tk_TM) FmtTimeFull(t time.Time) string {

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
