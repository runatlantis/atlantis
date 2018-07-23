package ar_MR

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type ar_MR struct {
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

// New returns a new instance of translator for the 'ar_MR' locale
func New() locales.Translator {
	return &ar_MR{
		locale:                 "ar_MR",
		pluralsCardinal:        []locales.PluralRule{1, 2, 3, 4, 5, 6},
		pluralsOrdinal:         []locales.PluralRule{6},
		pluralsRange:           []locales.PluralRule{1, 4, 5, 6},
		decimal:                ",",
		group:                  ".",
		minus:                  "؜-",
		percent:                "٪؜",
		perMille:               "؉",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		percentSuffix:          " ",
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "يناير", "فبراير", "مارس", "إبريل", "مايو", "يونيو", "يوليو", "أغشت", "شتمبر", "أكتوبر", "نوفمبر", "دجمبر"},
		monthsNarrow:           []string{"", "ي", "ف", "م", "إ", "و", "ن", "ل", "غ", "ش", "ك", "ب", "د"},
		monthsWide:             []string{"", "يناير", "فبراير", "مارس", "إبريل", "مايو", "يونيو", "يوليو", "أغشت", "شتمبر", "أكتوبر", "نوفمبر", "دجمبر"},
		daysAbbreviated:        []string{"الأحد", "الاثنين", "الثلاثاء", "الأربعاء", "الخميس", "الجمعة", "السبت"},
		daysNarrow:             []string{"ح", "ن", "ث", "ر", "خ", "ج", "س"},
		daysShort:              []string{"أحد", "إثنين", "ثلاثاء", "أربعاء", "خميس", "جمعة", "سبت"},
		daysWide:               []string{"الأحد", "الاثنين", "الثلاثاء", "الأربعاء", "الخميس", "الجمعة", "السبت"},
		periodsAbbreviated:     []string{"ص", "م"},
		periodsNarrow:          []string{"ص", "م"},
		periodsWide:            []string{"ص", "م"},
		erasAbbreviated:        []string{"ق.م", "م"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"قبل الميلاد", "ميلادي"},
		timezones:              map[string]string{"COT": "توقيت كولومبيا الرسمي", "CDT": "التوقيت الصيفي المركزي لأمريكا الشمالية", "HAST": "توقيت هاواي ألوتيان الرسمي", "ARST": "توقيت الأرجنتين الصيفي", "ART": "توقيت الأرجنتين الرسمي", "SAST": "توقيت جنوب أفريقيا", "WEZ": "توقيت غرب أوروبا الرسمي", "ACWDT": "توقيت غرب وسط أستراليا الصيفي", "HADT": "توقيت هاواي ألوتيان الصيفي", "PST": "توقيت المحيط الهادي الرسمي", "HEEG": "توقيت شرق غرينلاند الصيفي", "HNPM": "توقيت سانت بيير وميكولون الرسمي", "HAT": "توقيت نيوفاوندلاند الصيفي", "COST": "توقيت كولومبيا الصيفي", "BT": "توقيت بوتان", "ECT": "توقيت الإكوادور", "HNOG": "توقيت غرب غرينلاند الرسمي", "MESZ": "توقيت وسط أوروبا الصيفي", "AEDT": "توقيت شرق أستراليا الصيفي", "WAT": "توقيت غرب أفريقيا الرسمي", "WESZ": "توقيت غرب أوروبا الصيفي", "NZDT": "توقيت نيوزيلندا الصيفي", "GFT": "توقيت غايانا الفرنسية", "OESZ": "توقيت شرق أوروبا الصيفي", "UYST": "توقيت أوروغواي الصيفي", "WARST": "توقيت غرب الأرجنتين الصيفي", "AWST": "توقيت غرب أستراليا الرسمي", "HNT": "توقيت نيوفاوندلاند الرسمي", "HEPM": "توقيت سانت بيير وميكولون الصيفي", "TMST": "توقيت تركمانستان الصيفي", "CST": "التوقيت الرسمي المركزي لأمريكا الشمالية", "JDT": "توقيت اليابان الصيفي", "HKT": "توقيت هونغ كونغ الرسمي", "WART": "توقيت غرب الأرجنتين الرسمي", "EAT": "توقيت شرق أفريقيا", "CHAST": "توقيت تشاتام الرسمي", "ADT": "التوقيت الصيفي الأطلسي", "ACWST": "توقيت غرب وسط أستراليا الرسمي", "MST": "MST", "IST": "توقيت الهند", "HNNOMX": "التوقيت الرسمي لشمال غرب المكسيك", "MDT": "MDT", "CLST": "توقيت شيلي الصيفي", "HEOG": "توقيت غرب غرينلاند الصيفي", "HKST": "توقيت هونغ كونغ الصيفي", "VET": "توقيت فنزويلا", "HENOMX": "التوقيت الصيفي لشمال غرب المكسيك", "CAT": "توقيت وسط أفريقيا", "PDT": "توقيت المحيط الهادي الصيفي", "AWDT": "توقيت غرب أستراليا الصيفي", "AEST": "توقيت شرق أستراليا الرسمي", "EDT": "التوقيت الصيفي الشرقي لأمريكا الشمالية", "ACST": "توقيت وسط أستراليا الرسمي", "BOT": "توقيت بوليفيا", "WITA": "توقيت وسط إندونيسيا", "HECU": "توقيت كوبا الصيفي", "HNPMX": "توقيت المحيط الهادي الرسمي للمكسيك", "HEPMX": "توقيت المحيط الهادي الصيفي للمكسيك", "WAST": "توقيت غرب أفريقيا الصيفي", "HNEG": "توقيت شرق غرينلاند الرسمي", "LHST": "توقيت لورد هاو الرسمي", "CLT": "توقيت شيلي الرسمي", "HNCU": "توقيت كوبا الرسمي", "WIB": "توقيت غرب إندونيسيا", "JST": "توقيت اليابان الرسمي", "GMT": "توقيت غرينتش", "AST": "التوقيت الرسمي الأطلسي", "ChST": "توقيت تشامورو", "SGT": "توقيت سنغافورة", "∅∅∅": "توقيت أزورس الصيفي", "OEZ": "توقيت شرق أوروبا الرسمي", "UYT": "توقيت أوروغواي الرسمي", "CHADT": "توقيت تشاتام الصيفي", "MYT": "توقيت ماليزيا", "AKST": "التوقيت الرسمي لألاسكا", "AKDT": "توقيت ألاسكا الصيفي", "MEZ": "توقيت وسط أوروبا الرسمي", "WIT": "توقيت شرق إندونيسيا", "LHDT": "التوقيت الصيفي للورد هاو", "SRT": "توقيت سورينام", "TMT": "توقيت تركمانستان الرسمي", "GYT": "توقيت غيانا", "NZST": "توقيت نيوزيلندا الرسمي", "EST": "التوقيت الرسمي الشرقي لأمريكا الشمالية", "ACDT": "توقيت وسط أستراليا الصيفي"},
	}
}

// Locale returns the current translators string locale
func (ar *ar_MR) Locale() string {
	return ar.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'ar_MR'
func (ar *ar_MR) PluralsCardinal() []locales.PluralRule {
	return ar.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'ar_MR'
func (ar *ar_MR) PluralsOrdinal() []locales.PluralRule {
	return ar.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'ar_MR'
func (ar *ar_MR) PluralsRange() []locales.PluralRule {
	return ar.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'ar_MR'
func (ar *ar_MR) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	nMod100 := math.Mod(n, 100)

	if n == 0 {
		return locales.PluralRuleZero
	} else if n == 1 {
		return locales.PluralRuleOne
	} else if n == 2 {
		return locales.PluralRuleTwo
	} else if nMod100 >= 3 && nMod100 <= 10 {
		return locales.PluralRuleFew
	} else if nMod100 >= 11 && nMod100 <= 99 {
		return locales.PluralRuleMany
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'ar_MR'
func (ar *ar_MR) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'ar_MR'
func (ar *ar_MR) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := ar.CardinalPluralRule(num1, v1)
	end := ar.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleZero && end == locales.PluralRuleOne {
		return locales.PluralRuleZero
	} else if start == locales.PluralRuleZero && end == locales.PluralRuleTwo {
		return locales.PluralRuleZero
	} else if start == locales.PluralRuleZero && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleZero && end == locales.PluralRuleMany {
		return locales.PluralRuleMany
	} else if start == locales.PluralRuleZero && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleTwo {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleMany {
		return locales.PluralRuleMany
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleTwo && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleTwo && end == locales.PluralRuleMany {
		return locales.PluralRuleMany
	} else if start == locales.PluralRuleTwo && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleFew && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleFew && end == locales.PluralRuleMany {
		return locales.PluralRuleMany
	} else if start == locales.PluralRuleFew && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleMany && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleMany && end == locales.PluralRuleMany {
		return locales.PluralRuleMany
	} else if start == locales.PluralRuleMany && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleOne {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleTwo {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleMany {
		return locales.PluralRuleMany
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (ar *ar_MR) MonthAbbreviated(month time.Month) string {
	return ar.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (ar *ar_MR) MonthsAbbreviated() []string {
	return ar.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (ar *ar_MR) MonthNarrow(month time.Month) string {
	return ar.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (ar *ar_MR) MonthsNarrow() []string {
	return ar.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (ar *ar_MR) MonthWide(month time.Month) string {
	return ar.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (ar *ar_MR) MonthsWide() []string {
	return ar.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (ar *ar_MR) WeekdayAbbreviated(weekday time.Weekday) string {
	return ar.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (ar *ar_MR) WeekdaysAbbreviated() []string {
	return ar.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (ar *ar_MR) WeekdayNarrow(weekday time.Weekday) string {
	return ar.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (ar *ar_MR) WeekdaysNarrow() []string {
	return ar.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (ar *ar_MR) WeekdayShort(weekday time.Weekday) string {
	return ar.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (ar *ar_MR) WeekdaysShort() []string {
	return ar.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (ar *ar_MR) WeekdayWide(weekday time.Weekday) string {
	return ar.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (ar *ar_MR) WeekdaysWide() []string {
	return ar.daysWide
}

// Decimal returns the decimal point of number
func (ar *ar_MR) Decimal() string {
	return ar.decimal
}

// Group returns the group of number
func (ar *ar_MR) Group() string {
	return ar.group
}

// Group returns the minus sign of number
func (ar *ar_MR) Minus() string {
	return ar.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'ar_MR' and handles both Whole and Real numbers based on 'v'
func (ar *ar_MR) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ar.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ar.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(ar.minus) - 1; j >= 0; j-- {
			b = append(b, ar.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'ar_MR' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (ar *ar_MR) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 10
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ar.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(ar.minus) - 1; j >= 0; j-- {
			b = append(b, ar.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, ar.percentSuffix...)

	b = append(b, ar.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'ar_MR'
func (ar *ar_MR) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ar.currencies[currency]
	l := len(s) + len(symbol) + 6 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ar.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ar.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(ar.minus) - 1; j >= 0; j-- {
			b = append(b, ar.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ar.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, ar.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'ar_MR'
// in accounting notation.
func (ar *ar_MR) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ar.currencies[currency]
	l := len(s) + len(symbol) + 6 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ar.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ar.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		for j := len(ar.minus) - 1; j >= 0; j-- {
			b = append(b, ar.minus[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ar.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, ar.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, ar.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'ar_MR'
func (ar *ar_MR) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0xe2, 0x80, 0x8f, 0x2f}...)
	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0xe2, 0x80, 0x8f, 0x2f}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'ar_MR'
func (ar *ar_MR) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0xe2, 0x80, 0x8f, 0x2f}...)

	if t.Month() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Month()), 10)

	b = append(b, []byte{0xe2, 0x80, 0x8f, 0x2f}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'ar_MR'
func (ar *ar_MR) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ar.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'ar_MR'
func (ar *ar_MR) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, ar.daysWide[t.Weekday()]...)
	b = append(b, []byte{0xd8, 0x8c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ar.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'ar_MR'
func (ar *ar_MR) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ar.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ar.periodsAbbreviated[0]...)
	} else {
		b = append(b, ar.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'ar_MR'
func (ar *ar_MR) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ar.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ar.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ar.periodsAbbreviated[0]...)
	} else {
		b = append(b, ar.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'ar_MR'
func (ar *ar_MR) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ar.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ar.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ar.periodsAbbreviated[0]...)
	} else {
		b = append(b, ar.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'ar_MR'
func (ar *ar_MR) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ar.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ar.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ar.periodsAbbreviated[0]...)
	} else {
		b = append(b, ar.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := ar.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
