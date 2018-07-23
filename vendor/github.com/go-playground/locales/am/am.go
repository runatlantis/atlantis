package am

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type am struct {
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

// New returns a new instance of translator for the 'am' locale
func New() locales.Translator {
	return &am{
		locale:                 "am",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         []locales.PluralRule{6},
		pluralsRange:           []locales.PluralRule{2, 6},
		decimal:                ".",
		group:                  ",",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AU$", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "የቻይና ዩዋን", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ብር", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "₪", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JP¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "₩", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZ$", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "฿", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "NT$", "TZS", "UAH", "UAK", "UGS", "UGX", "US$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "₫", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "EC$", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: ")",
		monthsAbbreviated:      []string{"", "ጃንዩ", "ፌብሩ", "ማርች", "ኤፕሪ", "ሜይ", "ጁን", "ጁላይ", "ኦገስ", "ሴፕቴ", "ኦክቶ", "ኖቬም", "ዲሴም"},
		monthsNarrow:           []string{"", "ጃ", "ፌ", "ማ", "ኤ", "ሜ", "ጁ", "ጁ", "ኦ", "ሴ", "ኦ", "ኖ", "ዲ"},
		monthsWide:             []string{"", "ጃንዩወሪ", "ፌብሩወሪ", "ማርች", "ኤፕሪል", "ሜይ", "ጁን", "ጁላይ", "ኦገስት", "ሴፕቴምበር", "ኦክቶበር", "ኖቬምበር", "ዲሴምበር"},
		daysAbbreviated:        []string{"እሑድ", "ሰኞ", "ማክሰ", "ረቡዕ", "ሐሙስ", "ዓርብ", "ቅዳሜ"},
		daysNarrow:             []string{"እ", "ሰ", "ማ", "ረ", "ሐ", "ዓ", "ቅ"},
		daysShort:              []string{"እ", "ሰ", "ማ", "ረ", "ሐ", "ዓ", "ቅ"},
		daysWide:               []string{"እሑድ", "ሰኞ", "ማክሰኞ", "ረቡዕ", "ሐሙስ", "ዓርብ", "ቅዳሜ"},
		periodsAbbreviated:     []string{"ጥዋት", "ከሰዓት"},
		periodsNarrow:          []string{"ጠ", "ከ"},
		periodsWide:            []string{"ጥዋት", "ከሰዓት"},
		erasAbbreviated:        []string{"", ""},
		erasNarrow:             []string{"ዓ/ዓ", "ዓ/ም"},
		erasWide:               []string{"ዓመተ ዓለም", "ዓመተ ምሕረት"},
		timezones:              map[string]string{"SRT": "የሱሪናም ሰዓት", "TMST": "የቱርክመኒስታን ክረምት ሰዓት", "GYT": "የጉያና ሰዓት", "AEST": "የአውስትራሊያ ምዕራባዊ መደበኛ የሰዓት አቆጣጠር", "SAST": "የደቡብ አፍሪካ መደበኛ ሰዓት", "HNNOMX": "ሰሜናዊ ምእራብ የሜክሲኮ መደበኛ ሰዓት አቆጣጠር", "PST": "የፓስፊክ መደበኛ ሰዓት አቆጣጠር", "AWST": "የአውስትራሊያ ምስራቃዊ መደበኛ ሰዓት አቆጣጠር", "NZST": "የኒው ዚላንድ መደበኛ ሰዓት", "HEEG": "የምስራቅ ግሪንላንድ ክረምት ሰዓት", "MESZ": "የመካከለኛው አውሮፓ ክረምት ሰዓት", "ARST": "የአርጀንቲና የበጋ ሰዓት አቆጣጠር", "MST": "MST", "AST": "የአትላንቲክ መደበኛ የሰዓት አቆጣጠር", "HENOMX": "ሰሜናዊ ምእራብ የሜክሲኮ የቀን ሰዓት አቆጣጠር", "CHAST": "የቻታም መደበኛ ሰዓት", "PDT": "የፓስፊክ የቀን ሰዓት አቆጣጠር", "HEPMX": "የሜክሲኮ ፓሲፊክ የቀን ሰዓት አቆጣጠር", "ECT": "የኢኳዶር ሰዓት", "IST": "የህንድ መደበኛ ሰዓት", "WARST": "የአርጀንቲና ምስራቃዊ በጋ ሰዓት አቆጣጠር", "UYT": "የኡራጓይ መደበኛ ሰዓት", "OESZ": "የምስራቃዊ አውሮፓ ክረምት ሰዓት", "LHST": "የሎርድ ሆዌ መደበኛ የሰዓት አቆጣጠር", "LHDT": "የሎርድ ሆዌ የቀን ሰዓት አቆጣጠር", "CAT": "የመካከለኛው አፍሪካ ሰዓት", "COST": "የኮሎምቢያ ክረምት ሰዓት", "WAT": "የምዕራብ አፍሪካ መደበኛ ሰዓት", "ACDT": "የአውስትራሊያ መካከለኛ የቀን ሰዓት አቆጣጠር", "CLST": "የቺሊ ክረምት ሰዓት", "WITA": "የመካከለኛው ኢንዶኔዢያ ሰዓት", "OEZ": "የምስራቃዊ አውሮፓ መደበኛ ሰዓት", "HECU": "የኩባ የቀን ብርሃን ሰዓት", "EDT": "የምዕራባዊ የቀን ሰዓት አቆጣጠር", "ACWDT": "የአውስትራሊያ መካከለኛው ምስራቅ የቀን ሰዓት አቆጣጠር", "HAT": "የኒውፋውንድላንድ የቀን የሰዓት አቆጣጠር", "SGT": "የሲንጋፒር መደበኛ ሰዓት", "WIT": "የምስራቃዊ ኢንዶኔዢያ ሰዓት", "CST": "የመካከለኛ መደበኛ ሰዓት አቆጣጠር", "WAST": "የምዕራብ አፍሪካ ክረምት ሰዓት", "WIB": "የምዕራባዊ ኢንዶኔዢያ ሰዓት", "TMT": "የቱርክመኒስታን መደበኛ ሰዓት", "HNCU": "የኩባ መደበኛ ሰዓት", "AWDT": "የአውስትራሊያ ምስራቃዊ የቀን ሰዓት አቆጣጠር", "BT": "የቡታን ሰዓት", "JST": "የጃፓን መደበኛ ሰዓት", "AKDT": "የአላስካ የቀን ሰዓት አቆጣጠር", "HNEG": "የምስራቅ ግሪንላንድ መደበኛ ሰዓት", "WART": "የምዕራባዊ አርጀንቲና መደበኛ ሰዓት አቆጣጠር", "UYST": "የኡራጓይ ክረምት ሰዓት", "HNPM": "ቅዱስ የፒዬር እና ሚኴሎን መደበኛ ሰዓት", "CLT": "የቺሊ መደበኛ ሰዓት", "WEZ": "የምዕራባዊ አውሮፓ መደበኛ ሰዓት", "HEOG": "የምዕራብ ግሪንላንድ ክረምት ሰዓት", "MEZ": "የመካከለኛው አውሮፓ መደበኛ ሰዓት", "VET": "የቬኔዝዌላ ሰዓት", "HNPMX": "የሜክሲኮ ፓሲፊክ መደበኛ ሰዓት አቆጣጠር", "COT": "የኮሎምቢያ መደበኛ ሰዓት", "HADT": "የሃዋይ አሌኡት የቀን ሰዓት አቆጣጠር", "∅∅∅": "የአማዞን የቀን ሰዓት አቆጣጠር", "ADT": "የአትላንቲክ የቀን ሰዓት አቆጣጠር", "EST": "የምዕራባዊ መደበኛ የሰዓት አቆጣጠር", "MDT": "MDT", "HEPM": "ቅዱስ የፒዬር እና ሚኴሎን የቀን ብርሃን ሰዓት", "GMT": "ግሪንዊች ማዕከላዊ ሰዓት", "ChST": "የቻሞሮ መደበኛ ሰዓት", "CDT": "የመካከለኛ የቀን ሰዓት አቆጣጠር", "GFT": "የፈረንሳይ ጉያና ሰዓት", "HKT": "የሆንግ ኮንግ መደበኛ ሰዓት", "HNT": "የኒውፋውንድላንድ መደበኛ የሰዓት አቆጣጠር", "MYT": "የማሌይዢያ ሰዓት", "JDT": "የጃፓን የቀን ብርሃን ሰዓት", "AKST": "የአላስካ መደበኛ የሰዓት አቆጣጠር", "HKST": "የሆንግ ኮንግ ክረምት ሰዓት", "AEDT": "የአውስትራሊያ ምዕራባዊ የቀን ሰዓት አቆጣጠር", "HAST": "የሃዋይ አሌኡት መደበኛ ሰዓት አቆጣጠር", "ART": "የአርጀንቲና መደበኛ ሰዓት አቆጣጠር", "CHADT": "የቻታም የቀን ብርሃን ሰዓት", "WESZ": "የምዕራባዊ አውሮፓ ክረምት ሰዓት", "NZDT": "የኒው ዚላንድ የቀን ብርሃን ሰዓት", "BOT": "የቦሊቪያ ሰዓት", "ACST": "የአውስትራሊያ መካከለኛ መደበኛ የሰዓት አቆጣጠር", "EAT": "የምስራቅ አፍሪካ ሰዓት", "HNOG": "የምዕራብ ግሪንላንድ መደበኛ ሰዓት", "ACWST": "የአውስትራሊያ መካከለኛ ምስራቃዊ መደበኛ ሰዓት አቆጣጠር"},
	}
}

// Locale returns the current translators string locale
func (am *am) Locale() string {
	return am.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'am'
func (am *am) PluralsCardinal() []locales.PluralRule {
	return am.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'am'
func (am *am) PluralsOrdinal() []locales.PluralRule {
	return am.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'am'
func (am *am) PluralsRange() []locales.PluralRule {
	return am.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'am'
func (am *am) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)

	if (i == 0) || (n == 1) {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'am'
func (am *am) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'am'
func (am *am) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := am.CardinalPluralRule(num1, v1)
	end := am.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (am *am) MonthAbbreviated(month time.Month) string {
	return am.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (am *am) MonthsAbbreviated() []string {
	return am.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (am *am) MonthNarrow(month time.Month) string {
	return am.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (am *am) MonthsNarrow() []string {
	return am.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (am *am) MonthWide(month time.Month) string {
	return am.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (am *am) MonthsWide() []string {
	return am.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (am *am) WeekdayAbbreviated(weekday time.Weekday) string {
	return am.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (am *am) WeekdaysAbbreviated() []string {
	return am.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (am *am) WeekdayNarrow(weekday time.Weekday) string {
	return am.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (am *am) WeekdaysNarrow() []string {
	return am.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (am *am) WeekdayShort(weekday time.Weekday) string {
	return am.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (am *am) WeekdaysShort() []string {
	return am.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (am *am) WeekdayWide(weekday time.Weekday) string {
	return am.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (am *am) WeekdaysWide() []string {
	return am.daysWide
}

// Decimal returns the decimal point of number
func (am *am) Decimal() string {
	return am.decimal
}

// Group returns the group of number
func (am *am) Group() string {
	return am.group
}

// Group returns the minus sign of number
func (am *am) Minus() string {
	return am.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'am' and handles both Whole and Real numbers based on 'v'
func (am *am) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, am.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, am.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, am.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'am' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (am *am) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, am.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, am.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, am.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'am'
func (am *am) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := am.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, am.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, am.group[0])
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
		b = append(b, am.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, am.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'am'
// in accounting notation.
func (am *am) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := am.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, am.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, am.group[0])
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

		b = append(b, am.currencyNegativePrefix[0])

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
			b = append(b, am.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, am.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'am'
func (am *am) FmtDateShort(t time.Time) string {

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

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'am'
func (am *am) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, am.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'am'
func (am *am) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, am.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'am'
func (am *am) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, am.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20, 0xe1, 0x8d, 0xa3}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, am.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'am'
func (am *am) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, am.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, am.periodsAbbreviated[0]...)
	} else {
		b = append(b, am.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'am'
func (am *am) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, am.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, am.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, am.periodsAbbreviated[0]...)
	} else {
		b = append(b, am.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'am'
func (am *am) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, am.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, am.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, am.periodsAbbreviated[0]...)
	} else {
		b = append(b, am.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'am'
func (am *am) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, am.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, am.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, am.periodsAbbreviated[0]...)
	} else {
		b = append(b, am.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := am.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
