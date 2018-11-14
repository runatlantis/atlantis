package pa

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type pa struct {
	locale             string
	pluralsCardinal    []locales.PluralRule
	pluralsOrdinal     []locales.PluralRule
	pluralsRange       []locales.PluralRule
	decimal            string
	group              string
	minus              string
	percent            string
	perMille           string
	timeSeparator      string
	inifinity          string
	currencies         []string // idx = enum of currency code
	monthsAbbreviated  []string
	monthsNarrow       []string
	monthsWide         []string
	daysAbbreviated    []string
	daysNarrow         []string
	daysShort          []string
	daysWide           []string
	periodsAbbreviated []string
	periodsNarrow      []string
	periodsShort       []string
	periodsWide        []string
	erasAbbreviated    []string
	erasNarrow         []string
	erasWide           []string
	timezones          map[string]string
}

// New returns a new instance of translator for the 'pa' locale
func New() locales.Translator {
	return &pa{
		locale:             "pa",
		pluralsCardinal:    []locales.PluralRule{2, 6},
		pluralsOrdinal:     []locales.PluralRule{6},
		pluralsRange:       []locales.PluralRule{2, 6},
		decimal:            ".",
		group:              ",",
		minus:              "-",
		percent:            "%",
		perMille:           "‰",
		timeSeparator:      ":",
		inifinity:          "∞",
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "A$", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "₪", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JP¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "₩", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZ$", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "р.", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "฿", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "NT$", "TZS", "UAH", "UAK", "UGS", "UGX", "US$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "₫", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "EC$", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		monthsAbbreviated:  []string{"", "ਜਨ", "ਫ਼ਰ", "ਮਾਰਚ", "ਅਪ੍ਰੈ", "ਮਈ", "ਜੂਨ", "ਜੁਲਾ", "ਅਗ", "ਸਤੰ", "ਅਕਤੂ", "ਨਵੰ", "ਦਸੰ"},
		monthsNarrow:       []string{"", "ਜ", "ਫ਼", "ਮਾ", "ਅ", "ਮ", "ਜੂ", "ਜੁ", "ਅ", "ਸ", "ਅ", "ਨ", "ਦ"},
		monthsWide:         []string{"", "ਜਨਵਰੀ", "ਫ਼ਰਵਰੀ", "ਮਾਰਚ", "ਅਪ੍ਰੈਲ", "ਮਈ", "ਜੂਨ", "ਜੁਲਾਈ", "ਅਗਸਤ", "ਸਤੰਬਰ", "ਅਕਤੂਬਰ", "ਨਵੰਬਰ", "ਦਸੰਬਰ"},
		daysAbbreviated:    []string{"ਐਤ", "ਸੋਮ", "ਮੰਗਲ", "ਬੁੱਧ", "ਵੀਰ", "ਸ਼ੁੱਕਰ", "ਸ਼ਨਿੱਚਰ"},
		daysNarrow:         []string{"ਐ", "ਸੋ", "ਮੰ", "ਬੁੱ", "ਵੀ", "ਸ਼ੁੱ", "ਸ਼"},
		daysShort:          []string{"ਐਤ", "ਸੋਮ", "ਮੰਗ", "ਬੁੱਧ", "ਵੀਰ", "ਸ਼ੁੱਕ", "ਸ਼ਨਿੱ"},
		daysWide:           []string{"ਐਤਵਾਰ", "ਸੋਮਵਾਰ", "ਮੰਗਲਵਾਰ", "ਬੁੱਧਵਾਰ", "ਵੀਰਵਾਰ", "ਸ਼ੁੱਕਰਵਾਰ", "ਸ਼ਨਿੱਚਰਵਾਰ"},
		periodsAbbreviated: []string{"ਪੂ.ਦੁ.", "ਬਾ.ਦੁ."},
		periodsNarrow:      []string{"ਸ.", "ਸ਼."},
		periodsWide:        []string{"ਪੂ.ਦੁ.", "ਬਾ.ਦੁ."},
		erasAbbreviated:    []string{"ਈ. ਪੂ.", "ਸੰਨ"},
		erasNarrow:         []string{"ਈ.ਪੂ.", "ਸੰਨ"},
		erasWide:           []string{"ਈਸਵੀ ਪੂਰਵ", "ਈਸਵੀ ਸੰਨ"},
		timezones:          map[string]string{"∅∅∅": "ਬ੍ਰਾਜ਼ੀਲੀਆ ਗਰਮੀਆਂ ਦਾ ਵੇਲਾ", "HNEG": "ਪੂਰਬੀ ਗ੍ਰੀਨਲੈਂਡ ਮਿਆਰੀ ਵੇਲਾ", "HEPM": "ਸੈਂਟ ਪੀਅਰੇ ਅਤੇ ਮਿਕੇਲਨ ਪ੍ਰਕਾਸ਼ ਵੇਲਾ", "WITA": "ਮੱਧ ਇੰਡੋਨੇਸ਼ੀਆਈ ਵੇਲਾ", "TMT": "ਤੁਰਕਮੇਨਿਸਤਾਨ ਮਿਆਰੀ ਵੇਲਾ", "HEPMX": "ਮੈਕਸੀਕਨ ਪੈਸਿਫਿਕ ਪ੍ਰਕਾਸ਼ ਵੇਲਾ", "HEOG": "ਪੱਛਮੀ ਗ੍ਰੀਨਲੈਂਡ ਗਰਮੀਆਂ ਦਾ ਵੇਲਾ", "HNNOMX": "ਉੱਤਰ ਪੱਛਮੀ ਮੈਕਸੀਕੋ ਮਿਆਰੀ ਵੇਲਾ", "COT": "ਕੋਲੰਬੀਆ ਮਿਆਰੀ ਵੇਲਾ", "UYST": "ਉਰੂਗਵੇ ਗਰਮੀਆਂ ਦਾ ਵੇਲਾ", "ADT": "ਅਟਲਾਂਟਿਕ ਪ੍ਰਕਾਸ਼ ਵੇਲਾ", "HNPM": "ਸੈਂਟ ਪੀਅਰੇ ਅਤੇ ਮਿਕੇਲਨ ਮਿਆਰੀ ਵੇਲਾ", "EDT": "ਉੱਤਰੀ ਅਮਰੀਕੀ ਪੂਰਬੀ ਪ੍ਰਕਾਸ਼ ਵੇਲਾ", "HKT": "ਹਾਂਗ ਕਾਂਗ ਮਿਆਰੀ ਵੇਲਾ", "OEZ": "ਪੂਰਬੀ ਯੂਰਪੀ ਮਿਆਰੀ ਵੇਲਾ", "ChST": "ਚਾਮੋਰੋ ਮਿਆਰੀ ਵੇਲਾ", "BT": "ਭੂਟਾਨ ਵੇਲਾ", "ACDT": "ਆਸਟ੍ਰੇਲੀਆਈ ਕੇਂਦਰੀ ਪ੍ਰਕਾਸ਼ ਵੇਲਾ", "HKST": "ਹਾਂਗ ਕਾਂਗ ਗਰਮੀਆਂ ਦਾ ਵੇਲਾ", "UYT": "ਉਰੂਗਵੇ ਮਿਆਰੀ ਵੇਲਾ", "WESZ": "ਪੱਛਮੀ ਯੂਰਪੀ ਗਰਮੀਆਂ ਦਾ ਵੇਲਾ", "MYT": "ਮਲੇਸ਼ੀਆ ਵੇਲਾ", "TMST": "ਤੁਰਕਮੇਨਿਸਤਾਨ ਗਰਮੀਆਂ ਦਾ ਵੇਲਾ", "WEZ": "ਪੱਛਮੀ ਯੂਰਪੀ ਮਿਆਰੀ ਵੇਲਾ", "EST": "ਉੱਤਰੀ ਅਮਰੀਕੀ ਪੂਰਬੀ ਮਿਆਰੀ ਵੇਲਾ", "CST": "ਉੱਤਰੀ ਅਮਰੀਕੀ ਕੇਂਦਰੀ ਮਿਆਰੀ ਵੇਲਾ", "HNT": "ਨਿਊਫਾਉਂਡਲੈਂਡ ਮਿਆਰੀ ਵੇਲਾ", "AWST": "ਆਸਟ੍ਰੇਲੀਆਈ ਪੱਛਮੀ ਮਿਆਰੀ ਵੇਲਾ", "MST": "ਉੱਤਰੀ ਅਮਰੀਕੀ ਮਾਉਂਟੇਨ ਮਿਆਰੀ ਵੇਲਾ", "GFT": "ਫ੍ਰੈਂਚ ਗੁਏਨਾ ਵੇਲਾ", "ECT": "ਇਕਵੇਡੋਰ ਵੇਲਾ", "LHST": "ਲੌਰਡ ਹੋਵੇ ਮਿਆਰੀ ਵੇਲਾ", "VET": "ਵੈਨੇਜ਼ੂਏਲਾ ਵੇਲਾ", "PST": "ਉੱਤਰੀ ਅਮਰੀਕੀ ਪੈਸਿਫਿਕ ਮਿਆਰੀ ਵੇਲਾ", "PDT": "ਉੱਤਰੀ ਅਮਰੀਕੀ ਪੈਸਿਫਿਕ ਪ੍ਰਕਾਸ਼ ਵੇਲਾ", "MDT": "ਉੱਤਰੀ ਅਮਰੀਕੀ ਮਾਉਂਟੇਨ ਪ੍ਰਕਾਸ਼ ਵੇਲਾ", "WIB": "ਪੱਛਮੀ ਇੰਡੋਨੇਸ਼ੀਆ ਵੇਲਾ", "CLT": "ਚਿਲੀ ਮਿਆਰੀ ਵੇਲਾ", "WIT": "ਪੂਰਬੀ ਇੰਡੋਨੇਸ਼ੀਆ ਵੇਲਾ", "OESZ": "ਪੂਰਬੀ ਯੂਰਪੀ ਗਰਮੀਆਂ ਦਾ ਵੇਲਾ", "ARST": "ਅਰਜਨਟੀਨਾ ਗਰਮੀਆਂ ਦਾ ਵੇਲਾ", "CHAST": "ਚੈਥਮ ਮਿਆਰੀ ਵੇਲਾ", "CDT": "ਉੱਤਰੀ ਅਮਰੀਕੀ ਕੇਂਦਰੀ ਪ੍ਰਕਾਸ਼ ਵੇਲਾ", "NZDT": "ਨਿਊਜ਼ੀਲੈਂਡ ਪ੍ਰਕਾਸ਼ ਵੇਲਾ", "HNOG": "ਪੱਛਮੀ ਗ੍ਰੀਨਲੈਂਡ ਮਿਆਰੀ ਵੇਲਾ", "ACWST": "ਆਸਟ੍ਰੇਲੀਆਈ ਕੇਂਦਰੀ ਪੱਛਮੀ ਮਿਆਰੀ ਵੇਲਾ", "WARST": "ਪੱਛਮੀ ਅਰਜਨਟੀਨਾ ਗਰਮੀਆਂ ਦਾ ਵੇਲਾ", "HAT": "ਨਿਊਫਾਉਂਡਲੈਂਡ ਪ੍ਰਕਾਸ਼ ਵੇਲਾ", "HENOMX": "ਉੱਤਰ ਪੱਛਮੀ ਮੈਕਸੀਕੋ ਪ੍ਰਕਾਸ਼ ਵੇਲਾ", "SAST": "ਦੱਖਣੀ ਅਫ਼ਰੀਕਾ ਮਿਆਰੀ ਵੇਲਾ", "WAT": "ਪੱਛਮੀ ਅਫਰੀਕਾ ਮਿਆਰੀ ਵੇਲਾ", "GYT": "ਗੁਯਾਨਾ ਵੇਲਾ", "NZST": "ਨਿਊਜ਼ੀਲੈਂਡ ਮਿਆਰੀ ਵੇਲਾ", "HEEG": "ਪੂਰਬੀ ਗ੍ਰੀਨਲੈਂਡ ਗਰਮੀਆਂ ਦਾ ਵੇਲਾ", "CAT": "ਕੇਂਦਰੀ ਅਫਰੀਕਾ ਵੇਲਾ", "HADT": "ਹਵਾਈ-ਅਲੇਯੂਸ਼ਿਅਨ ਪ੍ਰਕਾਸ਼ ਵੇਲਾ", "AST": "ਅਟਲਾਂਟਿਕ ਮਿਆਰੀ ਵੇਲਾ", "JST": "ਜਪਾਨ ਮਿਆਰੀ ਵੇਲਾ", "COST": "ਕੋਲੰਬੀਆ ਗਰਮੀਆਂ ਦਾ ਵੇਲਾ", "GMT": "ਗ੍ਰੀਨਵਿਚ ਮੀਨ ਵੇਲਾ", "HECU": "ਕਿਊਬਾ ਪ੍ਰਕਾਸ਼ ਵੇਲਾ", "ACWDT": "ਆਸਟ੍ਰੇਲੀਆਈ ਕੇਂਦਰੀ ਪੱਛਮੀ ਪ੍ਰਕਾਸ਼ ਵੇਲਾ", "AKDT": "ਅਲਾਸਕਾ ਪ੍ਰਕਾਸ਼ ਵੇਲਾ", "SGT": "ਸਿੰਗਾਪੁਰ ਮਿਆਰੀ ਵੇਲਾ", "ACST": "ਆਸਟ੍ਰੇਲੀਆਈ ਕੇਂਦਰੀ ਮਿਆਰੀ ਵੇਲਾ", "WART": "ਪੱਛਮੀ ਅਰਜਨਟੀਨਾ ਮਿਆਰੀ ਵੇਲਾ", "CHADT": "ਚੈਥਮ ਪ੍ਰਕਾਸ਼ ਵੇਲਾ", "WAST": "ਪੱਛਮੀ ਅਫਰੀਕਾ ਗਰਮੀਆਂ ਦਾ ਵੇਲਾ", "JDT": "ਜਪਾਨ ਪ੍ਰਕਾਸ਼ ਵੇਲਾ", "AKST": "ਅਲਾਸਕਾ ਮਿਆਰੀ ਵੇਲਾ", "MEZ": "ਮੱਧ ਯੂਰਪੀ ਮਿਆਰੀ ਵੇਲਾ", "IST": "ਭਾਰਤੀ ਮਿਆਰੀ ਵੇਲਾ", "HAST": "ਹਵਾਈ-ਅਲੇਯੂਸ਼ਿਅਨ ਮਿਆਰੀ ਵੇਲਾ", "AEST": "ਆਸਟ੍ਰੇਲੀਆਈ ਪੂਰਬੀ ਮਿਆਰੀ ਵੇਲਾ", "AEDT": "ਆਸਟ੍ਰੇਲੀਆਈ ਪੂਰਬੀ ਪ੍ਰਕਾਸ਼ ਵੇਲਾ", "LHDT": "ਲੌਰਡ ਹੋਵੇ ਪ੍ਰਕਾਸ਼ ਵੇਲਾ", "SRT": "ਸੂਰੀਨਾਮ ਵੇਲਾ", "EAT": "ਪੂਰਬੀ ਅਫਰੀਕਾ ਵੇਲਾ", "CLST": "ਚਿਲੀ ਗਰਮੀਆਂ ਦਾ ਵੇਲਾ", "HNPMX": "ਮੈਕਸੀਕਨ ਪੈਸਿਫਿਕ ਮਿਆਰੀ ਵੇਲਾ", "MESZ": "ਮੱਧ ਯੂਰਪੀ ਗਰਮੀਆਂ ਦਾ ਵੇਲਾ", "BOT": "ਬੋਲੀਵੀਆ ਵੇਲਾ", "ART": "ਅਰਜਨਟੀਨਾ ਮਿਆਰੀ ਵੇਲਾ", "HNCU": "ਕਿਊਬਾ ਮਿਆਰੀ ਵੇਲਾ", "AWDT": "ਆਸਟ੍ਰੇਲੀਆਈ ਪੱਛਮੀ ਪ੍ਰਕਾਸ਼ ਵੇਲਾ"},
	}
}

// Locale returns the current translators string locale
func (pa *pa) Locale() string {
	return pa.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'pa'
func (pa *pa) PluralsCardinal() []locales.PluralRule {
	return pa.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'pa'
func (pa *pa) PluralsOrdinal() []locales.PluralRule {
	return pa.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'pa'
func (pa *pa) PluralsRange() []locales.PluralRule {
	return pa.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'pa'
func (pa *pa) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n >= 0 && n <= 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'pa'
func (pa *pa) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'pa'
func (pa *pa) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := pa.CardinalPluralRule(num1, v1)
	end := pa.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (pa *pa) MonthAbbreviated(month time.Month) string {
	return pa.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (pa *pa) MonthsAbbreviated() []string {
	return pa.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (pa *pa) MonthNarrow(month time.Month) string {
	return pa.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (pa *pa) MonthsNarrow() []string {
	return pa.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (pa *pa) MonthWide(month time.Month) string {
	return pa.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (pa *pa) MonthsWide() []string {
	return pa.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (pa *pa) WeekdayAbbreviated(weekday time.Weekday) string {
	return pa.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (pa *pa) WeekdaysAbbreviated() []string {
	return pa.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (pa *pa) WeekdayNarrow(weekday time.Weekday) string {
	return pa.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (pa *pa) WeekdaysNarrow() []string {
	return pa.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (pa *pa) WeekdayShort(weekday time.Weekday) string {
	return pa.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (pa *pa) WeekdaysShort() []string {
	return pa.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (pa *pa) WeekdayWide(weekday time.Weekday) string {
	return pa.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (pa *pa) WeekdaysWide() []string {
	return pa.daysWide
}

// Decimal returns the decimal point of number
func (pa *pa) Decimal() string {
	return pa.decimal
}

// Group returns the group of number
func (pa *pa) Group() string {
	return pa.group
}

// Group returns the minus sign of number
func (pa *pa) Minus() string {
	return pa.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'pa' and handles both Whole and Real numbers based on 'v'
func (pa *pa) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	inSecondary := false
	groupThreshold := 3

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, pa.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {

			if count == groupThreshold {
				b = append(b, pa.group[0])
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
		b = append(b, pa.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'pa' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (pa *pa) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, pa.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, pa.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, pa.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'pa'
func (pa *pa) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := pa.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	inSecondary := false
	groupThreshold := 3

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, pa.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {

			if count == groupThreshold {
				b = append(b, pa.group[0])
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
		b = append(b, pa.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, pa.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'pa'
// in accounting notation.
func (pa *pa) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := pa.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	inSecondary := false
	groupThreshold := 3

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, pa.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {

			if count == groupThreshold {
				b = append(b, pa.group[0])
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

		b = append(b, pa.minus[0])

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
			b = append(b, pa.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'pa'
func (pa *pa) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'pa'
func (pa *pa) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, pa.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'pa'
func (pa *pa) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, pa.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'pa'
func (pa *pa) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, pa.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, pa.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'pa'
func (pa *pa) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, pa.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, pa.periodsAbbreviated[0]...)
	} else {
		b = append(b, pa.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'pa'
func (pa *pa) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, pa.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, pa.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, pa.periodsAbbreviated[0]...)
	} else {
		b = append(b, pa.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'pa'
func (pa *pa) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, pa.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, pa.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, pa.periodsAbbreviated[0]...)
	} else {
		b = append(b, pa.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'pa'
func (pa *pa) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, pa.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, pa.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, pa.periodsAbbreviated[0]...)
	} else {
		b = append(b, pa.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := pa.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
