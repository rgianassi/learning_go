package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/rgianassi/learning/go/src/url_shortener/shorten"
)

func testNonExistentHash() error {
	client := &http.Client{}

	fmt.Println("Test non existent hash")

	response, err := client.Get("http://localhost:9090/1234567")

	if err != nil {
		fmt.Println("Got error on get for non existent hash:", err)
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusNotFound {
		fmt.Println("Expected not found response on non existent hash, but got:", response.StatusCode)
		return err
	}

	return nil
}

func testWeatherHash() error {
	client := &http.Client{}

	fmt.Println("Test weather hash")

	const weatherURL = "https://wttr.in/Florence"

	response, err := client.Get("http://localhost:9090/f495791")

	if err != nil {
		fmt.Println("Got error on get for weather hash:", err)
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		fmt.Println("Expected ok response on existent hash, but got:", response.StatusCode)
		return err
	}

	requestURL := response.Request.URL
	url := requestURL.String()

	if url != weatherURL {
		fmt.Println("Expected redirect to", weatherURL, ", but got:", url)
		return err
	}

	return nil
}

func testStatisticsJSON() error {
	client := &http.Client{}

	fmt.Println("Test statistics JSON")

	response, err := client.Get("http://localhost:9090/statistics?format=json")

	if err != nil {
		fmt.Println("Got error on get for statistics JSON:", err)
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		fmt.Println("Expected ok response on statistics JSON, but got:", response.StatusCode)
		return err
	}

	decoder := json.NewDecoder(response.Body)

	var stats shorten.StatsJSON

	if err := decoder.Decode(&stats); err != nil {
		fmt.Println("Unable to decode statistics JSON:", err)
		return err
	}

	gotTotalURL := stats.ServerStats.TotalURL
	if gotTotalURL != 1 {
		fmt.Println("Expected TotalURL on statistics JSON to be 1, but got:", gotTotalURL)
		return err
	}

	return nil
}

func trueMain() int {
	fmt.Println("Starting server...")

	executable := "build/url_shortener/http_server"
	loadFlag := "--load"
	loadParameter := "build/url_shortener/persistence.json"

	cmd := exec.Command(executable, loadFlag, loadParameter)
	err := cmd.Start()

	if err != nil {
		log.Println(err)
		return 1
	}

	fmt.Println("Server started")

	go func() {
		fmt.Println("Waiting server...")
		err := cmd.Wait()
		log.Println("Server finished with error:", err)
	}()

	time.Sleep(1 * time.Second) // to let the goroutine go

	defer func() {
		fmt.Println("Sending kill")

		cmd.Process.Signal(os.Kill)

		fmt.Println("Kill sent")

		time.Sleep(1 * time.Second) // to let the kill kill
	}()

	fmt.Println("Starting test...")

	if err := testNonExistentHash(); err != nil {
		return 1
	}

	if err := testWeatherHash(); err != nil {
		return 1
	}

	if err := testStatisticsJSON(); err != nil {
		return 1
	}

	fmt.Println("Exit")

	return 0
}

func main() {
	exitCode := trueMain()

	os.Exit(exitCode)
}
