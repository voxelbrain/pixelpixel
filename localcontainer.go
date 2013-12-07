package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type LocalContainerCreator struct {
	Root string
}

func NewLocalContainerCreator() *LocalContainerCreator {
	r := &LocalContainerCreator{
		Root: filepath.Join(os.TempDir(), "pixelpixel"),
	}
	return r
}

func (lcc *LocalContainerCreator) CreateContainer(fs *tar.Reader, envInjections []string) (Container, error) {
	dir := filepath.Join(lcc.Root, GenerateAlnumString(32))
	err := os.MkdirAll(dir, os.FileMode(0755))
	if err != nil {
		return nil, err
	}

	ctr := &localContainer{
		Root:        dir,
		LogBuffer:   &bytes.Buffer{},
		Terminating: make(chan bool),
	}

	purge := true
	defer func() {
		if purge {
			close(ctr.Terminating)
			ctr.Terminated = true
			os.RemoveAll(dir)
		}
	}()

	err = ctr.extractFileSystem(fs)
	if err != nil {
		return nil, err
	}

	err = ctr.compile()
	if err != nil {
		return ctr, nil
	}

	purge = false
	ctr.Cmd = exec.Command(filepath.Join(dir, "pixel"))
	ctr.Cmd.Dir = dir
	ctr.Cmd.Stdout = ctr.LogBuffer
	ctr.Cmd.Stderr = ctr.LogBuffer
	ctr.Port = <-portGenerator
	ctr.Cmd.Env = stringList(os.Environ(), envInjections, fmt.Sprintf("PORT=%d", ctr.Port))
	err = ctr.Cmd.Start()
	if err != nil {
		return ctr, err
	}

	go func() {
		ctr.Cmd.Wait()
		close(ctr.Terminating)
		ctr.Terminated = true
	}()

	return ctr, nil
}

type localContainer struct {
	Root        string
	Cmd         *exec.Cmd
	Terminating chan bool
	LogBuffer   *bytes.Buffer
	Port        int
	Terminated  bool
}

func (lc *localContainer) IsRunning() bool {
	return !lc.Terminated
}

func (lc *localContainer) Address() net.Addr {
	addr, _ := net.ResolveTCPAddr("tcp4", fmt.Sprintf("localhost:%d", lc.Port))
	return addr
}

func (lc *localContainer) Logs() string {
	return lc.LogBuffer.String()
}

func (lc *localContainer) SoftKill() {
	if !lc.IsRunning() || lc.Cmd == nil || lc.Cmd.Process == nil {
		return
	}
	lc.Cmd.Process.Signal(os.Interrupt)
}

func (lc *localContainer) HardKill() {
	if !lc.IsRunning() || lc.Cmd == nil || lc.Cmd.Process == nil {
		return
	}
	lc.Cmd.Process.Signal(os.Kill)
}

func (lc *localContainer) Wait() {
	select {
	case <-lc.Terminating:
	}
}

func (lc *localContainer) Cleanup() {
	lc.Wait()
	os.RemoveAll(lc.Root)
}

func (lc *localContainer) compile() error {
	cmd := exec.Command("go", "get", "-d")
	cmd.Dir = lc.Root
	cmd.Stdout = lc.LogBuffer
	cmd.Stderr = lc.LogBuffer
	if err := cmd.Run(); err != nil {
		return err
	}

	files, err := filepath.Glob(filepath.Join(lc.Root, "*.go"))
	if err != nil {
		return err
	}

	// add extension .exe to compiled program in windows environment
	filename := "pixel"
	if runtime.GOOS == "windows" {
		filename += ".exe"
	}

	cmd = exec.Command("go", stringList("build", "-o", filename, files)...)
	cmd.Dir = lc.Root
	cmd.Stdout = lc.LogBuffer
	cmd.Stderr = lc.LogBuffer
	return cmd.Run()
}

func (lc *localContainer) extractFileSystem(fs *tar.Reader) error {
	hdr, err := fs.Next()
	for err != io.EOF {
		if err != nil {
			return err
		}
		file := filepath.Join(lc.Root, hdr.Name)
		switch hdr.Typeflag {
		case tar.TypeDir:
			err := os.MkdirAll(file, os.FileMode(0755))
			if err != nil {
				return err
			}
		case tar.TypeReg:
			err := lc.writeFile(file, fs)
			if err != nil {
				return err
			}
		default:
			log.Printf("Encountered unknown file type 0x%02x, skipping", hdr.Typeflag)
		}

		hdr, err = fs.Next()
	}
	return nil
}

func (lc *localContainer) writeFile(file string, ff io.Reader) error {
	f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, os.FileMode(0644))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, ff)
	return err
}

// Stolen from $GOROOT/src/cmd/go/main.go
// stringList's arguments should be a sequence of string or []string values.
// stringList flattens them into a single []string.
func stringList(args ...interface{}) []string {
	var x []string
	for _, arg := range args {
		switch arg := arg.(type) {
		case []string:
			x = append(x, arg...)
		case string:
			x = append(x, arg)
		default:
			panic("stringList: invalid argument")
		}
	}
	return x
}

var (
	portGenerator <-chan int
)

func init() {
	c := make(chan int)
	portGenerator = c
	go func() {
		for {
			for i := 49000; i < 65535; i++ {
				c <- i
			}
		}
	}()
}
