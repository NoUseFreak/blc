package main

import (
    "os"
    "flag"
    "fmt"
    "net/http"
    "golang.org/x/net/html"
    "net/url"
    "strings"
    "bytes"
)

type OutputLogger struct {
    Verbosity string
}
func (ol OutputLogger) Output(parts ...string) {
    var buffer bytes.Buffer

    for _, part := range parts {
        buffer.WriteString(part)
        buffer.WriteString(" ")
    }

    fmt.Println(buffer.String())
}
func (ol OutputLogger) Debug(parts ...string) {
    if ol.Verbosity == "DEBUG" {
        ol.Output(parts...)
    }
}
func (ol OutputLogger) Error(parts ...string) {
    ol.Output(parts...)
}

var visited = make(map[string]bool) 
var baseUrl string
var logger OutputLogger

func main() {
    verbose := flag.Bool("v", false, "Be more verbose")
    flag.Parse()         
    args := flag.Args()
    logger = OutputLogger{}

    if *verbose {
        logger.Verbosity = "DEBUG"
    }   

    if len(args) < 1 {   
        logger.Debug("Please specify start page")  
        os.Exit(1)                                
    }
    queue := make(chan string)
    baseUrl = args[0]

    logger.Output("Checking links on", baseUrl)

    go func() {            
        queue <- baseUrl
    }()

    for uri := range queue {    
        enqueue(uri, queue)  
    }
    close(queue)

    logger.Output("Done")

    os.Exit(0)
}

/**
 * Enqueue a page
 */
func enqueue(url string, queue chan string) {
    if visited[url] {
        return
    }

    logger.Debug("fetching", url)
    links := retrieveLinks(url)
    visited[url] = true
    for _, link := range links {
        absolute := fixUrl(link, url) 
        if url == "" {
            continue
        }
        if visited[absolute] {
            continue
        }
        if !strings.Contains(absolute, baseUrl) {
            // continue
        }

        go func() { 
            queue <- absolute 
        }() 
    }
}

/**
 * Download a given url and return all urls found on that page.
 */
func retrieveLinks(url string) []string {  
    resp, err := http.Get(url)
    links := make([]string, 0)
    if err != nil {            
        logger.Error("Detected broken url", url)
        return links
    }
    defer resp.Body.Close()  

    page := html.NewTokenizer(resp.Body)
    for {
        tokenType := page.Next()
        if tokenType == html.ErrorToken {
            return links
        }
        token := page.Token()
        if tokenType == html.StartTagToken && token.DataAtom.String() == "a" {
            for _, attr := range token.Attr {
                if attr.Key == "href" {
                    links = append(links, attr.Val)
                }
            }
        }
    }
}


func fixUrl(href, base string) (string) {  
    uri, err := url.Parse(href)              
    if err != nil {                          
        return ""                              
    }                                        
    baseUrl, err := url.Parse(base)          
    if err != nil {                          
        return ""
    }
    uri = baseUrl.ResolveReference(uri)
    
    return uri.String()                      
}