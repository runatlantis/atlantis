package ml_IN

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type ml_IN struct {
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

// New returns a new instance of translator for the 'ml_IN' locale
func New() locales.Translator {
	return &ml_IN{
		locale:                 "ml_IN",
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
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: ")",
		monthsAbbreviated:      []string{"", "ജനു", "ഫെബ്രു", "മാർ", "ഏപ്രി", "മേയ്", "ജൂൺ", "ജൂലൈ", "ഓഗ", "സെപ്റ്റം", "ഒക്ടോ", "നവം", "ഡിസം"},
		monthsNarrow:           []string{"", "ജ", "ഫെ", "മാ", "ഏ", "മെ", "ജൂൺ", "ജൂ", "ഓ", "സെ", "ഒ", "ന", "ഡി"},
		monthsWide:             []string{"", "ജനുവരി", "ഫെബ്രുവരി", "മാർച്ച്", "ഏപ്രിൽ", "മേയ്", "ജൂൺ", "ജൂലൈ", "ഓഗസ്റ്റ്", "സെപ്റ്റംബർ", "ഒക്\u200cടോബർ", "നവംബർ", "ഡിസംബർ"},
		daysAbbreviated:        []string{"ഞായർ", "തിങ്കൾ", "ചൊവ്വ", "ബുധൻ", "വ്യാഴം", "വെള്ളി", "ശനി"},
		daysNarrow:             []string{"ഞ", "തി", "ചൊ", "ബു", "വ്യാ", "വെ", "ശ"},
		daysShort:              []string{"ഞാ", "തി", "ചൊ", "ബു", "വ്യാ", "വെ", "ശ"},
		daysWide:               []string{"ഞായറാഴ്\u200cച", "തിങ്കളാഴ്\u200cച", "ചൊവ്വാഴ്ച", "ബുധനാഴ്\u200cച", "വ്യാഴാഴ്\u200cച", "വെള്ളിയാഴ്\u200cച", "ശനിയാഴ്\u200cച"},
		periodsAbbreviated:     []string{"AM", "PM"},
		periodsNarrow:          []string{"AM", "PM"},
		periodsWide:            []string{"AM", "PM"},
		erasAbbreviated:        []string{"ക്രി.മു.", "എഡി"},
		erasNarrow:             []string{"ക്രി.മു.", "എഡി"},
		erasWide:               []string{"ക്രിസ്\u200cതുവിന് മുമ്പ്", "ആന്നോ ഡൊമിനി"},
		timezones:              map[string]string{"ADT": "അറ്റ്\u200cലാന്റിക് ഡേലൈറ്റ് സമയം", "AEST": "ഓസ്\u200cട്രേലിയൻ കിഴക്കൻ സ്റ്റാൻഡേർഡ് സമയം", "MST": "വടക്കെ അമേരിക്കൻ മൗണ്ടൻ സ്റ്റാൻഡേർഡ് സമയം", "MDT": "വടക്കെ അമേരിക്കൻ മൗണ്ടൻ ഡേലൈറ്റ് സമയം", "NZST": "ന്യൂസിലാൻഡ് സ്റ്റാൻഡേർഡ് സമയം", "NZDT": "ന്യൂസിലാൻഡ് ഡേലൈറ്റ് സമയം", "EAT": "കിഴക്കൻ ആഫ്രിക്ക സമയം", "ART": "അർജന്റീന സ്റ്റാൻഡേർഡ് സമയം", "PST": "വടക്കെ അമേരിക്കൻ പസഫിക് സ്റ്റാൻഡേർഡ് സമയം", "JST": "ജപ്പാൻ സ്റ്റാൻഡേർഡ് സമയം", "MYT": "മലേഷ്യ സമയം", "GFT": "ഫ്രഞ്ച് ഗയാന സമയം", "JDT": "ജപ്പാൻ ഡേലൈറ്റ് സമയം", "VET": "വെനിസ്വേല സമയം", "WIB": "പടിഞ്ഞാറൻ ഇന്തോനേഷ്യ സമയം", "WAT": "പടിഞ്ഞാറൻ ആഫ്രിക്ക സ്റ്റാൻഡേർഡ് സമയം", "HKST": "ഹോങ്കോങ്ങ് ഗ്രീഷ്\u200cമകാല സമയം", "OEZ": "കിഴക്കൻ യൂറോപ്യൻ സ്റ്റാൻഡേർഡ് സമയം", "HNCU": "ക്യൂബ സ്റ്റാൻഡേർഡ് സമയം", "AWDT": "ഓസ്\u200cട്രേലിയൻ പടിഞ്ഞാറൻ ഡേലൈറ്റ് സമയം", "∅∅∅": "എയ്ക്കർ വേനൽക്കാല സമയം", "CHAST": "ചാത്തം സ്റ്റാൻഡേർഡ് സമയം", "WESZ": "പടിഞ്ഞാറൻ യൂറോപ്യൻ ഗ്രീഷ്\u200cമകാല സമയം", "ARST": "അർജന്റീന ഗ്രീഷ്\u200cമകാല സമയം", "ChST": "ചമോറോ സ്റ്റാൻഡേർഡ് സമയം", "WIT": "കിഴക്കൻ ഇന്തോനേഷ്യ സമയം", "CAT": "മധ്യ ആഫ്രിക്ക സമയം", "MEZ": "സെൻട്രൽ യൂറോപ്യൻ സ്റ്റാൻഡേർഡ് സമയം", "WARST": "പടിഞ്ഞാറൻ അർജന്റീന ഗ്രീഷ്\u200cമകാല സമയം", "HAT": "ന്യൂഫൗണ്ട്\u200cലാന്റ് ഡേലൈറ്റ് സമയം", "HNNOMX": "വടക്കുപടിഞ്ഞാറൻ മെക്\u200cസിക്കൻ സ്റ്റാൻഡേർഡ് സമയം", "SRT": "സുരിനെയിം സമയം", "CLT": "ചിലി സ്റ്റാൻഡേർഡ് സമയം", "SAST": "ദക്ഷിണാഫ്രിക്ക സ്റ്റാൻഡേർഡ് സമയം", "EDT": "വടക്കെ അമേരിക്കൻ കിഴക്കൻ ഡേലൈറ്റ് സമയം", "CHADT": "ചാത്തം ഗ്രീഷ്\u200cമകാല സമയം", "HECU": "ക്യൂബ ഡേലൈറ്റ് സമയം", "BT": "ഭൂട്ടാൻ സമയം", "AKST": "അലാസ്ക സ്റ്റാൻഡേർഡ് സമയം", "UYT": "ഉറുഗ്വേ സ്റ്റാൻഡേർഡ് സമയം", "AWST": "ഓസ്\u200cട്രേലിയൻ പടിഞ്ഞാറൻ സ്റ്റാൻഡേർഡ് സമയം", "AEDT": "ഓസ്\u200cട്രേലിയൻ കിഴക്കൻ ഡേലൈറ്റ് സമയം", "IST": "ഇന്ത്യൻ സ്റ്റാൻഡേർഡ് സമയം", "HEPM": "സെന്റ് പിയറി ആൻഡ് മിക്വലൻ ഡേലൈറ്റ് സമയം", "HENOMX": "വടക്കുപടിഞ്ഞാറൻ മെക്സിക്കൻ ഡേലൈറ്റ് സമയം", "TMST": "തുർക്ക്\u200cമെനിസ്ഥാൻ ഗ്രീഷ്\u200cമകാല സമയം", "UYST": "ഉറുഗ്വേ ഗ്രീഷ്\u200cമകാല സമയം", "GMT": "ഗ്രീൻവിച്ച് മീൻ സമയം", "AKDT": "അലാസ്\u200cക ഡേലൈറ്റ് സമയം", "ACST": "ഓസ്ട്രേലിയൻ സെൻട്രൽ സ്റ്റാൻഡേർഡ് സമയം", "ACWST": "ഓസ്ട്രേലിയൻ സെൻട്രൽ പടിഞ്ഞാറൻ സ്റ്റാൻഡേർഡ് സമയം", "COST": "കൊളംബിയ ഗ്രീഷ്\u200cമകാല സമയം", "HNPMX": "മെക്\u200cസിക്കൻ പസഫിക് സ്റ്റാൻഡേർഡ് സമയം", "WAST": "പടിഞ്ഞാറൻ ആഫ്രിക്ക ഗ്രീഷ്\u200cമകാല സമയം", "ECT": "ഇക്വഡോർ സമയം", "SGT": "സിംഗപ്പൂർ സ്റ്റാൻഡേർഡ് സമയം", "ACDT": "ഓസ്ട്രേലിയൻ സെൻട്രൽ ഡേലൈറ്റ് സമയം", "WITA": "മധ്യ ഇന്തോനേഷ്യ സമയം", "AST": "അറ്റ്\u200cലാന്റിക് സ്റ്റാൻഡേർഡ് സമയം", "WEZ": "പടിഞ്ഞാറൻ യൂറോപ്യൻ സ്റ്റാൻഡേർഡ് സമയം", "HEOG": "പടിഞ്ഞാറൻ ഗ്രീൻലാൻഡ് ഗ്രീഷ്\u200cമകാല സമയം", "LHST": "ലോർഡ് ഹോവ് സ്റ്റാൻഡേർഡ് സമയം", "CLST": "ചിലി ഗ്രീഷ്\u200cമകാല സമയം", "HADT": "ഹവായ്-അലൂഷ്യൻ ഡേലൈറ്റ് സമയം", "GYT": "ഗയാന സമയം", "PDT": "വടക്കെ അമേരിക്കൻ പസഫിക് ഡേലൈറ്റ് സമയം", "HNOG": "പടിഞ്ഞാറൻ ഗ്രീൻലാൻഡ് സ്റ്റാൻഡേർഡ് സമയം", "MESZ": "സെൻട്രൽ യൂറോപ്യൻ ഗ്രീഷ്മകാല സമയം", "WART": "പടിഞ്ഞാറൻ അർജന്റീന സ്റ്റാൻഡേർഡ് സമയം", "HNPM": "സെന്റ് പിയറി ആൻഡ് മിക്വലൻ സ്റ്റാൻഡേർഡ് സമയം", "HNT": "ന്യൂഫൗണ്ട്\u200cലാന്റ് സ്റ്റാൻഡേർഡ് സമയം", "TMT": "തുർക്ക്\u200cമെനിസ്ഥാൻ സ്റ്റാൻഡേർഡ് സമയം", "HAST": "ഹവായ്-അലൂഷ്യൻ സ്റ്റാൻഡേർഡ് സമയം", "CST": "വടക്കെ അമേരിക്കൻ സെൻട്രൽ സ്റ്റാൻഡേർഡ് സമയം", "HNEG": "കിഴക്കൻ ഗ്രീൻലാൻഡ് സ്റ്റാൻഡേർഡ് സമയം", "HEEG": "കിഴക്കൻ ഗ്രീൻലാൻഡ് ഗ്രീഷ്\u200cമകാല സമയം", "EST": "വടക്കെ അമേരിക്കൻ കിഴക്കൻ സ്റ്റാൻഡേർഡ് സമയം", "LHDT": "ലോർഡ് ഹോവ് ഡേലൈറ്റ് സമയം", "COT": "കൊളംബിയ സ്റ്റാൻഡേർഡ് സമയം", "OESZ": "കിഴക്കൻ യൂറോപ്യൻ ഗ്രീഷ്മകാല സമയം", "CDT": "വടക്കെ അമേരിക്കൻ സെൻട്രൽ ഡേലൈറ്റ് സമയം", "ACWDT": "ഓസ്ട്രേലിയൻ സെൻട്രൽ പടിഞ്ഞാറൻ ഡേലൈറ്റ് സമയം", "HKT": "ഹോങ്കോങ്ങ് സ്റ്റാൻഡേർഡ് സമയം", "HEPMX": "മെക്സിക്കൻ പസഫിക് ഡേലൈറ്റ് സമയം", "BOT": "ബൊളീവിയ സമയം"},
	}
}

// Locale returns the current translators string locale
func (ml *ml_IN) Locale() string {
	return ml.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'ml_IN'
func (ml *ml_IN) PluralsCardinal() []locales.PluralRule {
	return ml.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'ml_IN'
func (ml *ml_IN) PluralsOrdinal() []locales.PluralRule {
	return ml.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'ml_IN'
func (ml *ml_IN) PluralsRange() []locales.PluralRule {
	return ml.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'ml_IN'
func (ml *ml_IN) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'ml_IN'
func (ml *ml_IN) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'ml_IN'
func (ml *ml_IN) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := ml.CardinalPluralRule(num1, v1)
	end := ml.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (ml *ml_IN) MonthAbbreviated(month time.Month) string {
	return ml.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (ml *ml_IN) MonthsAbbreviated() []string {
	return ml.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (ml *ml_IN) MonthNarrow(month time.Month) string {
	return ml.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (ml *ml_IN) MonthsNarrow() []string {
	return ml.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (ml *ml_IN) MonthWide(month time.Month) string {
	return ml.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (ml *ml_IN) MonthsWide() []string {
	return ml.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (ml *ml_IN) WeekdayAbbreviated(weekday time.Weekday) string {
	return ml.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (ml *ml_IN) WeekdaysAbbreviated() []string {
	return ml.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (ml *ml_IN) WeekdayNarrow(weekday time.Weekday) string {
	return ml.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (ml *ml_IN) WeekdaysNarrow() []string {
	return ml.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (ml *ml_IN) WeekdayShort(weekday time.Weekday) string {
	return ml.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (ml *ml_IN) WeekdaysShort() []string {
	return ml.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (ml *ml_IN) WeekdayWide(weekday time.Weekday) string {
	return ml.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (ml *ml_IN) WeekdaysWide() []string {
	return ml.daysWide
}

// Decimal returns the decimal point of number
func (ml *ml_IN) Decimal() string {
	return ml.decimal
}

// Group returns the group of number
func (ml *ml_IN) Group() string {
	return ml.group
}

// Group returns the minus sign of number
func (ml *ml_IN) Minus() string {
	return ml.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'ml_IN' and handles both Whole and Real numbers based on 'v'
func (ml *ml_IN) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	inSecondary := false
	groupThreshold := 3

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ml.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {

			if count == groupThreshold {
				b = append(b, ml.group[0])
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
		b = append(b, ml.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'ml_IN' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (ml *ml_IN) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ml.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ml.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, ml.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'ml_IN'
func (ml *ml_IN) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ml.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ml.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ml.group[0])
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
		b = append(b, ml.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ml.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'ml_IN'
// in accounting notation.
func (ml *ml_IN) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ml.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ml.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ml.group[0])
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

		b = append(b, ml.currencyNegativePrefix[0])

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
			b = append(b, ml.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, ml.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'ml_IN'
func (ml *ml_IN) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'ml_IN'
func (ml *ml_IN) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, ml.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'ml_IN'
func (ml *ml_IN) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, ml.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'ml_IN'
func (ml *ml_IN) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, ml.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, ml.daysWide[t.Weekday()]...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'ml_IN'
func (ml *ml_IN) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ml.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ml.periodsAbbreviated[0]...)
	} else {
		b = append(b, ml.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'ml_IN'
func (ml *ml_IN) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ml.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ml.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ml.periodsAbbreviated[0]...)
	} else {
		b = append(b, ml.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'ml_IN'
func (ml *ml_IN) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ml.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ml.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ml.periodsAbbreviated[0]...)
	} else {
		b = append(b, ml.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'ml_IN'
func (ml *ml_IN) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ml.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ml.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ml.periodsAbbreviated[0]...)
	} else {
		b = append(b, ml.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := ml.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
