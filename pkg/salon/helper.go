package salon

import (
	"io/fs"
	"path/filepath"
	"strings"
)

func findAllFiles(fsys fs.FS, dir, suffix string) ([]string, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading directory %s", dir)
	}
	var result = []string{}
	for _, entry := range entries {
		name := filepath.ToSlash(filepath.Join(dir, entry.Name()))
		if entry.IsDir() {
			entries2, err := findAllFiles(fsys, name, suffix)
			if err != nil {
				return nil, err
			}
			result = append(result, entries2...)
		} else {
			if strings.HasSuffix(entry.Name(), suffix) {
				result = append(result, name)
			}
		}
	}
	return result, nil
}
