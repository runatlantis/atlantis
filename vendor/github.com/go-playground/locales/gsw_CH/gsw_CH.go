package gsw_CH

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type gsw_CH struct {
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

// New returns a new instance of translator for the 'gsw_CH' locale
func New() locales.Translator {
	return &gsw_CH{
		locale:                 "gsw_CH",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         []locales.PluralRule{6},
		pluralsRange:           []locales.PluralRule{2, 6},
		decimal:                ".",
		group:                  "’",
		minus:                  "−",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		percentSuffix:          " ",
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "Jan", "Feb", "Mär", "Apr", "Mai", "Jun", "Jul", "Aug", "Sep", "Okt", "Nov", "Dez"},
		monthsNarrow:           []string{"", "J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"},
		monthsWide:             []string{"", "Januar", "Februar", "März", "April", "Mai", "Juni", "Juli", "Auguscht", "Septämber", "Oktoober", "Novämber", "Dezämber"},
		daysAbbreviated:        []string{"Su.", "Mä.", "Zi.", "Mi.", "Du.", "Fr.", "Sa."},
		daysNarrow:             []string{"S", "M", "D", "M", "D", "F", "S"},
		daysWide:               []string{"Sunntig", "Määntig", "Ziischtig", "Mittwuch", "Dunschtig", "Friitig", "Samschtig"},
		periodsAbbreviated:     []string{"vorm.", "nam."},
		periodsWide:            []string{"am Vormittag", "am Namittag"},
		erasAbbreviated:        []string{"v. Chr.", "n. Chr."},
		erasNarrow:             []string{"v. Chr.", "n. Chr."},
		erasWide:               []string{"v. Chr.", "n. Chr."},
		timezones:              map[string]string{"MEZ": "Mitteleuropäischi Schtandardziit", "LHST": "LHST", "LHDT": "LHDT", "VET": "VET", "COST": "COST", "PST": "PST", "JDT": "JDT", "ACST": "ACST", "HENOMX": "HENOMX", "CAT": "Zentralafrikanischi Ziit", "CHADT": "CHADT", "WIB": "WIB", "ACWST": "ACWST", "HNEG": "HNEG", "HNNOMX": "HNNOMX", "ART": "ART", "COT": "COT", "WARST": "WARST", "AEST": "AEST", "AST": "AST", "WITA": "WITA", "WIT": "WIT", "HNCU": "HNCU", "ADT": "ADT", "SAST": "Süüdafrikanischi ziit", "NZST": "NZST", "HEEG": "HEEG", "HNOG": "HNOG", "ARST": "ARST", "ChST": "ChST", "AWDT": "AWDT", "GFT": "GFT", "ECT": "ECT", "HAST": "HAST", "UYST": "UYST", "GMT": "GMT", "WEZ": "Weschteuropäischi Schtandardziit", "∅∅∅": "Acre-Summerziit", "BT": "BT", "SGT": "SGT", "EST": "EST", "WART": "WART", "HECU": "HECU", "MYT": "MYT", "HKST": "HKST", "MST": "MST", "MDT": "MDT", "PDT": "PDT", "OEZ": "Oschteuropäischi Schtandardziit", "NZDT": "NZDT", "ACWDT": "ACWDT", "HKT": "HKT", "HEPM": "HEPM", "SRT": "SRT", "EAT": "Oschtafrikanischi Ziit", "TMST": "TMST", "OESZ": "Oschteuropäischi Summerziit", "HNPMX": "HNPMX", "CDT": "Amerika-Zentraal Summerziit", "BOT": "BOT", "AKST": "Alaska-Schtandardziit", "TMT": "TMT", "GYT": "GYT", "EDT": "EDT", "HAT": "HAT", "HEPMX": "HEPMX", "CST": "Amerika-Zentraal Schtandardziit", "WAST": "Weschtafrikanischi Summerziit", "JST": "JST", "MESZ": "Mitteleuropäischi Summerziit", "HADT": "HADT", "UYT": "UYT", "WAT": "Weschtafrikanischi Schtandardziit", "ACDT": "ACDT", "CHAST": "CHAST", "AEDT": "AEDT", "CLST": "CLST", "WESZ": "Weschteuropäischi Summerziit", "AKDT": "Alaska-Summerziit", "HEOG": "HEOG", "IST": "IST", "HNPM": "HNPM", "HNT": "HNT", "CLT": "CLT", "AWST": "AWST"},
	}
}

// Locale returns the current translators string locale
func (gsw *gsw_CH) Locale() string {
	return gsw.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'gsw_CH'
func (gsw *gsw_CH) PluralsCardinal() []locales.PluralRule {
	return gsw.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'gsw_CH'
func (gsw *gsw_CH) PluralsOrdinal() []locales.PluralRule {
	return gsw.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'gsw_CH'
func (gsw *gsw_CH) PluralsRange() []locales.PluralRule {
	return gsw.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'gsw_CH'
func (gsw *gsw_CH) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'gsw_CH'
func (gsw *gsw_CH) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'gsw_CH'
func (gsw *gsw_CH) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := gsw.CardinalPluralRule(num1, v1)
	end := gsw.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (gsw *gsw_CH) MonthAbbreviated(month time.Month) string {
	return gsw.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (gsw *gsw_CH) MonthsAbbreviated() []string {
	return gsw.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (gsw *gsw_CH) MonthNarrow(month time.Month) string {
	return gsw.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (gsw *gsw_CH) MonthsNarrow() []string {
	return gsw.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (gsw *gsw_CH) MonthWide(month time.Month) string {
	return gsw.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (gsw *gsw_CH) MonthsWide() []string {
	return gsw.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (gsw *gsw_CH) WeekdayAbbreviated(weekday time.Weekday) string {
	return gsw.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (gsw *gsw_CH) WeekdaysAbbreviated() []string {
	return gsw.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (gsw *gsw_CH) WeekdayNarrow(weekday time.Weekday) string {
	return gsw.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (gsw *gsw_CH) WeekdaysNarrow() []string {
	return gsw.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (gsw *gsw_CH) WeekdayShort(weekday time.Weekday) string {
	return gsw.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (gsw *gsw_CH) WeekdaysShort() []string {
	return gsw.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (gsw *gsw_CH) WeekdayWide(weekday time.Weekday) string {
	return gsw.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (gsw *gsw_CH) WeekdaysWide() []string {
	return gsw.daysWide
}

// Decimal returns the decimal point of number
func (gsw *gsw_CH) Decimal() string {
	return gsw.decimal
}

// Group returns the group of number
func (gsw *gsw_CH) Group() string {
	return gsw.group
}

// Group returns the minus sign of number
func (gsw *gsw_CH) Minus() string {
	return gsw.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'gsw_CH' and handles both Whole and Real numbers based on 'v'
func (gsw *gsw_CH) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 4 + 3*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, gsw.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(gsw.group) - 1; j >= 0; j-- {
					b = append(b, gsw.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(gsw.minus) - 1; j >= 0; j-- {
			b = append(b, gsw.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'gsw_CH' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (gsw *gsw_CH) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 7
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, gsw.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(gsw.minus) - 1; j >= 0; j-- {
			b = append(b, gsw.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, gsw.percentSuffix...)

	b = append(b, gsw.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'gsw_CH'
func (gsw *gsw_CH) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := gsw.currencies[currency]
	l := len(s) + len(symbol) + 6 + 3*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, gsw.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(gsw.group) - 1; j >= 0; j-- {
					b = append(b, gsw.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(gsw.minus) - 1; j >= 0; j-- {
			b = append(b, gsw.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, gsw.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, gsw.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'gsw_CH'
// in accounting notation.
func (gsw *gsw_CH) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := gsw.currencies[currency]
	l := len(s) + len(symbol) + 6 + 3*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, gsw.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(gsw.group) - 1; j >= 0; j-- {
					b = append(b, gsw.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		for j := len(gsw.minus) - 1; j >= 0; j-- {
			b = append(b, gsw.minus[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, gsw.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, gsw.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, gsw.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'gsw_CH'
func (gsw *gsw_CH) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Month() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Month()), 10)

	b = append(b, []byte{0x2e}...)

	if t.Year() > 9 {
		b = append(b, strconv.Itoa(t.Year())[2:]...)
	} else {
		b = append(b, strconv.Itoa(t.Year())[1:]...)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'gsw_CH'
func (gsw *gsw_CH) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Month() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Month()), 10)

	b = append(b, []byte{0x2e}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'gsw_CH'
func (gsw *gsw_CH) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, gsw.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'gsw_CH'
func (gsw *gsw_CH) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, gsw.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, gsw.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'gsw_CH'
func (gsw *gsw_CH) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, gsw.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'gsw_CH'
func (gsw *gsw_CH) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, gsw.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, gsw.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'gsw_CH'
func (gsw *gsw_CH) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, gsw.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, gsw.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'gsw_CH'
func (gsw *gsw_CH) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, gsw.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, gsw.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := gsw.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
