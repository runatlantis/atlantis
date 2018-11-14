package el_CY

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type el_CY struct {
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

// New returns a new instance of translator for the 'el_CY' locale
func New() locales.Translator {
	return &el_CY{
		locale:                 "el_CY",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         []locales.PluralRule{6},
		pluralsRange:           []locales.PluralRule{2, 6},
		decimal:                ",",
		group:                  ".",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "Ιαν", "Φεβ", "Μαρ", "Απρ", "Μαΐ", "Ιουν", "Ιουλ", "Αυγ", "Σεπ", "Οκτ", "Νοε", "Δεκ"},
		monthsNarrow:           []string{"", "Ι", "Φ", "Μ", "Α", "Μ", "Ι", "Ι", "Α", "Σ", "Ο", "Ν", "Δ"},
		monthsWide:             []string{"", "Ιανουαρίου", "Φεβρουαρίου", "Μαρτίου", "Απριλίου", "Μαΐου", "Ιουνίου", "Ιουλίου", "Αυγούστου", "Σεπτεμβρίου", "Οκτωβρίου", "Νοεμβρίου", "Δεκεμβρίου"},
		daysAbbreviated:        []string{"Κυρ", "Δευ", "Τρί", "Τετ", "Πέμ", "Παρ", "Σάβ"},
		daysNarrow:             []string{"Κ", "Δ", "Τ", "Τ", "Π", "Π", "Σ"},
		daysShort:              []string{"Κυ", "Δε", "Τρ", "Τε", "Πέ", "Πα", "Σά"},
		daysWide:               []string{"Κυριακή", "Δευτέρα", "Τρίτη", "Τετάρτη", "Πέμπτη", "Παρασκευή", "Σάββατο"},
		periodsAbbreviated:     []string{"π.μ.", "μ.μ."},
		periodsNarrow:          []string{"πμ", "μμ"},
		periodsWide:            []string{"π.μ.", "μ.μ."},
		erasAbbreviated:        []string{"π.Χ.", "μ.Χ."},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"προ Χριστού", "μετά Χριστόν"},
		timezones:              map[string]string{"GYT": "Ώρα Γουιάνας", "GMT": "Μέση ώρα Γκρίνουιτς", "AEST": "Χειμερινή ώρα Ανατολικής Αυστραλίας", "WIB": "Ώρα Δυτικής Ινδονησίας", "MESZ": "Θερινή ώρα Κεντρικής Ευρώπης", "IST": "Ώρα Ινδίας", "SRT": "Ώρα Σουρινάμ", "ChST": "Ώρα Τσαμόρο", "PDT": "Θερινή ώρα Ειρηνικού", "HEPMX": "Θερινή ώρα Ειρηνικού Μεξικού", "CDT": "Κεντρική θερινή ώρα Βόρειας Αμερικής", "JST": "Χειμερινή ώρα Ιαπωνίας", "HNPM": "Χειμερινή ώρα Σεν Πιερ και Μικελόν", "CHAST": "Χειμερινή ώρα Τσάταμ", "SAST": "Χειμερινή ώρα Νότιας Αφρικής", "JDT": "Θερινή ώρα Ιαπωνίας", "HNEG": "Χειμερινή ώρα Ανατολικής Γροιλανδίας", "EDT": "Ανατολική θερινή ώρα Βόρειας Αμερικής", "HNNOMX": "Χειμερινή ώρα Βορειοδυτικού Μεξικού", "HADT": "Θερινή ώρα Χαβάης-Αλεούτιων Νήσων", "AST": "Χειμερινή ώρα Ατλαντικού", "ACST": "Χειμερινή ώρα Κεντρικής Αυστραλίας", "VET": "Ώρα Βενεζουέλας", "SGT": "Ώρα Σιγκαπούρης", "HKT": "Χειμερινή ώρα Χονγκ Κονγκ", "HAST": "Χειμερινή ώρα Χαβάης-Αλεούτιων Νήσων", "∅∅∅": "Θερινή ώρα Αμαζονίου", "AWDT": "Θερινή ώρα Δυτικής Αυστραλίας", "ADT": "Θερινή ώρα Ατλαντικού", "COST": "Θερινή ώρα Κολομβίας", "WEZ": "Χειμερινή ώρα Δυτικής Ευρώπης", "BOT": "Ώρα Βολιβίας", "HEEG": "Θερινή ώρα Ανατολικής Γροιλανδίας", "HKST": "Θερινή ώρα Χονγκ Κονγκ", "LHST": "Χειμερινή ώρα Λορντ Χάου", "HNT": "Χειμερινή ώρα Νέας Γης", "HAT": "Θερινή ώρα Νέας Γης", "HENOMX": "Θερινή ώρα Βορειοδυτικού Μεξικού", "TMT": "Χειμερινή ώρα Τουρκμενιστάν", "ART": "Χειμερινή ώρα Αργεντινής", "UYST": "Θερινή ώρα Ουρουγουάης", "HNOG": "Χειμερινή ώρα Δυτικής Γροιλανδίας", "EST": "Ανατολική χειμερινή ώρα Βόρειας Αμερικής", "WAST": "Θερινή ώρα Δυτικής Αφρικής", "AKST": "Χειμερινή ώρα Αλάσκας", "CLST": "Θερινή ώρα Χιλής", "WIT": "Ώρα Ανατολικής Ινδονησίας", "ARST": "Θερινή ώρα Αργεντινής", "HNCU": "Χειμερινή ώρα Κούβας", "HNPMX": "Χειμερινή ώρα Ειρηνικού Μεξικού", "AKDT": "Θερινή ώρα Αλάσκας", "MEZ": "Χειμερινή ώρα Κεντρικής Ευρώπης", "WITA": "Ώρα Κεντρικής Ινδονησίας", "COT": "Χειμερινή ώρα Κολομβίας", "GFT": "Ώρα Γαλλικής Γουιάνας", "ACDT": "Θερινή ώρα Κεντρικής Αυστραλίας", "HEOG": "Θερινή ώρα Δυτικής Γροιλανδίας", "OESZ": "Θερινή ώρα Ανατολικής Ευρώπης", "UYT": "Χειμερινή ώρα Ουρουγουάης", "WAT": "Χειμερινή ώρα Δυτικής Αφρικής", "LHDT": "Θερινή ώρα Λορντ Χάου", "CST": "Κεντρική χειμερινή ώρα Βόρειας Αμερικής", "WESZ": "Θερινή ώρα Δυτικής Ευρώπης", "ACWDT": "Θερινή ώρα Κεντροδυτικής Αυστραλίας", "WARST": "Θερινή ώρα Δυτικής Αργεντινής", "MST": "Χειμερινή ώρα Μακάο", "MDT": "Θερινή ώρα Μακάο", "TMST": "Θερινή ώρα Τουρκμενιστάν", "CHADT": "Θερινή ώρα Τσάταμ", "AEDT": "Θερινή ώρα Ανατολικής Αυστραλίας", "BT": "Ώρα Μπουτάν", "ACWST": "Χειμερινή ώρα Κεντροδυτικής Αυστραλίας", "CAT": "Ώρα Κεντρικής Αφρικής", "OEZ": "Χειμερινή ώρα Ανατολικής Ευρώπης", "AWST": "Χειμερινή ώρα Δυτικής Αυστραλίας", "NZDT": "Θερινή ώρα Νέας Ζηλανδίας", "MYT": "Ώρα Μαλαισίας", "ECT": "Ώρα Ισημερινού", "EAT": "Ώρα Ανατολικής Αφρικής", "CLT": "Χειμερινή ώρα Χιλής", "HECU": "Θερινή ώρα Κούβας", "PST": "Χειμερινή ώρα Ειρηνικού", "NZST": "Χειμερινή ώρα Νέας Ζηλανδίας", "WART": "Χειμερινή ώρα Δυτικής Αργεντινής", "HEPM": "Θερινή ώρα Σεν Πιερ και Μικελόν"},
	}
}

// Locale returns the current translators string locale
func (el *el_CY) Locale() string {
	return el.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'el_CY'
func (el *el_CY) PluralsCardinal() []locales.PluralRule {
	return el.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'el_CY'
func (el *el_CY) PluralsOrdinal() []locales.PluralRule {
	return el.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'el_CY'
func (el *el_CY) PluralsRange() []locales.PluralRule {
	return el.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'el_CY'
func (el *el_CY) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'el_CY'
func (el *el_CY) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'el_CY'
func (el *el_CY) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := el.CardinalPluralRule(num1, v1)
	end := el.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (el *el_CY) MonthAbbreviated(month time.Month) string {
	return el.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (el *el_CY) MonthsAbbreviated() []string {
	return el.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (el *el_CY) MonthNarrow(month time.Month) string {
	return el.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (el *el_CY) MonthsNarrow() []string {
	return el.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (el *el_CY) MonthWide(month time.Month) string {
	return el.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (el *el_CY) MonthsWide() []string {
	return el.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (el *el_CY) WeekdayAbbreviated(weekday time.Weekday) string {
	return el.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (el *el_CY) WeekdaysAbbreviated() []string {
	return el.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (el *el_CY) WeekdayNarrow(weekday time.Weekday) string {
	return el.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (el *el_CY) WeekdaysNarrow() []string {
	return el.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (el *el_CY) WeekdayShort(weekday time.Weekday) string {
	return el.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (el *el_CY) WeekdaysShort() []string {
	return el.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (el *el_CY) WeekdayWide(weekday time.Weekday) string {
	return el.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (el *el_CY) WeekdaysWide() []string {
	return el.daysWide
}

// Decimal returns the decimal point of number
func (el *el_CY) Decimal() string {
	return el.decimal
}

// Group returns the group of number
func (el *el_CY) Group() string {
	return el.group
}

// Group returns the minus sign of number
func (el *el_CY) Minus() string {
	return el.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'el_CY' and handles both Whole and Real numbers based on 'v'
func (el *el_CY) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, el.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, el.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, el.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'el_CY' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (el *el_CY) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, el.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, el.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, el.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'el_CY'
func (el *el_CY) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := el.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, el.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, el.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, el.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, el.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, el.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'el_CY'
// in accounting notation.
func (el *el_CY) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := el.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, el.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, el.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, el.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, el.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, el.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, el.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'el_CY'
func (el *el_CY) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'el_CY'
func (el *el_CY) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, el.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'el_CY'
func (el *el_CY) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, el.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'el_CY'
func (el *el_CY) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, el.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, el.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'el_CY'
func (el *el_CY) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, el.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, el.periodsAbbreviated[0]...)
	} else {
		b = append(b, el.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'el_CY'
func (el *el_CY) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, el.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, el.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, el.periodsAbbreviated[0]...)
	} else {
		b = append(b, el.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'el_CY'
func (el *el_CY) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, el.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, el.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, el.periodsAbbreviated[0]...)
	} else {
		b = append(b, el.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'el_CY'
func (el *el_CY) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, el.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, el.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, el.periodsAbbreviated[0]...)
	} else {
		b = append(b, el.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := el.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
