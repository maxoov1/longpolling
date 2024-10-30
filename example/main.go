package main

import (
	"net/http"
	"time"

	"github.com/maxoov1/longpolling"
)

func responseHandler(response *http.Response, err error) {
	println(response.Status)
}

func main() {
	client := longpolling.NewClient(
		"https://example.com/",
		longpolling.WithTimeout(5*time.Second),
		longpolling.WithResponseHandler(responseHandler),
	)

	client.Start()
	select {}
}
