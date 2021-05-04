// Command server is the realtime server visualization component.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/go-jakarta/slides/32-3dviz/geoip"
	"github.com/kenshaw/goji"
	"github.com/uber/h3-go/v3"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:3000", "listen addr")
	ip := flag.String("defip", "123.45.67.89", "default ip address")
	res := flag.Int("res", 4, "h3 resolution")
	queue := flag.String("queue", "3dviz", "sqs queue name")
	flag.Parse()
	if err := run(*addr, *ip, *res, *queue); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(addr, ip string, res int, queue string) error {
	s, err := newServer(ip, res, queue)
	if err != nil {
		return err
	}
	return http.ListenAndServe(addr, s)
}

type Server struct {
	*goji.Mux
	ip   net.IP
	res  int
	g    *geoip.Geoip
	sess *session.Session
	svc  *sqs.SQS
	qurl string
}

func newServer(ip string, res int, queue string) (*Server, error) {
	g, err := geoip.New()
	if err != nil {
		return nil, err
	}
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	svc := sqs.New(sess)
	urlres, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &queue,
	})
	if err != nil {
		return nil, err
	}
	log.Printf("queue: %s", *urlres.QueueUrl)
	s := &Server{
		Mux:  goji.New(),
		ip:   net.ParseIP(ip),
		res:  res,
		g:    g,
		sess: sess,
		svc:  svc,
		qurl: *urlres.QueueUrl,
	}
	s.HandleFunc(goji.Get("/*"), s.index)
	return s, nil
}

func (s *Server) index(res http.ResponseWriter, req *http.Request) {
	remote := req.RemoteAddr
	if s := req.URL.Query().Get("remote"); s != "" {
		remote = s
	}
	ip, err := s.resolveIP(remote)
	if err != nil {
		errf(res, "unknown remote: %v", err)
		return
	}
	country, lat, lon, err := s.g.Lookup(ip)
	if err != nil {
		errf(res, "unable to geolocate %s: %v", ip, err)
		return
	}
	hex := fmt.Sprintf("%x", h3.FromGeo(h3.GeoCoord{
		Latitude:  lat,
		Longitude: lon,
	}, s.res))
	msg := Message{
		"country": country,
		"lat":     lat,
		"lon":     lon,
		"hex":     hex,
		"path":    req.URL.Path,
	}
	go s.queue(msg)
	fmt.Fprintf(res, "%s", msg)
}

func (s *Server) resolveIP(remote string) (net.IP, error) {
	host := remote
	if i := strings.LastIndex(host, ":"); i != -1 {
		var err error
		if host, _, err = net.SplitHostPort(remote); err != nil {
			return net.IP{}, err
		}
	}
	ip := net.ParseIP(host)
	if host == "127.0.0.1" {
		copy(ip, s.ip)
	}
	return ip, nil
}

func (s *Server) queue(msg Message) error {
	_, err := s.svc.SendMessage(&sqs.SendMessageInput{
		QueueUrl:    &s.qurl,
		MessageBody: aws.String(msg.String()),
	})
	if err != nil {
		return qerrf("unable to send message: %v", err)
	}
	return nil
}

type Message map[string]interface{}

func (msg Message) String() string {
	buf, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return string(buf)
}

func errf(res http.ResponseWriter, s string, v ...interface{}) {
	err := fmt.Sprintf(s, v...)
	log.Print("http error:", err)
	http.Error(res, err, http.StatusInternalServerError)
}

func qerrf(s string, v ...interface{}) error {
	err := fmt.Errorf(s, v...)
	log.Print("sqs error:", err)
	return err
}
