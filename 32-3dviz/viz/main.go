package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/uber/h3-go/v3"
)

func main() {
	exp := flag.Duration("exp", 10*time.Minute, "expire duration")
	queue := flag.String("queue", "3dviz", "sqs queue name")
	res := flag.Int("res", 2, "resolution")
	flag.Parse()
	if err := run(context.Background(), *exp, *queue, *res); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, exp time.Duration, queue string, res int) error {
	v, err := newViz(exp, queue, res)
	if err != nil {
		return err
	}
	go v.poll(ctx)
	return newApp(v).run(ctx)
}

type Viz struct {
	exp      time.Duration
	sess     *session.Session
	svc      *sqs.SQS
	qurl     string
	res      int
	messages []*Message
	sync.RWMutex
}

func newViz(exp time.Duration, queue string, res int) (*Viz, error) {
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
	return &Viz{
		exp:  exp,
		sess: sess,
		svc:  svc,
		qurl: *urlres.QueueUrl,
		res:  res,
	}, nil
}

func (v *Viz) poll(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil && err != context.Canceled {
				return errf("context done: %v", err)
			}
			return nil
		case msg := <-v.read(ctx):
			if msg == nil {
				continue
			}
			msg.Expires = time.Now().Add(v.exp)
			msg.Parent = h3.ToParent(h3.FromString(msg.Hex), v.res)
			if err := v.add(msg); err != nil {
				return errf("unable to add: %v", err)
			}
		}
	}
}

func (v *Viz) read(ctx context.Context) <-chan *Message {
	ch := make(chan *Message, 1)
	go func() {
		defer close(ch)
		res, err := v.svc.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl: &v.qurl,
		})
		if err != nil {
			return
		}
		for _, z := range res.Messages {
			msg := new(Message)
			if err := json.Unmarshal([]byte(*z.Body), msg); err != nil {
				return
			}
			select {
			case <-ctx.Done():
				return
			case ch <- msg:
			}
		}
	}()
	return ch
}

func (v *Viz) add(msg *Message) error {
	v.Lock()
	defer v.Unlock()
	now := time.Now()
	var messages []*Message
	for _, m := range v.messages {
		if now.After(m.Expires) {
			continue
		}
		messages = append(messages, m)
	}
	v.messages = append(messages, msg)
	return nil
}

type Message struct {
	Expires time.Time  `json:"-"`
	Country string     `json:"country"`
	Lat     float64    `json:"lat"`
	Lon     float64    `json:"lon"`
	Hex     string     `json:"hex"`
	Parent  h3.H3Index `json:"-"`
	Path    string     `json:"path"`
}

func errf(s string, v ...interface{}) error {
	err := fmt.Errorf(s, v...)
	log.Printf("error: %v", err)
	return err
}
