package fo

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type fo struct {
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

// New returns a new instance of translator for the 'fo' locale
func New() locales.Translator {
	return &fo{
		locale:                 "fo",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ",",
		group:                  ".",
		minus:                  "−",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "A$", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "kr", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "₪", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JP¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "₩", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZ$", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "NT$", "TZS", "UAH", "UAK", "UGS", "UGX", "US$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "₫", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "EC$", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		percentSuffix:          " ",
		currencyPositiveSuffix: " ",
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: " )",
		monthsAbbreviated:      []string{"", "jan.", "feb.", "mar.", "apr.", "mai", "jun.", "jul.", "aug.", "sep.", "okt.", "nov.", "des."},
		monthsNarrow:           []string{"", "J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"},
		monthsWide:             []string{"", "januar", "februar", "mars", "apríl", "mai", "juni", "juli", "august", "september", "oktober", "november", "desember"},
		daysAbbreviated:        []string{"sun.", "mán.", "týs.", "mik.", "hós.", "frí.", "ley."},
		daysNarrow:             []string{"S", "M", "T", "M", "H", "F", "L"},
		daysShort:              []string{"su.", "má.", "tý.", "mi.", "hó.", "fr.", "le."},
		daysWide:               []string{"sunnudagur", "mánadagur", "týsdagur", "mikudagur", "hósdagur", "fríggjadagur", "leygardagur"},
		periodsAbbreviated:     []string{"AM", "PM"},
		periodsNarrow:          []string{"AM", "PM"},
		periodsWide:            []string{"AM", "PM"},
		erasAbbreviated:        []string{"f.Kr.", "e.Kr."},
		erasNarrow:             []string{"fKr", "eKr"},
		erasWide:               []string{"fyri Krist", "eftir Krist"},
		timezones:              map[string]string{"HENOMX": "Northwest Mexico summartíð", "CLT": "Kili vanlig tíð", "HEPMX": "Mexican Pacific summartíð", "WEZ": "Vesturevropa vanlig tíð", "ECT": "Ekvador tíð", "EST": "Eastern vanlig tíð", "MESZ": "Miðevropa summartíð", "WART": "Vestur Argentina vanlig tíð", "HAT": "Newfoundland summartíð", "VET": "Venesuela tíð", "WARST": "Vestur Argentina summartíð", "SRT": "Surinam tíð", "CLST": "Kili summartíð", "HNCU": "Cuba vanlig tíð", "PDT": "Pacific summartíð", "BT": "Butan tíð", "GFT": "Franska Gujana tíð", "ACWST": "miðvestur Avstralia vanlig tíð", "HEEG": "Eystur grønlendsk summartíð", "HKST": "Hong Kong summartíð", "MDT": "MDT", "AEDT": "eystur Avstralia summartíð", "BOT": "Bolivia tíð", "JST": "Japan vanlig tíð", "ACST": "mið Avstralia vanlig tíð", "WIT": "Eystur Indonesia tíð", "TMT": "Turkmenistan vanlig tíð", "COST": "Kolombia summartíð", "GMT": "Greenwich Mean tíð", "HNNOMX": "Northwest Mexico vanlig tíð", "CAT": "Miðafrika tíð", "HADT": "Hawaii-Aleutian summartíð", "COT": "Kolombia vanlig tíð", "CDT": "Central summartíð", "OESZ": "Eysturevropa summartíð", "∅∅∅": "Azorurnar summartíð", "HAST": "Hawaii-Aleutian vanlig tíð", "CHADT": "Chatham summartíð", "HNPMX": "Mexican Pacific vanlig tíð", "MEZ": "Miðevropa vanlig tíð", "AST": "Atlantic vanlig tíð", "AEST": "eystur Avstralia vanlig tíð", "NZDT": "Nýsæland summartíð", "EDT": "Eastern summartíð", "HEOG": "Vestur grønlendsk summartíð", "HKT": "Hong Kong vanlig tíð", "HNPM": "St. Pierre & Miquelon vanlig tíð", "OEZ": "Eysturevropa vanlig tíð", "UYST": "Uruguai summartíð", "AWDT": "vestur Avstralia summartíð", "ACWDT": "miðvestur Avstralia summartíð", "LHDT": "Lord Howe summartíð", "WITA": "Mið Indonesia tíð", "ADT": "Atlantic summartíð", "HEPM": "St. Pierre & Miquelon summartíð", "MST": "MST", "CST": "Central vanlig tíð", "PST": "Pacific vanlig tíð", "WIB": "Vestur Indonesia tíð", "AKDT": "Alaska summartíð", "SGT": "Singapor tíð", "HNEG": "Eystur grønlendsk vanlig tíð", "LHST": "Lord Howe vanlig tíð", "HNT": "Newfoundland vanlig tíð", "UYT": "Uruguai vanlig tíð", "ChST": "Chamorro vanlig tíð", "EAT": "Eysturafrika tíð", "CHAST": "Chatham vanlig tíð", "AWST": "vestur Avstralia vanlig tíð", "SAST": "Suðurafrika vanlig tíð", "WESZ": "Vesturevropa summartíð", "HNOG": "Vestur grønlendsk vanlig tíð", "GYT": "Gujana tíð", "HECU": "Cuba summartíð", "WAT": "Vesturafrika vanlig tíð", "NZST": "Nýsæland vanlig tíð", "JDT": "Japan summartíð", "AKST": "Alaska vanlig tíð", "ACDT": "mið Avstralia summartíð", "IST": "India tíð", "TMST": "Turkmenistan summartíð", "ART": "Argentina vanlig tíð", "ARST": "Argentina summartíð", "WAST": "Vesturafrika summartíð", "MYT": "Malaisia tíð"},
	}
}

// Locale returns the current translators string locale
func (fo *fo) Locale() string {
	return fo.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'fo'
func (fo *fo) PluralsCardinal() []locales.PluralRule {
	return fo.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'fo'
func (fo *fo) PluralsOrdinal() []locales.PluralRule {
	return fo.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'fo'
func (fo *fo) PluralsRange() []locales.PluralRule {
	return fo.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'fo'
func (fo *fo) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'fo'
func (fo *fo) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'fo'
func (fo *fo) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (fo *fo) MonthAbbreviated(month time.Month) string {
	return fo.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (fo *fo) MonthsAbbreviated() []string {
	return fo.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (fo *fo) MonthNarrow(month time.Month) string {
	return fo.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (fo *fo) MonthsNarrow() []string {
	return fo.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (fo *fo) MonthWide(month time.Month) string {
	return fo.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (fo *fo) MonthsWide() []string {
	return fo.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (fo *fo) WeekdayAbbreviated(weekday time.Weekday) string {
	return fo.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (fo *fo) WeekdaysAbbreviated() []string {
	return fo.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (fo *fo) WeekdayNarrow(weekday time.Weekday) string {
	return fo.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (fo *fo) WeekdaysNarrow() []string {
	return fo.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (fo *fo) WeekdayShort(weekday time.Weekday) string {
	return fo.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (fo *fo) WeekdaysShort() []string {
	return fo.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (fo *fo) WeekdayWide(weekday time.Weekday) string {
	return fo.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (fo *fo) WeekdaysWide() []string {
	return fo.daysWide
}

// Decimal returns the decimal point of number
func (fo *fo) Decimal() string {
	return fo.decimal
}

// Group returns the group of number
func (fo *fo) Group() string {
	return fo.group
}

// Group returns the minus sign of number
func (fo *fo) Minus() string {
	return fo.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'fo' and handles both Whole and Real numbers based on 'v'
func (fo *fo) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, fo.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, fo.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(fo.minus) - 1; j >= 0; j-- {
			b = append(b, fo.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'fo' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (fo *fo) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 7
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, fo.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(fo.minus) - 1; j >= 0; j-- {
			b = append(b, fo.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, fo.percentSuffix...)

	b = append(b, fo.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'fo'
func (fo *fo) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := fo.currencies[currency]
	l := len(s) + len(symbol) + 6 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, fo.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, fo.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(fo.minus) - 1; j >= 0; j-- {
			b = append(b, fo.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, fo.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, fo.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'fo'
// in accounting notation.
func (fo *fo) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := fo.currencies[currency]
	l := len(s) + len(symbol) + 8 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, fo.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, fo.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, fo.currencyNegativePrefix[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, fo.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, fo.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, fo.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'fo'
func (fo *fo) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'fo'
func (fo *fo) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'fo'
func (fo *fo) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, fo.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'fo'
func (fo *fo) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, fo.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, fo.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'fo'
func (fo *fo) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, fo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'fo'
func (fo *fo) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, fo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, fo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'fo'
func (fo *fo) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, fo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, fo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'fo'
func (fo *fo) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, fo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, fo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := fo.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
