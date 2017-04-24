// Copyright © 2016 Crowdfire Inc <opensource@crowdfireapp.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

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
		{`{"authUid": "123-xy"}`, true},
		{`{"authUid": "123-xy"`, false},
		{"xyz", false},
	}

	for _, c := range cases {
		if actual := sendMessageToURL(c.msg); actual != c.expected {
			t.Errorf("sendMessageToURL(%#v) expected %#v, but got %#v", c.msg, c.expected, actual)
		}
	}

	defer ts.Close()
}

func TestEncodeURL(t *testing.T) {
	// test 1 - simple url
	url1 := "https://example.com"
	expectedEncodedURL1 := "https://example.com"
	encodedURL1 := encodeURL(url1)
	if expectedEncodedURL1 != encodedURL1 {
		t.Errorf("encodeURL(%#v) expected %#v, but got %#v", url1, expectedEncodedURL1, encodedURL1)
	}

	// test 2 - simple url with extra path
	url2 := "https://example.com/path"
	expectedEncodedURL2 := "https://example.com/path"
	encodedURL2 := encodeURL(url2)
	if expectedEncodedURL2 != encodedURL2 {
		t.Errorf("encodeURL(%#v) expected %#v, but got %#v", url2, expectedEncodedURL2, encodedURL2)
	}

	// test 3 - url with query params without special character
	url3 := "https://example.com/path?key1=name&key2=age"
	expectedEncodedURL3 := "https://example.com/path?key1=name&key2=age"
	encodedURL3 := encodeURL(url3)
	if expectedEncodedURL3 != encodedURL3 {
		t.Errorf("encodeURL(%#v) expected %#v, but got %#v", url3, expectedEncodedURL3, encodedURL3)
	}

	// test 4 - url with query params with special character
	url4 := "https://example.com/path?key1=name surname&key2=age"
	expectedEncodedURL4 := "https://example.com/path?key1=name+surname&key2=age"
	encodedURL4 := encodeURL(url4)
	if expectedEncodedURL4 != encodedURL4 {
		t.Errorf("encodeURL(%#v) expected %#v, but got %#v", url4, expectedEncodedURL4, encodedURL4)
	}

}
