package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

/*

Write another program that performs an end to end test of the URL shortener:

it starts an (already compiled) binary of the the url shortener server,
in a sub process, with an already existing url file --load.

using an HTTP client, it performs some HTTP requests of the different handlers

it verifies their responses

either exits 0 if all went well or exit 1 and write what went wrong on std out

Scenario:

call / on an non-existing SHA (check the http.StatusCode)

call / on a existing SHA

call /statistics, unmarshall the json and checks it

*/

func main() {
	fmt.Println("Starting server...")
	cmd := exec.Command("build/url_shortener/http_server", "--load", "build/url_shortener/persistence.json")
	err := cmd.Start()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Server started")
	go func() {
		fmt.Println("Waiting server...")
		err := cmd.Wait()
		log.Println("Command finished with error:", err)
	}()
	time.Sleep(60 * time.Second)
	fmt.Println("Sending kill")
	cmd.Process.Signal(os.Kill)
	fmt.Println("Kill sent")
}
