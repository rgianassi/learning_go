package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/rgianassi/learning/go/src/url_shortener/shorten"
)

const (
	exitCodeOk    = 0
	exitCodeError = 1

	loadFlag = "--load"

	oneSecond = 1 * time.Second

	sourcePersistenceFile = "src/url_shortener/cmd/end_to_end_tester/persistence.json"
	persistenceFile       = "build/url_shortener/persistence.json"

	serverExecutable = "build/url_shortener/http_server"

	weatherURLRome              = "https://wttr.in/Rome"
	shortWeatherURLRome         = "87aefef"
	shortenWeatherURLRome       = "http://localhost:9090/shorten?url=https://wttr.in/Rome"
	weatherURLFlorence          = "https://wttr.in/Florence"
	fullShortWeatherURLFlorence = "http://localhost:9090/f495791"
	fullShortNonExistentURL     = "http://localhost:9090/1234567"

	redirectURL = "http://localhost:9090/87aefef"

	statisticsURL = "http://localhost:9090/statistics?format=json"

	shortURLRegExp = `">(.*) ->`
)

// https://stackoverflow.com/a/21061062
func copyFromFileToFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return out.Close()
}

func testNonExistentHash() error {
	client := &http.Client{}

	response, err := client.Get(fullShortNonExistentURL)

	if err != nil {
		return fmt.Errorf("testNonExistentHash: got error on get: %v", err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusNotFound {
		return fmt.Errorf("testNonExistentHash: unexpected status code: %v", response.StatusCode)
	}

	return nil
}

func testWeatherHash() error {
	client := &http.Client{}

	response, err := client.Get(fullShortWeatherURLFlorence)

	if err != nil {
		return fmt.Errorf("testWeatherHash: got error on get: %v", err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("testWeatherHash: unexpected status code: %v", response.StatusCode)
	}

	requestURL := response.Request.URL
	url := requestURL.String()

	if url != weatherURLFlorence {
		return fmt.Errorf("testWeatherHash: unexpected URL, wanted: %v, got: %v", weatherURLFlorence, url)
	}

	return nil
}

func testStatisticsJSON() error {
	client := &http.Client{}

	response, err := client.Get(statisticsURL)

	if err != nil {
		return fmt.Errorf("testStatisticsJSON: got error on get: %v", err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("testStatisticsJSON: unexpected status code: %v", response.StatusCode)
	}

	decoder := json.NewDecoder(response.Body)

	var stats shorten.StatsJSON

	if err := decoder.Decode(&stats); err != nil {
		return fmt.Errorf("testStatisticsJSON: unable to decode, error: %v", err)
	}

	gotTotalURL := stats.ServerStats.TotalURL
	if gotTotalURL != 1 {
		return fmt.Errorf("testStatisticsJSON: expected TotalURL on statistics JSON to be 1, but got: %v", gotTotalURL)
	}

	return nil
}

func testShortenURLAddingANonExistentURL() error {
	client := &http.Client{}

	response, err := client.Get(shortenWeatherURLRome)

	if err != nil {
		return fmt.Errorf("testShortenURLAddingANonExistentURL: got error on get: %v", err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("testShortenURLAddingANonExistentURL: unexpected status code: %v", response.StatusCode)
	}

	bodyByte, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return fmt.Errorf("testShortenURLAddingANonExistentURL: error reading body: %v", err)
	}

	body := string(bodyByte)
	re := regexp.MustCompile(shortURLRegExp)
	gotShortURL := strings.TrimSpace(re.FindStringSubmatch(body)[1])
	if gotShortURL == "" {
		return fmt.Errorf("testShortenURLAddingANonExistentURL: no short URL on shorten page")
	}

	return nil
}

func testRedirectURL() error {
	client := &http.Client{}

	response, err := client.Get(redirectURL)

	if err != nil {
		return fmt.Errorf("testRedirectURL: got error on get: %v", err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("testRedirectURL: unexpected status code: %v", response.StatusCode)
	}

	requestURL := response.Request.URL
	url := requestURL.String()

	if url != weatherURLRome {
		return fmt.Errorf("testRedirectURL: wrong redirect, wanted: %v, got: %v", weatherURLRome, url)
	}

	return nil
}

func testURLAddedToPersistenceFile() error {
	urlShortener := shorten.NewURLShortener()

	f, err := os.Open(persistenceFile)
	if err != nil {
		return fmt.Errorf("testURLAddedToPersistenceFile: error opening persistence file: %v", err)
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	if err := urlShortener.UnpersistFrom(reader); err != nil {
		return fmt.Errorf("testURLAddedToPersistenceFile: error unpersisting persistence file: %v", err)
	}

	gotLongURL, err := urlShortener.GetURL(shortWeatherURLRome)
	if err != nil {
		return fmt.Errorf("testURLAddedToPersistenceFile: error on GetURL call: %v", err)
	}

	if weatherURLRome != gotLongURL {
		return fmt.Errorf("testURLAddedToPersistenceFile: persisted URL not found, wanted: %v, got: %v, error: %v", weatherURLRome, gotLongURL, err)
	}

	return nil
}

func trueMain() int {
	cmd := exec.Command(serverExecutable, loadFlag, persistenceFile)
	err := cmd.Start()

	if err != nil {
		log.Println("trueMain: error starting server:", err)
		return exitCodeError
	}

	time.Sleep(oneSecond) // to let the goroutine go

	defer func() {
		cmd.Process.Signal(syscall.SIGINT)
		time.Sleep(oneSecond) // to let the kill kill
		cmd.Wait()
	}()

	if err := testNonExistentHash(); err != nil {
		return exitCodeError
	}

	if err := testWeatherHash(); err != nil {
		return exitCodeError
	}

	if err := testStatisticsJSON(); err != nil {
		return exitCodeError
	}

	if err := testShortenURLAddingANonExistentURL(); err != nil {
		return exitCodeError
	}

	if err := testRedirectURL(); err != nil {
		return exitCodeError
	}

	return exitCodeOk
}

func main() {
	if err := copyFromFileToFile(sourcePersistenceFile, persistenceFile); err != nil {
		log.Println("main: persistence file not copied. Error:", err)
		os.Exit(exitCodeError)
	}

	exitCode := trueMain()

	if err := testURLAddedToPersistenceFile(); err != nil {
		exitCode = exitCodeError
	}

	os.Exit(exitCode)
}
