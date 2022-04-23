package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
)

func get(url string) *http.Response {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err, url)
	}

	backo := backoff.NewExponentialBackOff()
	for {
		res1, err := http.Get(url)
		if err != nil {
			log.Fatal(err, url)
		}
		if res1.StatusCode < 300 {
			break
		}

		body, err := ioutil.ReadAll(res.Body)
		fmt.Println(string(body))
		sleepTime := backo.NextBackOff()
		dur := time.Duration(sleepTime)
		fmt.Printf("request to {%s} failed with {%d | %s}. Sleeping {%s} till next try\n", url, res1.StatusCode, res1.Status, sleepTime)
		time.Sleep(dur)
	}
	return res
}
