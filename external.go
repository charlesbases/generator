package generator

// Package is the import path of a Go package.
// for example: "google.golang.org/protobuf/compiler/protogen"
type Package struct {
	// path 包路径
	path string
	// alias 别名
	alias string
	// standard 是否标准库包
	standard bool
}

// NewPackage .
// 简洁包名例如: "github.com/pkg/errors", 程序可以自处理 alias
// 复杂包名例如: "github.com/redis/go-redis/v9", "github.com/robfig/cron/v3", 需手动调用 Alias()，否则会导致生成的文件出错
func NewPackage(pkg string) *Package {
	return &Package{path: pkg}
}

// Alias .
func (exp *Package) Alias(alias string) *Package {
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
