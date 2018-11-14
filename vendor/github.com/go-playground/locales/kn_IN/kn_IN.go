package kn_IN

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type kn_IN struct {
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

// New returns a new instance of translator for the 'kn_IN' locale
func New() locales.Translator {
	return &kn_IN{
		locale:             "kn_IN",
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
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		monthsAbbreviated:  []string{"", "ಜನವರಿ", "ಫೆಬ್ರವರಿ", "ಮಾರ್ಚ್", "ಏಪ್ರಿ", "ಮೇ", "ಜೂನ್", "ಜುಲೈ", "ಆಗ", "ಸೆಪ್ಟೆಂ", "ಅಕ್ಟೋ", "ನವೆಂ", "ಡಿಸೆಂ"},
		monthsNarrow:       []string{"", "ಜ", "ಫೆ", "ಮಾ", "ಏ", "ಮೇ", "ಜೂ", "ಜು", "ಆ", "ಸೆ", "ಅ", "ನ", "ಡಿ"},
		monthsWide:         []string{"", "ಜನವರಿ", "ಫೆಬ್ರವರಿ", "ಮಾರ್ಚ್", "ಏಪ್ರಿಲ್", "ಮೇ", "ಜೂನ್", "ಜುಲೈ", "ಆಗಸ್ಟ್", "ಸೆಪ್ಟೆಂಬರ್", "ಅಕ್ಟೋಬರ್", "ನವೆಂಬರ್", "ಡಿಸೆಂಬರ್"},
		daysAbbreviated:    []string{"ಭಾನು", "ಸೋಮ", "ಮಂಗಳ", "ಬುಧ", "ಗುರು", "ಶುಕ್ರ", "ಶನಿ"},
		daysNarrow:         []string{"ಭಾ", "ಸೋ", "ಮಂ", "ಬು", "ಗು", "ಶು", "ಶ"},
		daysShort:          []string{"ಭಾನು", "ಸೋಮ", "ಮಂಗಳ", "ಬುಧ", "ಗುರು", "ಶುಕ್ರ", "ಶನಿ"},
		daysWide:           []string{"ಭಾನುವಾರ", "ಸೋಮವಾರ", "ಮಂಗಳವಾರ", "ಬುಧವಾರ", "ಗುರುವಾರ", "ಶುಕ್ರವಾರ", "ಶನಿವಾರ"},
		periodsAbbreviated: []string{"ಪೂರ್ವಾಹ್ನ", "ಅಪರಾಹ್ನ"},
		periodsNarrow:      []string{"ಪೂ", "ಅ"},
		periodsWide:        []string{"ಪೂರ್ವಾಹ್ನ", "ಅಪರಾಹ್ನ"},
		erasAbbreviated:    []string{"ಕ್ರಿ.ಪೂ", "ಕ್ರಿ.ಶ"},
		erasNarrow:         []string{"", ""},
		erasWide:           []string{"ಕ್ರಿಸ್ತ ಪೂರ್ವ", "ಕ್ರಿಸ್ತ ಶಕ"},
		timezones:          map[string]string{"ADT": "ಅಟ್ಲಾಂಟಿಕ್ ದಿನದ ಸಮಯ", "SAST": "ದಕ್ಷಿಣ ಆಫ್ರಿಕಾ ಪ್ರಮಾಣಿತ ಸಮಯ", "IST": "ಭಾರತೀಯ ಪ್ರಮಾಣಿತ ಸಮಯ", "CLT": "ಚಿಲಿ ಪ್ರಮಾಣಿತ ಸಮಯ", "ChST": "ಚಮೋರೋ ಪ್ರಮಾಣಿತ ಸಮಯ", "AEST": "ಆಸ್ಟ್ರೇಲಿಯಾದ ಪೂರ್ವ ಪ್ರಮಾಣಿತ ಸಮಯ", "AKDT": "\u200cಅಲಾಸ್ಕಾ ಹಗಲು ಸಮಯ", "HNOG": "ಪಶ್ಚಿಮ ಗ್ರೀನ್\u200cಲ್ಯಾಂಡ್ ಪ್ರಮಾಣಿತ ಸಮಯ", "OESZ": "ಪೂರ್ವ ಯುರೋಪಿಯನ್ ಬೇಸಿಗೆ ಸಮಯ", "ART": "ಅರ್ಜೆಂಟೀನಾ ಪ್ರಮಾಣಿತ ಸಮಯ", "MESZ": "ಮಧ್ಯ ಯುರೋಪಿಯನ್ ಬೇಸಿಗೆ ಸಮಯ", "LHST": "ಲಾರ್ಡ್ ಹೋವ್ ಪ್ರಮಾಣಿತ ಸಮಯ", "VET": "ವೆನಿಜುವೆಲಾ ಸಮಯ", "HNCU": "ಕ್ಯೂಬಾ ಪ್ರಮಾಣಿತ ಸಮಯ", "MST": "ಉತ್ತರ ಅಮೆರಿಕದ ಪರ್ವತ ಪ್ರಮಾಣಿತ ಸಮಯ", "BT": "ಭೂತಾನ್ ಸಮಯ", "SGT": "ಸಿಂಗಾಪುರ್ ಪ್ರಮಾಣಿತ ಸಮಯ", "HEOG": "ಪಶ್ಚಿಮ ಗ್ರೀನ್\u200cಲ್ಯಾಂಡ್ ಬೇಸಿಗೆ ಸಮಯ", "HEPM": "ಸೇಂಟ್ ಪಿಯರ್ ಮತ್ತು ಮಿಕ್ವೆಲನ್ ಹಗಲು ಸಮಯ", "EAT": "ಪೂರ್ವ ಆಫ್ರಿಕಾ ಸಮಯ", "HEPMX": "ಮೆಕ್ಸಿಕನ್ ಪೆಸಿಫಿಕ್ ಹಗಲು ಸಮಯ", "WAST": "ಪಶ್ಚಿಮ ಆಫ್ರಿಕಾ ಬೇಸಿಗೆ ಸಮಯ", "JDT": "ಜಪಾನ್ ಹಗಲು ಸಮಯ", "AKST": "ಅಲಸ್ಕಾ ಪ್ರಮಾಣಿತ ಸಮಯ", "LHDT": "ಲಾರ್ಡ್ ಹೋವ್ ಬೆಳಗಿನ ಸಮಯ", "AWDT": "ಆಸ್ಟ್ರೇಲಿಯಾದ ಪಶ್ಚಿಮ ಹಗಲು ಸಮಯ", "HNT": "ನ್ಯೂಫೌಂಡ್\u200cಲ್ಯಾಂಡ್ ಪ್ರಮಾಣಿತ ಸಮಯ", "HECU": "ಕ್ಯೂಬಾ ದಿನದ ಸಮಯ", "HNPM": "ಸೇಂಟ್ ಪಿಯರ್ ಮತ್ತು ಮಿಕ್ವೆಲನ್ ಪ್ರಮಾಣಿತ ಸಮಯ", "HNNOMX": "ವಾಯವ್ಯ ಮೆಕ್ಸಿಕೊ ಪ್ರಮಾಣಿತ ಸಮಯ", "CLST": "ಚಿಲಿ ಬೇಸಿಗೆ ಸಮಯ", "GMT": "ಗ್ರೀನ್\u200cವಿಚ್ ಸರಾಸರಿ ಕಾಲಮಾನ", "CDT": "ಉತ್ತರ ಅಮೆರಿಕದ ಕೇಂದ್ರೀಯ ದಿನದ ಸಮಯ", "MDT": "ಉತ್ತರ ಅಮೆರಿಕದ ಪರ್ವತ ದಿನದ ಸಮಯ", "WARST": "ಪಶ್ಚಿಮ ಅರ್ಜೆಂಟೀನಾ ಬೇಸಿಗೆ ಸಮಯ", "HADT": "ಹವಾಯಿ-ಅಲ್ಯುಟಿಯನ್ ಹಗಲು ಸಮಯ", "GYT": "ಗಯಾನಾ ಸಮಯ", "∅∅∅": "ಬ್ರೆಸಿಲಿಯಾ ಬೇಸಿಗೆ ಸಮಯ", "ACWST": "ಆಸ್ಟ್ರೇಲಿಯಾದ ಕೇಂದ್ರ ಪಶ್ಚಿಮ ಪ್ರಮಾಣಿತ ಸಮಯ", "TMST": "ತುರ್ಕ್\u200cಮೇನಿಸ್ತಾನ್ ಬೇಸಿಗೆ ಸಮಯ", "OEZ": "ಪೂರ್ವ ಯುರೋಪಿಯನ್ ಪ್ರಮಾಣಿತ ಸಮಯ", "HAST": "ಹವಾಯಿ-ಅಲ್ಯುಟಿಯನ್ ಪ್ರಮಾಣಿತ ಸಮಯ", "HKT": "ಹಾಂಗ್ ಕಾಂಗ್ ಪ್ರಮಾಣಿತ ಸಮಯ", "WART": "ಪಶ್ಚಿಮ ಅರ್ಜೆಂಟೀನಾ ಪ್ರಮಾಣಿತ ಸಮಯ", "SRT": "ಸುರಿನೇಮ್ ಸಮಯ", "GFT": "ಫ್ರೆಂಚ್ ಗಯಾನಾ ಸಮಯ", "NZST": "ನ್ಯೂಜಿಲ್ಯಾಂಡ್ ಪ್ರಮಾಣಿತ ಸಮಯ", "MYT": "ಮಲೇಷಿಯಾ ಸಮಯ", "ACST": "ಆಸ್ಟ್ರೇಲಿಯಾದ ಕೇಂದ್ರ ಪ್ರಮಾಣಿತ ಸಮಯ", "ACWDT": "ಆಸ್ಟ್ರೇಲಿಯಾದ ಕೇಂದ್ರ ಪಶ್ಚಿಮ ಹಗಲು ಸಮಯ", "WESZ": "ಪಶ್ಚಿಮ ಯುರೋಪಿಯನ್ ಬೇಸಿಗೆ ಸಮಯ", "TMT": "ತುರ್ಕ್\u200cಮೇನಿಸ್ತಾನ್ ಪ್ರಮಾಣಿತ ಸಮಯ", "COST": "ಕೊಲಂಬಿಯಾ ಬೇಸಿಗೆ ಸಮಯ", "PDT": "ಉತ್ತರ ಅಮೆರಿಕದ ಪೆಸಿಫಿಕ್ ದಿನದ ಸಮಯ", "ACDT": "ಆಸ್ಟ್ರೇಲಿಯಾದ ಕೇಂದ್ರ ಹಗಲು ಸಮಯ", "HEEG": "ಪೂರ್ವ ಗ್ರೀನ್\u200cಲ್ಯಾಂಡ್ ಬೇಸಿಗೆ ಸಮಯ", "MEZ": "ಮಧ್ಯ ಯುರೋಪಿಯನ್ ಪ್ರಮಾಣಿತ ಸಮಯ", "HENOMX": "ವಾಯವ್ಯ ಮೆಕ್ಸಿಕೊ ಹಗಲು ಸಮಯ", "HNPMX": "ಮೆಕ್ಸಿಕನ್ ಪೆಸಿಫಿಕ್ ಪ್ರಮಾಣಿತ ಸಮಯ", "WEZ": "ಪಶ್ಚಿಮ ಯುರೋಪಿಯನ್ ಪ್ರಮಾಣಿತ ಸಮಯ", "ECT": "ಈಕ್ವೆಡಾರ್ ಸಮಯ", "UYST": "ಉರುಗ್ವೇ ಬೇಸಿಗೆ ಸಮಯ", "CAT": "ಮಧ್ಯ ಆಫ್ರಿಕಾ ಸಮಯ", "PST": "ಉತ್ತರ ಅಮೆರಿಕದ ಪೆಸಿಫಿಕ್ ಪ್ರಮಾಣಿತ ಸಮಯ", "BOT": "ಬೊಲಿವಿಯಾ ಸಮಯ", "HNEG": "ಪೂರ್ವ ಗ್ರೀನ್\u200cಲ್ಯಾಂಡ್ ಪ್ರಮಾಣಿತ ಸಮಯ", "HAT": "ನ್ಯೂಫೌಂಡ್\u200cಲ್ಯಾಂಡ್ ದಿನದ ಸಮಯ", "WITA": "ಮಧ್ಯ ಇಂಡೋನೇಷಿಯಾ ಸಮಯ", "NZDT": "ನ್ಯೂಜಿಲ್ಯಾಂಡ್ ಹಗಲು ಸಮಯ", "EDT": "ಉತ್ತರ ಅಮೆರಿಕದ ಪೂರ್ವದ ದಿನದ ಸಮಯ", "HKST": "ಹಾಂಗ್ ಕಾಂಗ್ ಬೇಸಿಗೆ ಸಮಯ", "CHAST": "ಚಥಾಮ್ ಪ್ರಮಾಣಿತ ಸಮಯ", "CHADT": "ಚಥಾಮ್ ಹಗಲು ಸಮಯ", "CST": "ಉತ್ತರ ಅಮೆರಿಕದ ಕೇಂದ್ರ ಪ್ರಮಾಣಿತ ಸಮಯ", "AWST": "ಆಸ್ಟ್ರೇಲಿಯಾದ ಪಶ್ಚಿಮ ಪ್ರಮಾಣಿತ ಸಮಯ", "AEDT": "ಪೂರ್ವ ಆಸ್ಟ್ರೇಲಿಯಾದ ಹಗಲು ಸಮಯ", "UYT": "ಉರುಗ್ವೇ ಪ್ರಮಾಣಿತ ಸಮಯ", "AST": "ಅಟ್ಲಾಂಟಿಕ್ ಪ್ರಮಾಣಿತ ಸಮಯ", "WAT": "ಪಶ್ಚಿಮ ಆಫ್ರಿಕಾ ಪ್ರಮಾಣಿತ ಸಮಯ", "WIB": "ಪಶ್ಚಿಮ ಇಂಡೋನೇಷಿಯ ಸಮಯ", "COT": "ಕೊಲಂಬಿಯಾ ಪ್ರಮಾಣಿತ ಸಮಯ", "JST": "ಜಪಾನ್ ಪ್ರಮಾಣಿತ ಸಮಯ", "EST": "ಉತ್ತರ ಅಮೆರಿಕದ ಪೂರ್ವದ ಪ್ರಮಾಣಿತ ಸಮಯ", "WIT": "ಪೂರ್ವ ಇಂಡೋನೇಷಿಯಾ ಸಮಯ", "ARST": "ಅರ್ಜೆಂಟಿನಾ ಬೇಸಿಗೆ ಸಮಯ"},
	}
}

// Locale returns the current translators string locale
func (kn *kn_IN) Locale() string {
	return kn.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'kn_IN'
func (kn *kn_IN) PluralsCardinal() []locales.PluralRule {
	return kn.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'kn_IN'
func (kn *kn_IN) PluralsOrdinal() []locales.PluralRule {
	return kn.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'kn_IN'
func (kn *kn_IN) PluralsRange() []locales.PluralRule {
	return kn.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'kn_IN'
func (kn *kn_IN) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)

	if (i == 0) || (n == 1) {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'kn_IN'
func (kn *kn_IN) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'kn_IN'
func (kn *kn_IN) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := kn.CardinalPluralRule(num1, v1)
	end := kn.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (kn *kn_IN) MonthAbbreviated(month time.Month) string {
	return kn.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (kn *kn_IN) MonthsAbbreviated() []string {
	return kn.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (kn *kn_IN) MonthNarrow(month time.Month) string {
	return kn.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (kn *kn_IN) MonthsNarrow() []string {
	return kn.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (kn *kn_IN) MonthWide(month time.Month) string {
	return kn.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (kn *kn_IN) MonthsWide() []string {
	return kn.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (kn *kn_IN) WeekdayAbbreviated(weekday time.Weekday) string {
	return kn.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (kn *kn_IN) WeekdaysAbbreviated() []string {
	return kn.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (kn *kn_IN) WeekdayNarrow(weekday time.Weekday) string {
	return kn.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (kn *kn_IN) WeekdaysNarrow() []string {
	return kn.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (kn *kn_IN) WeekdayShort(weekday time.Weekday) string {
	return kn.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (kn *kn_IN) WeekdaysShort() []string {
	return kn.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (kn *kn_IN) WeekdayWide(weekday time.Weekday) string {
	return kn.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (kn *kn_IN) WeekdaysWide() []string {
	return kn.daysWide
}

// Decimal returns the decimal point of number
func (kn *kn_IN) Decimal() string {
	return kn.decimal
}

// Group returns the group of number
func (kn *kn_IN) Group() string {
	return kn.group
}

// Group returns the minus sign of number
func (kn *kn_IN) Minus() string {
	return kn.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'kn_IN' and handles both Whole and Real numbers based on 'v'
func (kn *kn_IN) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, kn.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, kn.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, kn.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'kn_IN' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (kn *kn_IN) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, kn.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, kn.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, kn.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'kn_IN'
func (kn *kn_IN) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := kn.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, kn.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, kn.group[0])
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
		b = append(b, kn.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, kn.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'kn_IN'
// in accounting notation.
func (kn *kn_IN) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := kn.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, kn.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, kn.group[0])
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

		b = append(b, kn.minus[0])

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
			b = append(b, kn.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'kn_IN'
func (kn *kn_IN) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'kn_IN'
func (kn *kn_IN) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, kn.monthsAbbreviated[t.Month()]...)
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

// FmtDateLong returns the long date representation of 't' for 'kn_IN'
func (kn *kn_IN) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, kn.monthsWide[t.Month()]...)
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

// FmtDateFull returns the full date representation of 't' for 'kn_IN'
func (kn *kn_IN) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, kn.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, kn.monthsWide[t.Month()]...)
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

// FmtTimeShort returns the short time representation of 't' for 'kn_IN'
func (kn *kn_IN) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	if h < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, kn.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, kn.periodsAbbreviated[0]...)
	} else {
		b = append(b, kn.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'kn_IN'
func (kn *kn_IN) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	if h < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, kn.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, kn.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, kn.periodsAbbreviated[0]...)
	} else {
		b = append(b, kn.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'kn_IN'
func (kn *kn_IN) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	if h < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, kn.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, kn.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, kn.periodsAbbreviated[0]...)
	} else {
		b = append(b, kn.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'kn_IN'
func (kn *kn_IN) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	if h < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, kn.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, kn.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, kn.periodsAbbreviated[0]...)
	} else {
		b = append(b, kn.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := kn.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
