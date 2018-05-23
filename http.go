package main

import (
	"bytes"
	"fmt"
	"net/http"
)

func readBody(resp *http.Response) (body []byte) {
	var buf *bytes.Buffer
	if resp.ContentLength > 0 {
		buf = bytes.NewBuffer(make([]byte, 0, resp.ContentLength))
	} else {
		buf = bytes.NewBuffer(make([]byte, 100))
	}

	n, err := buf.ReadFrom(resp.Body)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	body = buf.Bytes()[int64(buf.Len())-n:]

	return
}
