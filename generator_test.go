package generator

import "testing"

func TestImport(t *testing.T) {
	var (
		bufio    = NewExternalPackage("bufio").Alias("_")
		bytes    = NewExternalPackage("bytes").Alias("_")
		fmt      = NewExternalPackage("fmt").Alias("_")
		ioutil   = NewExternalPackage("io/ioutil").Alias("_")
		filepath = NewExternalPackage("path/filepath").Alias("_")

		protogen = NewExternalPackage("google.golang.org/protobuf/compiler/protogen").Alias("_")
	)

	Run(func(plugin *Plugin) error {
		{
			f := plugin.NewGeneratedFile("import_test.go", "testing", bufio, bytes, fmt, ioutil, filepath, protogen)
			f.Writer("// Code generated by generator. DO DOT EDIT.")
			f.Writer("package testing")
		}
		return nil
	})
}

func TestGenerator(t *testing.T) {

}
