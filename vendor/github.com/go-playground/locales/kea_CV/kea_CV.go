package kea_CV

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type kea_CV struct {
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

// New returns a new instance of translator for the 'kea_CV' locale
func New() locales.Translator {
	return &kea_CV{
		locale:                 "kea_CV",
		pluralsCardinal:        []locales.PluralRule{6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ",",
		group:                  " ",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositiveSuffix: " ",
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: " )",
		monthsAbbreviated:      []string{"", "Jan", "Feb", "Mar", "Abr", "Mai", "Jun", "Jul", "Ago", "Set", "Otu", "Nuv", "Diz"},
		monthsNarrow:           []string{"", "J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"},
		monthsWide:             []string{"", "Janeru", "Febreru", "Marsu", "Abril", "Maiu", "Junhu", "Julhu", "Agostu", "Setenbru", "Otubru", "Nuvenbru", "Dizenbru"},
		daysAbbreviated:        []string{"dum", "sig", "ter", "kua", "kin", "ses", "sab"},
		daysNarrow:             []string{"D", "S", "T", "K", "K", "S", "S"},
		daysShort:              []string{"du", "si", "te", "ku", "ki", "se", "sa"},
		daysWide:               []string{"dumingu", "sigunda-fera", "tersa-fera", "kuarta-fera", "kinta-fera", "sesta-fera", "sabadu"},
		periodsAbbreviated:     []string{"am", "pm"},
		periodsNarrow:          []string{"a", "p"},
		periodsWide:            []string{"am", "pm"},
		erasAbbreviated:        []string{"AK", "DK"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"Antis di Kristu", "Dispos di Kristu"},
		timezones:              map[string]string{"CHAST": "CHAST", "AWDT": "Ora di Verãu di Australia Osidental", "AEST": "Ora Padrãu di Australia Oriental", "JST": "JST", "LHST": "LHST", "CDT": "Ora Sentral di Verãu", "AWST": "Ora Padrãu di Australia Osidental", "MST": "Ora di Montanha Padrãu", "MYT": "MYT", "HNEG": "HNEG", "IST": "IST", "ART": "ART", "ACWDT": "Ora di Verãu di Australia Sentru-Osidental", "HKST": "HKST", "WITA": "WITA", "GYT": "GYT", "HEPMX": "HEPMX", "AKST": "AKST", "EST": "Ora Oriental Padrãu", "ADT": "Ora di Verãu di Atlantiku", "HNPM": "HNPM", "CLT": "CLT", "NZST": "NZST", "BOT": "BOT", "SRT": "SRT", "UYST": "UYST", "HKT": "HKT", "WART": "WART", "HEPM": "HEPM", "OESZ": "Ora di Verãu di Europa Oriental", "ARST": "ARST", "HNPMX": "HNPMX", "WESZ": "Ora di Verãu di Europa Osidental", "ACWST": "Ora Padrãu di Australia Sentru-Osidental", "HNOG": "HNOG", "WARST": "WARST", "HNT": "HNT", "HAT": "HAT", "VET": "VET", "CLST": "CLST", "PST": "Ora di Pasifiku Padrãu", "AEDT": "Ora di Verãu di Australia Oriental", "WEZ": "Ora Padrãu di Europa Osidental", "AKDT": "AKDT", "HEEG": "HEEG", "MESZ": "Ora di Verãu di Europa Sentral", "TMT": "TMT", "WAT": "Ora Padrãu di Afrika Osidental", "JDT": "JDT", "NZDT": "NZDT", "EDT": "Ora Oriental di Verãu", "CAT": "Ora di Afrika Sentral", "GFT": "GFT", "SGT": "SGT", "ACST": "Ora Padrãu di Australia Sentral", "EAT": "Ora di Afrika Oriental", "WIT": "WIT", "MDT": "Ora di Verãu di Montanha", "TMST": "TMST", "HAST": "HAST", "HNCU": "HNCU", "HECU": "HECU", "AST": "Ora Padrãu di Atlantiku", "HEOG": "HEOG", "MEZ": "Ora Padrãu di Europa Sentral", "GMT": "GMT", "∅∅∅": "∅∅∅", "PDT": "Ora di Pasifiku di Verãu", "SAST": "Ora di Sul di Afrika", "ACDT": "Ora di Verãu di Australia Sentral", "HNNOMX": "HNNOMX", "HENOMX": "HENOMX", "OEZ": "Ora Padrãu di Europa Oriental", "COST": "COST", "ChST": "ChST", "CST": "Ora Sentral Padrãu", "WIB": "WIB", "ECT": "ECT", "LHDT": "LHDT", "HADT": "HADT", "COT": "COT", "UYT": "UYT", "CHADT": "CHADT", "WAST": "Ora di Verão di Afrika Osidental", "BT": "BT"},
	}
}

// Locale returns the current translators string locale
func (kea *kea_CV) Locale() string {
	return kea.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'kea_CV'
func (kea *kea_CV) PluralsCardinal() []locales.PluralRule {
	return kea.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'kea_CV'
func (kea *kea_CV) PluralsOrdinal() []locales.PluralRule {
	return kea.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'kea_CV'
func (kea *kea_CV) PluralsRange() []locales.PluralRule {
	return kea.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'kea_CV'
func (kea *kea_CV) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'kea_CV'
func (kea *kea_CV) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'kea_CV'
func (kea *kea_CV) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (kea *kea_CV) MonthAbbreviated(month time.Month) string {
	return kea.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (kea *kea_CV) MonthsAbbreviated() []string {
	return kea.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (kea *kea_CV) MonthNarrow(month time.Month) string {
	return kea.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (kea *kea_CV) MonthsNarrow() []string {
	return kea.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (kea *kea_CV) MonthWide(month time.Month) string {
	return kea.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (kea *kea_CV) MonthsWide() []string {
	return kea.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (kea *kea_CV) WeekdayAbbreviated(weekday time.Weekday) string {
	return kea.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (kea *kea_CV) WeekdaysAbbreviated() []string {
	return kea.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (kea *kea_CV) WeekdayNarrow(weekday time.Weekday) string {
	return kea.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (kea *kea_CV) WeekdaysNarrow() []string {
	return kea.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (kea *kea_CV) WeekdayShort(weekday time.Weekday) string {
	return kea.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (kea *kea_CV) WeekdaysShort() []string {
	return kea.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (kea *kea_CV) WeekdayWide(weekday time.Weekday) string {
	return kea.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (kea *kea_CV) WeekdaysWide() []string {
	return kea.daysWide
}

// Decimal returns the decimal point of number
func (kea *kea_CV) Decimal() string {
	return kea.decimal
}

// Group returns the group of number
func (kea *kea_CV) Group() string {
	return kea.group
}

// Group returns the minus sign of number
func (kea *kea_CV) Minus() string {
	return kea.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'kea_CV' and handles both Whole and Real numbers based on 'v'
func (kea *kea_CV) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, kea.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(kea.group) - 1; j >= 0; j-- {
					b = append(b, kea.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, kea.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'kea_CV' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (kea *kea_CV) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, kea.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, kea.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, kea.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'kea_CV'
func (kea *kea_CV) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := kea.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, kea.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(kea.group) - 1; j >= 0; j-- {
					b = append(b, kea.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, kea.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, kea.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, kea.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'kea_CV'
// in accounting notation.
func (kea *kea_CV) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := kea.currencies[currency]
	l := len(s) + len(symbol) + 6 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, kea.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(kea.group) - 1; j >= 0; j-- {
					b = append(b, kea.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, kea.currencyNegativePrefix[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, kea.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, kea.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, kea.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'kea_CV'
func (kea *kea_CV) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2f}...)
	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0x2f}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'kea_CV'
func (kea *kea_CV) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, kea.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'kea_CV'
func (kea *kea_CV) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20, 0x64, 0x69}...)
	b = append(b, []byte{0x20}...)
	b = append(b, kea.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20, 0x64, 0x69}...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'kea_CV'
func (kea *kea_CV) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, kea.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20, 0x64, 0x69}...)
	b = append(b, []byte{0x20}...)
	b = append(b, kea.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20, 0x64, 0x69}...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'kea_CV'
func (kea *kea_CV) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, kea.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'kea_CV'
func (kea *kea_CV) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, kea.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, kea.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'kea_CV'
func (kea *kea_CV) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, kea.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, kea.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'kea_CV'
func (kea *kea_CV) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, kea.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, kea.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := kea.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
