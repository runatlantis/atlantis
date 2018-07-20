package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/fr"
	"github.com/go-playground/pure"
	"github.com/go-playground/pure/examples/middleware/logging-recovery"
	"github.com/go-playground/universal-translator"
)

var (
	tmpls    *template.Template
	utrans   *ut.UniversalTranslator
	transKey = struct {
		name string
	}{
		name: "transKey",
	}
)

// Translator wraps ut.Translator in order to handle errors transparently
// it is totally optional but recommended as it can now be used directly in
// templates and nobody can add translations where they're not supposed to.
type Translator interface {
	locales.Translator

	// creates the translation for the locale given the 'key' and params passed in.
	// wraps ut.Translator.T to handle errors
	T(key interface{}, params ...string) string

	// creates the cardinal translation for the locale given the 'key', 'num' and 'digit' arguments
	//  and param passed in.
	// wraps ut.Translator.C to handle errors
	C(key interface{}, num float64, digits uint64, param string) string

	// creates the ordinal translation for the locale given the 'key', 'num' and 'digit' arguments
	// and param passed in.
	// wraps ut.Translator.O to handle errors
	O(key interface{}, num float64, digits uint64, param string) string

	//  creates the range translation for the locale given the 'key', 'num1', 'digit1', 'num2' and
	//  'digit2' arguments and 'param1' and 'param2' passed in
	// wraps ut.Translator.R to handle errors
	R(key interface{}, num1 float64, digits1 uint64, num2 float64, digits2 uint64, param1, param2 string) string

	// Currency returns the type used by the given locale.
	Currency() currency.Type
}

// implements Translator interface definition above.
type translator struct {
	locales.Translator
	trans ut.Translator
}

var _ Translator = (*translator)(nil)

func (t *translator) T(key interface{}, params ...string) string {

	s, err := t.trans.T(key, params...)
	if err != nil {
		log.Printf("issue translating key: '%v' error: '%s'", key, err)
	}

	return s
}

func (t *translator) C(key interface{}, num float64, digits uint64, param string) string {

	s, err := t.trans.C(key, num, digits, param)
	if err != nil {
		log.Printf("issue translating cardinal key: '%v' error: '%s'", key, err)
	}

	return s
}

func (t *translator) O(key interface{}, num float64, digits uint64, param string) string {

	s, err := t.trans.C(key, num, digits, param)
	if err != nil {
		log.Printf("issue translating ordinal key: '%v' error: '%s'", key, err)
	}

	return s
}

func (t *translator) R(key interface{}, num1 float64, digits1 uint64, num2 float64, digits2 uint64, param1, param2 string) string {

	s, err := t.trans.R(key, num1, digits1, num2, digits2, param1, param2)
	if err != nil {
		log.Printf("issue translating range key: '%v' error: '%s'", key, err)
	}

	return s
}

func (t *translator) Currency() currency.Type {

	// choose your own locale. The reason it isn't mapped for you is because many
	// countries have multiple currencies; it's up to you and you're application how
	// and which currencies to use. I recommend adding a function it to to your custon translator
	// interface like defined above.
	switch t.Locale() {
	case "en":
		return currency.USD
	case "fr":
		return currency.EUR
	default:
		return currency.USD
	}
}

func main() {

	en := en.New()
	utrans = ut.New(en, en, fr.New())
	setup()

	tmpls, _ = template.ParseFiles("home.tmpl")

	r := pure.New()
	r.Use(middleware.LoggingAndRecovery(true), translatorMiddleware)
	r.Get("/", home)

	log.Println("Running on Port :8080")
	log.Println("Try me with URL http://localhost:8080/?locale=en")
	log.Println("and then http://localhost:8080/?locale=fr")
	http.ListenAndServe(":8080", r.Serve())
}

func home(w http.ResponseWriter, r *http.Request) {

	// get locale translator ( could be wrapped into a helper function )
	t := r.Context().Value(transKey).(Translator)

	s := struct {
		Trans       Translator
		Now         time.Time
		PositiveNum float64
		NegativeNum float64
		Percent     float64
	}{
		Trans:       t,
		Now:         time.Now(),
		PositiveNum: 1234576.45,
		NegativeNum: -35900394.34,
		Percent:     96.76,
	}

	if err := tmpls.ExecuteTemplate(w, "home", s); err != nil {
		log.Fatal(err)
	}
}

func translatorMiddleware(next http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// there are many ways to check, this is just checking for query param &
		// Accept-Language header but can be expanded to Cookie's etc....

		params := r.URL.Query()

		locale := params.Get("locale")
		var t ut.Translator

		if len(locale) > 0 {

			var found bool

			if t, found = utrans.GetTranslator(locale); found {
				goto END
			}
		}

		// get and parse the "Accept-Language" http header and return an array
		t, _ = utrans.FindTranslator(pure.AcceptedLanguages(r)...)
	END:
		// I would normally wrap ut.Translator with one with my own functions in order
		// to handle errors and be able to use all functions from translator within the templates.
		r = r.WithContext(context.WithValue(r.Context(), transKey, &translator{trans: t, Translator: t.(locales.Translator)}))

		next(w, r)
	}
}

func setup() {

	en, _ := utrans.FindTranslator("en")
	en.AddCardinal("days-left", "There is {0} day left", locales.PluralRuleOne, false)
	en.AddCardinal("days-left", "There are {0} days left", locales.PluralRuleOther, false)

	fr, _ := utrans.FindTranslator("fr")
	fr.AddCardinal("days-left", "Il reste {0} jour", locales.PluralRuleOne, false)
	fr.AddCardinal("days-left", "Il reste {0} jours", locales.PluralRuleOther, false)

	err := utrans.VerifyTranslations()
	if err != nil {
		log.Fatal(err)
	}
}
