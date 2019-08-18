package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

type RequestType int

//Some categories for the different requests
const (
	UNKNOWN RequestType = iota
	JSON
	HTTP_POST
	BLOB //Unused, could be a valid category
)

//RequestStats represents a request made to a load tested server
type RequestStats struct {
	RequestNumber int
	RequestType   RequestType
	RequestTime   time.Duration
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Invalid amount of program arguments")
	}
	concurrentUsers, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("Invalid number %v", err)
	}

	//Channel which all results are put on
	//Multiply by 2 because we make 2 request in our simulation
	//An improvement would be to not hardcode and make the channel capacity dynamic because this could cause channel deadlock bugs if not handled properly
	requestResultsChannel := make(chan RequestStats, concurrentUsers*2)

	loadTestStartTime := time.Now()
	//Limit the amount of concurrent requests being made because of file handles limit caused by too many open sockets
	//adjust as the target machine is capable
	concurrencyLimit := 200
	sem := make(chan bool, concurrencyLimit)
	go func() {
		for i := 0; i < concurrentUsers; i++ {
			go func(requestNumber int) {
				simulateUserInteraction(requestNumber, requestResultsChannel)
				<-sem
			}(i)
			sem <- true
		}
		//Make sure the last goroutines are finished
		for i := 0; i < cap(sem); i++ {
			sem <- true
		}

		close(requestResultsChannel)
	}()

	//collecting the data which is being produced concurrently
	allRequestStats := []RequestStats{}
	for result := range requestResultsChannel {
		allRequestStats = append(allRequestStats, result)
	}

	log.Printf("Load test is done. Time taken = %v", time.Since(loadTestStartTime))

	requestTypeTostring := func(r RequestType) string {
		switch r {
		case JSON:
			return "JSON"
		case HTTP_POST:
			return "POST"
		default:
			return "UNKNOWN"
		}
	}

	fileName := "request_stats.csv"
	f, _ := os.Create(fileName)
	defer f.Close()
	//Write csv headers
	_, err = f.Write([]byte("RequestNumber;RequestType;RequestTime(ms)\n"))
	if err != nil {
		log.Fatal(err)
	}
	//Write request stats to file
	for _, stat := range allRequestStats {
		_, err = f.Write([]byte(
			fmt.Sprintf(
				"%d;%f;%s\n",
				stat.RequestNumber,
				float64(stat.RequestTime)/float64(time.Millisecond), //Request time in miliseconds
				requestTypeTostring(stat.RequestType),
			),
		))
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println(fmt.Sprintf("Program done. Wrote stats to file: %s", fileName))
}

//simulateUserInteraction makes http requests to simulate user interaction and measures the time the request take and sends associated data on requestResultsChannel
func simulateUserInteraction(requestNumber int, requestResultsChannel chan RequestStats) {
	jsonRequestTime, err := measureTime(func() error {
		resp, err := http.Get("http://localhost:8080/json")
		if err != nil {
			return fmt.Errorf("HTTP request failed. Error = %v", err)
		}
		defer resp.Body.Close()
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	requestResultsChannel <- RequestStats{
		RequestNumber: requestNumber,
		RequestType:   JSON,
		RequestTime:   jsonRequestTime,
	}

	//time.Sleep(1 * time.Second) //Simulate user taking some time to talk to the backend again

	postRequestTime, err := measureTime(func() error {
		data := url.Values{}
		data.Set("form_data", "Hello world!")
		resp, err := http.Post("http://localhost:8080/echo", "", bytes.NewBufferString(data.Encode()))
		if err != nil {
			return fmt.Errorf("HTTP request failed. Error = %v", err)
		}
		defer resp.Body.Close()
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	requestResultsChannel <- RequestStats{
		RequestNumber: requestNumber,
		RequestType:   HTTP_POST,
		RequestTime:   postRequestTime,
	}
}

// measureTime measures the time a function takes to complete
func measureTime(fn func() error) (time.Duration, error) {
	startTime := time.Now()
	err := fn()
	if err != nil {
		return 0, err
	}
	return time.Since(startTime), nil
}
