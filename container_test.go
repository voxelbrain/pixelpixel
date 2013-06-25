package main

import (
	"archive/tar"
	"bytes"
	"io"
	"math"
	"path/filepath"
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
	id, err := lcm.NewContainer(fs)
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
	id, err := lcm.NewContainer(fs)
	if err != nil {
		t.Fatalf("Could not start container: %s", err)
	}
	defer lcm.DestroyContainer(id)

	<-lcm.WaitFor(id)
	output := lcm.containers[id].Logs.String()
	if output != "Hello World" {
		t.Fatalf("Unexpected output of container: %s", output)
	}
}

func TestWaitFor(t *testing.T) {
	buf := makeFs(map[string]interface{}{
		"main.go": `package main

		import (
			"time"
		)

		func main() {
			time.Sleep(500 * time.Millisecond)
		}`,
	})
	fs := tar.NewReader(bytes.NewReader(buf))

	lcm := NewLocalContainerManager()
	start := time.Now()
	id, err := lcm.NewContainer(fs)
	if err != nil {
		t.Fatalf("Could not start container: %s", err)
	}
	defer lcm.DestroyContainer(id)

	select {
	case <-lcm.WaitFor(id):
	case <-time.After(600 * time.Millisecond):
	}

	length := time.Now().Sub(start)
	if math.Abs(float64(length-500*time.Millisecond)) > float64(10*time.Millisecond) {
		t.Fatalf("Unexpected amount of time until death")
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
