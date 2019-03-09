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
	Dump    bool
}

//NewClient create a rest client
func NewClient(dump bool) *Client {
	return &Client{
		client:  &http.Client{},
		Headers: map[string]string{},
		Dump:    dump,
	}
}

func (c *Client) do(method string, url string, payload interface{}) ([]byte, error) {
	// be sure that method is in uppercase
	method = strings.ToUpper(method)
	if c.Dump {
		fmt.Printf("request > %s %s\n", method, url)
	}
	// create payload io reader
	var reader io.Reader
	if payload != nil {
		// marshall payload
		jsonpayload, err := json.Marshal(payload)
		if err != nil {
			return []byte(""), fmt.Errorf("cannot serialise payload")
		}
		if c.Dump {
			fmt.Printf("payload ---\njsonpayload\npayload ---")
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
	// dump request headers
	if c.Dump {
		fmt.Println("request headers ---")
		for key, value := range request.Header {
			if key == "Authorization" {
				parts := strings.Split(value[0], " ")
				parts[len(parts)-1] = "---HIDDEN---"
				fmt.Printf("%s: %s\n", key, strings.Join(parts, " "))
				continue
			}
			fmt.Printf("%s: %s\n", key, strings.Join(value, "; "))

		}
		fmt.Println("request headers ---")
	}
	// do the request
	response, err := c.client.Do(request)
	if err != nil {
		return []byte(""), fmt.Errorf("error executing request")
	}
	// dump response headers
	if c.Dump {
		fmt.Println("response headers ---")
		for key, value := range response.Header {
			fmt.Printf("%s: %s\n", key, strings.Join(value, "; "))
		}
		fmt.Println("response headers ---")
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
	if c.Dump {
		fmt.Printf("response ---\n%s\nresponse ---\n", string(data))
	}
	if response.StatusCode >= 300 || response.StatusCode < 200 {
		return data, fmt.Errorf(response.Status)
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
