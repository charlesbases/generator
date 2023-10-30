package generator

import "path/filepath"

// externalPackage is the import path of a Go package.
// for example: "google.golang.org/protobuf/compiler/protogen"
type externalPackage struct {
	// path 包路径
	path string
	// alias 别名
	alias string
	// standard 是否标准库包
	standard bool
}

// ExternalPackage .
func ExternalPackage(pkg string) *externalPackage {
	return &externalPackage{path: pkg, alias: trim(filepath.Base(pkg)), standard: IsStandard(pkg)}
}

// Alias .
func (exp *externalPackage) Alias(alias string) *externalPackage {
	exp.alias = alias
	return exp
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
