package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/middleware"
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

// PostJSON posts json data to a URI.
func (c *Client) PostJSON(function string, json []byte) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", c.BaseEndpoint, function)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create json post request")
	}
	req.Header.Set("Content-Type", "application/json")

	client := middleware.LoggingClient{}
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
func (c *Client) PostFile(function string, filename string, params map[string]string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", c.BaseEndpoint, function)

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// add the file
	f, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read file")
	}
	defer f.Close()

	fw, err := w.CreateFormFile("file", filename)
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

	client := &middleware.LoggingClient{}
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

	client := &middleware.LoggingClient{}
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
