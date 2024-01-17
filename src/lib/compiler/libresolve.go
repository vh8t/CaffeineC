package compiler

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/vyPal/CaffeineC/lib/cache"
	"github.com/vyPal/CaffeineC/lib/project"
)

func ResolveImportPath(path string, pcache cache.PackageCache) (string, error) {
	if strings.HasPrefix(path, "./") || strings.HasPrefix(path, "/") || strings.HasPrefix(path, "../") {
		return path, nil
	} else {
		found, pkg, fp, err := pcache.ResolvePackage(path)
		if err != nil {
			return "", err
		}
		if found {
			conf, err := project.GetCfConf(pkg.Path)
			if err != nil {
				return "", err
			}
			if conf.SourceDir == "" {
				color.Yellow("Package %s doesn't have a configured source directory. Using src/", pkg.Name)
				conf.SourceDir = "src"
			}
			if !strings.HasSuffix(fp, ".cffc") {
				fp += ".cffc"
			}
			return filepath.Join(pkg.Path, conf.SourceDir, fp), nil
		} else {
			return fmt.Sprintf("./%s", path), nil
		}
	}
}