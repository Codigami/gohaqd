package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var ts *httptest.Server

func init() {
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data map[string]interface{}
		bodyBytes, _ := ioutil.ReadAll(r.Body)
		if err := json.Unmarshal(bodyBytes, &data); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Invalid JSON.")
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))

	url = ts.URL
}

func TestSendMessageToURL(t *testing.T) {
	cases := []struct {
		msg      string
		expected bool
	}{
		{"{\"authUid\": \"123-xy\"}", true},
		{"{\"authUid\": \"123-xy\"", false},
		{"xyz", false},
	}

	for _, c := range cases {
		if actual := sendMessageToURL(c.msg); actual != c.expected {
			t.Errorf("sendMessageToURL(%#v) expected %#v, but got %#v", c.msg, c.expected, actual)
		}
	}

	defer ts.Close()
}
