//
//   Copyright © 2021 Uncharted Software Inc.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package rest

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil/api/middleware"
)

// ClientCtor repressents a client constructor to instantiate a rest client.
type ClientCtor func() *Client

// Client represents a basic REST client.
type Client struct {
	BaseEndpoint string
}

// NewClient instantiates and returns a new REST client constructor.
func NewClient(endpoint string) ClientCtor {
	return func() *Client {
		return &Client{
			BaseEndpoint: endpoint,
		}
	}
}

// Get performs a get using the provided params as query string parameters.
func (c *Client) Get(function string, params map[string]string) ([]byte, error) {
	url := function
	if c.BaseEndpoint != "" {
		url = fmt.Sprintf("%s/%s", c.BaseEndpoint, function)
	}

	queryString := ""
	for name, value := range params {
		queryString = fmt.Sprintf("%s&%s=%s", queryString, name, value)
	}
	// skip first &
	if len(queryString) > 0 {
		queryString = queryString[1:]
		url = fmt.Sprintf("%s?%s", url, queryString)
	}

	client := createClient()
	resp, err := client.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get result from json post")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read response body")
	}

	return body, nil
}

// PostJSON posts json data to a URI.
func (c *Client) PostJSON(function string, json []byte) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", c.BaseEndpoint, function)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create json post request")
	}
	req.Header.Set("Content-Type", "application/json")

	client := createClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get result from json post")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read response body")
	}

	return body, nil
}

// PostFile submits a file in a POST request using a multipart form.
func (c *Client) PostFile(function string, fileParameterName string, filename string, params map[string]string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", c.BaseEndpoint, function)

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// add the file
	f, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read file")
	}
	defer f.Close()

	fw, err := w.CreateFormFile(fileParameterName, filename)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create form request")
	}
	if _, err = io.Copy(fw, f); err != nil {
		return nil, errors.Wrap(err, "Unable to copy file")
	}

	// add the parameters
	for name, value := range params {
		err := w.WriteField(name, value)
		if err != nil {
			return nil, errors.Wrap(err, "unable to add parameter field")
		}
	}
	w.Close()

	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create request")
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	client := createClient()
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to post request")
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", res.Status)
	}

	result, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read result")
	}

	return result, nil
}

// PostRequest submits a post request with the provided parameters.
func (c *Client) PostRequest(function string, params map[string]string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", c.BaseEndpoint, function)

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// add the parameters
	for name, value := range params {
		err := w.WriteField(name, value)
		if err != nil {
			return nil, errors.Wrap(err, "unable to add parameter field")
		}
	}
	w.Close()

	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create request")
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	client := createClient()
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to post request")
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", res.Status)
	}

	result, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read result")
	}

	return result, nil
}

// PostRequestRaw submits a post request with the provided parameters
// submitted as a raw string.
func (c *Client) PostRequestRaw(function string, params map[string]interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", c.BaseEndpoint, function)
	b, err := json.Marshal(params)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal parameters")
	}

	// interface requires double marshalling to have a raw string
	b, err = json.Marshal(string(b))
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal (*2) parameters")
	}

	res, err := http.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, errors.Wrap(err, "Unable to post request")
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", res.Status)
	}

	result, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read result")
	}

	return result, nil
}

func createClient() *middleware.LoggingClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := &middleware.LoggingClient{
		Client: http.Client{
			Transport: tr,
		},
	}
	return client
}
