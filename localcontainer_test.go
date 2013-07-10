package main

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestContainerLogs(t *testing.T) {
	buf := makeFs(map[string]interface{}{
		"main.go": `package main

		import (
			"fmt"
			"os"
		)

		func main() {
			fmt.Printf("Hello World")
			fmt.Fprintf(os.Stderr, "Hello Error")
		}`,
	})
	fs := tar.NewReader(bytes.NewReader(buf))

	lcm := NewLocalContainerManager()
	id, err := lcm.NewContainer(fs, nil)
	if err != nil {
		t.Fatalf("Could not start container: %s", err)
	}
	defer lcm.DestroyContainer(id, true)

	<-lcm.WaitFor(id)
	output := lcm.containers[id].Logs.String()
	if output != "Hello WorldHello Error" {
		t.Fatalf("Unexpected output of container: %s", output)
	}
}

func TestSubfolderHandling(t *testing.T) {
	buf := makeFs(map[string]interface{}{
		"main.go": `package main

		import (
			"fmt"
			"./subpackage"
		)

		func main() {
			fmt.Printf("%s", subpackage.TheConstant)
		}`,
		"subpackage": map[string]interface{}{
			"subpackage.go": `package subpackage

			var (
				TheConstant = "Hello World"
			)`,
		},
	})
	fs := tar.NewReader(bytes.NewReader(buf))

	lcm := NewLocalContainerManager()
	id, err := lcm.NewContainer(fs, nil)
	if err != nil {
		t.Fatalf("Could not start container: %s", err)
	}
	defer lcm.DestroyContainer(id, true)

	<-lcm.WaitFor(id)
	output, err := lcm.Logs(id)
	if err != nil {
		t.Fatalf("Could not get logs: %s", err)
	}
	if string(output) != "Hello World" {
		t.Fatalf("Unexpected output of container: %s", output)
	}
}

func TestWaitFor(t *testing.T) {
	buf := makeFs(map[string]interface{}{
		"main.go": `package main

		func main() {
		}`,
	})
	fs := tar.NewReader(bytes.NewReader(buf))

	lcm := NewLocalContainerManager()
	id, err := lcm.NewContainer(fs, nil)
	if err != nil {
		t.Fatalf("Could not start container: %s", err)
	}
	defer lcm.DestroyContainer(id, true)

	select {
	case <-lcm.WaitFor(id):
		// Nop
	case <-time.After(1 * time.Second):
		t.Fatalf("Timeout occured")
	}
}

func TestKilling(t *testing.T) {
	buf := makeFs(map[string]interface{}{
		"main.go": `package main

		import (
			"time"
			"os"
			"os/signal"
			"fmt"
		)

		func main() {
			fmt.Printf("hai")
			c := make(chan os.Signal)
			signal.Notify(c, os.Interrupt)
			go func() {
				for {
					<-c
				}
			}()
			time.Sleep(10 * time.Second)
		}`,
	})
	fs := tar.NewReader(bytes.NewReader(buf))

	lcm := NewLocalContainerManager()
	id, err := lcm.NewContainer(fs, nil)
	if err != nil {
		t.Fatalf("Could not start container: %s", err)
	}
	done := lcm.WaitFor(id)
	lcm.DestroyContainer(id, true)

	select {
	case <-done:
		// Nop
	case <-time.After(3 * time.Second):
		t.Fatalf("Timeout occured")
	}
}

func TestPurge(t *testing.T) {
	buf := makeFs(map[string]interface{}{
		"main.go": `package main

		func main() {
		}`,
	})
	fs := tar.NewReader(bytes.NewReader(buf))

	lcm := NewLocalContainerManager()
	id, err := lcm.NewContainer(fs, nil)
	if err != nil {
		t.Fatalf("Could not start container: %s", err)
	}
	path := lcm.containers[id].Root
	destroyed := lcm.containers[id].destroyed
	lcm.DestroyContainer(id, true)

	select {
	case <-destroyed:
		// Nop
	case <-time.After(1 * time.Second):
		t.Fatalf("Timeout occured")
	}

	if info, err := os.Stat(path); err == nil && info.IsDir() {
		t.Fatalf("Folder %s was not purged", path)
	}

	fs = tar.NewReader(bytes.NewReader(buf))
	id, err = lcm.NewContainer(fs, nil)
	if err != nil {
		t.Fatalf("Could not start container: %s", err)
	}
	path = lcm.containers[id].Root
	destroyed = lcm.containers[id].destroyed
	lcm.DestroyContainer(id, false)

	select {
	case <-destroyed:
		// Nop
	case <-time.After(1 * time.Second):
		t.Fatalf("Timeout occured")
	}

	if info, err := os.Stat(path); err != nil || !info.IsDir() {
		t.Fatalf("Folder %s was purged", path)
	}

	os.RemoveAll(path)
}

func TestTwoSequentialContainers(t *testing.T) {
	buf := makeFs(map[string]interface{}{
		"main.go": `package main

		func main() {
		}`,
	})

	fs1 := tar.NewReader(bytes.NewReader(buf))
	fs2 := tar.NewReader(bytes.NewReader(buf))
	lcm := NewLocalContainerManager()

	id1, err := lcm.NewContainer(fs1, nil)
	if err != nil {
		t.Fatalf("Could not start container: %s", err)
	}
	lcm.DestroyContainer(id1, true)

	id2, err := lcm.NewContainer(fs2, nil)
	if err != nil {
		t.Fatalf("Could not start container: %s", err)
	}
	lcm.DestroyContainer(id2, true)

	count := 0
	for {
		select {
		case <-lcm.WaitFor(id1):
			count++
		case <-lcm.WaitFor(id2):
			count++
		case <-time.After(1 * time.Second):
			t.Fatalf("Timeout occured")
		}
		if count == 2 {
			return
		}
	}
}

func TestInvalidTar(t *testing.T) {
	data := bytes.NewReader([]byte(`This is obivously not a valid tar`))
	fs := tar.NewReader(data)

	lcm := NewLocalContainerManager()
	_, err := lcm.NewContainer(fs, nil)
	if err == nil {
		t.Fatalf("Corrupted tar file was accepted")
	}
}

func TestPortAssignment(t *testing.T) {
	buf := makeFs(map[string]interface{}{
		"main.go": `package main

		import (
			"fmt"
			"os"
		)

		func main() {
			fmt.Printf("localhost:%s", os.Getenv("PORT"))
		}`,
	})
	fs := tar.NewReader(bytes.NewReader(buf))

	lcm := NewLocalContainerManager()
	id1, err := lcm.NewContainer(fs, nil)
	if err != nil {
		t.Fatalf("Could not start container: %s", err)
	}
	defer lcm.DestroyContainer(id1, true)

	select {
	case <-lcm.WaitFor(id1):
		// Nop
	case <-time.After(1 * time.Second):
		t.Fatalf("Timeout occured")
	}

	fs = tar.NewReader(bytes.NewReader(buf))
	id2, err := lcm.NewContainer(fs, nil)
	if err != nil {
		t.Fatalf("Could not start container: %s", err)
	}
	defer lcm.DestroyContainer(id2, true)

	select {
	case <-lcm.WaitFor(id2):
		// Nop
	case <-time.After(1 * time.Second):
		t.Fatalf("Timeout occured")
	}

	logs1, err := lcm.Logs(id1)
	if err != nil {
		t.Fatalf("Could not get logs: %s", err)
	}
	port1, err := lcm.SocketAddress(id1)
	if err != nil {
		t.Fatalf("Could not get port: %s", err)
	}
	if string(logs1) != port1 {
		t.Fatalf("Specified and injected ports differ. Injected %d, got %s", port1, logs1)
	}

	logs2, err := lcm.Logs(id2)
	if err != nil {
		t.Fatalf("Could not get logs: %s", err)
	}
	port2, err := lcm.SocketAddress(id2)
	if err != nil {
		t.Fatalf("Could not get port: %s", err)
	}
	if string(logs2) != port2 {
		t.Fatalf("Specified and injected ports differ. Injected %d, got %s", port2, logs2)
	}

	if reflect.DeepEqual(logs1, logs2) {
		t.Fatalf("Same port was assigned")
	}
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
