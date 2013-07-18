package main

import (
	"fmt"
	"log"
)

func logs(key string) {
	log.Printf("Logs for \"%s\":", key)
	_, body, err := makeApiCall("GET", "/"+key+"/logs", nil)
	if err != nil {
		log.Fatalf("Could not reqeust logs: %s", err)
	}

	fmt.Printf("%s", body)
}
