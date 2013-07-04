package main

import (
	"archive/tar"
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/voxelbrain/goptions"
)

var (
	options = struct {
		Server string        `goptions:"-s, --server, description='Pixelpixel server to push to'"`
		Help   goptions.Help `goptions:"-h, --help, description='Show this help'"`
	}{
		Server: "localhost:8080",
	}
)

func main() {
	goptions.ParseAndFail(&options)

	fs, err := createFs(".")
	if err != nil {
		log.Fatalf("Could not create filesystem: %s", err)
	}

	resp, err := http.Post("http://"+options.Server+"/pixels/", "application/tar", bytes.NewReader(fs))
	if err != nil {
		log.Fatalf("Could not upload filesystem: %s", err)
	}
	defer resp.Body.Close()
	buf := &bytes.Buffer{}
	io.Copy(buf, resp.Body)

	if resp.StatusCode != 200 {
		log.Fatalf("Server did not accept filesystem: %s", buf.String())
	}
	log.Printf("Your ID: %s", buf.String())
}

func createFs(path string) ([]byte, error) {
	buf := &bytes.Buffer{}
	fs := tar.NewWriter(buf)
	err := filepath.Walk(path, func(path string, info os.FileInfo, _ error) error {
		if path == "." {
			return nil
		}
		log.Printf("Adding %s", path)
		if info.IsDir() {
			err := fs.WriteHeader(&tar.Header{
				Name:     info.Name(),
				Typeflag: tar.TypeDir,
			})
			return err
		}

		err := fs.WriteHeader(&tar.Header{
			Name:     info.Name(),
			Mode:     int64(info.Mode()),
			Size:     info.Size(),
			Typeflag: tar.TypeReg,
		})
		if err != nil {
			log.Printf("Could not start next file: %s", err)
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			log.Printf("Could not read %s: %s", path, err)
			return err
		}
		defer f.Close()
		io.Copy(fs, f)
		return nil
	})
	fs.Flush()
	fs.Close()
	return buf.Bytes(), err
}
