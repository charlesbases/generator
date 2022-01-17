package generator

import (
	"path"
	"strings"
)

// ExternalPackage is the import path of a Go package.
// for example: "google.golang.org/protobuf/compiler/protogen"
type ExternalPackage struct {
	Path  string
	alias string
}

// ExternalIdent is a Go identifier, consisting of a name and import path.
type ExternalIdent struct {
	Package *ExternalPackage
	Name    string
}

// NewExternalPackage .
func NewExternalPackage(p string) *ExternalPackage {
	return &ExternalPackage{Path: p, alias: trim(path.Base(p))}
}

// Alias .
func (exp *ExternalPackage) Alias(a string) *ExternalPackage {
	exp.alias = a
	return exp
}

// Ident .
func (extp *ExternalPackage) Ident(name string) ExternalIdent {
	return ExternalIdent{
		Package: extp,
		Name:    name,
	}
}

// string return ExternalPackage.Alias + "." + ExternalIdent.Name
func (exti *ExternalIdent) string() string {
	var bui strings.Builder
	bui.Grow(len(exti.Package.alias) + len(exti.Name) + 1)
	bui.WriteString(exti.Package.alias)
	bui.WriteString(".")
	bui.WriteString(exti.Name)
	return bui.String()
}

// trim .
func trim(s string) string {
	for i := len(s) - 1; i > -1; i-- {
		switch s[i] {
		case '-', '_':
			return s[i+1:]
		}
	}
	return s
}
