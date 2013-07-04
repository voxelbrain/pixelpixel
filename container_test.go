package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
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
	defer lcm.DestroyContainer(id)

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
	defer lcm.DestroyContainer(id)

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
	defer lcm.DestroyContainer(id)

	select {
	case <-lcm.WaitFor(id):
		// Nop
	case <-time.After(1 * time.Second):
		t.Fatalf("Timeout occured")
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
			fmt.Printf("%s", os.Getenv("PORT"))
		}`,
	})
	fs := tar.NewReader(bytes.NewReader(buf))

	lcm := NewLocalContainerManager()
	id1, err := lcm.NewContainer(fs, nil)
	if err != nil {
		t.Fatalf("Could not start container: %s", err)
	}
	defer lcm.DestroyContainer(id1)

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
	defer lcm.DestroyContainer(id2)

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
	port1, err := lcm.Port(id1)
	if err != nil {
		t.Fatalf("Could not get port: %s", err)
	}
	if string(logs1) != fmt.Sprintf("%d", port1) {
		t.Fatalf("Specified and injected ports differ. Injected %d, got %s", port1, logs1)
	}

	logs2, err := lcm.Logs(id2)
	if err != nil {
		t.Fatalf("Could not get logs: %s", err)
	}
	port2, err := lcm.Port(id2)
	if err != nil {
		t.Fatalf("Could not get port: %s", err)
	}
	if string(logs2) != fmt.Sprintf("%d", port2) {
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
