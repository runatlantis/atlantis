package hsb_DE

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type hsb_DE struct {
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

// New returns a new instance of translator for the 'hsb_DE' locale
func New() locales.Translator {
	return &hsb_DE{
		locale:                 "hsb_DE",
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
		timezones:              map[string]string{"TMT": "turkmenski standardny čas", "CDT": "sewjeroameriski centralny lětni čas", "WAST": "zapadoafriski lětni čas", "MYT": "malajziski čas", "COST": "kolumbiski lětni čas", "CLST": "chilski lětni čas", "UYT": "uruguayski standardny čas", "AEDT": "wuchodoawstralski lětni čas", "HNOG": "zapadogrönlandski standardny čas", "PDT": "sewjeroameriski pacifiski lětni čas", "AWDT": "zapadoawstralski lětni čas", "WEZ": "zapadoeuropski standardny čas", "MST": "MST", "MDT": "MDT", "HAST": "hawaiisko-aleutski standardny čas", "ChST": "chamorroski čas", "HEPMX": "mexiski pacifiski lětni čas", "BOT": "boliwiski čas", "BT": "bhutanski čas", "IST": "indiski čas", "LHDT": "lětni čas kupy Lord-Howe", "HNNOMX": "mexiski sewjerozapadny standardny čas", "OESZ": "wuchodoeuropski lětni čas", "NZST": "nowoseelandski standardny čas", "SGT": "Singapurski čas", "HKST": "Hongkongski lětni čas", "WAT": "zapadoafriski standardny čas", "VET": "venezuelski čas", "CAT": "centralnoafriski čas", "ARST": "argentinski lětni čas", "UYST": "uruguayski lětni čas", "WIB": "zapadoindoneski čas", "HEEG": "wuchodogrönlandski lětni čas", "EDT": "sewjeroameriski wuchodny lětni čas", "WIT": "wuchodoindoneski", "OEZ": "wuchodoeuropski standardny čas", "GYT": "guyanski čas", "CST": "sewjeroameriski centralny standardny čas", "SAST": "južnoafriski čas", "WESZ": "zapadoeuropski lětni čas", "NZDT": "nowoseelandski lětni čas", "AKDT": "alaskaski lětni čas", "HNT": "nowofundlandski standardny čas", "HAT": "nowofundlandski lětni čas", "HENOMX": "mexiski sewjerozapadny lětni čas", "CHAST": "chathamski standardny čas", "AST": "atlantiski standardny čas", "ECT": "ekwadorski čas", "LHST": "standardny čas kupy Lord-Howe", "WITA": "srjedźoindoneski čas", "SRT": "surinamski čas", "HNCU": "kubaski standardny čas", "HECU": "kubaski lětni čas", "ART": "argentinski standardny čas", "CHADT": "chathamski lětni čas", "HNPMX": "mexiski pacifiski standardny čas", "WART": "zapadoargentinski standardny čas", "HEPM": "lětni čas kupow St. Pierre a Miquelon", "EAT": "wuchodoafriski čas", "HADT": "hawaiisko-aleutski lětni čas", "PST": "sewjeroameriski pacifiski standardny čas", "ADT": "atlantiski lětni čas", "WARST": "zapadoargentinski lětni čas", "TMST": "turkmenski lětni čas", "AEST": "wuchodoawstralski standardny čas", "GFT": "francoskoguyanski čas", "EST": "sewjeroameriski wuchodny standardny čas", "ACWDT": "sjedźozapadny awstralski lětni čas", "CLT": "chilski standardny čas", "AWST": "zapadoawstralski standardny čas", "HEOG": "zapadogrönlandski lětni čas", "∅∅∅": "Amaconaski lětni čas", "ACDT": "srjedźoawstralski lětni čas", "HKT": "Hongkongski standardny čas", "HNPM": "standardny čas kupow St. Pierre a Miquelon", "JST": "japanski standardny čas", "HNEG": "wuchodogrönlandski standardny čas", "ACST": "srjedźoawstralski standardny čas", "MESZ": "srjedźoeuropski lětni čas", "MEZ": "srjedźoeuropski standardny čas", "COT": "kolumbiski standardny čas", "GMT": "Greenwichski čas", "JDT": "japanski lětni čas", "AKST": "alaskaski standardny čas", "ACWST": "srjedźozapadny awstralski standardny čas"},
	}
}

// Locale returns the current translators string locale
func (hsb *hsb_DE) Locale() string {
	return hsb.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'hsb_DE'
func (hsb *hsb_DE) PluralsCardinal() []locales.PluralRule {
	return hsb.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'hsb_DE'
func (hsb *hsb_DE) PluralsOrdinal() []locales.PluralRule {
	return hsb.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'hsb_DE'
func (hsb *hsb_DE) PluralsRange() []locales.PluralRule {
	return hsb.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'hsb_DE'
func (hsb *hsb_DE) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

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

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'hsb_DE'
func (hsb *hsb_DE) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'hsb_DE'
func (hsb *hsb_DE) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (hsb *hsb_DE) MonthAbbreviated(month time.Month) string {
	return hsb.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (hsb *hsb_DE) MonthsAbbreviated() []string {
	return hsb.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (hsb *hsb_DE) MonthNarrow(month time.Month) string {
	return hsb.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (hsb *hsb_DE) MonthsNarrow() []string {
	return hsb.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (hsb *hsb_DE) MonthWide(month time.Month) string {
	return hsb.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (hsb *hsb_DE) MonthsWide() []string {
	return hsb.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (hsb *hsb_DE) WeekdayAbbreviated(weekday time.Weekday) string {
	return hsb.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (hsb *hsb_DE) WeekdaysAbbreviated() []string {
	return hsb.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (hsb *hsb_DE) WeekdayNarrow(weekday time.Weekday) string {
	return hsb.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (hsb *hsb_DE) WeekdaysNarrow() []string {
	return hsb.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (hsb *hsb_DE) WeekdayShort(weekday time.Weekday) string {
	return hsb.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (hsb *hsb_DE) WeekdaysShort() []string {
	return hsb.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (hsb *hsb_DE) WeekdayWide(weekday time.Weekday) string {
	return hsb.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (hsb *hsb_DE) WeekdaysWide() []string {
	return hsb.daysWide
}

// Decimal returns the decimal point of number
func (hsb *hsb_DE) Decimal() string {
	return hsb.decimal
}

// Group returns the group of number
func (hsb *hsb_DE) Group() string {
	return hsb.group
}

// Group returns the minus sign of number
func (hsb *hsb_DE) Minus() string {
	return hsb.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'hsb_DE' and handles both Whole and Real numbers based on 'v'
func (hsb *hsb_DE) FmtNumber(num float64, v uint64) string {

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

// FmtPercent returns 'num' with digits/precision of 'v' for 'hsb_DE' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (hsb *hsb_DE) FmtPercent(num float64, v uint64) string {
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

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'hsb_DE'
func (hsb *hsb_DE) FmtCurrency(num float64, v uint64, currency currency.Type) string {

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

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'hsb_DE'
// in accounting notation.
func (hsb *hsb_DE) FmtAccounting(num float64, v uint64, currency currency.Type) string {

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

// FmtDateShort returns the short date representation of 't' for 'hsb_DE'
func (hsb *hsb_DE) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'hsb_DE'
func (hsb *hsb_DE) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'hsb_DE'
func (hsb *hsb_DE) FmtDateLong(t time.Time) string {

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

// FmtDateFull returns the full date representation of 't' for 'hsb_DE'
func (hsb *hsb_DE) FmtDateFull(t time.Time) string {

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

// FmtTimeShort returns the short time representation of 't' for 'hsb_DE'
func (hsb *hsb_DE) FmtTimeShort(t time.Time) string {

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

// FmtTimeMedium returns the medium time representation of 't' for 'hsb_DE'
func (hsb *hsb_DE) FmtTimeMedium(t time.Time) string {

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

// FmtTimeLong returns the long time representation of 't' for 'hsb_DE'
func (hsb *hsb_DE) FmtTimeLong(t time.Time) string {

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

// FmtTimeFull returns the full time representation of 't' for 'hsb_DE'
func (hsb *hsb_DE) FmtTimeFull(t time.Time) string {

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
