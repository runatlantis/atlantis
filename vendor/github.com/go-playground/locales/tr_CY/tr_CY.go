package tr_CY

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type tr_CY struct {
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

// New returns a new instance of translator for the 'tr_CY' locale
func New() locales.Translator {
	return &tr_CY{
		locale:                 "tr_CY",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         []locales.PluralRule{6},
		pluralsRange:           []locales.PluralRule{2, 6},
		decimal:                ",",
		group:                  ".",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: ")",
		monthsAbbreviated:      []string{"", "Oca", "Şub", "Mar", "Nis", "May", "Haz", "Tem", "Ağu", "Eyl", "Eki", "Kas", "Ara"},
		monthsNarrow:           []string{"", "O", "Ş", "M", "N", "M", "H", "T", "A", "E", "E", "K", "A"},
		monthsWide:             []string{"", "Ocak", "Şubat", "Mart", "Nisan", "Mayıs", "Haziran", "Temmuz", "Ağustos", "Eylül", "Ekim", "Kasım", "Aralık"},
		daysAbbreviated:        []string{"Paz", "Pzt", "Sal", "Çar", "Per", "Cum", "Cmt"},
		daysNarrow:             []string{"P", "P", "S", "Ç", "P", "C", "C"},
		daysShort:              []string{"Pa", "Pt", "Sa", "Ça", "Pe", "Cu", "Ct"},
		daysWide:               []string{"Pazar", "Pazartesi", "Salı", "Çarşamba", "Perşembe", "Cuma", "Cumartesi"},
		periodsAbbreviated:     []string{"ÖÖ", "ÖS"},
		periodsNarrow:          []string{"öö", "ös"},
		periodsWide:            []string{"ÖÖ", "ÖS"},
		erasAbbreviated:        []string{"MÖ", "MS"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"Milattan Önce", "Milattan Sonra"},
		timezones:              map[string]string{"CHADT": "Chatham Yaz Saati", "WESZ": "Batı Avrupa Yaz Saati", "AKDT": "Alaska Yaz Saati", "MEZ": "Orta Avrupa Standart Saati", "TMT": "Türkmenistan Standart Saati", "COST": "Kolombiya Yaz Saati", "GFT": "Fransız Guyanası Saati", "HNPM": "Saint Pierre ve Miquelon Standart Saati", "AST": "Atlantik Standart Saati", "MESZ": "Orta Avrupa Yaz Saati", "CHAST": "Chatham Standart Saati", "JDT": "Japonya Yaz Saati", "ACDT": "Orta Avustralya Yaz Saati", "MST": "Makao Standart Saati", "HADT": "Hawaii-Aleut Yaz Saati", "IST": "Hindistan Standart Saati", "WART": "Batı Arjantin Standart Saati", "VET": "Venezuela Saati", "OEZ": "Doğu Avrupa Standart Saati", "COT": "Kolombiya Standart Saati", "CLT": "Şili Standart Saati", "∅∅∅": "Amazon Yaz Saati", "AWDT": "Batı Avustralya Yaz Saati", "HNPMX": "Meksika Pasifik Kıyısı Standart Saati", "ADT": "Atlantik Yaz Saati", "HNOG": "Batı Grönland Standart Saati", "HENOMX": "Kuzeybatı Meksika Yaz Saati", "MDT": "Makao Yaz Saati", "CDT": "Kuzey Amerika Merkezi Yaz Saati", "BOT": "Bolivya Saati", "ECT": "Ekvador Saati", "EDT": "Kuzey Amerika Doğu Yaz Saati", "HNEG": "Doğu Grönland Standart Saati", "HKT": "Hong Kong Standart Saati", "SRT": "Surinam Saati", "CLST": "Şili Yaz Saati", "LHDT": "Lord Howe Yaz Saati", "AKST": "Alaska Standart Saati", "SGT": "Singapur Standart Saati", "ACST": "Orta Avustralya Standart Saati", "HEEG": "Doğu Grönland Yaz Saati", "HNT": "Newfoundland Standart Saati", "OESZ": "Doğu Avrupa Yaz Saati", "NZST": "Yeni Zelanda Standart Saati", "CST": "Kuzey Amerika Merkezi Standart Saati", "AWST": "Batı Avustralya Standart Saati", "WAST": "Batı Afrika Yaz Saati", "WEZ": "Batı Avrupa Standart Saati", "MYT": "Malezya Saati", "WITA": "Orta Endonezya Saati", "TMST": "Türkmenistan Yaz Saati", "UYST": "Uruguay Yaz Saati", "BT": "Butan Saati", "LHST": "Lord Howe Standart Saati", "CAT": "Orta Afrika Saati", "SAST": "Güney Afrika Standart Saati", "UYT": "Uruguay Standart Saati", "ChST": "Chamorro Saati", "PDT": "Kuzey Amerika Pasifik Yaz Saati", "AEST": "Doğu Avustralya Standart Saati", "EST": "Kuzey Amerika Doğu Standart Saati", "WIT": "Doğu Endonezya Saati", "GYT": "Guyana Saati", "JST": "Japonya Standart Saati", "HKST": "Hong Kong Yaz Saati", "AEDT": "Doğu Avustralya Yaz Saati", "WIB": "Batı Endonezya Saati", "GMT": "Greenwich Ortalama Saati", "HAT": "Newfoundland Yaz Saati", "HNNOMX": "Kuzeybatı Meksika Standart Saati", "EAT": "Doğu Afrika Saati", "PST": "Kuzey Amerika Pasifik Standart Saati", "ACWST": "İç Batı Avustralya Standart Saati", "HEOG": "Batı Grönland Yaz Saati", "HEPM": "Saint Pierre ve Miquelon Yaz Saati", "ARST": "Arjantin Yaz Saati", "HECU": "Küba Yaz Saati", "HNCU": "Küba Standart Saati", "HEPMX": "Meksika Pasifik Kıyısı Yaz Saati", "WAT": "Batı Afrika Standart Saati", "NZDT": "Yeni Zelanda Yaz Saati", "ACWDT": "İç Batı Avustralya Yaz Saati", "WARST": "Batı Arjantin Yaz Saati", "HAST": "Hawaii-Aleut Standart Saati", "ART": "Arjantin Standart Saati"},
	}
}

// Locale returns the current translators string locale
func (tr *tr_CY) Locale() string {
	return tr.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'tr_CY'
func (tr *tr_CY) PluralsCardinal() []locales.PluralRule {
	return tr.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'tr_CY'
func (tr *tr_CY) PluralsOrdinal() []locales.PluralRule {
	return tr.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'tr_CY'
func (tr *tr_CY) PluralsRange() []locales.PluralRule {
	return tr.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'tr_CY'
func (tr *tr_CY) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'tr_CY'
func (tr *tr_CY) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'tr_CY'
func (tr *tr_CY) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := tr.CardinalPluralRule(num1, v1)
	end := tr.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (tr *tr_CY) MonthAbbreviated(month time.Month) string {
	return tr.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (tr *tr_CY) MonthsAbbreviated() []string {
	return tr.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (tr *tr_CY) MonthNarrow(month time.Month) string {
	return tr.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (tr *tr_CY) MonthsNarrow() []string {
	return tr.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (tr *tr_CY) MonthWide(month time.Month) string {
	return tr.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (tr *tr_CY) MonthsWide() []string {
	return tr.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (tr *tr_CY) WeekdayAbbreviated(weekday time.Weekday) string {
	return tr.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (tr *tr_CY) WeekdaysAbbreviated() []string {
	return tr.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (tr *tr_CY) WeekdayNarrow(weekday time.Weekday) string {
	return tr.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (tr *tr_CY) WeekdaysNarrow() []string {
	return tr.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (tr *tr_CY) WeekdayShort(weekday time.Weekday) string {
	return tr.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (tr *tr_CY) WeekdaysShort() []string {
	return tr.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (tr *tr_CY) WeekdayWide(weekday time.Weekday) string {
	return tr.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (tr *tr_CY) WeekdaysWide() []string {
	return tr.daysWide
}

// Decimal returns the decimal point of number
func (tr *tr_CY) Decimal() string {
	return tr.decimal
}

// Group returns the group of number
func (tr *tr_CY) Group() string {
	return tr.group
}

// Group returns the minus sign of number
func (tr *tr_CY) Minus() string {
	return tr.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'tr_CY' and handles both Whole and Real numbers based on 'v'
func (tr *tr_CY) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, tr.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, tr.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, tr.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'tr_CY' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (tr *tr_CY) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, tr.decimal[0])
			inWhole = true

			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, tr.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, tr.minus[0])
	}

	b = append(b, tr.percent[0])

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'tr_CY'
func (tr *tr_CY) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := tr.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, tr.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, tr.group[0])
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
		b = append(b, tr.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, tr.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'tr_CY'
// in accounting notation.
func (tr *tr_CY) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := tr.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, tr.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, tr.group[0])
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

		b = append(b, tr.currencyNegativePrefix[0])

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
			b = append(b, tr.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, tr.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'tr_CY'
func (tr *tr_CY) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

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

// FmtDateMedium returns the medium date representation of 't' for 'tr_CY'
func (tr *tr_CY) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, tr.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'tr_CY'
func (tr *tr_CY) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, tr.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'tr_CY'
func (tr *tr_CY) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, tr.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, tr.daysWide[t.Weekday()]...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'tr_CY'
func (tr *tr_CY) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, tr.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, tr.periodsAbbreviated[0]...)
	} else {
		b = append(b, tr.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'tr_CY'
func (tr *tr_CY) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, tr.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, tr.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, tr.periodsAbbreviated[0]...)
	} else {
		b = append(b, tr.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'tr_CY'
func (tr *tr_CY) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, tr.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, tr.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, tr.periodsAbbreviated[0]...)
	} else {
		b = append(b, tr.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'tr_CY'
func (tr *tr_CY) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, tr.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, tr.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, tr.periodsAbbreviated[0]...)
	} else {
		b = append(b, tr.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := tr.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
