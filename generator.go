package generator

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const rfc = "2006-01-02 15:04:05.000"

// Plugin .
type Plugin struct {
	files []*generatedFile
}

// generatedFile .
type generatedFile struct {
	filename    string
	packagepath string

	buf *bytes.Buffer

	externalPackages map[string]*Package
	usedPackageNames map[string]struct{}
}

// Run .
func Run(fn func(plugin *Plugin)) error {
	return run(fn)
}

// run .
func run(fn func(p *Plugin)) error {
	var p = &Plugin{files: make([]*generatedFile, 0)}
	fn(p)
	return p.output()
}

// NewFile .
func (p *Plugin) NewFile(name string, path string, exts ...*Package) *generatedFile {
	f := &generatedFile{
		filename:         name,
		packagepath:      abs(path),
		buf:              new(bytes.Buffer),
		externalPackages: make(map[string]*Package),
		usedPackageNames: make(map[string]struct{}),
	}
	p.files = append(p.files, f)

	f.imports(exts...)
	f.Writer("// Code generated by https://github.com/charlesbases/reverse. DO NOT EDIT.")
	f.Writer("// date:", time.Now().Format(rfc), "\n")

	return f
}

// Import .
func (f *generatedFile) Import(exts ...*Package) {
	f.imports(exts...)
}

// Writer .
func (f *generatedFile) Writer(v ...any) {
	if len(v) != 0 {
		fmt.Fprint(f.buf, v...)
	}
	fmt.Fprintln(f.buf)
}

// imports .
func (f *generatedFile) imports(exts ...*Package) {
	for _, ext := range exts {
		ext.standard = IsStandard(ext.path)

		if ext.alias == "" {
			ext.alias = trim(filepath.Base(ext.path))
		}

		if _, found := f.externalPackages[ext.path]; !found {
			f.externalPackages[ext.path] = ext
			if ext.alias != "." && ext.alias != "_" {
				if _, found := f.usedPackageNames[ext.alias]; found {
					var index = 1
					for {
						if _, found := f.usedPackageNames[ext.alias+strconv.Itoa(index)]; found {
							index++
						} else {
							ext.alias += strconv.Itoa(index)
							break
						}
					}
				}
			}
			f.usedPackageNames[ext.alias] = struct{}{}
		}
	}
}

// content .
func (f *generatedFile) content() ([]byte, error) {
	if !strings.HasSuffix(f.filename, ".go") {
		return f.buf.Bytes(), nil
	}

	// Reformat generated code.
	original := f.buf.Bytes()
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, "", original, parser.ParseComments)
	if err != nil {
		// Print out the bad code with line numbers.
		// This should never happen in practice, but it can while changing generated code
		// so consider this a debugging aid.
		var src bytes.Buffer
		s := bufio.NewScanner(bytes.NewReader(original))
		for line := 1; s.Scan(); line++ {
			fmt.Fprintf(&src, "%5d\t%s\n", line, s.Bytes())
		}
		return nil, fmt.Errorf("%s: unparsable Go source: %v\n%s", f.filename, err, src.String())
	}

	var imports = make([]*Package, 0, len(f.externalPackages))
	for _, externalPackage := range f.externalPackages {
		imports = append(imports, externalPackage)
	}

	sort.Slice(imports, func(i, j int) bool {
		if imports[i].standard && !imports[j].standard {
			return true
		}
		if imports[j].standard && !imports[i].standard {
			return false
		}
		return imports[i].path < imports[j].path
	})

	// Modify the AST to include a new import block.
	if len(imports) > 0 {
		// Insert block after package statement or
		// possible comment attached to the end of the package statement.
		pos := file.Package
		tokFile := fileSet.File(file.Package)
		pkgLine := tokFile.Line(file.Package)
		for _, c := range file.Comments {
			if tokFile.Line(c.Pos()) > pkgLine {
				break
			}
			pos = c.End()
		}

		// Construct the import block.
		decl := &ast.GenDecl{
			Tok:    token.IMPORT,
			TokPos: pos,
			Lparen: pos,
			Rparen: pos,
		}
		for _, pkg := range imports {
			if pkg.alias == filepath.Base(pkg.path) {
				pkg.alias = ""
			}

			decl.Specs = append(decl.Specs, &ast.ImportSpec{
				Name: &ast.Ident{
					Name:    pkg.alias,
					NamePos: pos,
				},
				Path: &ast.BasicLit{
					Kind:     token.STRING,
					Value:    strconv.Quote(pkg.path),
					ValuePos: pos,
				},
				EndPos: pos,
			})
		}
		file.Decls = append([]ast.Decl{decl}, file.Decls...)
	}

	var out bytes.Buffer
	if err = (&printer.Config{Mode: printer.TabIndent | printer.UseSpaces, Tabwidth: 8}).Fprint(&out, fileSet, file); err != nil {
		return nil, fmt.Errorf("%s: can not reformat Go source: %v", f.filename, err)
	}
	return out.Bytes(), nil
}

// output .
func (p *Plugin) output() error {
	for _, f := range p.files {
		content, err := f.content()
		if err != nil {
			return err
		}

		if err := os.MkdirAll(f.packagepath, 0755); err != nil {
			return err
		}

		file, err := os.OpenFile(filepath.Join(f.packagepath, f.filename), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0664)
		if err != nil {
			return err
		}
		file.Write(content)
		file.Close()
	}

	return p.format()
}

// Format .
func (p *Plugin) format() error {
	var args = make([]string, 0, len(p.files)+1)
	args = append(args, "-w")
	for _, file := range p.files {
		if strings.HasSuffix(file.filename, ".go") {
			args = append(args, filepath.Join(file.packagepath, file.filename))
		}
	}
	if len(args) > 1 {
		// use goimports
		if exec.Command("goimports", args...).Run() != nil {
			return errors.New("`goimports` not found, please execute `go install golang.org/x/tools/cmd/goimports@latest` to install.")
		}
	}
	return nil
}

func abs(p string) string {
	p, _ = filepath.Abs(p)
	return p
}
