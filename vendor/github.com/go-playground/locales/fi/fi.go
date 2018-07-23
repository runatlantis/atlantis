package fi

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type fi struct {
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

// New returns a new instance of translator for the 'fi' locale
func New() locales.Translator {
	return &fi{
		locale:                 "fi",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         []locales.PluralRule{6},
		pluralsRange:           []locales.PluralRule{6},
		decimal:                ",",
		group:                  " ",
		minus:                  "−",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ".",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "mk", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		percentSuffix:          " ",
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "tammik.", "helmik.", "maalisk.", "huhtik.", "toukok.", "kesäk.", "heinäk.", "elok.", "syysk.", "lokak.", "marrask.", "jouluk."},
		monthsNarrow:           []string{"", "T", "H", "M", "H", "T", "K", "H", "E", "S", "L", "M", "J"},
		monthsWide:             []string{"", "tammikuuta", "helmikuuta", "maaliskuuta", "huhtikuuta", "toukokuuta", "kesäkuuta", "heinäkuuta", "elokuuta", "syyskuuta", "lokakuuta", "marraskuuta", "joulukuuta"},
		daysAbbreviated:        []string{"su", "ma", "ti", "ke", "to", "pe", "la"},
		daysNarrow:             []string{"S", "M", "T", "K", "T", "P", "L"},
		daysShort:              []string{"su", "ma", "ti", "ke", "to", "pe", "la"},
		daysWide:               []string{"sunnuntaina", "maanantaina", "tiistaina", "keskiviikkona", "torstaina", "perjantaina", "lauantaina"},
		periodsAbbreviated:     []string{"ap.", "ip."},
		periodsNarrow:          []string{"ap.", "ip."},
		periodsWide:            []string{"ap.", "ip."},
		erasAbbreviated:        []string{"eKr.", "jKr."},
		erasNarrow:             []string{"eKr", "jKr"},
		erasWide:               []string{"ennen Kristuksen syntymää", "jälkeen Kristuksen syntymän"},
		timezones:              map[string]string{"ACDT": "Keski-Australian kesäaika", "OEZ": "Itä-Euroopan normaaliaika", "AEDT": "Itä-Australian kesäaika", "WEZ": "Länsi-Euroopan normaaliaika", "HNOG": "Länsi-Grönlannin normaaliaika", "HEOG": "Länsi-Grönlannin kesäaika", "AWST": "Länsi-Australian normaaliaika", "ARST": "Argentiinan kesäaika", "NZDT": "Uuden-Seelannin kesäaika", "HADT": "Havaijin-Aleuttien kesäaika", "GMT": "Greenwichin normaaliaika", "HNPMX": "Meksikon Tyynenmeren normaaliaika", "AST": "Kanadan Atlantin normaaliaika", "JDT": "Japanin kesäaika", "CST": "Yhdysvaltain keskinen normaaliaika", "AEST": "Itä-Australian normaaliaika", "MDT": "Macaon kesäaika", "MEZ": "Keski-Euroopan normaaliaika", "HEPM": "Saint-Pierren ja Miquelonin kesäaika", "PST": "Yhdysvaltain Tyynenmeren normaaliaika", "ACWST": "Läntisen Keski-Australian normaaliaika", "IST": "Intian aika", "TMST": "Turkmenistanin kesäaika", "COT": "Kolumbian normaaliaika", "GYT": "Guyanan aika", "ChST": "Tšamorron aika", "WAT": "Länsi-Afrikan normaaliaika", "SGT": "Singaporen aika", "AKDT": "Alaskan kesäaika", "EST": "Yhdysvaltain itäinen normaaliaika", "HKST": "Hongkongin kesäaika", "HNPM": "Saint-Pierren ja Miquelonin normaaliaika", "HAST": "Havaijin-Aleuttien normaaliaika", "CHAST": "Chathamin normaaliaika", "CHADT": "Chathamin kesäaika", "GFT": "Ranskan Guayanan aika", "WAST": "Länsi-Afrikan kesäaika", "HEPMX": "Meksikon Tyynenmeren kesäaika", "BT": "Bhutanin aika", "HNEG": "Itä-Grönlannin normaaliaika", "LHDT": "Lord Howen kesäaika", "HAT": "Newfoundlandin kesäaika", "WIT": "Itä-Indonesian aika", "EAT": "Itä-Afrikan aika", "WESZ": "Länsi-Euroopan kesäaika", "JST": "Japanin normaaliaika", "NZST": "Uuden-Seelannin normaaliaika", "AKST": "Alaskan normaaliaika", "WITA": "Keski-Indonesian aika", "SRT": "Surinamen aika", "TMT": "Turkmenistanin normaaliaika", "CLST": "Chilen kesäaika", "SAST": "Etelä-Afrikan aika", "BOT": "Bolivian aika", "MESZ": "Keski-Euroopan kesäaika", "ART": "Argentiinan normaaliaika", "CDT": "Yhdysvaltain keskinen kesäaika", "MYT": "Malesian aika", "HKT": "Hongkongin normaaliaika", "HENOMX": "Luoteis-Meksikon kesäaika", "OESZ": "Itä-Euroopan kesäaika", "HECU": "Kuuban kesäaika", "HEEG": "Itä-Grönlannin kesäaika", "LHST": "Lord Howen normaaliaika", "HNT": "Newfoundlandin normaaliaika", "CAT": "Keski-Afrikan aika", "COST": "Kolumbian kesäaika", "ACST": "Keski-Australian normaaliaika", "WARST": "Länsi-Argentiinan kesäaika", "VET": "Venezuelan aika", "CLT": "Chilen normaaliaika", "HNCU": "Kuuban normaaliaika", "ACWDT": "Läntisen Keski-Australian kesäaika", "∅∅∅": "Azorien kesäaika", "HNNOMX": "Luoteis-Meksikon normaaliaika", "MST": "Macaon normaaliaika", "PDT": "Yhdysvaltain Tyynenmeren kesäaika", "AWDT": "Länsi-Australian kesäaika", "ADT": "Kanadan Atlantin kesäaika", "EDT": "Yhdysvaltain itäinen kesäaika", "WART": "Länsi-Argentiinan normaaliaika", "UYT": "Uruguayn normaaliaika", "UYST": "Uruguayn kesäaika", "WIB": "Länsi-Indonesian aika", "ECT": "Ecuadorin aika"},
	}
}

// Locale returns the current translators string locale
func (fi *fi) Locale() string {
	return fi.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'fi'
func (fi *fi) PluralsCardinal() []locales.PluralRule {
	return fi.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'fi'
func (fi *fi) PluralsOrdinal() []locales.PluralRule {
	return fi.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'fi'
func (fi *fi) PluralsRange() []locales.PluralRule {
	return fi.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'fi'
func (fi *fi) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)

	if i == 1 && v == 0 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'fi'
func (fi *fi) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'fi'
func (fi *fi) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (fi *fi) MonthAbbreviated(month time.Month) string {
	return fi.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (fi *fi) MonthsAbbreviated() []string {
	return fi.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (fi *fi) MonthNarrow(month time.Month) string {
	return fi.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (fi *fi) MonthsNarrow() []string {
	return fi.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (fi *fi) MonthWide(month time.Month) string {
	return fi.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (fi *fi) MonthsWide() []string {
	return fi.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (fi *fi) WeekdayAbbreviated(weekday time.Weekday) string {
	return fi.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (fi *fi) WeekdaysAbbreviated() []string {
	return fi.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (fi *fi) WeekdayNarrow(weekday time.Weekday) string {
	return fi.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (fi *fi) WeekdaysNarrow() []string {
	return fi.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (fi *fi) WeekdayShort(weekday time.Weekday) string {
	return fi.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (fi *fi) WeekdaysShort() []string {
	return fi.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (fi *fi) WeekdayWide(weekday time.Weekday) string {
	return fi.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (fi *fi) WeekdaysWide() []string {
	return fi.daysWide
}

// Decimal returns the decimal point of number
func (fi *fi) Decimal() string {
	return fi.decimal
}

// Group returns the group of number
func (fi *fi) Group() string {
	return fi.group
}

// Group returns the minus sign of number
func (fi *fi) Minus() string {
	return fi.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'fi' and handles both Whole and Real numbers based on 'v'
func (fi *fi) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, fi.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(fi.group) - 1; j >= 0; j-- {
					b = append(b, fi.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(fi.minus) - 1; j >= 0; j-- {
			b = append(b, fi.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'fi' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (fi *fi) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 7
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, fi.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(fi.minus) - 1; j >= 0; j-- {
			b = append(b, fi.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, fi.percentSuffix...)

	b = append(b, fi.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'fi'
func (fi *fi) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := fi.currencies[currency]
	l := len(s) + len(symbol) + 6 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, fi.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(fi.group) - 1; j >= 0; j-- {
					b = append(b, fi.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(fi.minus) - 1; j >= 0; j-- {
			b = append(b, fi.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, fi.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, fi.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'fi'
// in accounting notation.
func (fi *fi) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := fi.currencies[currency]
	l := len(s) + len(symbol) + 6 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, fi.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(fi.group) - 1; j >= 0; j-- {
					b = append(b, fi.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		for j := len(fi.minus) - 1; j >= 0; j-- {
			b = append(b, fi.minus[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, fi.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, fi.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, fi.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'fi'
func (fi *fi) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e}...)
	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'fi'
func (fi *fi) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e}...)
	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'fi'
func (fi *fi) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, fi.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'fi'
func (fi *fi) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, []byte{0x63, 0x63, 0x63, 0x63, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, fi.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'fi'
func (fi *fi) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'fi'
func (fi *fi) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'fi'
func (fi *fi) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'fi'
func (fi *fi) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := fi.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
