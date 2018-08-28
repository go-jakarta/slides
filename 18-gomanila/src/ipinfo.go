package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func main() {
	// read from web
	res, err := http.Get("http://ifconfig.me/all")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}

	// read the body
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}

	// loop over body
	vals := map[string]interface{}{}
	for _, line := range strings.Split(string(body), "\n") {
		v := strings.SplitN(string(line), ":", 2)
		if len(v) > 1 {
			vals[v[0]] = strings.TrimSpace(v[1])
		}
	}

	// encode json
	buf, err := json.MarshalIndent(vals, "", "\t")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}

	os.Stdout.Write(buf)
}
