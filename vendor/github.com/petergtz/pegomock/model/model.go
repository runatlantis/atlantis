// Copyright 2012 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package model contains the data model necessary for generating mock implementations.
package model

import (
	"fmt"
	"io"
	"strings"
)

// Package is a Go package. It may be a subset.
type Package struct {
	Name       string
	Interfaces []*Interface
	DotImports []string
}

func (pkg *Package) Print(w io.Writer) {
	fmt.Fprintf(w, "package %s\n", pkg.Name)
	for _, intf := range pkg.Interfaces {
		intf.Print(w)
	}
}

// Imports returns the imports needed by the Package as a set of import paths.
func (pkg *Package) Imports() map[string]bool {
	im := make(map[string]bool)
	for _, intf := range pkg.Interfaces {
		intf.addImports(im)
	}
	return im
}

// Interface is a Go interface.
type Interface struct {
	Name    string
	Methods []*Method
}

func (intf *Interface) Print(w io.Writer) {
	fmt.Fprintf(w, "interface %s\n", intf.Name)
	for _, m := range intf.Methods {
		m.Print(w)
	}
}

func (intf *Interface) addImports(im map[string]bool) {
	for _, m := range intf.Methods {
		m.addImports(im)
	}
}

// Method is a single method of an interface.
type Method struct {
	Name     string
	In, Out  []*Parameter
	Variadic *Parameter // may be nil
}

func (m *Method) Print(w io.Writer) {
	fmt.Fprintf(w, "  - method %s\n", m.Name)
	if len(m.In) > 0 {
		fmt.Fprintf(w, "    in:\n")
		for _, p := range m.In {
			p.Print(w)
		}
	}
	if m.Variadic != nil {
		fmt.Fprintf(w, "    ...:\n")
		m.Variadic.Print(w)
	}
	if len(m.Out) > 0 {
		fmt.Fprintf(w, "    out:\n")
		for _, p := range m.Out {
			p.Print(w)
		}
	}
}

func (m *Method) addImports(im map[string]bool) {
	for _, p := range m.In {
		p.Type.addImports(im)
	}
	if m.Variadic != nil {
		m.Variadic.Type.addImports(im)
	}
	for _, p := range m.Out {
		p.Type.addImports(im)
	}
}

// Parameter is an argument or return parameter of a method.
type Parameter struct {
	Name string // may be empty
	Type Type
}

func (p *Parameter) Print(w io.Writer) {
	n := p.Name
	if n == "" {
		n = `""`
	}
	fmt.Fprintf(w, "    - %v: %v\n", n, p.Type.String(nil, ""))
}

// Type is a Go type.
type Type interface {
	String(pm map[string]string, pkgOverride string) string
	addImports(im map[string]bool)
}

// ArrayType is an array or slice type.
type ArrayType struct {
	Len  int // -1 for slices, >= 0 for arrays
	Type Type
}

func (at *ArrayType) String(pm map[string]string, pkgOverride string) string {
	s := "[]"
	if at.Len > -1 {
		s = fmt.Sprintf("[%d]", at.Len)
	}
	return s + at.Type.String(pm, pkgOverride)
}

func (at *ArrayType) addImports(im map[string]bool) { at.Type.addImports(im) }

// ChanType is a channel type.
type ChanType struct {
	Dir  ChanDir // 0, 1 or 2
	Type Type
}

func (ct *ChanType) String(pm map[string]string, pkgOverride string) string {
	s := ct.Type.String(pm, pkgOverride)
	if ct.Dir == RecvDir {
		return "<-chan " + s
	}
	if ct.Dir == SendDir {
		return "chan<- " + s
	}
	return "chan " + s
}

func (ct *ChanType) addImports(im map[string]bool) { ct.Type.addImports(im) }

// ChanDir is a channel direction.
type ChanDir int

const (
	RecvDir ChanDir = 1
	SendDir ChanDir = 2
)

// FuncType is a function type.
type FuncType struct {
	In, Out  []*Parameter
	Variadic *Parameter // may be nil
}

func (ft *FuncType) String(pm map[string]string, pkgOverride string) string {
	args := make([]string, len(ft.In))
	for i, p := range ft.In {
		args[i] = p.Type.String(pm, pkgOverride)
	}
	if ft.Variadic != nil {
		args = append(args, "..."+ft.Variadic.Type.String(pm, pkgOverride))
	}
	rets := make([]string, len(ft.Out))
	for i, p := range ft.Out {
		rets[i] = p.Type.String(pm, pkgOverride)
	}
	retString := strings.Join(rets, ", ")
	if nOut := len(ft.Out); nOut == 1 {
		retString = " " + retString
	} else if nOut > 1 {
		retString = " (" + retString + ")"
	}
	return "func(" + strings.Join(args, ", ") + ")" + retString
}

func (ft *FuncType) addImports(im map[string]bool) {
	for _, p := range ft.In {
		p.Type.addImports(im)
	}
	if ft.Variadic != nil {
		ft.Variadic.Type.addImports(im)
	}
	for _, p := range ft.Out {
		p.Type.addImports(im)
	}
}

// MapType is a map type.
type MapType struct {
	Key, Value Type
}

func (mt *MapType) String(pm map[string]string, pkgOverride string) string {
	return "map[" + mt.Key.String(pm, pkgOverride) + "]" + mt.Value.String(pm, pkgOverride)
}

func (mt *MapType) addImports(im map[string]bool) {
	mt.Key.addImports(im)
	mt.Value.addImports(im)
}

// NamedType is an exported type in a package.
type NamedType struct {
	Package string // may be empty
	Type    string // TODO: should this be typed Type?
}

func (nt *NamedType) String(pm map[string]string, pkgOverride string) string {
	// TODO: is this right?
	if pkgOverride == nt.Package {
		return nt.Type
	}
	return pm[nt.Package] + "." + nt.Type
}
func (nt *NamedType) addImports(im map[string]bool) {
	if nt.Package != "" {
		im[nt.Package] = true
	}
}

// PointerType is a pointer to another type.
type PointerType struct {
	Type Type
}

func (pt *PointerType) String(pm map[string]string, pkgOverride string) string {
	return "*" + pt.Type.String(pm, pkgOverride)
}
func (pt *PointerType) addImports(im map[string]bool) { pt.Type.addImports(im) }

// PredeclaredType is a predeclared type such as "int".
type PredeclaredType string

func (pt PredeclaredType) String(pm map[string]string, pkgOverride string) string { return string(pt) }
func (pt PredeclaredType) addImports(im map[string]bool)                          {}
