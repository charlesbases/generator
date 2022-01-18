package generator

import (
	"io/ioutil"
	"path"
	"runtime"
	"strings"
)

var standardPackages = make(map[string]bool, 0)

func init() {
	fi, _ := ioutil.ReadDir(path.Join(runtime.GOROOT(), "src"))
	for _, item := range fi {
		if item.IsDir() {
			standardPackages[item.Name()] = true
		}
	}
}

// IsStandard 是否为标准库包
func IsStandard(p string) bool {
	if standardPackages[p] {
		return true
	}
	for inter := range standardPackages {
		if strings.HasPrefix(p, inter+"/") {
			return true
		}
	}
	return false
}
