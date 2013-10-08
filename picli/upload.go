package main

import (
	"archive/tar"
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func upload(key string) {
	fs, err := createFs(".")
	if err != nil {
		log.Fatalf("Could not create filesystem: %s", err)
	}

	var code int
	var body string
	if key != "" {
		code, body, err = makeApiCall("PUT", "/"+key, bytes.NewReader(fs))
	} else {
		log.Printf("Creating a new pixel")
		code, body, err = makeApiCall("POST", "/", bytes.NewReader(fs))
	}

	if code >= 300 || err != nil {
		log.Printf("%s", string(body))
		log.Fatalf("Could not upload new pixel. Status code: %d, Error: %s", code, err)
	}

	f, err := os.Create(keyFile)
	if err != nil {
		log.Fatalf("Could not save key file (key=%s): %s", body, err)
	}
	defer f.Close()
	io.WriteString(f, body)
	log.Printf("Pixel %s uploaded", strings.TrimSpace(body))
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
				Name:     path,
				Typeflag: tar.TypeDir,
			})
			return err
		}

		err := fs.WriteHeader(&tar.Header{
			Name:     path,
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
