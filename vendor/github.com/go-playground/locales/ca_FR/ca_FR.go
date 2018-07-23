package ca_FR

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type ca_FR struct {
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

// New returns a new instance of translator for the 'ca_FR' locale
func New() locales.Translator {
	return &ca_FR{
		locale:                 "ca_FR",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         []locales.PluralRule{2, 3, 4, 6},
		pluralsRange:           []locales.PluralRule{6},
		decimal:                ",",
		group:                  ".",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "F", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositiveSuffix: " ",
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: " )",
		monthsAbbreviated:      []string{"", "de gen.", "de febr.", "de març", "d’abr.", "de maig", "de juny", "de jul.", "d’ag.", "de set.", "d’oct.", "de nov.", "de des."},
		monthsNarrow:           []string{"", "GN", "FB", "MÇ", "AB", "MG", "JN", "JL", "AG", "ST", "OC", "NV", "DS"},
		monthsWide:             []string{"", "de gener", "de febrer", "de març", "d’abril", "de maig", "de juny", "de juliol", "d’agost", "de setembre", "d’octubre", "de novembre", "de desembre"},
		daysAbbreviated:        []string{"dg.", "dl.", "dt.", "dc.", "dj.", "dv.", "ds."},
		daysNarrow:             []string{"dg", "dl", "dt", "dc", "dj", "dv", "ds"},
		daysShort:              []string{"dg.", "dl.", "dt.", "dc.", "dj.", "dv.", "ds."},
		daysWide:               []string{"diumenge", "dilluns", "dimarts", "dimecres", "dijous", "divendres", "dissabte"},
		periodsAbbreviated:     []string{"a. m.", "p. m."},
		periodsNarrow:          []string{"a. m.", "p. m."},
		periodsWide:            []string{"a. m.", "p. m."},
		erasAbbreviated:        []string{"aC", "dC"},
		erasNarrow:             []string{"aC", "dC"},
		erasWide:               []string{"abans de Crist", "després de Crist"},
		timezones:              map[string]string{"VET": "Hora de Veneçuela", "HNPMX": "Hora estàndard del Pacífic de Mèxic", "MST": "Hora estàndard de muntanya d’Amèrica del Nord", "AEST": "Hora estàndard d’Austràlia Oriental", "WEZ": "Hora estàndard de l’Oest d’Europa", "NZDT": "Hora d’estiu de Nova Zelanda", "EST": "Hora estàndard oriental d’Amèrica del Nord", "MESZ": "Hora d’estiu del Centre d’Europa", "HKT": "Hora estàndard de Hong Kong", "CST": "Hora estàndard central d’Amèrica del Nord", "CDT": "Hora d’estiu central d’Amèrica del Nord", "HNT": "Hora estàndard de Terranova", "CLT": "Hora estàndard de Xile", "HKST": "Hora d’estiu de Hong Kong", "IST": "Hora estàndard de l’Índia", "WITA": "Hora central d’Indonèsia", "COT": "Hora estàndard de Colòmbia", "UYT": "Hora estàndard de l’Uruguai", "HNEG": "Hora estàndard de l’Est de Grenlàndia", "ADT": "Hora d’estiu de l’Atlàntic", "WAT": "Hora estàndard de l’Àfrica Occidental", "EDT": "Hora d’estiu oriental d’Amèrica del Nord", "ACST": "Hora estàndard d’Austràlia Central", "LHST": "Hora estàndard de Lord Howe", "UYST": "Hora d’estiu de l’Uruguai", "CHAST": "Hora estàndard de Chatham", "HEOG": "Hora d’estiu de l’Oest de Grenlàndia", "HNNOMX": "Hora estàndard del nord-oest de Mèxic", "CAT": "Hora de l’Àfrica Central", "AWST": "Hora estàndard d’Austràlia Occidental", "ACWDT": "Hora d’estiu d’Austràlia centre-occidental", "AWDT": "Hora d’estiu d’Austràlia Occidental", "AKDT": "Hora d’estiu d’Alaska", "SGT": "Hora de Singapur", "HENOMX": "Hora d’estiu del nord-oest de Mèxic", "SRT": "Hora de Surinam", "HECU": "Hora d’estiu de Cuba", "PST": "Hora estàndard del Pacífic", "WAST": "Hora d’estiu de l’Àfrica Occidental", "WARST": "Hora d’estiu de l’oest de l’Argentina", "OEZ": "Hora estàndard de l’Est d’Europa", "AEDT": "Hora d’estiu d’Austràlia Oriental", "MDT": "Hora d’estiu de muntanya d’Amèrica del Nord", "BOT": "Hora de Bolívia", "ACWST": "Hora estàndard d’Austràlia centre-occidental", "WIT": "Hora de l’est d’Indonèsia", "TMST": "Hora d’estiu del Turkmenistan", "HEPMX": "Hora d’estiu del Pacífic de Mèxic", "WIB": "Hora de l’oest d’Indonèsia", "GFT": "Hora de la Guaiana Francesa", "HAST": "Hora estàndard de Hawaii-Aleutianes", "HADT": "Hora d’estiu de Hawaii-Aleutianes", "HNCU": "Hora estàndard de Cuba", "WESZ": "Hora d’estiu de l’Oest d’Europa", "AST": "Hora estàndard de l’Atlàntic", "SAST": "Hora estàndard del sud de l’Àfrica", "ACDT": "Hora d’estiu d’Austràlia Central", "WART": "Hora estàndard de l’oest de l’Argentina", "ART": "Hora estàndard de l’Argentina", "GYT": "Hora de Guyana", "PDT": "Hora d’estiu del Pacífic", "HNOG": "Hora estàndard de l’Oest de Grenlàndia", "EAT": "Hora de l’Àfrica Oriental", "JST": "Hora estàndard del Japó", "HEEG": "Hora d’estiu de l’Est de Grenlàndia", "ChST": "Hora de Chamorro", "HNPM": "Hora estàndard de Saint-Pierre i Miquelon", "AKST": "Hora estàndard d’Alaska", "ECT": "Hora de l’Equador", "MEZ": "Hora estàndard del Centre d’Europa", "CLST": "Hora d’estiu de Xile", "TMT": "Hora estàndard del Turkmenistan", "GMT": "Hora del Meridià de Greenwich", "MYT": "Hora de Malàisia", "OESZ": "Hora d’estiu de l’Est d’Europa", "BT": "Hora de Bhutan", "JDT": "Hora d’estiu del Japó", "HAT": "Hora d’estiu de Terranova", "ARST": "Hora d’estiu de l’Argentina", "CHADT": "Hora d’estiu de Chatham", "∅∅∅": "Hora d’estiu de Brasília", "HEPM": "Hora d’estiu de Saint-Pierre i Miquelon", "COST": "Hora d’estiu de Colòmbia", "NZST": "Hora estàndard de Nova Zelanda", "LHDT": "Horari d’estiu de Lord Howe"},
	}
}

// Locale returns the current translators string locale
func (ca *ca_FR) Locale() string {
	return ca.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'ca_FR'
func (ca *ca_FR) PluralsCardinal() []locales.PluralRule {
	return ca.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'ca_FR'
func (ca *ca_FR) PluralsOrdinal() []locales.PluralRule {
	return ca.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'ca_FR'
func (ca *ca_FR) PluralsRange() []locales.PluralRule {
	return ca.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'ca_FR'
func (ca *ca_FR) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)

	if i == 1 && v == 0 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'ca_FR'
func (ca *ca_FR) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 || n == 3 {
		return locales.PluralRuleOne
	} else if n == 2 {
		return locales.PluralRuleTwo
	} else if n == 4 {
		return locales.PluralRuleFew
	}

	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'ca_FR'
func (ca *ca_FR) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (ca *ca_FR) MonthAbbreviated(month time.Month) string {
	return ca.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (ca *ca_FR) MonthsAbbreviated() []string {
	return ca.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (ca *ca_FR) MonthNarrow(month time.Month) string {
	return ca.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (ca *ca_FR) MonthsNarrow() []string {
	return ca.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (ca *ca_FR) MonthWide(month time.Month) string {
	return ca.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (ca *ca_FR) MonthsWide() []string {
	return ca.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (ca *ca_FR) WeekdayAbbreviated(weekday time.Weekday) string {
	return ca.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (ca *ca_FR) WeekdaysAbbreviated() []string {
	return ca.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (ca *ca_FR) WeekdayNarrow(weekday time.Weekday) string {
	return ca.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (ca *ca_FR) WeekdaysNarrow() []string {
	return ca.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (ca *ca_FR) WeekdayShort(weekday time.Weekday) string {
	return ca.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (ca *ca_FR) WeekdaysShort() []string {
	return ca.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (ca *ca_FR) WeekdayWide(weekday time.Weekday) string {
	return ca.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (ca *ca_FR) WeekdaysWide() []string {
	return ca.daysWide
}

// Decimal returns the decimal point of number
func (ca *ca_FR) Decimal() string {
	return ca.decimal
}

// Group returns the group of number
func (ca *ca_FR) Group() string {
	return ca.group
}

// Group returns the minus sign of number
func (ca *ca_FR) Minus() string {
	return ca.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'ca_FR' and handles both Whole and Real numbers based on 'v'
func (ca *ca_FR) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ca.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ca.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ca.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'ca_FR' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (ca *ca_FR) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ca.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ca.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, ca.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'ca_FR'
func (ca *ca_FR) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ca.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ca.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ca.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ca.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ca.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, ca.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'ca_FR'
// in accounting notation.
func (ca *ca_FR) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ca.currencies[currency]
	l := len(s) + len(symbol) + 6 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ca.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ca.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, ca.currencyNegativePrefix[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ca.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, ca.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, ca.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'ca_FR'
func (ca *ca_FR) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2f}...)
	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0x2f}...)

	if t.Year() > 9 {
		b = append(b, strconv.Itoa(t.Year())[2:]...)
	} else {
		b = append(b, strconv.Itoa(t.Year())[1:]...)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'ca_FR'
func (ca *ca_FR) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ca.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'ca_FR'
func (ca *ca_FR) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ca.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20, 0x64, 0x65}...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'ca_FR'
func (ca *ca_FR) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, ca.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ca.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20, 0x64, 0x65}...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'ca_FR'
func (ca *ca_FR) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ca.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'ca_FR'
func (ca *ca_FR) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ca.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ca.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'ca_FR'
func (ca *ca_FR) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ca.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ca.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'ca_FR'
func (ca *ca_FR) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ca.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ca.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := ca.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
