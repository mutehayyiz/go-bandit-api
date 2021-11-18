package main

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {

	//init test runner
	tr = NewTestRunner()
	defer tr.Server.Close()

	// get config
	config, err := LoadConfig("config.conf")
	if err != nil {
		panic(err)
	}

	// connect database
	err = StorageConnect(&config.Storage)
	if err != nil {
		panic(err)
	}

	ret := m.Run()
	os.Exit(ret)
}

type TestCase struct {
	Method string
	Path   string
	Body   string
}

type TestRunner struct {
	Server *httptest.Server
	Client *http.Client
}

var tr *TestRunner

func NewTestRunner() *TestRunner {
	return &TestRunner{
		Server: httptest.NewServer(GenerateRouter()),
		Client: &http.Client{},
	}
}

func (t *TestRunner) Run(tc *TestCase) ([]byte, error) {
	reader := strings.NewReader(tc.Body)

	req, err := http.NewRequest(tc.Method, tr.Server.URL+tc.Path, reader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.Client.Do(req)
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return bytes, err
}

var testData = &struct {
	ID  string
	Url string
}{
	Url: "https://github.com/anxolerd/dvpwa",
}

func TestNewScan(t *testing.T) {
	tc := TestCase{
		Method: "POST",
		Path:   "/scan",
		Body:   fmt.Sprintf(`{"url":"%s"}`, testData.Url),
	}

	bytes, err := tr.Run(&tc)
	assert.NoError(t, err)

	var res map[string]string
	err = json.Unmarshal(bytes, &res)
	assert.NoError(t, err)

	assert.Empty(t, res["error"])
	assert.NotEmpty(t, res["id"])

	testData.ID = res["id"]
}

func TestGetScan(t *testing.T) {
	tc := TestCase{
		Method: "GET",
		Path:   fmt.Sprintf(`/scan/%s`, testData.ID),
	}

	bytes, err := tr.Run(&tc)
	assert.NoError(t, err)

	var sc Scan
	err = json.Unmarshal(bytes, &sc)
	assert.NoError(t, err)

	assert.Equal(t, testData.ID, sc.ID)
	assert.Equal(t, testData.Url, sc.URL)
	assert.NotEqual(t, "error", sc.Status)
	assert.Empty(t, sc.Error)
}
