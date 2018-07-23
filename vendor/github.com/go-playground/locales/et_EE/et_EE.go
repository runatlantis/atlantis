package et_EE

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type et_EE struct {
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

// New returns a new instance of translator for the 'et_EE' locale
func New() locales.Translator {
	return &et_EE{
		locale:                 "et_EE",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         []locales.PluralRule{6},
		pluralsRange:           []locales.PluralRule{6},
		decimal:                ",",
		group:                  " ",
		minus:                  "−",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositiveSuffix: " ",
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: " )",
		monthsAbbreviated:      []string{"", "jaan", "veebr", "märts", "apr", "mai", "juuni", "juuli", "aug", "sept", "okt", "nov", "dets"},
		monthsNarrow:           []string{"", "J", "V", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"},
		monthsWide:             []string{"", "jaanuar", "veebruar", "märts", "aprill", "mai", "juuni", "juuli", "august", "september", "oktoober", "november", "detsember"},
		daysAbbreviated:        []string{"P", "E", "T", "K", "N", "R", "L"},
		daysNarrow:             []string{"P", "E", "T", "K", "N", "R", "L"},
		daysShort:              []string{"P", "E", "T", "K", "N", "R", "L"},
		daysWide:               []string{"pühapäev", "esmaspäev", "teisipäev", "kolmapäev", "neljapäev", "reede", "laupäev"},
		periodsAbbreviated:     []string{"AM", "PM"},
		periodsNarrow:          []string{"AM", "PM"},
		periodsWide:            []string{"AM", "PM"},
		erasAbbreviated:        []string{"eKr", "pKr"},
		erasNarrow:             []string{"eKr", "pKr"},
		erasWide:               []string{"enne Kristust", "pärast Kristust"},
		timezones:              map[string]string{"SGT": "Singapuri standardaeg", "SRT": "Suriname aeg", "CLT": "Tšiili standardaeg", "UYT": "Uruguay standardaeg", "CHADT": "Chathami suveaeg", "AWST": "Lääne-Austraalia standardaeg", "AWDT": "Lääne-Austraalia suveaeg", "HAST": "Hawaii-Aleuudi standardaeg", "GMT": "Greenwichi aeg", "LHDT": "Lord Howe’i suveaeg", "HNPM": "Saint-Pierre’i ja Miqueloni standardaeg", "GYT": "Guyana aeg", "OESZ": "Ida-Euroopa suveaeg", "ARST": "Argentina suveaeg", "CAT": "Kesk-Aafrika aeg", "TMST": "Türkmenistani suveaeg", "EDT": "Idaranniku suveaeg", "WITA": "Kesk-Indoneesia aeg", "ART": "Argentina standardaeg", "HNCU": "Kuuba standardaeg", "JDT": "Jaapani suveaeg", "HNT": "Newfoundlandi standardaeg", "WIB": "Lääne-Indoneesia aeg", "JST": "Jaapani standardaeg", "ACDT": "Kesk-Austraalia suveaeg", "CLST": "Tšiili suveaeg", "TMT": "Türkmenistani standardaeg", "HEPMX": "Mehhiko Vaikse ookeani suveaeg", "MDT": "Mäestikuvööndi suveaeg", "AEST": "Ida-Austraalia standardaeg", "ACST": "Kesk-Austraalia standardaeg", "WAT": "Lääne-Aafrika standardaeg", "WAST": "Lääne-Aafrika suveaeg", "MYT": "Malaisia \u200b\u200baeg", "WARST": "Lääne-Argentina suveaeg", "ADT": "Atlandi suveaeg", "AEDT": "Ida-Austraalia suveaeg", "HNOG": "Lääne-Gröönimaa standardaeg", "HKST": "Hongkongi suveaeg", "IST": "India aeg", "COT": "Colombia standardaeg", "AST": "Atlandi standardaeg", "HEEG": "Ida-Gröönimaa suveaeg", "ECT": "Ecuadori aeg", "MESZ": "Kesk-Euroopa suveaeg", "WART": "Lääne-Argentina standardaeg", "HAT": "Newfoundlandi suveaeg", "BT": "Bhutani aeg", "BOT": "Boliivia aeg", "MST": "Mäestikuvööndi standardaeg", "NZST": "Uus-Meremaa standardaeg", "AKDT": "Alaska suveaeg", "MEZ": "Kesk-Euroopa standardaeg", "HNNOMX": "Loode-Mehhiko standardaeg", "HENOMX": "Loode-Mehhiko suveaeg", "CST": "Kesk-Ameerika standardaeg", "PST": "Vaikse ookeani standardaeg", "HEPM": "Saint-Pierre’i ja Miqueloni suveaeg", "OEZ": "Ida-Euroopa standardaeg", "HECU": "Kuuba suveaeg", "∅∅∅": "Acre suveaeg", "NZDT": "Uus-Meremaa suveaeg", "EAT": "Ida-Aafrika aeg", "COST": "Colombia suveaeg", "HNEG": "Ida-Gröönimaa standardaeg", "HEOG": "Lääne-Gröönimaa suveaeg", "CHAST": "Chathami standardaeg", "CDT": "Kesk-Ameerika suveaeg", "WESZ": "Lääne-Euroopa suveaeg", "AKST": "Alaska standardaeg", "WIT": "Ida-Indoneesia aeg", "ChST": "Tšamorro standardaeg", "HNPMX": "Mehhiko Vaikse ookeani standardaeg", "SAST": "Lõuna-Aafrika standardaeg", "GFT": "Prantsuse Guajaana aeg", "EST": "Idaranniku standardaeg", "ACWST": "Austraalia Kesk-Lääne standardaeg", "ACWDT": "Austraalia Kesk-Lääne suveaeg", "HKT": "Hongkongi standardaeg", "LHST": "Lord Howe’i standardaeg", "PDT": "Vaikse ookeani suveaeg", "WEZ": "Lääne-Euroopa standardaeg", "UYST": "Uruguay suveaeg", "VET": "Venezuela aeg", "HADT": "Hawaii-Aleuudi suveaeg"},
	}
}

// Locale returns the current translators string locale
func (et *et_EE) Locale() string {
	return et.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'et_EE'
func (et *et_EE) PluralsCardinal() []locales.PluralRule {
	return et.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'et_EE'
func (et *et_EE) PluralsOrdinal() []locales.PluralRule {
	return et.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'et_EE'
func (et *et_EE) PluralsRange() []locales.PluralRule {
	return et.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'et_EE'
func (et *et_EE) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)

	if i == 1 && v == 0 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'et_EE'
func (et *et_EE) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'et_EE'
func (et *et_EE) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (et *et_EE) MonthAbbreviated(month time.Month) string {
	return et.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (et *et_EE) MonthsAbbreviated() []string {
	return et.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (et *et_EE) MonthNarrow(month time.Month) string {
	return et.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (et *et_EE) MonthsNarrow() []string {
	return et.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (et *et_EE) MonthWide(month time.Month) string {
	return et.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (et *et_EE) MonthsWide() []string {
	return et.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (et *et_EE) WeekdayAbbreviated(weekday time.Weekday) string {
	return et.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (et *et_EE) WeekdaysAbbreviated() []string {
	return et.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (et *et_EE) WeekdayNarrow(weekday time.Weekday) string {
	return et.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (et *et_EE) WeekdaysNarrow() []string {
	return et.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (et *et_EE) WeekdayShort(weekday time.Weekday) string {
	return et.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (et *et_EE) WeekdaysShort() []string {
	return et.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (et *et_EE) WeekdayWide(weekday time.Weekday) string {
	return et.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (et *et_EE) WeekdaysWide() []string {
	return et.daysWide
}

// Decimal returns the decimal point of number
func (et *et_EE) Decimal() string {
	return et.decimal
}

// Group returns the group of number
func (et *et_EE) Group() string {
	return et.group
}

// Group returns the minus sign of number
func (et *et_EE) Minus() string {
	return et.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'et_EE' and handles both Whole and Real numbers based on 'v'
func (et *et_EE) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, et.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(et.group) - 1; j >= 0; j-- {
					b = append(b, et.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(et.minus) - 1; j >= 0; j-- {
			b = append(b, et.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'et_EE' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (et *et_EE) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 5
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, et.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(et.minus) - 1; j >= 0; j-- {
			b = append(b, et.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, et.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'et_EE'
func (et *et_EE) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := et.currencies[currency]
	l := len(s) + len(symbol) + 6 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, et.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(et.group) - 1; j >= 0; j-- {
					b = append(b, et.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(et.minus) - 1; j >= 0; j-- {
			b = append(b, et.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, et.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, et.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'et_EE'
// in accounting notation.
func (et *et_EE) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := et.currencies[currency]
	l := len(s) + len(symbol) + 8 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, et.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(et.group) - 1; j >= 0; j-- {
					b = append(b, et.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, et.currencyNegativePrefix[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, et.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, et.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, et.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'et_EE'
func (et *et_EE) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'et_EE'
func (et *et_EE) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, et.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'et_EE'
func (et *et_EE) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, et.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'et_EE'
func (et *et_EE) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, et.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, et.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'et_EE'
func (et *et_EE) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, et.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'et_EE'
func (et *et_EE) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, et.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, et.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'et_EE'
func (et *et_EE) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, et.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, et.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'et_EE'
func (et *et_EE) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, et.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, et.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := et.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
