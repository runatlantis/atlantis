package hsb

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type hsb struct {
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

// New returns a new instance of translator for the 'hsb' locale
func New() locales.Translator {
	return &hsb{
		locale:                 "hsb",
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
		monthsAbbreviated:      []string{"", "jan.", "feb.", "měr.", "apr.", "mej.", "jun.", "jul.", "awg.", "sep.", "okt.", "now.", "dec."},
		monthsNarrow:           []string{"", "j", "f", "m", "a", "m", "j", "j", "a", "s", "o", "n", "d"},
		monthsWide:             []string{"", "januara", "februara", "měrca", "apryla", "meje", "junija", "julija", "awgusta", "septembra", "oktobra", "nowembra", "decembra"},
		daysAbbreviated:        []string{"nje", "pón", "wut", "srj", "štw", "pja", "sob"},
		daysNarrow:             []string{"n", "p", "w", "s", "š", "p", "s"},
		daysShort:              []string{"nj", "pó", "wu", "sr", "št", "pj", "so"},
		daysWide:               []string{"njedźela", "póndźela", "wutora", "srjeda", "štwórtk", "pjatk", "sobota"},
		periodsAbbreviated:     []string{"dopołdnja", "popołdnju"},
		periodsNarrow:          []string{"dop.", "pop."},
		periodsWide:            []string{"dopołdnja", "popołdnju"},
		erasAbbreviated:        []string{"př.Chr.n.", "po Chr.n."},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"před Chrystowym narodźenjom", "po Chrystowym narodźenju"},
		timezones:              map[string]string{"HEPM": "lětni čas kupow St. Pierre a Miquelon", "SRT": "surinamski čas", "TMT": "turkmenski standardny čas", "UYT": "uruguayski standardny čas", "MST": "sewjeroameriski hórski standardny čas", "SAST": "južnoafriski čas", "MYT": "malajziski čas", "HAT": "nowofundlandski lětni čas", "HNCU": "kubaski standardny čas", "HAST": "hawaiisko-aleutski standardny čas", "ART": "argentinski standardny čas", "AKDT": "alaskaski lětni čas", "HENOMX": "mexiski sewjerozapadny lětni čas", "EAT": "wuchodoafriski čas", "OESZ": "wuchodoeuropski lětni čas", "CLT": "chilski standardny čas", "ADT": "atlantiski lětni čas", "JST": "japanski standardny čas", "ECT": "ekwadorski čas", "CAT": "centralnoafriski čas", "HECU": "kubaski lětni čas", "AWDT": "zapadoawstralski lětni čas", "WAT": "zapadoafriski standardny čas", "NZDT": "nowoseelandski lětni čas", "BOT": "boliwiski čas", "IST": "indiski čas", "GYT": "guyanski čas", "CDT": "sewjeroameriski centralny lětni čas", "AEST": "wuchodoawstralski standardny čas", "BT": "bhutanski čas", "AKST": "alaskaski standardny čas", "AEDT": "wuchodoawstralski lětni čas", "WIB": "zapadoindoneski čas", "SGT": "Singapurski čas", "ACWST": "srjedźozapadny awstralski standardny čas", "VET": "venezuelski čas", "WITA": "srjedźoindoneski čas", "CLST": "chilski lětni čas", "UYST": "uruguayski lětni čas", "CST": "sewjeroameriski centralny standardny čas", "AWST": "zapadoawstralski standardny čas", "NZST": "nowoseelandski standardny čas", "HKST": "Hongkongski lětni čas", "ACDT": "srjedźoawstralski lětni čas", "HNT": "nowofundlandski standardny čas", "HNPM": "standardny čas kupow St. Pierre a Miquelon", "LHST": "standardny čas kupy Lord-Howe", "TMST": "turkmenski lětni čas", "MEZ": "srjedźoeuropski standardny čas", "WIT": "wuchodoindoneski", "CHAST": "chathamski standardny čas", "PST": "sewjeroameriski pacifiski standardny čas", "JDT": "japanski lětni čas", "GFT": "francoskoguyanski čas", "HNOG": "zapadogrönlandski standardny čas", "WAST": "zapadoafriski lětni čas", "HNEG": "wuchodogrönlandski standardny čas", "OEZ": "wuchodoeuropski standardny čas", "PDT": "sewjeroameriski pacifiski lětni čas", "WESZ": "zapadoeuropski lětni čas", "EST": "sewjeroameriski wuchodny standardny čas", "EDT": "sewjeroameriski wuchodny lětni čas", "HNNOMX": "mexiski sewjerozapadny standardny čas", "COT": "kolumbiski standardny čas", "ChST": "chamorroski čas", "MDT": "sewjeroameriski hórski lětni čas", "ACWDT": "sjedźozapadny awstralski lětni čas", "HEEG": "wuchodogrönlandski lětni čas", "MESZ": "srjedźoeuropski lětni čas", "CHADT": "chathamski lětni čas", "AST": "atlantiski standardny čas", "∅∅∅": "∅∅∅", "ACST": "srjedźoawstralski standardny čas", "ARST": "argentinski lětni čas", "HADT": "hawaiisko-aleutski lětni čas", "COST": "kolumbiski lětni čas", "WEZ": "zapadoeuropski standardny čas", "HEOG": "zapadogrönlandski lětni čas", "WART": "zapadoargentinski standardny čas", "WARST": "zapadoargentinski lětni čas", "GMT": "Greenwichski čas", "HNPMX": "mexiski pacifiski standardny čas", "HEPMX": "mexiski pacifiski lětni čas", "HKT": "Hongkongski standardny čas", "LHDT": "lětni čas kupy Lord-Howe"},
	}
}

// Locale returns the current translators string locale
func (hsb *hsb) Locale() string {
	return hsb.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'hsb'
func (hsb *hsb) PluralsCardinal() []locales.PluralRule {
	return hsb.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'hsb'
func (hsb *hsb) PluralsOrdinal() []locales.PluralRule {
	return hsb.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'hsb'
func (hsb *hsb) PluralsRange() []locales.PluralRule {
	return hsb.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'hsb'
func (hsb *hsb) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

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

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'hsb'
func (hsb *hsb) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'hsb'
func (hsb *hsb) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (hsb *hsb) MonthAbbreviated(month time.Month) string {
	return hsb.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (hsb *hsb) MonthsAbbreviated() []string {
	return hsb.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (hsb *hsb) MonthNarrow(month time.Month) string {
	return hsb.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (hsb *hsb) MonthsNarrow() []string {
	return hsb.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (hsb *hsb) MonthWide(month time.Month) string {
	return hsb.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (hsb *hsb) MonthsWide() []string {
	return hsb.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (hsb *hsb) WeekdayAbbreviated(weekday time.Weekday) string {
	return hsb.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (hsb *hsb) WeekdaysAbbreviated() []string {
	return hsb.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (hsb *hsb) WeekdayNarrow(weekday time.Weekday) string {
	return hsb.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (hsb *hsb) WeekdaysNarrow() []string {
	return hsb.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (hsb *hsb) WeekdayShort(weekday time.Weekday) string {
	return hsb.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (hsb *hsb) WeekdaysShort() []string {
	return hsb.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (hsb *hsb) WeekdayWide(weekday time.Weekday) string {
	return hsb.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (hsb *hsb) WeekdaysWide() []string {
	return hsb.daysWide
}

// Decimal returns the decimal point of number
func (hsb *hsb) Decimal() string {
	return hsb.decimal
}

// Group returns the group of number
func (hsb *hsb) Group() string {
	return hsb.group
}

// Group returns the minus sign of number
func (hsb *hsb) Minus() string {
	return hsb.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'hsb' and handles both Whole and Real numbers based on 'v'
func (hsb *hsb) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, hsb.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, hsb.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, hsb.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'hsb' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (hsb *hsb) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 5
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, hsb.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, hsb.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, hsb.percentSuffix...)

	b = append(b, hsb.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'hsb'
func (hsb *hsb) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := hsb.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, hsb.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, hsb.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, hsb.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, hsb.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, hsb.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'hsb'
// in accounting notation.
func (hsb *hsb) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := hsb.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, hsb.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, hsb.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, hsb.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, hsb.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, hsb.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, hsb.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'hsb'
func (hsb *hsb) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'hsb'
func (hsb *hsb) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'hsb'
func (hsb *hsb) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, hsb.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'hsb'
func (hsb *hsb) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, hsb.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, hsb.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'hsb'
func (hsb *hsb) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, hsb.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20, 0x68, 0x6f, 0x64, 0xc5, 0xba}...)
	b = append(b, []byte{0x2e}...)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'hsb'
func (hsb *hsb) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, hsb.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, hsb.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'hsb'
func (hsb *hsb) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, hsb.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, hsb.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'hsb'
func (hsb *hsb) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, hsb.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, hsb.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := hsb.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
