package te

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type te struct {
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

// New returns a new instance of translator for the 'te' locale
func New() locales.Translator {
	return &te{
		locale:                 "te",
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
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "A$", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "₪", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JP¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "₩", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZ$", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "฿", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "NT$", "TZS", "UAH", "UAK", "UGS", "UGX", "$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "₫", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "EC$", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: ")",
		monthsAbbreviated:      []string{"", "జన", "ఫిబ్ర", "మార్చి", "ఏప్రి", "మే", "జూన్", "జులై", "ఆగ", "సెప్టెం", "అక్టో", "నవం", "డిసెం"},
		monthsNarrow:           []string{"", "జ", "ఫి", "మా", "ఏ", "మే", "జూ", "జు", "ఆ", "సె", "అ", "న", "డి"},
		monthsWide:             []string{"", "జనవరి", "ఫిబ్రవరి", "మార్చి", "ఏప్రిల్", "మే", "జూన్", "జులై", "ఆగస్టు", "సెప్టెంబర్", "అక్టోబర్", "నవంబర్", "డిసెంబర్"},
		daysAbbreviated:        []string{"ఆది", "సోమ", "మంగళ", "బుధ", "గురు", "శుక్ర", "శని"},
		daysNarrow:             []string{"ఆ", "సో", "మ", "బు", "గు", "శు", "శ"},
		daysShort:              []string{"ఆది", "సోమ", "మం", "బుధ", "గురు", "శుక్ర", "శని"},
		daysWide:               []string{"ఆదివారం", "సోమవారం", "మంగళవారం", "బుధవారం", "గురువారం", "శుక్రవారం", "శనివారం"},
		periodsAbbreviated:     []string{"AM", "PM"},
		periodsNarrow:          []string{"ఉ", "సా"},
		periodsWide:            []string{"AM", "PM"},
		erasAbbreviated:        []string{"క్రీపూ", "క్రీశ"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"క్రీస్తు పూర్వం", "క్రీస్తు శకం"},
		timezones:              map[string]string{"CHAST": "చాథమ్ ప్రామాణిక సమయం", "SAST": "దక్షిణ ఆఫ్రికా ప్రామాణిక సమయం", "BOT": "బొలీవియా సమయం", "JST": "జపాన్ ప్రామాణిక సమయం", "EST": "తూర్పు ప్రామాణిక సమయం", "IST": "భారతదేశ సమయం", "COST": "కొలంబియా వేసవి సమయం", "UYT": "ఉరుగ్వే ప్రామాణిక సమయం", "HEPM": "సెయింట్ పియర్ మరియు మిక్వెలాన్ పగటి వెలుతురు సమయం", "CST": "మధ్యమ ప్రామాణిక సమయం", "WIB": "పశ్చిమ ఇండోనేషియా సమయం", "ECT": "ఈక్వడార్ సమయం", "HKST": "హాంకాంగ్ వేసవి సమయం", "∅∅∅": "అమెజాన్ వేసవి సమయం", "GMT": "గ్రీన్\u200cవిచ్ సగటు సమయం", "BT": "భూటాన్ సమయం", "NZDT": "న్యూజిల్యాండ్ పగటి వెలుతురు సమయం", "HNOG": "పశ్చిమ గ్రీన్\u200cల్యాండ్ ప్రామాణిక సమయం", "AWST": "ఆస్ట్రేలియన్ పశ్చిమ ప్రామాణిక సమయం", "MYT": "మలేషియా సమయం", "HAT": "న్యూఫౌండ్\u200cల్యాండ్ పగటి వెలుతురు సమయం", "VET": "వెనిజులా సమయం", "MDT": "మకావ్ వేసవి సమయం", "ACWDT": "ఆస్ట్రేలియా మధ్యమ పశ్చిమ పగటి వెలుతురు సమయం", "AEST": "ఆస్ట్రేలియన్ తూర్పు ప్రామాణిక సమయం", "ACST": "ఆస్ట్రేలియా మధ్యమ ప్రామాణిక సమయం", "WIT": "తూర్పు ఇండోనేషియా సమయం", "AWDT": "ఆస్ట్రేలియన్ పశ్చిమ పగటి వెలుతురు సమయం", "HNEG": "తూర్పు గ్రీన్\u200cల్యాండ్ ప్రామాణిక సమయం", "WARST": "పశ్చిమ అర్జెంటీనా వేసవి సమయం", "ChST": "చామర్రో ప్రామాణిక సమయం", "PDT": "పసిఫిక్ పగటి వెలుతురు సమయం", "HNPMX": "మెక్సికన్ పసిఫిక్ ప్రామాణిక సమయం", "ACDT": "ఆస్ట్రేలియా మధ్యమ పగటి వెలుతురు సమయం", "WART": "పశ్చిమ అర్జెంటీనా ప్రామాణిక సమయం", "HNPM": "సెయింట్ పియెర్ మరియు మిక్వెలాన్ ప్రామాణిక సమయం", "CLT": "చిలీ ప్రామాణిక సమయం", "OESZ": "తూర్పు యూరోపియన్ వేసవి సమయం", "CDT": "మధ్యమ పగటి వెలుతురు సమయం", "PST": "పసిఫిక్ ప్రామాణిక సమయం", "TMT": "తుర్క్\u200cమెనిస్తాన్ ప్రామాణిక సమయం", "HECU": "క్యూబా పగటి వెలుతురు సమయం", "EDT": "తూర్పు పగటి వెలుతురు సమయం", "ART": "అర్జెంటీనా ప్రామాణిక సమయం", "JDT": "జపాన్ పగటి వెలుతురు సమయం", "HEOG": "పశ్చిమ గ్రీన్\u200cల్యాండ్ వేసవి సమయం", "MESZ": "సెంట్రల్ యూరోపియన్ వేసవి సమయం", "ADT": "అట్లాంటిక్ పగటి వెలుతురు సమయం", "GFT": "ఫ్రెంచ్ గయానా సమయం", "AKST": "అలాస్కా ప్రామాణిక సమయం", "CLST": "చిలీ వేసవి సమయం", "ARST": "ఆర్జెంటీనా వేసవి సమయం", "ACWST": "మధ్యమ ఆస్ట్రేలియన్ పశ్చిమ ప్రామాణిక సమయం", "LHST": "లార్డ్ హోవ్ ప్రామాణిక సమయం", "WITA": "సెంట్రల్ ఇండోనేషియా సమయం", "HADT": "హవాయ్-అల్యూషియన్ పగటి వెలుతురు సమయం", "WAT": "పశ్చిమ ఆఫ్రికా ప్రామాణిక సమయం", "COT": "కొలంబియా ప్రామాణిక సమయం", "HAST": "హవాయ్-అల్యూషియన్ ప్రామాణిక సమయం", "AEDT": "ఆస్ట్రేలియన్ తూర్పు పగటి వెలుతురు సమయం", "HKT": "హాంకాంగ్ ప్రామాణిక సమయం", "HNT": "న్యూఫౌండ్\u200cల్యాండ్ ప్రామాణిక సమయం", "HNNOMX": "వాయువ్య మెక్సికో ప్రామాణిక సమయం", "MST": "మకావ్ ప్రామాణిక సమయం", "CAT": "సెంట్రల్ ఆఫ్రికా సమయం", "CHADT": "చాథమ్ పగటి వెలుతురు సమయం", "HNCU": "క్యూబా ప్రామాణిక సమయం", "AST": "అట్లాంటిక్ ప్రామాణిక సమయం", "WAST": "పశ్చిమ ఆఫ్రికా వేసవి సమయం", "AKDT": "అలాస్కా పగటి వెలుతురు సమయం", "SGT": "సింగపూర్ ప్రామాణిక సమయం", "UYST": "ఉరుగ్వే వేసవి సమయం", "GYT": "గయానా సమయం", "WESZ": "పశ్చిమ యూరోపియన్ వేసవి సమయం", "LHDT": "లార్డ్ హోవ్ పగటి సమయం", "SRT": "సూరినామ్ సమయం", "OEZ": "తూర్పు యూరోపియన్ ప్రామాణిక సమయం", "HEPMX": "మెక్సికన్ పసిఫిక్ పగటి వెలుతురు సమయం", "WEZ": "పశ్చిమ యూరోపియన్ ప్రామాణిక సమయం", "NZST": "న్యూజిల్యాండ్ ప్రామాణిక సమయం", "HEEG": "తూర్పు గ్రీన్\u200cల్యాండ్ వేసవి సమయం", "MEZ": "సెంట్రల్ యూరోపియన్ ప్రామాణిక సమయం", "HENOMX": "వాయువ్య మెక్సికో పగటి వెలుతురు సమయం", "TMST": "తుర్క్\u200cమెనిస్తాన్ వేసవి సమయం", "EAT": "తూర్పు ఆఫ్రికా సమయం"},
	}
}

// Locale returns the current translators string locale
func (te *te) Locale() string {
	return te.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'te'
func (te *te) PluralsCardinal() []locales.PluralRule {
	return te.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'te'
func (te *te) PluralsOrdinal() []locales.PluralRule {
	return te.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'te'
func (te *te) PluralsRange() []locales.PluralRule {
	return te.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'te'
func (te *te) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'te'
func (te *te) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'te'
func (te *te) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := te.CardinalPluralRule(num1, v1)
	end := te.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (te *te) MonthAbbreviated(month time.Month) string {
	return te.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (te *te) MonthsAbbreviated() []string {
	return te.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (te *te) MonthNarrow(month time.Month) string {
	return te.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (te *te) MonthsNarrow() []string {
	return te.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (te *te) MonthWide(month time.Month) string {
	return te.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (te *te) MonthsWide() []string {
	return te.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (te *te) WeekdayAbbreviated(weekday time.Weekday) string {
	return te.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (te *te) WeekdaysAbbreviated() []string {
	return te.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (te *te) WeekdayNarrow(weekday time.Weekday) string {
	return te.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (te *te) WeekdaysNarrow() []string {
	return te.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (te *te) WeekdayShort(weekday time.Weekday) string {
	return te.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (te *te) WeekdaysShort() []string {
	return te.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (te *te) WeekdayWide(weekday time.Weekday) string {
	return te.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (te *te) WeekdaysWide() []string {
	return te.daysWide
}

// Decimal returns the decimal point of number
func (te *te) Decimal() string {
	return te.decimal
}

// Group returns the group of number
func (te *te) Group() string {
	return te.group
}

// Group returns the minus sign of number
func (te *te) Minus() string {
	return te.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'te' and handles both Whole and Real numbers based on 'v'
func (te *te) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	inSecondary := false
	groupThreshold := 3

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, te.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {

			if count == groupThreshold {
				b = append(b, te.group[0])
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
		b = append(b, te.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'te' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (te *te) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, te.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, te.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, te.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'te'
func (te *te) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := te.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	inSecondary := false
	groupThreshold := 3

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, te.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {

			if count == groupThreshold {
				b = append(b, te.group[0])
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

	if num < 0 {
		b = append(b, te.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, te.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'te'
// in accounting notation.
func (te *te) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := te.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	inSecondary := false
	groupThreshold := 3

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, te.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {

			if count == groupThreshold {
				b = append(b, te.group[0])
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

		b = append(b, te.currencyNegativePrefix[0])

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
			b = append(b, te.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, te.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'te'
func (te *te) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2d}...)

	if t.Month() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Month()), 10)

	b = append(b, []byte{0x2d}...)

	if t.Year() > 9 {
		b = append(b, strconv.Itoa(t.Year())[2:]...)
	} else {
		b = append(b, strconv.Itoa(t.Year())[1:]...)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'te'
func (te *te) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, te.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'te'
func (te *te) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, te.monthsWide[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'te'
func (te *te) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, te.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, te.daysWide[t.Weekday()]...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'te'
func (te *te) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, te.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, te.periodsAbbreviated[0]...)
	} else {
		b = append(b, te.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'te'
func (te *te) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, te.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, te.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, te.periodsAbbreviated[0]...)
	} else {
		b = append(b, te.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'te'
func (te *te) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, te.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, te.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, te.periodsAbbreviated[0]...)
	} else {
		b = append(b, te.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'te'
func (te *te) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, te.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, te.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, te.periodsAbbreviated[0]...)
	} else {
		b = append(b, te.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := te.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
