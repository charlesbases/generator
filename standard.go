package generator

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var standardPackages = make(map[string]struct{}, 0)

func init() {
	fi, _ := os.ReadDir(filepath.Join(runtime.GOROOT(), "src"))
	for _, item := range fi {
		if item.IsDir() {
			standardPackages[item.Name()] = struct{}{}
		}
	}
}

// IsStandard 是否为标准库包
func IsStandard(p string) bool {
	if _, found := standardPackages[p]; found {
		return true
	}
	for inter := range standardPackages {
		if strings.HasPrefix(p, inter+"/") {
			return true
		}
	}
	return false
}
