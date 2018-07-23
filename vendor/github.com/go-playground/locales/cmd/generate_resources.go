package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/go-playground/locales"

	"golang.org/x/text/unicode/cldr"

	"text/template"
)

const (
	locDir      = "../%s"
	locFilename = locDir + "/%s.go"
)

var (
	tfuncs = template.FuncMap{
		"is_multibyte": func(s string) bool {
			return len([]byte(s)) > 1
		},
		"reverse_bytes": func(s string) string {
			b := make([]byte, 0, 8)

			for j := len(s) - 1; j >= 0; j-- {
				b = append(b, s[j])
			}

			return fmt.Sprintf("%#v", b)
		},
		"byte_count": func(s ...string) string {
			var count int

			for i := 0; i < len(s); i++ {
				count += len([]byte(s[i]))
			}

			return strconv.Itoa(count)
		},
	}
	prVarFuncs = map[string]string{
		"n": "n := math.Abs(num)\n",
		"i": "i := int64(n)\n",
		// "v": "v := ...", // inherently available as argument
		"w": "w := locales.W(n, v)\n",
		"f": "f := locales.F(n, v)\n",
		"t": "t := locales.T(n, v)\n",
	}

	translators            = make(map[string]*translator)
	baseTranslators        = make(map[string]*translator)
	globalCurrenciesMap    = make(map[string]struct{}) // ["USD"] = "$" currency code, just all currencies for mapping to enum
	globCurrencyIdxMap     = make(map[string]int)      // ["USD"] = 0
	globalCurrencies       = make([]string, 0, 100)    // array of currency codes index maps to enum
	tmpl                   *template.Template
	nModRegex              = regexp.MustCompile("(n%[0-9]+)")
	iModRegex              = regexp.MustCompile("(i%[0-9]+)")
	wModRegex              = regexp.MustCompile("(w%[0-9]+)")
	fModRegex              = regexp.MustCompile("(f%[0-9]+)")
	tModRegex              = regexp.MustCompile("(t%[0-9]+)")
	groupLenRegex          = regexp.MustCompile(",([0-9#]+)\\.")
	groupLenPercentRegex   = regexp.MustCompile(",([0-9#]+)$")
	secondaryGroupLenRegex = regexp.MustCompile(",([0-9#]+),")
	requiredNumRegex       = regexp.MustCompile("([0-9]+)\\.")
	requiredDecimalRegex   = regexp.MustCompile("\\.([0-9]+)")

	enInheritance = map[string]string{
		"en_150": "en_001", "en_AG": "en_001", "en_AI": "en_001", "en_AU": "en_001", "en_BB": "en_001", "en_BE": "en_001", "en_BM": "en_001", "en_BS": "en_001", "en_BW": "en_001", "en_BZ": "en_001", "en_CA": "en_001", "en_CC": "en_001", "en_CK": "en_001", "en_CM": "en_001", "en_CX": "en_001", "en_CY": "en_001", "en_DG": "en_001", "en_DM": "en_001", "en_ER": "en_001", "en_FJ": "en_001", "en_FK": "en_001", "en_FM": "en_001", "en_GB": "en_001", "en_GD": "en_001", "en_GG": "en_001", "en_GH": "en_001", "en_GI": "en_001", "en_GM": "en_001", "en_GY": "en_001", "en_HK": "en_001", "en_IE": "en_001", "en_IL": "en_001", "en_IM": "en_001", "en_IN": "en_001", "en_IO": "en_001", "en_JE": "en_001", "en_JM": "en_001", "en_KE": "en_001", "en_KI": "en_001", "en_KN": "en_001", "en_KY": "en_001", "en_LC": "en_001", "en_LR": "en_001", "en_LS": "en_001", "en_MG": "en_001", "en_MO": "en_001", "en_MS": "en_001", "en_MT": "en_001", "en_MU": "en_001", "en_MW": "en_001", "en_MY": "en_001", "en_NA": "en_001", "en_NF": "en_001", "en_NG": "en_001", "en_NR": "en_001", "en_NU": "en_001", "en_NZ": "en_001", "en_PG": "en_001", "en_PH": "en_001", "en_PK": "en_001", "en_PN": "en_001", "en_PW": "en_001", "en_RW": "en_001", "en_SB": "en_001", "en_SC": "en_001", "en_SD": "en_001", "en_SG": "en_001", "en_SH": "en_001", "en_SL": "en_001", "en_SS": "en_001", "en_SX": "en_001", "en_SZ": "en_001", "en_TC": "en_001", "en_TK": "en_001", "en_TO": "en_001", "en_TT": "en_001", "en_TV": "en_001", "en_TZ": "en_001", "en_UG": "en_001", "en_VC": "en_001", "en_VG": "en_001", "en_VU": "en_001", "en_WS": "en_001", "en_ZA": "en_001", "en_ZM": "en_001", "en_ZW": "en_001", }
	en150Inheritance = map[string]string{"en_AT": "en_150", "en_CH": "en_150", "en_DE": "en_150", "en_DK": "en_150", "en_FI": "en_150", "en_NL": "en_150", "en_SE": "en_150", "en_SI": "en_150"}
	es419Inheritance = map[string]string{
		"es_AR": "es_419", "es_BO": "es_419", "es_BR": "es_419", "es_BZ": "es_419", "es_CL": "es_419", "es_CO": "es_419", "es_CR": "es_419", "es_CU": "es_419", "es_DO": "es_419", "es_EC": "es_419", "es_GT": "es_419", "es_HN": "es_419", "es_MX": "es_419", "es_NI": "es_419", "es_PA": "es_419", "es_PE": "es_419", "es_PR": "es_419", "es_PY": "es_419", "es_SV": "es_419", "es_US": "es_419", "es_UY": "es_419", "es_VE": "es_419",
	}
	rootInheritance = map[string]string{
		"az_Arab": "root", "az_Cyrl": "root", "bm_Nkoo": "root", "bs_Cyrl": "root", "en_Dsrt": "root", "en_Shaw": "root", "ha_Arab": "root", "iu_Latn": "root", "mn_Mong": "root", "ms_Arab": "root", "pa_Arab": "root", "shi_Latn": "root", "sr_Latn": "root", "uz_Arab": "root", "uz_Cyrl": "root", "vai_Latn": "root", "zh_Hant": "root", "yue_Hans": "root",
	}
	ptPtInheritance = map[string]string{
		"pt_AO": "pt_PT", "pt_CH": "pt_PT", "pt_CV": "pt_PT", "pt_GQ": "pt_PT", "pt_GW": "pt_PT", "pt_LU": "pt_PT", "pt_MO": "pt_PT", "pt_MZ": "pt_PT", "pt_ST": "pt_PT", "pt_TL": "pt_PT",
	}
	zhHantHKInheritance = map[string]string{
		"zh_Hant_MO": "zh_Hant_HK",
	}

	inheritMaps = []map[string]string{ enInheritance, en150Inheritance, es419Inheritance, rootInheritance, ptPtInheritance, zhHantHKInheritance}
)

type translator struct {
	Locale     string
	BaseLocale string
	// InheritedLocale string
	Plurals        string
	CardinalFunc   string
	PluralsOrdinal string
	OrdinalFunc    string
	PluralsRange   string
	RangeFunc      string
	Decimal        string
	Group          string
	Minus          string
	Percent        string
	PerMille       string
	TimeSeparator  string
	Infinity       string
	Currencies     string

	// FmtNumber vars
	FmtNumberExists            bool
	FmtNumberGroupLen          int
	FmtNumberSecondaryGroupLen int
	FmtNumberMinDecimalLen     int

	// FmtPercent vars
	FmtPercentExists            bool
	FmtPercentGroupLen          int
	FmtPercentSecondaryGroupLen int
	FmtPercentMinDecimalLen     int
	FmtPercentPrefix            string
	FmtPercentSuffix            string
	FmtPercentInPrefix          bool
	FmtPercentLeft              bool

	// FmtCurrency vars
	FmtCurrencyExists            bool
	FmtCurrencyGroupLen          int
	FmtCurrencySecondaryGroupLen int
	FmtCurrencyMinDecimalLen     int
	FmtCurrencyPrefix            string
	FmtCurrencySuffix            string
	FmtCurrencyInPrefix          bool
	FmtCurrencyLeft              bool
	FmtCurrencyNegativeExists    bool
	FmtCurrencyNegativePrefix    string
	FmtCurrencyNegativeSuffix    string
	FmtCurrencyNegativeInPrefix  bool
	FmtCurrencyNegativeLeft      bool

	// Date & Time
	FmtCalendarExists bool

	FmtMonthsAbbreviated string
	FmtMonthsNarrow      string
	FmtMonthsWide        string

	FmtDaysAbbreviated string
	FmtDaysNarrow      string
	FmtDaysShort       string
	FmtDaysWide        string

	FmtPeriodsAbbreviated string
	FmtPeriodsNarrow      string
	FmtPeriodsShort       string
	FmtPeriodsWide        string

	FmtErasAbbreviated string
	FmtErasNarrow      string
	FmtErasWide        string

	FmtTimezones string

	// calculation only fields below this point...
	DecimalNumberFormat          string
	PercentNumberFormat          string
	CurrencyNumberFormat         string
	NegativeCurrencyNumberFormat string

	// Dates
	FmtDateFull   string
	FmtDateLong   string
	FmtDateMedium string
	FmtDateShort  string

	// Times
	FmtTimeFull   string
	FmtTimeLong   string
	FmtTimeMedium string
	FmtTimeShort  string

	// timezones per locale by type
	timezones map[string]*zoneAbbrev // key = type eg. America_Eastern zone Abbrev will be long form eg. Eastern Standard Time, Pacific Standard Time.....
}

type zoneAbbrev struct {
	standard string
	daylight string
}

var timezones = map[string]*zoneAbbrev{} // key = type eg. America_Eastern zone Abbrev eg. EST & EDT

func main() {

	var err error

	// load template
	tmpl, err = template.New("all").Funcs(tfuncs).ParseGlob("*.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	// load CLDR recourses
	var decoder cldr.Decoder

	cldr, err := decoder.DecodePath("data/core")
	if err != nil {
		panic("failed decode CLDR data; " + err.Error())
	}

	preProcess(cldr)
	postProcess(cldr)

	var currencies string

	for i, curr := range globalCurrencies {

		if i == 0 {
			currencies = curr + " Type = iota\n"
			continue
		}

		currencies += curr + "\n"
	}

	if err = os.MkdirAll(fmt.Sprintf(locDir, "currency"), 0777); err != nil {
		log.Fatal(err)
	}

	filename := fmt.Sprintf(locFilename, "currency", "currency")

	output, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer output.Close()

	if err := tmpl.ExecuteTemplate(output, "currencies", currencies); err != nil {
		log.Fatal(err)
	}

	output.Close()

	// after file written run gofmt on file to ensure best formatting
	cmd := exec.Command("goimports", "-w", filename)
	if err = cmd.Run(); err != nil {
		log.Panic("failed execute \"goimports\" for file ", filename, ": ", err)
	}

	cmd = exec.Command("gofmt", "-s", "-w", filename)
	if err = cmd.Run(); err != nil {
		log.Panic("failed execute \"gofmt\" for file ", filename, ": ", err)
	}

	for _, trans := range translators {
		fmt.Println("Writing Data:", trans.Locale)

		if err = os.MkdirAll(fmt.Sprintf(locDir, trans.Locale), 0777); err != nil {
			log.Fatal(err)
		}

		filename = fmt.Sprintf(locFilename, trans.Locale, trans.Locale)

		output, err := os.Create(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer output.Close()

		if err := tmpl.ExecuteTemplate(output, "translator", trans); err != nil {
			log.Fatal(err)
		}

		output.Close()

		// after file written run gofmt on file to ensure best formatting
		cmd := exec.Command("goimports", "-w", filename)
		if err = cmd.Run(); err != nil {
			log.Panic("failed execute \"goimports\" for file ", filename, ": ", err)
		}

		// this simplifies some syntax that I can;t find an option for in goimports, namely '-s'
		cmd = exec.Command("gofmt", "-s", "-w", filename)
		if err = cmd.Run(); err != nil {
			log.Panic("failed execute \"gofmt\" for file ", filename, ": ", err)
		}

		filename = fmt.Sprintf(locFilename, trans.Locale, trans.Locale+"_test")

		if _, err := os.Stat(filename); err == nil {
			fmt.Println("*************** test file exists, skipping:", filename)
			continue
		}

		output, err = os.Create(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer output.Close()

		if err := tmpl.ExecuteTemplate(output, "tests", trans); err != nil {
			log.Fatal(err)
		}

		output.Close()

		// after file written run gofmt on file to ensure best formatting
		cmd = exec.Command("goimports", "-w", filename)
		if err = cmd.Run(); err != nil {
			log.Panic("failed execute \"goimports\" for file ", filename, ": ", err)
		}

		// this simplifies some syntax that I can;t find an option for in goimports, namely '-s'
		cmd = exec.Command("gofmt", "-s", "-w", filename)
		if err = cmd.Run(); err != nil {
			log.Panic("failed execute \"gofmt\" for file ", filename, ": ", err)
		}
	}
}

func applyOverrides(trans *translator) {

	if trans.BaseLocale == "ru" {
		trans.PercentNumberFormat = "#,##0%"
	}
}

func postProcess(cldr *cldr.CLDR) {

	for _, v := range timezones {

		// no DST
		if len(v.daylight) == 0 {
			v.daylight = v.standard
		}
	}

	var inherited *translator
	var base *translator
	var inheritedFound, baseFound bool

	for _, trans := range translators {

		fmt.Println("Post Processing:", trans.Locale)

		// cardinal plural rules
		trans.CardinalFunc, trans.Plurals = parseCardinalPluralRuleFunc(cldr, trans.Locale, trans.BaseLocale)

		//ordinal plural rules
		trans.OrdinalFunc, trans.PluralsOrdinal = parseOrdinalPluralRuleFunc(cldr, trans.BaseLocale)

		// range plural rules
		trans.RangeFunc, trans.PluralsRange = parseRangePluralRuleFunc(cldr, trans.BaseLocale)

		// ignore base locales
		if trans.BaseLocale == trans.Locale {
			inheritedFound = false
			baseFound = false
		} else {

			inheritedFound = false

			for _, inheritMap := range(inheritMaps) {
				if inherit, found := inheritMap[trans.Locale]; found {
					inherited, inheritedFound = translators[inherit]
					break;
				}
			}

			base, baseFound = baseTranslators[trans.BaseLocale]
		}

		// Numbers

		if len(trans.Decimal) == 0 {

			if inheritedFound {
				trans.Decimal = inherited.Decimal
			}

			if len(trans.Decimal) == 0 && baseFound {
				trans.Decimal = base.Decimal
			}

			if len(trans.Decimal) == 0 {
				trans.Decimal = ""
			}
		}

		if len(trans.Group) == 0 {

			if inheritedFound {
				trans.Group = inherited.Group
			}

			if len(trans.Group) == 0 && baseFound {
				trans.Group = base.Group
			}

			if len(trans.Group) == 0 {
				trans.Group = ""
			}
		}

		if len(trans.Minus) == 0 {

			if inheritedFound {
				trans.Minus = inherited.Minus
			}

			if len(trans.Minus) == 0 && baseFound {
				trans.Minus = base.Minus
			}

			if len(trans.Minus) == 0 {
				trans.Minus = ""
			}
		}

		if len(trans.Percent) == 0 {

			if inheritedFound {
				trans.Percent = inherited.Percent
			}

			if len(trans.Percent) == 0 && baseFound {
				trans.Percent = base.Percent
			}

			if len(trans.Percent) == 0 {
				trans.Percent = ""
			}
		}

		if len(trans.PerMille) == 0 {

			if inheritedFound {
				trans.PerMille = inherited.PerMille
			}

			if len(trans.PerMille) == 0 && baseFound {
				trans.PerMille = base.PerMille
			}

			if len(trans.PerMille) == 0 {
				trans.PerMille = ""
			}
		}

		if len(trans.TimeSeparator) == 0 && inheritedFound {
			trans.TimeSeparator = inherited.TimeSeparator
		}

		if len(trans.TimeSeparator) == 0 && baseFound {
			trans.TimeSeparator = base.TimeSeparator
		}

		if len(trans.Infinity) == 0 && inheritedFound {
			trans.Infinity = inherited.Infinity
		}

		if len(trans.Infinity) == 0 && baseFound {
			trans.Infinity = base.Infinity
		}

		// Currency

		// number values

		if len(trans.DecimalNumberFormat) == 0 && inheritedFound {
			trans.DecimalNumberFormat = inherited.DecimalNumberFormat
		}

		if len(trans.DecimalNumberFormat) == 0 && baseFound {
			trans.DecimalNumberFormat = base.DecimalNumberFormat
		}

		if len(trans.PercentNumberFormat) == 0 && inheritedFound {
			trans.PercentNumberFormat = inherited.PercentNumberFormat
		}

		if len(trans.PercentNumberFormat) == 0 && baseFound {
			trans.PercentNumberFormat = base.PercentNumberFormat
		}

		if len(trans.CurrencyNumberFormat) == 0 && inheritedFound {
			trans.CurrencyNumberFormat = inherited.CurrencyNumberFormat
		}

		if len(trans.CurrencyNumberFormat) == 0 && baseFound {
			trans.CurrencyNumberFormat = base.CurrencyNumberFormat
		}

		if len(trans.NegativeCurrencyNumberFormat) == 0 && inheritedFound {
			trans.NegativeCurrencyNumberFormat = inherited.NegativeCurrencyNumberFormat
		}

		if len(trans.NegativeCurrencyNumberFormat) == 0 && baseFound {
			trans.NegativeCurrencyNumberFormat = base.NegativeCurrencyNumberFormat
		}

		// date values

		if len(trans.FmtDateFull) == 0 && inheritedFound {
			trans.FmtDateFull = inherited.FmtDateFull
		}

		if len(trans.FmtDateFull) == 0 && baseFound {
			trans.FmtDateFull = base.FmtDateFull
		}

		if len(trans.FmtDateLong) == 0 && inheritedFound {
			trans.FmtDateLong = inherited.FmtDateLong
		}

		if len(trans.FmtDateLong) == 0 && baseFound {
			trans.FmtDateLong = base.FmtDateLong
		}

		if len(trans.FmtDateMedium) == 0 && inheritedFound {
			trans.FmtDateMedium = inherited.FmtDateMedium
		}

		if len(trans.FmtDateMedium) == 0 && baseFound {
			trans.FmtDateMedium = base.FmtDateMedium
		}

		if len(trans.FmtDateShort) == 0 && inheritedFound {
			trans.FmtDateShort = inherited.FmtDateShort
		}

		if len(trans.FmtDateShort) == 0 && baseFound {
			trans.FmtDateShort = base.FmtDateShort
		}

		// time values

		if len(trans.FmtTimeFull) == 0 && inheritedFound {
			trans.FmtTimeFull = inherited.FmtTimeFull
		}

		if len(trans.FmtTimeFull) == 0 && baseFound {
			trans.FmtTimeFull = base.FmtTimeFull
		}

		if len(trans.FmtTimeLong) == 0 && inheritedFound {
			trans.FmtTimeLong = inherited.FmtTimeLong
		}

		if len(trans.FmtTimeLong) == 0 && baseFound {
			trans.FmtTimeLong = base.FmtTimeLong
		}

		if len(trans.FmtTimeMedium) == 0 && inheritedFound {
			trans.FmtTimeMedium = inherited.FmtTimeMedium
		}

		if len(trans.FmtTimeMedium) == 0 && baseFound {
			trans.FmtTimeMedium = base.FmtTimeMedium
		}

		if len(trans.FmtTimeShort) == 0 && inheritedFound {
			trans.FmtTimeShort = inherited.FmtTimeShort
		}

		if len(trans.FmtTimeShort) == 0 && baseFound {
			trans.FmtTimeShort = base.FmtTimeShort
		}

		// month values

		if len(trans.FmtMonthsAbbreviated) == 0 && inheritedFound {
			trans.FmtMonthsAbbreviated = inherited.FmtMonthsAbbreviated
		}

		if len(trans.FmtMonthsAbbreviated) == 0 && baseFound {
			trans.FmtMonthsAbbreviated = base.FmtMonthsAbbreviated
		}

		if len(trans.FmtMonthsNarrow) == 0 && inheritedFound {
			trans.FmtMonthsNarrow = inherited.FmtMonthsNarrow
		}

		if len(trans.FmtMonthsNarrow) == 0 && baseFound {
			trans.FmtMonthsNarrow = base.FmtMonthsNarrow
		}

		if len(trans.FmtMonthsWide) == 0 && inheritedFound {
			trans.FmtMonthsWide = inherited.FmtMonthsWide
		}

		if len(trans.FmtMonthsWide) == 0 && baseFound {
			trans.FmtMonthsWide = base.FmtMonthsWide
		}

		// day values

		if len(trans.FmtDaysAbbreviated) == 0 && inheritedFound {
			trans.FmtDaysAbbreviated = inherited.FmtDaysAbbreviated
		}

		if len(trans.FmtDaysAbbreviated) == 0 && baseFound {
			trans.FmtDaysAbbreviated = base.FmtDaysAbbreviated
		}

		if len(trans.FmtDaysNarrow) == 0 && inheritedFound {
			trans.FmtDaysNarrow = inherited.FmtDaysNarrow
		}

		if len(trans.FmtDaysNarrow) == 0 && baseFound {
			trans.FmtDaysNarrow = base.FmtDaysNarrow
		}

		if len(trans.FmtDaysShort) == 0 && inheritedFound {
			trans.FmtDaysShort = inherited.FmtDaysShort
		}

		if len(trans.FmtDaysShort) == 0 && baseFound {
			trans.FmtDaysShort = base.FmtDaysShort
		}

		if len(trans.FmtDaysWide) == 0 && inheritedFound {
			trans.FmtDaysWide = inherited.FmtDaysWide
		}

		if len(trans.FmtDaysWide) == 0 && baseFound {
			trans.FmtDaysWide = base.FmtDaysWide
		}

		// period values

		if len(trans.FmtPeriodsAbbreviated) == 0 && inheritedFound {
			trans.FmtPeriodsAbbreviated = inherited.FmtPeriodsAbbreviated
		}

		if len(trans.FmtPeriodsAbbreviated) == 0 && baseFound {
			trans.FmtPeriodsAbbreviated = base.FmtPeriodsAbbreviated
		}

		if len(trans.FmtPeriodsNarrow) == 0 && inheritedFound {
			trans.FmtPeriodsNarrow = inherited.FmtPeriodsNarrow
		}

		if len(trans.FmtPeriodsNarrow) == 0 && baseFound {
			trans.FmtPeriodsNarrow = base.FmtPeriodsNarrow
		}

		if len(trans.FmtPeriodsShort) == 0 && inheritedFound {
			trans.FmtPeriodsShort = inherited.FmtPeriodsShort
		}

		if len(trans.FmtPeriodsShort) == 0 && baseFound {
			trans.FmtPeriodsShort = base.FmtPeriodsShort
		}

		if len(trans.FmtPeriodsWide) == 0 && inheritedFound {
			trans.FmtPeriodsWide = inherited.FmtPeriodsWide
		}

		if len(trans.FmtPeriodsWide) == 0 && baseFound {
			trans.FmtPeriodsWide = base.FmtPeriodsWide
		}

		// era values

		if len(trans.FmtErasAbbreviated) == 0 && inheritedFound {
			trans.FmtErasAbbreviated = inherited.FmtErasAbbreviated
		}

		if len(trans.FmtErasAbbreviated) == 0 && baseFound {
			trans.FmtErasAbbreviated = base.FmtErasAbbreviated
		}

		if len(trans.FmtErasNarrow) == 0 && inheritedFound {
			trans.FmtErasNarrow = inherited.FmtErasNarrow
		}

		if len(trans.FmtErasNarrow) == 0 && baseFound {
			trans.FmtErasNarrow = base.FmtErasNarrow
		}

		if len(trans.FmtErasWide) == 0 && inheritedFound {
			trans.FmtErasWide = inherited.FmtErasWide
		}

		if len(trans.FmtErasWide) == 0 && baseFound {
			trans.FmtErasWide = base.FmtErasWide
		}

		ldml := cldr.RawLDML(trans.Locale)

		currencies := make([]string, len(globalCurrencies), len(globalCurrencies))

		var kval string

		for k, v := range globCurrencyIdxMap {

			kval = k
			// if kval[:len(kval)-1] != " " {
			// 	kval += " "
			// }

			currencies[v] = kval
		}

		// some just have no data...
		if ldml.Numbers != nil {

			if ldml.Numbers.Currencies != nil {
				for _, currency := range ldml.Numbers.Currencies.Currency {

					if len(currency.Symbol) == 0 {
						continue
					}

					if len(currency.Symbol[0].Data()) == 0 {
						continue
					}

					if len(currency.Type) == 0 {
						continue
					}

					currencies[globCurrencyIdxMap[currency.Type]] = currency.Symbol[0].Data()
				}
			}
		}

		trans.Currencies = fmt.Sprintf("%#v", currencies)

		// timezones

		if (trans.timezones == nil || len(trans.timezones) == 0) && inheritedFound {
			trans.timezones = inherited.timezones
		}

		if (trans.timezones == nil || len(trans.timezones) == 0) && baseFound {
			trans.timezones = base.timezones
		}

		// make sure all inherited timezones are part of sub locale timezones
		if inheritedFound {

			var ok bool

			for k, v := range inherited.timezones {

				if _, ok = trans.timezones[k]; ok {
					continue
				}

				trans.timezones[k] = v
			}
		}

		// make sure all base timezones are part of sub locale timezones
		if baseFound {

			var ok bool

			for k, v := range base.timezones {

				if _, ok = trans.timezones[k]; ok {
					continue
				}

				trans.timezones[k] = v
			}
		}

		applyOverrides(trans)

		parseDecimalNumberFormat(trans)
		parsePercentNumberFormat(trans)
		parseCurrencyNumberFormat(trans)
	}

	for _, trans := range translators {

		fmt.Println("Final Processing:", trans.Locale)

		// if it's still nill.....
		if trans.timezones == nil {
			trans.timezones = make(map[string]*zoneAbbrev)
		}

		tz := make(map[string]string) // key = abbrev locale eg. EST, EDT, MST, PST... value = long locale eg. Eastern Standard Time, Pacific Time.....

		for k, v := range timezones {

			ttz, ok := trans.timezones[k]
			if !ok {
				ttz = v
				trans.timezones[k] = v
			}

			tz[v.standard] = ttz.standard
			tz[v.daylight] = ttz.daylight
		}

		trans.FmtTimezones = fmt.Sprintf("%#v", tz)

		if len(trans.TimeSeparator) == 0 {
			trans.TimeSeparator = ":"
		}

		trans.FmtDateShort, trans.FmtDateMedium, trans.FmtDateLong, trans.FmtDateFull = parseDateFormats(trans, trans.FmtDateShort, trans.FmtDateMedium, trans.FmtDateLong, trans.FmtDateFull)
		trans.FmtTimeShort, trans.FmtTimeMedium, trans.FmtTimeLong, trans.FmtTimeFull = parseDateFormats(trans, trans.FmtTimeShort, trans.FmtTimeMedium, trans.FmtTimeLong, trans.FmtTimeFull)
	}
}

// preprocesses maps, array etc... just requires multiple passes no choice....
func preProcess(cldrVar *cldr.CLDR) {

	for _, l := range cldrVar.Locales() {

		fmt.Println("Pre Processing:", l)

		split := strings.SplitN(l, "_", 2)
		baseLocale := split[0]
		// inheritedLocale := baseLocale

		// // one of the inherited english locales
		// // http://cldr.unicode.org/development/development-process/design-proposals/english-inheritance
		// if l == "en_001" || l == "en_GB" {
		// 	inheritedLocale = l
		// }

		trans := &translator{
			Locale:     l,
			BaseLocale: baseLocale,
			// InheritedLocale: inheritedLocale,
		}

		// if is a base locale
		if len(split) == 1 {
			baseTranslators[baseLocale] = trans
		}

		// baseTranslators[l] = trans
		// baseTranslators[baseLocale] = trans // allowing for unofficial fallback if none exists
		translators[l] = trans

		// get number, currency and datetime symbols

		// number values
		ldml := cldrVar.RawLDML(l)

		// some just have no data...
		if ldml.Numbers != nil {

			if len(ldml.Numbers.Symbols) > 0 {

				symbol := ldml.Numbers.Symbols[0]

				// Try to get the default numbering system instead of the first one
				systems := ldml.Numbers.DefaultNumberingSystem
				// There shouldn't really be more than one DefaultNumberingSystem
				if len(systems) > 0 {
					if dns := systems[0].Data(); dns != "" {
						for k := range ldml.Numbers.Symbols {
							if ldml.Numbers.Symbols[k].NumberSystem == dns {
								symbol = ldml.Numbers.Symbols[k]
								break
							}
						}
					}
				}

				if len(symbol.Decimal) > 0 {
					trans.Decimal = symbol.Decimal[0].Data()
				}
				if len(symbol.Group) > 0 {
					trans.Group = symbol.Group[0].Data()
				}
				if len(symbol.MinusSign) > 0 {
					trans.Minus = symbol.MinusSign[0].Data()
				}
				if len(symbol.PercentSign) > 0 {
					trans.Percent = symbol.PercentSign[0].Data()
				}
				if len(symbol.PerMille) > 0 {
					trans.PerMille = symbol.PerMille[0].Data()
				}

				if len(symbol.TimeSeparator) > 0 {
					trans.TimeSeparator = symbol.TimeSeparator[0].Data()
				}

				if len(symbol.Infinity) > 0 {
					trans.Infinity = symbol.Infinity[0].Data()
				}
			}

			if ldml.Numbers.Currencies != nil {

				for _, currency := range ldml.Numbers.Currencies.Currency {

					if len(strings.TrimSpace(currency.Type)) == 0 {
						continue
					}

					globalCurrenciesMap[currency.Type] = struct{}{}
				}
			}

			if len(ldml.Numbers.DecimalFormats) > 0 && len(ldml.Numbers.DecimalFormats[0].DecimalFormatLength) > 0 {

				for _, dfl := range ldml.Numbers.DecimalFormats[0].DecimalFormatLength {
					if len(dfl.Type) == 0 {
						trans.DecimalNumberFormat = dfl.DecimalFormat[0].Pattern[0].Data()
						break
					}
				}
			}

			if len(ldml.Numbers.PercentFormats) > 0 && len(ldml.Numbers.PercentFormats[0].PercentFormatLength) > 0 {

				for _, dfl := range ldml.Numbers.PercentFormats[0].PercentFormatLength {
					if len(dfl.Type) == 0 {
						trans.PercentNumberFormat = dfl.PercentFormat[0].Pattern[0].Data()
						break
					}
				}
			}

			if len(ldml.Numbers.CurrencyFormats) > 0 && len(ldml.Numbers.CurrencyFormats[0].CurrencyFormatLength) > 0 {

				if len(ldml.Numbers.CurrencyFormats[0].CurrencyFormatLength[0].CurrencyFormat) > 1 {

					split := strings.SplitN(ldml.Numbers.CurrencyFormats[0].CurrencyFormatLength[0].CurrencyFormat[1].Pattern[0].Data(), ";", 2)

					trans.CurrencyNumberFormat = split[0]

					if len(split) > 1 && len(split[1]) > 0 {
						trans.NegativeCurrencyNumberFormat = split[1]
					} else {
						trans.NegativeCurrencyNumberFormat = trans.CurrencyNumberFormat
					}
				} else {
					trans.CurrencyNumberFormat = ldml.Numbers.CurrencyFormats[0].CurrencyFormatLength[0].CurrencyFormat[0].Pattern[0].Data()
					trans.NegativeCurrencyNumberFormat = trans.CurrencyNumberFormat
				}
			}
		}

		if ldml.Dates != nil {

			if ldml.Dates.TimeZoneNames != nil {

				for _, zone := range ldml.Dates.TimeZoneNames.Metazone {

					for _, short := range zone.Short {

						if len(short.Standard) > 0 {
							za, ok := timezones[zone.Type]
							if !ok {
								za = new(zoneAbbrev)
								timezones[zone.Type] = za
							}
							za.standard = short.Standard[0].Data()
						}

						if len(short.Daylight) > 0 {
							za, ok := timezones[zone.Type]
							if !ok {
								za = new(zoneAbbrev)
								timezones[zone.Type] = za
							}
							za.daylight = short.Daylight[0].Data()
						}
					}

					for _, long := range zone.Long {

						if trans.timezones == nil {
							trans.timezones = make(map[string]*zoneAbbrev)
						}

						if len(long.Standard) > 0 {
							za, ok := trans.timezones[zone.Type]
							if !ok {
								za = new(zoneAbbrev)
								trans.timezones[zone.Type] = za
							}
							za.standard = long.Standard[0].Data()
						}

						za, ok := trans.timezones[zone.Type]
						if !ok {
							za = new(zoneAbbrev)
							trans.timezones[zone.Type] = za
						}

						if len(long.Daylight) > 0 {
							za.daylight = long.Daylight[0].Data()
						} else {
							za.daylight = za.standard
						}
					}
				}
			}

			if ldml.Dates.Calendars != nil {

				var calendar *cldr.Calendar

				for _, cal := range ldml.Dates.Calendars.Calendar {
					if cal.Type == "gregorian" {
						calendar = cal
					}
				}

				if calendar != nil {

					if calendar.DateFormats != nil {

						for _, datefmt := range calendar.DateFormats.DateFormatLength {

							switch datefmt.Type {
							case "full":
								trans.FmtDateFull = datefmt.DateFormat[0].Pattern[0].Data()

							case "long":
								trans.FmtDateLong = datefmt.DateFormat[0].Pattern[0].Data()

							case "medium":
								trans.FmtDateMedium = datefmt.DateFormat[0].Pattern[0].Data()

							case "short":
								trans.FmtDateShort = datefmt.DateFormat[0].Pattern[0].Data()
							}
						}
					}

					if calendar.TimeFormats != nil {

						for _, datefmt := range calendar.TimeFormats.TimeFormatLength {

							switch datefmt.Type {
							case "full":
								trans.FmtTimeFull = datefmt.TimeFormat[0].Pattern[0].Data()
							case "long":
								trans.FmtTimeLong = datefmt.TimeFormat[0].Pattern[0].Data()
							case "medium":
								trans.FmtTimeMedium = datefmt.TimeFormat[0].Pattern[0].Data()
							case "short":
								trans.FmtTimeShort = datefmt.TimeFormat[0].Pattern[0].Data()
							}
						}
					}

					if calendar.Months != nil {

						// month context starts at 'format', but there is also has 'stand-alone'
						// I'm making the decision to use the 'stand-alone' if, and only if,
						// the value does not exist in the 'format' month context
						var abbrSet, narrSet, wideSet bool

						for _, monthctx := range calendar.Months.MonthContext {

							for _, months := range monthctx.MonthWidth {

								var monthData []string

								for _, m := range months.Month {

									if len(m.Data()) == 0 {
										continue
									}

									switch m.Type {
									case "1":
										monthData = append(monthData, m.Data())
									case "2":
										monthData = append(monthData, m.Data())
									case "3":
										monthData = append(monthData, m.Data())
									case "4":
										monthData = append(monthData, m.Data())
									case "5":
										monthData = append(monthData, m.Data())
									case "6":
										monthData = append(monthData, m.Data())
									case "7":
										monthData = append(monthData, m.Data())
									case "8":
										monthData = append(monthData, m.Data())
									case "9":
										monthData = append(monthData, m.Data())
									case "10":
										monthData = append(monthData, m.Data())
									case "11":
										monthData = append(monthData, m.Data())
									case "12":
										monthData = append(monthData, m.Data())
									}
								}

								if len(monthData) > 0 {

									// making array indexes line up with month values
									// so I'll have an extra empty value, it's way faster
									// than a switch over all type values...
									monthData = append(make([]string, 1, len(monthData)+1), monthData...)

									switch months.Type {
									case "abbreviated":
										if !abbrSet {
											abbrSet = true
											trans.FmtMonthsAbbreviated = fmt.Sprintf("%#v", monthData)
										}
									case "narrow":
										if !narrSet {
											narrSet = true
											trans.FmtMonthsNarrow = fmt.Sprintf("%#v", monthData)
										}
									case "wide":
										if !wideSet {
											wideSet = true
											trans.FmtMonthsWide = fmt.Sprintf("%#v", monthData)
										}
									}
								}
							}
						}
					}

					if calendar.Days != nil {

						// day context starts at 'format', but there is also has 'stand-alone'
						// I'm making the decision to use the 'stand-alone' if, and only if,
						// the value does not exist in the 'format' day context
						var abbrSet, narrSet, shortSet, wideSet bool

						for _, dayctx := range calendar.Days.DayContext {

							for _, days := range dayctx.DayWidth {

								var dayData []string

								for _, d := range days.Day {

									switch d.Type {
									case "sun":
										dayData = append(dayData, d.Data())
									case "mon":
										dayData = append(dayData, d.Data())
									case "tue":
										dayData = append(dayData, d.Data())
									case "wed":
										dayData = append(dayData, d.Data())
									case "thu":
										dayData = append(dayData, d.Data())
									case "fri":
										dayData = append(dayData, d.Data())
									case "sat":
										dayData = append(dayData, d.Data())
									}
								}

								if len(dayData) > 0 {
									switch days.Type {
									case "abbreviated":
										if !abbrSet {
											abbrSet = true
											trans.FmtDaysAbbreviated = fmt.Sprintf("%#v", dayData)
										}
									case "narrow":
										if !narrSet {
											narrSet = true
											trans.FmtDaysNarrow = fmt.Sprintf("%#v", dayData)
										}
									case "short":
										if !shortSet {
											shortSet = true
											trans.FmtDaysShort = fmt.Sprintf("%#v", dayData)
										}
									case "wide":
										if !wideSet {
											wideSet = true
											trans.FmtDaysWide = fmt.Sprintf("%#v", dayData)
										}
									}
								}
							}
						}
					}

					if calendar.DayPeriods != nil {

						// day periods context starts at 'format', but there is also has 'stand-alone'
						// I'm making the decision to use the 'stand-alone' if, and only if,
						// the value does not exist in the 'format' day period context
						var abbrSet, narrSet, shortSet, wideSet bool

						for _, ctx := range calendar.DayPeriods.DayPeriodContext {

							for _, width := range ctx.DayPeriodWidth {

								// [0] = AM
								// [0] = PM
								ampm := make([]string, 2, 2)

								for _, d := range width.DayPeriod {

									if d.Type == "am" {
										ampm[0] = d.Data()
										continue
									}

									if d.Type == "pm" {
										ampm[1] = d.Data()
									}
								}

								switch width.Type {
								case "abbreviated":
									if !abbrSet {
										abbrSet = true
										trans.FmtPeriodsAbbreviated = fmt.Sprintf("%#v", ampm)
									}
								case "narrow":
									if !narrSet {
										narrSet = true
										trans.FmtPeriodsNarrow = fmt.Sprintf("%#v", ampm)
									}
								case "short":
									if !shortSet {
										shortSet = true
										trans.FmtPeriodsShort = fmt.Sprintf("%#v", ampm)
									}
								case "wide":
									if !wideSet {
										wideSet = true
										trans.FmtPeriodsWide = fmt.Sprintf("%#v", ampm)
									}
								}
							}
						}
					}

					if calendar.Eras != nil {

						// [0] = BC
						// [0] = AD
						abbrev := make([]string, 2, 2)
						narr := make([]string, 2, 2)
						wide := make([]string, 2, 2)

						if calendar.Eras.EraAbbr != nil {

							if len(calendar.Eras.EraAbbr.Era) == 4 {
								abbrev[0] = calendar.Eras.EraAbbr.Era[0].Data()
								abbrev[1] = calendar.Eras.EraAbbr.Era[2].Data()
							} else if len(calendar.Eras.EraAbbr.Era) == 2 {
								abbrev[0] = calendar.Eras.EraAbbr.Era[0].Data()
								abbrev[1] = calendar.Eras.EraAbbr.Era[1].Data()
							}
						}

						if calendar.Eras.EraNarrow != nil {

							if len(calendar.Eras.EraNarrow.Era) == 4 {
								narr[0] = calendar.Eras.EraNarrow.Era[0].Data()
								narr[1] = calendar.Eras.EraNarrow.Era[2].Data()
							} else if len(calendar.Eras.EraNarrow.Era) == 2 {
								narr[0] = calendar.Eras.EraNarrow.Era[0].Data()
								narr[1] = calendar.Eras.EraNarrow.Era[1].Data()
							}
						}

						if calendar.Eras.EraNames != nil {

							if len(calendar.Eras.EraNames.Era) == 4 {
								wide[0] = calendar.Eras.EraNames.Era[0].Data()
								wide[1] = calendar.Eras.EraNames.Era[2].Data()
							} else if len(calendar.Eras.EraNames.Era) == 2 {
								wide[0] = calendar.Eras.EraNames.Era[0].Data()
								wide[1] = calendar.Eras.EraNames.Era[1].Data()
							}
						}

						trans.FmtErasAbbreviated = fmt.Sprintf("%#v", abbrev)
						trans.FmtErasNarrow = fmt.Sprintf("%#v", narr)
						trans.FmtErasWide = fmt.Sprintf("%#v", wide)
					}
				}
			}
		}
	}

	for k := range globalCurrenciesMap {
		globalCurrencies = append(globalCurrencies, k)
	}

	sort.Strings(globalCurrencies)

	for i, loc := range globalCurrencies {
		globCurrencyIdxMap[loc] = i
	}
}

func parseDateFormats(trans *translator, shortFormat, mediumFormat, longFormat, fullFormat string) (short, medium, long, full string) {

	// Short Date Parsing

	short = parseDateTimeFormat(trans.BaseLocale, shortFormat, 2)
	medium = parseDateTimeFormat(trans.BaseLocale, mediumFormat, 2)
	long = parseDateTimeFormat(trans.BaseLocale, longFormat, 1)
	full = parseDateTimeFormat(trans.BaseLocale, fullFormat, 0)

	// End Short Data Parsing

	return
}

func parseDateTimeFormat(baseLocale, format string, eraScore uint8) (results string) {

	// rules:
	// y = four digit year
	// yy = two digit year

	// var b []byte

	var inConstantText bool
	var start int

	for i := 0; i < len(format); i++ {

		switch format[i] {

		// time separator
		case ':':

			if inConstantText {
				inConstantText = false
				results += "b = append(b, " + fmt.Sprintf("%#v", []byte(format[start:i])) + "...)\n"
			}

			results += "b = append(b, " + baseLocale + ".timeSeparator...)"
		case '\'':

			i++
			startI := i

			// peek to see if ''
			if len(format) != i && format[i] == '\'' {

				if inConstantText {
					inConstantText = false
					results += "b = append(b, " + fmt.Sprintf("%#v", []byte(format[start:i-1])) + "...)\n"
				} else {
					inConstantText = true
					start = i
				}

				continue
			}

			// not '' so whatever comes between '' is constant

			if len(format) != i {

				// advance i to the next single quote + 1
				for ; i < len(format); i++ {
					if format[i] == '\'' {

						if inConstantText {
							inConstantText = false
							b := []byte(format[start : startI-1])
							b = append(b, []byte(format[startI:i])...)

							results += "b = append(b, " + fmt.Sprintf("%#v", b) + "...)\n"

						} else {
							results += "b = append(b, " + fmt.Sprintf("%#v", []byte(format[startI:i])) + "...)\n"
						}

						break
					}
				}
			}

		// 24 hour
		case 'H':

			if inConstantText {
				inConstantText = false
				results += "b = append(b, " + fmt.Sprintf("%#v", []byte(format[start:i])) + "...)\n"
			}

			// peek
			// two digit hour required?
			if len(format) != i+1 && format[i+1] == 'H' {
				i++
				results += `
					if t.Hour() < 10 {
						b = append(b, '0')
					}

				`
			}

			results += "b = strconv.AppendInt(b, int64(t.Hour()), 10)\n"

		// hour
		case 'h':

			if inConstantText {
				inConstantText = false
				results += "b = append(b, " + fmt.Sprintf("%#v", []byte(format[start:i])) + "...)\n"
			}

			results += `
				h := t.Hour()

				if h > 12 {
					h -= 12
				}

			`

			// peek
			// two digit hour required?
			if len(format) != i+1 && format[i+1] == 'h' {
				i++
				results += `
					if h < 10 {
						b = append(b, '0')
					}

				`
			}

			results += "b = strconv.AppendInt(b, int64(h), 10)\n"

		// minute
		case 'm':

			if inConstantText {
				inConstantText = false
				results += "b = append(b, " + fmt.Sprintf("%#v", []byte(format[start:i])) + "...)\n"
			}

			// peek
			// two digit minute required?
			if len(format) != i+1 && format[i+1] == 'm' {
				i++
				results += `

					if t.Minute() < 10 {
						b = append(b, '0')
					}

				`
			}

			results += "b = strconv.AppendInt(b, int64(t.Minute()), 10)\n"

		// second
		case 's':

			if inConstantText {
				inConstantText = false
				results += "b = append(b, " + fmt.Sprintf("%#v", []byte(format[start:i])) + "...)\n"
			}

			// peek
			// two digit minute required?
			if len(format) != i+1 && format[i+1] == 's' {
				i++
				results += `

					if t.Second() < 10 {
						b = append(b, '0')
					}

				`
			}

			results += "b = strconv.AppendInt(b, int64(t.Second()), 10)\n"

		// day period
		case 'a':

			if inConstantText {
				inConstantText = false
				results += "b = append(b, " + fmt.Sprintf("%#v", []byte(format[start:i])) + "...)\n"
			}

			// only used with 'h', patterns should not contains 'a' without 'h' so not checking

			// choosing to use abbreviated, didn't see any rules about which should be used with which
			// date format....

			results += `

				if t.Hour() < 12 {
					b = append(b, ` + baseLocale + `.periodsAbbreviated[0]...)
				} else {
					b = append(b, ` + baseLocale + `.periodsAbbreviated[1]...)
				}

			`

		// timezone
		case 'z', 'v':

			if inConstantText {
				inConstantText = false
				results += "b = append(b, " + fmt.Sprintf("%#v", []byte(format[start:i])) + "...)\n"
			}

			// consume multiple, only handling Abbrev tz from time.Time for the moment...

			var count int

			if format[i] == 'z' {
				for j := i; j < len(format); j++ {
					if format[j] == 'z' {
						count++
					} else {
						break
					}
				}
			}

			if format[i] == 'v' {
				for j := i; j < len(format); j++ {
					if format[j] == 'v' {
						count++
					} else {
						break
					}
				}
			}

			i += count - 1

			// using the timezone on the Go time object, eg. EST, EDT, MST.....

			if count < 4 {

				results += `

					tz, _ := t.Zone()
					b = append(b, tz...)

				`
			} else {

				results += `
					tz, _ := t.Zone()

					if btz, ok := ` + baseLocale + `.timezones[tz]; ok {
						b = append(b, btz...)
					} else {
						b = append(b, tz...)
					}

				`
			}

		// day
		case 'd':

			if inConstantText {
				inConstantText = false
				results += "b = append(b, " + fmt.Sprintf("%#v", []byte(format[start:i])) + "...)\n"
			}

			// peek
			// two digit day required?
			if len(format) != i+1 && format[i+1] == 'd' {
				i++
				results += `

					if t.Day() < 10 {
						b = append(b, '0')
					}

				`
			}

			results += "b = strconv.AppendInt(b, int64(t.Day()), 10)\n"

		// month
		case 'M':

			if inConstantText {
				inConstantText = false
				results += "b = append(b, " + fmt.Sprintf("%#v", []byte(format[start:i])) + "...)\n"
			}

			var count int

			// count # of M's
			for j := i; j < len(format); j++ {
				if format[j] == 'M' {
					count++
				} else {
					break
				}
			}

			switch count {

			// Numeric form, at least 1 digit
			case 1:

				results += "b = strconv.AppendInt(b, int64(t.Month()), 10)\n"

			// Number form, at least 2 digits (padding with 0)
			case 2:

				results += `

				if t.Month() < 10 {
					b = append(b, '0')
				}
					
				b = strconv.AppendInt(b, int64(t.Month()), 10)

				`

			// Abbreviated form
			case 3:

				results += "b = append(b, " + baseLocale + ".monthsAbbreviated[t.Month()]...)\n"

			// Full/Wide form
			case 4:

				results += "b = append(b, " + baseLocale + ".monthsWide[t.Month()]...)\n"

			// Narrow form - only used in where context makes it clear, such as headers in a calendar.
			// Should be one character wherever possible.
			case 5:

				results += "b = append(b, " + baseLocale + ".monthsNarrow[t.Month()]...)\n"
			}

			// skip over M's
			i += count - 1

		// year
		case 'y':

			if inConstantText {
				inConstantText = false
				results += "b = append(b, " + fmt.Sprintf("%#v", []byte(format[start:i])) + "...)\n"
			}

			// peek
			// two digit year
			if len(format) != i+1 && format[i+1] == 'y' {
				i++
				results += `

					if t.Year() > 9 {
						b = append(b, strconv.Itoa(t.Year())[2:]...)
					} else {
						b = append(b, strconv.Itoa(t.Year())[1:]...)
					}

				`
			} else {
				// four digit year
				results += `

					if t.Year() > 0 {
						b = strconv.AppendInt(b, int64(t.Year()), 10)
					} else {
						b = strconv.AppendInt(b, int64(-t.Year()), 10)
					}

				`
			}

		// weekday
		// I know I only see 'EEEE' in the xml, but just in case handled all posibilities
		case 'E':

			if inConstantText {
				inConstantText = false
				results += "b = append(b, " + fmt.Sprintf("%#v", []byte(format[start:i])) + "...)\n"
			}

			var count int

			// count # of E's
			for j := i; j < len(format); j++ {
				if format[j] == 'E' {
					count++
				} else {
					break
				}
			}

			switch count {

			// Narrow
			case 1:

				results += "b = append(b, " + baseLocale + ".daysNarrow[t.Weekday()]...)\n"

			// Short
			case 2:

				results += "b = append(b, " + baseLocale + ".daysShort[t.Weekday()]...)\n"

			// Abbreviated
			case 3:

				results += "b = append(b, " + baseLocale + ".daysAbbreviated[t.Weekday()]...)\n"

			// Full/Wide
			case 4:

				results += "b = append(b, " + baseLocale + ".daysWide[t.Weekday()]...)\n"
			}

			// skip over E's
			i += count - 1

		// era eg. AD, BC
		case 'G':

			if inConstantText {
				inConstantText = false
				results += "b = append(b, " + fmt.Sprintf("%#v", []byte(format[start:i])) + "...)\n"
			}

			switch eraScore {
			case 0:
				results += `

				if t.Year() < 0 {
					b = append(b, ` + baseLocale + `.erasWide[0]...)
				} else {
					b = append(b, ` + baseLocale + `.erasWide[1]...)
				}

			`
			case 1, 2:
				results += `

				if t.Year() < 0 {
					b = append(b, ` + baseLocale + `.erasAbbreviated[0]...)
				} else {
					b = append(b, ` + baseLocale + `.erasAbbreviated[1]...)
				}

			`
			}

		default:
			// append all non matched text as they are constants
			if !inConstantText {
				inConstantText = true
				start = i
			}
		}
	}

	// if we were inConstantText when the string ended, add what's left.
	if inConstantText {
		// inContantText = false
		results += "b = append(b, " + fmt.Sprintf("%#v", []byte(format[start:])) + "...)\n"
	}

	return
}

func parseCurrencyNumberFormat(trans *translator) {

	if len(trans.CurrencyNumberFormat) == 0 {
		return
	}

	trans.FmtCurrencyExists = true
	negativeEqual := trans.CurrencyNumberFormat == trans.NegativeCurrencyNumberFormat

	match := groupLenRegex.FindString(trans.CurrencyNumberFormat)
	if len(match) > 0 {
		trans.FmtCurrencyGroupLen = len(match) - 2
	}

	match = requiredDecimalRegex.FindString(trans.CurrencyNumberFormat)
	if len(match) > 0 {
		trans.FmtCurrencyMinDecimalLen = len(match) - 1
	}

	match = secondaryGroupLenRegex.FindString(trans.CurrencyNumberFormat)
	if len(match) > 0 {
		trans.FmtCurrencySecondaryGroupLen = len(match) - 2
	}

	idx := 0

	for idx = 0; idx < len(trans.CurrencyNumberFormat); idx++ {
		if trans.CurrencyNumberFormat[idx] == '#' || trans.CurrencyNumberFormat[idx] == '0' {
			trans.FmtCurrencyPrefix = trans.CurrencyNumberFormat[:idx]
			break
		}
	}

	for idx = len(trans.CurrencyNumberFormat) - 1; idx >= 0; idx-- {
		if trans.CurrencyNumberFormat[idx] == '#' || trans.CurrencyNumberFormat[idx] == '0' {
			idx++
			trans.FmtCurrencySuffix = trans.CurrencyNumberFormat[idx:]
			break
		}
	}

	for idx = 0; idx < len(trans.FmtCurrencyPrefix); idx++ {
		if trans.FmtCurrencyPrefix[idx] == '¤' {

			trans.FmtCurrencyInPrefix = true
			trans.FmtCurrencyPrefix = strings.Replace(trans.FmtCurrencyPrefix, string(trans.FmtCurrencyPrefix[idx]), "", 1)

			if idx == 0 {
				trans.FmtCurrencyLeft = true
			} else {
				trans.FmtCurrencyLeft = false
			}

			break
		}
	}

	for idx = 0; idx < len(trans.FmtCurrencySuffix); idx++ {
		if trans.FmtCurrencySuffix[idx] == '¤' {

			trans.FmtCurrencyInPrefix = false
			trans.FmtCurrencySuffix = strings.Replace(trans.FmtCurrencySuffix, string(trans.FmtCurrencySuffix[idx]), "", 1)

			if idx == 0 {
				trans.FmtCurrencyLeft = true
			} else {
				trans.FmtCurrencyLeft = false
			}

			break
		}
	}

	// if len(trans.FmtCurrencyPrefix) > 0 {
	// 	trans.FmtCurrencyPrefix = fmt.Sprintf("%#v", []byte(trans.FmtCurrencyPrefix))
	// }

	// if len(trans.FmtCurrencySuffix) > 0 {
	// 	trans.FmtCurrencySuffix = fmt.Sprintf("%#v", []byte(trans.FmtCurrencySuffix))
	// }

	// no need to parse again if true....
	if negativeEqual {

		trans.FmtCurrencyNegativePrefix = trans.FmtCurrencyPrefix
		trans.FmtCurrencyNegativeSuffix = trans.FmtCurrencySuffix
		trans.FmtCurrencyNegativeInPrefix = trans.FmtCurrencyInPrefix
		trans.FmtCurrencyNegativeLeft = trans.FmtCurrencyLeft

		return
	}

	trans.FmtCurrencyNegativeExists = true

	for idx = 0; idx < len(trans.NegativeCurrencyNumberFormat); idx++ {
		if trans.NegativeCurrencyNumberFormat[idx] == '#' || trans.NegativeCurrencyNumberFormat[idx] == '0' {

			trans.FmtCurrencyNegativePrefix = trans.NegativeCurrencyNumberFormat[:idx]
			break
		}
	}

	for idx = len(trans.NegativeCurrencyNumberFormat) - 1; idx >= 0; idx-- {
		if trans.NegativeCurrencyNumberFormat[idx] == '#' || trans.NegativeCurrencyNumberFormat[idx] == '0' {
			idx++
			trans.FmtCurrencyNegativeSuffix = trans.NegativeCurrencyNumberFormat[idx:]
			break
		}
	}

	for idx = 0; idx < len(trans.FmtCurrencyNegativePrefix); idx++ {
		if trans.FmtCurrencyNegativePrefix[idx] == '¤' {

			trans.FmtCurrencyNegativeInPrefix = true
			trans.FmtCurrencyNegativePrefix = strings.Replace(trans.FmtCurrencyNegativePrefix, string(trans.FmtCurrencyNegativePrefix[idx]), "", 1)

			if idx == 0 {
				trans.FmtCurrencyNegativeLeft = true
			} else {
				trans.FmtCurrencyNegativeLeft = false
			}

			break
		}
	}

	for idx = 0; idx < len(trans.FmtCurrencyNegativeSuffix); idx++ {
		if trans.FmtCurrencyNegativeSuffix[idx] == '¤' {

			trans.FmtCurrencyNegativeInPrefix = false
			trans.FmtCurrencyNegativeSuffix = strings.Replace(trans.FmtCurrencyNegativeSuffix, string(trans.FmtCurrencyNegativeSuffix[idx]), "", 1)

			if idx == 0 {
				trans.FmtCurrencyNegativeLeft = true
			} else {
				trans.FmtCurrencyNegativeLeft = false
			}

			break
		}
	}

	// if len(trans.FmtCurrencyNegativePrefix) > 0 {
	// 	trans.FmtCurrencyNegativePrefix = fmt.Sprintf("%#v", []byte(trans.FmtCurrencyNegativePrefix))
	// }

	// if len(trans.FmtCurrencyNegativeSuffix) > 0 {
	// 	trans.FmtCurrencyNegativeSuffix = fmt.Sprintf("%#v", []byte(trans.FmtCurrencyNegativeSuffix))
	// }

	return
}

func parsePercentNumberFormat(trans *translator) {

	if len(trans.PercentNumberFormat) == 0 {
		return
	}

	trans.FmtPercentExists = true

	match := groupLenPercentRegex.FindString(trans.PercentNumberFormat)
	if len(match) > 0 {
		trans.FmtPercentGroupLen = len(match) - 1
	}

	match = requiredDecimalRegex.FindString(trans.PercentNumberFormat)
	if len(match) > 0 {
		trans.FmtPercentMinDecimalLen = len(match) - 1
	}

	match = secondaryGroupLenRegex.FindString(trans.PercentNumberFormat)
	if len(match) > 0 {
		trans.FmtPercentSecondaryGroupLen = len(match) - 2
	}

	idx := 0

	for idx = 0; idx < len(trans.PercentNumberFormat); idx++ {
		if trans.PercentNumberFormat[idx] == '#' || trans.PercentNumberFormat[idx] == '0' {
			trans.FmtPercentPrefix = trans.PercentNumberFormat[:idx]
			break
		}
	}

	for idx = len(trans.PercentNumberFormat) - 1; idx >= 0; idx-- {
		if trans.PercentNumberFormat[idx] == '#' || trans.PercentNumberFormat[idx] == '0' {
			idx++
			trans.FmtPercentSuffix = trans.PercentNumberFormat[idx:]
			break
		}
	}

	for idx = 0; idx < len(trans.FmtPercentPrefix); idx++ {
		if trans.FmtPercentPrefix[idx] == '%' {

			trans.FmtPercentInPrefix = true
			trans.FmtPercentPrefix = strings.Replace(trans.FmtPercentPrefix, string(trans.FmtPercentPrefix[idx]), "", 1)

			if idx == 0 {
				trans.FmtPercentLeft = true
			} else {
				trans.FmtPercentLeft = false
			}

			break
		}
	}

	for idx = 0; idx < len(trans.FmtPercentSuffix); idx++ {
		if trans.FmtPercentSuffix[idx] == '%' {

			trans.FmtPercentInPrefix = false
			trans.FmtPercentSuffix = strings.Replace(trans.FmtPercentSuffix, string(trans.FmtPercentSuffix[idx]), "", 1)

			if idx == 0 {
				trans.FmtPercentLeft = true
			} else {
				trans.FmtPercentLeft = false
			}

			break
		}
	}

	// if len(trans.FmtPercentPrefix) > 0 {
	// 	trans.FmtPercentPrefix = fmt.Sprintf("%#v", []byte(trans.FmtPercentPrefix))
	// }

	// if len(trans.FmtPercentSuffix) > 0 {
	// 	trans.FmtPercentSuffix = fmt.Sprintf("%#v", []byte(trans.FmtPercentSuffix))
	// }

	return
}

func parseDecimalNumberFormat(trans *translator) {

	if len(trans.DecimalNumberFormat) == 0 {
		return
	}

	trans.FmtNumberExists = true

	formats := strings.SplitN(trans.DecimalNumberFormat, ";", 2)

	match := groupLenRegex.FindString(formats[0])
	if len(match) > 0 {
		trans.FmtNumberGroupLen = len(match) - 2
	}

	match = requiredDecimalRegex.FindString(formats[0])
	if len(match) > 0 {
		trans.FmtNumberMinDecimalLen = len(match) - 1
	}

	match = secondaryGroupLenRegex.FindString(formats[0])
	if len(match) > 0 {
		trans.FmtNumberSecondaryGroupLen = len(match) - 2
	}

	return
}

type sortRank struct {
	Rank  uint8
	Value string
}

type ByRank []sortRank

func (a ByRank) Len() int           { return len(a) }
func (a ByRank) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByRank) Less(i, j int) bool { return a[i].Rank < a[j].Rank }

type ByPluralRule []locales.PluralRule

func (a ByPluralRule) Len() int           { return len(a) }
func (a ByPluralRule) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPluralRule) Less(i, j int) bool { return a[i] < a[j] }

// TODO: refine generated code a bit, some combinations end up with same plural rule,
// could check all at once; but it works and that's step 1 complete
func parseRangePluralRuleFunc(current *cldr.CLDR, baseLocale string) (results string, plurals string) {

	var pluralRange *struct {
		cldr.Common
		Locales     string `xml:"locales,attr"`
		PluralRange []*struct {
			cldr.Common
			Start  string `xml:"start,attr"`
			End    string `xml:"end,attr"`
			Result string `xml:"result,attr"`
		} `xml:"pluralRange"`
	}

	var pluralArr []locales.PluralRule

	for _, pr := range current.Supplemental().Plurals[1].PluralRanges {

		locs := strings.Split(pr.Locales, " ")

		for _, loc := range locs {

			if loc == baseLocale {
				pluralRange = pr
			}
		}
	}

	// no range plural rules for locale
	if pluralRange == nil {
		plurals = "nil"
		results = "return locales.PluralRuleUnknown"
		return
	}

	mp := make(map[string]struct{})

	// pre-process if all the same
	for _, rule := range pluralRange.PluralRange {
		mp[rule.Result] = struct{}{}
	}

	for k := range mp {
		psI := pluralStringToInt(k)
		pluralArr = append(pluralArr, psI)
	}

	if len(mp) == 1 {
		results += "return locales." + pluralStringToString(pluralRange.PluralRange[0].Result)
		plurals = fmt.Sprintf("%#v", pluralArr)
		return
	}

	multiple := len(pluralRange.PluralRange) > 1

	if multiple {
		results += "start := " + baseLocale + ".CardinalPluralRule(num1, v1)\n"
		results += "end := " + baseLocale + ".CardinalPluralRule(num2, v2)\n\n"
	}

	first := true

	// pre parse for variables
	for i, rule := range pluralRange.PluralRange {

		if i == len(pluralRange.PluralRange)-1 {

			if multiple {
				results += "\n\n"
			}
			results += "return locales." + pluralStringToString(rule.Result)
			continue
		}

		if first {
			results += "if"
			first = false
		} else {
			results += "else if"
		}

		results += " start == locales." + pluralStringToString(rule.Start) + " && end == locales." + pluralStringToString(rule.End) + " {\n return locales." + pluralStringToString(rule.Result) + "\n} "

	}

	if multiple {
		results = "\n" + results + "\n"
	}

	if len(pluralArr) == 0 {
		plurals = "nil"
	} else {

		ints := make([]int, len(pluralArr))
		for i := 0; i < len(pluralArr); i++ {
			ints[i] = int(pluralArr[i])
		}

		sort.Ints(ints)

		for i := 0; i < len(ints); i++ {
			pluralArr[i] = locales.PluralRule(ints[i])
		}

		plurals = fmt.Sprintf("%#v", pluralArr)
	}

	return

}

// TODO: cleanup function logic perhaps write a lexer... but it's working right now, and
// I'm already farther down the rabbit hole than I'd like and so pulling the chute here.
func parseOrdinalPluralRuleFunc(current *cldr.CLDR, baseLocale string) (results string, plurals string) {

	var prOrdinal *struct {
		cldr.Common
		Locales    string "xml:\"locales,attr\""
		PluralRule []*struct {
			cldr.Common
			Count string "xml:\"count,attr\""
		} "xml:\"pluralRule\""
	}

	var pluralArr []locales.PluralRule

	// idx 0 is ordinal rules
	for _, pr := range current.Supplemental().Plurals[0].PluralRules {

		locs := strings.Split(pr.Locales, " ")

		for _, loc := range locs {

			if loc == baseLocale {

				prOrdinal = pr
				// for _, pl := range pr.PluralRule {
				// 	fmt.Println(pl.Count, pl.Common.Data())
				// }
			}
		}
	}

	// no plural rules for locale
	if prOrdinal == nil {
		plurals = "nil"
		results = "return locales.PluralRuleUnknown"
		return
	}

	vals := make(map[string]struct{})
	first := true

	// pre parse for variables
	for _, rule := range prOrdinal.PluralRule {

		ps1 := pluralStringToString(rule.Count)
		psI := pluralStringToInt(rule.Count)
		pluralArr = append(pluralArr, psI)

		data := strings.Replace(strings.Replace(strings.Replace(strings.TrimSpace(strings.SplitN(rule.Common.Data(), "@", 2)[0]), " = ", " == ", -1), " or ", " || ", -1), " and ", " && ", -1)

		if len(data) == 0 {
			if len(prOrdinal.PluralRule) == 1 {

				results = "return locales." + ps1

			} else {

				results += "\n\nreturn locales." + ps1
				// results += "else {\nreturn locales." + locales.PluralStringToString(rule.Count) + ", nil\n}"
			}

			continue
		}

		// // All need n, so always add
		// if strings.Contains(data, "n") {
		// 	vals[prVarFuncs["n"]] = struct{}{}
		// }

		if strings.Contains(data, "i") {
			vals[prVarFuncs["i"]] = struct{}{}
		}

		// v is inherently avaialable as an argument
		// if strings.Contains(data, "v") {
		// 	vals[prVarFuncs["v"]] = struct{}{}
		// }

		if strings.Contains(data, "w") {
			vals[prVarFuncs["w"]] = struct{}{}
		}

		if strings.Contains(data, "f") {
			vals[prVarFuncs["f"]] = struct{}{}
		}

		if strings.Contains(data, "t") {
			vals[prVarFuncs["t"]] = struct{}{}
		}

		if first {
			results += "if "
			first = false
		} else {
			results += "else if "
		}

		stmt := ""

		// real work here
		//
		// split by 'or' then by 'and' allowing to better
		// determine bracketing for formula

		ors := strings.Split(data, "||")

		for _, or := range ors {

			stmt += "("

			ands := strings.Split(strings.TrimSpace(or), "&&")

			for _, and := range ands {

				inArg := false
				pre := ""
				lft := ""
				preOperator := ""
				args := strings.Split(strings.TrimSpace(and), " ")

				for _, a := range args {

					if inArg {
						// check to see if is a value range 2..9

						multiRange := strings.Count(a, "..") > 1
						cargs := strings.Split(strings.TrimSpace(a), ",")
						hasBracket := len(cargs) > 1
						bracketAdded := false
						lastWasRange := false

						for _, carg := range cargs {

							if rng := strings.Split(carg, ".."); len(rng) > 1 {

								if multiRange {
									pre += " ("
								} else {
									pre += " "
								}

								switch preOperator {
								case "==":
									pre += lft + " >= " + rng[0] + " && " + lft + "<=" + rng[1]
								case "!=":
									pre += "(" + lft + " < " + rng[0] + " || " + lft + " > " + rng[1] + ")"
								}

								if multiRange {
									pre += ") || "
								} else {
									pre += " || "
								}

								lastWasRange = true
								continue
							}

							if lastWasRange {
								pre = strings.TrimRight(pre, " || ") + " && "
							}

							lastWasRange = false

							if hasBracket && !bracketAdded {
								pre += "("
								bracketAdded = true
							}

							// single comma separated values
							switch preOperator {
							case "==":
								pre += " " + lft + preOperator + carg + " || "
							case "!=":
								pre += " " + lft + preOperator + carg + " && "
							}

						}

						pre = strings.TrimRight(pre, " || ")
						pre = strings.TrimRight(pre, " && ")
						pre = strings.TrimRight(pre, " || ")

						if hasBracket && bracketAdded {
							pre += ")"
						}

						continue
					}

					if strings.Contains(a, "=") || a == ">" || a == "<" {
						inArg = true
						preOperator = a
						continue
					}

					lft += a
				}

				stmt += pre + " && "
			}

			stmt = strings.TrimRight(stmt, " && ") + ") || "
		}

		stmt = strings.TrimRight(stmt, " || ")

		results += stmt

		results += " {\n"

		// return plural rule here
		results += "return locales." + ps1 + "\n"

		results += "}"
	}

	pre := "\n"

	// always needed
	vals[prVarFuncs["n"]] = struct{}{}

	sorted := make([]sortRank, 0, len(vals))

	for k := range vals {
		switch k[:1] {
		case "n":
			sorted = append(sorted, sortRank{
				Value: prVarFuncs["n"],
				Rank:  1,
			})
		case "i":
			sorted = append(sorted, sortRank{
				Value: prVarFuncs["i"],
				Rank:  2,
			})
		case "w":
			sorted = append(sorted, sortRank{
				Value: prVarFuncs["w"],
				Rank:  3,
			})
		case "f":
			sorted = append(sorted, sortRank{
				Value: prVarFuncs["f"],
				Rank:  4,
			})
		case "t":
			sorted = append(sorted, sortRank{
				Value: prVarFuncs["t"],
				Rank:  5,
			})
		}
	}

	sort.Sort(ByRank(sorted))

	for _, k := range sorted {
		pre += k.Value
	}

	if len(results) == 0 {
		results = "return locales.PluralRuleUnknown"
	} else {

		if !strings.HasPrefix(results, "return") {

			results = manyToSingleVars(results)
			// pre += "\n"
			results = pre + results
		}
	}

	if len(pluralArr) == 0 {
		plurals = "nil"
	} else {
		plurals = fmt.Sprintf("%#v", pluralArr)
	}

	return
}

// TODO: cleanup function logic perhaps write a lexer... but it's working right now, and
// I'm already farther down the rabbit hole than I'd like and so pulling the chute here.
//
// updated to also accept actual locale as 'pt_PT' exists in cardinal rules different from 'pt'
func parseCardinalPluralRuleFunc(current *cldr.CLDR, locale, baseLocale string) (results string, plurals string) {

	var prCardinal *struct {
		cldr.Common
		Locales    string "xml:\"locales,attr\""
		PluralRule []*struct {
			cldr.Common
			Count string "xml:\"count,attr\""
		} "xml:\"pluralRule\""
	}

	var pluralArr []locales.PluralRule
	var inBaseLocale bool
	l := locale

FIND:
	// idx 2 is cardinal rules
	for _, pr := range current.Supplemental().Plurals[2].PluralRules {

		locs := strings.Split(pr.Locales, " ")

		for _, loc := range locs {

			if loc == l {
				prCardinal = pr
			}
		}
	}

	// no plural rules for locale
	if prCardinal == nil {

		if !inBaseLocale {
			inBaseLocale = true
			l = baseLocale
			goto FIND
		}

		plurals = "nil"
		results = "return locales.PluralRuleUnknown"
		return
	}

	vals := make(map[string]struct{})
	first := true

	// pre parse for variables
	for _, rule := range prCardinal.PluralRule {

		ps1 := pluralStringToString(rule.Count)
		psI := pluralStringToInt(rule.Count)
		pluralArr = append(pluralArr, psI)

		data := strings.Replace(strings.Replace(strings.Replace(strings.TrimSpace(strings.SplitN(rule.Common.Data(), "@", 2)[0]), " = ", " == ", -1), " or ", " || ", -1), " and ", " && ", -1)

		if len(data) == 0 {
			if len(prCardinal.PluralRule) == 1 {

				results = "return locales." + ps1

			} else {

				results += "\n\nreturn locales." + ps1
				// results += "else {\nreturn locales." + locales.PluralStringToString(rule.Count) + ", nil\n}"
			}

			continue
		}

		// // All need n, so always add
		// if strings.Contains(data, "n") {
		// 	vals[prVarFuncs["n"]] = struct{}{}
		// }

		if strings.Contains(data, "i") {
			vals[prVarFuncs["i"]] = struct{}{}
		}

		// v is inherently avaialable as an argument
		// if strings.Contains(data, "v") {
		// 	vals[prVarFuncs["v"]] = struct{}{}
		// }

		if strings.Contains(data, "w") {
			vals[prVarFuncs["w"]] = struct{}{}
		}

		if strings.Contains(data, "f") {
			vals[prVarFuncs["f"]] = struct{}{}
		}

		if strings.Contains(data, "t") {
			vals[prVarFuncs["t"]] = struct{}{}
		}

		if first {
			results += "if "
			first = false
		} else {
			results += "else if "
		}

		stmt := ""

		// real work here
		//
		// split by 'or' then by 'and' allowing to better
		// determine bracketing for formula

		ors := strings.Split(data, "||")

		for _, or := range ors {

			stmt += "("

			ands := strings.Split(strings.TrimSpace(or), "&&")

			for _, and := range ands {

				inArg := false
				pre := ""
				lft := ""
				preOperator := ""
				args := strings.Split(strings.TrimSpace(and), " ")

				for _, a := range args {

					if inArg {
						// check to see if is a value range 2..9

						multiRange := strings.Count(a, "..") > 1
						cargs := strings.Split(strings.TrimSpace(a), ",")
						hasBracket := len(cargs) > 1
						bracketAdded := false
						lastWasRange := false

						for _, carg := range cargs {

							if rng := strings.Split(carg, ".."); len(rng) > 1 {

								if multiRange {
									pre += " ("
								} else {
									pre += " "
								}

								switch preOperator {
								case "==":
									pre += lft + " >= " + rng[0] + " && " + lft + "<=" + rng[1]
								case "!=":
									pre += "(" + lft + " < " + rng[0] + " || " + lft + " > " + rng[1] + ")"
								}

								if multiRange {
									pre += ") || "
								} else {
									pre += " || "
								}

								lastWasRange = true
								continue
							}

							if lastWasRange {
								pre = strings.TrimRight(pre, " || ") + " && "
							}

							lastWasRange = false

							if hasBracket && !bracketAdded {
								pre += "("
								bracketAdded = true
							}

							// single comma separated values
							switch preOperator {
							case "==":
								pre += " " + lft + preOperator + carg + " || "
							case "!=":
								pre += " " + lft + preOperator + carg + " && "
							}

						}

						pre = strings.TrimRight(pre, " || ")
						pre = strings.TrimRight(pre, " && ")
						pre = strings.TrimRight(pre, " || ")

						if hasBracket && bracketAdded {
							pre += ")"
						}

						continue
					}

					if strings.Contains(a, "=") || a == ">" || a == "<" {
						inArg = true
						preOperator = a
						continue
					}

					lft += a
				}

				stmt += pre + " && "
			}

			stmt = strings.TrimRight(stmt, " && ") + ") || "
		}

		stmt = strings.TrimRight(stmt, " || ")

		results += stmt

		results += " {\n"

		// return plural rule here
		results += "return locales." + ps1 + "\n"

		results += "}"
	}

	pre := "\n"

	// always needed
	vals[prVarFuncs["n"]] = struct{}{}

	sorted := make([]sortRank, 0, len(vals))

	for k := range vals {
		switch k[:1] {
		case "n":
			sorted = append(sorted, sortRank{
				Value: prVarFuncs["n"],
				Rank:  1,
			})
		case "i":
			sorted = append(sorted, sortRank{
				Value: prVarFuncs["i"],
				Rank:  2,
			})
		case "w":
			sorted = append(sorted, sortRank{
				Value: prVarFuncs["w"],
				Rank:  3,
			})
		case "f":
			sorted = append(sorted, sortRank{
				Value: prVarFuncs["f"],
				Rank:  4,
			})
		case "t":
			sorted = append(sorted, sortRank{
				Value: prVarFuncs["t"],
				Rank:  5,
			})
		}
	}

	sort.Sort(ByRank(sorted))

	for _, k := range sorted {
		pre += k.Value
	}

	if len(results) == 0 {
		results = "return locales.PluralRuleUnknown"
	} else {

		if !strings.HasPrefix(results, "return") {

			results = manyToSingleVars(results)
			// pre += "\n"
			results = pre + results
		}
	}

	if len(pluralArr) == 0 {
		plurals = "nil"
	} else {
		plurals = fmt.Sprintf("%#v", pluralArr)
	}

	return
}

func manyToSingleVars(input string) (results string) {

	matches := nModRegex.FindAllString(input, -1)
	mp := make(map[string][]string) // map of formula to variable
	var found bool
	var split []string
	var variable string

	for _, formula := range matches {

		if _, found = mp[formula]; found {
			continue
		}

		split = strings.SplitN(formula, "%", 2)

		mp[formula] = []string{split[1], "math.Mod(" + split[0] + ", " + split[1] + ")"}
	}

	for k, v := range mp {
		variable = "nMod" + v[0]
		results += variable + " := " + v[1] + "\n"
		input = strings.Replace(input, k, variable, -1)
	}

	matches = iModRegex.FindAllString(input, -1)
	mp = make(map[string][]string) // map of formula to variable

	for _, formula := range matches {

		if _, found = mp[formula]; found {
			continue
		}

		split = strings.SplitN(formula, "%", 2)

		mp[formula] = []string{split[1], formula}
	}

	for k, v := range mp {
		variable = "iMod" + v[0]
		results += variable + " := " + v[1] + "\n"
		input = strings.Replace(input, k, variable, -1)
	}

	matches = wModRegex.FindAllString(input, -1)
	mp = make(map[string][]string) // map of formula to variable

	for _, formula := range matches {

		if _, found = mp[formula]; found {
			continue
		}

		split = strings.SplitN(formula, "%", 2)

		mp[formula] = []string{split[1], formula}
	}

	for k, v := range mp {
		variable = "wMod" + v[0]
		results += variable + " := " + v[1] + "\n"
		input = strings.Replace(input, k, variable, -1)
	}

	matches = fModRegex.FindAllString(input, -1)
	mp = make(map[string][]string) // map of formula to variable

	for _, formula := range matches {

		if _, found = mp[formula]; found {
			continue
		}

		split = strings.SplitN(formula, "%", 2)

		mp[formula] = []string{split[1], formula}
	}

	for k, v := range mp {
		variable = "fMod" + v[0]
		results += variable + " := " + v[1] + "\n"
		input = strings.Replace(input, k, variable, -1)
	}

	matches = tModRegex.FindAllString(input, -1)
	mp = make(map[string][]string) // map of formula to variable

	for _, formula := range matches {

		if _, found = mp[formula]; found {
			continue
		}

		split = strings.SplitN(formula, "%", 2)

		mp[formula] = []string{split[1], formula}
	}

	for k, v := range mp {
		variable = "tMod" + v[0]
		results += variable + " := " + v[1] + "\n"
		input = strings.Replace(input, k, variable, -1)
	}

	results = results + "\n" + input

	return
}

// pluralStringToInt returns the enum value of 'plural' provided
func pluralStringToInt(plural string) locales.PluralRule {

	switch plural {
	case "zero":
		return locales.PluralRuleZero
	case "one":
		return locales.PluralRuleOne
	case "two":
		return locales.PluralRuleTwo
	case "few":
		return locales.PluralRuleFew
	case "many":
		return locales.PluralRuleMany
	case "other":
		return locales.PluralRuleOther
	default:
		return locales.PluralRuleUnknown
	}
}

func pluralStringToString(pr string) string {

	pr = strings.TrimSpace(pr)

	switch pr {
	case "zero":
		return "PluralRuleZero"
	case "one":
		return "PluralRuleOne"
	case "two":
		return "PluralRuleTwo"
	case "few":
		return "PluralRuleFew"
	case "many":
		return "PluralRuleMany"
	case "other":
		return "PluralRuleOther"
	default:
		return "PluralRuleUnknown"
	}
}
