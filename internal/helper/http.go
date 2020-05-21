package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
)

func HttpPathTrim(path string) string {
	path = regexp.MustCompile(`^/+`).ReplaceAllString(path, "")
	path = regexp.MustCompile(`/+$`).ReplaceAllString(path, "")
	return path
}

func HttpPathID(path string) (id, name, pattern string) {
	rx := regexp.MustCompile(`/{([^:]+):(.*)}$`)
	if tmp := rx.FindAllStringSubmatch(path, -1); tmp != nil && len(tmp) > 0 && len(tmp[0]) == 3 {
		id = tmp[0][0]
		name = tmp[0][1]
		pattern = tmp[0][2]
	}
	return
}

func HttpQuery(method, url string, head map[string]string, body map[string]interface{}) (int, map[string][]string, []byte, error) {
	// create request
	if req, err := http.NewRequest(method, url, nil); err != nil {
		return 0, nil, nil, err
	} else if req != nil {
		// request header
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Pragma", "no-cache")
		if head != nil && len(head) > 0 {
			for k, v := range head {
				req.Header.Set(k, v)
			}
		}
		// request data
		if body != nil && len(body) > 0 {
			data, err := json.Marshal(body)
			if err != nil {
				return 0, nil, nil, err
			}
			req.Body = ioutil.NopCloser(bytes.NewBuffer(data))
		}
		// send
		clt := &http.Client{}
		if resp, err := clt.Do(req); err != nil {
			return 0, nil, nil, err
		} else if resp != nil {
			defer func() {
				_ = resp.Body.Close()
			}()
			if b, err := ioutil.ReadAll(resp.Body); err != nil {
				return resp.StatusCode, resp.Header, nil, err
			} else {
				return resp.StatusCode, resp.Header, b, err
			}
		}
	}
	return 0, nil, nil, fmt.Errorf("missing request body")
}
