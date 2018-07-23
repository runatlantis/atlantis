package ks_IN

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type ks_IN struct {
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
	currencyPositivePrefix string
	currencyNegativePrefix string
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

// New returns a new instance of translator for the 'ks_IN' locale
func New() locales.Translator {
	return &ks_IN{
		locale:                 "ks_IN",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ".",
		group:                  ",",
		minus:                  "‎-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositivePrefix: " ",
		currencyNegativePrefix: " ",
		monthsNarrow:           []string{"", "ج", "ف", "م", "ا", "م", "ج", "ج", "ا", "س", "س", "ا", "ن"},
		monthsWide:             []string{"", "جنؤری", "فرؤری", "مارٕچ", "اپریل", "میٔ", "جوٗن", "جوٗلایی", "اگست", "ستمبر", "اکتوٗبر", "نومبر", "دسمبر"},
		daysAbbreviated:        []string{"آتھوار", "ژٔنٛدٕروار", "بوٚموار", "بودوار", "برٛٮ۪سوار", "جُمہ", "بٹوار"},
		daysNarrow:             []string{"ا", "ژ", "ب", "ب", "ب", "ج", "ب"},
		daysWide:               []string{"اَتھوار", "ژٔنٛدرٕروار", "بوٚموار", "بودوار", "برٛٮ۪سوار", "جُمہ", "بٹوار"},
		erasAbbreviated:        []string{"بی سی", "اے ڈی"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"قبٕل مسیٖح", "عیٖسوی سنہٕ"},
		timezones:              map[string]string{"NZDT": "نِوزِلینڑ ڑےلایٔٹ ٹایِم", "HNEG": "مشرِقی گریٖن لینڑُک سٹینڑاڑ ٹایِم", "∅∅∅": "اٮ۪زورٕس سَمَر ٹ", "HAST": "حَواے اٮ۪لیوٗٹِیَن سٹینڑاڑ ٹایِم", "UYT": "یوٗرٮ۪گوَے سٹینڑاڑ ٹایِم", "WIB": "مغرِبی اِنڑونیشِیا ٹایِم", "GFT": "فرٛٮ۪نٛچ گیوٗٮ۪نا ٹایِم", "HEPM": "سینٛٹ پَیری مِقیوٗلَن ڑےلایِٔٹ ٹایِم", "COST": "کولومبِیا سَمَر ٹایِم", "ADT": "اٮ۪ٹلانٹِک ڈےلایِٔٹ ٹایِم", "AEST": "آسٹریلِیَن مشرقی سٹینڑاڑ ٹایِم", "HEOG": "مغرِبی گریٖن لینڑُک سَمَر ٹایِم", "CHAST": "کٮ۪تھَم سٹینڑاڑ ٹایِم", "JDT": "جاپٲنۍ ڑےلایِٔٹ ٹایِم", "HEEG": "مشرِقی گریٖن لینڑُک سَمَر ٹایِم", "OEZ": "مشرقی یوٗرپی سٹینڑاڑ ٹایِم", "OESZ": "مشرقی یوٗرپی سَمَر ٹایِم", "GMT": "گرٛیٖن وِچ میٖن ٹایِم", "CDT": "مرکزی ڈےلایِٔٹ ٹایِم", "AST": "اٮ۪ٹلانٹِک سٹینڑاڑ ٹایِم", "HNT": "نیوٗ فاونڑلینڑ سٹینڑاڑ ٹایِم", "TMST": "تُرکمٮ۪نِستان سَمَر ٹایِم", "HECU": "کیوٗبا ڈےلایِٔٹ ٹایِم", "SAST": "جنوٗبی افریقا ٹایِم", "ACWDT": "آسٹریلِیَن مرکزی مغربی ڈےلایِٔٹ ٹایِم", "ACWST": "آسٹریلِیَن مرکزی مغربی سٹینڑاڑ ٹایِم", "HKT": "حانگ کانٛگ سٹینڑاڑ ٹایِم", "WITA": "مرکزی اِنڑونیشِیا ٹایِم", "SRT": "سُرِنام ٹایِم", "CLT": "چِلی سٹینڑاڑ ٹایِم", "JST": "جاپٲنۍ سٹینڑاڑ ٹایِم", "GYT": "گُیَنا ٹایِم", "ChST": "کٮ۪مورو سٹینڑاڑ ٹایِم", "HNOG": "مغرِبی گریٖن لینڑُک سٹینڑاڑ ٹایِم", "WART": "مغربی ارجٮ۪نٹیٖنا سٹینڑاڑ ٹایِم", "WARST": "مغربی ارجٮ۪نٹیٖنا سَمَر ٹایِم", "HNPM": "سینٛٹ پَیری مِقیوٗلَن سٹینڑاڑ ٹایِم", "HNNOMX": "HNNOMX", "COT": "کولومبِیا سٹینڑاڑ ٹایِم", "PST": "پیسِفِک سٹینڑاڑ ٹایِم", "PDT": "پیسِفِک ڈےلایِٔٹ ٹایِم", "AWST": "آسٹریلِیَن مغرِبی سٹینڑاڑ ٹایِم", "MESZ": "مرکزی یوٗرپی سَمَر ٹایِم", "IST": "ہِنٛدوستان", "LHDT": "لعاڑ ڑےلایٔٹ ٹایِم", "TMT": "تُرکمٮ۪نِستان سٹینڑاڑ ٹایِم", "HNCU": "کیوٗبا سٹینڑاڑ ٹایِم", "WEZ": "مغرِبی یوٗرپی سٹینڑاڑ ٹایِم", "NZST": "نِوزِلینڑ سٹینڑاڑ ٹایِم", "MYT": "مَلیشِیا ٹایِم", "SGT": "سِنٛگاپوٗر ٹایِم", "MEZ": "مرکزی یوٗرپی سٹینڑاڑ ٹایِم", "HAT": "نیوٗ فاونڑ لینڑ ڑےلایِٔٹ ٹایِم", "MST": "مَکَعوٗ سٹینڑاڑ ٹایِم", "ARST": "ارجٮ۪نٹیٖنا سَمَر ٹایِم", "CHADT": "چٮ۪تھَم سَمَر ٹایِم", "AEDT": "آسٹریلِیَن مشرقی ڈےلایِٔٹ ٹایِم", "WESZ": "مغرِبی یوٗرِپی سَمَر ٹایِم", "AKST": "اٮ۪لاسکا سٹینڑاڑ ٹایِم", "ACST": "آسٹریلِیَن مرکزی سٹینڑاڑ ٹایِم", "BOT": "بولِوِیا ٹایِم", "LHST": "لعاڑ حووے سٹینڑاڑ ٹایِم", "CAT": "مرکزی افریٖقا ٹایِم", "CST": "مرکزی سٹینڑاڑ ٹایِم", "EST": "مشرقی سٹینڑاڑ ٹایِم", "EDT": "مشرقی ڈےلایِٔٹ ٹایِم", "HENOMX": "HENOMX", "HADT": "حَواے اٮ۪لیوٗٹِیَن سَمَر ٹایِم", "ART": "ارجٮ۪نٹیٖنا سٹینڑاڑ ٹایِم", "HNPMX": "HNPMX", "BT": "بوٗٹان ٹایِم", "MDT": "مَکَعوٗ سَمَر ٹایِم", "EAT": "مشرقی افریٖقا ٹایِم", "HEPMX": "HEPMX", "ACDT": "آسٹریلِیَن مرکزی ڈےلایِٔٹ ٹایِم", "CLST": "چِلی سَمَر ٹایِم", "UYST": "یوٗرٮ۪گوَے سَمَر ٹایِم", "WAT": "مغربی افریٖقا سٹینڑاڑ ٹایِم", "WAST": "مغربی افریٖقا سَمَر ٹایِم", "AKDT": "اٮ۪لاسکا ڈےلایِٔٹ ٹایِم", "HKST": "حانٛگ کانٛگ سَمَر ٹایِم", "VET": "وٮ۪نٮ۪زیوٗلا ٹایِم", "WIT": "مشرِقی اِنڑونیشِیا ٹایِم", "AWDT": "آسٹریلِیَن مغرِبیٖ ڈےلایٔٹ ٹایِم", "ECT": "اِکویڑَر ٹایِم"},
	}
}

// Locale returns the current translators string locale
func (ks *ks_IN) Locale() string {
	return ks.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'ks_IN'
func (ks *ks_IN) PluralsCardinal() []locales.PluralRule {
	return ks.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'ks_IN'
func (ks *ks_IN) PluralsOrdinal() []locales.PluralRule {
	return ks.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'ks_IN'
func (ks *ks_IN) PluralsRange() []locales.PluralRule {
	return ks.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'ks_IN'
func (ks *ks_IN) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'ks_IN'
func (ks *ks_IN) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'ks_IN'
func (ks *ks_IN) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (ks *ks_IN) MonthAbbreviated(month time.Month) string {
	return ks.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (ks *ks_IN) MonthsAbbreviated() []string {
	return nil
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (ks *ks_IN) MonthNarrow(month time.Month) string {
	return ks.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (ks *ks_IN) MonthsNarrow() []string {
	return ks.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (ks *ks_IN) MonthWide(month time.Month) string {
	return ks.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (ks *ks_IN) MonthsWide() []string {
	return ks.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (ks *ks_IN) WeekdayAbbreviated(weekday time.Weekday) string {
	return ks.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (ks *ks_IN) WeekdaysAbbreviated() []string {
	return ks.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (ks *ks_IN) WeekdayNarrow(weekday time.Weekday) string {
	return ks.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (ks *ks_IN) WeekdaysNarrow() []string {
	return ks.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (ks *ks_IN) WeekdayShort(weekday time.Weekday) string {
	return ks.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (ks *ks_IN) WeekdaysShort() []string {
	return ks.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (ks *ks_IN) WeekdayWide(weekday time.Weekday) string {
	return ks.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (ks *ks_IN) WeekdaysWide() []string {
	return ks.daysWide
}

// Decimal returns the decimal point of number
func (ks *ks_IN) Decimal() string {
	return ks.decimal
}

// Group returns the group of number
func (ks *ks_IN) Group() string {
	return ks.group
}

// Group returns the minus sign of number
func (ks *ks_IN) Minus() string {
	return ks.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'ks_IN' and handles both Whole and Real numbers based on 'v'
func (ks *ks_IN) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 5 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	inSecondary := false
	groupThreshold := 3

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ks.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {

			if count == groupThreshold {
				b = append(b, ks.group[0])
				count = 1

				if !inSecondary {
					inSecondary = true
					groupThreshold = 2
				}
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(ks.minus) - 1; j >= 0; j-- {
			b = append(b, ks.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'ks_IN' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (ks *ks_IN) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 6
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ks.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(ks.minus) - 1; j >= 0; j-- {
			b = append(b, ks.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, ks.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'ks_IN'
func (ks *ks_IN) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ks.currencies[currency]
	l := len(s) + len(symbol) + 7 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	inSecondary := false
	groupThreshold := 3

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ks.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {

			if count == groupThreshold {
				b = append(b, ks.group[0])
				count = 1

				if !inSecondary {
					inSecondary = true
					groupThreshold = 2
				}
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	for j := len(symbol) - 1; j >= 0; j-- {
		b = append(b, symbol[j])
	}

	for j := len(ks.currencyPositivePrefix) - 1; j >= 0; j-- {
		b = append(b, ks.currencyPositivePrefix[j])
	}

	if num < 0 {
		for j := len(ks.minus) - 1; j >= 0; j-- {
			b = append(b, ks.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ks.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'ks_IN'
// in accounting notation.
func (ks *ks_IN) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ks.currencies[currency]
	l := len(s) + len(symbol) + 7 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	inSecondary := false
	groupThreshold := 3

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ks.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {

			if count == groupThreshold {
				b = append(b, ks.group[0])
				count = 1

				if !inSecondary {
					inSecondary = true
					groupThreshold = 2
				}
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

		for j := len(ks.currencyNegativePrefix) - 1; j >= 0; j-- {
			b = append(b, ks.currencyNegativePrefix[j])
		}

		for j := len(ks.minus) - 1; j >= 0; j-- {
			b = append(b, ks.minus[j])
		}

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(ks.currencyPositivePrefix) - 1; j >= 0; j-- {
			b = append(b, ks.currencyPositivePrefix[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ks.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'ks_IN'
func (ks *ks_IN) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0x2f}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2f}...)

	if t.Year() > 9 {
		b = append(b, strconv.Itoa(t.Year())[2:]...)
	} else {
		b = append(b, strconv.Itoa(t.Year())[1:]...)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'ks_IN'
func (ks *ks_IN) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, ks.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'ks_IN'
func (ks *ks_IN) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, ks.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'ks_IN'
func (ks *ks_IN) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, ks.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, ks.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'ks_IN'
func (ks *ks_IN) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ks.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ks.periodsAbbreviated[0]...)
	} else {
		b = append(b, ks.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'ks_IN'
func (ks *ks_IN) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ks.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ks.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ks.periodsAbbreviated[0]...)
	} else {
		b = append(b, ks.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'ks_IN'
func (ks *ks_IN) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ks.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ks.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ks.periodsAbbreviated[0]...)
	} else {
		b = append(b, ks.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'ks_IN'
func (ks *ks_IN) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ks.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ks.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ks.periodsAbbreviated[0]...)
	} else {
		b = append(b, ks.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := ks.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
