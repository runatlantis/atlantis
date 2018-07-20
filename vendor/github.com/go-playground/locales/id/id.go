package id

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type id struct {
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

// New returns a new instance of translator for the 'id' locale
func New() locales.Translator {
	return &id{
		locale:             "id",
		pluralsCardinal:    []locales.PluralRule{6},
		pluralsOrdinal:     []locales.PluralRule{6},
		pluralsRange:       []locales.PluralRule{6},
		decimal:            ",",
		group:              ".",
		minus:              "-",
		percent:            "%",
		perMille:           "‰",
		timeSeparator:      ".",
		inifinity:          "∞",
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AU$", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "Rp", "IEP", "ILP", "ILR", "₪", "Rs", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JP¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "₩", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZ$", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "฿", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "NT$", "TZS", "UAH", "UAK", "UGS", "UGX", "US$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "₫", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "EC$", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		monthsAbbreviated:  []string{"", "Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agt", "Sep", "Okt", "Nov", "Des"},
		monthsNarrow:       []string{"", "J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"},
		monthsWide:         []string{"", "Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"},
		daysAbbreviated:    []string{"Min", "Sen", "Sel", "Rab", "Kam", "Jum", "Sab"},
		daysNarrow:         []string{"M", "S", "S", "R", "K", "J", "S"},
		daysShort:          []string{"Min", "Sen", "Sel", "Rab", "Kam", "Jum", "Sab"},
		daysWide:           []string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"},
		periodsAbbreviated: []string{"AM", "PM"},
		periodsNarrow:      []string{"AM", "PM"},
		periodsWide:        []string{"AM", "PM"},
		erasAbbreviated:    []string{"SM", "M"},
		erasNarrow:         []string{"SM", "M"},
		erasWide:           []string{"Sebelum Masehi", "Masehi"},
		timezones:          map[string]string{"CLT": "Waktu Standar Cile", "HAST": "Waktu Standar Hawaii-Aleutian", "PDT": "Waktu Musim Panas Pasifik", "BOT": "Waktu Bolivia", "BT": "Waktu Bhutan", "SGT": "Waktu Standar Singapura", "EDT": "Waktu Musim Panas Timur", "HKST": "Waktu Musim Panas Hong Kong", "ACWDT": "Waktu Musim Panas Barat Tengah Australia", "SRT": "Waktu Suriname", "HNCU": "Waktu Standar Kuba", "CAT": "Waktu Afrika Tengah", "EAT": "Waktu Afrika Timur", "AWST": "Waktu Standar Barat Australia", "HNOG": "Waktu Standar Greenland Barat", "MESZ": "Waktu Musim Panas Eropa Tengah", "WART": "Waktu Standar Argentina Bagian Barat", "LHST": "Waktu Standar Lord Howe", "WIT": "Waktu Indonesia Timur", "WIB": "Waktu Indonesia Barat", "SAST": "Waktu Standar Afrika Selatan", "GFT": "Waktu Guyana Prancis", "IST": "Waktu India", "AWDT": "Waktu Musim Panas Barat Australia", "CDT": "Waktu Musim Panas Tengah", "ADT": "Waktu Musim Panas Atlantik", "WESZ": "Waktu Musim Panas Eropa Barat", "ECT": "Waktu Ekuador", "HEEG": "Waktu Musim Panas Greenland Timur", "COT": "Waktu Standar Kolombia", "HEPMX": "Waktu Musim Panas Pasifik Meksiko", "AEDT": "Waktu Musim Panas Timur Australia", "AST": "Waktu Standar Atlantik", "AKDT": "Waktu Musim Panas Alaska", "EST": "Waktu Standar Timur", "HAT": "Waktu Musim Panas Newfoundland", "HENOMX": "Waktu Musim Panas Meksiko Barat Laut", "COST": "Waktu Musim Panas Kolombia", "WARST": "Waktu Musim Panas Argentina Bagian Barat", "HNPM": "Waktu Standar Saint Pierre dan Miquelon", "MST": "Waktu Standar Makau", "UYT": "Waktu Standar Uruguay", "WAT": "Waktu Standar Afrika Barat", "TMST": "Waktu Musim Panas Turkmenistan", "HNPMX": "Waktu Standar Pasifik Meksiko", "WEZ": "Waktu Standar Eropa Barat", "NZST": "Waktu Standar Selandia Baru", "ACWST": "Waktu Standar Barat Tengah Australia", "OESZ": "Waktu Musim Panas Eropa Timur", "HADT": "Waktu Musim Panas Hawaii-Aleutian", "CHAST": "Waktu Standar Chatham", "HEPM": "Waktu Musim Panas Saint Pierre dan Miquelon", "TMT": "Waktu Standar Turkmenistan", "ART": "Waktu Standar Argentina", "ARST": "Waktu Musim Panas Argentina", "HECU": "Waktu Musim Panas Kuba", "MYT": "Waktu Malaysia", "JST": "Waktu Standar Jepang", "NZDT": "Waktu Musim Panas Selandia Baru", "ACDT": "Waktu Musim Panas Tengah Australia", "∅∅∅": "Waktu Musim Panas Azores", "AEST": "Waktu Standar Timur Australia", "JDT": "Waktu Musim Panas Jepang", "HNEG": "Waktu Standar Greenland Timur", "ACST": "Waktu Standar Tengah Australia", "UYST": "Waktu Musim Panas Uruguay", "GMT": "Greenwich Mean Time", "WITA": "Waktu Indonesia Tengah", "CLST": "Waktu Musim Panas Cile", "PST": "Waktu Standar Pasifik", "WAST": "Waktu Musim Panas Afrika Barat", "AKST": "Waktu Standar Alaska", "MEZ": "Waktu Standar Eropa Tengah", "HKT": "Waktu Standar Hong Kong", "LHDT": "Waktu Musim Panas Lord Howe", "VET": "Waktu Venezuela", "CHADT": "Waktu Musim Panas Chatham", "CST": "Waktu Standar Tengah", "HEOG": "Waktu Musim Panas Greenland Barat", "HNNOMX": "Waktu Standar Meksiko Barat Laut", "MDT": "Waktu Musim Panas Makau", "GYT": "Waktu Guyana", "ChST": "Waktu Standar Chamorro", "HNT": "Waktu Standar Newfoundland", "OEZ": "Waktu Standar Eropa Timur"},
	}
}

// Locale returns the current translators string locale
func (id *id) Locale() string {
	return id.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'id'
func (id *id) PluralsCardinal() []locales.PluralRule {
	return id.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'id'
func (id *id) PluralsOrdinal() []locales.PluralRule {
	return id.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'id'
func (id *id) PluralsRange() []locales.PluralRule {
	return id.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'id'
func (id *id) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'id'
func (id *id) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'id'
func (id *id) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (id *id) MonthAbbreviated(month time.Month) string {
	return id.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (id *id) MonthsAbbreviated() []string {
	return id.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (id *id) MonthNarrow(month time.Month) string {
	return id.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (id *id) MonthsNarrow() []string {
	return id.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (id *id) MonthWide(month time.Month) string {
	return id.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (id *id) MonthsWide() []string {
	return id.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (id *id) WeekdayAbbreviated(weekday time.Weekday) string {
	return id.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (id *id) WeekdaysAbbreviated() []string {
	return id.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (id *id) WeekdayNarrow(weekday time.Weekday) string {
	return id.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (id *id) WeekdaysNarrow() []string {
	return id.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (id *id) WeekdayShort(weekday time.Weekday) string {
	return id.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (id *id) WeekdaysShort() []string {
	return id.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (id *id) WeekdayWide(weekday time.Weekday) string {
	return id.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (id *id) WeekdaysWide() []string {
	return id.daysWide
}

// Decimal returns the decimal point of number
func (id *id) Decimal() string {
	return id.decimal
}

// Group returns the group of number
func (id *id) Group() string {
	return id.group
}

// Group returns the minus sign of number
func (id *id) Minus() string {
	return id.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'id' and handles both Whole and Real numbers based on 'v'
func (id *id) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, id.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, id.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, id.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'id' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (id *id) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, id.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, id.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, id.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'id'
func (id *id) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := id.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, id.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, id.group[0])
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
		b = append(b, id.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, id.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'id'
// in accounting notation.
func (id *id) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := id.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, id.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, id.group[0])
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

		b = append(b, id.minus[0])

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
			b = append(b, id.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'id'
func (id *id) FmtDateShort(t time.Time) string {

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

	if t.Year() > 9 {
		b = append(b, strconv.Itoa(t.Year())[2:]...)
	} else {
		b = append(b, strconv.Itoa(t.Year())[1:]...)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'id'
func (id *id) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, id.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'id'
func (id *id) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, id.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'id'
func (id *id) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, id.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, id.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'id'
func (id *id) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'id'
func (id *id) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

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

// FmtTimeLong returns the long time representation of 't' for 'id'
func (id *id) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

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

// FmtTimeFull returns the full time representation of 't' for 'id'
func (id *id) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

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

	if btz, ok := id.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
