package main

import (
	"context"
	"log"

	"github.com/chromedp/chromedp"
)

func main() {
	// create chrome instance
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	// run tasks
	var res string
	err := chromedp.Run(ctx, googleSearch("site:brank.as", &res))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("first search result: %s", res)
}

func googleSearch(q string, res *string) chromedp.Tasks {
	if res == nil {
		panic("res cannot be nil")
	}
	return chromedp.Tasks{
		chromedp.Navigate(`https://www.google.com`),
		chromedp.WaitVisible(`#hplogo`, chromedp.ByID),
		chromedp.SendKeys(`#lst-ib`, q+"\n", chromedp.ByID),
		chromedp.WaitVisible(`#res`, chromedp.ByID),
		chromedp.Text(`#res div.rc:nth-child(1)`, res),
	}
}
