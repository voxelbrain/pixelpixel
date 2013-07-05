package main

import (
	"archive/tar"
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

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

const (
	keyFile = ".key"
)

func main() {
	goptions.ParseAndFail(&options)

	key := loadKey()

	fs, err := createFs(".")
	if err != nil {
		log.Fatalf("Could not create filesystem: %s", err)
	}

	var req *http.Request
	if key == "" {
		url := "http://" + path.Join(options.Server, "pixels") + "/"
		req, _ = http.NewRequest("POST", url, bytes.NewReader(fs))
	} else {
		url := "http://" + path.Join(options.Server, "pixels", key)
		req, _ = http.NewRequest("PUT", url, bytes.NewReader(fs))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Could not upload filesystem: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		log.Fatalf("An error occured! Status code: %d", resp.StatusCode)
	}

	buf := &bytes.Buffer{}
	io.Copy(buf, resp.Body)
	if key == "" {
		f, err := os.Create(keyFile)
		if err != nil {
			log.Fatalf("Could not save key file (key=%s): %s", buf.String(), err)
		}
		defer f.Close()
		f.Write(buf.Bytes())
	}
}

func createFs(path string) ([]byte, error) {
	buf := &bytes.Buffer{}
	fs := tar.NewWriter(buf)
	err := filepath.Walk(path, func(path string, info os.FileInfo, _ error) error {
		if strings.HasPrefix(path, ".") {
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

func loadKey() string {
	f, err := os.Open(keyFile)
	if err != nil {
		return ""
	}
	defer f.Close()
	key, err := ioutil.ReadAll(f)
	if err != nil {
		return ""
	}
	return string(key)
}
