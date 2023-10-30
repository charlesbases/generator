package generator

import (
	"testing"
)

func TestImport(t *testing.T) {
	var (
		bufio    = NewPackage("bufio").Alias("_")
		bytes    = NewPackage("bytes").Alias("_")
		fmt      = NewPackage("fmt").Alias("_")
		ioutil   = NewPackage("io/ioutil").Alias("_")
		flag     = NewPackage("flag").Alias("fg")
		exec     = NewPackage("os/exec").Alias("exec")
		path     = NewPackage("path")
		filepath = NewPackage("path/filepath").Alias("_")
		err1     = NewPackage("errors")
		err2     = NewPackage("github.com/pkg/errors")

		protogen = NewPackage("google.golang.org/protobuf/compiler/protogen").Alias("_")
	)

	Run(func(p *Plugin) {
		{
			f := p.NewFile("import_test.go", "testing", bufio, bytes, fmt, ioutil, filepath, protogen, flag, exec, path, err1, err2)
			f.Writer("package testing")

			f.Import(NewPackage("time"))
			f.Writer("  var _ =", "time.Now()")
		}
	})
}

func TestGenerator(t *testing.T) {

}
