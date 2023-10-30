package generator

import (
	"testing"
)

func TestImport(t *testing.T) {
	var (
		bufio    = ExternalPackage("bufio").Alias("_")
		bytes    = ExternalPackage("bytes").Alias("_")
		fmt      = ExternalPackage("fmt").Alias("_")
		ioutil   = ExternalPackage("io/ioutil").Alias("_")
		flag     = ExternalPackage("flag").Alias("fg")
		exec     = ExternalPackage("os/exec").Alias("exec")
		path     = ExternalPackage("path")
		filepath = ExternalPackage("path/filepath").Alias("_")
		err1     = ExternalPackage("errors")
		err2     = ExternalPackage("github.com/pkg/errors")

		protogen = ExternalPackage("google.golang.org/protobuf/compiler/protogen").Alias("_")
	)

	Run(func(p *Plugin) {
		{
			f := p.NewFile("import_test.go", "testing", bufio, bytes, fmt, ioutil, filepath, protogen, flag, exec, path, err1, err2)
			f.Writer("package testing")

			f.Import(ExternalPackage("time"))
			f.Writer("  var _ =", "time.Now()")
		}
	})
}

func TestGenerator(t *testing.T) {

}
