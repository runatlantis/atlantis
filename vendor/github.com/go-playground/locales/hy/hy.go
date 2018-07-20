package hy

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type hy struct {
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

// New returns a new instance of translator for the 'hy' locale
func New() locales.Translator {
	return &hy{
		locale:                 "hy",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         []locales.PluralRule{2, 6},
		pluralsRange:           []locales.PluralRule{2, 6},
		decimal:                ",",
		group:                  " ",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "֏", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "A$", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "₪", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JP¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "₩", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZ$", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "฿", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "NT$", "TZS", "UAH", "UAK", "UGS", "UGX", "$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "₫", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "EC$", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "հնվ", "փտվ", "մրտ", "ապր", "մյս", "հնս", "հլս", "օգս", "սեպ", "հոկ", "նոյ", "դեկ"},
		monthsNarrow:           []string{"", "Հ", "Փ", "Մ", "Ա", "Մ", "Հ", "Հ", "Օ", "Ս", "Հ", "Ն", "Դ"},
		monthsWide:             []string{"", "հունվարի", "փետրվարի", "մարտի", "ապրիլի", "մայիսի", "հունիսի", "հուլիսի", "օգոստոսի", "սեպտեմբերի", "հոկտեմբերի", "նոյեմբերի", "դեկտեմբերի"},
		daysAbbreviated:        []string{"կիր", "երկ", "երք", "չրք", "հնգ", "ուր", "շբթ"},
		daysNarrow:             []string{"Կ", "Ե", "Ե", "Չ", "Հ", "Ո", "Շ"},
		daysShort:              []string{"կր", "եկ", "եք", "չք", "հգ", "ու", "շբ"},
		daysWide:               []string{"կիրակի", "երկուշաբթի", "երեքշաբթի", "չորեքշաբթի", "հինգշաբթի", "ուրբաթ", "շաբաթ"},
		periodsAbbreviated:     []string{"ԿԱ", "ԿՀ"},
		periodsNarrow:          []string{"ա", "հ"},
		periodsWide:            []string{"ԿԱ", "ԿՀ"},
		erasAbbreviated:        []string{"մ.թ.ա.", "մ.թ."},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"Քրիստոսից առաջ", "Քրիստոսից հետո"},
		timezones:              map[string]string{"PDT": "Խաղաղօվկիանոսյան ամառային ժամանակ", "WEZ": "Արևմտյան Եվրոպայի ստանդարտ ժամանակ", "WIB": "Արևմտյան Ինդոնեզիայի ժամանակ", "GFT": "Ֆրանսիական Գվիանայի ժամանակ", "HEPM": "Սեն Պիեռ և Միքելոնի ամառային ժամանակ", "VET": "Վենեսուելայի ժամանակ", "CAT": "Կենտրոնական Աֆրիկայի ժամանակ", "HADT": "Հավայան-ալեության ամառային ժամանակ", "AKST": "Ալյասկայի ստանդարտ ժամանակ", "HAT": "Նյուֆաունդլենդի ամառային ժամանակ", "MST": "MST", "CLT": "Չիլիի ստանդարտ ժամանակ", "SAST": "Հարավային Աֆրիկայի ժամանակ", "BOT": "Բոլիվիայի ժամանակ", "HNEG": "Արևելյան Գրենլանդիայի ստանդարտ ժամանակ", "∅∅∅": "Ազորյան կղզիների ամառային ժամանակ", "OESZ": "Արևելյան Եվրոպայի ամառային ժամանակ", "AST": "Ատլանտյան ստանդարտ ժամանակ", "HEOG": "Արևմտյան Գրենլանդիայի ամառային ժամանակ", "HNCU": "Կուբայի ստանդարտ ժամանակ", "MYT": "Մալայզիայի ժամանակ", "HNOG": "Արևմտյան Գրենլանդիայի ստանդարտ ժամանակ", "MESZ": "Կենտրոնական Եվրոպայի ամառային ժամանակ", "ChST": "Չամոռոյի ժամանակ", "ADT": "Ատլանտյան ամառային ժամանակ", "HKT": "Հոնկոնգի ստանդարտ ժամանակ", "IST": "Հնդկաստանի ստանդարտ ժամանակ", "SRT": "Սուրինամի ժամանակ", "CHAST": "Չաթեմ կղզու ստանդարտ ժամանակ", "ACWST": "Կենտրոնական Ավստրալիայի արևմտյան ստանդարտ ժամանակ", "ACWDT": "Կենտրոնական Ավստրալիայի արևմտյան ամառային ժամանակ", "HEEG": "Արևելյան Գրենլանդիայի ամառային ժամանակ", "MEZ": "Կենտրոնական Եվրոպայի ստանդարտ ժամանակ", "COST": "Կոլումբիայի ամառային ժամանակ", "HNPMX": "Մեքսիկայի խաղաղօվկիանոսյան ստանդարտ ժամանակ", "PST": "Խաղաղօվկիանոսյան ստանդարտ ժամանակ", "AEST": "Արևելյան Ավստրալիայի ստանդարտ ժամանակ", "WART": "Արևմտյան Արգենտինայի ստնադարտ ժամանակ", "HNPM": "Սեն Պիեռ և Միքելոնի ստանդարտ ժամանակ", "WITA": "Կենտրոնական Ինդոնեզիայի ժամանակ", "WIT": "Արևելյան Ինդոնեզիայի ժամանակ", "WAST": "Արևմտյան Աֆրիկայի ամառային ժամանակ", "JDT": "Ճապոնիայի ամառային ժամանակ", "EST": "Արևելյան Ամերիկայի ստանդարտ ժամանակ", "CLST": "Չիլիի ամառային ժամանակ", "TMT": "Թուրքմենստանի ստանդարտ ժամանակ", "COT": "Կոլումբիայի ստանդարտ ժամանակ", "EDT": "Արևելյան Ամերիկայի ամառային ժամանակ", "ART": "Արգենտինայի ստնադարտ ժամանակ", "GYT": "Գայանայի ժամանակ", "HEPMX": "Մեքսիկայի խաղաղօվկիանոսյան ամառային ժամանակ", "AEDT": "Արևելյան Ավստրալիայի ամառային ժամանակ", "WESZ": "Արևմտյան Եվրոպայի ամառային ժամանակ", "SGT": "Սինգապուրի ժամանակ", "LHST": "Լորդ Հաուի ստանդարտ ժամանակ", "LHDT": "Լորդ Հաուի ամառային ժամանակ", "EAT": "Արևելյան Աֆրիկայի ժամանակ", "AWDT": "Արևմտյան Ավստրալիայի ամառային ժամանակ", "ACDT": "Կենտրոնական Ավստրալիայի ամառային ժամանակ", "CDT": "Կենտրոնական Ամերիկայի ամառային ժամանակ", "TMST": "Թուրքմենստանի ամառային ժամանակ", "UYST": "Ուրուգվայի ամառային ժամանակ", "WAT": "Արևմտյան Աֆրիկայի ստանդարտ ժամանակ", "ECT": "Էկվադորի ժամանակ", "GMT": "Գրինվիչի ժամանակ", "NZDT": "Նոր Զելանդիայի ամառային ժամանակ", "ACST": "Կենտրոնական Ավստրալիայի ստանդարտ ժամանակ", "HKST": "Հոնկոնգի ամառային ժամանակ", "HNNOMX": "Հյուսիսարևմտյան Մեքսիկայի ստանդարտ ժամանակ", "ARST": "Արգենտինայի ամառային ժամանակ", "CHADT": "Չաթեմ կղզու ամառային ժամանակ", "NZST": "Նոր Զելանդիայի ստանդարտ ժամանակ", "HENOMX": "Հյուսիսարևմտյան Մեքսիկայի ամառային ժամանակ", "OEZ": "Արևելյան Եվրոպայի ստանդարտ ժամանակ", "HAST": "Հավայան-ալեության ստանդարտ ժամանակ", "UYT": "Ուրուգվայի ստանդարտ ժամանակ", "CST": "Կենտրոնական Ամերիկայի ստանդարտ ժամանակ", "JST": "Ճապոնիայի ստանդարտ ժամանակ", "BT": "Բութանի ժամանակ", "AKDT": "Ալյասկայի ամառային ժամանակ", "WARST": "Արևմտյան Արգենտինայի ամառային ժամանակ", "HNT": "Նյուֆաունդլենդի ստանդարտ ժամանակ", "MDT": "MDT", "HECU": "Կուբայի ամառային ժամանակ", "AWST": "Արևմտյան Ավստրալիայի ստանդարտ ժամանակ"},
	}
}

// Locale returns the current translators string locale
func (hy *hy) Locale() string {
	return hy.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'hy'
func (hy *hy) PluralsCardinal() []locales.PluralRule {
	return hy.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'hy'
func (hy *hy) PluralsOrdinal() []locales.PluralRule {
	return hy.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'hy'
func (hy *hy) PluralsRange() []locales.PluralRule {
	return hy.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'hy'
func (hy *hy) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)

	if i == 0 || i == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'hy'
func (hy *hy) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'hy'
func (hy *hy) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := hy.CardinalPluralRule(num1, v1)
	end := hy.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (hy *hy) MonthAbbreviated(month time.Month) string {
	return hy.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (hy *hy) MonthsAbbreviated() []string {
	return hy.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (hy *hy) MonthNarrow(month time.Month) string {
	return hy.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (hy *hy) MonthsNarrow() []string {
	return hy.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (hy *hy) MonthWide(month time.Month) string {
	return hy.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (hy *hy) MonthsWide() []string {
	return hy.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (hy *hy) WeekdayAbbreviated(weekday time.Weekday) string {
	return hy.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (hy *hy) WeekdaysAbbreviated() []string {
	return hy.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (hy *hy) WeekdayNarrow(weekday time.Weekday) string {
	return hy.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (hy *hy) WeekdaysNarrow() []string {
	return hy.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (hy *hy) WeekdayShort(weekday time.Weekday) string {
	return hy.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (hy *hy) WeekdaysShort() []string {
	return hy.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (hy *hy) WeekdayWide(weekday time.Weekday) string {
	return hy.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (hy *hy) WeekdaysWide() []string {
	return hy.daysWide
}

// Decimal returns the decimal point of number
func (hy *hy) Decimal() string {
	return hy.decimal
}

// Group returns the group of number
func (hy *hy) Group() string {
	return hy.group
}

// Group returns the minus sign of number
func (hy *hy) Minus() string {
	return hy.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'hy' and handles both Whole and Real numbers based on 'v'
func (hy *hy) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, hy.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(hy.group) - 1; j >= 0; j-- {
					b = append(b, hy.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, hy.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'hy' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (hy *hy) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, hy.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, hy.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, hy.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'hy'
func (hy *hy) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := hy.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, hy.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(hy.group) - 1; j >= 0; j-- {
					b = append(b, hy.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, hy.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, hy.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, hy.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'hy'
// in accounting notation.
func (hy *hy) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := hy.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, hy.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(hy.group) - 1; j >= 0; j-- {
					b = append(b, hy.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, hy.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, hy.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, hy.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, hy.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'hy'
func (hy *hy) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'hy'
func (hy *hy) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, hy.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20, 0xd5, 0xa9, 0x2e}...)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'hy'
func (hy *hy) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, hy.monthsWide[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20, 0xd5, 0xa9, 0x2e}...)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'hy'
func (hy *hy) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20, 0xd5, 0xa9, 0x2e, 0x20}...)
	b = append(b, hy.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, hy.daysWide[t.Weekday()]...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'hy'
func (hy *hy) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, hy.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'hy'
func (hy *hy) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, hy.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, hy.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'hy'
func (hy *hy) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, hy.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, hy.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'hy'
func (hy *hy) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, hy.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, hy.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := hy.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
