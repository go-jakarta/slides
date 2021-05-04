package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	redis "gopkg.in/redis.v2"

	fcm "github.com/NaySoftware/go-fcm"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/cskr/pubsub"
	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/s3"
	"github.com/goamz/goamz/sqs"
	"github.com/golang/groupcache"
	"github.com/kenshaw/jwt"
	"github.com/kenshaw/sdhook"
	"github.com/sirupsen/logrus"
	ses "github.com/sourcegraph/go-ses"

	bigquery "github.com/dailyburn/bigquery/client"
)

func myMiddleware(ctxt context.Context) context.Context {
	myLogger := globalLogger.WithFields(map[string]interface{}{
		"field1": reqData,
	})
	return context.WithValue(LoggerKey, myLogger)
}

func myHandler(ctxt context.Context, res http.ResponseWriter, req *http.Request) {
	logger := ctxt.Value(LoggerKey).(Logger)
	logger.Printf("log message")
}

var signer = jwt.RS256.New(keyset)

func createToken(ctxt context.Context) ([]byte, error) {
	return signer.Encode(map[string]interface{}{
		"user_id":      ctxt.Value(UserIDKey),
		"session_data": "something",
	})
}

func redisExample() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)

	err := client.Set("key", "value", 0).Err()
}

func memcacheExample() {
	mc := memcache.New("10.0.0.1:11211", "10.0.0.2:11211", "10.0.0.3:11212")
	mc.Set(&memcache.Item{Key: "foo", Value: []byte("my value")})
	it, err := mc.Get("foo")
}

func groupcacheExample() {
	peers := groupcache.NewHTTPPool("http://127.0.0.1:8080")
	peers.SetBasePath("/cache/")
	getter := groupcache.GetterFunc(func(ctx groupcache.Context, key string, dest groupcache.Sink) error {
		// key == gopher.png
		dest.SetString(strconv.Itoa(0) + ":" + key)
		return nil
	})
	groupcache.NewGroup("thumbnail", 1<<20, getter)
}

func sqlExample() {
	db, err := sql.Open("<sql driver name>", "user:pass@/database")
	stmt, err := db.Prepare("INSERT INTO table SET value = ?...")
	res, err := stmt.Exec(val1 /* ... */)
}

func firebaseExample() {
	//	db, err := firebase.NewDatabaseRef( /* ... */ )
	//	r := db.Ref("path/to/something")
	//	id, err := r.Push(map[string]interface{}{
	//		"my_field": 15,
	//	})
}

func bigqueryExample() {
	client := bigquery.New("/path/to/creds")
	rows, headers, err := client.Query("dataset", "project", "select * from publicdata:samples.shakespeare limit 100;")
}

func sqsExample() {
	conn := sqs.New(aws.Auth{
		AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	}, aws.USEast)

	q, err := conn.CreateQueue(queueName)
	if err != nil {
		log.Fatalf(err.Error())
	}

	// send
	q.SendMessage(batch)

	// receive
	res, err := q.ReceiveMessage(10)
}

func sesExample() {
	from := "notify@sourcegraph.com"
	to := "success@simulator.amazonses.com"

	// uses environment AWS_ACCESS_KEY_ID and AWS_SECRET_KEY
	res, err := ses.EnvConfig.SendEmail(from, to, "Hello, world!", "Here is the message body.")
}

func pubsubExample() {
	ps := pubsub.New(1)

	// pub
	ps.Pub("t1", msg1, msg2 /* ... */)

	// sub
	ch1 := ps.Sub("t1")
	for _, v := range ch1 {
		/* ... */
	}
}

func fcmExample() {
	c := fcm.NewFcmClient("server key")
	c.NewFcmMsgTo(topic, map[string]string{
		"msg": "Hello World1",
		"sum": "Happy Day",
	})

	status, err := c.Send()
	if err == nil {
		status.PrintResults()
	}
}

func stackdriverLogrusExample() {
	logger := logrus.New()
	h, err := sdhook.New(
		sdhook.GoogleServiceAccountCredentialsFile("/path/to/credentials.json"),
	)
	logger.Hooks.Add(h)

	ctxt = context.WithValue(context.Background(), LoggerKey, logger.WithFields(logrus.Fields{
		"field1": "value1",
		"field2": "value2",
	}))

	logger.Printf("logging something")
}

func s3Example() {
	s3client := s3.New(&aws.Config{
		Region:           "",
		Endpoint:         "s3.amazonaws.com",
		S3ForcePathStyle: true,
		Credentials:      creds,
		LogLevel:         0,
	})

	/* ... a lot of other initialization stuff ... ! */

	s3client.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(bucketName),
		Key:           aws.String(path),
		ACL:           aws.String("public-read"),
		Body:          fileBytes,
		ContentLength: aws.Long(size),
		ContentType:   aws.String(fileType),
		Metadata: map[string]*string{
			"Key": aws.String("MetadataValue"),
		},
	})
}
