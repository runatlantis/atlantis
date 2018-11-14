package lo

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type lo struct {
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

// New returns a new instance of translator for the 'lo' locale
func New() locales.Translator {
	return &lo{
		locale:                 "lo",
		pluralsCardinal:        []locales.PluralRule{6},
		pluralsOrdinal:         []locales.PluralRule{2, 6},
		pluralsRange:           []locales.PluralRule{6},
		decimal:                ",",
		group:                  ".",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "A$", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "₪", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JP¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "₩", "KWD", "KYD", "KZT", "₭", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZ$", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "฿", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "NT$", "TZS", "UAH", "UAK", "UGS", "UGX", "US$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "₫", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "EC$", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyNegativePrefix: "-",
		monthsAbbreviated:      []string{"", "ມ.ກ.", "ກ.ພ.", "ມ.ນ.", "ມ.ສ.", "ພ.ພ.", "ມິ.ຖ.", "ກ.ລ.", "ສ.ຫ.", "ກ.ຍ.", "ຕ.ລ.", "ພ.ຈ.", "ທ.ວ."},
		monthsNarrow:           []string{"", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"},
		monthsWide:             []string{"", "ມັງກອນ", "ກຸມພາ", "ມີນາ", "ເມສາ", "ພຶດສະພາ", "ມິຖຸນາ", "ກໍລະກົດ", "ສິງຫາ", "ກັນຍາ", "ຕຸລາ", "ພະຈິກ", "ທັນວາ"},
		daysAbbreviated:        []string{"ອາທິດ", "ຈັນ", "ອັງຄານ", "ພຸດ", "ພະຫັດ", "ສຸກ", "ເສົາ"},
		daysNarrow:             []string{"ອາ", "ຈ", "ອ", "ພ", "ພຫ", "ສຸ", "ສ"},
		daysShort:              []string{"ອາ.", "ຈ.", "ອ.", "ພ.", "ພຫ.", "ສຸ.", "ສ."},
		daysWide:               []string{"ວັນອາທິດ", "ວັນຈັນ", "ວັນອັງຄານ", "ວັນພຸດ", "ວັນພະຫັດ", "ວັນສຸກ", "ວັນເສົາ"},
		periodsAbbreviated:     []string{"ກ່ອນທ່ຽງ", "ຫຼັງທ່ຽງ"},
		periodsNarrow:          []string{"ກທ", "ຫຼທ"},
		periodsWide:            []string{"ກ່ອນທ່ຽງ", "ຫຼັງທ່ຽງ"},
		erasAbbreviated:        []string{"ກ່ອນ ຄ.ສ.", "ຄ.ສ."},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"ກ່ອນຄຣິດສັກກະລາດ", "ຄຣິດສັກກະລາດ"},
		timezones:              map[string]string{"WAST": "ເວ\u200bລາ\u200bລະ\u200bດູ\u200bຮ້ອນ\u200bອາ\u200bຟຣິ\u200bກາ\u200bຕາ\u200bເວັນ\u200bຕົກ", "EDT": "ເວລາກາງເວັນຕາເວັນອອກ", "HKST": "\u200bເວ\u200bລາ\u200bລະ\u200bດູ\u200bຮ້ອນ\u200bຮອງ\u200bກົງ", "ChST": "ເວ\u200bລາ\u200bຈາ\u200bໂມ\u200bໂຣ", "AWDT": "ເວ\u200bລາ\u200bຕອນ\u200bທ່ຽງ\u200bອອສ\u200bເຕຣ\u200bລຽນ\u200bຕາ\u200bເວັນ\u200bຕົກ", "SAST": "ເວ\u200bລາ\u200bອາ\u200bຟຣິ\u200bກາ\u200bໃຕ້", "BT": "ເວ\u200bລາ\u200bພູ\u200bຖານ", "SGT": "ເວ\u200bລາ\u200bສິງ\u200bກະ\u200bໂປ", "HEEG": "ເວລາລະດູຮ້ອນກຣີນແລນຕາເວັນອອກ", "ART": "\u200bເວ\u200bລາ\u200bມາດ\u200bຕະ\u200bຖານອາ\u200bເຈນ\u200bທິ\u200bນາ", "HNPMX": "ເວລາມາດຕະຖານແປຊິຟິກເມັກຊິກັນ", "∅∅∅": "ເວ\u200bລາ\u200bລະ\u200bດູ\u200bຮ້ອນອາ\u200bໂຊ\u200bເຣ\u200bສ", "LHST": "ເວ\u200bລາ\u200bມາດ\u200bຕະ\u200bຖານ\u200bລອດ\u200bເຮົາ", "SRT": "ເວ\u200bລາ\u200bຊຸ\u200bຣິ\u200bນາມ", "HADT": "ເວລາຕອນທ່ຽງຮາວາຍ-ເອລູທຽນ", "ARST": "\u200bເວ\u200bລາ\u200bລະ\u200bດູ\u200bຮ້ອນ\u200bອາ\u200bເຈນ\u200bທິ\u200bນາ", "TMST": "ເວລາລະດູຮ້ອນຕວກເມນິສຖານ", "OEZ": "ເວ\u200bລາ\u200bມາ\u200bດ\u200bຕະ\u200bຖານ\u200bຢູ\u200bໂຣບ\u200bຕາ\u200bເວັນ\u200bອອກ", "GYT": "ເວລາກາຍອານາ", "BOT": "ເວ\u200bລາ\u200bໂບ\u200bລິ\u200bເວຍ", "GFT": "ເວ\u200bລາ\u200bເຟ\u200bຣນ\u200bຊ໌\u200bເກຍ\u200bນາ", "HNOG": "ເວລາມາດຕະຖານກຣີນແລນຕາເວັນຕົກ", "HNPM": "\u200bເວ\u200bລາມາດ\u200bຕະ\u200bຖານເຊນ\u200bປີ\u200bແອ ແລະ\u200bມິ\u200bກົວ\u200bລອນ", "CLST": "ເວ\u200bລາ\u200bລະ\u200bດູ\u200bຮ້ອນຊິ\u200bລີ", "HECU": "ເວລາກາງເວັນຄິວບາ", "PDT": "ເວລາກາງເວັນແປຊິຟິກ", "AWST": "ເວ\u200bລາ\u200bມາ\u200bດ\u200bຕະ\u200bຖານອອສ\u200bເຕຣ\u200bລຽນ\u200bຕາ\u200bເວັນ\u200bຕົກ", "WESZ": "ເວ\u200bລາ\u200bລະ\u200bດູ\u200bຮ້ອນຢູ\u200bໂຣບ\u200bຕາ\u200bເວັນ\u200bຕົກ", "MESZ": "\u200bເວ\u200bລາ\u200bລະ\u200bດູ\u200bຮ້ອນ\u200bຢູ\u200bໂຣບ\u200bກາງ", "HAST": "ເວລາມາດຕະຖານຮາວາຍ-ເອລູທຽນ", "UYST": "ເວ\u200bລາ\u200bລະ\u200bດູ\u200bຮ້ອນ\u200bອູ\u200bຣູ\u200bກວຍ", "EST": "ເວລາມາດຕະຖານຕາເວັນອອກ", "IST": "ເວລາ ອິນເດຍ", "CHAST": "ເວ\u200bລາ\u200bມາດ\u200bຕະ\u200bຖານ\u200bຊາ\u200bທາມ", "AEST": "ເວ\u200bລາ\u200bມາດຕະຖານ\u200b\u200b\u200bອອສ\u200bເຕຣ\u200bລຽນ\u200bຕາ\u200bເວັນ\u200bອອກ", "TMT": "ເວລາມາດຕະຖານຕວກເມນິສຖານ", "HAT": "ເວລາກາງເວັນນິວຟາວແລນ", "CAT": "ເວ\u200bລາ\u200bອາ\u200bຟຣິ\u200bກາ\u200bກາງ", "WEZ": "ເວ\u200bລາ\u200bມາດ\u200bຕະ\u200bຖານຢູ\u200bໂຣບ\u200bຕາ\u200bເວັນ\u200bຕົກ", "ACWDT": "ເວ\u200bລາ\u200bຕອນ\u200bທ່ຽງ\u200bອອສ\u200bເຕຣ\u200bລຽນ\u200bກາງ\u200bຕາ\u200bເວັນ\u200bຕົກ", "HKT": "ເວ\u200bລາ\u200bມາດ\u200bຕະ\u200bຖານ\u200bຮອງ\u200bກົງ", "LHDT": "\u200bເວ\u200bລ\u200bສາ\u200bຕອນ\u200b\u200bທ່ຽງ\u200bລອດ\u200bເຮົາ\u200b", "WART": "ເວ\u200bລາ\u200bມາດ\u200bຕະ\u200bຖານເວ\u200bສ\u200bເທິນອາ\u200bເຈນ\u200bທິ\u200bນາ", "HNT": "ເວ\u200bລາ\u200bມາດ\u200bຕະ\u200bຖານ\u200bນິວ\u200bຟາວ\u200bແລນ", "ACWST": "ເວ\u200bລາ\u200bມາດ\u200bຕະ\u200bຖານອອສ\u200bເຕຣ\u200bລຽນ\u200bກາງ\u200bຕາ\u200bເວັນ\u200bຕົກ", "COT": "ເວລາມາດຕະຖານໂຄລຳເບຍ", "WIB": "ເວ\u200bລາ\u200bອິນ\u200bໂດ\u200bເນ\u200bເຊຍ\u200bຕາ\u200bເວັນ\u200bຕົກ", "AEDT": "ເວ\u200bລາ\u200bຕອນ\u200bທ່ຽງ\u200bອອສ\u200bເຕຣ\u200bລຽນ\u200bຕາ\u200bເວັນ\u200bອອກ", "NZST": "ເວ\u200bລາ\u200bມາດ\u200bຕະ\u200bຖານນິວ\u200bຊີ\u200bແລນ", "AKDT": "ເວລາກາງເວັນອະແລສກາ", "ECT": "ເວ\u200bລາ\u200bເອ\u200bກົວ\u200bດໍ", "HEPM": "\u200bເວ\u200bລາຕອນ\u200bທ່ຽງເຊນ\u200bປີ\u200bແອ ແລະ\u200bມິ\u200bກົວ\u200bລອນ", "HNNOMX": "\u200bເວ\u200bລາ\u200bມາດ\u200bຕະ\u200bຖານນອດ\u200bເວ\u200bສ\u200bເມັກ\u200bຊິ\u200bໂກ", "GMT": "ເວ\u200bລາກຣີນ\u200bວິ\u200bຊ", "HEPMX": "ເວລາກາງເວັນແປຊິຟິກເມັກຊິກັນ", "WITA": "ເວ\u200bລາ\u200bອິນ\u200bໂດ\u200bເນ\u200bເຊຍ\u200bກາງ", "MDT": "ເວລາລະດູຮ້ອນມາເກົາ", "EAT": "ເວ\u200bລາ\u200bອາ\u200bຟຣິ\u200bກາ\u200bຕາ\u200bເວັນ\u200bອອກ", "CLT": "ເວ\u200bລາ\u200bມາດ\u200bຕະ\u200bຖານຊິ\u200bລີ", "COST": "ເວລາລະດູຮ້ອນໂຄລໍາເບຍ", "CDT": "ເວລາກາງເວັນກາງ", "JST": "ເວ\u200bລາ\u200bມາດ\u200bຕະ\u200bຖານ\u200bຍີ່\u200bປຸ່ນ", "ACDT": "ເວ\u200bລາ\u200bຕອນ\u200bທ່ຽງ\u200bອອສ\u200bເຕຣ\u200bເລຍ\u200bກາງ", "HNEG": "ເວລາມາດຕະຖານຕາເວັນອອກກຣີນແລນ", "HENOMX": "ເວລາກາງເວັນເມັກຊິກັນນອດເວສ", "CHADT": "ເວ\u200bລາ\u200bຕອນ\u200bທ່ຽງ\u200bຊາ\u200bທາມ", "JDT": "ເວ\u200bລາ\u200bຕອນ\u200bທ່ຽງ\u200bຍີ່\u200bປຸ່ນ", "HEOG": "ເວລາຕອນທ່ຽງກຣີນແລນຕາເວັນຕົກ", "MEZ": "ເວ\u200bລາ\u200bມາດ\u200bຕະ\u200bຖານ\u200bຢູ\u200bໂຣບກາງ", "MST": "ເວລາມາດຕະຖານມາເກົາ", "WAT": "ເວ\u200bລາ\u200bມາດ\u200bຕະ\u200bຖານ\u200bອາ\u200bຟຣິ\u200bກາ\u200bຕາ\u200bເວັນ\u200bຕົກ", "NZDT": "ເວ\u200bລາ\u200bຕອນ\u200bທ່ຽງ\u200bນິວ\u200bຊີ\u200bແລນ", "AST": "ເວລາມາດຕະຖານຂອງອາແລນຕິກ", "ADT": "ເວລາກາງເວັນຂອງອາແລນຕິກ", "AKST": "ເວລາມາດຕະຖານອະແລສກາ", "WIT": "ເວ\u200bລາ\u200bອິນ\u200bໂດ\u200bເນ\u200bເຊຍ\u200bຕາ\u200bເວັນ\u200bອອກ", "UYT": "ເວ\u200bລາ\u200bມາດ\u200bຕະ\u200bຖານ\u200bອູ\u200bຣູ\u200bກວຍ", "HNCU": "ເວລາມາດຕະຖານຂອງຄິວບາ", "CST": "ເວລາມາດຕະຖານກາງ", "PST": "ເວລາມາດຕະຖານແປຊິຟິກ", "MYT": "ເວ\u200bລາ\u200bມາ\u200bເລ\u200bເຊຍ", "ACST": "ເວ\u200bລາມາດ\u200bຕະ\u200bຖານອອ\u200bສ\u200bເຕຣ\u200bເລຍ\u200bກ\u200bາງ", "WARST": "ເວ\u200bລາ\u200bລະ\u200bດູ\u200bຮ້ອນເວ\u200bສ\u200bເທິນອາ\u200bເຈນ\u200bທິ\u200bນາ", "VET": "ເວ\u200bລາ\u200bເວ\u200bເນ\u200bຊູ\u200bເອ\u200bລາ", "OESZ": "ເວ\u200bລາ\u200bລະ\u200bດູ\u200bຮ້ອນຢູ\u200bໂຣບ\u200bຕາ\u200bເວັນ\u200bອອກ"},
	}
}

// Locale returns the current translators string locale
func (lo *lo) Locale() string {
	return lo.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'lo'
func (lo *lo) PluralsCardinal() []locales.PluralRule {
	return lo.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'lo'
func (lo *lo) PluralsOrdinal() []locales.PluralRule {
	return lo.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'lo'
func (lo *lo) PluralsRange() []locales.PluralRule {
	return lo.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'lo'
func (lo *lo) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'lo'
func (lo *lo) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'lo'
func (lo *lo) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (lo *lo) MonthAbbreviated(month time.Month) string {
	return lo.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (lo *lo) MonthsAbbreviated() []string {
	return lo.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (lo *lo) MonthNarrow(month time.Month) string {
	return lo.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (lo *lo) MonthsNarrow() []string {
	return lo.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (lo *lo) MonthWide(month time.Month) string {
	return lo.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (lo *lo) MonthsWide() []string {
	return lo.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (lo *lo) WeekdayAbbreviated(weekday time.Weekday) string {
	return lo.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (lo *lo) WeekdaysAbbreviated() []string {
	return lo.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (lo *lo) WeekdayNarrow(weekday time.Weekday) string {
	return lo.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (lo *lo) WeekdaysNarrow() []string {
	return lo.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (lo *lo) WeekdayShort(weekday time.Weekday) string {
	return lo.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (lo *lo) WeekdaysShort() []string {
	return lo.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (lo *lo) WeekdayWide(weekday time.Weekday) string {
	return lo.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (lo *lo) WeekdaysWide() []string {
	return lo.daysWide
}

// Decimal returns the decimal point of number
func (lo *lo) Decimal() string {
	return lo.decimal
}

// Group returns the group of number
func (lo *lo) Group() string {
	return lo.group
}

// Group returns the minus sign of number
func (lo *lo) Minus() string {
	return lo.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'lo' and handles both Whole and Real numbers based on 'v'
func (lo *lo) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'lo' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (lo *lo) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, lo.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, lo.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, lo.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'lo'
func (lo *lo) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := lo.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, lo.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, lo.group[0])
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
		b = append(b, lo.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, lo.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'lo'
// in accounting notation.
func (lo *lo) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := lo.currencies[currency]
	l := len(s) + len(symbol) + 3 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, lo.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, lo.group[0])
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

		b = append(b, lo.currencyNegativePrefix[0])

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
			b = append(b, lo.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'lo'
func (lo *lo) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2f}...)
	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0x2f}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'lo'
func (lo *lo) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, lo.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'lo'
func (lo *lo) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, lo.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'lo'
func (lo *lo) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, lo.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20, 0xe0, 0xba, 0x97, 0xe0, 0xba, 0xb5, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, lo.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() < 0 {
		b = append(b, lo.erasWide[0]...)
	} else {
		b = append(b, lo.erasWide[1]...)
	}

	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'lo'
func (lo *lo) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, lo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'lo'
func (lo *lo) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, lo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, lo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'lo'
func (lo *lo) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, []byte{0x20, 0xe0, 0xbb, 0x82, 0xe0, 0xba, 0xa1, 0xe0, 0xba, 0x87, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20, 0xe0, 0xba, 0x99, 0xe0, 0xba, 0xb2, 0xe0, 0xba, 0x97, 0xe0, 0xba, 0xb5, 0x20}...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20, 0xe0, 0xba, 0xa7, 0xe0, 0xba, 0xb4, 0xe0, 0xba, 0x99, 0xe0, 0xba, 0xb2, 0xe0, 0xba, 0x97, 0xe0, 0xba, 0xb5, 0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'lo'
func (lo *lo) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, []byte{0x20, 0xe0, 0xbb, 0x82, 0xe0, 0xba, 0xa1, 0xe0, 0xba, 0x87, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20, 0xe0, 0xba, 0x99, 0xe0, 0xba, 0xb2, 0xe0, 0xba, 0x97, 0xe0, 0xba, 0xb5, 0x20}...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20, 0xe0, 0xba, 0xa7, 0xe0, 0xba, 0xb4, 0xe0, 0xba, 0x99, 0xe0, 0xba, 0xb2, 0xe0, 0xba, 0x97, 0xe0, 0xba, 0xb5, 0x20}...)

	tz, _ := t.Zone()

	if btz, ok := lo.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
