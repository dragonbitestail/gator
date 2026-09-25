package main

import (
	"bytes"
	"context"
	"errors"
  "fmt"
	"html"
	"io"
	//"log"
  "net/http"
	"encoding/xml"
	"golang.org/x/net/html/charset"
	_ "github.com/dragonbitestail/gator/pkg/logging"
)

type requestType string

const (
    reqGetC requestType  = "GET"
    reqPostC requestType = "POST"
)


type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	logr.Debug("fetchFeed()", "feedURL", feedURL)

	rssFeed := RSSFeed{}
	resp, bodyBytes, err := getResponse(ctx, feedURL, reqGetC)
	if err != nil {
		logr.Error("fetchFeed() getResponse returning error", "feedURL", feedURL)
		return nil, err
	}

	// Check resp
	logr.Debug("fetchFeed() eval'ing response", "StatusCode", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		logr.Error("fetchFeed() http Status not OK", "feedURL", feedURL)
		return nil, errors.New("Response not OK: " + resp.Status)
	}

	// Decode XML to RSSFeed
	// https://stackoverflow.com/questions/6002619/unmarshal-an-iso-8859-1-xml-input-in-go
	reader := bytes.NewReader(bodyBytes)
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel
	err = decoder.Decode(&rssFeed)
	if err != nil {
		logr.Error("fetchFeed() xml Unmarshal failed")
		return nil, err
	}

	err = unescapeEntitiesHTML(&rssFeed)
	if err != nil {
		logr.Error("fetchFeed() unescapeEntitiesHTML() returned error", "feedURL", feedURL)
		return nil, err
	}

	return &rssFeed, nil
}

func unescapeEntitiesHTML(rss *RSSFeed) error {
	logr.Debug("unescapeEntitiesHTML()", "rss.Channel.Title", rss.Channel.Title)

  rss.Channel.Title = html.UnescapeString(rss.Channel.Title)
  rss.Channel.Description = html.UnescapeString(rss.Channel.Description)
	for _, rssItem := range rss.Channel.Item {
	  rssItem.Title = html.UnescapeString(rss.Channel.Title)
  	rssItem.Description = html.UnescapeString(rss.Channel.Description)
	}
	return nil
}

func getResponse(ctx context.Context, uri string, rType requestType) (*http.Response, []byte, error) {
    // Create new request:
    req, err := http.NewRequestWithContext(ctx, string(rType), uri, nil)
    if err != nil {
      logr.Error("getResponse(): Returning error creating request.")
      return nil, nil, err
    }

    // Set header on the request:
    req.Header.Set("user-agent", "gator")

    // Make the request using '&'.
		// Ensures you won't accidentally duplicate the client state when passing it around your code.
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
      logr.Error("getResponse(): Returning error making request.")
      return nil, nil, err
    }
    defer resp.Body.Close()

		bodyBytes, err := getResponseBody(resp)
		if err != nil {
			return nil, nil, err
		}
    return resp, bodyBytes, nil
}


func getResponseBody(resp *http.Response) ([]byte, error) {
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		logr.Error("getResponseBody() error reading from response body")
		return nil, err
	}
	return b, nil
}


func printAllRespHeaders(respHeader http.Header){
	fmt.Println("RESPONSE Headers::")
	for field, value := range respHeader {
	    fmt.Printf("%s => %s\n", field, value)
	}
	return
}

func getRespLastModified(respHeader http.Header, headerName string) (string, error) {

  // Read header from response:
  lmHeader := respHeader.Get(headerName)
  fmt.Printf("getLastModified(): last modified: \"%s\"\n", lmHeader)

  // Delete header from response:
  // res.Header.Del("last-modified")

  return lmHeader, nil
} // Body.Close() called from defer to tell server we are done.
