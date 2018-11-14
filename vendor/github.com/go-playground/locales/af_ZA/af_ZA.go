package af_ZA

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type af_ZA struct {
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

// New returns a new instance of translator for the 'af_ZA' locale
func New() locales.Translator {
	return &af_ZA{
		locale:                 "af_ZA",
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
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
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
		timezones:              map[string]string{"TMST": "Turkmenistan-somertyd", "ARST": "Argentinië-somertyd", "BOT": "Bolivia-tyd", "AEDT": "Oostelike Australiese dagligtyd", "WAT": "Wes-Afrika-standaardtyd", "JDT": "Japan-dagligtyd", "CLT": "Chili-standaardtyd", "OESZ": "Oos-Europese somertyd", "GYT": "Guyana-tyd", "PDT": "Pasifiese dagligtyd", "HNPMX": "Meksikaanse Pasifiese standaardtyd", "NZST": "Nieu-Seeland-standaardtyd", "EST": "Noord-Amerikaanse oostelike standaardtyd", "WART": "Wes-Argentinië-standaardtyd", "ChST": "Chamorro-standaardtyd", "BT": "Bhoetan-tyd", "ACWST": "Sentraal-westelike Australiese standaard-tyd", "HEEG": "Oos-Groenland-somertyd", "MST": "MST", "CAT": "Sentraal-Afrika-tyd", "TMT": "Turkmenistan-standaardtyd", "COST": "Colombië-somertyd", "CST": "Noord-Amerikaanse sentrale standaardtyd", "HADT": "Hawaii-Aleoete-dagligtyd", "∅∅∅": "Amasone-somertyd", "AKDT": "Alaska-dagligtyd", "MESZ": "Sentraal-Europese somertyd", "LHST": "Lord Howe-standaardtyd", "MDT": "MDT", "OEZ": "Oos-Europese standaardtyd", "UYST": "Uruguay-somertyd", "SAST": "Suid-Afrika-standaardtyd", "WAST": "Wes-Afrika-somertyd", "HAT": "Newfoundland-dagligtyd", "PST": "Pasifiese standaardtyd", "WEZ": "Wes-Europese standaardtyd", "AKST": "Alaska-standaardtyd", "SGT": "Singapoer-standaardtyd", "HNOG": "Wes-Groenland-standaardtyd", "HEPMX": "Meksikaanse Pasifiese dagligtyd", "MYT": "Maleisië-tyd", "HAST": "Hawaii-Aleoete-standaardtyd", "AWDT": "Westelike Australiese dagligtyd", "VET": "Venezuela-tyd", "HNPM": "Sint-Pierre en Miquelon-standaardtyd", "HEPM": "Sint-Pierre en Miquelon-dagligtyd", "CHADT": "Chatham-dagligtyd", "HECU": "Kuba-dagligtyd", "AEST": "Oostelike Australiese standaardtyd", "HNEG": "Oos-Groenland-standaardtyd", "WARST": "Wes-Argentinië-somertyd", "MEZ": "Sentraal-Europese standaardtyd", "SRT": "Suriname-tyd", "WIT": "Oos-Indonesië-tyd", "UYT": "Uruguay-standaardtyd", "HNCU": "Kuba-standaardtyd", "AST": "Atlantiese standaardtyd", "WESZ": "Wes-Europese somertyd", "JST": "Japan-standaardtyd", "GFT": "Frans-Guiana-tyd", "ACWDT": "Sentraal-westelike Australiese dagligtyd", "WITA": "Sentraal-Indonesiese tyd", "ACDT": "Sentraal-Australiese dagligtyd", "HEOG": "Wes-Groenland-somertyd", "HKT": "Hongkong-standaardtyd", "HNNOMX": "Noordwes-Meksiko-standaardtyd", "EAT": "Oos-Afrika-tyd", "ART": "Argentinië-standaardtyd", "NZDT": "Nieu-Seeland-dagligtyd", "EDT": "Noord-Amerikaanse oostelike dagligtyd", "HKST": "Hongkong-somertyd", "IST": "Indië-standaardtyd", "LHDT": "Lord Howe-dagligtyd", "HENOMX": "Noordwes-Meksiko-dagligtyd", "GMT": "Greenwich-tyd", "CHAST": "Chatham-standaardtyd", "CDT": "Noord-Amerikaanse sentrale dagligtyd", "ECT": "Ecuador-tyd", "ACST": "Sentraal-Australiese standaardtyd", "HNT": "Newfoundland-standaardtyd", "CLST": "Chili-somertyd", "COT": "Colombië-standaardtyd", "AWST": "Westelike Australiese standaardtyd", "ADT": "Atlantiese dagligtyd", "WIB": "Wes-Indonesië-tyd"},
	}
}

// Locale returns the current translators string locale
func (af *af_ZA) Locale() string {
	return af.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'af_ZA'
func (af *af_ZA) PluralsCardinal() []locales.PluralRule {
	return af.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'af_ZA'
func (af *af_ZA) PluralsOrdinal() []locales.PluralRule {
	return af.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'af_ZA'
func (af *af_ZA) PluralsRange() []locales.PluralRule {
	return af.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'af_ZA'
func (af *af_ZA) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'af_ZA'
func (af *af_ZA) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'af_ZA'
func (af *af_ZA) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (af *af_ZA) MonthAbbreviated(month time.Month) string {
	return af.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (af *af_ZA) MonthsAbbreviated() []string {
	return af.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (af *af_ZA) MonthNarrow(month time.Month) string {
	return af.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (af *af_ZA) MonthsNarrow() []string {
	return af.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (af *af_ZA) MonthWide(month time.Month) string {
	return af.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (af *af_ZA) MonthsWide() []string {
	return af.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (af *af_ZA) WeekdayAbbreviated(weekday time.Weekday) string {
	return af.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (af *af_ZA) WeekdaysAbbreviated() []string {
	return af.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (af *af_ZA) WeekdayNarrow(weekday time.Weekday) string {
	return af.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (af *af_ZA) WeekdaysNarrow() []string {
	return af.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (af *af_ZA) WeekdayShort(weekday time.Weekday) string {
	return af.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (af *af_ZA) WeekdaysShort() []string {
	return af.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (af *af_ZA) WeekdayWide(weekday time.Weekday) string {
	return af.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (af *af_ZA) WeekdaysWide() []string {
	return af.daysWide
}

// Decimal returns the decimal point of number
func (af *af_ZA) Decimal() string {
	return af.decimal
}

// Group returns the group of number
func (af *af_ZA) Group() string {
	return af.group
}

// Group returns the minus sign of number
func (af *af_ZA) Minus() string {
	return af.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'af_ZA' and handles both Whole and Real numbers based on 'v'
func (af *af_ZA) FmtNumber(num float64, v uint64) string {

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

// FmtPercent returns 'num' with digits/precision of 'v' for 'af_ZA' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (af *af_ZA) FmtPercent(num float64, v uint64) string {
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

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'af_ZA'
func (af *af_ZA) FmtCurrency(num float64, v uint64, currency currency.Type) string {

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

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'af_ZA'
// in accounting notation.
func (af *af_ZA) FmtAccounting(num float64, v uint64, currency currency.Type) string {

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

// FmtDateShort returns the short date representation of 't' for 'af_ZA'
func (af *af_ZA) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'af_ZA'
func (af *af_ZA) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'af_ZA'
func (af *af_ZA) FmtDateLong(t time.Time) string {

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

// FmtDateFull returns the full date representation of 't' for 'af_ZA'
func (af *af_ZA) FmtDateFull(t time.Time) string {

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

// FmtTimeShort returns the short time representation of 't' for 'af_ZA'
func (af *af_ZA) FmtTimeShort(t time.Time) string {

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

// FmtTimeMedium returns the medium time representation of 't' for 'af_ZA'
func (af *af_ZA) FmtTimeMedium(t time.Time) string {

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

// FmtTimeLong returns the long time representation of 't' for 'af_ZA'
func (af *af_ZA) FmtTimeLong(t time.Time) string {

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

// FmtTimeFull returns the full time representation of 't' for 'af_ZA'
func (af *af_ZA) FmtTimeFull(t time.Time) string {

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
