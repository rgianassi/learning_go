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
func copy(src, dst string) error {
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
		fmt.Println("got error on get for non existent hash:", err)
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusNotFound {
		fmt.Println("expected not found response on non existent hash, but got:", response.StatusCode)
		return err
	}

	return nil
}

func testWeatherHash() error {
	client := &http.Client{}

	response, err := client.Get(fullShortWeatherURLFlorence)

	if err != nil {
		fmt.Println("got error on get for weather hash:", err)
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		fmt.Println("expected ok response on existent hash, but got:", response.StatusCode)
		return err
	}

	requestURL := response.Request.URL
	url := requestURL.String()

	if url != weatherURLFlorence {
		fmt.Println("expected redirect to", weatherURLFlorence, ", but got:", url)
		return err
	}

	return nil
}

func testStatisticsJSON() error {
	client := &http.Client{}

	response, err := client.Get(statisticsURL)

	if err != nil {
		fmt.Println("got error on get for statistics JSON:", err)
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		fmt.Println("expected ok response on statistics JSON, but got:", response.StatusCode)
		return err
	}

	decoder := json.NewDecoder(response.Body)

	var stats shorten.StatsJSON

	if err := decoder.Decode(&stats); err != nil {
		fmt.Println("unable to decode statistics JSON:", err)
		return err
	}

	gotTotalURL := stats.ServerStats.TotalURL
	if gotTotalURL != 1 {
		fmt.Println("expected TotalURL on statistics JSON to be 1, but got:", gotTotalURL)
		return err
	}

	return nil
}

func testShortenURLAddingANonExistentURL() error {
	client := &http.Client{}

	response, err := client.Get(shortenWeatherURLRome)

	if err != nil {
		fmt.Println("got error on get for shorten weather URL:", err)
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		fmt.Println("expected ok response for shorten weather URL, but got:", response.StatusCode)
		return err
	}

	bodyByte, err := ioutil.ReadAll(response.Body)

	if err != nil {
		fmt.Println("got error for shorten weather URL reading body:", err)
		return err
	}

	body := string(bodyByte)
	re := regexp.MustCompile(shortURLRegExp)
	gotShortURL := strings.TrimSpace(re.FindStringSubmatch(body)[1])
	if gotShortURL == "" {
		msg := "expected short URL on shorten weather URL call, but nothing got"
		fmt.Println(msg)
		return fmt.Errorf(msg)
	}

	return nil
}

func testRedirectURL() error {
	client := &http.Client{}

	response, err := client.Get(redirectURL)

	if err != nil {
		fmt.Println("got error on get for redirect weather URL hash:", err)
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		fmt.Println("expected ok response on redirect weather URL hash, but got:", response.StatusCode)
		return err
	}

	requestURL := response.Request.URL
	url := requestURL.String()

	if url != weatherURLRome {
		fmt.Println("expected redirect to", weatherURLRome, ", but got:", url)
		return err
	}

	return nil
}

func testURLAddedToPersistenceFile() error {
	urlShortener := shorten.NewURLShortener()

	f, err := os.Open(persistenceFile)
	if err != nil {
		log.Println("error unpersisting URL added to persistence file:", err)
		return err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	if err := urlShortener.UnpersistFrom(reader); err != nil {
		fmt.Println("unable to unpersist URL added to persistence file:", persistenceFile, "err:", err)
		return err
	}

	gotLongURL, err := urlShortener.GetURL(shortWeatherURLRome)
	if err != nil {
		fmt.Println("unable to get URL added to persistence file:", err)
		return err
	}

	if weatherURLRome != gotLongURL {
		fmt.Println("unable to find URL added to persistence file:", err)
		return err
	}

	return nil
}

func trueMain(serverClosed chan struct{}) int {
	cmd := exec.Command(serverExecutable, loadFlag, persistenceFile)
	err := cmd.Start()

	if err != nil {
		log.Println(err)
		return exitCodeError
	}

	go func(serverClosed chan struct{}) {
		cmd.Wait()
		close(serverClosed)
	}(serverClosed)

	time.Sleep(oneSecond) // to let the goroutine go

	defer func() {
		cmd.Process.Signal(syscall.SIGINT)
		time.Sleep(oneSecond) // to let the kill kill
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
	serverClosed := make(chan struct{})

	if err := copy(sourcePersistenceFile, persistenceFile); err != nil {
		log.Println("persistence file not copied. Error:", err)
		os.Exit(exitCodeError)
	}

	exitCode := trueMain(serverClosed)

	<-serverClosed
	if err := testURLAddedToPersistenceFile(); err != nil {
		exitCode = exitCodeError
	}

	os.Exit(exitCode)
}
