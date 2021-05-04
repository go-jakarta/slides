package geoip

//go:generate go run gen.go

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"io/ioutil"
	"net"
	"strings"

	maxminddb "github.com/oschwald/maxminddb-golang"
)

//go:embed GeoLite2-City.mmdb.gz
var GeoLite2CityMmdbGz []byte

// Geoip is a geoip util.
type Geoip struct {
	db *maxminddb.Reader
}

// New creates a new geoip util.
func New() (*Geoip, error) {
	r, err := gzip.NewReader(bytes.NewReader(GeoLite2CityMmdbGz))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	db, err := maxminddb.FromBytes(buf)
	if err != nil {
		return nil, err
	}
	return &Geoip{
		db: db,
	}, nil
}

func (g *Geoip) Lookup(ip net.IP) (string, float64, float64, error) {
	var root struct {
		Country struct {
			ISOCode string `maxminddb:"iso_code"`
		} `maxminddb:"country"`
		Location struct {
			Latitude  float64 `maxminddb:"latitude"`
			Longitude float64 `maxminddb:"longitude"`
		} `maxminddb:"location"`
	}
	if err := g.db.Lookup(ip, &root); err != nil {
		return "", 0.0, 0.0, err
	}
	return strings.ToLower(root.Country.ISOCode), root.Location.Latitude, root.Location.Longitude, nil
}
