package main

import "fmt"
import "github.com/mmcdole/gofeed"
//import "github.com/SlyMarbo/rss"
import "github.com/gorilla/feeds"
import "sort"
import "time"
import "net/http"
import "log"

const README_URL = "https://raw.githubusercontent.com/chase-seibert/engineering-manager-blogs/master/README.md"

func getUrls(baseUrl string) []string {
	return []string{
		"https://ibenstewart.com/feed",
		"https://danielrichnak.com/feed",
		"https://chase-seibert.github.io/blog/feed.xml",
		"https://chelseatroy.com/feed/",
		"https://medium.com/feed/dakshp",
		"https://www.leadsv.com/insight/?format=rss",
		"http://www.kendallmiller.co/kendall-miller-blog?format=RSS",
		"https://matthewnewkirk.com/feed/",
		"http://randsinrepose.com/feed/",
		"https://medium.com/feed/@royrapoport",
		"https://introvertedengineer.com/feed",
	}
}

func fetchUrl(url string, ch chan<-*gofeed.Feed) {
  fmt.Printf("%v\n", url)
  fp := gofeed.NewParser()
  feed, _ := fp.ParseURL(url)
  // TODO: https://github.com/mmcdole/gofeed/issues/83
  // See: https://github.com/mmcdole/gofeed#default-mappings
  //feed, _ := rss.Fetch(url)
  //fmt.Println(feed.Items[0].PublishedParsed)
  //fmt.Println(feed.String())
  ch <- feed
}

func fetchUrls(urls []string) []*gofeed.Feed {
  allFeeds := make([]*gofeed.Feed, 0)
  ch := make(chan *gofeed.Feed)
  for _, url := range urls {
    go fetchUrl(url, ch)
  }
  for range urls {
    feed := <- ch
    allFeeds = append(allFeeds, feed)
  }
  //fmt.Printf("%#v", allFeeds)
  return allFeeds
}

// TODO: there must be a shorter syntax for this
type byPublished []*gofeed.Feed

func (s byPublished) Len() int {
    return len(s)
}

func (s byPublished) Swap(i, j int) {
    s[i], s[j] = s[j], s[i]
}

func (s byPublished) Less(i, j int) bool {
    return s[i].Items[0].PublishedParsed.Before(*s[j].Items[0].PublishedParsed)
}

func getAuthor(feed *gofeed.Feed) string {
  if feed.Author != nil {
    return feed.Author.Name
  }
  if feed.Items[0].Author != nil {
    return feed.Items[0].Author.Name
  }
  // TODO: handle better
  fmt.Printf("Could not determine author for %v", feed.Link)
  return "Unknown Author"
}

func combineallFeeds(allFeeds []*gofeed.Feed) *feeds.Feed {
  feed := &feeds.Feed{
      // TODO: where to pull this metadata from?
      Title:       "Engineering Manager Blogs",
      Link:        &feeds.Link{Href: "https://github.com/chase-seibert/engineering-manager-blogs"},
      Description: "Collection of Engineering Manager Blog RSS Feeds",
      Author:      &feeds.Author{Name: "Chase Seibert", Email: "chase.seibert@gmail.com"},
      Created:     time.Now(),
  }
  sort.Sort(byPublished(allFeeds))
  for _, sourceFeed := range allFeeds {
    // TODO: interleave ALL items and then sort?
    item := sourceFeed.Items[0]
    feed.Items = append(feed.Items, &feeds.Item{
      Title:       item.Title,
      Link:        &feeds.Link{Href: item.Link},
      Description: item.Description,
      //Author:      &feeds.Author{Name: item.Author.Name, Email: item.Author.Email},
      Author:      &feeds.Author{Name: getAuthor(sourceFeed)},
      Created:     *item.PublishedParsed,
      //Updated:     *item.UpdatedParsed,
      Content:     item.Content,
    })
  }
  return feed
}

func handler(w http.ResponseWriter, r *http.Request) {
  urls := getUrls(README_URL)
  allFeeds := fetchUrls(urls)
	//fmt.Printf("%#v", allFeeds)
  combinedFeed := combineallFeeds(allFeeds)
  //fmt.Printf("%#v", combinedFeed)
  atom, _ := combinedFeed.ToAtom()
  fmt.Fprintf(w, atom)
}

func main() {
  http.HandleFunc("/", handler)
  log.Fatal(http.ListenAndServe(":8080", nil))
}

/*
- Hard code URLs, fetch in serial, print to screen
- Parse RSS and produce new stream
- Parallelize fetching, error handling
- Get working as RSS server
- Serve RSS XML somewhere
- Fetch list of URLs dynamically
- Caching
- lint/format
*/
