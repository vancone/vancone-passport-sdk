package util

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/vancone/vancone-web-common-go/pkg/logger"
)

func Request(uri string, method string, body string, auth bool) []byte {
	client := &http.Client{}
	request, err := http.NewRequest(method, uri, strings.NewReader(body))
	if err != nil {
		logger.Error(err.Error())
		return nil
	}

	request.Header.Add("Content-Type", "application/json")
	if auth {
		request.Header.Add("passport-token", ApplyToken())
	}

	response, err := client.Do(request)
	if err != nil {
		logger.Error(err.Error())
		return nil
	}
	defer response.Body.Close()

	resp, _ := ioutil.ReadAll(response.Body)
	return resp
}

func RequestWithAuth(uri string, method string, body string) []byte {
	return Request(uri, method, body, true)
}
