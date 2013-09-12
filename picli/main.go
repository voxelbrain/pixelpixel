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
	"time"

	"github.com/voxelbrain/goptions"
)

var (
	options = struct {
		Server string        `goptions:"-s, --server, description='Pixelpixel server to push to'"`
		Help   goptions.Help `goptions:"-h, --help, description='Show this help'"`
		goptions.Verbs
		Upload struct{} `goptions:"upload"`
		Logs   struct{} `goptions:"logs"`
		Format struct{} `goptions:"format"`
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
	key := prepareKey()

	switch options.Verbs {
	case "upload":
		options.Server = validateServer(options.Server)
		upload(key)
	case "logs":
		options.Server = validateServer(options.Server)
		logs(key)
	case "format":
		format()
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

func prepareKey() string {
	f, err := os.Open(keyFile)
	if err != nil {
		return ""
	}
	defer f.Close()
	key, err := ioutil.ReadAll(f)
	if err != nil {
		return ""
	}
	skey := strings.TrimSpace(string(key))
	code, _, err := makeApiCall("GET", "/"+skey, nil)
	if err == nil && code == http.StatusOK {
		return skey
	}
	return ""
}

func slashify(url string) string {
	url = path.Clean(url)
	if !strings.HasSuffix(url, "/") {
		url = url + "/"
	}
	return "http://" + strings.TrimPrefix(url, "/")
}

const (
	DEFAULT_REMOTE_SERVER = "pixelpixel.haxigon.com"
)

func validateServer(server string) string {
	t := time.AfterFunc(3*time.Second, func() {
		log.Fatalf("Attempt to connect to server timed out")
	})
	defer func() {
		t.Stop()
	}()

	resp, err := http.Get("http://" + path.Join(server, "handshake"))
	if err == nil && readAll(resp.Body) == "PIXELPIXEL OK" {
		return server
	}

	resp, err = http.Get("http://" + path.Join(DEFAULT_REMOTE_SERVER, "handshake"))
	if err == nil && readAll(resp.Body) == "PIXELPIXEL OK" {
		return DEFAULT_REMOTE_SERVER
	}
	log.Fatalf("Could not connect to server")
	return ""
}

func readAll(r io.Reader) string {
	data, _ := ioutil.ReadAll(r)
	return string(data)
}
