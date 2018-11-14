package ko

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type ko struct {
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

// New returns a new instance of translator for the 'ko' locale
func New() locales.Translator {
	return &ko{
		locale:                 "ko",
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
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AU$", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "₪", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JP¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "₩", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZ$", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "NT$", "TZS", "UAH", "UAK", "UGS", "UGX", "US$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "₫", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "EC$", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: ")",
		monthsAbbreviated:      []string{"", "1월", "2월", "3월", "4월", "5월", "6월", "7월", "8월", "9월", "10월", "11월", "12월"},
		monthsNarrow:           []string{"", "1월", "2월", "3월", "4월", "5월", "6월", "7월", "8월", "9월", "10월", "11월", "12월"},
		monthsWide:             []string{"", "1월", "2월", "3월", "4월", "5월", "6월", "7월", "8월", "9월", "10월", "11월", "12월"},
		daysAbbreviated:        []string{"일", "월", "화", "수", "목", "금", "토"},
		daysNarrow:             []string{"일", "월", "화", "수", "목", "금", "토"},
		daysShort:              []string{"일", "월", "화", "수", "목", "금", "토"},
		daysWide:               []string{"일요일", "월요일", "화요일", "수요일", "목요일", "금요일", "토요일"},
		periodsAbbreviated:     []string{"AM", "PM"},
		periodsNarrow:          []string{"AM", "PM"},
		periodsWide:            []string{"오전", "오후"},
		erasAbbreviated:        []string{"BC", "AD"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"기원전", "서기"},
		timezones:              map[string]string{"CLST": "칠레 하계 표준시", "CHADT": "채텀 하계 표준시", "PDT": "미 태평양 하계 표준시", "SAST": "남아프리카 시간", "NZST": "뉴질랜드 표준시", "WARST": "아르헨티나 서부 하계 표준시", "COST": "콜롬비아 하계 표준시", "PST": "미 태평양 표준시", "AKST": "알래스카 표준시", "MEZ": "중부 유럽 표준시", "WAST": "서아프리카 하계 표준시", "HNEG": "그린란드 동부 표준시", "HNOG": "그린란드 서부 표준시", "HAT": "뉴펀들랜드 하계 표준시", "WAT": "서아프리카 표준시", "EST": "미 동부 표준시", "TMST": "투르크메니스탄 하계 표준시", "ART": "아르헨티나 표준시", "GMT": "그리니치 표준시", "AWDT": "오스트레일리아 서부 하계 표준시", "AEST": "오스트레일리아 동부 표준시", "AEDT": "오스트레일리아 동부 하계 표준시", "ACST": "오스트레일리아 중부 표준시", "ACWST": "오스트레일리아 중서부 표준시", "HKST": "홍콩 하계 표준시", "WART": "아르헨티나 서부 표준시", "HNNOMX": "멕시코 북서부 표준시", "AST": "대서양 표준시", "ACWDT": "오스트레일리아 중서부 하계 표준시", "WITA": "중부 인도네시아 시간", "TMT": "투르크메니스탄 표준시", "HEPMX": "멕시코 태평양 하계 표준시", "JST": "일본 표준시", "EDT": "미 동부 하계 표준시", "LHST": "로드 하우 표준시", "LHDT": "로드 하우 하계 표준시", "HENOMX": "멕시코 북서부 하계 표준시", "EAT": "동아프리카 시간", "HAST": "하와이 알류샨 표준시", "COT": "콜롬비아 표준시", "GFT": "프랑스령 가이아나 시간", "ECT": "에콰도르 시간", "∅∅∅": "아마존 하계 표준시", "AWST": "오스트레일리아 서부 표준시", "VET": "베네수엘라 시간", "UYST": "우루과이 하계 표준시", "BOT": "볼리비아 시간", "SGT": "싱가포르 표준시", "HEOG": "그린란드 서부 하계 표준시", "HEPM": "세인트피에르 미클롱 하계 표준시", "CLT": "칠레 표준시", "HNT": "뉴펀들랜드 표준시", "MST": "마카오 표준 시간", "CAT": "중앙아프리카 시간", "ARST": "아르헨티나 하계 표준시", "WESZ": "서유럽 하계 표준시", "IST": "인도 표준시", "OEZ": "동유럽 표준시", "ADT": "미 대서양 하계 표준시", "ACDT": "오스트레일리아 중부 하계 표준시", "GYT": "가이아나 시간", "HNPM": "세인트피에르 미클롱 표준시", "WEZ": "서유럽 표준시", "WIB": "서부 인도네시아 시간", "WIT": "동부 인도네시아 시간", "OESZ": "동유럽 하계 표준시", "ChST": "차모로 시간", "HECU": "쿠바 하계 표준시", "CST": "미 중부 표준시", "CDT": "미 중부 하계 표준시", "BT": "부탄 시간", "AKDT": "알래스카 하계 표준시", "HEEG": "그린란드 동부 하계 표준시", "MDT": "마카오 하계 표준시", "MYT": "말레이시아 시간", "HKT": "홍콩 표준시", "JDT": "일본 하계 표준시", "NZDT": "뉴질랜드 하계 표준시", "SRT": "수리남 시간", "HADT": "하와이 알류샨 하계 표준시", "UYT": "우루과이 표준시", "CHAST": "채텀 표준시", "HNCU": "쿠바 표준시", "HNPMX": "멕시코 태평양 표준시", "MESZ": "중부 유럽 하계 표준시"},
	}
}

// Locale returns the current translators string locale
func (ko *ko) Locale() string {
	return ko.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'ko'
func (ko *ko) PluralsCardinal() []locales.PluralRule {
	return ko.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'ko'
func (ko *ko) PluralsOrdinal() []locales.PluralRule {
	return ko.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'ko'
func (ko *ko) PluralsRange() []locales.PluralRule {
	return ko.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'ko'
func (ko *ko) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'ko'
func (ko *ko) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'ko'
func (ko *ko) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (ko *ko) MonthAbbreviated(month time.Month) string {
	return ko.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (ko *ko) MonthsAbbreviated() []string {
	return ko.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (ko *ko) MonthNarrow(month time.Month) string {
	return ko.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (ko *ko) MonthsNarrow() []string {
	return ko.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (ko *ko) MonthWide(month time.Month) string {
	return ko.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (ko *ko) MonthsWide() []string {
	return ko.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (ko *ko) WeekdayAbbreviated(weekday time.Weekday) string {
	return ko.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (ko *ko) WeekdaysAbbreviated() []string {
	return ko.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (ko *ko) WeekdayNarrow(weekday time.Weekday) string {
	return ko.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (ko *ko) WeekdaysNarrow() []string {
	return ko.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (ko *ko) WeekdayShort(weekday time.Weekday) string {
	return ko.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (ko *ko) WeekdaysShort() []string {
	return ko.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (ko *ko) WeekdayWide(weekday time.Weekday) string {
	return ko.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (ko *ko) WeekdaysWide() []string {
	return ko.daysWide
}

// Decimal returns the decimal point of number
func (ko *ko) Decimal() string {
	return ko.decimal
}

// Group returns the group of number
func (ko *ko) Group() string {
	return ko.group
}

// Group returns the minus sign of number
func (ko *ko) Minus() string {
	return ko.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'ko' and handles both Whole and Real numbers based on 'v'
func (ko *ko) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ko.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ko.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ko.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'ko' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (ko *ko) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ko.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ko.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, ko.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'ko'
func (ko *ko) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ko.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ko.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ko.group[0])
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
		b = append(b, ko.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ko.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'ko'
// in accounting notation.
func (ko *ko) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ko.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ko.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ko.group[0])
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

		b = append(b, ko.currencyNegativePrefix[0])

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
			b = append(b, ko.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, ko.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'ko'
func (ko *ko) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 9 {
		b = append(b, strconv.Itoa(t.Year())[2:]...)
	} else {
		b = append(b, strconv.Itoa(t.Year())[1:]...)
	}

	b = append(b, []byte{0x2e, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e}...)

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'ko'
func (ko *ko) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2e, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e}...)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'ko'
func (ko *ko) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0xeb, 0x85, 0x84, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0xec, 0x9b, 0x94, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0xec, 0x9d, 0xbc}...)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'ko'
func (ko *ko) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0xeb, 0x85, 0x84, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0xec, 0x9b, 0x94, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0xec, 0x9d, 0xbc, 0x20}...)
	b = append(b, ko.daysWide[t.Weekday()]...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'ko'
func (ko *ko) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 12 {
		b = append(b, ko.periodsAbbreviated[0]...)
	} else {
		b = append(b, ko.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ko.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'ko'
func (ko *ko) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 12 {
		b = append(b, ko.periodsAbbreviated[0]...)
	} else {
		b = append(b, ko.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ko.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ko.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'ko'
func (ko *ko) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 12 {
		b = append(b, ko.periodsAbbreviated[0]...)
	} else {
		b = append(b, ko.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, []byte{0xec, 0x8b, 0x9c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0xeb, 0xb6, 0x84, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0xec, 0xb4, 0x88, 0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'ko'
func (ko *ko) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 12 {
		b = append(b, ko.periodsAbbreviated[0]...)
	} else {
		b = append(b, ko.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, []byte{0xec, 0x8b, 0x9c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0xeb, 0xb6, 0x84, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0xec, 0xb4, 0x88, 0x20}...)

	tz, _ := t.Zone()

	if btz, ok := ko.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
