package uz_Cyrl_UZ

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type uz_Cyrl_UZ struct {
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

// New returns a new instance of translator for the 'uz_Cyrl_UZ' locale
func New() locales.Translator {
	return &uz_Cyrl_UZ{
		locale:                 "uz_Cyrl_UZ",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         []locales.PluralRule{6},
		pluralsRange:           []locales.PluralRule{2, 6},
		decimal:                ",",
		group:                  " ",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "yan", "fev", "mar", "apr", "may", "iyn", "iyl", "avg", "sen", "okt", "noy", "dek"},
		monthsNarrow:           []string{"", "Y", "F", "M", "A", "M", "I", "I", "A", "S", "O", "N", "D"},
		monthsWide:             []string{"", "yanvar", "fevral", "mart", "aprel", "may", "iyun", "iyul", "avgust", "sentabr", "oktabr", "noyabr", "dekabr"},
		daysAbbreviated:        []string{"Yak", "Dush", "Sesh", "Chor", "Pay", "Jum", "Shan"},
		daysNarrow:             []string{"Y", "D", "S", "C", "P", "J", "S"},
		daysShort:              []string{"Ya", "Du", "Se", "Ch", "Pa", "Ju", "Sh"},
		daysWide:               []string{"yakshanba", "dushanba", "seshanba", "chorshanba", "payshanba", "juma", "shanba"},
		periodsAbbreviated:     []string{"TO", "TK"},
		periodsNarrow:          []string{"TO", "TK"},
		periodsWide:            []string{"TO", "TK"},
		erasAbbreviated:        []string{"m.a.", "milodiy"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"miloddan avvalgi", "milodiy"},
		timezones:              map[string]string{"BOT": "Boliviya vaqti", "GYT": "Gayana vaqti", "AWDT": "G‘arbiy Avstraliya yozgi vaqti", "WESZ": "G‘arbiy Yevropa yozgi vaqti", "NZDT": "Yangi Zelandiya yozgi vaqti", "HEPM": "Sen-Pyer va Mikelon yozgi vaqti", "AEDT": "Sharqiy Avstraliya yozgi vaqti", "GFT": "Fransuz Gvianasi vaqti", "BT": "Butan vaqti", "WARST": "Gʻarbiy Argentina yozgi vaqti", "WIB": "Gʻarbiy Indoneziya vaqti", "JDT": "Yaponiya yozgi vaqti", "AKST": "Alyaska standart vaqti", "HECU": "Kuba yozgi vaqti", "PST": "Tinch okeani standart vaqti", "HEPMX": "Meksika Tinch okeani yozgi vaqti", "AST": "Atlantika standart vaqti", "WITA": "Markaziy Indoneziya vaqti", "WAT": "Gʻarbiy Afrika standart vaqti", "WEZ": "G‘arbiy Yevropa standart vaqti", "HNOG": "G‘arbiy Grenlandiya standart vaqti", "MEZ": "Markaziy Yevropa standart vaqti", "HNPM": "Sen-Pyer va Mikelon standart vaqti", "HNNOMX": "Shimoli-g‘arbiy Meksika standart vaqti", "CLST": "Chili yozgi vaqti", "COT": "Kolumbiya standart vaqti", "UYT": "Urugvay standart vaqti", "ECT": "Ekvador vaqti", "IST": "Hindiston standart vaqti", "OEZ": "Sharqiy Yevropa standart vaqti", "NZST": "Yangi Zelandiya standart vaqti", "LHDT": "Lord-Xau yozgi vaqti", "HENOMX": "Shimoli-g‘arbiy Meksika yozgi vaqti", "HEEG": "Sharqiy Grenlandiya yozgi vaqti", "HKST": "Gonkong yozgi vaqti", "SRT": "Surinam vaqti", "OESZ": "Sharqiy Yevropa yozgi vaqti", "CHAST": "Chatem standart vaqti", "WAST": "Gʻarbiy Afrika yozgi vaqti", "AKDT": "Alyaska yozgi vaqti", "ChST": "Chamorro standart vaqti", "HNCU": "Kuba standart vaqti", "CAT": "Markaziy Afrika vaqti", "∅∅∅": "Amazonka yozgi vaqti", "CHADT": "Chatem yozgi vaqti", "SGT": "Singapur vaqti", "HAT": "Nyufaundlend yozgi vaqti", "GMT": "Grinvich o‘rtacha vaqti", "HNEG": "Sharqiy Grenlandiya standart vaqti", "EAT": "Sharqiy Afrika vaqti", "HAST": "Gavayi-aleut standart vaqti", "HNPMX": "Meksika Tinch okeani standart vaqti", "MST": "Tog‘ standart vaqti (AQSH)", "TMT": "Turkmaniston standart vaqti", "ACWDT": "Markaziy Avstraliya g‘arbiy yozgi vaqti", "WART": "Gʻarbiy Argentina standart vaqti", "TMST": "Turkmaniston yozgi vaqti", "ARST": "Argentina yozgi vaqti", "CST": "Markaziy Amerika standart vaqti", "JST": "Yaponiya standart vaqti", "ACST": "Markaziy Avstraliya standart vaqti", "AWST": "G‘arbiy Avstraliya standart vaqti", "HKT": "Gonkong standart vaqti", "CLT": "Chili standart vaqti", "WIT": "Sharqiy Indoneziya vaqti", "HADT": "Gavayi-aleut yozgi vaqti", "UYST": "Urugvay yozgi vaqti", "CDT": "Markaziy Amerika yozgi vaqti", "PDT": "Tinch okeani yozgi vaqti", "VET": "Venesuela vaqti", "AEST": "Sharqiy Avstraliya standart vaqti", "EST": "Sharqiy Amerika standart vaqti", "EDT": "Sharqiy Amerika yozgi vaqti", "MESZ": "Markaziy Yevropa yozgi vaqti", "ACDT": "Markaziy Avstraliya yozgi vaqti", "ACWST": "Markaziy Avstraliya g‘arbiy standart vaqti", "HNT": "Nyufaundlend standart vaqti", "ART": "Argentina standart vaqti", "COST": "Kolumbiya yozgi vaqti", "MDT": "Tog‘ yozgi vaqti (AQSH)", "MYT": "Malayziya vaqti", "ADT": "Atlantika yozgi vaqti", "SAST": "Janubiy Afrika standart vaqti", "HEOG": "G‘arbiy Grenlandiya yozgi vaqti", "LHST": "Lord-Xau standart vaqti"},
	}
}

// Locale returns the current translators string locale
func (uz *uz_Cyrl_UZ) Locale() string {
	return uz.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'uz_Cyrl_UZ'
func (uz *uz_Cyrl_UZ) PluralsCardinal() []locales.PluralRule {
	return uz.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'uz_Cyrl_UZ'
func (uz *uz_Cyrl_UZ) PluralsOrdinal() []locales.PluralRule {
	return uz.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'uz_Cyrl_UZ'
func (uz *uz_Cyrl_UZ) PluralsRange() []locales.PluralRule {
	return uz.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'uz_Cyrl_UZ'
func (uz *uz_Cyrl_UZ) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'uz_Cyrl_UZ'
func (uz *uz_Cyrl_UZ) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'uz_Cyrl_UZ'
func (uz *uz_Cyrl_UZ) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := uz.CardinalPluralRule(num1, v1)
	end := uz.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (uz *uz_Cyrl_UZ) MonthAbbreviated(month time.Month) string {
	return uz.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (uz *uz_Cyrl_UZ) MonthsAbbreviated() []string {
	return uz.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (uz *uz_Cyrl_UZ) MonthNarrow(month time.Month) string {
	return uz.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (uz *uz_Cyrl_UZ) MonthsNarrow() []string {
	return uz.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (uz *uz_Cyrl_UZ) MonthWide(month time.Month) string {
	return uz.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (uz *uz_Cyrl_UZ) MonthsWide() []string {
	return uz.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (uz *uz_Cyrl_UZ) WeekdayAbbreviated(weekday time.Weekday) string {
	return uz.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (uz *uz_Cyrl_UZ) WeekdaysAbbreviated() []string {
	return uz.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (uz *uz_Cyrl_UZ) WeekdayNarrow(weekday time.Weekday) string {
	return uz.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (uz *uz_Cyrl_UZ) WeekdaysNarrow() []string {
	return uz.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (uz *uz_Cyrl_UZ) WeekdayShort(weekday time.Weekday) string {
	return uz.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (uz *uz_Cyrl_UZ) WeekdaysShort() []string {
	return uz.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (uz *uz_Cyrl_UZ) WeekdayWide(weekday time.Weekday) string {
	return uz.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (uz *uz_Cyrl_UZ) WeekdaysWide() []string {
	return uz.daysWide
}

// Decimal returns the decimal point of number
func (uz *uz_Cyrl_UZ) Decimal() string {
	return uz.decimal
}

// Group returns the group of number
func (uz *uz_Cyrl_UZ) Group() string {
	return uz.group
}

// Group returns the minus sign of number
func (uz *uz_Cyrl_UZ) Minus() string {
	return uz.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'uz_Cyrl_UZ' and handles both Whole and Real numbers based on 'v'
func (uz *uz_Cyrl_UZ) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, uz.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(uz.group) - 1; j >= 0; j-- {
					b = append(b, uz.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, uz.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'uz_Cyrl_UZ' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (uz *uz_Cyrl_UZ) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, uz.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, uz.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, uz.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'uz_Cyrl_UZ'
func (uz *uz_Cyrl_UZ) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := uz.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, uz.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(uz.group) - 1; j >= 0; j-- {
					b = append(b, uz.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, uz.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, uz.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, uz.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'uz_Cyrl_UZ'
// in accounting notation.
func (uz *uz_Cyrl_UZ) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := uz.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, uz.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(uz.group) - 1; j >= 0; j-- {
					b = append(b, uz.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, uz.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, uz.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, uz.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, uz.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'uz_Cyrl_UZ'
func (uz *uz_Cyrl_UZ) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2f}...)

	if t.Month() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Month()), 10)

	b = append(b, []byte{0x2f}...)

	if t.Year() > 9 {
		b = append(b, strconv.Itoa(t.Year())[2:]...)
	} else {
		b = append(b, strconv.Itoa(t.Year())[1:]...)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'uz_Cyrl_UZ'
func (uz *uz_Cyrl_UZ) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2d}...)
	b = append(b, uz.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'uz_Cyrl_UZ'
func (uz *uz_Cyrl_UZ) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2d}...)
	b = append(b, uz.monthsWide[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'uz_Cyrl_UZ'
func (uz *uz_Cyrl_UZ) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, uz.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2d}...)
	b = append(b, uz.monthsWide[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'uz_Cyrl_UZ'
func (uz *uz_Cyrl_UZ) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, uz.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'uz_Cyrl_UZ'
func (uz *uz_Cyrl_UZ) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, uz.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, uz.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'uz_Cyrl_UZ'
func (uz *uz_Cyrl_UZ) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, uz.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, uz.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20, 0x28}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	b = append(b, []byte{0x29}...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'uz_Cyrl_UZ'
func (uz *uz_Cyrl_UZ) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, uz.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, uz.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20, 0x28}...)

	tz, _ := t.Zone()

	if btz, ok := uz.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	b = append(b, []byte{0x29}...)

	return string(b)
}
