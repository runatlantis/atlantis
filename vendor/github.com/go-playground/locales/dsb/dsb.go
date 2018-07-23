package dsb

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type dsb struct {
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

// New returns a new instance of translator for the 'dsb' locale
func New() locales.Translator {
	return &dsb{
		locale:                 "dsb",
		pluralsCardinal:        []locales.PluralRule{2, 3, 4, 6},
		pluralsOrdinal:         []locales.PluralRule{6},
		pluralsRange:           nil,
		decimal:                ",",
		group:                  ".",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "₪", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "₩", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZ$", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "zł", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "฿", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "NT$", "TZS", "UAH", "UAK", "UGS", "UGX", "$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "₫", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "EC$", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		percentSuffix:          " ",
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "jan.", "feb.", "měr.", "apr.", "maj.", "jun.", "jul.", "awg.", "sep.", "okt.", "now.", "dec."},
		monthsNarrow:           []string{"", "j", "f", "m", "a", "m", "j", "j", "a", "s", "o", "n", "d"},
		monthsWide:             []string{"", "januara", "februara", "měrca", "apryla", "maja", "junija", "julija", "awgusta", "septembra", "oktobra", "nowembra", "decembra"},
		daysAbbreviated:        []string{"nje", "pón", "wał", "srj", "stw", "pět", "sob"},
		daysNarrow:             []string{"n", "p", "w", "s", "s", "p", "s"},
		daysShort:              []string{"nj", "pó", "wa", "sr", "st", "pě", "so"},
		daysWide:               []string{"njeźela", "pónjeźele", "wałtora", "srjoda", "stwórtk", "pětk", "sobota"},
		periodsAbbreviated:     []string{"dopołdnja", "wótpołdnja"},
		periodsNarrow:          []string{"dop.", "wótp."},
		periodsWide:            []string{"dopołdnja", "wótpołdnja"},
		erasAbbreviated:        []string{"pś.Chr.n.", "pó Chr.n."},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"pśed Kristusowym naroźenim", "pó Kristusowem naroźenju"},
		timezones:              map[string]string{"CLT": "Chilski standardny cas", "HAST": "Hawaiisko-aleutski standardny cas", "LHDT": "lěśojski cas kupy Lord-Howe", "WART": "Pódwjacornoargentinski standardny cas", "HNT": "Nowofundlandski standardny cas", "VET": "Venezuelski cas", "JST": "Japański standardny cas", "ACWDT": "Srjejźopódwjacorny awstralski lěśojski cas", "EST": "Pódpołnocnoameriski pódzajtšny standardny cas", "ACDT": "Srjejźoawstralski lěśojski cas", "HKST": "Hongkongski lěśojski cas", "MDT": "MDT", "CAT": "Srjejźoafriski cas", "UYST": "Uruguayski lěśojski cas", "HNCU": "Kubański standardny cas", "COT": "Kolumbiski standardny cas", "AWDT": "Pódwjacornoawstralski lěśojski cas", "HNPMX": "Mexiski pacifiski standardny cas", "WARST": "Pódwjacornoargentinski lěśojski cas", "CHAST": "Chathamski standardny cas", "CST": "Pódpołnocnoameriski centralny standardny cas", "CHADT": "Chathamski lěśojski cas", "SAST": "Pódpołdnjowoafriski cas", "GFT": "Francojskoguyański cas", "HNEG": "Pódzajtšnogrönlandski standardny cas", "LHST": "Standardny cas kupy Lord-Howe", "HEPM": "St.-Pierre-a-Miqueloński lěśojski cas", "EAT": "Pódzajtšnoafriski cas", "OESZ": "Pódzajtšnoeuropski lěśojski cas", "MESZ": "Srjejźoeuropski lěśojski cas", "NZST": "Nowoseelandski standardny cas", "HNPM": "St.-Pierre-a-Miqueloński standardny cas", "WIT": "Pódzajtšnoindoneski", "AST": "Atlantiski standardny cas", "AEST": "Pódzajtšnoawstralski standardny cas", "MST": "MST", "PDT": "Pódpołnocnoameriski pacifiski lěśojski cas", "ADT": "Atlantiski lěśojski cas", "EDT": "Pódpołnocnoameriski pódzajtšny lěśojski cas", "WAST": "Pódwjacornoafriski lěśojski cas", "NZDT": "Nowoseelandski lěśojski cas", "SGT": "Singapurski cas", "HNNOMX": "Mexiski dłujkowjacorny standardny cas", "TMT": "Turkmeniski standardny cas", "HADT": "Hawaiisko-aleutski lěśojski cas", "COST": "Kolumbiski lěśojski cas", "ChST": "Chamorrski cas", "JDT": "Japański lěśojski cas", "∅∅∅": "Peruski lěśojski cas", "CLST": "Chilski lěśojski cas", "OEZ": "Pódzajtšnoeuropski standardny cas", "GMT": "Greenwichski cas", "GYT": "Guyański cas", "UYT": "Uruguayski standardny cas", "HECU": "Kubański lěśojski cas", "WESZ": "Pódwjacornoeuropski lěśojski cas", "IST": "Indiski cas", "WITA": "Srjejźoindoneski cas", "HENOMX": "Mexiski dłujkowjacorny lěśojski cas", "ART": "Argentinski standardny cas", "BOT": "Boliwiski cas", "HEEG": "Pódzajtšnogrönlandski lěśojski cas", "MYT": "Malajziski cas", "ACST": "Srjejźoawstralski standardny cas", "HEOG": "Pódwjacornogrönlandski lěśojski cas", "HKT": "Hongkongski standardny cas", "AEDT": "Pódzajtšnoawstralski lěśojski cas", "WEZ": "Pódwjacornoeuropski standardny cas", "WIB": "Pódwjacornoindoneski cas", "BT": "Bhutański cas", "HNOG": "Pódwjacornogrönlandski standardny cas", "ARST": "Argentinski lěśojski cas", "CDT": "Pódpołnocnoameriski centralny lěśojski cas", "AWST": "Pódwjacornoawstralski standardny cas", "AKST": "Alaskojski standardny cas", "AKDT": "Alaskojski lěśojski cas", "ECT": "Ekuadorski cas", "HAT": "Nowofundlandski lěśojski cas", "WAT": "Pódwjacornoafriski standardny cas", "ACWST": "Srjejźopódwjacorny awstralski standardny cas", "MEZ": "Srjejźoeuropski standardny cas", "SRT": "Surinamski cas", "TMST": "Turkmeniski lěśojski cas", "PST": "Pódpołnocnoameriski pacifiski standardny cas", "HEPMX": "Mexiski pacifiski lěśojski cas"},
	}
}

// Locale returns the current translators string locale
func (dsb *dsb) Locale() string {
	return dsb.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'dsb'
func (dsb *dsb) PluralsCardinal() []locales.PluralRule {
	return dsb.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'dsb'
func (dsb *dsb) PluralsOrdinal() []locales.PluralRule {
	return dsb.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'dsb'
func (dsb *dsb) PluralsRange() []locales.PluralRule {
	return dsb.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'dsb'
func (dsb *dsb) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)
	f := locales.F(n, v)
	iMod100 := i % 100
	fMod100 := f % 100

	if (v == 0 && iMod100 == 1) || (fMod100 == 1) {
		return locales.PluralRuleOne
	} else if (v == 0 && iMod100 == 2) || (fMod100 == 2) {
		return locales.PluralRuleTwo
	} else if (v == 0 && iMod100 >= 3 && iMod100 <= 4) || (fMod100 >= 3 && fMod100 <= 4) {
		return locales.PluralRuleFew
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'dsb'
func (dsb *dsb) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'dsb'
func (dsb *dsb) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (dsb *dsb) MonthAbbreviated(month time.Month) string {
	return dsb.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (dsb *dsb) MonthsAbbreviated() []string {
	return dsb.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (dsb *dsb) MonthNarrow(month time.Month) string {
	return dsb.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (dsb *dsb) MonthsNarrow() []string {
	return dsb.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (dsb *dsb) MonthWide(month time.Month) string {
	return dsb.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (dsb *dsb) MonthsWide() []string {
	return dsb.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (dsb *dsb) WeekdayAbbreviated(weekday time.Weekday) string {
	return dsb.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (dsb *dsb) WeekdaysAbbreviated() []string {
	return dsb.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (dsb *dsb) WeekdayNarrow(weekday time.Weekday) string {
	return dsb.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (dsb *dsb) WeekdaysNarrow() []string {
	return dsb.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (dsb *dsb) WeekdayShort(weekday time.Weekday) string {
	return dsb.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (dsb *dsb) WeekdaysShort() []string {
	return dsb.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (dsb *dsb) WeekdayWide(weekday time.Weekday) string {
	return dsb.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (dsb *dsb) WeekdaysWide() []string {
	return dsb.daysWide
}

// Decimal returns the decimal point of number
func (dsb *dsb) Decimal() string {
	return dsb.decimal
}

// Group returns the group of number
func (dsb *dsb) Group() string {
	return dsb.group
}

// Group returns the minus sign of number
func (dsb *dsb) Minus() string {
	return dsb.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'dsb' and handles both Whole and Real numbers based on 'v'
func (dsb *dsb) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, dsb.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, dsb.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, dsb.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'dsb' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (dsb *dsb) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 5
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, dsb.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, dsb.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, dsb.percentSuffix...)

	b = append(b, dsb.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'dsb'
func (dsb *dsb) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := dsb.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, dsb.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, dsb.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, dsb.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, dsb.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, dsb.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'dsb'
// in accounting notation.
func (dsb *dsb) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := dsb.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, dsb.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, dsb.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, dsb.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, dsb.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, dsb.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, dsb.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'dsb'
func (dsb *dsb) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e}...)
	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Year() > 9 {
		b = append(b, strconv.Itoa(t.Year())[2:]...)
	} else {
		b = append(b, strconv.Itoa(t.Year())[1:]...)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'dsb'
func (dsb *dsb) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'dsb'
func (dsb *dsb) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, dsb.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'dsb'
func (dsb *dsb) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, dsb.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, dsb.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'dsb'
func (dsb *dsb) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, dsb.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'dsb'
func (dsb *dsb) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, dsb.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, dsb.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'dsb'
func (dsb *dsb) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, dsb.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, dsb.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'dsb'
func (dsb *dsb) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, dsb.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, dsb.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := dsb.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
