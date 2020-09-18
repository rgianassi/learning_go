package main

import (
	"fmt"
	"log"
	"net/http"
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

func testNonExistentHash() {
	client := &http.Client{}

	fmt.Println("Test non existent hash")

	response, err := client.Get("http://localhost:9090/1234567")

	if err != nil {
		fmt.Println("Got error on get for non existent hash:", err)
		os.Exit(1)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusNotFound {
		fmt.Println("Expected not found response on non existent hash, but got:", response.StatusCode)
		os.Exit(1)
	}
}

func testDevelerHash() {
	client := &http.Client{}

	fmt.Println("Test develer hash")

	response, err := client.Get("http://localhost:9090/ac60366")

	if err != nil {
		fmt.Println("Got error on get for develer hash:", err)
		os.Exit(1)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		fmt.Println("Expected ok response on existent hash, but got:", response.StatusCode)
		os.Exit(1)
	}

	url, err := response.Location()

	if err != nil {
		fmt.Println("Got error on develer location:", err)
		os.Exit(1)
	}

	if url.Path != "http://www.develer.com" {
		fmt.Println("Expected redirect to http://www.develer.com, but got:", url.Path)
		os.Exit(1)
	}
}

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

	time.Sleep(1 * time.Second)

	defer func() {
		fmt.Println("Sending kill")
		cmd.Process.Signal(os.Kill)
		fmt.Println("Kill sent")
		time.Sleep(1 * time.Second)
	}()

	fmt.Println("Starting test...")

	testNonExistentHash()
	testDevelerHash()

	fmt.Println("Exit")
	os.Exit(0)
}
