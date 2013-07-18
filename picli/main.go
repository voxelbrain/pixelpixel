package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/voxelbrain/goptions"
)

var (
	options = struct {
		Server string        `goptions:"-s, --server, description='Pixelpixel server to push to'"`
		Help   goptions.Help `goptions:"-h, --help, description='Show this help'"`
		goptions.Verbs
		Upload struct{} `goptions:"upload"`
		Logs   struct{} `goptions:"logs"`
	}{
		Server: "localhost:8080",
	}
)

const (
	keyFile = ".key"
)

func main() {
	fs := goptions.NewFlagSet("picli", &options)
	err := fs.Parse(os.Args[1:])
	key := loadKey()

	switch options.Verbs {
	case "upload":
		upload(key)
	case "logs":
	default:
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
		fs.PrintHelp(os.Stderr)
		return
	}

}

func makeApiCall(method, subpath string, body io.Reader) (int, string, error) {
	url := path.Join(options.Server, "pixels") + "/"
	url = slashify(url + subpath)
	log.Printf("URL: %s", url)
	req, _ := http.NewRequest(method, url, body)
	resp, err := http.DefaultClient.Do(req)
	buf := &bytes.Buffer{}

	code := -1
	if resp != nil {
		defer resp.Body.Close()
		code = resp.StatusCode
		io.Copy(buf, resp.Body)
	}

	// TODO: Sadly, this seems to be necassary. Can't seem to get it
	// working without it. I always get `malformed HTTP response ""`
	http.DefaultTransport.(*http.Transport).CloseIdleConnections()

	return code, buf.String(), err
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
	return strings.TrimSpace(string(key))
}

func slashify(url string) string {
	url = path.Clean(url)
	if !strings.HasSuffix(url, "/") {
		url = url + "/"
	}
	return "http://" + strings.TrimPrefix(url, "/")
}
