package el_GR

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type el_GR struct {
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

// New returns a new instance of translator for the 'el_GR' locale
func New() locales.Translator {
	return &el_GR{
		locale:                 "el_GR",
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
		timezones:              map[string]string{"MST": "Χειμερινή ώρα Μακάο", "EAT": "Ώρα Ανατολικής Αφρικής", "GYT": "Ώρα Γουιάνας", "UYT": "Χειμερινή ώρα Ουρουγουάης", "HEEG": "Θερινή ώρα Ανατολικής Γροιλανδίας", "LHDT": "Θερινή ώρα Λορντ Χάου", "HEPM": "Θερινή ώρα Σεν Πιερ και Μικελόν", "HENOMX": "Θερινή ώρα Βορειοδυτικού Μεξικού", "HNPMX": "Χειμερινή ώρα Ειρηνικού Μεξικού", "BOT": "Ώρα Βολιβίας", "∅∅∅": "Θερινή ώρα Αζορών", "VET": "Ώρα Βενεζουέλας", "CAT": "Ώρα Κεντρικής Αφρικής", "ART": "Χειμερινή ώρα Αργεντινής", "HECU": "Θερινή ώρα Κούβας", "PST": "Χειμερινή ώρα Ειρηνικού", "BT": "Ώρα Μπουτάν", "NZST": "Χειμερινή ώρα Νέας Ζηλανδίας", "IST": "Ώρα Ινδίας", "GFT": "Ώρα Γαλλικής Γουιάνας", "WAT": "Χειμερινή ώρα Δυτικής Αφρικής", "WAST": "Θερινή ώρα Δυτικής Αφρικής", "JDT": "Θερινή ώρα Ιαπωνίας", "CLT": "Χειμερινή ώρα Χιλής", "OESZ": "Θερινή ώρα Ανατολικής Ευρώπης", "AEDT": "Θερινή ώρα Ανατολικής Αυστραλίας", "HNEG": "Χειμερινή ώρα Ανατολικής Γροιλανδίας", "TMT": "Χειμερινή ώρα Τουρκμενιστάν", "LHST": "Χειμερινή ώρα Λορντ Χάου", "HNNOMX": "Χειμερινή ώρα Βορειοδυτικού Μεξικού", "HAST": "Χειμερινή ώρα Χαβάης-Αλεούτιων Νήσων", "ARST": "Θερινή ώρα Αργεντινής", "GMT": "Μέση ώρα Γκρίνουιτς", "PDT": "Θερινή ώρα Ειρηνικού", "SAST": "Χειμερινή ώρα Νότιας Αφρικής", "MEZ": "Χειμερινή ώρα Κεντρικής Ευρώπης", "ACST": "Χειμερινή ώρα Κεντρικής Αυστραλίας", "ACWST": "Χειμερινή ώρα Κεντροδυτικής Αυστραλίας", "AKDT": "Θερινή ώρα Αλάσκας", "TMST": "Θερινή ώρα Τουρκμενιστάν", "OEZ": "Χειμερινή ώρα Ανατολικής Ευρώπης", "AWST": "Χειμερινή ώρα Δυτικής Αυστραλίας", "MYT": "Ώρα Μαλαισίας", "AKST": "Χειμερινή ώρα Αλάσκας", "WARST": "Θερινή ώρα Δυτικής Αργεντινής", "WITA": "Ώρα Κεντρικής Ινδονησίας", "HEPMX": "Θερινή ώρα Ειρηνικού Μεξικού", "AEST": "Χειμερινή ώρα Ανατολικής Αυστραλίας", "ACWDT": "Θερινή ώρα Κεντροδυτικής Αυστραλίας", "HKST": "Θερινή ώρα Χονγκ Κονγκ", "WART": "Χειμερινή ώρα Δυτικής Αργεντινής", "HNT": "Χειμερινή ώρα Νέας Γης", "HNPM": "Χειμερινή ώρα Σεν Πιερ και Μικελόν", "ChST": "Ώρα Τσαμόρο", "WESZ": "Θερινή ώρα Δυτικής Ευρώπης", "WIB": "Ώρα Δυτικής Ινδονησίας", "HKT": "Χειμερινή ώρα Χονγκ Κονγκ", "CLST": "Θερινή ώρα Χιλής", "HADT": "Θερινή ώρα Χαβάης-Αλεούτιων Νήσων", "UYST": "Θερινή ώρα Ουρουγουάης", "CST": "Κεντρική χειμερινή ώρα Βόρειας Αμερικής", "NZDT": "Θερινή ώρα Νέας Ζηλανδίας", "ACDT": "Θερινή ώρα Κεντρικής Αυστραλίας", "HEOG": "Θερινή ώρα Δυτικής Γροιλανδίας", "HAT": "Θερινή ώρα Νέας Γης", "CHAST": "Χειμερινή ώρα Τσάταμ", "COST": "Θερινή ώρα Κολομβίας", "ADT": "Θερινή ώρα Ατλαντικού", "SGT": "Ώρα Σιγκαπούρης", "HNOG": "Χειμερινή ώρα Δυτικής Γροιλανδίας", "CDT": "Κεντρική θερινή ώρα Βόρειας Αμερικής", "MDT": "Θερινή ώρα Μακάο", "SRT": "Ώρα Σουρινάμ", "COT": "Χειμερινή ώρα Κολομβίας", "CHADT": "Θερινή ώρα Τσάταμ", "HNCU": "Χειμερινή ώρα Κούβας", "EDT": "Ανατολική θερινή ώρα Βόρειας Αμερικής", "MESZ": "Θερινή ώρα Κεντρικής Ευρώπης", "AWDT": "Θερινή ώρα Δυτικής Αυστραλίας", "AST": "Χειμερινή ώρα Ατλαντικού", "JST": "Χειμερινή ώρα Ιαπωνίας", "EST": "Ανατολική χειμερινή ώρα Βόρειας Αμερικής", "WIT": "Ώρα Ανατολικής Ινδονησίας", "ECT": "Ώρα Ισημερινού", "WEZ": "Χειμερινή ώρα Δυτικής Ευρώπης"},
	}
}

// Locale returns the current translators string locale
func (el *el_GR) Locale() string {
	return el.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'el_GR'
func (el *el_GR) PluralsCardinal() []locales.PluralRule {
	return el.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'el_GR'
func (el *el_GR) PluralsOrdinal() []locales.PluralRule {
	return el.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'el_GR'
func (el *el_GR) PluralsRange() []locales.PluralRule {
	return el.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'el_GR'
func (el *el_GR) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'el_GR'
func (el *el_GR) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'el_GR'
func (el *el_GR) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

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
func (el *el_GR) MonthAbbreviated(month time.Month) string {
	return el.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (el *el_GR) MonthsAbbreviated() []string {
	return el.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (el *el_GR) MonthNarrow(month time.Month) string {
	return el.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (el *el_GR) MonthsNarrow() []string {
	return el.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (el *el_GR) MonthWide(month time.Month) string {
	return el.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (el *el_GR) MonthsWide() []string {
	return el.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (el *el_GR) WeekdayAbbreviated(weekday time.Weekday) string {
	return el.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (el *el_GR) WeekdaysAbbreviated() []string {
	return el.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (el *el_GR) WeekdayNarrow(weekday time.Weekday) string {
	return el.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (el *el_GR) WeekdaysNarrow() []string {
	return el.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (el *el_GR) WeekdayShort(weekday time.Weekday) string {
	return el.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (el *el_GR) WeekdaysShort() []string {
	return el.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (el *el_GR) WeekdayWide(weekday time.Weekday) string {
	return el.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (el *el_GR) WeekdaysWide() []string {
	return el.daysWide
}

// Decimal returns the decimal point of number
func (el *el_GR) Decimal() string {
	return el.decimal
}

// Group returns the group of number
func (el *el_GR) Group() string {
	return el.group
}

// Group returns the minus sign of number
func (el *el_GR) Minus() string {
	return el.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'el_GR' and handles both Whole and Real numbers based on 'v'
func (el *el_GR) FmtNumber(num float64, v uint64) string {

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

// FmtPercent returns 'num' with digits/precision of 'v' for 'el_GR' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (el *el_GR) FmtPercent(num float64, v uint64) string {
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

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'el_GR'
func (el *el_GR) FmtCurrency(num float64, v uint64, currency currency.Type) string {

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

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'el_GR'
// in accounting notation.
func (el *el_GR) FmtAccounting(num float64, v uint64, currency currency.Type) string {

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

// FmtDateShort returns the short date representation of 't' for 'el_GR'
func (el *el_GR) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'el_GR'
func (el *el_GR) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'el_GR'
func (el *el_GR) FmtDateLong(t time.Time) string {

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

// FmtDateFull returns the full date representation of 't' for 'el_GR'
func (el *el_GR) FmtDateFull(t time.Time) string {

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

// FmtTimeShort returns the short time representation of 't' for 'el_GR'
func (el *el_GR) FmtTimeShort(t time.Time) string {

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

// FmtTimeMedium returns the medium time representation of 't' for 'el_GR'
func (el *el_GR) FmtTimeMedium(t time.Time) string {

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

// FmtTimeLong returns the long time representation of 't' for 'el_GR'
func (el *el_GR) FmtTimeLong(t time.Time) string {

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

// FmtTimeFull returns the full time representation of 't' for 'el_GR'
func (el *el_GR) FmtTimeFull(t time.Time) string {

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
