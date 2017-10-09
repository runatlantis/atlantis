package gomock

import (
	"encoding/gob"
	"fmt"
	"reflect"

	"github.com/petergtz/pegomock/model"
)

func init() {
	gob.Register(&model.ArrayType{})
	gob.Register(&model.ChanType{})
	gob.Register(&model.FuncType{})
	gob.Register(&model.MapType{})
	gob.Register(&model.NamedType{})
	gob.Register(&model.PointerType{})
	gob.Register(model.PredeclaredType(""))
}

func InterfaceFromInterfaceType(it reflect.Type) (*model.Interface, error) {
	if it.Kind() != reflect.Interface {
		return nil, fmt.Errorf("%v is not an interface", it)
	}
	intf := &model.Interface{}

	for i := 0; i < it.NumMethod(); i++ {
		mt := it.Method(i)
		// TODO: need to skip unexported methods? or just raise an error?
		m := &model.Method{
			Name: mt.Name,
		}

		var err error
		m.In, m.Variadic, m.Out, err = funcArgsFromType(mt.Type)
		if err != nil {
			return nil, err
		}

		intf.Methods = append(intf.Methods, m)
	}

	return intf, nil
}

// t's Kind must be a reflect.Func.
func funcArgsFromType(t reflect.Type) (in []*model.Parameter, variadic *model.Parameter, out []*model.Parameter, err error) {
	nin := t.NumIn()
	if t.IsVariadic() {
		nin--
	}
	var p *model.Parameter
	for i := 0; i < nin; i++ {
		p, err = parameterFromType(t.In(i))
		if err != nil {
			return
		}
		in = append(in, p)
	}
	if t.IsVariadic() {
		p, err = parameterFromType(t.In(nin).Elem())
		if err != nil {
			return
		}
		variadic = p
	}
	for i := 0; i < t.NumOut(); i++ {
		p, err = parameterFromType(t.Out(i))
		if err != nil {
			return
		}
		out = append(out, p)
	}
	return
}

func parameterFromType(t reflect.Type) (*model.Parameter, error) {
	tt, err := typeFromType(t)
	if err != nil {
		return nil, err
	}
	return &model.Parameter{Type: tt}, nil
}

var errorType = reflect.TypeOf((*error)(nil)).Elem()

var byteType = reflect.TypeOf(byte(0))

func typeFromType(t reflect.Type) (model.Type, error) {
	// Hack workaround for https://golang.org/issue/3853.
	// This explicit check should not be necessary.
	if t == byteType {
		return model.PredeclaredType("byte"), nil
	}

	if imp := t.PkgPath(); imp != "" {
		return &model.NamedType{
			Package: imp,
			Type:    t.Name(),
		}, nil
	}

	// only unnamed or predeclared types after here

	// Lots of types have element types. Let's do the parsing and error checking for all of them.
	var elemType model.Type
	switch t.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Ptr, reflect.Slice:
		var err error
		elemType, err = typeFromType(t.Elem())
		if err != nil {
			return nil, err
		}
	}

	switch t.Kind() {
	case reflect.Array:
		return &model.ArrayType{
			Len:  t.Len(),
			Type: elemType,
		}, nil
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.String:
		return model.PredeclaredType(t.Kind().String()), nil
	case reflect.Chan:
		var dir model.ChanDir
		switch t.ChanDir() {
		case reflect.RecvDir:
			dir = model.RecvDir
		case reflect.SendDir:
			dir = model.SendDir
		}
		return &model.ChanType{
			Dir:  dir,
			Type: elemType,
		}, nil
	case reflect.Func:
		in, variadic, out, err := funcArgsFromType(t)
		if err != nil {
			return nil, err
		}
		return &model.FuncType{
			In:       in,
			Out:      out,
			Variadic: variadic,
		}, nil
	case reflect.Interface:
		// Two special interfaces.
		if t.NumMethod() == 0 {
			return model.PredeclaredType("interface{}"), nil
		}
		if t == errorType {
			return model.PredeclaredType("error"), nil
		}
	case reflect.Map:
		kt, err := typeFromType(t.Key())
		if err != nil {
			return nil, err
		}
		return &model.MapType{
			Key:   kt,
			Value: elemType,
		}, nil
	case reflect.Ptr:
		return &model.PointerType{
			Type: elemType,
		}, nil
	case reflect.Slice:
		return &model.ArrayType{
			Len:  -1,
			Type: elemType,
		}, nil
	case reflect.Struct:
		if t.NumField() == 0 {
			return model.PredeclaredType("struct{}"), nil
		}
	}

	// TODO: Struct, UnsafePointer
	return nil, fmt.Errorf("can't yet turn %v (%v) into a model.Type", t, t.Kind())
}
