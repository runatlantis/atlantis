package main

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
)

// Generate matchers:
//
//     go generate github.com/petergtz/pegomock/internal/generate_matchers

//go:generate go run matcher_generation.go

func main() {
	err := ioutil.WriteFile(
		"../../matcher_factories.go",
		[]byte(GenerateDefaultMatchersFile()),
		0644)
	if err != nil {
		panic(err)
	}
}

func GenerateDefaultMatchersFile() string {
	return fmt.Sprintf(`package pegomock

import (
	"reflect"
)

%s
	`, GenerateDefaultMatchers())
}

func GenerateDefaultMatchers() string {
	result := ""
	for _, kind := range primitiveKinds {
		result += GenerateEqMatcherFactory(kind) +
			GenerateAnyMatcherFactory(kind) +
			GenerateAnySliceMatcherFactory(kind)
	}
	return result
}

var primitiveKinds = []reflect.Kind{
	reflect.Bool,
	reflect.Int,
	reflect.Int8,
	reflect.Int16,
	reflect.Int32,
	reflect.Int64,
	reflect.Uint,
	reflect.Uint8,
	reflect.Uint16,
	reflect.Uint32,
	reflect.Uint64,
	reflect.Uintptr,
	reflect.Float32,
	reflect.Float64,
	reflect.Complex64,
	reflect.Complex128,
	reflect.String,
}

func GenerateEqMatcherFactory(kind reflect.Kind) string {
	return fmt.Sprintf(`func Eq%s(value %s) %s {
	RegisterMatcher(&EqMatcher{Value: value})
	return %s
}

`, strings.Title(kind.String()), kind, kind, nullOf(kind))
}

func GenerateAnyMatcherFactory(kind reflect.Kind) string {
	return fmt.Sprintf(`func Any%s() %s {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf((%s)(%s))))
	return %s
}

`, strings.Title(kind.String()), kind, kind.String(), nullOf(kind), nullOf(kind))
}

func GenerateAnySliceMatcherFactory(kind reflect.Kind) string {
	return fmt.Sprintf(`func Any%sSlice() []%s {
	RegisterMatcher(NewAnyMatcher(reflect.SliceOf(reflect.TypeOf((%s)(%s)))))
	return nil
}

`, strings.Title(kind.String()), kind.String(), kind.String(), nullOf(kind))
}

// TODO generate:
// Eq Slice matchers
// generate chan, func matchers

func nullOf(kind reflect.Kind) string {
	switch {
	case kind == reflect.Bool:
		return `false`
	case reflect.Int <= kind && kind <= reflect.Complex128:
		return `0`
	case kind == reflect.String:
		return `""`
	default:
		return "nil"
	}
}
