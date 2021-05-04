// +build ignore

package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	geoipURL = "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&suffix=tar.gz&license_key=%s"
)

func main() {
	license := flag.String("license", "", "maxmind license")
	flag.Parse()
	if err := run(context.Background(), *license); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, license string) error {
	if license == "" {
		return errors.New("missing -license")
	}
	urlstr := fmt.Sprintf(geoipURL, license)
	log.Printf("RETRIEVING: %s", urlstr)
	req, err := http.NewRequest("GET", urlstr, nil)
	if err != nil {
		return err
	}
	cl := &http.Client{}
	res, err := cl.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status: %d", res.StatusCode)
	}
	gz, err := gzip.NewReader(res.Body)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	r, w := tar.NewReader(gz), gzip.NewWriter(buf)
loop:
	// process tar
	for {
		h, err := r.Next()
		switch {
		case err == io.EOF:
			break loop
		case err != nil:
			return err
		}
		if !strings.HasSuffix(h.Name, ".mmdb") {
			continue
		}
		if _, err = io.Copy(w, r); err != nil {
			return err
		}
	}
	if err := w.Close(); err != nil {
		return err
	}
	return ioutil.WriteFile("GeoLite2-City.mmdb.gz", buf.Bytes(), 0644)
}
