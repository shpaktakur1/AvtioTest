package rest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"
)

func executeRequest(a *App, r *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	a.Router.ServeHTTP(recorder, r)
	return recorder
}

func checkResponseCode(t *testing.T, name string, expected int, got int) {
	if expected != got {
		t.Errorf("%v: Expected response code %v, got %v", name, expected, got)
	}
}

func checkResponseBody(t *testing.T, name string, expected string, got string) {
	if expected != got {
		t.Errorf("%v: Expected response body %v, got %v", name, expected, got)
	}
}

func TestApp_authorization(t *testing.T) {
	a := &App{}
	a.Authorization = &BasicAuthorizer{Username: "Alladin", Password: "Open Sesame"}
	a.Initialize(0, nil, nil, 500, 1, nil)
	a.Cache.Set("string", "something", 0)

	type actionTest struct {
		name string

		method string
		url    string
		body   io.Reader

		authorize bool
		user      string
		password  string

		expectedCode int
		expectedBody string
	}

	var tests = []actionTest{
		{"NoAuth", "GET", "/string", nil, false, "", "", http.StatusUnauthorized, "Unauthorized\n"},
		{"WrongAuth", "GET", "/string", nil, true, "admin", "admin", http.StatusUnauthorized, "Unauthorized\n"},
		{"CorrectAuth", "GET", "/string", nil, true, "Alladin", "Open Sesame", http.StatusOK, `{"type":0,"data":"something"}`},
	}

	for _, tt := range tests {
		req, _ := http.NewRequest(tt.method, tt.url, tt.body)
		if tt.authorize {
			req.SetBasicAuth(tt.user, tt.password)
		}
		response := executeRequest(a, req)
		checkResponseCode(t, tt.name, tt.expectedCode, response.Code)
		checkResponseBody(t, tt.name, tt.expectedBody, response.Body.String())
	}

}

func TestApp_ttl(t *testing.T) {
	a := &App{}
	a.Initialize(0, nil, nil, 500, 1, nil)
	a.Cache.Set("string", "something", 10)
}

func TestApp_actions(t *testing.T) {
	a := &App{}
	a.Initialize(0, nil, nil, 500, 1, nil)
	a.Cache.Set("string", "something", 0)
	a.Cache.Set("string_to_remove", "something_else", 0)
	a.Cache.Set("list", []interface{}{1, "abc", nil, 3.2, []int{1, 2, 3}}, 0)
	a.Cache.Set("map", map[string]interface{}{"key": "value", "int": 1, "map": map[int]int{42: 24}}, 0)

	type actionTest struct {
		name string

		method string
		url    string
		body   io.Reader

		expectedCode int
		expectedBody string
	}

	var tests = []actionTest{
		{"Get invalid key", "GET", "/invalid", nil, http.StatusBadRequest, `{"error":"key not found"}`},
		{"Get string", "GET", "/string", nil, http.StatusOK, `{"type":0,"data":"something"}`},
		{"Get list", "GET", "/list", nil, http.StatusOK, `{"type":1,"data":[1,"abc",null,3.2,[1,2,3]]}`},
		{"Remove item", "DELETE", "/string_to_remove", nil, http.StatusOK, `"OK"`},
		{"Attempt to remove non-existent item", "DELETE", "/invalid", nil, http.StatusOK, `"OK"`},

		{"Get string[0]", "GET", "/string/0", nil, http.StatusBadRequest, `{"error":"cant Get item at index"}`},

		{"Get list[0] with int value", "GET", "/list/0", nil, http.StatusOK, `1`},
		{"Get list[1] with str value", "GET", "/list/1", nil, http.StatusOK, `"abc"`},
		{"Get list[2] with nil value", "GET", "/list/2", nil, http.StatusOK, `null`},
		{"Get list[3] with float value", "GET", "/list/3", nil, http.StatusOK, `3.2`},
		{"Get list[4] with []int value", "GET", "/list/4", nil, http.StatusOK, `[1,2,3]`},
		{"Get list[5] not exist", "GET", "/list/5", nil, http.StatusBadRequest, `{"error":"cant Get item at index"}`},

		{"Get map[key] with str", "GET", "/map/key", nil, http.StatusOK, `"value"`},
		{"Get map[key] with map[int]int", "GET", "/map/map", nil, http.StatusOK, `{"42":24}`},
		{"Get map[invalid] with no value", "GET", "/map/invalid", nil, http.StatusBadRequest, `{"error":"cant Get item at index"}`},

		{"Set new map", "POST", "/new_map", bytes.NewBufferString(`{"hello":"world"}`), http.StatusOK, `{"type":2,"data":{"hello":"world"}}`},
		{"Post negative TTL", "POST", "/ill?ttl=-15s", bytes.NewBufferString(`{"hello":"world"}`), http.StatusBadRequest, `{"error":"TTL should be positive"}`},
		{"Post malformed TTL", "POST", "/ill2?ttl=15z", bytes.NewBufferString(`{"hello":"world"}`), http.StatusBadRequest, `{"error":"Malformed duration"}`},
	}

	for _, tt := range tests {
		req, _ := http.NewRequest(tt.method, tt.url, tt.body)
		response := executeRequest(a, req)
		checkResponseCode(t, tt.name, tt.expectedCode, response.Code)
		checkResponseBody(t, tt.name, tt.expectedBody, response.Body.String())
	}

	tests = []actionTest{
		{"Get previously removed item", "GET", "/string_to_remove", nil, http.StatusBadRequest, `{"error":"key not found"}`},
		{"Get newly inserted map", "GET", "/new_map", nil, http.StatusOK, `{"type":2,"data":{"hello":"world"}}`},
		{"Ensure that negative ttl is not saved", "GET", "/ill", nil, http.StatusBadRequest, `{"error":"key not found"}`},
		{"Ensure that malformed ttl is not saved", "GET", "/ill2", nil, http.StatusBadRequest, `{"error":"key not found"}`},
	}

	for _, tt := range tests {
		req, _ := http.NewRequest(tt.method, tt.url, tt.body)
		response := executeRequest(a, req)
		checkResponseCode(t, tt.name, tt.expectedCode, response.Code)
		checkResponseBody(t, tt.name, tt.expectedBody, response.Body.String())
	}

	kt := actionTest{"Get keys", "GET", "/", nil, http.StatusOK, `["map","new_list","new_map","string","new_string","list"]`}
	req, _ := http.NewRequest(kt.method, kt.url, kt.body)
	response := executeRequest(a, req)
	checkResponseCode(t, kt.name, kt.expectedCode, response.Code)
	var keys []string
	expected := []string{"map", "new_map", "string", "list"}
	decoder := json.NewDecoder(response.Body)
	decoder.Decode(&keys)
	sort.Strings(keys)
	sort.Strings(expected)

	if !reflect.DeepEqual(keys, expected) {
		t.Errorf("Unexpected final state of keys: expected %v, got %v", expected, keys)
	}
}
