package ta_MY

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type ta_MY struct {
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

// New returns a new instance of translator for the 'ta_MY' locale
func New() locales.Translator {
	return &ta_MY{
		locale:                 "ta_MY",
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
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "RM", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "S$", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositivePrefix: " ",
		currencyNegativePrefix: " ",
		monthsAbbreviated:      []string{"", "ஜன.", "பிப்.", "மார்.", "ஏப்.", "மே", "ஜூன்", "ஜூலை", "ஆக.", "செப்.", "அக்.", "நவ.", "டிச."},
		monthsNarrow:           []string{"", "ஜ", "பி", "மா", "ஏ", "மே", "ஜூ", "ஜூ", "ஆ", "செ", "அ", "ந", "டி"},
		monthsWide:             []string{"", "ஜனவரி", "பிப்ரவரி", "மார்ச்", "ஏப்ரல்", "மே", "ஜூன்", "ஜூலை", "ஆகஸ்ட்", "செப்டம்பர்", "அக்டோபர்", "நவம்பர்", "டிசம்பர்"},
		daysAbbreviated:        []string{"ஞாயி.", "திங்.", "செவ்.", "புத.", "வியா.", "வெள்.", "சனி"},
		daysNarrow:             []string{"ஞா", "தி", "செ", "பு", "வி", "வெ", "ச"},
		daysShort:              []string{"ஞா", "தி", "செ", "பு", "வி", "வெ", "ச"},
		daysWide:               []string{"ஞாயிறு", "திங்கள்", "செவ்வாய்", "புதன்", "வியாழன்", "வெள்ளி", "சனி"},
		periodsAbbreviated:     []string{"முற்பகல்", "பிற்பகல்"},
		periodsNarrow:          []string{"மு.ப", "பி.ப"},
		periodsWide:            []string{"முற்பகல்", "பிற்பகல்"},
		erasAbbreviated:        []string{"கி.மு.", "கி.பி."},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"கிறிஸ்துவுக்கு முன்", "அன்னோ டோமினி"},
		timezones:              map[string]string{"CDT": "மத்திய பகலொளி நேரம்", "UYT": "உருகுவே நிலையான நேரம்", "HNCU": "கியூபா நிலையான நேரம்", "JST": "ஜப்பான் நிலையான நேரம்", "IST": "இந்திய நிலையான நேரம்", "EAT": "கிழக்கு ஆப்பிரிக்க நேரம்", "HADT": "ஹவாய்-அலேஷியன் பகலொளி நேரம்", "MST": "மவுன்டைன் நிலையான நேரம்", "MYT": "மலேஷிய நேரம்", "JDT": "ஜப்பான் பகலொளி நேரம்", "AKDT": "அலாஸ்கா பகலொளி நேரம்", "EST": "கிழக்கத்திய நிலையான நேரம்", "HNNOMX": "வடமேற்கு மெக்ஸிகோ நிலையான நேரம்", "AWST": "ஆஸ்திரேலியன் மேற்கத்திய நிலையான நேரம்", "WAT": "மேற்கு ஆப்பிரிக்க நிலையான நேரம்", "ACDT": "ஆஸ்திரேலியன் மத்திய பகலொளி நேரம்", "HKST": "ஹாங்காங் கோடை நேரம்", "MEZ": "மத்திய ஐரோப்பிய நிலையான நேரம்", "HENOMX": "வடமேற்கு மெக்ஸிகோ பகலொளி நேரம்", "CLST": "சிலி கோடை நேரம்", "GYT": "கயானா நேரம்", "WIB": "மேற்கத்திய இந்தோனேசிய நேரம்", "ADT": "அட்லாண்டிக் பகலொளி நேரம்", "WAST": "மேற்கு ஆப்பிரிக்க கோடை நேரம்", "BOT": "பொலிவியா நேரம்", "BT": "பூடான் நேரம்", "HNEG": "கிழக்கு கிரீன்லாந்து நிலையான நேரம்", "SRT": "சுரினாம் நேரம்", "CHAST": "சத்தாம் நிலையான நேரம்", "HNPMX": "மெக்ஸிகன் பசிபிக் நிலையான நேரம்", "ECT": "ஈக்வடார் நேரம்", "ACWST": "ஆஸ்திரேலியன் மத்திய மேற்கத்திய நிலையான நேரம்", "VET": "வெனிசுலா நேரம்", "WIT": "கிழக்கத்திய இந்தோனேசிய நேரம்", "TMST": "துர்க்மெனிஸ்தான் கோடை நேரம்", "OESZ": "கிழக்கத்திய ஐரோப்பிய கோடை நேரம்", "PDT": "பசிபிக் பகலொளி நேரம்", "ACWDT": "ஆஸ்திரேலியன் மத்திய மேற்கத்திய பகலொளி நேரம்", "EDT": "கிழக்கத்திய பகலொளி நேரம்", "WESZ": "மேற்கத்திய ஐரோப்பிய கோடை நேரம்", "GFT": "ஃபிரஞ்சு கயானா நேரம்", "HEOG": "மேற்கு கிரீன்லாந்து கோடை நேரம்", "CAT": "மத்திய ஆப்பிரிக்க நேரம்", "CHADT": "சத்தாம் பகலொளி நேரம்", "AST": "அட்லாண்டிக் நிலையான நேரம்", "HEPMX": "மெக்ஸிகன் பசிபிக் பகலொளி நேரம்", "LHDT": "லார்ட் ஹோவ் பகலொளி நேரம்", "WITA": "மத்திய இந்தோனேசிய நேரம்", "HNT": "நியூஃபவுண்ட்லாந்து நிலையான நேரம்", "ARST": "அர்ஜென்டினா கோடை நேரம்", "HECU": "கியூபா பகலொளி நேரம்", "WEZ": "மேற்கத்திய ஐரோப்பிய நிலையான நேரம்", "NZST": "நியூசிலாந்து நிலையான நேரம்", "WARST": "மேற்கத்திய அர்ஜென்டினா கோடை நேரம்", "COT": "கொலம்பியா நிலையான நேரம்", "COST": "கொலம்பியா கோடை நேரம்", "SAST": "தென் ஆப்பிரிக்க நிலையான நேரம்", "MESZ": "மத்திய ஐரோப்பிய கோடை நேரம்", "WART": "மேற்கத்திய அர்ஜென்டினா நிலையான நேரம்", "OEZ": "கிழக்கத்திய ஐரோப்பிய நிலையான நேரம்", "CST": "மத்திய நிலையான நேரம்", "SGT": "சிங்கப்பூர் நிலையான நேரம்", "HNOG": "மேற்கு கிரீன்லாந்து நிலையான நேரம்", "HAT": "நியூஃபவுண்ட்லாந்து பகலொளி நேரம்", "CLT": "சிலி நிலையான நேரம்", "TMT": "துர்க்மெனிஸ்தான் நிலையான நேரம்", "AEST": "ஆஸ்திரேலியன் கிழக்கத்திய நிலையான நேரம்", "AEDT": "ஆஸ்திரேலியன் கிழக்கத்திய பகலொளி நேரம்", "AKST": "அலாஸ்கா நிலையான நேரம்", "ACST": "ஆஸ்திரேலியன் மத்திய நிலையான நேரம்", "HEEG": "கிழக்கு கிரீன்லாந்து கோடை நேரம்", "HKT": "ஹாங்காங் நிலையான நேரம்", "LHST": "லார்ட் ஹோவ் நிலையான நேரம்", "HEPM": "செயின்ட் பியரி & மிக்குயிலான் பகலொளி நேரம்", "UYST": "உருகுவே கோடை நேரம்", "PST": "பசிபிக் நிலையான நேரம்", "∅∅∅": "அமேசான் கோடை நேரம்", "GMT": "கிரீன்விச் சராசரி நேரம்", "ChST": "சாமோரோ நிலையான நேரம்", "AWDT": "ஆஸ்திரேலியன் மேற்கத்திய பகலொளி நேரம்", "MDT": "மவுன்டைன் பகலொளி நேரம்", "NZDT": "நியூசிலாந்து பகலொளி நேரம்", "HAST": "ஹவாய்-அலேஷியன் நிலையான நேரம்", "ART": "அர்ஜென்டினா நிலையான நேரம்", "HNPM": "செயின்ட் பியரி & மிக்குயிலான் நிலையான நேரம்"},
	}
}

// Locale returns the current translators string locale
func (ta *ta_MY) Locale() string {
	return ta.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'ta_MY'
func (ta *ta_MY) PluralsCardinal() []locales.PluralRule {
	return ta.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'ta_MY'
func (ta *ta_MY) PluralsOrdinal() []locales.PluralRule {
	return ta.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'ta_MY'
func (ta *ta_MY) PluralsRange() []locales.PluralRule {
	return ta.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'ta_MY'
func (ta *ta_MY) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'ta_MY'
func (ta *ta_MY) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'ta_MY'
func (ta *ta_MY) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := ta.CardinalPluralRule(num1, v1)
	end := ta.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (ta *ta_MY) MonthAbbreviated(month time.Month) string {
	return ta.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (ta *ta_MY) MonthsAbbreviated() []string {
	return ta.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (ta *ta_MY) MonthNarrow(month time.Month) string {
	return ta.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (ta *ta_MY) MonthsNarrow() []string {
	return ta.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (ta *ta_MY) MonthWide(month time.Month) string {
	return ta.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (ta *ta_MY) MonthsWide() []string {
	return ta.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (ta *ta_MY) WeekdayAbbreviated(weekday time.Weekday) string {
	return ta.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (ta *ta_MY) WeekdaysAbbreviated() []string {
	return ta.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (ta *ta_MY) WeekdayNarrow(weekday time.Weekday) string {
	return ta.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (ta *ta_MY) WeekdaysNarrow() []string {
	return ta.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (ta *ta_MY) WeekdayShort(weekday time.Weekday) string {
	return ta.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (ta *ta_MY) WeekdaysShort() []string {
	return ta.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (ta *ta_MY) WeekdayWide(weekday time.Weekday) string {
	return ta.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (ta *ta_MY) WeekdaysWide() []string {
	return ta.daysWide
}

// Decimal returns the decimal point of number
func (ta *ta_MY) Decimal() string {
	return ta.decimal
}

// Group returns the group of number
func (ta *ta_MY) Group() string {
	return ta.group
}

// Group returns the minus sign of number
func (ta *ta_MY) Minus() string {
	return ta.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'ta_MY' and handles both Whole and Real numbers based on 'v'
func (ta *ta_MY) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ta.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ta.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ta.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'ta_MY' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (ta *ta_MY) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ta.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ta.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, ta.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'ta_MY'
func (ta *ta_MY) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ta.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ta.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ta.group[0])
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

	for j := len(ta.currencyPositivePrefix) - 1; j >= 0; j-- {
		b = append(b, ta.currencyPositivePrefix[j])
	}

	if num < 0 {
		b = append(b, ta.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ta.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'ta_MY'
// in accounting notation.
func (ta *ta_MY) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ta.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ta.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ta.group[0])
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

		for j := len(ta.currencyNegativePrefix) - 1; j >= 0; j-- {
			b = append(b, ta.currencyNegativePrefix[j])
		}

		b = append(b, ta.minus[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(ta.currencyPositivePrefix) - 1; j >= 0; j-- {
			b = append(b, ta.currencyPositivePrefix[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ta.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'ta_MY'
func (ta *ta_MY) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'ta_MY'
func (ta *ta_MY) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ta.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'ta_MY'
func (ta *ta_MY) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ta.monthsWide[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'ta_MY'
func (ta *ta_MY) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, ta.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ta.monthsWide[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'ta_MY'
func (ta *ta_MY) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 12 {
		b = append(b, ta.periodsAbbreviated[0]...)
	} else {
		b = append(b, ta.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ta.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'ta_MY'
func (ta *ta_MY) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 12 {
		b = append(b, ta.periodsAbbreviated[0]...)
	} else {
		b = append(b, ta.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ta.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ta.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'ta_MY'
func (ta *ta_MY) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 12 {
		b = append(b, ta.periodsAbbreviated[0]...)
	} else {
		b = append(b, ta.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ta.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ta.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'ta_MY'
func (ta *ta_MY) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 12 {
		b = append(b, ta.periodsAbbreviated[0]...)
	} else {
		b = append(b, ta.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ta.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ta.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := ta.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
