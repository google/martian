package recording

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

type recording struct {
	path string
}

func New(path string) *recording {
	return &recording{
		path: path,
	}
}

func (r *recording) Start() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (r *recording) Open() error {
	if _, err := os.Stat(r.path); os.IsNotExist(err) {
		r.initialize()
		return nil
	}

	zr, err := zip.OpenReader(r.path)
	if err != nil {
		return err
	}

	tmp, err := ioutil.TempDir("", "martian.replay.")
	if err != nil {
		return err
	}

	for _, f := range zr.File {
		p := filepath.Join(tmp, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(p, f.Mode())
			continue
		}

		tgtf, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		defer tgtf.Close()
		if err != nil {
			return err
		}

		srcf, err := f.Open()
		defer srcf.Close()
		if err != nil {
			return err
		}
		_, err = io.Copy(tgtf, srcf)
		if err != nil {
			return err
		}
	}
}

func (r *recording) initialize() {

}
