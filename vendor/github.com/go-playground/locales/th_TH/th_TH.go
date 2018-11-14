package th_TH

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type th_TH struct {
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

// New returns a new instance of translator for the 'th_TH' locale
func New() locales.Translator {
	return &th_TH{
		locale:                 "th_TH",
		pluralsCardinal:        []locales.PluralRule{6},
		pluralsOrdinal:         []locales.PluralRule{6},
		pluralsRange:           []locales.PluralRule{6},
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
		monthsAbbreviated:      []string{"", "ม.ค.", "ก.พ.", "มี.ค.", "เม.ย.", "พ.ค.", "มิ.ย.", "ก.ค.", "ส.ค.", "ก.ย.", "ต.ค.", "พ.ย.", "ธ.ค."},
		monthsNarrow:           []string{"", "ม.ค.", "ก.พ.", "มี.ค.", "เม.ย.", "พ.ค.", "มิ.ย.", "ก.ค.", "ส.ค.", "ก.ย.", "ต.ค.", "พ.ย.", "ธ.ค."},
		monthsWide:             []string{"", "มกราคม", "กุมภาพันธ์", "มีนาคม", "เมษายน", "พฤษภาคม", "มิถุนายน", "กรกฎาคม", "สิงหาคม", "กันยายน", "ตุลาคม", "พฤศจิกายน", "ธันวาคม"},
		daysAbbreviated:        []string{"อา.", "จ.", "อ.", "พ.", "พฤ.", "ศ.", "ส."},
		daysNarrow:             []string{"อา", "จ", "อ", "พ", "พฤ", "ศ", "ส"},
		daysShort:              []string{"อา.", "จ.", "อ.", "พ.", "พฤ.", "ศ.", "ส."},
		daysWide:               []string{"วันอาทิตย์", "วันจันทร์", "วันอังคาร", "วันพุธ", "วันพฤหัสบดี", "วันศุกร์", "วันเสาร์"},
		periodsAbbreviated:     []string{"ก่อนเที่ยง", "หลังเที่ยง"},
		periodsNarrow:          []string{"a", "p"},
		periodsWide:            []string{"ก่อนเที่ยง", "หลังเที่ยง"},
		erasAbbreviated:        []string{"ปีก่อน ค.ศ.", "ค.ศ."},
		erasNarrow:             []string{"ก่อน ค.ศ.", "ค.ศ."},
		erasWide:               []string{"ปีก่อนคริสต์ศักราช", "คริสต์ศักราช"},
		timezones:              map[string]string{"AST": "เวลามาตรฐานแอตแลนติก", "SAST": "เวลาแอฟริกาใต้", "ACWDT": "เวลาออมแสงทางตะวันตกตอนกลางของออสเตรเลีย", "LHST": "เวลามาตรฐานลอร์ดโฮว์", "HADT": "เวลาออมแสงฮาวาย-อะลูเชียน", "UYST": "เวลาฤดูร้อนอุรุกวัย", "MDT": "เวลาออมแสงแถบภูเขาในอเมริกาเหนือ", "BT": "เวลาภูฏาน", "IST": "เวลาอินเดีย", "HNPM": "เวลามาตรฐานแซงปีแยร์และมีเกอลง", "HNNOMX": "เวลามาตรฐานเม็กซิโกตะวันตกเฉียงเหนือ", "WIT": "เวลาอินโดนีเซียฝั่งตะวันออก", "WITA": "เวลาอินโดนีเซียตอนกลาง", "CLT": "เวลามาตรฐานชิลี", "CHADT": "เวลาออมแสงแชทัม", "PDT": "เวลาออมแสงแปซิฟิกในอเมริกาเหนือ", "AWST": "เวลามาตรฐานทางตะวันตกของออสเตรเลีย", "WEZ": "เวลามาตรฐานยุโรปตะวันตก", "JST": "เวลามาตรฐานญี่ปุ่น", "EDT": "เวลาออมแสงทางตะวันออกในอเมริกาเหนือ", "HEEG": "เวลาฤดูร้อนกรีนแลนด์ตะวันออก", "TMT": "เวลามาตรฐานเติร์กเมนิสถาน", "TMST": "เวลาฤดูร้อนเติร์กเมนิสถาน", "HAST": "เวลามาตรฐานฮาวาย-อะลูเชียน", "CDT": "เวลาออมแสงตอนกลางในอเมริกาเหนือ", "NZDT": "เวลาออมแสงนิวซีแลนด์", "COST": "เวลาฤดูร้อนโคลอมเบีย", "CHAST": "เวลามาตรฐานแชทัม", "ADT": "เวลาออมแสงของแอตแลนติก", "HKT": "เวลามาตรฐานฮ่องกง", "LHDT": "เวลาออมแสงลอร์ดโฮว์", "GMT": "เวลามาตรฐานกรีนิช", "GYT": "เวลากายอานา", "AEDT": "เวลาออมแสงทางตะวันออกของออสเตรเลีย", "EST": "เวลามาตรฐานทางตะวันออกในอเมริกาเหนือ", "HKST": "เวลาฤดูร้อนฮ่องกง", "WART": "เวลามาตรฐานทางตะวันตกของอาร์เจนตินา", "AEST": "เวลามาตรฐานทางตะวันออกของออสเตรเลีย", "NZST": "เวลามาตรฐานนิวซีแลนด์", "JDT": "เวลาออมแสงญี่ปุ่น", "ACWST": "เวลามาตรฐานทางตะวันตกตอนกลางของออสเตรเลีย", "VET": "เวลาเวเนซุเอลา", "ART": "เวลามาตรฐานอาร์เจนตินา", "COT": "เวลามาตรฐานโคลอมเบีย", "HNCU": "เวลามาตรฐานคิวบา", "∅∅∅": "เวลาฤดูร้อนบราซิเลีย", "BOT": "เวลาโบลิเวีย", "HEOG": "เวลาฤดูร้อนกรีนแลนด์ตะวันตก", "HENOMX": "เวลาออมแสงเม็กซิโกตะวันตกเฉียงเหนือ", "UYT": "เวลามาตรฐานอุรุกวัย", "HECU": "เวลาออมแสงของคิวบา", "HNPMX": "เวลามาตรฐานแปซิฟิกเม็กซิโก", "MYT": "เวลามาเลเซีย", "SRT": "เวลาซูรินาเม", "OEZ": "เวลามาตรฐานยุโรปตะวันออก", "AWDT": "เวลาออมแสงทางตะวันตกของออสเตรเลีย", "WAT": "เวลามาตรฐานแอฟริกาตะวันตก", "SGT": "เวลาสิงคโปร์", "CST": "เวลามาตรฐานตอนกลางในอเมริกาเหนือ", "PST": "เวลามาตรฐานแปซิฟิกในอเมริกาเหนือ", "ECT": "เวลาเอกวาดอร์", "EAT": "เวลาแอฟริกาตะวันออก", "CLST": "เวลาฤดูร้อนชิลี", "ChST": "เวลาชามอร์โร", "HEPMX": "เวลาออมแสงแปซิฟิกเม็กซิโก", "WAST": "เวลาฤดูร้อนแอฟริกาตะวันตก", "HAT": "เวลาออมแสงนิวฟันด์แลนด์", "WIB": "เวลาอินโดนีเซียฝั่งตะวันตก", "AKST": "เวลามาตรฐานอะแลสกา", "ACDT": "เวลาออมแสงทางตอนกลางของออสเตรเลีย", "MEZ": "เวลามาตรฐานยุโรปกลาง", "WARST": "เวลาฤดูร้อนทางตะวันตกของอาร์เจนตินา", "HNT": "เวลามาตรฐานนิวฟันด์แลนด์", "MST": "เวลามาตรฐานแถบภูเขาในอเมริกาเหนือ", "WESZ": "เวลาฤดูร้อนยุโรปตะวันตก", "AKDT": "เวลาออมแสงของอะแลสกา", "HNOG": "เวลามาตรฐานกรีนแลนด์ตะวันตก", "CAT": "เวลาแอฟริกากลาง", "OESZ": "เวลาฤดูร้อนยุโรปตะวันออก", "ARST": "เวลาฤดูร้อนอาร์เจนตินา", "GFT": "เวลาเฟรนช์เกียนา", "ACST": "เวลามาตรฐานทางตอนกลางของออสเตรเลีย", "HNEG": "เวลามาตรฐานกรีนแลนด์ตะวันออก", "MESZ": "เวลาฤดูร้อนยุโรปกลาง", "HEPM": "เวลาออมแสงของแซงปีแยร์และมีเกอลง"},
	}
}

// Locale returns the current translators string locale
func (th *th_TH) Locale() string {
	return th.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'th_TH'
func (th *th_TH) PluralsCardinal() []locales.PluralRule {
	return th.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'th_TH'
func (th *th_TH) PluralsOrdinal() []locales.PluralRule {
	return th.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'th_TH'
func (th *th_TH) PluralsRange() []locales.PluralRule {
	return th.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'th_TH'
func (th *th_TH) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'th_TH'
func (th *th_TH) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'th_TH'
func (th *th_TH) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (th *th_TH) MonthAbbreviated(month time.Month) string {
	return th.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (th *th_TH) MonthsAbbreviated() []string {
	return th.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (th *th_TH) MonthNarrow(month time.Month) string {
	return th.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (th *th_TH) MonthsNarrow() []string {
	return th.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (th *th_TH) MonthWide(month time.Month) string {
	return th.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (th *th_TH) MonthsWide() []string {
	return th.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (th *th_TH) WeekdayAbbreviated(weekday time.Weekday) string {
	return th.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (th *th_TH) WeekdaysAbbreviated() []string {
	return th.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (th *th_TH) WeekdayNarrow(weekday time.Weekday) string {
	return th.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (th *th_TH) WeekdaysNarrow() []string {
	return th.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (th *th_TH) WeekdayShort(weekday time.Weekday) string {
	return th.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (th *th_TH) WeekdaysShort() []string {
	return th.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (th *th_TH) WeekdayWide(weekday time.Weekday) string {
	return th.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (th *th_TH) WeekdaysWide() []string {
	return th.daysWide
}

// Decimal returns the decimal point of number
func (th *th_TH) Decimal() string {
	return th.decimal
}

// Group returns the group of number
func (th *th_TH) Group() string {
	return th.group
}

// Group returns the minus sign of number
func (th *th_TH) Minus() string {
	return th.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'th_TH' and handles both Whole and Real numbers based on 'v'
func (th *th_TH) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, th.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, th.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, th.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'th_TH' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (th *th_TH) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, th.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, th.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, th.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'th_TH'
func (th *th_TH) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := th.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, th.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, th.group[0])
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
		b = append(b, th.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, th.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'th_TH'
// in accounting notation.
func (th *th_TH) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := th.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, th.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, th.group[0])
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

		b = append(b, th.currencyNegativePrefix[0])

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
			b = append(b, th.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, th.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'th_TH'
func (th *th_TH) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'th_TH'
func (th *th_TH) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, th.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'th_TH'
func (th *th_TH) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, th.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() < 0 {
		b = append(b, th.erasAbbreviated[0]...)
	} else {
		b = append(b, th.erasAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'th_TH'
func (th *th_TH) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, th.daysWide[t.Weekday()]...)
	b = append(b, []byte{0xe0, 0xb8, 0x97, 0xe0, 0xb8, 0xb5, 0xe0, 0xb9, 0x88, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, th.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() < 0 {
		b = append(b, th.erasWide[0]...)
	} else {
		b = append(b, th.erasWide[1]...)
	}

	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'th_TH'
func (th *th_TH) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, th.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'th_TH'
func (th *th_TH) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, th.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, th.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'th_TH'
func (th *th_TH) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, []byte{0x20, 0xe0, 0xb8, 0x99, 0xe0, 0xb8, 0xb2, 0xe0, 0xb8, 0xac, 0xe0, 0xb8, 0xb4, 0xe0, 0xb8, 0x81, 0xe0, 0xb8, 0xb2, 0x20}...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20, 0xe0, 0xb8, 0x99, 0xe0, 0xb8, 0xb2, 0xe0, 0xb8, 0x97, 0xe0, 0xb8, 0xb5, 0x20}...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20, 0xe0, 0xb8, 0xa7, 0xe0, 0xb8, 0xb4, 0xe0, 0xb8, 0x99, 0xe0, 0xb8, 0xb2, 0xe0, 0xb8, 0x97, 0xe0, 0xb8, 0xb5, 0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'th_TH'
func (th *th_TH) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, []byte{0x20, 0xe0, 0xb8, 0x99, 0xe0, 0xb8, 0xb2, 0xe0, 0xb8, 0xac, 0xe0, 0xb8, 0xb4, 0xe0, 0xb8, 0x81, 0xe0, 0xb8, 0xb2, 0x20}...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20, 0xe0, 0xb8, 0x99, 0xe0, 0xb8, 0xb2, 0xe0, 0xb8, 0x97, 0xe0, 0xb8, 0xb5, 0x20}...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20, 0xe0, 0xb8, 0xa7, 0xe0, 0xb8, 0xb4, 0xe0, 0xb8, 0x99, 0xe0, 0xb8, 0xb2, 0xe0, 0xb8, 0x97, 0xe0, 0xb8, 0xb5, 0x20}...)

	tz, _ := t.Zone()

	if btz, ok := th.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
