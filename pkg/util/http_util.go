package util

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func Request(uri string, method string, body string, auth bool) []byte {
	client := &http.Client{}
	request, err := http.NewRequest(method, uri, strings.NewReader(body))
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	request.Header.Add("Content-Type", "application/json")
	if auth {
		request.Header.Add("passport-token", getToken())
	}

	response, err := client.Do(request)
	if err != nil {
		log.Println(err.Error())
		return nil
	}
	defer response.Body.Close()

	resp, _ := ioutil.ReadAll(response.Body)
	return resp
}

func RequestWithAuth(uri string, method string, body string) []byte {
	return Request(uri, method, body, true)
}
