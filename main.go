package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"os"
	"unicode"
	"unicode/utf8"
)

var (
	fname      = flag.String("f", "", "file name")
	structName = flag.String("s", "", "model struct")
)

func main() {
	flag.Parse()
	if *fname == "" {
		log.Print("flag -f needed")
		flag.Usage()
		os.Exit(1)
	}
	if *structName == "" {
		log.Print("flag -s needed")
		flag.Usage()
		os.Exit(1)
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, *fname, nil, 0)
	if err != nil {
		log.Fatal(err)
	}

	info := types.Info{
		Defs: make(map[*ast.Ident]types.Object),
	}
	conf := types.Config{
		Importer: importer.Default(),
	}
	pkg, err := conf.Check("", fset, []*ast.File{f}, &info)
	if err != nil {
		log.Fatal(err)
	}
	g := &Generator{
		pkg:        pkg,
		structName: *structName,
		imports:    make(map[string]*types.Package),
		GenFile:    &GenFile{},
	}

	var strct *types.Struct
	for d, v := range info.Defs {
		if d.Name != g.structName {
			continue
		}
		strct = v.Type().Underlying().(*types.Struct)
		break
	}
	if strct == nil {
		log.Fatalf("struct %q not found", g.structName)
	}

	for i := 0; i < strct.NumFields(); i++ {
		f := strct.Field(i)
		typ := g.parse(f.Type())
		field := Field{
			name: f.Name(),
			typ:  typ,
		}
		g.fields = append(g.fields, field)
	}

	source := g.generate()
	source, err = format.Source(source)
	if err != nil {
		log.Fatalf("failed to format source: %v", err)
	}
	fmt.Printf("%s", source)

}

type GenFile struct {
	buf bytes.Buffer
}

func (g *GenFile) P(v ...interface{}) {
	for _, x := range v {
		fmt.Fprint(&g.buf, x)
	}
	fmt.Fprintln(&g.buf)
}

type Field struct {
	name string
	typ  string
}

type Generator struct {
	pkg        *types.Package
	structName string
	imports    map[string]*types.Package
	fields     []Field
	*GenFile
}

func (g *Generator) generate() []byte {
	g.P("// Code generated by cfgen. DO NOT EDIT.")
	g.P()
	g.P("package ", g.pkg.Name())
	g.P()
	g.P("import (")
	for _, pkg := range g.imports {
		g.P(fmt.Sprintf("%q", pkg.Path()))
	}
	g.P(")")
	g.P()
	g.P("var (")
	for _, f := range g.fields {
		g.P(fmt.Sprintf("default%s %s", f.name, f.typ))
	}
	g.P(")")
	g.P()
	g.P("type base struct{}")
	g.P()
	for _, f := range g.fields {
		g.P(fmt.Sprintf("func (base) %s() %s { return default%s }", f.name, f.typ, f.name))
	}
	g.P()
	g.P(fmt.Sprintf("func New() %s {", g.structName))
	g.P("return base{}")
	g.P("}")
	g.P()
	g.P(fmt.Sprintf("type %s interface {", g.structName))
	for _, f := range g.fields {
		g.P(fmt.Sprintf("%s() %s", f.name, f.typ))
	}
	g.P("}")
	g.P()
	g.set()
	return g.GenFile.buf.Bytes()
}

func (g *Generator) set() {
	for _, f := range g.fields {
		isLower := false
		r, size := utf8.DecodeRuneInString(f.name)
		l := unicode.ToLower(r)
		if r == l {
			isLower = true
		}
		field := string(l) + f.name[size:]
		if isLower {
			field = "_" + field
		}

		g.P(fmt.Sprintf("type cfg%s struct {", f.name))
		g.P(g.structName)
		g.P(fmt.Sprintf("%s %s", field, f.typ))
		g.P("}")
		g.P()
		g.P(fmt.Sprintf("func (cfg cfg%s) %s() %s {", f.name, f.name, f.typ))
		g.P(fmt.Sprintf("return cfg.%s", field))
		g.P("}")
		g.P()
		g.P(fmt.Sprintf("func Set%s(cfg %s, %c %s) %s {", f.name, g.structName, l, f.typ, g.structName))
		g.P(fmt.Sprintf("return cfg%s{", f.name))
		g.P(fmt.Sprintf("%s: cfg,", g.structName))
		g.P(fmt.Sprintf("%s: %c,", field, l))
		g.P("}")
		g.P("}")
	}
}

func (g *Generator) parse(typ types.Type) string {
	switch t := typ.(type) {
	case *types.Basic:
		return g.parseBasic(t)
	case *types.Named:
		return g.parseNamed(t)
	case *types.Array:
		return g.parseArray(t)
	case *types.Pointer:
		return g.parsePointer(t)
	case *types.Struct:
		return g.parseStruct(t)
	case *types.Signature:
		return g.parseSignature(t)
	case *types.Slice:
		return g.parseSlice(t)
	case *types.Map:
		return g.parseMap(t)
	case *types.Interface:
		return g.parseInterface(t)
	case *types.Chan:
		return g.parseChan(t)
	default:
	}
	return "nil"
}

func (g *Generator) parseBasic(t *types.Basic) string {
	return t.Name()
}

func (g *Generator) parseNamed(t *types.Named) string {
	name := ""
	pkg := t.Obj().Pkg()
	if pkg != nil && pkg.Path() != g.pkg.Path() {
		g.imports[pkg.Path()] = pkg
		name += pkg.Name() + "."
	}
	name += t.Obj().Name()
	return name
}

func (g *Generator) parseArray(t *types.Array) string {
	typ := g.parse(t.Elem())
	return fmt.Sprintf("[%d]%s", t.Len(), typ)
}

func (g *Generator) parsePointer(t *types.Pointer) string {
	return "*" + g.parse(t.Elem())
}

func (g *Generator) parseStruct(t *types.Struct) string {
	out := "struct{"
	for i := 0; i < t.NumFields(); i++ {
		f := t.Field(i)
		name, typ := f.Name(), g.parse(f.Type())
		out += fmt.Sprintf("%s %s;", name, typ)
	}
	out += "}"
	return out
}

func (g *Generator) parseSignature(t *types.Signature) string {
	params := g.tuple(t.Params())
	results := g.tuple(t.Results())
	return fmt.Sprintf("func %s %s", params, results)
}

func (g *Generator) parseSlice(t *types.Slice) string {
	return "[]" + g.parse(t.Elem())
}

func (g *Generator) parseMap(t *types.Map) string {
	k := g.parse(t.Key())
	v := g.parse(t.Elem())
	return fmt.Sprintf("map[%s]%s", k, v)
}

func (g *Generator) parseInterface(t *types.Interface) string {
	out := "interface{"
	for i := 0; i < t.NumEmbeddeds(); i++ {
		out += g.parse(t.EmbeddedType(i)) + ";"
	}
	for i := 0; i < t.NumExplicitMethods(); i++ {
		m := t.ExplicitMethod(i)
		typ := g.parse(m.Type())
		method := m.Name() + typ[len("func"):]
		out += method + ";"
	}
	out += "}"
	return out
}

func (g *Generator) parseChan(t *types.Chan) string {
	out := "chan"
	switch t.Dir() {
	case types.SendOnly:
		out += "<-"
	case types.RecvOnly:
		out = "<-" + out
	default:
	}
	out += " " + g.parse(t.Elem())
	return out
}

func (g *Generator) tuple(t *types.Tuple) string {
	out := "("
	for i := 0; i < t.Len(); i++ {
		f := t.At(i)
		name, typ := f.Name(), g.parse(f.Type())
		out += fmt.Sprintf("%s %s,", name, typ)
	}
	out += ")"
	return out
}
