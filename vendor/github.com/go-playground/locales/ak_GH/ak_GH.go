package ak_GH

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type ak_GH struct {
	locale             string
	pluralsCardinal    []locales.PluralRule
	pluralsOrdinal     []locales.PluralRule
	pluralsRange       []locales.PluralRule
	decimal            string
	group              string
	minus              string
	percent            string
	perMille           string
	timeSeparator      string
	inifinity          string
	currencies         []string // idx = enum of currency code
	monthsAbbreviated  []string
	monthsNarrow       []string
	monthsWide         []string
	daysAbbreviated    []string
	daysNarrow         []string
	daysShort          []string
	daysWide           []string
	periodsAbbreviated []string
	periodsNarrow      []string
	periodsShort       []string
	periodsWide        []string
	erasAbbreviated    []string
	erasNarrow         []string
	erasWide           []string
	timezones          map[string]string
}

// New returns a new instance of translator for the 'ak_GH' locale
func New() locales.Translator {
	return &ak_GH{
		locale:             "ak_GH",
		pluralsCardinal:    []locales.PluralRule{2, 6},
		pluralsOrdinal:     nil,
		pluralsRange:       []locales.PluralRule{6},
		decimal:            ".",
		group:              ",",
		timeSeparator:      ":",
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		monthsAbbreviated:  []string{"", "S-Ɔ", "K-Ɔ", "E-Ɔ", "E-O", "E-K", "O-A", "A-K", "D-Ɔ", "F-Ɛ", "Ɔ-A", "Ɔ-O", "M-Ɔ"},
		monthsWide:         []string{"", "Sanda-Ɔpɛpɔn", "Kwakwar-Ɔgyefuo", "Ebɔw-Ɔbenem", "Ebɔbira-Oforisuo", "Esusow Aketseaba-Kɔtɔnimba", "Obirade-Ayɛwohomumu", "Ayɛwoho-Kitawonsa", "Difuu-Ɔsandaa", "Fankwa-Ɛbɔ", "Ɔbɛsɛ-Ahinime", "Ɔberɛfɛw-Obubuo", "Mumu-Ɔpɛnimba"},
		daysAbbreviated:    []string{"Kwe", "Dwo", "Ben", "Wuk", "Yaw", "Fia", "Mem"},
		daysNarrow:         []string{"K", "D", "B", "W", "Y", "F", "M"},
		daysWide:           []string{"Kwesida", "Dwowda", "Benada", "Wukuda", "Yawda", "Fida", "Memeneda"},
		periodsAbbreviated: []string{"AN", "EW"},
		periodsWide:        []string{"AN", "EW"},
		erasAbbreviated:    []string{"AK", "KE"},
		erasNarrow:         []string{"", ""},
		erasWide:           []string{"Ansa Kristo", "Kristo Ekyiri"},
		timezones:          map[string]string{"ECT": "ECT", "WART": "WART", "VET": "VET", "ART": "ART", "GYT": "GYT", "UYT": "UYT", "HNPMX": "HNPMX", "AST": "AST", "MESZ": "MESZ", "LHDT": "LHDT", "HAT": "HAT", "CDT": "CDT", "WAT": "WAT", "BOT": "BOT", "ACST": "ACST", "ACDT": "ACDT", "TMST": "TMST", "OEZ": "OEZ", "WEZ": "WEZ", "NZDT": "NZDT", "EST": "EST", "ACWDT": "ACWDT", "WAST": "WAST", "HENOMX": "HENOMX", "ARST": "ARST", "GMT": "GMT", "ChST": "ChST", "CHADT": "CHADT", "HNCU": "HNCU", "AWST": "AWST", "HKT": "HKT", "GFT": "GFT", "OESZ": "OESZ", "HECU": "HECU", "PDT": "PDT", "ADT": "ADT", "WESZ": "WESZ", "WIB": "WIB", "MYT": "MYT", "IST": "IST", "SRT": "SRT", "TMT": "TMT", "NZST": "NZST", "WARST": "WARST", "HEPM": "HEPM", "CAT": "CAT", "EAT": "EAT", "COT": "COT", "UYST": "UYST", "HEPMX": "HEPMX", "ACWST": "ACWST", "HEEG": "HEEG", "HEOG": "HEOG", "MEZ": "MEZ", "∅∅∅": "∅∅∅", "HNNOMX": "HNNOMX", "WIT": "WIT", "CST": "CST", "AKDT": "AKDT", "SGT": "SGT", "HNT": "HNT", "MDT": "MDT", "CLST": "CLST", "HAST": "HAST", "HADT": "HADT", "CHAST": "CHAST", "AKST": "AKST", "HNEG": "HNEG", "HNOG": "HNOG", "MST": "MST", "AEST": "AEST", "HNPM": "HNPM", "BT": "BT", "LHST": "LHST", "WITA": "WITA", "AWDT": "AWDT", "SAST": "SAST", "HKST": "HKST", "CLT": "CLT", "COST": "COST", "PST": "PST", "JST": "JST", "AEDT": "AEDT", "JDT": "JDT", "EDT": "EDT"},
	}
}

// Locale returns the current translators string locale
func (ak *ak_GH) Locale() string {
	return ak.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'ak_GH'
func (ak *ak_GH) PluralsCardinal() []locales.PluralRule {
	return ak.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'ak_GH'
func (ak *ak_GH) PluralsOrdinal() []locales.PluralRule {
	return ak.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'ak_GH'
func (ak *ak_GH) PluralsRange() []locales.PluralRule {
	return ak.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'ak_GH'
func (ak *ak_GH) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n >= 0 && n <= 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'ak_GH'
func (ak *ak_GH) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'ak_GH'
func (ak *ak_GH) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (ak *ak_GH) MonthAbbreviated(month time.Month) string {
	return ak.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (ak *ak_GH) MonthsAbbreviated() []string {
	return ak.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (ak *ak_GH) MonthNarrow(month time.Month) string {
	return ak.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (ak *ak_GH) MonthsNarrow() []string {
	return nil
}

// MonthWide returns the locales wide month given the 'month' provided
func (ak *ak_GH) MonthWide(month time.Month) string {
	return ak.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (ak *ak_GH) MonthsWide() []string {
	return ak.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (ak *ak_GH) WeekdayAbbreviated(weekday time.Weekday) string {
	return ak.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (ak *ak_GH) WeekdaysAbbreviated() []string {
	return ak.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (ak *ak_GH) WeekdayNarrow(weekday time.Weekday) string {
	return ak.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (ak *ak_GH) WeekdaysNarrow() []string {
	return ak.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (ak *ak_GH) WeekdayShort(weekday time.Weekday) string {
	return ak.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (ak *ak_GH) WeekdaysShort() []string {
	return ak.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (ak *ak_GH) WeekdayWide(weekday time.Weekday) string {
	return ak.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (ak *ak_GH) WeekdaysWide() []string {
	return ak.daysWide
}

// Decimal returns the decimal point of number
func (ak *ak_GH) Decimal() string {
	return ak.decimal
}

// Group returns the group of number
func (ak *ak_GH) Group() string {
	return ak.group
}

// Group returns the minus sign of number
func (ak *ak_GH) Minus() string {
	return ak.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'ak_GH' and handles both Whole and Real numbers based on 'v'
func (ak *ak_GH) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'ak_GH' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (ak *ak_GH) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'ak_GH'
func (ak *ak_GH) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ak.currencies[currency]
	l := len(s) + len(symbol) + 1 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ak.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ak.group[0])
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
		b = append(b, ak.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ak.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'ak_GH'
// in accounting notation.
func (ak *ak_GH) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ak.currencies[currency]
	l := len(s) + len(symbol) + 1 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ak.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ak.group[0])
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

		b = append(b, ak.minus[0])

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
			b = append(b, ak.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'ak_GH'
func (ak *ak_GH) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 9 {
		b = append(b, strconv.Itoa(t.Year())[2:]...)
	} else {
		b = append(b, strconv.Itoa(t.Year())[1:]...)
	}

	b = append(b, []byte{0x2f}...)

	if t.Month() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Month()), 10)

	b = append(b, []byte{0x2f}...)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'ak_GH'
func (ak *ak_GH) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, ak.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'ak_GH'
func (ak *ak_GH) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, ak.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'ak_GH'
func (ak *ak_GH) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, ak.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, ak.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'ak_GH'
func (ak *ak_GH) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ak.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ak.periodsAbbreviated[0]...)
	} else {
		b = append(b, ak.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'ak_GH'
func (ak *ak_GH) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ak.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ak.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ak.periodsAbbreviated[0]...)
	} else {
		b = append(b, ak.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'ak_GH'
func (ak *ak_GH) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ak.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ak.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ak.periodsAbbreviated[0]...)
	} else {
		b = append(b, ak.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'ak_GH'
func (ak *ak_GH) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ak.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ak.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ak.periodsAbbreviated[0]...)
	} else {
		b = append(b, ak.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := ak.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
