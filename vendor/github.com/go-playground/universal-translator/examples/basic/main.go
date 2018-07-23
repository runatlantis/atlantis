package main

import (
	"fmt"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/en_CA"
	"github.com/go-playground/locales/fr"
	"github.com/go-playground/locales/nl"
	"github.com/go-playground/universal-translator"
)

// only one instance as translators within are shared + goroutine safe
var universalTraslator *ut.UniversalTranslator

func main() {

	// NOTE: this example is omitting a lot of error checking for brevity
	e := en.New()
	universalTraslator = ut.New(e, e, en_CA.New(), nl.New(), fr.New())

	en, _ := universalTraslator.GetTranslator("en")

	// generally used after parsing an http 'Accept-Language' header
	// and this will try to find a matching locale you support or
	// fallback locale.
	// en, _ := ut.FindTranslator([]string{"en", "en_CA", "nl"})

	// this will help
	fmt.Println("Cardinal Plural Rules:", en.PluralsCardinal())
	fmt.Println("Ordinal Plural Rules:", en.PluralsOrdinal())
	fmt.Println("Range Plural Rules:", en.PluralsRange())

	// add basic language only translations
	// last param indicates if it's ok to override the translation if one already exists
	en.Add("welcome", "Welcome {0} to our test", false)

	// add language translations dependant on cardinal plural rules
	en.AddCardinal("days", "You have {0} day left to register", locales.PluralRuleOne, false)
	en.AddCardinal("days", "You have {0} days left to register", locales.PluralRuleOther, false)

	// add language translations dependant on ordinal plural rules
	en.AddOrdinal("day-of-month", "{0}st", locales.PluralRuleOne, false)
	en.AddOrdinal("day-of-month", "{0}nd", locales.PluralRuleTwo, false)
	en.AddOrdinal("day-of-month", "{0}rd", locales.PluralRuleFew, false)
	en.AddOrdinal("day-of-month", "{0}th", locales.PluralRuleOther, false)

	// add language translations dependant on range plural rules
	// NOTE: only one plural rule for range in 'en' locale
	en.AddRange("between", "It's {0}-{1} days away", locales.PluralRuleOther, false)

	// now lets use the translations we just added, in the same order we added them

	fmt.Println(en.T("welcome", "Joeybloggs"))

	fmt.Println(en.C("days", 1, 0, en.FmtNumber(1, 0))) // you'd normally have variables defined for 1 and 0
	fmt.Println(en.C("days", 2, 0, en.FmtNumber(2, 0)))
	fmt.Println(en.C("days", 10456.25, 2, en.FmtNumber(10456.25, 2)))

	fmt.Println(en.O("day-of-month", 1, 0, en.FmtNumber(1, 0)))
	fmt.Println(en.O("day-of-month", 2, 0, en.FmtNumber(2, 0)))
	fmt.Println(en.O("day-of-month", 3, 0, en.FmtNumber(3, 0)))
	fmt.Println(en.O("day-of-month", 4, 0, en.FmtNumber(4, 0)))
	fmt.Println(en.O("day-of-month", 10456.25, 0, en.FmtNumber(10456.25, 0)))

	fmt.Println(en.R("between", 0, 0, 1, 0, en.FmtNumber(0, 0), en.FmtNumber(1, 0)))
	fmt.Println(en.R("between", 1, 0, 2, 0, en.FmtNumber(1, 0), en.FmtNumber(2, 0)))
	fmt.Println(en.R("between", 1, 0, 100, 0, en.FmtNumber(1, 0), en.FmtNumber(100, 0)))
}
