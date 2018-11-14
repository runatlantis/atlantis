package to_TO

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type to_TO struct {
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

// New returns a new instance of translator for the 'to_TO' locale
func New() locales.Translator {
	return &to_TO{
		locale:                 "to_TO",
		pluralsCardinal:        []locales.PluralRule{6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ".",
		group:                  ",",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		percentSuffix:          " ",
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "Sān", "Fēp", "Maʻa", "ʻEpe", "Mē", "Sun", "Siu", "ʻAok", "Sep", "ʻOka", "Nōv", "Tīs"},
		monthsNarrow:           []string{"", "S", "F", "M", "E", "M", "S", "S", "A", "S", "O", "N", "T"},
		monthsWide:             []string{"", "Sānuali", "Fēpueli", "Maʻasi", "ʻEpeleli", "Mē", "Sune", "Siulai", "ʻAokosi", "Sepitema", "ʻOkatopa", "Nōvema", "Tīsema"},
		daysAbbreviated:        []string{"Sāp", "Mōn", "Tūs", "Pul", "Tuʻa", "Fal", "Tok"},
		daysNarrow:             []string{"S", "M", "T", "P", "T", "F", "T"},
		daysShort:              []string{"Sāp", "Mōn", "Tūs", "Pul", "Tuʻa", "Fal", "Tok"},
		daysWide:               []string{"Sāpate", "Mōnite", "Tūsite", "Pulelulu", "Tuʻapulelulu", "Falaite", "Tokonaki"},
		periodsAbbreviated:     []string{"AM", "PM"},
		periodsNarrow:          []string{"AM", "PM"},
		periodsWide:            []string{"hengihengi", "efiafi"},
		erasAbbreviated:        []string{"KM", "TS"},
		erasNarrow:             []string{"KM", "TS"},
		erasWide:               []string{"ki muʻa", "taʻu ʻo Sīsū"},
		timezones:              map[string]string{"HNNOMX": "houa fakamekisikou-tokelauhihifo taimi totonu", "HEOG": "houa fakafonuamata-hihifo taimi liliu", "VET": "houa fakavenesuela", "OESZ": "houa fakaʻeulope-hahake taimi liliu", "GMT": "houa fakakiliniuisi mālie", "PDT": "houa fakaʻamelika-tokelau pasifika taimi liliu", "AEDT": "houa fakaʻaositelēlia-hahake taimi liliu", "HNEG": "houa fakafonuamata-hahake taimi totonu", "WART": "houa fakaʻasenitina-hihifo taimi totonu", "WITA": "houa fakaʻinitonisia-loto", "HENOMX": "houa fakamekisikou-tokelauhihifo taimi liliu", "TMT": "houa fakatūkimenisitani taimi totonu", "ACDT": "houa fakaʻaositelēlia-loto taimi liliu", "HNOG": "houa fakafonuamata-hihifo taimi totonu", "NZST": "houa fakanuʻusila taimi totonu", "AKST": "houa fakaʻalasika taimi totonu", "HNPM": "houa fakasā-piea-mo-mikeloni taimi totonu", "COST": "houa fakakolomipia taimi liliu", "UYT": "houa fakaʻulukuai taimi totonu", "UYST": "houa fakaʻulukuai taimi liliu", "WESZ": "houa fakaʻeulope-hihifo taimi liliu", "ACWST": "houa fakaʻaositelēlia-loto-hihifo taimi totonu", "ART": "houa fakaʻasenitina taimi totonu", "COT": "houa fakakolomipia taimi totonu", "WARST": "houa fakaʻasenitina-hihifo taimi liliu", "EAT": "houa fakaʻafelika-hahake", "HECU": "houa fakakiupa taimi liliu", "WEZ": "houa fakaʻeulope-hihifo taimi totonu", "ECT": "houa fakaʻekuetoa", "HEPM": "houa fakasā-piea-mo-mikeloni taimi liliu", "HAST": "houa fakahauaʻi taimi totonu", "ARST": "houa fakaʻasenitina taimi liliu", "MYT": "houa fakamaleisia", "GFT": "houa fakakuiana-fakafalanisē", "NZDT": "houa fakanuʻusila taimi liliu", "HEEG": "houa fakafonuamata-hahake taimi liliu", "MEZ": "houa fakaʻeulope-loto taimi totonu", "HNT": "houa fakafonuaʻilofoʻou taimi totonu", "CAT": "houa fakaʻafelika-loto", "MST": "houa fakaʻamelika-tokelau moʻunga taimi totonu", "∅∅∅": "houa fakaʻakelī taimi liliu", "SAST": "houa fakaʻafelika-tonga", "ChST": "houa fakakamolo", "AKDT": "houa fakaʻalasika taimi liliu", "HKST": "houa fakahongi-kongi taimi liliu", "WIT": "houa fakaʻinitonisia-hahake", "HNCU": "houa fakakiupa taimi totonu", "AEST": "houa fakaʻaositelēlia-hahake taimi totonu", "WAT": "houa fakaʻafelika-hihifo taimi totonu", "EDT": "houa fakaʻamelika-tokelau hahake taimi liliu", "CST": "houa fakaʻamelika-tokelau loto taimi totonu", "CDT": "houa fakaʻamelika-tokelau loto taimi liliu", "AST": "houa fakaʻamelika-tokelau ʻatalanitiki taimi totonu", "EST": "houa fakaʻamelika-tokelau hahake taimi totonu", "JDT": "houa fakasiapani taimi liliu", "LHST": "houa fakamotuʻeikihoue taimi totonu", "LHDT": "houa fakamotuʻeikihoue taimi liliu", "CLT": "houa fakasili taimi totonu", "OEZ": "houa fakaʻeulope-hahake taimi totonu", "WIB": "houa fakaʻinitonisia-hihifo", "JST": "houa fakasiapani taimi totonu", "ACWDT": "houa fakaʻaositelēlia-loto-hihifo taimi liliu", "HKT": "houa fakahongi-kongi taimi totonu", "HAT": "houa fakafonuaʻilofoʻou taimi liliu", "SRT": "houa fakasuliname", "GYT": "houa fakakuiana", "HNPMX": "houa fakamekisikou-pasifika taimi totonu", "BT": "houa fakapūtani", "BOT": "houa fakapolīvia", "HEPMX": "houa fakamekisikou-pasifika taimi liliu", "SGT": "houa fakasingapoa", "HADT": "houa fakahauaʻi taimi liliu", "CHAST": "houa fakasatihami taimi totonu", "CHADT": "houa fakasatihami taimi liliu", "ADT": "houa fakaʻamelika-tokelau ʻatalanitiki taimi liliu", "MDT": "houa fakaʻamelika-tokelau moʻunga taimi liliu", "MESZ": "houa fakaʻeulope-loto taimi liliu", "IST": "houa fakaʻinitia", "CLST": "houa fakasili taimi liliu", "TMST": "houa fakatūkimenisitani taimi liliu", "PST": "houa fakaʻamelika-tokelau pasifika taimi totonu", "AWST": "houa fakaʻaositelēlia-hihifo taimi totonu", "WAST": "houa fakaʻafelika-hihifo taimi liliu", "ACST": "houa fakaʻaositelēlia-loto taimi totonu", "AWDT": "houa fakaʻaositelēlia-hihifo taimi liliu"},
	}
}

// Locale returns the current translators string locale
func (to *to_TO) Locale() string {
	return to.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'to_TO'
func (to *to_TO) PluralsCardinal() []locales.PluralRule {
	return to.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'to_TO'
func (to *to_TO) PluralsOrdinal() []locales.PluralRule {
	return to.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'to_TO'
func (to *to_TO) PluralsRange() []locales.PluralRule {
	return to.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'to_TO'
func (to *to_TO) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'to_TO'
func (to *to_TO) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'to_TO'
func (to *to_TO) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (to *to_TO) MonthAbbreviated(month time.Month) string {
	return to.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (to *to_TO) MonthsAbbreviated() []string {
	return to.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (to *to_TO) MonthNarrow(month time.Month) string {
	return to.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (to *to_TO) MonthsNarrow() []string {
	return to.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (to *to_TO) MonthWide(month time.Month) string {
	return to.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (to *to_TO) MonthsWide() []string {
	return to.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (to *to_TO) WeekdayAbbreviated(weekday time.Weekday) string {
	return to.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (to *to_TO) WeekdaysAbbreviated() []string {
	return to.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (to *to_TO) WeekdayNarrow(weekday time.Weekday) string {
	return to.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (to *to_TO) WeekdaysNarrow() []string {
	return to.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (to *to_TO) WeekdayShort(weekday time.Weekday) string {
	return to.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (to *to_TO) WeekdaysShort() []string {
	return to.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (to *to_TO) WeekdayWide(weekday time.Weekday) string {
	return to.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (to *to_TO) WeekdaysWide() []string {
	return to.daysWide
}

// Decimal returns the decimal point of number
func (to *to_TO) Decimal() string {
	return to.decimal
}

// Group returns the group of number
func (to *to_TO) Group() string {
	return to.group
}

// Group returns the minus sign of number
func (to *to_TO) Minus() string {
	return to.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'to_TO' and handles both Whole and Real numbers based on 'v'
func (to *to_TO) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, to.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, to.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, to.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'to_TO' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (to *to_TO) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 5
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, to.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, to.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, to.percentSuffix...)

	b = append(b, to.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'to_TO'
func (to *to_TO) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := to.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, to.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, to.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, to.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, to.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, to.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'to_TO'
// in accounting notation.
func (to *to_TO) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := to.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, to.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, to.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, to.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, to.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, to.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, to.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'to_TO'
func (to *to_TO) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'to_TO'
func (to *to_TO) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, to.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'to_TO'
func (to *to_TO) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, to.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'to_TO'
func (to *to_TO) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, to.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, to.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'to_TO'
func (to *to_TO) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, to.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, to.periodsAbbreviated[0]...)
	} else {
		b = append(b, to.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'to_TO'
func (to *to_TO) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, to.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, to.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, to.periodsAbbreviated[0]...)
	} else {
		b = append(b, to.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'to_TO'
func (to *to_TO) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, to.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, to.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, to.periodsAbbreviated[0]...)
	} else {
		b = append(b, to.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'to_TO'
func (to *to_TO) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, to.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, to.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, to.periodsAbbreviated[0]...)
	} else {
		b = append(b, to.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := to.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
