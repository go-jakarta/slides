package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	// read from web
	res, err := http.Get("http://ifconfig.co/json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
	defer res.Body.Close()

	// read the body
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}

	// decode json
	var vals map[string]interface{}
	err = json.Unmarshal(body, &vals)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}

	for k, v := range vals {
		fmt.Fprintf(os.Stdout, "%s: %v\n", k, v)
	}
}
