package main

import (
	"fmt"

	"strings"

	"golang.org/x/net/html"

	"net/http"
	"os"
)

func getHref(t html.Token) (ok bool, href string) {
	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
			ok = true
		}
	}
	return
}

func crawl(url string, ch chan string, chFinished chan bool) {
	// when we are done with crawl funciton then the func() [in front of the defer keyword] will run

	resp, err := http.Get(url)

	// defer is the key word to use when we want a funcaiton to run at the end of a particulare function
	defer func() {
		//we want to puclish true to say that we are done proccessig with url
		chFinished <- true
	}()

	if err != nil {
		fmt.Println("error : failed to crawl : ", url)
		return
	}
	// it has a lot of html tags
	b := resp.Body

	defer b.Close()
	// it will convert the body into a lot of tokens
	z := html.NewTokenizer(b)

	for {
		// we are going over those tokens oneby one
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			return
		case tt == html.StartTagToken:
			// if its a start token we take that value and put it into t
			t := z.Token()
			// its a variable that becomes true if the data is equal to a
			isAnchor := t.Data == "a"
			if !isAnchor {
				continue
			}
			ok, url := getHref(t)
			if !ok {
				continue
			}
			// if we have recieved the url we check that if it starts with a http

			hasProto := strings.Index(url, "http") == 0

			if hasProto {
				// if it has http then we publish it to the channel
				ch <- url
			}
		}
	}
}

func main() {
	// we want to parse multiple different urls to scrap , as we are going through these urls we are going to scrape them we are going to find uniqiue urls in those pages and we want to store it in those found urls maps
	// the bollean is for checking if we found the urls or not
	//basically we want to find unique urls
	foundUrls := make(map[string]bool)
	// this variable is going to stire the args we are passing from the command line
	seedUrls := os.Args[1:]

	// in this channel we are going to output out all the different urls that we find , all those urls that need to be scraped
	chUrls := make(chan string)
	// where we are going to say true in the sense that we are done , finished with this page
	chFinished := make(chan bool)

	//goroutines are basically routines that are called fucntion , multiple proceesess that call other functions
	// for each of those urls we want to start a new goroutine

	// so we range ovre all the urls that the user will send them
	for _, url := range seedUrls {
		// every routine takes the url that needs to be scraped using the crawl function
		// it takes the chaneel url which takes all the unique urls which we will find on this particulare url that the user has given us
		// then we pass the finishe dchannel so we can tell when we are finished with this with this goroutine
		go crawl(url, chUrls, chFinished)
	}

	// now we need to subscribe to those channel how ? we need to write the select statement

	// first we go over those urls
	for c := 0; c < len(seedUrls); {
		// since we have 2 different channels we have 2 cases
		select {
		// when we want to get a value from a channel we write like this
		// the value that we get from this channel will be output into url
		// whenever we get a response from a channel we want to perform an action
		// we dont want the channel urls to stay blocked  till the other channel output something we want both of them to be completely concurrent completely indepemdent of each other
		case url := <-chUrls:
			// we are going to store the found url in this map that we have already created
			foundUrls[url] = true
			// and if the other channel is outputing something else we want to do something else
			// if this channel has a response we are going to the next channel
		case <-chFinished:
			c++
		}

		// then we are going to print out all the urls that we have found
		fmt.Println("\nFound", len(foundUrls), "unique urls:  \n ")

		for url, _ := range foundUrls {
			fmt.Println("-" + url)
		}
		// you have to close channels to remove deadlocks
		close(chUrls)
	}
}
