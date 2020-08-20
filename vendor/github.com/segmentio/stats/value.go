package stats

import (
	"math"
	"strconv"
	"time"
)

type Value struct {
	typ  Type
	pad  int32
	bits uint64
}

func ValueOf(v interface{}) Value {
	switch x := v.(type) {
	case nil:
		return Value{}
	case bool:
		return boolValue(x)
	case int:
		return intValue(x)
	case int8:
		return int8Value(x)
	case int16:
		return int16Value(x)
	case int32:
		return int32Value(x)
	case int64:
		return int64Value(x)
	case uint:
		return uintValue(x)
	case uint8:
		return uint8Value(x)
	case uint16:
		return uint16Value(x)
	case uint32:
		return uint32Value(x)
	case uint64:
		return uint64Value(x)
	case uintptr:
		return uintptrValue(x)
	case float32:
		return float32Value(x)
	case float64:
		return float64Value(x)
	case time.Duration:
		return durationValue(x)
	default:
		panic("stats.ValueOf received a value of unsupported type")
	}
}

func boolValue(v bool) Value {
	return Value{typ: Bool, bits: boolBits(v)}
}

func intValue(v int) Value {
	return int64Value(int64(v))
}

func int8Value(v int8) Value {
	return int64Value(int64(v))
}

func int16Value(v int16) Value {
	return int64Value(int64(v))
}

func int32Value(v int32) Value {
	return int64Value(int64(v))
}

func int64Value(v int64) Value {
	return Value{typ: Int, bits: uint64(v)}
}

func uintValue(v uint) Value {
	return uint64Value(uint64(v))
}

func uint8Value(v uint8) Value {
	return uint64Value(uint64(v))
}

func uint16Value(v uint16) Value {
	return uint64Value(uint64(v))
}

func uint32Value(v uint32) Value {
	return uint64Value(uint64(v))
}

func uint64Value(v uint64) Value {
	return Value{typ: Uint, bits: v}
}

func uintptrValue(v uintptr) Value {
	return uint64Value(uint64(v))
}

func float32Value(v float32) Value {
	return float64Value(float64(v))
}

func float64Value(v float64) Value {
	return Value{typ: Float, bits: math.Float64bits(v)}
}

func durationValue(v time.Duration) Value {
	return Value{typ: Duration, bits: uint64(v)}
}

func (v Value) Type() Type {
	return v.typ
}

func (v Value) Bool() bool {
	return v.bits != 0
}

func (v Value) Int() int64 {
	return int64(v.bits)
}

func (v Value) Uint() uint64 {
	return v.bits
}

func (v Value) Float() float64 {
	return math.Float64frombits(v.bits)
}

func (v Value) Duration() time.Duration {
	return time.Duration(v.bits)
}

func (v Value) Interface() interface{} {
	switch v.Type() {
	case Null:
		return nil
	case Bool:
		return v.Bool()
	case Int:
		return v.Int()
	case Uint:
		return v.Uint()
	case Float:
		return v.Float()
	case Duration:
		return v.Duration()
	default:
		panic("unknown type found in a stats.Value")
	}
}

func (v Value) String() string {
	switch v.Type() {
	case Null:
		return "<nil>"
	case Bool:
		return strconv.FormatBool(v.Bool())
	case Int:
		return strconv.FormatInt(v.Int(), 10)
	case Uint:
		return strconv.FormatUint(v.Uint(), 10)
	case Float:
		return strconv.FormatFloat(v.Float(), 'g', -1, 64)
	case Duration:
		return v.Duration().String()
	default:
		return "<unknown>"
	}
}

type Type int32

const (
	Null Type = iota
	Bool
	Int
	Uint
	Float
	Duration
)

func (t Type) String() string {
	switch t {
	case Null:
		return "<nil>"
	case Bool:
		return "bool"
	case Int:
		return "int64"
	case Uint:
		return "uint64"
	case Float:
		return "float64"
	case Duration:
		return "time.Duration"
	default:
		return "<unknown>"
	}
}

func (t Type) GoString() string {
	switch t {
	case Null:
		return "stats.Null"
	case Bool:
		return "stats.Bool"
	case Int:
		return "stats.Int"
	case Uint:
		return "stats.Uint"
	case Float:
		return "stats.Float"
	case Duration:
		return "stats.Duration"
	default:
		return "stats.Type(" + strconv.Itoa(int(t)) + ")"
	}
}

func boolBits(v bool) uint64 {
	if v {
		return 1
	}
	return 0
}
