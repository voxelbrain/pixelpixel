package main

import (
	"fmt"
	"log"
)

func logs(key string) {
	fmt.Printf("Logs for \"%s\":\n", key)
	_, body, err := makeApiCall("GET", "/"+key+"/logs", nil)
	if err != nil {
		log.Fatalf("Could not reqeust logs: %s", err)
	}

	fmt.Printf("%s", body)
}
