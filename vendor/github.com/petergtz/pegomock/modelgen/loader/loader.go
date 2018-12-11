package loader

import (
	"errors"
	"fmt"
	"go/ast"
	"go/types"

	"github.com/petergtz/pegomock/model"
	"golang.org/x/tools/go/loader"
)

func GenerateModel(importPath string, interfaceName string) (*model.Package, error) {
	var conf loader.Config
	conf.Import(importPath)
	program, e := conf.Load()
	if e != nil {
		panic(e)
	}
	info := program.Imported[importPath]

	for def := range info.Defs {
		if def.Name == interfaceName && def.Obj.Kind == ast.Typ {
			interfacetype, ok := def.Obj.Decl.(*ast.TypeSpec).Type.(*ast.InterfaceType)
			if ok {
				g := &modelGenerator{info: info}
				iface := &model.Interface{
					Name:    interfaceName,
					Methods: g.modelMethodsFrom(interfacetype.Methods),
				}
				return &model.Package{
					Name:       info.Pkg.Name(),
					Interfaces: []*model.Interface{iface},
				}, nil
			}
		}
	}

	return nil, errors.New("Did not find interface name \"" + interfaceName + "\"")
}

type modelGenerator struct {
	info *loader.PackageInfo
}

func (g *modelGenerator) modelMethodsFrom(astMethods *ast.FieldList) (modelMethods []*model.Method) {
	for _, astMethod := range astMethods.List {
		modelMethods = append(modelMethods, g.modelMethodFrom(astMethod))
	}
	return
}

func (g *modelGenerator) modelMethodFrom(astMethod *ast.Field) *model.Method {
	in, out, variadic := g.signatureFrom(astMethod.Type.(*ast.FuncType))
	return &model.Method{Name: astMethod.Names[0].Name, In: in, Variadic: variadic, Out: out}
}

func (g *modelGenerator) signatureFrom(astFuncType *ast.FuncType) (in, out []*model.Parameter, variadic *model.Parameter) {
	in, variadic = g.inParamsFrom(astFuncType.Params)
	out = g.outParamsFrom(astFuncType.Results)
	return
}

func (g *modelGenerator) inParamsFrom(params *ast.FieldList) (in []*model.Parameter, variadic *model.Parameter) {
	for _, param := range params.List {
		for _, name := range param.Names {
			if ellipsisType, isEllipsisType := param.Type.(*ast.Ellipsis); isEllipsisType {
				variadic = g.newParam(name.Name, ellipsisType.Elt)
			} else {
				in = append(in, g.newParam(name.Name, param.Type))
			}
		}
		if len(param.Names) == 0 {
			if ellipsisType, isEllipsisType := param.Type.(*ast.Ellipsis); isEllipsisType {
				variadic = g.newParam("", ellipsisType.Elt)
			} else {
				in = append(in, g.newParam("", param.Type))
			}
		}
	}
	return
}

func (g *modelGenerator) outParamsFrom(results *ast.FieldList) (out []*model.Parameter) {
	if results != nil {
		for _, param := range results.List {
			for _, name := range param.Names {
				out = append(out, g.newParam(name.Name, param.Type))
			}
			if len(param.Names) == 0 {
				out = append(out, g.newParam("", param.Type))
			}
		}
	}
	return
}

func (g *modelGenerator) newParam(name string, typ ast.Expr) *model.Parameter {
	return &model.Parameter{
		Name: name,
		Type: g.modelTypeFrom(g.info.TypeOf(typ)),
	}
}

func (g *modelGenerator) modelTypeFrom(typesType types.Type) model.Type {
	switch typedTyp := typesType.(type) {
	case *types.Basic:
		if !predeclared(typedTyp.Kind()) {
			panic(fmt.Sprintf("Unexpected Basic Type %v", typedTyp.Name()))
		}
		return model.PredeclaredType(typedTyp.Name())
	case *types.Pointer:
		return &model.PointerType{
			Type: g.modelTypeFrom(typedTyp.Elem()),
		}
	case *types.Array:
		return &model.ArrayType{
			Len:  int(typedTyp.Len()),
			Type: g.modelTypeFrom(typedTyp.Elem()),
		}
	case *types.Slice:
		return &model.ArrayType{
			Len:  -1,
			Type: g.modelTypeFrom(typedTyp.Elem()),
		}
	case *types.Map:
		return &model.MapType{
			Key:   g.modelTypeFrom(typedTyp.Key()),
			Value: g.modelTypeFrom(typedTyp.Elem()),
		}
	case *types.Chan:
		return &model.ChanType{
			Dir:  model.ChanDir(typedTyp.Dir()),
			Type: g.modelTypeFrom(typedTyp.Elem()),
		}
	case *types.Named:
		if typedTyp.Obj().Pkg() == nil {
			return model.PredeclaredType(typedTyp.Obj().Name())
		}
		return &model.NamedType{
			Package: typedTyp.Obj().Pkg().Path(),
			Type:    typedTyp.Obj().Name(),
		}
	case *types.Interface:
		return model.PredeclaredType(typedTyp.String())
	case *types.Signature:
		in, variadic := g.generateInParamsFrom(typedTyp.Params())
		out := g.generateOutParamsFrom(typedTyp.Results())
		return &model.FuncType{In: in, Out: out, Variadic: variadic}
	default:
		panic(fmt.Sprintf("Unknown types.Type: %v (%T)", typesType, typesType))
	}
}

func (g *modelGenerator) generateInParamsFrom(params *types.Tuple) (in []*model.Parameter, variadic *model.Parameter) {
	// TODO: variadic

	for i := 0; i < params.Len(); i++ {
		in = append(in, &model.Parameter{
			Name: params.At(i).Name(),
			Type: g.modelTypeFrom(params.At(i).Type()),
		})
	}
	return
}

func (g *modelGenerator) generateOutParamsFrom(params *types.Tuple) (out []*model.Parameter) {
	for i := 0; i < params.Len(); i++ {
		out = append(out, &model.Parameter{
			Name: params.At(i).Name(),
			Type: g.modelTypeFrom(params.At(i).Type()),
		})
	}
	return
}

func predeclared(basicKind types.BasicKind) bool {
	return basicKind >= types.Bool && basicKind <= types.String
}
