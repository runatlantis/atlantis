package stats

import (
	"reflect"
	"time"
	"unsafe"
)

type structField struct {
	typ reflect.Type
	off uintptr
}

func (f structField) pointer(ptr unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(uintptr(ptr) + f.off)
}

func (f structField) value(ptr unsafe.Pointer) reflect.Value {
	return reflect.NewAt(f.typ, f.pointer(ptr))
}

func (f structField) bool(ptr unsafe.Pointer) bool {
	return *(*bool)(f.pointer(ptr))
}

func (f structField) int(ptr unsafe.Pointer) int {
	return *(*int)(f.pointer(ptr))
}

func (f structField) int8(ptr unsafe.Pointer) int8 {
	return *(*int8)(f.pointer(ptr))
}

func (f structField) int16(ptr unsafe.Pointer) int16 {
	return *(*int16)(f.pointer(ptr))
}

func (f structField) int32(ptr unsafe.Pointer) int32 {
	return *(*int32)(f.pointer(ptr))
}

func (f structField) int64(ptr unsafe.Pointer) int64 {
	return *(*int64)(f.pointer(ptr))
}

func (f structField) uint(ptr unsafe.Pointer) uint {
	return *(*uint)(f.pointer(ptr))
}

func (f structField) uint8(ptr unsafe.Pointer) uint8 {
	return *(*uint8)(f.pointer(ptr))
}

func (f structField) uint16(ptr unsafe.Pointer) uint16 {
	return *(*uint16)(f.pointer(ptr))
}

func (f structField) uint32(ptr unsafe.Pointer) uint32 {
	return *(*uint32)(f.pointer(ptr))
}

func (f structField) uint64(ptr unsafe.Pointer) uint64 {
	return *(*uint64)(f.pointer(ptr))
}

func (f structField) uintptr(ptr unsafe.Pointer) uintptr {
	return *(*uintptr)(f.pointer(ptr))
}

func (f structField) float32(ptr unsafe.Pointer) float32 {
	return *(*float32)(f.pointer(ptr))
}

func (f structField) float64(ptr unsafe.Pointer) float64 {
	return *(*float64)(f.pointer(ptr))
}

func (f structField) duration(ptr unsafe.Pointer) time.Duration {
	return *(*time.Duration)(f.pointer(ptr))
}

func (f structField) string(ptr unsafe.Pointer) string {
	return *(*string)(f.pointer(ptr))
}

var (
	boolType     = reflect.TypeOf(false)
	intType      = reflect.TypeOf(int(0))
	int8Type     = reflect.TypeOf(int8(0))
	int16Type    = reflect.TypeOf(int16(0))
	int32Type    = reflect.TypeOf(int32(0))
	int64Type    = reflect.TypeOf(int64(0))
	uintType     = reflect.TypeOf(uint(0))
	uint8Type    = reflect.TypeOf(uint8(0))
	uint16Type   = reflect.TypeOf(uint16(0))
	uint32Type   = reflect.TypeOf(uint32(0))
	uint64Type   = reflect.TypeOf(uint64(0))
	uintptrType  = reflect.TypeOf(uintptr(0))
	float32Type  = reflect.TypeOf(float32(0))
	float64Type  = reflect.TypeOf(float64(0))
	durationType = reflect.TypeOf(time.Duration(0))
	stringType   = reflect.TypeOf("")
)
