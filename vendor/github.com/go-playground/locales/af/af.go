package af

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type af struct {
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

// New returns a new instance of translator for the 'af' locale
func New() locales.Translator {
	return &af{
		locale:                 "af",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         []locales.PluralRule{6},
		pluralsRange:           []locales.PluralRule{6},
		decimal:                ",",
		group:                  " ",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "A$", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "₪", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JP¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "₩", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZ$", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "฿", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "NT$", "TZS", "UAH", "UAK", "UGS", "UGX", "US$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "₫", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "EC$", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "R", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: ")",
		monthsAbbreviated:      []string{"", "Jan.", "Feb.", "Mrt.", "Apr.", "Mei", "Jun.", "Jul.", "Aug.", "Sep.", "Okt.", "Nov.", "Des."},
		monthsNarrow:           []string{"", "J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"},
		monthsWide:             []string{"", "Januarie", "Februarie", "Maart", "April", "Mei", "Junie", "Julie", "Augustus", "September", "Oktober", "November", "Desember"},
		daysAbbreviated:        []string{"So.", "Ma.", "Di.", "Wo.", "Do.", "Vr.", "Sa."},
		daysNarrow:             []string{"S", "M", "D", "W", "D", "V", "S"},
		daysShort:              []string{"So.", "Ma.", "Di.", "Wo.", "Do.", "Vr.", "Sa."},
		daysWide:               []string{"Sondag", "Maandag", "Dinsdag", "Woensdag", "Donderdag", "Vrydag", "Saterdag"},
		periodsAbbreviated:     []string{"vm.", "nm."},
		periodsNarrow:          []string{"v", "n"},
		periodsWide:            []string{"vm.", "nm."},
		erasAbbreviated:        []string{"v.C.", "n.C."},
		erasNarrow:             []string{"v.C.", "n.C."},
		erasWide:               []string{"voor Christus", "na Christus"},
		timezones:              map[string]string{"GMT": "Greenwich-tyd", "CHADT": "Chatham-dagligtyd", "CDT": "Noord-Amerikaanse sentrale dagligtyd", "HEPMX": "Meksikaanse Pasifiese dagligtyd", "GFT": "Frans-Guiana-tyd", "NZDT": "Nieu-Seeland-dagligtyd", "HKST": "Hongkong-somertyd", "LHST": "Lord Howe-standaardtyd", "WITA": "Sentraal-Indonesiese tyd", "CLT": "Chili-standaardtyd", "AST": "Atlantiese standaardtyd", "WAST": "Wes-Afrika-somertyd", "AKDT": "Alaska-dagligtyd", "ACST": "Sentraal-Australiese standaardtyd", "MEZ": "Sentraal-Europese standaardtyd", "HNPM": "Sint-Pierre en Miquelon-standaardtyd", "CLST": "Chili-somertyd", "MST": "Noord-Amerikaanse berg-standaardtyd", "MDT": "Noord-Amerikaanse berg-dagligtyd", "EDT": "Noord-Amerikaanse oostelike dagligtyd", "HENOMX": "Noordwes-Meksiko-dagligtyd", "CST": "Noord-Amerikaanse sentrale standaardtyd", "AWST": "Westelike Australiese standaardtyd", "SAST": "Suid-Afrika-standaardtyd", "JDT": "Japan-dagligtyd", "ACDT": "Sentraal-Australiese dagligtyd", "WART": "Wes-Argentinië-standaardtyd", "HNT": "Newfoundland-standaardtyd", "HNNOMX": "Noordwes-Meksiko-standaardtyd", "TMST": "Turkmenistan-somertyd", "HNCU": "Kuba-standaardtyd", "BT": "Bhoetan-tyd", "UYT": "Uruguay-standaardtyd", "WIB": "Wes-Indonesië-tyd", "HKT": "Hongkong-standaardtyd", "SRT": "Suriname-tyd", "AWDT": "Westelike Australiese dagligtyd", "HNPMX": "Meksikaanse Pasifiese standaardtyd", "MYT": "Maleisië-tyd", "AKST": "Alaska-standaardtyd", "HNEG": "Oos-Groenland-standaardtyd", "LHDT": "Lord Howe-dagligtyd", "PDT": "Pasifiese dagligtyd", "AEST": "Oostelike Australiese standaardtyd", "HEPM": "Sint-Pierre en Miquelon-dagligtyd", "WIT": "Oos-Indonesië-tyd", "UYST": "Uruguay-somertyd", "ADT": "Atlantiese dagligtyd", "BOT": "Bolivia-tyd", "HNOG": "Wes-Groenland-standaardtyd", "HEOG": "Wes-Groenland-somertyd", "EST": "Noord-Amerikaanse oostelike standaardtyd", "ACWST": "Sentraal-westelike Australiese standaard-tyd", "EAT": "Oos-Afrika-tyd", "OESZ": "Oos-Europese somertyd", "COST": "Colombië-somertyd", "WAT": "Wes-Afrika-standaardtyd", "ECT": "Ecuador-tyd", "WARST": "Wes-Argentinië-somertyd", "VET": "Venezuela-tyd", "ARST": "Argentinië-somertyd", "PST": "Pasifiese standaardtyd", "WESZ": "Wes-Europese somertyd", "JST": "Japan-standaardtyd", "IST": "Indië-standaardtyd", "HAST": "Hawaii-Aleoete-standaardtyd", "GYT": "Guyana-tyd", "ChST": "Chamorro-standaardtyd", "HECU": "Kuba-dagligtyd", "AEDT": "Oostelike Australiese dagligtyd", "MESZ": "Sentraal-Europese somertyd", "OEZ": "Oos-Europese standaardtyd", "ART": "Argentinië-standaardtyd", "ACWDT": "Sentraal-westelike Australiese dagligtyd", "HEEG": "Oos-Groenland-somertyd", "HAT": "Newfoundland-dagligtyd", "TMT": "Turkmenistan-standaardtyd", "CAT": "Sentraal-Afrika-tyd", "CHAST": "Chatham-standaardtyd", "∅∅∅": "Brasilia-somertyd", "WEZ": "Wes-Europese standaardtyd", "NZST": "Nieu-Seeland-standaardtyd", "SGT": "Singapoer-standaardtyd", "HADT": "Hawaii-Aleoete-dagligtyd", "COT": "Colombië-standaardtyd"},
	}
}

// Locale returns the current translators string locale
func (af *af) Locale() string {
	return af.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'af'
func (af *af) PluralsCardinal() []locales.PluralRule {
	return af.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'af'
func (af *af) PluralsOrdinal() []locales.PluralRule {
	return af.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'af'
func (af *af) PluralsRange() []locales.PluralRule {
	return af.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'af'
func (af *af) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'af'
func (af *af) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'af'
func (af *af) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (af *af) MonthAbbreviated(month time.Month) string {
	return af.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (af *af) MonthsAbbreviated() []string {
	return af.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (af *af) MonthNarrow(month time.Month) string {
	return af.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (af *af) MonthsNarrow() []string {
	return af.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (af *af) MonthWide(month time.Month) string {
	return af.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (af *af) MonthsWide() []string {
	return af.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (af *af) WeekdayAbbreviated(weekday time.Weekday) string {
	return af.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (af *af) WeekdaysAbbreviated() []string {
	return af.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (af *af) WeekdayNarrow(weekday time.Weekday) string {
	return af.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (af *af) WeekdaysNarrow() []string {
	return af.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (af *af) WeekdayShort(weekday time.Weekday) string {
	return af.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (af *af) WeekdaysShort() []string {
	return af.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (af *af) WeekdayWide(weekday time.Weekday) string {
	return af.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (af *af) WeekdaysWide() []string {
	return af.daysWide
}

// Decimal returns the decimal point of number
func (af *af) Decimal() string {
	return af.decimal
}

// Group returns the group of number
func (af *af) Group() string {
	return af.group
}

// Group returns the minus sign of number
func (af *af) Minus() string {
	return af.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'af' and handles both Whole and Real numbers based on 'v'
func (af *af) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, af.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(af.group) - 1; j >= 0; j-- {
					b = append(b, af.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, af.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'af' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (af *af) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, af.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, af.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, af.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'af'
func (af *af) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := af.currencies[currency]
	l := len(s) + len(symbol) + 2 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, af.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(af.group) - 1; j >= 0; j-- {
					b = append(b, af.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	for j := len(symbol) - 1; j >= 0; j-- {
		b = append(b, symbol[j])
	}

	if num < 0 {
		b = append(b, af.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, af.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'af'
// in accounting notation.
func (af *af) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := af.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, af.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(af.group) - 1; j >= 0; j-- {
					b = append(b, af.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		b = append(b, af.currencyNegativePrefix[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, af.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, af.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'af'
func (af *af) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'af'
func (af *af) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, af.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'af'
func (af *af) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, af.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'af'
func (af *af) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, af.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, af.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'af'
func (af *af) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, af.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'af'
func (af *af) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, af.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, af.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'af'
func (af *af) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, af.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, af.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'af'
func (af *af) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, af.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, af.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := af.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
