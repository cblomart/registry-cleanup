package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	jsonMmime         = "application/json"
	headerContentType = "Content-Type"
	headerAccept      = "Accept"
)

//Client is a simple rest client
type Client struct {
	client  *http.Client
	Headers map[string]string
}

//NewClient create a rest client
func NewClient() *Client {
	return &Client{
		client:  &http.Client{},
		Headers: map[string]string{},
	}
}

func (c *Client) do(method string, url string, payload interface{}) ([]byte, error) {
	// be sure that method is in uppercase
	method = strings.ToUpper(method)
	// create payload io reader
	var reader io.Reader
	if payload != nil {
		// marshall payload
		jsonpayload, err := json.Marshal(payload)
		if err != nil {
			return []byte(""), fmt.Errorf("cannot serialise payload")
		}
		reader = bytes.NewBuffer(jsonpayload)
	}
	// create the request
	request, err := http.NewRequest(method, url, reader)
	if err != nil {
		return []byte(""), fmt.Errorf("cannot create request")
	}
	// set default rest headers
	if payload != nil {
		request.Header.Set(headerContentType, jsonMmime)
	}
	if method != "HEAD" {
		request.Header.Set(headerAccept, jsonMmime)
	}
	// set the requested headers
	for key, value := range c.Headers {
		request.Header.Set(key, value)
	}
	// do the request
	response, err := c.client.Do(request)
	if err != nil {
		return []byte(""), fmt.Errorf("error executing request")
	}
	if method == "HEAD" {
		body, err := json.Marshal(response.Header)
		if err != nil {
			return []byte(""), fmt.Errorf("cannot marshal response headers")
		}
		return body, nil
	}
	// read the response
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return []byte(""), fmt.Errorf("cannot read response body")
	}
	return data, nil
}

//Get does a get request
func (c *Client) Get(url string, payload interface{}, v interface{}) error {
	data, err := c.do("GET", url, payload)
	if err != nil {
		return err
	}
	if v != nil {
		return json.Unmarshal(data, v)
	}
	return nil
}

//Head does a get request
func (c *Client) Head(url string, payload interface{}, v interface{}) error {
	data, err := c.do("HEAD", url, payload)
	if err != nil {
		return err
	}
	if v != nil {
		return json.Unmarshal(data, v)
	}
	return nil
}

//Delete does a delete request
func (c *Client) Delete(url string, payload interface{}, v interface{}) error {
	data, err := c.do("DELETE", url, payload)
	if err != nil {
		return err
	}
	if v != nil {
		return json.Unmarshal(data, v)
	}
	return nil
}

//Post does a post request
func (c *Client) Post(url string, payload interface{}, v interface{}) error {
	data, err := c.do("POST", url, payload)
	if err != nil {
		return err
	}
	if v != nil {
		return json.Unmarshal(data, v)
	}
	return nil
}
