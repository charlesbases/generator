package generator

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Plugin .
type Plugin struct {
	files []*GeneratedFile
}

// GeneratedFile .
type GeneratedFile struct {
	filename    string
	packagepath string

	buf *bytes.Buffer

	externalPackages map[string]*ExternalPackage
	usedPackageNames map[string]bool
}

// Run .
func Run(fn func(plugin *Plugin) error) {
	if err := run(fn); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}

// run .
func run(fn func(plugin *Plugin) error) error {
	var plugin = &Plugin{files: make([]*GeneratedFile, 0)}

	if err := fn(plugin); err != nil {
		return err
	}

	return plugin.marshal()
}

// NewGeneratedFile .
// 初始化 externals 后，可不使用 ExternalPackage.Ident。须确保包名不冲突！
func (plugin *Plugin) NewGeneratedFile(name string, path string, externals ...*ExternalPackage) *GeneratedFile {
	f := &GeneratedFile{
		filename:         name,
		packagepath:      abs(path),
		buf:              new(bytes.Buffer),
		externalPackages: make(map[string]*ExternalPackage),
		usedPackageNames: make(map[string]bool),
	}

	for _, item := range externals {
		f.externalPackages[item.Path] = item
		f.usedPackageNames[item.alias] = true
	}

	plugin.files = append(plugin.files, f)
	return f
}

// Writer .
func (f *GeneratedFile) Writer(v ...interface{}) {
	for _, item := range v {
		switch item := item.(type) {
		case ExternalIdent:
			fmt.Fprint(f.buf, f.wrap(&item))
		default:
			fmt.Fprint(f.buf, item)
		}
	}
	fmt.Fprintln(f.buf)
}

// wrap .
func (f *GeneratedFile) wrap(exti *ExternalIdent) string {
	// if f.packagepath == exti.Package.Path {
	// 	return exti.Name
	// }

	// 外部包是否被引用
	if _, find := f.externalPackages[exti.Package.Path]; !find {
		// 包名是否使用
		if _, used := f.usedPackageNames[exti.Package.alias]; used {
			var suffix = 1
			for f.usedPackageNames[exti.Package.alias+strconv.Itoa(suffix)] {
				suffix++
			}
			exti.Package.alias += strconv.Itoa(suffix)
		}
		f.usedPackageNames[exti.Package.alias] = true
		f.externalPackages[exti.Package.Path] = exti.Package
	}

	return exti.string()
}

// content .
func (f *GeneratedFile) content() ([]byte, error) {
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
		return nil, fmt.Errorf("%v: unparsable Go source: %v\n%v", f.filename, err, src.String())
	}

	var imports = make([][2]string, 0, len(f.externalPackages))
	var isStandard = make(map[string]bool, len(f.externalPackages))
	for _, externalPackage := range f.externalPackages {
		isStandard[externalPackage.Path] = IsStandard(externalPackage.Path)
		imports = append(imports, [2]string{externalPackage.alias, externalPackage.Path})
	}

	sort.Slice(imports, func(i, j int) bool {
		var _i, _j = isStandard[imports[i][1]], isStandard[imports[j][1]]
		if _i && !_j {
			return true
		}
		if _j && !_i {
			return false
		}
		return imports[i][1] < imports[j][1]
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
		for _, item := range imports {
			decl.Specs = append(decl.Specs, &ast.ImportSpec{
				Name: &ast.Ident{
					Name:    item[0],
					NamePos: pos,
				},
				Path: &ast.BasicLit{
					Kind:     token.STRING,
					Value:    strconv.Quote(item[1]),
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

// marshal .
func (plugin *Plugin) marshal() error {
	for _, f := range plugin.files {
		content, err := f.content()
		if err != nil {
			return err
		}

		if err := os.MkdirAll(f.packagepath, 0755); err != nil {
			return err
		}

		file, err := os.OpenFile(path.Join(f.packagepath, f.filename), os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
		if err != nil {
			return err
		}
		file.Write(content)
		file.Close()
	}
	return nil
}

func abs(p string) string {
	p, _ = filepath.Abs(p)
	return p
}
