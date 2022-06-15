package main

import (
	"encoding/json"
	"io"
	"net/http"
)

type ApiResp struct {
	Servers []string `json:"servers"`
}

func callApi(endpoint string) (*ApiResp, error) {
	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}

	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	var api *ApiResp
	err = json.NewDecoder(resp.Body).Decode(&api)
	if err != nil {
		return nil, err
	}
	return api, err
}
