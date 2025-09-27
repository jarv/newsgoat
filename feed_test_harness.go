package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Lorem Ipsum word pools for generating content
var loremWords = []string{
	"lorem", "ipsum", "dolor", "sit", "amet", "consectetur", "adipiscing", "elit",
	"sed", "do", "eiusmod", "tempor", "incididunt", "ut", "labore", "et", "dolore",
	"magna", "aliqua", "enim", "ad", "minim", "veniam", "quis", "nostrud",
	"exercitation", "ullamco", "laboris", "nisi", "aliquip", "ex", "ea", "commodo",
	"consequat", "duis", "aute", "irure", "in", "reprehenderit", "voluptate",
	"velit", "esse", "cillum", "fugiat", "nulla", "pariatur", "excepteur", "sint",
	"occaecat", "cupidatat", "non", "proident", "sunt", "culpa", "qui", "officia",
	"deserunt", "mollit", "anim", "id", "est", "laborum",
}

var loremTitles = []string{
	"Lorem Ipsum Technology News", "Dolor Sit Tech Blog", "Consectetur Programming Tips",
	"Adipiscing Development Updates", "Sed Do Software Review", "Eiusmod Code Chronicles",
	"Tempor Tech Times", "Incididunt Innovation Hub", "Labore Dev Digest",
	"Dolore Digital Daily", "Magna Tech Magazine", "Aliqua Algorithm Alert",
}

// Generate random lorem ipsum text
func generateLoremText(wordCount int) string {
	if wordCount <= 0 {
		wordCount = 10 + rand.Intn(20) // 10-30 words
	}

	words := make([]string, wordCount)
	for i := 0; i < wordCount; i++ {
		words[i] = loremWords[rand.Intn(len(loremWords))]
	}

	text := strings.Join(words, " ")
	// Capitalize first letter
	if len(text) > 0 {
		text = strings.ToUpper(string(text[0])) + text[1:]
	}

	return text + "."
}

// Generate a dummy RSS feed
func generateDummyFeed(title string, articleCount int) string {
	if title == "" {
		title = loremTitles[rand.Intn(len(loremTitles))]
	}

	if articleCount <= 0 {
		articleCount = 1 + rand.Intn(100) // 1-100 articles
	}

	// Start RSS feed
	rss := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
<channel>
<title>%s</title>
<description>%s</description>
<link>http://example.com</link>
<language>en-us</language>
<lastBuildDate>%s</lastBuildDate>
<generator>NewsGoat Feed Test Harness</generator>
`, title, generateLoremText(15), time.Now().Format(time.RFC1123))

	// Add articles
	for i := 0; i < articleCount; i++ {
		articleTitle := generateLoremText(3 + rand.Intn(7)) // 3-10 words
		articleTitle = strings.TrimSuffix(articleTitle, ".") // Remove period from titles

		description := generateLoremText(20 + rand.Intn(30)) // 20-50 words
		content := generateLoremText(50 + rand.Intn(100))    // 50-150 words

		// Random publish date within last 30 days
		publishDate := time.Now().AddDate(0, 0, -rand.Intn(30))

		guid := fmt.Sprintf("http://example.com/article/%d", i+1)

		rss += fmt.Sprintf(`
<item>
<title>%s</title>
<description>%s</description>
<content:encoded><![CDATA[%s]]></content:encoded>
<link>%s</link>
<guid>%s</guid>
<pubDate>%s</pubDate>
</item>`, articleTitle, description, content, guid, guid, publishDate.Format(time.RFC1123))
	}

	rss += `
</channel>
</rss>`

	return rss
}

func feedHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	// Parse status parameter (default: 200)
	status := 200
	if statusParam := query.Get("status"); statusParam != "" {
		if parsedStatus, err := strconv.Atoi(statusParam); err == nil {
			status = parsedStatus
		}
	}

	// Parse delay parameter (default: random 0-5 seconds)
	var delay time.Duration
	if delayParam := query.Get("delay"); delayParam != "" {
		if delaySeconds, err := strconv.Atoi(delayParam); err == nil && delaySeconds >= 0 {
			delay = time.Duration(delaySeconds) * time.Second
		}
	} else {
		// Random delay between 0 and 5 seconds
		delay = time.Duration(rand.Intn(6)) * time.Second
	}

	// Parse title parameter
	title := query.Get("title")
	if title == "" {
		title = loremTitles[rand.Intn(len(loremTitles))]
	}

	// Parse articles parameter (default: random 1-100)
	articleCount := 0
	if articlesParam := query.Get("articles"); articlesParam != "" {
		if parsedCount, err := strconv.Atoi(articlesParam); err == nil && parsedCount > 0 {
			articleCount = parsedCount
		}
	}

	// Check for conditional request headers
	ifNoneMatch := r.Header.Get("If-None-Match")
	ifModifiedSince := r.Header.Get("If-Modified-Since")

	requestType := "UNCONDITIONAL"
	if ifNoneMatch != "" || ifModifiedSince != "" {
		requestType = "CONDITIONAL"
	}

	// Log the incoming request
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📥 REQUEST: %s %s\n", requestType, r.URL.Path)
	fmt.Printf("   Feed Title: %s\n", title)
	fmt.Printf("   User-Agent: %s\n", r.Header.Get("User-Agent"))
	if ifNoneMatch != "" {
		fmt.Printf("   If-None-Match: %s\n", ifNoneMatch)
	}
	if ifModifiedSince != "" {
		fmt.Printf("   If-Modified-Since: %s\n", ifModifiedSince)
	}

	// Apply delay
	if delay > 0 {
		fmt.Printf("   ⏱️  Applying delay: %v\n", delay)
		time.Sleep(delay)
	}

	// Generate ETag based on current time (changes every second)
	now := time.Now()
	etag := fmt.Sprintf(`"etag-%d"`, now.Unix())
	lastModified := now.UTC().Format(http.TimeFormat)

	// Check if content matches (for 304 response)
	// If client sent If-None-Match and it matches current ETag, return 304
	if ifNoneMatch != "" && ifNoneMatch == etag {
		fmt.Printf("📤 RESPONSE: 304 Not Modified\n")
		fmt.Printf("   ETag: %s\n", etag)
		fmt.Printf("   Last-Modified: %s\n", lastModified)
		fmt.Println("   ✨ Conditional request matched - no content sent!")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		w.Header().Set("ETag", etag)
		w.Header().Set("Last-Modified", lastModified)
		w.Header().Set("Cache-Control", "max-age=3600")
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// Generate feed content
	feedContent := generateDummyFeed(title, articleCount)

	// Set response headers
	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.Header().Set("ETag", etag)
	w.Header().Set("Last-Modified", lastModified)
	w.Header().Set("Cache-Control", "max-age=3600")

	// Log the response
	fmt.Printf("📤 RESPONSE: %d OK\n", status)
	fmt.Printf("   ETag: %s\n", etag)
	fmt.Printf("   Last-Modified: %s\n", lastModified)
	fmt.Printf("   Cache-Control: max-age=3600\n")
	fmt.Printf("   Articles: %d\n", articleCount)
	fmt.Printf("   Content-Length: %d bytes\n", len(feedContent))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	w.WriteHeader(status)

	// Write response
	if status >= 200 && status < 300 {
		if _, err := w.Write([]byte(feedContent)); err != nil {
			fmt.Printf("❌ Error writing feed content: %v\n", err)
		}
	} else {
		// For error status codes, write a simple error message
		if _, err := fmt.Fprintf(w, "HTTP %d Error", status); err != nil {
			fmt.Printf("❌ Error writing error response: %v\n", err)
		}
	}
}

func runFeedTestHarness() error {
	// Initialize random number generator (Go 1.20+ automatically seeds)

	port := ":8080"

	http.HandleFunc("/", feedHandler)
	http.HandleFunc("/feed.xml", feedHandler)
	http.HandleFunc("/rss.xml", feedHandler)
	http.HandleFunc("/feed", feedHandler)

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🐐 NewsGoat Feed Test Harness")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("   Listening on: http://localhost%s\n", port)
	fmt.Println()
	fmt.Println("✨ Features:")
	fmt.Println("   • Supports conditional requests (ETag, Last-Modified)")
	fmt.Println("   • Returns 304 Not Modified when ETag matches")
	fmt.Println("   • Sets Cache-Control: max-age=3600")
	fmt.Println("   • Logs request type (CONDITIONAL vs UNCONDITIONAL)")
	fmt.Println()
	fmt.Println("📖 Example URLs:")
	fmt.Printf("   http://localhost%s/feed.xml\n", port)
	fmt.Printf("   http://localhost%s/feed.xml?title=Tech+News&articles=10\n", port)
	fmt.Printf("   http://localhost%s/feed.xml?delay=5&articles=3\n", port)
	fmt.Printf("   http://localhost%s/feed.xml?status=500\n", port)
	fmt.Println()
	fmt.Println("💡 Testing conditional requests:")
	fmt.Println("   1. Add the feed URL to NewsGoat")
	fmt.Println("   2. Refresh once (UNCONDITIONAL - stores ETag)")
	fmt.Println("   3. Refresh again within same second (304 Not Modified)")
	fmt.Println("   4. Wait 1+ second and refresh (200 OK - new ETag)")
	fmt.Println()
	fmt.Println("🔧 Query Parameters:")
	fmt.Println("   title=...    Feed title (default: random)")
	fmt.Println("   articles=N   Number of articles (default: random 1-100)")
	fmt.Println("   delay=N      Response delay in seconds (default: random 0-5)")
	fmt.Println("   status=N     HTTP status code (default: 200)")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	return http.ListenAndServe(port, nil)
}