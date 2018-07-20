package dsb_DE

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type dsb_DE struct {
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

// New returns a new instance of translator for the 'dsb_DE' locale
func New() locales.Translator {
	return &dsb_DE{
		locale:                 "dsb_DE",
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
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
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
		timezones:              map[string]string{"SRT": "Surinamski cas", "UYT": "Uruguayski standardny cas", "∅∅∅": "Brasília lěśojski cas", "AEST": "Pódzajtšnoawstralski standardny cas", "SAST": "Pódpołdnjowoafriski cas", "WAST": "Pódwjacornoafriski lěśojski cas", "ECT": "Ekuadorski cas", "HEPM": "St.-Pierre-a-Miqueloński lěśojski cas", "TMT": "Turkmeniski standardny cas", "COT": "Kolumbiski standardny cas", "COST": "Kolumbiski lěśojski cas", "OESZ": "Pódzajtšnoeuropski lěśojski cas", "CHADT": "Chathamski lěśojski cas", "JST": "Japański standardny cas", "MEZ": "Srjejźoeuropski standardny cas", "HENOMX": "Mexiski dłujkowjacorny lěśojski cas", "HEPMX": "Mexiski pacifiski lěśojski cas", "AST": "Atlantiski standardny cas", "ACWST": "Srjejźopódwjacorny awstralski standardny cas", "EDT": "Pódpołnocnoameriski pódzajtšny lěśojski cas", "HAST": "Hawaiisko-aleutski standardny cas", "GMT": "Greenwichski cas", "BOT": "Boliwiski cas", "LHDT": "lěśojski cas kupy Lord-Howe", "ADT": "Atlantiski lěśojski cas", "GFT": "Francojskoguyański cas", "NZST": "Nowoseelandski standardny cas", "HEOG": "Pódwjacornogrönlandski lěśojski cas", "HNNOMX": "Mexiski dłujkowjacorny standardny cas", "CST": "Pódpołnocnoameriski centralny standardny cas", "PST": "Pódpołnocnoameriski pacifiski standardny cas", "HKST": "Hongkongski lěśojski cas", "HNT": "Nowofundlandski standardny cas", "CLT": "Chilski standardny cas", "ChST": "Chamorrski cas", "HECU": "Kubański lěśojski cas", "HNPMX": "Mexiski pacifiski standardny cas", "NZDT": "Nowoseelandski lěśojski cas", "HNEG": "Pódzajtšnogrönlandski standardny cas", "ART": "Argentinski standardny cas", "ARST": "Argentinski lěśojski cas", "AEDT": "Pódzajtšnoawstralski lěśojski cas", "ACWDT": "Srjejźopódwjacorny awstralski lěśojski cas", "WART": "Pódwjacornoargentinski standardny cas", "MESZ": "Srjejźoeuropski lěśojski cas", "CDT": "Pódpołnocnoameriski centralny lěśojski cas", "MST": "Pódpołnocnoameriski górski standardny cas", "WESZ": "Pódwjacornoeuropski lěśojski cas", "MYT": "Malajziski cas", "AKST": "Alaskojski standardny cas", "HNOG": "Pódwjacornogrönlandski standardny cas", "EST": "Pódpołnocnoameriski pódzajtšny standardny cas", "AWDT": "Pódwjacornoawstralski lěśojski cas", "WAT": "Pódwjacornoafriski standardny cas", "ACDT": "Srjejźoawstralski lěśojski cas", "WITA": "Srjejźoindoneski cas", "VET": "Venezuelski cas", "WIT": "Pódzajtšnoindoneski", "WEZ": "Pódwjacornoeuropski standardny cas", "AKDT": "Alaskojski lěśojski cas", "HAT": "Nowofundlandski lěśojski cas", "CAT": "Srjejźoafriski cas", "UYST": "Uruguayski lěśojski cas", "ACST": "Srjejźoawstralski standardny cas", "WARST": "Pódwjacornoargentinski lěśojski cas", "IST": "Indiski cas", "HADT": "Hawaiisko-aleutski lěśojski cas", "GYT": "Guyański cas", "HNCU": "Kubański standardny cas", "JDT": "Japański lěśojski cas", "SGT": "Singapurski cas", "CLST": "Chilski lěśojski cas", "PDT": "Pódpołnocnoameriski pacifiski lěśojski cas", "WIB": "Pódwjacornoindoneski cas", "HEEG": "Pódzajtšnogrönlandski lěśojski cas", "HNPM": "St.-Pierre-a-Miqueloński standardny cas", "OEZ": "Pódzajtšnoeuropski standardny cas", "AWST": "Pódwjacornoawstralski standardny cas", "TMST": "Turkmeniski lěśojski cas", "CHAST": "Chathamski standardny cas", "MDT": "Pódpołnocnoameriski górski lěśojski cas", "BT": "Bhutański cas", "HKT": "Hongkongski standardny cas", "LHST": "Standardny cas kupy Lord-Howe", "EAT": "Pódzajtšnoafriski cas"},
	}
}

// Locale returns the current translators string locale
func (dsb *dsb_DE) Locale() string {
	return dsb.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'dsb_DE'
func (dsb *dsb_DE) PluralsCardinal() []locales.PluralRule {
	return dsb.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'dsb_DE'
func (dsb *dsb_DE) PluralsOrdinal() []locales.PluralRule {
	return dsb.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'dsb_DE'
func (dsb *dsb_DE) PluralsRange() []locales.PluralRule {
	return dsb.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'dsb_DE'
func (dsb *dsb_DE) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

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

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'dsb_DE'
func (dsb *dsb_DE) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'dsb_DE'
func (dsb *dsb_DE) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (dsb *dsb_DE) MonthAbbreviated(month time.Month) string {
	return dsb.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (dsb *dsb_DE) MonthsAbbreviated() []string {
	return dsb.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (dsb *dsb_DE) MonthNarrow(month time.Month) string {
	return dsb.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (dsb *dsb_DE) MonthsNarrow() []string {
	return dsb.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (dsb *dsb_DE) MonthWide(month time.Month) string {
	return dsb.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (dsb *dsb_DE) MonthsWide() []string {
	return dsb.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (dsb *dsb_DE) WeekdayAbbreviated(weekday time.Weekday) string {
	return dsb.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (dsb *dsb_DE) WeekdaysAbbreviated() []string {
	return dsb.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (dsb *dsb_DE) WeekdayNarrow(weekday time.Weekday) string {
	return dsb.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (dsb *dsb_DE) WeekdaysNarrow() []string {
	return dsb.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (dsb *dsb_DE) WeekdayShort(weekday time.Weekday) string {
	return dsb.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (dsb *dsb_DE) WeekdaysShort() []string {
	return dsb.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (dsb *dsb_DE) WeekdayWide(weekday time.Weekday) string {
	return dsb.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (dsb *dsb_DE) WeekdaysWide() []string {
	return dsb.daysWide
}

// Decimal returns the decimal point of number
func (dsb *dsb_DE) Decimal() string {
	return dsb.decimal
}

// Group returns the group of number
func (dsb *dsb_DE) Group() string {
	return dsb.group
}

// Group returns the minus sign of number
func (dsb *dsb_DE) Minus() string {
	return dsb.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'dsb_DE' and handles both Whole and Real numbers based on 'v'
func (dsb *dsb_DE) FmtNumber(num float64, v uint64) string {

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

// FmtPercent returns 'num' with digits/precision of 'v' for 'dsb_DE' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (dsb *dsb_DE) FmtPercent(num float64, v uint64) string {
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

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'dsb_DE'
func (dsb *dsb_DE) FmtCurrency(num float64, v uint64, currency currency.Type) string {

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

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'dsb_DE'
// in accounting notation.
func (dsb *dsb_DE) FmtAccounting(num float64, v uint64, currency currency.Type) string {

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

// FmtDateShort returns the short date representation of 't' for 'dsb_DE'
func (dsb *dsb_DE) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'dsb_DE'
func (dsb *dsb_DE) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'dsb_DE'
func (dsb *dsb_DE) FmtDateLong(t time.Time) string {

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

// FmtDateFull returns the full date representation of 't' for 'dsb_DE'
func (dsb *dsb_DE) FmtDateFull(t time.Time) string {

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

// FmtTimeShort returns the short time representation of 't' for 'dsb_DE'
func (dsb *dsb_DE) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, dsb.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'dsb_DE'
func (dsb *dsb_DE) FmtTimeMedium(t time.Time) string {

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

// FmtTimeLong returns the long time representation of 't' for 'dsb_DE'
func (dsb *dsb_DE) FmtTimeLong(t time.Time) string {

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

// FmtTimeFull returns the full time representation of 't' for 'dsb_DE'
func (dsb *dsb_DE) FmtTimeFull(t time.Time) string {

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
