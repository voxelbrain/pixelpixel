package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/surma-dump/gouuid"
)

type LocalContainerManager struct {
	Root       string
	m          *sync.Mutex
	containers map[ContainerId]*localContainer
}

type localContainer struct {
	Id   ContainerId
	Root string
	Cmd  *exec.Cmd
	Done chan bool
	Logs *bytes.Buffer
}

func NewLocalContainerManager() *LocalContainerManager {
	r := &LocalContainerManager{
		m:          &sync.Mutex{},
		containers: map[ContainerId]*localContainer{},
		Root:       filepath.Join(os.TempDir(), "pixelpixel"),
	}
	return r
}

func (lcm *LocalContainerManager) NewContainer(fs *tar.Reader, envInjections []string) (ContainerId, error) {
	id := ContainerId(gouuid.New().String())
	dir := filepath.Join(lcm.Root, id.String())
	err := os.MkdirAll(dir, os.FileMode(0755))
	if err != nil {
		return id, err
	}

	ctr := &localContainer{
		Id:   id,
		Root: dir,
		Logs: &bytes.Buffer{},
		Done: make(chan bool),
	}

	err = ctr.extractFileSystem(fs)
	if err != nil {
		return id, err
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return id, err
	}
	ctr.Cmd = exec.Command("go", stringList("run", files)...)
	ctr.Cmd.Dir = dir
	ctr.Cmd.Stdout = ctr.Logs
	ctr.Cmd.Stderr = ctr.Logs
	ctr.Cmd.Env = stringList(os.Environ(), envInjections, fmt.Sprintf("PORT=%d", <-portGenerator))
	err = ctr.Cmd.Start()
	if err != nil {
		return id, err
	}

	go func(c chan bool) {
		ctr.Cmd.Wait()
		close(c)
	}(ctr.Done)

	lcm.m.Lock()
	lcm.containers[id] = ctr
	lcm.m.Unlock()

	return id, nil
}

func (lcm *LocalContainerManager) DestroyContainer(id ContainerId) error {
	ctr, ok := lcm.containers[id]
	if !ok {
		return os.ErrNotExist
	}

	ctr.Cmd.Process.Signal(os.Interrupt)
	c := make(chan error)
	go func() {
		c <- ctr.Cmd.Wait()
	}()
	select {
	case <-c:
		// Nop
	case <-time.After(5 * time.Second):
		ctr.Cmd.Process.Signal(os.Kill)
	}

	lcm.m.Lock()
	delete(lcm.containers, id)
	lcm.m.Unlock()

	os.RemoveAll(filepath.Join(lcm.Root, ctr.Id.String()))
	return nil
}

func (lcm *LocalContainerManager) WaitFor(id ContainerId) chan bool {
	ctr, ok := lcm.containers[id]
	if !ok {
		return nil
	}
	return ctr.Done
}

func (lcm *LocalContainerManager) Logs(id ContainerId) ([]byte, error) {
	ctr, ok := lcm.containers[id]
	if !ok {
		return nil, os.ErrNotExist
	}
	return ctr.Logs.Bytes(), nil
}

func (lc *localContainer) extractFileSystem(fs *tar.Reader) error {
	hdr, err := fs.Next()
	for err == nil {
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
	go func() {
		c := make(chan int)
		portGenerator = c
		for i := 49000; i < 65535; i++ {
			c <- i
		}
	}()
}
