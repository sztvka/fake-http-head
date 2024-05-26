package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"golang.org/x/net/http2"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"time"
)

type Result struct {
	Headers        map[string]interface{} `json:"headers"`
	Status         string                 `json:"status"`
	HttpStatusCode int                    `json:"http_status_code"`
}

type Input struct {
	Url            string            `json:"url"`
	ProxyUrl       string            `json:"proxy_url"`
	TimeoutMs      uint64            `json:"timeout_ms"`
	RequestHeaders map[string]string `json:"headers"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Print(`{"status": "error missing json input"}`)
		os.Exit(1)
	}
	var input Input
	err := json.Unmarshal([]byte(os.Args[1]), &input)
	if err != nil {
		fmt.Print(`{"status": "error invalid json input"}`)
		os.Exit(1)
	}

	fmt.Print(FakeHead(input.Url, input.ProxyUrl, input.TimeoutMs, input.RequestHeaders))
}

func FakeHead(rawUrl string, rawProxyUrl string, timeoutMs uint64, requestHeaders map[string]string) string {
	proxyURL, err := url.Parse(rawProxyUrl)
	if err != nil {
		panic(err)
	}
	proxy := http.ProxyURL(proxyURL)

	transport := &http.Transport{
		Proxy: proxy,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	err = http2.ConfigureTransport(transport)
	if err != nil {
		panic(err)
	}

	client := &http.Client{
		Transport: transport,
	}

	req, err := http.NewRequest("GET", rawUrl, nil)
	if err != nil {
		panic(err)
	}
	for key, value := range requestHeaders {
		req.Header.Add(key, value)
	}

	ctx, cancel := context.WithCancel(context.Background())
	timeoutChan := time.NewTimer(time.Duration(timeoutMs) * time.Millisecond).C
	defer cancel()
	req = req.WithContext(ctx)
	c1 := make(chan string, 1)

	go func() {
		resp, respErr := client.Do(req)
		if respErr != nil {
			errorResult, marshalErr := json.Marshal(Result{
				Headers: make(map[string]interface{}),
				Status:  "error",
			})
			if marshalErr != nil {
				panic(marshalErr)
			}
			c1 <- string(errorResult)
		}

		cancel()
		//got ctx cancelled
		if resp == nil || respErr != nil {
			return
		}
		responseHeaders, marshalErr := json.Marshal(Result{
			Headers:        flattenSingleElementArrays(resp.Header),
			Status:         "ok",
			HttpStatusCode: resp.StatusCode,
		})
		if marshalErr != nil {
			panic(marshalErr)
		}

		if resp.Body != nil {
			resp.Body.Close()
		}
		c1 <- string(responseHeaders)
	}()

	select {
	case <-timeoutChan:
		cancel()
		timeoutResult, marshalErr := json.Marshal(Result{Status: "timeout"})
		if marshalErr != nil {
			panic(marshalErr)
		}
		return string(timeoutResult)
	case result := <-c1:
		close(c1)
		return result
	}
}

func flattenSingleElementArrays(originalMap map[string][]string) map[string]interface{} {
	flattenedMap := make(map[string]interface{})
	for key, value := range originalMap {
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Slice && v.Len() == 1 {
			flattenedMap[key] = v.Index(0).String()
		} else {
			flattenedMap[key] = value
		}
	}
	return flattenedMap
}
