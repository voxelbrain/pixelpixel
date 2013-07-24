package main

import (
	"archive/tar"
	"bytes"
	"io"
	"math/rand"
	"path/filepath"
	"time"
)

const (
	chars = `abcdefghijklmnopqrstuvwxyz1234567890`
)

func GenerateAlnumString(length int) string {
	key := make([]byte, length)
	idx := rand.Perm(len(chars))
	for i := 0; i < length; i++ {
		key[i] = chars[idx[i]]
	}
	return string(key)
}

func StopContainer(ctr Container) {
	ctr.SoftKill()
	timer := time.AfterFunc(2*time.Second, func() {
		ctr.HardKill()
	})
	ctr.Wait()
	timer.Stop()
}

func makeFs(fs map[string]interface{}) []byte {
	buf := &bytes.Buffer{}
	w := tar.NewWriter(buf)
	makeFsPrefix(w, "", fs)
	w.Close()
	return buf.Bytes()
}

func makeFsPrefix(w *tar.Writer, prefix string, fs map[string]interface{}) {
	for item, content := range fs {
		path := filepath.Join(prefix, item)
		switch x := content.(type) {
		case string:
			w.WriteHeader(&tar.Header{
				Name:     path,
				Typeflag: tar.TypeReg,
				Size:     int64(len([]byte(x))),
			})
			io.WriteString(w, x)
		case map[string]interface{}:
			w.WriteHeader(&tar.Header{
				Name:     path,
				Typeflag: tar.TypeDir,
			})
			makeFsPrefix(w, path, x)
		default:
			panic("Invalid item type in Fs declaration")
		}
	}
}
