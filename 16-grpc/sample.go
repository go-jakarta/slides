package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	http.HandleFunc("/api/echo", func(res http.ResponseWriter, req *http.Request) {
		buf, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(res, "invalid body", 500)
			return
		}
		defer req.Body.Close()

		var d struct {
			Msg string `json:"msg"`
		}
		err = json.Unmarshal(buf, &d)
		if err != nil {
			http.Error(res, "invalid body", 500)
			return
		}
		fmt.Fprintf(res, `{ "msg": "%s" }`, d.Msg)
	})
	http.ListenAndServe(":8080", nil)
}
