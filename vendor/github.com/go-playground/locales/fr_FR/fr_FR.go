package fr_FR

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type fr_FR struct {
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

// New returns a new instance of translator for the 'fr_FR' locale
func New() locales.Translator {
	return &fr_FR{
		locale:                 "fr_FR",
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
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		percentSuffix:          " ",
		currencyPositiveSuffix: " ",
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: " )",
		monthsAbbreviated:      []string{"", "janv.", "févr.", "mars", "avr.", "mai", "juin", "juil.", "août", "sept.", "oct.", "nov.", "déc."},
		monthsNarrow:           []string{"", "J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"},
		monthsWide:             []string{"", "janvier", "février", "mars", "avril", "mai", "juin", "juillet", "août", "septembre", "octobre", "novembre", "décembre"},
		daysAbbreviated:        []string{"dim.", "lun.", "mar.", "mer.", "jeu.", "ven.", "sam."},
		daysNarrow:             []string{"D", "L", "M", "M", "J", "V", "S"},
		daysShort:              []string{"di", "lu", "ma", "me", "je", "ve", "sa"},
		daysWide:               []string{"dimanche", "lundi", "mardi", "mercredi", "jeudi", "vendredi", "samedi"},
		periodsAbbreviated:     []string{"AM", "PM"},
		periodsNarrow:          []string{"AM", "PM"},
		periodsWide:            []string{"AM", "PM"},
		erasAbbreviated:        []string{"av. J.-C.", "ap. J.-C."},
		erasNarrow:             []string{"av. J.-C.", "ap. J.-C."},
		erasWide:               []string{"avant Jésus-Christ", "après Jésus-Christ"},
		timezones:              map[string]string{"MDT": "heure d’été des Rocheuses", "HNNOMX": "heure normale du Nord-Ouest du Mexique", "OEZ": "heure normale d’Europe de l’Est", "SGT": "heure de Singapour", "HKT": "heure normale de Hong Kong", "SRT": "heure du Suriname", "ARST": "heure d’été de l’Argentine", "AST": "heure normale de l’Atlantique", "∅∅∅": "heure d’été de l’Acre", "NZST": "heure normale de la Nouvelle-Zélande", "CDT": "heure d’été du Centre", "MYT": "heure de la Malaisie", "WART": "heure normale de l’Ouest argentin", "CHADT": "heure d’été des îles Chatham", "HADT": "heure d’été d’Hawaii - Aléoutiennes", "CST": "heure normale du centre nord-américain", "BOT": "heure de Bolivie", "ACWST": "heure normale du centre-ouest de l’Australie", "HKST": "heure d’été de Hong Kong", "UYT": "heure normale de l’Uruguay", "JDT": "heure d’été du Japon", "AKDT": "heure d’été de l’Alaska", "EAT": "heure normale d’Afrique de l’Est", "CHAST": "heure normale des îles Chatham", "MST": "heure normale des Rocheuses", "WIT": "heure de l’Est indonésien", "CLT": "heure normale du Chili", "HEPM": "heure d’été de Saint-Pierre-et-Miquelon", "UYST": "heure d’été de l’Uruguay", "HECU": "heure d’été de Cuba", "AEDT": "heure d’été de l’Est de l’Australie", "NZDT": "heure d’été de la Nouvelle-Zélande", "VET": "heure du Venezuela", "HEOG": "heure d’été de l’Ouest du Groenland", "WARST": "heure d’été de l’Ouest argentin", "HENOMX": "heure d’été du Nord-Ouest du Mexique", "GMT": "heure moyenne de Greenwich", "ADT": "heure d’été de l’Atlantique", "AKST": "heure normale de l’Alaska", "HNOG": "heure normale de l’Ouest du Groenland", "HNPM": "heure normale de Saint-Pierre-et-Miquelon", "ART": "heure normale d’Argentine", "HNPMX": "heure normale du Pacifique mexicain", "SAST": "heure normale d’Afrique méridionale", "GFT": "heure de la Guyane française", "ACDT": "heure d’été du centre de l’Australie", "ChST": "heure des Chamorro", "JST": "heure normale du Japon", "EDT": "heure d’été de l’Est", "HAT": "heure d’été de Terre-Neuve", "MESZ": "heure d’été d’Europe centrale", "CAT": "heure normale d’Afrique centrale", "COST": "heure d’été de Colombie", "HAST": "heure normale d’Hawaii - Aléoutiennes", "GYT": "heure du Guyana", "WAST": "heure d’été d’Afrique de l’Ouest", "EST": "heure normale de l’Est nord-américain", "MEZ": "heure normale d’Europe centrale", "PDT": "heure d’été du Pacifique", "COT": "heure normale de Colombie", "OESZ": "heure d’été d’Europe de l’Est", "HEPMX": "heure d’été du Pacifique mexicain", "AWST": "heure normale de l’Ouest de l’Australie", "ECT": "heure de l’Équateur", "ACST": "heure normale du centre de l’Australie", "WITA": "heure du Centre indonésien", "AWDT": "heure d’été de l’Ouest de l’Australie", "WIB": "heure de l’Ouest indonésien", "CLST": "heure d’été du Chili", "HNCU": "heure normale de Cuba", "ACWDT": "heure d’été du centre-ouest de l’Australie", "TMST": "heure d’été du Turkménistan", "WAT": "heure normale d’Afrique de l’Ouest", "WESZ": "heure d’été d’Europe de l’Ouest", "HNEG": "heure normale de l’Est du Groenland", "HNT": "heure normale de Terre-Neuve", "TMT": "heure normale du Turkménistan", "AEST": "heure normale de l’Est de l’Australie", "BT": "heure du Bhoutan", "HEEG": "heure d’été de l’Est du Groenland", "LHDT": "heure d’été de Lord Howe", "PST": "heure normale du Pacifique nord-américain", "WEZ": "heure normale d’Europe de l’Ouest", "IST": "heure de l’Inde", "LHST": "heure normale de Lord Howe"},
	}
}

// Locale returns the current translators string locale
func (fr *fr_FR) Locale() string {
	return fr.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'fr_FR'
func (fr *fr_FR) PluralsCardinal() []locales.PluralRule {
	return fr.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'fr_FR'
func (fr *fr_FR) PluralsOrdinal() []locales.PluralRule {
	return fr.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'fr_FR'
func (fr *fr_FR) PluralsRange() []locales.PluralRule {
	return fr.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'fr_FR'
func (fr *fr_FR) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)

	if i == 0 || i == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'fr_FR'
func (fr *fr_FR) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'fr_FR'
func (fr *fr_FR) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := fr.CardinalPluralRule(num1, v1)
	end := fr.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (fr *fr_FR) MonthAbbreviated(month time.Month) string {
	return fr.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (fr *fr_FR) MonthsAbbreviated() []string {
	return fr.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (fr *fr_FR) MonthNarrow(month time.Month) string {
	return fr.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (fr *fr_FR) MonthsNarrow() []string {
	return fr.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (fr *fr_FR) MonthWide(month time.Month) string {
	return fr.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (fr *fr_FR) MonthsWide() []string {
	return fr.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (fr *fr_FR) WeekdayAbbreviated(weekday time.Weekday) string {
	return fr.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (fr *fr_FR) WeekdaysAbbreviated() []string {
	return fr.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (fr *fr_FR) WeekdayNarrow(weekday time.Weekday) string {
	return fr.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (fr *fr_FR) WeekdaysNarrow() []string {
	return fr.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (fr *fr_FR) WeekdayShort(weekday time.Weekday) string {
	return fr.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (fr *fr_FR) WeekdaysShort() []string {
	return fr.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (fr *fr_FR) WeekdayWide(weekday time.Weekday) string {
	return fr.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (fr *fr_FR) WeekdaysWide() []string {
	return fr.daysWide
}

// Decimal returns the decimal point of number
func (fr *fr_FR) Decimal() string {
	return fr.decimal
}

// Group returns the group of number
func (fr *fr_FR) Group() string {
	return fr.group
}

// Group returns the minus sign of number
func (fr *fr_FR) Minus() string {
	return fr.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'fr_FR' and handles both Whole and Real numbers based on 'v'
func (fr *fr_FR) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, fr.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(fr.group) - 1; j >= 0; j-- {
					b = append(b, fr.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, fr.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'fr_FR' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (fr *fr_FR) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 5
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, fr.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, fr.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, fr.percentSuffix...)

	b = append(b, fr.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'fr_FR'
func (fr *fr_FR) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := fr.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, fr.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(fr.group) - 1; j >= 0; j-- {
					b = append(b, fr.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, fr.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, fr.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, fr.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'fr_FR'
// in accounting notation.
func (fr *fr_FR) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := fr.currencies[currency]
	l := len(s) + len(symbol) + 6 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, fr.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(fr.group) - 1; j >= 0; j-- {
					b = append(b, fr.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, fr.currencyNegativePrefix[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, fr.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, fr.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, fr.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'fr_FR'
func (fr *fr_FR) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'fr_FR'
func (fr *fr_FR) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, fr.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'fr_FR'
func (fr *fr_FR) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, fr.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'fr_FR'
func (fr *fr_FR) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, fr.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, fr.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'fr_FR'
func (fr *fr_FR) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, fr.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'fr_FR'
func (fr *fr_FR) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, fr.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, fr.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'fr_FR'
func (fr *fr_FR) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, fr.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, fr.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'fr_FR'
func (fr *fr_FR) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, fr.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, fr.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := fr.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
