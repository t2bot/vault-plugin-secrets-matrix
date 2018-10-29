package matrix

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
)

func combineUrl(base string, path string) (string, error) {
	u, err := url.Parse(path)
	if err != nil {
		return "", err
	}

	b, err := url.Parse(base)
	if err != nil {
		return "", err
	}

	return b.ResolveReference(u).String(), nil
}

func getJson(base string, path string) (map[string]interface{}, error) {
	route, err := combineUrl(base, path)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(route)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	var m map[string]interface{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func postJson(base string, path string, reqBody map[string]interface{}) (map[string]interface{}, error) {
	route, err := combineUrl(base, path)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(route, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	var m map[string]interface{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}
