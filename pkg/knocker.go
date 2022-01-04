package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

// Knocker is http client wrapper that Keep connection strings and Bearer token.
type Knocker struct {
	Host  string
	Port  int
	Token string
}

// Knock send request with Bearer token and unmarshal response json.
func (k *Knocker) Knock(method string, path string, param io.Reader, v interface{}) (response *http.Response, body string, err error) {
	url := fmt.Sprintf("http://%s:%d%s", k.Host, k.Port, path)
	request, err := http.NewRequest(method, url, param)
	if err != nil {
		return nil, "", err
	}
	if k.Token != "" {
		request.Header.Add("Authorization", "Bearer "+k.Token)
	}
	res, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, "", err
	}
	defer func() {
		err := res.Body.Close()
		if err != nil {
			panic(err)
		}
	}()
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, "", err
	}

	if v != nil {
		if err := json.Unmarshal(resBody, v); err != nil {
			return nil, "", err
		}
	}

	return res, string(resBody), nil
}

// Auth send request and keep response string in Token.
func (k *Knocker) Auth(method string, path string, param io.Reader) (response *http.Response, body string, err error) {
	response, body, err = k.Knock(method, path, param, nil)
	k.Token = body
	return
}
