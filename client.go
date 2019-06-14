package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Client struct {
	client *http.Client
	url    string
}

func NewClient(url string) *Client {
	client := http.Client{
		Timeout: time.Second * 10, // Maximum of 10 secs
	}
	return &Client{
		client: &client,
		url:    url,
	}
}

func (client *Client) GetUnspentOuts(address string) ([]byte, error) {
	postData := fmt.Sprintf(`{"address":"%s", "amount":"0", "mixin":4, "use_dust":false, "dust_threshold":"1000000000"}`, address)
	return client.Call(client.url, "/get_unspent_outs", []byte(postData))
}

func (client *Client) GetRandomOuts(amounts []string, mixin uint32) ([]byte, error) {
	amountsByte, err := json.Marshal(amounts)
	if err != nil {
		return nil, err
	}
	postData := fmt.Sprintf(`{"amounts":%s, "count":%d}`, string(amountsByte), mixin)
	return client.Call(client.url, "/get_random_outs", []byte(postData))
}

func (client *Client) SubmitRawTx(tx string) ([]byte, error) {
	postData := fmt.Sprintf(`{"tx":"%s"}`, tx)
	return client.Call(client.url, "/submit_raw_tx", []byte(postData))
}

func (client *Client) Call(url, path string, message []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", url+path, bytes.NewBuffer(message))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
