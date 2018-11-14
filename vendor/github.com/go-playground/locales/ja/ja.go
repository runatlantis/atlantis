package ja

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type ja struct {
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

// New returns a new instance of translator for the 'ja' locale
func New() locales.Translator {
	return &ja{
		locale:                 "ja",
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
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "A$", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "元", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "₪", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "￥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "₩", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZ$", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "NT$", "TZS", "UAH", "UAK", "UGS", "UGX", "$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "₫", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "EC$", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: ")",
		monthsAbbreviated:      []string{"", "1月", "2月", "3月", "4月", "5月", "6月", "7月", "8月", "9月", "10月", "11月", "12月"},
		monthsNarrow:           []string{"", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"},
		monthsWide:             []string{"", "1月", "2月", "3月", "4月", "5月", "6月", "7月", "8月", "9月", "10月", "11月", "12月"},
		daysAbbreviated:        []string{"日", "月", "火", "水", "木", "金", "土"},
		daysNarrow:             []string{"日", "月", "火", "水", "木", "金", "土"},
		daysShort:              []string{"日", "月", "火", "水", "木", "金", "土"},
		daysWide:               []string{"日曜日", "月曜日", "火曜日", "水曜日", "木曜日", "金曜日", "土曜日"},
		periodsAbbreviated:     []string{"午前", "午後"},
		periodsNarrow:          []string{"午前", "午後"},
		periodsWide:            []string{"午前", "午後"},
		erasAbbreviated:        []string{"紀元前", "西暦"},
		erasNarrow:             []string{"BC", "AD"},
		erasWide:               []string{"紀元前", "西暦"},
		timezones:              map[string]string{"AKDT": "アラスカ夏時間", "HNOG": "グリーンランド西部標準時", "HECU": "キューバ夏時間", "IST": "インド標準時", "UYST": "ウルグアイ夏時間", "PST": "アメリカ太平洋標準時", "WIB": "インドネシア西部時間", "WAST": "西アフリカ夏時間", "HKT": "香港標準時", "MDT": "マカオ夏時間", "CLT": "チリ標準時", "JDT": "日本夏時間", "NZST": "ニュージーランド標準時", "GMT": "グリニッジ標準時", "ACST": "オーストラリア中部標準時", "ACDT": "オーストラリア中部夏時間", "ACWST": "オーストラリア中西部標準時", "MESZ": "中央ヨーロッパ夏時間", "WART": "西部アルゼンチン標準時", "HNCU": "キューバ標準時", "MYT": "マレーシア時間", "HNEG": "グリーンランド東部標準時", "HEEG": "グリーンランド東部夏時間", "COST": "コロンビア夏時間", "ChST": "チャモロ時間", "CHAST": "チャタム標準時", "UYT": "ウルグアイ標準時", "AST": "大西洋標準時", "SGT": "シンガポール標準時", "HAT": "ニューファンドランド夏時間", "WIT": "インドネシア東部時間", "OEZ": "東ヨーロッパ標準時", "HADT": "ハワイ・アリューシャン夏時間", "GFT": "仏領ギアナ時間", "BT": "ブータン時間", "HNT": "ニューファンドランド標準時", "HNPMX": "メキシコ太平洋標準時", "CST": "アメリカ中部標準時", "AEDT": "オーストラリア東部夏時間", "WAT": "西アフリカ標準時", "HEOG": "グリーンランド西部夏時間", "EDT": "アメリカ東部夏時間", "WITA": "インドネシア中部時間", "HENOMX": "メキシコ北西部夏時間", "OESZ": "東ヨーロッパ夏時間", "AKST": "アラスカ標準時", "HKST": "香港夏時間", "CLST": "チリ夏時間", "CHADT": "チャタム夏時間", "NZDT": "ニュージーランド夏時間", "AWDT": "オーストラリア西部夏時間", "AEST": "オーストラリア東部標準時", "LHDT": "ロードハウ夏時間", "VET": "ベネズエラ時間", "HNNOMX": "メキシコ北西部標準時", "EAT": "東アフリカ時間", "HAST": "ハワイ・アリューシャン標準時", "CDT": "アメリカ中部夏時間", "HNPM": "サンピエール・ミクロン標準時", "TMST": "トルクメニスタン夏時間", "CAT": "中央アフリカ時間", "COT": "コロンビア標準時", "HEPMX": "メキシコ太平洋夏時間", "MEZ": "中央ヨーロッパ標準時", "MST": "マカオ標準時", "SRT": "スリナム時間", "AWST": "オーストラリア西部標準時", "ART": "アルゼンチン標準時", "PDT": "アメリカ太平洋夏時間", "ADT": "大西洋夏時間", "ACWDT": "オーストラリア中西部夏時間", "∅∅∅": "アゾレス夏時間", "LHST": "ロードハウ標準時", "WARST": "西部アルゼンチン夏時間", "TMT": "トルクメニスタン標準時", "SAST": "南アフリカ標準時", "BOT": "ボリビア時間", "HEPM": "サンピエール・ミクロン夏時間", "ARST": "アルゼンチン夏時間", "WEZ": "西ヨーロッパ標準時", "WESZ": "西ヨーロッパ夏時間", "ECT": "エクアドル時間", "EST": "アメリカ東部標準時", "GYT": "ガイアナ時間", "JST": "日本標準時"},
	}
}

// Locale returns the current translators string locale
func (ja *ja) Locale() string {
	return ja.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'ja'
func (ja *ja) PluralsCardinal() []locales.PluralRule {
	return ja.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'ja'
func (ja *ja) PluralsOrdinal() []locales.PluralRule {
	return ja.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'ja'
func (ja *ja) PluralsRange() []locales.PluralRule {
	return ja.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'ja'
func (ja *ja) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'ja'
func (ja *ja) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'ja'
func (ja *ja) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (ja *ja) MonthAbbreviated(month time.Month) string {
	return ja.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (ja *ja) MonthsAbbreviated() []string {
	return ja.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (ja *ja) MonthNarrow(month time.Month) string {
	return ja.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (ja *ja) MonthsNarrow() []string {
	return ja.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (ja *ja) MonthWide(month time.Month) string {
	return ja.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (ja *ja) MonthsWide() []string {
	return ja.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (ja *ja) WeekdayAbbreviated(weekday time.Weekday) string {
	return ja.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (ja *ja) WeekdaysAbbreviated() []string {
	return ja.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (ja *ja) WeekdayNarrow(weekday time.Weekday) string {
	return ja.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (ja *ja) WeekdaysNarrow() []string {
	return ja.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (ja *ja) WeekdayShort(weekday time.Weekday) string {
	return ja.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (ja *ja) WeekdaysShort() []string {
	return ja.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (ja *ja) WeekdayWide(weekday time.Weekday) string {
	return ja.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (ja *ja) WeekdaysWide() []string {
	return ja.daysWide
}

// Decimal returns the decimal point of number
func (ja *ja) Decimal() string {
	return ja.decimal
}

// Group returns the group of number
func (ja *ja) Group() string {
	return ja.group
}

// Group returns the minus sign of number
func (ja *ja) Minus() string {
	return ja.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'ja' and handles both Whole and Real numbers based on 'v'
func (ja *ja) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ja.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ja.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ja.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'ja' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (ja *ja) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ja.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ja.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, ja.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'ja'
func (ja *ja) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ja.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ja.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ja.group[0])
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
		b = append(b, ja.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ja.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'ja'
// in accounting notation.
func (ja *ja) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ja.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ja.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ja.group[0])
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

		b = append(b, ja.currencyNegativePrefix[0])

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
			b = append(b, ja.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, ja.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'ja'
func (ja *ja) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2f}...)

	if t.Month() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Month()), 10)

	b = append(b, []byte{0x2f}...)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'ja'
func (ja *ja) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2f}...)

	if t.Month() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Month()), 10)

	b = append(b, []byte{0x2f}...)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'ja'
func (ja *ja) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0xe5, 0xb9, 0xb4}...)
	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0xe6, 0x9c, 0x88}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0xe6, 0x97, 0xa5}...)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'ja'
func (ja *ja) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0xe5, 0xb9, 0xb4}...)
	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0xe6, 0x9c, 0x88}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0xe6, 0x97, 0xa5}...)
	b = append(b, ja.daysWide[t.Weekday()]...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'ja'
func (ja *ja) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ja.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'ja'
func (ja *ja) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ja.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ja.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'ja'
func (ja *ja) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ja.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ja.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'ja'
func (ja *ja) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, []byte{0xe6, 0x99, 0x82}...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0xe5, 0x88, 0x86}...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0xe7, 0xa7, 0x92, 0x20}...)

	tz, _ := t.Zone()

	if btz, ok := ja.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
