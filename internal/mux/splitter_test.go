package mux

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
	"time"
)

const (
	host string = "127.0.0.1"
	port int    = 33480
)

func TestSplitter(t *testing.T) {
	wg := sync.WaitGroup{}
	splitter := NewSplitter()
	if err := splitter.NewRoute([]string{http.MethodPost}, "/v1/test", runnable); err != nil {
		t.Error(err)
	}
	if err := splitter.NewRoute([]string{http.MethodPost}, "/v2/sql/{table:[A-Za-z_0-9]+}/search", runnable); err != nil {
		t.Error(err)
	}
	if err := splitter.NewRoute([]string{http.MethodPost}, "/v3/file/{name:[A-Za-z]+[-][A-Za-z0-9]+[.][A-Za-z0-9]{3}}", runnable); err != nil {
		t.Error(err)
	}
	if err := splitter.NewRoute([]string{http.MethodPost}, "/v4/{table:[A-Za-z0-9]+}/{id:[A-Za-z0-9]+}", runnable); err != nil {
		t.Error(err)
	}
	if err := splitter.NewRoute([]string{http.MethodGet}, "/v5/user/", runnableList); err != nil {
		t.Error(err)
	}
	if err := splitter.NewRoute([]string{http.MethodGet}, "/v5/user/{id:[0-9]+}/", runnableView); err != nil {
		t.Error(err)
	}
	if err := splitter.NewRoute([]string{http.MethodPost}, "/v5/user/", runnableCreate); err != nil {
		t.Error(err)
	}
	if err := splitter.NewRoute([]string{http.MethodPut}, "/v5/user/{id:[0-9]+}/", runnableUpdate); err != nil {
		t.Error(err)
	}
	if err := splitter.NewRoute([]string{http.MethodDelete}, "/v5/user/{id:[0-9]+}/", runnableDelete); err != nil {
		t.Error(err)
	}
	// server
	server := http.Server{}
	server.Handler = splitter
	server.Addr = fmt.Sprintf("%s:%d", "", port)
	server.WriteTimeout = 24 * time.Hour
	server.ReadTimeout = 24 * time.Hour
	go func() {
		wg.Add(1)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			t.Errorf("mux[server]: %s\n", err.Error())
		}
		wg.Done()
	}()
	defer func() {
		if err := server.Shutdown(context.Background()); err != nil {
			t.Fatalf("mux[server]: %s\n", err.Error())
		}
		wg.Wait()
	}()
	time.Sleep(500 * time.Millisecond)
	// check url
	if code, _, err := httpQueryToMap(http.MethodPost, fmt.Sprintf("http://%s:%d/%s", host, port, "v1/test")); err != nil {
		t.Error(err)
	} else if code != http.StatusAccepted {
		t.Errorf("mux[url-path]: wrong response status (%d)", code)
	} else if code != http.StatusAccepted {
		t.Errorf("mux[url-path]: wrong response status (%d)", code)
	} else if code, _, err := httpQueryToMap(http.MethodPost, fmt.Sprintf("http://%s:%d/%s", host, port, "v1/test/")); err != nil {
		t.Error(err)
	} else if code != http.StatusAccepted {
		t.Errorf("mux[url-path]: wrong response status (%d)", code)
	}
	// check url parameters
	if code, body, err := httpQueryToMap(http.MethodPost, fmt.Sprintf("http://%s:%d/%s", host, port, "v2/sql/users/search")); err != nil {
		t.Error(err)
	} else if code != http.StatusAccepted {
		t.Errorf("mux[url-parameter]: wrong response status (%d)", code)
	} else if body == nil || len(body) == 0 {
		t.Errorf("mux[url-parameter]: wrong response data (%q)", body)
	} else if arg, ok := body["parameters"]; !ok || arg == nil || len(arg) == 0 {
		t.Errorf("mux[url-parameter]: wrong field 'parameters' (%q)", body)
	} else if v, ok := arg["table"]; !ok || v != "users" {
		t.Errorf("mux[url-parameter]: wrong field 'parameters' (%q)", arg)
	}
	// check url id
	if code, body, err := httpQueryToMap(http.MethodPost, fmt.Sprintf("http://%s:%d/%s", host, port, "v3/file/sound-001.mp3")); err != nil {
		t.Error(err)
	} else if code != http.StatusAccepted {
		t.Errorf("mux[url-id]: wrong response status (%d)", code)
	} else if body == nil || len(body) == 0 {
		t.Errorf("mux[url-id]: wrong response data (%q)", body)
	} else if arg, ok := body["id"]; !ok || arg == nil {
		t.Errorf("mux[url-id]: wrong field `id` (%q)", body)
	} else if n, ok := arg["name"]; !ok || n != "name" {
		t.Errorf("mux[url-id]: wrong field `name` (%q)", arg)
	} else if v, ok := arg["value"]; !ok || v != "sound-001.mp3" {
		t.Errorf("mux[url-id]: wrong field `value` (%q)", arg)
	} else if arg, ok := body["parameters"]; !ok || arg == nil || len(arg) != 0 {
		t.Errorf("mux[url-id]: wrong field 'parameters' (%q)", body)
	} else if code, _, err := httpQueryToMap(http.MethodPost, fmt.Sprintf("http://%s:%d/%s", host, port, "v3/file/sound-001.mp")); err != nil {
		t.Error(err)
	} else if code != http.StatusNotFound {
		t.Errorf("mux[url-id]: wrong response status (%d)", code)
	} else if code, _, err := httpQueryToMap(http.MethodPost, fmt.Sprintf("http://%s:%d/%s", host, port, "v3/file/sound.mp3")); err != nil {
		t.Error(err)
	} else if code != http.StatusNotFound {
		t.Errorf("mux[url-id]: wrong response status (%d)", code)
	} else if code, _, err := httpQueryToMap(http.MethodPost, fmt.Sprintf("http://%s:%d/%s", host, port, "v3/file/sound-001mp3")); err != nil {
		t.Error(err)
	} else if code != http.StatusNotFound {
		t.Errorf("mux[url-id]: wrong response status (%d)", code)
	} else if code, _, err := httpQueryToMap(http.MethodPost, fmt.Sprintf("http://%s:%d/%s", host, port, "v3/file")); err != nil {
		t.Error(err)
	} else if code != http.StatusNotFound {
		t.Errorf("mux[url-id]: wrong response status (%d)", code)
	} else if code, _, err := httpQueryToMap(http.MethodPost, fmt.Sprintf("http://%s:%d/%s", host, port, "v3/file/ sound-001.mp3")); err != nil {
		t.Error(err)
	} else if code != http.StatusNotFound {
		t.Errorf("mux[url-id]: wrong response status (%d)", code)
	}
	// check url id and parameters
	if code, body, err := httpQueryToMap(http.MethodPost, fmt.Sprintf("http://%s:%d/%s", host, port, "v4/session/123")); err != nil {
		t.Error(err)
	} else if code != http.StatusAccepted {
		t.Errorf("mux[url-id-and-parameters]: wrong response status (%d)", code)
	} else if body == nil || len(body) == 0 {
		t.Errorf("mux[url-id-and-parameters]: wrong response data (%q)", body)
	} else if arg, ok := body["id"]; !ok || arg == nil {
		t.Errorf("mux[url-id-and-parameters]: wrong field `id` (%q)", body)
	} else if n, ok := arg["name"]; !ok || n != "id" {
		t.Errorf("mux[url-id-and-parameters]: wrong field `name` (%q)", arg)
	} else if v, ok := arg["value"]; !ok || v != "123" {
		t.Errorf("mux[url-id-and-parameters]: wrong field `value` (%q)", arg)
	} else if arg, ok := body["parameters"]; !ok || arg == nil || len(arg) == 0 {
		t.Errorf("mux[url-id-and-parameters]: wrong field 'parameters' (%q)", body)
	} else if v, ok := arg["table"]; !ok || v != "session" {
		t.Errorf("mux[url-id-and-parameters]: wrong field 'table' (%q)", arg)
	}
	// check rest url
	if code, body, err := httpQuery(http.MethodGet, fmt.Sprintf("http://%s:%d/%s?test=1", host, port, "v5/user")); err != nil {
		t.Error(err)
	} else if code != http.StatusOK {
		t.Errorf("mux[rest-list]: wrong response status (%d)", code)
	} else if string(body) != "list" {
		t.Errorf("mux[rest-list]: wrong response body (%s)", string(body))
	}
	if code, body, err := httpQuery(http.MethodGet, fmt.Sprintf("http://%s:%d/%s?test=2", host, port, "v5/user/1")); err != nil {
		t.Error(err)
	} else if code != http.StatusOK {
		t.Errorf("mux[rest-view]: wrong response status (%d)", code)
	} else if string(body) != "view" {
		t.Errorf("mux[rest-view]: wrong response body (%s)", string(body))
	}
	if code, body, err := httpQuery(http.MethodPost, fmt.Sprintf("http://%s:%d/%s?test=3", host, port, "v5/user")); err != nil {
		t.Error(err)
	} else if code != http.StatusOK {
		t.Errorf("mux[rest-create]: wrong response status (%d)", code)
	} else if string(body) != "create" {
		t.Errorf("mux[rest-create]: wrong response body (%s)", string(body))
	}
	if code, body, err := httpQuery(http.MethodPut, fmt.Sprintf("http://%s:%d/%s?test=4", host, port, "v5/user/2")); err != nil {
		t.Error(err)
	} else if code != http.StatusOK {
		t.Errorf("mux[rest-update]: wrong response status (%d)", code)
	} else if string(body) != "update" {
		t.Errorf("mux[rest-update]: wrong response body (%s)", string(body))
	}
	if code, body, err := httpQuery(http.MethodDelete, fmt.Sprintf("http://%s:%d/%s?test=5", host, port, "v5/user/3")); err != nil {
		t.Error(err)
	} else if code != http.StatusOK {
		t.Errorf("mux[rest-delete]: wrong response status (%d)", code)
	} else if string(body) != "delete" {
		t.Errorf("mux[rest-delete]: wrong response body (%s)", string(body))
	}
}

func runnable(w http.ResponseWriter, r *Request) {
	w.Header().Set("Time", time.Now().Format("20101011"))
	d := struct {
		Parameters map[string]string   `json:"parameters"`
		ID         map[string]string   `json:"id"`
		Query      map[string][]string `json:"query"`
	}{
		Parameters: r.URL.parameters,
		ID: map[string]string{
			"value": string(r.URL.ID.Value),
			"name":  r.URL.ID.Name,
		},
		Query: r.URL.Query(),
	}
	b, err := json.Marshal(d)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write(b)
}

func runnableList(w http.ResponseWriter, _ *Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("list"))
}

func runnableView(w http.ResponseWriter, _ *Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("view"))
}

func runnableCreate(w http.ResponseWriter, _ *Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("create"))
}

func runnableUpdate(w http.ResponseWriter, _ *Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("update"))
}

func runnableDelete(w http.ResponseWriter, _ *Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("delete"))
}

func httpQueryToMap(method, url string) (int, map[string]map[string]string, error) {
	var data map[string]map[string]string = nil
	if code, body, err := httpQuery(method, url); err != nil || len(body) == 0 {
		return code, data, err
	} else if err = json.Unmarshal(body, &data); err != nil {
		return code, nil, fmt.Errorf("%s in '%s'", err.Error(), string(body))
	} else if data != nil {
		return code, data, err
	}
	return 0, nil, fmt.Errorf("missing request body")
}

func httpQuery(method, url string) (int, []byte, error) {
	// create request
	if req, err := http.NewRequest(method, url, nil); err != nil {
		return 0, nil, err
	} else if req != nil {
		// request header
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Pragma", "no-cache")
		// send
		clt := &http.Client{}
		if resp, err := clt.Do(req); err != nil {
			return 0, nil, err
		} else if resp != nil {
			defer func() {
				_ = resp.Body.Close()
			}()
			if b, err := ioutil.ReadAll(resp.Body); err != nil {
				return resp.StatusCode, nil, err
			} else if len(b) == 0 {
				return resp.StatusCode, nil, nil
			} else {
				return resp.StatusCode, b, err
			}
		}
	}
	return 0, nil, fmt.Errorf("missing request body")
}
