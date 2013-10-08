package main

import (
	"archive/tar"
	"bytes"
	"os"
	"sync"
	"testing"
	"time"
)

var (
	lcc = NewLocalContainerCreator()
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

	ctr, err := lcc.CreateContainer(fs, nil)
	if err != nil {
		t.Fatalf("Could not start container: %s", err)
	}
	defer ctr.Cleanup()

	ctr.Wait()
	output := ctr.Logs()
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

	ctr, err := lcc.CreateContainer(fs, nil)
	if err != nil {
		t.Fatalf("Could not start container: %s", err)
	}
	defer ctr.Cleanup()

	ctr.Wait()
	output := ctr.Logs()
	if err != nil {
		t.Fatalf("Could not get logs: %s", err)
	}
	if output != "Hello World" {
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

	ctr, err := lcc.CreateContainer(fs, nil)
	if err != nil {
		t.Fatalf("Could not start container: %s", err)
	}
	defer ctr.Cleanup()

	timer := time.AfterFunc(1*time.Second, func() {
		t.Fatalf("Termination timeout")
	})
	ctr.Wait()
	timer.Stop()
}

func TestPurge(t *testing.T) {
	buf := makeFs(map[string]interface{}{
		"main.go": `package main

		func main() {
		}`,
	})
	fs := tar.NewReader(bytes.NewReader(buf))

	ctr, err := lcc.CreateContainer(fs, nil)
	if err != nil {
		t.Fatalf("Could not start container: %s", err)
	}

	timer := time.AfterFunc(1*time.Second, func() {
		t.Fatalf("Timeout occured")
	})
	ctr.Cleanup()
	timer.Stop()

	if info, err := os.Stat(ctr.(*localContainer).Root); err == nil && info.IsDir() {
		t.Fatalf("Folder %s was not purged", ctr.(*localContainer).Root)
	}
}

func TestTwoSequentialContainers(t *testing.T) {
	buf := makeFs(map[string]interface{}{
		"main.go": `package main

		func main() {
		}`,
	})

	fs1 := tar.NewReader(bytes.NewReader(buf))
	fs2 := tar.NewReader(bytes.NewReader(buf))

	ctr1, err := lcc.CreateContainer(fs1, nil)
	if err != nil {
		t.Fatalf("Could not start container: %s", err)
	}
	go ctr1.Cleanup()

	ctr2, err := lcc.CreateContainer(fs2, nil)
	if err != nil {
		t.Fatalf("Could not start container: %s", err)
	}
	go ctr2.Cleanup()

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		ctr1.Wait()
		wg.Done()
	}()
	go func() {
		ctr2.Wait()
		wg.Done()
	}()
	timer := time.AfterFunc(1*time.Second, func() {
		t.Fatalf("Timeout occured")
	})
	wg.Wait()
	timer.Stop()
}

func TestInvalidTar(t *testing.T) {
	data := bytes.NewReader([]byte(`This is obivously not a valid tar`))
	fs := tar.NewReader(data)

	_, err := lcc.CreateContainer(fs, nil)
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
			fmt.Printf("127.0.0.1:%s", os.Getenv("PORT"))
		}`,
	})
	fs := tar.NewReader(bytes.NewReader(buf))

	ctr1, err := lcc.CreateContainer(fs, nil)
	if err != nil {
		t.Fatalf("Could not start container: %s", err)
	}
	go ctr1.Cleanup()

	timer := time.AfterFunc(1*time.Second, func() {
		t.Fatalf("Timeout occured")
	})
	ctr1.Wait()
	timer.Stop()

	fs = tar.NewReader(bytes.NewReader(buf))
	ctr2, err := lcc.CreateContainer(fs, nil)
	if err != nil {
		t.Fatalf("Could not start container: %s", err)
	}
	go ctr2.Cleanup()

	timer = time.AfterFunc(1*time.Second, func() {
		t.Fatalf("Timeout occured")
	})
	ctr2.Wait()
	timer.Stop()

	if ctr1.Logs() != ctr1.Address().String() {
		t.Fatalf("Specified and injected ports differ. Injected %s, got %s", ctr1.Address(), ctr1.Logs())
	}
	if ctr2.Logs() != ctr2.Address().String() {
		t.Fatalf("Specified and injected ports differ. Injected %s, got %s", ctr2.Address(), ctr2.Logs())
	}
	if ctr1.Logs() == ctr2.Logs() {
		t.Fatalf("Same port was assigned")
	}
}

func TestIsRunning(t *testing.T) {
	buf := makeFs(map[string]interface{}{
		"main.go": `package main

		func main() {
			panic("CRASH")
		}`,
	})
	fs := tar.NewReader(bytes.NewReader(buf))

	ctr, err := lcc.CreateContainer(fs, nil)
	if err != nil {
		t.Fatalf("Could not start container: %s", err)
	}
	defer ctr.Cleanup()

	if !ctr.IsRunning() {
		t.Fatalf("Container has IsRunning() = false after start")
	}
	timer := time.AfterFunc(1*time.Second, func() {
		t.Fatalf("Termination timeout")
	})
	ctr.Wait()
	if ctr.IsRunning() {
		t.Fatalf("Container has IsRunning() = true after Wait()")
	}
	timer.Stop()
}

func TestInvalidCode(t *testing.T) {
	buf := makeFs(map[string]interface{}{
		"main.go": `package main

		func main() {
			asdfsdafsafwhatisthis?
		}`,
	})
	fs := tar.NewReader(bytes.NewReader(buf))

	ctr, err := lcc.CreateContainer(fs, nil)
	if err != nil {
		t.Fatalf("Could not start container: %s", err)
	}
	defer ctr.Cleanup()

	timer := time.AfterFunc(1*time.Second, func() {
		t.Fatalf("Termination timeout")
	})
	ctr.Wait()
	timer.Stop()
}
