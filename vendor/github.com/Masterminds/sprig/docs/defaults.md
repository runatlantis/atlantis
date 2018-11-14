# Default Functions

Sprig provides tools for setting default values for templates.

## default

To set a simple default value, use `default`:

```
default "foo" .Bar
```

In the above, if `.Bar` evaluates to a non-empty value, it will be used. But if
it is empty, `foo` will be returned instead.

The definition of "empty" depends on type:

- Numeric: 0
- String: ""
- Lists: `[]`
- Dicts: `{}`
- Boolean: `false`
- And always `nil` (aka null)

For structs, there is no definition of empty, so a struct will never return the
default.

## empty

The `empty` function returns `true` if the given value is considered empty, and
`false` otherwise. The empty values are listed in the `default` section.

```
empty .Foo
```

Note that in Go template conditionals, emptiness is calculated for you. Thus,
you rarely need `if empty .Foo`. Instead, just use `if .Foo`.

## coalesce

The `coalesce` function takes a list of values and returns the first non-empty
one.

```
coalesce 0 1 2
```

The above returns `1`.

This function is useful for scanning through multiple variables or values:

```
coalesce .name .parent.name "Matt"
```

The above will first check to see if `.name` is empty. If it is not, it will return
that value. If it _is_ empty, `coalesce` will evaluate `.parent.name` for emptiness.
Finally, if both `.name` and `.parent.name` are empty, it will return `Matt`.

## toJson

The `toJson` function encodes an item into a JSON string.

```
toJson .Item
```

The above returns JSON string representation of `.Item`.

## toPrettyJson

The `toPrettyJson` function encodes an item into a pretty (indented) JSON string.

```
toPrettyJson .Item
```

The above returns indented JSON string representation of `.Item`.

## ternary

The `ternary` function takes two values, and a test value. If the test value is
true, the first value will be returned. If the test value is empty, the second
value will be returned. This is similar to the c ternary operator.

### true test value

```
ternary "foo" "bar" true
```

or

```
true | ternary "foo" "bar"
```

The above returns `"foo"`.

### false test value

```
ternary "foo" "bar" false
```

or

```
false | ternary "foo" "bar"
```

The above returns `"bar"`.
