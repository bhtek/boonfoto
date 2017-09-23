package filescanner

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ScannerVisitorFunc func(path string, modTime time.Time)

type filescanner struct {
	fsvisit ScannerVisitorFunc
}

func (fs filescanner) visit(path string, f os.FileInfo, err error) error {
	if !f.IsDir() && strings.HasSuffix(f.Name(), ".jpg") {
		fs.fsvisit(path, f.ModTime())
	}

	return nil
}

func Scan(path string, svc ScannerVisitorFunc) {
	fs := filescanner{svc}
	filepath.Walk(path, fs.visit)
}
