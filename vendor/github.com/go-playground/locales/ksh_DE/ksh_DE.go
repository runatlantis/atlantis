package ksh_DE

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type ksh_DE struct {
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

// New returns a new instance of translator for the 'ksh_DE' locale
func New() locales.Translator {
	return &ksh_DE{
		locale:                 "ksh_DE",
		pluralsCardinal:        []locales.PluralRule{1, 2, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ",",
		group:                  " ",
		minus:                  "−",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		percentSuffix:          " ",
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "Jan", "Fäb", "Mäz", "Apr", "Mai", "Jun", "Jul", "Ouj", "Säp", "Okt", "Nov", "Dez"},
		monthsNarrow:           []string{"", "J", "F", "M", "A", "M", "J", "J", "O", "S", "O", "N", "D"},
		monthsWide:             []string{"", "Jannewa", "Fäbrowa", "Määz", "Aprell", "Mai", "Juuni", "Juuli", "Oujoß", "Septämber", "Oktohber", "Novämber", "Dezämber"},
		daysAbbreviated:        []string{"Su.", "Mo.", "Di.", "Me.", "Du.", "Fr.", "Sa."},
		daysNarrow:             []string{"S", "M", "D", "M", "D", "F", "S"},
		daysShort:              []string{"Su", "Mo", "Di", "Me", "Du", "Fr", "Sa"},
		daysWide:               []string{"Sunndaach", "Mohndaach", "Dinnsdaach", "Metwoch", "Dunnersdaach", "Friidaach", "Samsdaach"},
		periodsAbbreviated:     []string{"v.M.", "n.M."},
		periodsWide:            []string{"Uhr vörmiddaachs", "Uhr nommendaachs"},
		erasAbbreviated:        []string{"v. Chr.", "n. Chr."},
		erasNarrow:             []string{"vC", "nC"},
		erasWide:               []string{"vür Krestos", "noh Krestos"},
		timezones:              map[string]string{"HAT": "HAT", "MDT": "MDT", "TMST": "TMST", "HADT": "HADT", "ART": "ART", "ARST": "ARST", "HKT": "HKT", "MESZ": "Meddel-Europpa sing Summerzick", "BOT": "BOT", "UYST": "UYST", "AEDT": "AEDT", "WITA": "WITA", "CLT": "CLT", "COT": "COT", "CHADT": "CHADT", "ADT": "ADT", "WEZ": "Weß-Europpa sing jewöhnlijje Zick", "ACDT": "ACDT", "LHDT": "LHDT", "BT": "BT", "SAST": "Söd-Affrekaanesche Zigg", "AKST": "AKST", "ACWDT": "ACWDT", "HNEG": "HNEG", "UYT": "UYT", "AWST": "AWST", "HNPMX": "HNPMX", "HENOMX": "HENOMX", "MST": "MST", "AWDT": "AWDT", "WESZ": "Weß-Europpa sing Summerzick", "WIB": "WIB", "EST": "EST", "COST": "COST", "CAT": "Zentraal-Affrekaanesche Zigg", "CST": "CST", "EDT": "EDT", "LHST": "LHST", "OESZ": "Oß-Europpa sing Summerzick", "WART": "WART", "TMT": "TMT", "MEZ": "Meddel-Europpa sing jewöhnlijje Zick", "CLST": "CLST", "WIT": "WIT", "HEEG": "HEEG", "HKST": "HKST", "EAT": "Oß-Affrekaanesche Zigg", "OEZ": "Oß-Europpa sing jewöhnlijje Zick", "GMT": "Greenwich sing Standat-Zick", "HECU": "HECU", "NZST": "NZST", "GFT": "GFT", "∅∅∅": "de Azore ier Summerzick", "VET": "VET", "JDT": "JDT", "AEST": "AEST", "WAT": "Jewöhnlijje Wäß-Affrekaanesche Zigg", "WAST": "Wäß-Affrekaanesche Sommerzigg", "HNNOMX": "HNNOMX", "HNCU": "HNCU", "ECT": "ECT", "ACST": "ACST", "HEPM": "HEPM", "HEPMX": "HEPMX", "CDT": "CDT", "GYT": "GYT", "PST": "PST", "NZDT": "NZDT", "MYT": "MYT", "JST": "JST", "HNPM": "HNPM", "HAST": "HAST", "CHAST": "CHAST", "AKDT": "AKDT", "HEOG": "HEOG", "WARST": "WARST", "ChST": "ChST", "AST": "AST", "SGT": "SGT", "ACWST": "ACWST", "HNT": "HNT", "SRT": "SRT", "PDT": "PDT", "HNOG": "HNOG", "IST": "IST"},
	}
}

// Locale returns the current translators string locale
func (ksh *ksh_DE) Locale() string {
	return ksh.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'ksh_DE'
func (ksh *ksh_DE) PluralsCardinal() []locales.PluralRule {
	return ksh.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'ksh_DE'
func (ksh *ksh_DE) PluralsOrdinal() []locales.PluralRule {
	return ksh.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'ksh_DE'
func (ksh *ksh_DE) PluralsRange() []locales.PluralRule {
	return ksh.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'ksh_DE'
func (ksh *ksh_DE) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 0 {
		return locales.PluralRuleZero
	} else if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'ksh_DE'
func (ksh *ksh_DE) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'ksh_DE'
func (ksh *ksh_DE) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (ksh *ksh_DE) MonthAbbreviated(month time.Month) string {
	return ksh.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (ksh *ksh_DE) MonthsAbbreviated() []string {
	return ksh.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (ksh *ksh_DE) MonthNarrow(month time.Month) string {
	return ksh.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (ksh *ksh_DE) MonthsNarrow() []string {
	return ksh.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (ksh *ksh_DE) MonthWide(month time.Month) string {
	return ksh.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (ksh *ksh_DE) MonthsWide() []string {
	return ksh.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (ksh *ksh_DE) WeekdayAbbreviated(weekday time.Weekday) string {
	return ksh.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (ksh *ksh_DE) WeekdaysAbbreviated() []string {
	return ksh.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (ksh *ksh_DE) WeekdayNarrow(weekday time.Weekday) string {
	return ksh.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (ksh *ksh_DE) WeekdaysNarrow() []string {
	return ksh.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (ksh *ksh_DE) WeekdayShort(weekday time.Weekday) string {
	return ksh.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (ksh *ksh_DE) WeekdaysShort() []string {
	return ksh.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (ksh *ksh_DE) WeekdayWide(weekday time.Weekday) string {
	return ksh.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (ksh *ksh_DE) WeekdaysWide() []string {
	return ksh.daysWide
}

// Decimal returns the decimal point of number
func (ksh *ksh_DE) Decimal() string {
	return ksh.decimal
}

// Group returns the group of number
func (ksh *ksh_DE) Group() string {
	return ksh.group
}

// Group returns the minus sign of number
func (ksh *ksh_DE) Minus() string {
	return ksh.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'ksh_DE' and handles both Whole and Real numbers based on 'v'
func (ksh *ksh_DE) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ksh.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(ksh.group) - 1; j >= 0; j-- {
					b = append(b, ksh.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(ksh.minus) - 1; j >= 0; j-- {
			b = append(b, ksh.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'ksh_DE' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (ksh *ksh_DE) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 7
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ksh.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(ksh.minus) - 1; j >= 0; j-- {
			b = append(b, ksh.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, ksh.percentSuffix...)

	b = append(b, ksh.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'ksh_DE'
func (ksh *ksh_DE) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ksh.currencies[currency]
	l := len(s) + len(symbol) + 6 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ksh.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(ksh.group) - 1; j >= 0; j-- {
					b = append(b, ksh.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(ksh.minus) - 1; j >= 0; j-- {
			b = append(b, ksh.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ksh.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, ksh.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'ksh_DE'
// in accounting notation.
func (ksh *ksh_DE) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ksh.currencies[currency]
	l := len(s) + len(symbol) + 6 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ksh.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(ksh.group) - 1; j >= 0; j-- {
					b = append(b, ksh.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		for j := len(ksh.minus) - 1; j >= 0; j-- {
			b = append(b, ksh.minus[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ksh.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, ksh.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, ksh.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'ksh_DE'
func (ksh *ksh_DE) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0x2e, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'ksh_DE'
func (ksh *ksh_DE) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, ksh.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x2e, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'ksh_DE'
func (ksh *ksh_DE) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, ksh.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'ksh_DE'
func (ksh *ksh_DE) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, ksh.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20, 0x64, 0xc3, 0xa4}...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, ksh.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'ksh_DE'
func (ksh *ksh_DE) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ksh.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'ksh_DE'
func (ksh *ksh_DE) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ksh.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ksh.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'ksh_DE'
func (ksh *ksh_DE) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ksh.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ksh.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'ksh_DE'
func (ksh *ksh_DE) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ksh.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ksh.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := ksh.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
