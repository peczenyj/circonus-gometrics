// Copyright 2016 Circonus, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var (
	testMetricCluster = MetricCluster{
		Name: "test",
		CID:  "/metric_cluster/1234",
		Queries: []MetricQuery{
			MetricQuery{
				Query: "*Req*",
				Type:  "average",
			},
		},
		Description: "",
		Tags:        []string{},
	}
)

func testMetricClusterServer() *httptest.Server {
	f := func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/metric_cluster/1234": // handle GET/PUT/DELETE
			switch r.Method {
			case "PUT": // update
				defer r.Body.Close()
				b, err := ioutil.ReadAll(r.Body)
				if err != nil {
					panic(err)
				}
				w.WriteHeader(200)
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintln(w, string(b))
			case "GET": // get by id/cid
				ret, err := json.Marshal(testMetricCluster)
				if err != nil {
					panic(err)
				}
				w.WriteHeader(200)
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintln(w, string(ret))
			case "DELETE": // delete
				w.WriteHeader(200)
				fmt.Fprintln(w, "")
			default:
				w.WriteHeader(500)
				fmt.Fprintln(w, "unsupported")
			}
		case "/metric_cluster":
			switch r.Method {
			case "GET":
				reqURL := r.URL.String()
				var c []MetricCluster
				if reqURL == "/metric_cluster?search=web+servers" {
					c = []MetricCluster{testMetricCluster}
				} else if reqURL == "/metric_cluster?f_tags_has=dc%3Asfo1" {
					c = []MetricCluster{testMetricCluster}
				} else if reqURL == "/metric_cluster?f_tags_has=dc%3Asfo1&search=web+servers" {
					c = []MetricCluster{testMetricCluster}
				} else if reqURL == "/metric_cluster" {
					c = []MetricCluster{testMetricCluster}
				} else {
					c = []MetricCluster{}
				}
				if len(c) > 0 {
					ret, err := json.Marshal(c)
					if err != nil {
						panic(err)
					}
					w.WriteHeader(200)
					w.Header().Set("Content-Type", "application/json")
					fmt.Fprintln(w, string(ret))
				} else {
					w.WriteHeader(404)
					fmt.Fprintln(w, fmt.Sprintf("not found: %s %s", r.Method, reqURL))
				}
			case "POST": // create
				defer r.Body.Close()
				b, err := ioutil.ReadAll(r.Body)
				if err != nil {
					panic(err)
				}
				w.WriteHeader(200)
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintln(w, string(b))
			default:
				w.WriteHeader(500)
				fmt.Fprintln(w, "unsupported")
			}
		default:
			w.WriteHeader(500)
			fmt.Fprintln(w, "unsupported")
		}
	}

	return httptest.NewServer(http.HandlerFunc(f))
}

func TestFetchMetricClusterByCID(t *testing.T) {
	server := testMetricClusterServer()
	defer server.Close()

	var apih *API
	var err error
	var cluster *MetricCluster

	ac := &Config{
		TokenKey: "abc123",
		TokenApp: "test",
		URL:      server.URL,
	}
	apih, err = NewAPI(ac)
	if err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	t.Log("Testing invalid CID")
	{
		cid := "/invalid"
		expectedError := errors.New("Invalid metric cluster CID [/invalid]")
		_, err = apih.FetchMetricCluster(CIDType(&cid), "")
		if err == nil {
			t.Fatalf("Expected error")
		}
		if err.Error() != expectedError.Error() {
			t.Fatalf("Expected %+v got '%+v'", expectedError, err)
		}
	}

	t.Log("Testing valid CID")
	{
		cluster, err = apih.FetchMetricCluster(CIDType(&testMetricCluster.CID), "")
		if err != nil {
			t.Fatalf("Expected no error, got '%v'", err)
		}

		actualType := reflect.TypeOf(cluster)
		expectedType := "*api.MetricCluster"
		if actualType.String() != expectedType {
			t.Fatalf("Expected %s, got %s", expectedType, actualType.String())
		}

		if cluster.CID != testMetricCluster.CID {
			t.Fatalf("CIDs do not match: %+v != %+v\n", cluster, testMetricCluster)
		}
	}

	t.Log("Testing valid CID w/extras 'metrics'")
	{
		cluster, err = apih.FetchMetricCluster(CIDType(&testMetricCluster.CID), "metrics")
		if err != nil {
			t.Fatalf("Expected no error, got '%v'", err)
		}

		if cluster.CID != testMetricCluster.CID {
			t.Fatalf("CIDs do not match: %+v != %+v\n", cluster, testMetricCluster)
		}
	}

	t.Log("Testing valid CID w/extras 'uuids'")
	{
		cluster, err = apih.FetchMetricCluster(CIDType(&testMetricCluster.CID), "uuids")
		if err != nil {
			t.Fatalf("Expected no error, got '%v'", err)
		}

		if cluster.CID != testMetricCluster.CID {
			t.Fatalf("CIDs do not match: %+v != %+v\n", cluster, testMetricCluster)
		}
	}
}

func TestCreateMetricCluster(t *testing.T) {
	server := testMetricClusterServer()
	defer server.Close()

	var apih *API
	var err error

	ac := &Config{
		TokenKey: "abc123",
		TokenApp: "test",
		URL:      server.URL,
	}
	apih, err = NewAPI(ac)
	if err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	cluster, err := apih.CreateMetricCluster(&testMetricCluster)
	if err != nil {
		t.Fatalf("Expected no error, got '%v'", err)
	}

	actualType := reflect.TypeOf(cluster)
	expectedType := "*api.MetricCluster"
	if actualType.String() != expectedType {
		t.Fatalf("Expected %s, got %s", expectedType, actualType.String())
	}
}

func TestUpdateMetricCluster(t *testing.T) {
	server := testMetricClusterServer()
	defer server.Close()

	ac := &Config{
		TokenKey: "abc123",
		TokenApp: "test",
		URL:      server.URL,
	}
	apih, err := NewAPI(ac)
	if err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	t.Log("valid config")
	{
		cluster, err := apih.UpdateMetricCluster(&testMetricCluster)
		if err != nil {
			t.Fatalf("Expected no error, got '%v'", err)
		}

		actualType := reflect.TypeOf(cluster)
		expectedType := "*api.MetricCluster"
		if actualType.String() != expectedType {
			t.Fatalf("Expected %s, got %s", expectedType, actualType.String())
		}
	}

	t.Log("invalid CID in config")
	{
		expectedError := errors.New("Invalid metric cluster CID [/invalid]")
		x := &MetricCluster{CID: "/invalid"}
		_, err := apih.UpdateMetricCluster(x)
		if err == nil {
			t.Fatal("Expected an error")
		}
		if err.Error() != expectedError.Error() {
			t.Fatalf("Expected %+v got '%+v'", expectedError, err)
		}
	}
}

func TestSearchMetricClusters(t *testing.T) {
	server := testMetricClusterServer()
	defer server.Close()

	var apih *API

	ac := &Config{
		TokenKey: "abc123",
		TokenApp: "test",
		URL:      server.URL,
	}
	apih, err := NewAPI(ac)
	if err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	search := SearchQueryType("web servers")
	filter := SearchFilterType(map[string][]string{"f_tags_has": []string{"dc:sfo1"}})

	t.Log("no search, no filter")
	{
		clusters, err := apih.SearchMetricClusters(nil, nil)
		if err != nil {
			t.Fatalf("Expected no error, got '%v'", err)
		}

		actualType := reflect.TypeOf(clusters)
		expectedType := "*[]api.MetricCluster"
		if actualType.String() != expectedType {
			t.Fatalf("Expected %s, got %s", expectedType, actualType.String())
		}
	}

	t.Log("search, no filter")
	{
		clusters, err := apih.SearchMetricClusters(&search, nil)
		if err != nil {
			t.Fatalf("Expected no error, got '%v'", err)
		}

		actualType := reflect.TypeOf(clusters)
		expectedType := "*[]api.MetricCluster"
		if actualType.String() != expectedType {
			t.Fatalf("Expected %s, got %s", expectedType, actualType.String())
		}
	}

	t.Log("no search, filter")
	{
		clusters, err := apih.SearchMetricClusters(nil, &filter)
		if err != nil {
			t.Fatalf("Expected no error, got '%v'", err)
		}

		actualType := reflect.TypeOf(clusters)
		expectedType := "*[]api.MetricCluster"
		if actualType.String() != expectedType {
			t.Fatalf("Expected %s, got %s", expectedType, actualType.String())
		}
	}

	t.Log("search, filter")
	{
		clusters, err := apih.SearchMetricClusters(&search, &filter)
		if err != nil {
			t.Fatalf("Expected no error, got '%v'", err)
		}

		actualType := reflect.TypeOf(clusters)
		expectedType := "*[]api.MetricCluster"
		if actualType.String() != expectedType {
			t.Fatalf("Expected %s, got %s", expectedType, actualType.String())
		}
	}
}
