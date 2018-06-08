package test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

func httpDo(method, url, trackId string, reqest []byte, t *testing.T) error {
	client := &http.Client{}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqest))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	t.Log(resp.Header.Get("TrackId"))

	t.Log(string(body))

	return nil
}
