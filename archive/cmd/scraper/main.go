package main

import (
	"encoding/json"
	"fmt"
	"log"
	"resty.dev/v3"
)

func main() {
	client := resty.New()
	defer client.Close()

	formMap := map[string]string{
		"log":       "mnichols",
		"pwd":       "<redacted>",
		"wp-submit": "Log In",
	}
	// login
	client.SetRedirectPolicy(resty.NoRedirectPolicy())
	res, err := client.R().
		EnableTrace().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetMultipartFormData(formMap).
		Post("https://woodworkingmasterclasses.com/wp-login.php")
	defer res.Body.Close()
	fmt.Println(err, res)
	fmt.Println(res.Request.TraceInfo())
	fmt.Println("status", res.StatusCode())
	for _, c := range res.Cookies() {
		fmt.Println(c.Name, c.Value)
	}
	cookies := res.Header()["Set-Cookie"]
	for _, c := range cookies {
		fmt.Println(c)
	}

	searchURL := "https://woodworkingmasterclasses.com/wp-admin/admin-ajax.php"
	searchFormParams := map[string]string{
		"action":                      "wwmc_search_videos",
		"searchData[categories][0]":   "1458",
		"searchData[sorting]":         "asc",
		"searchData[projects]":        "all",
		"searchData[showIntros]":      "show",
		"searchData[showEpisodes]":    "",
		"searchData[showStandalones]": "show",
		"searchData[watchedVideos]":   "show",
		"searchData[keywords]":        "",
	}
	// request page of HTML from search
	client.SetRedirectPolicy(resty.FlexibleRedirectPolicy(5))
	var searchResults = []*PostSummaryData{}

	searchRes, err := client.R().
		EnableTrace().
		SetCookies(res.Cookies()).
		SetMultipartFormData(searchFormParams).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetDoNotParseResponse(true).
		Post(searchURL)

	//SetQueryString("sorting=date%20asc&projects=all&showIntros=show&showEpisodes=&showStandalones=show&watchedVideos=show&keywords=#main").
	//Get("https://woodworkingmasterclasses.com/video-library/")
	defer searchRes.Body.Close()
	//fmt.Println(err, searchRes)
	fmt.Println(searchRes.Request.TraceInfo())
	fmt.Println("status", searchRes.StatusCode())

	dec := json.NewDecoder(searchRes.Body)
	if err = dec.Decode(&searchResults); err != nil {
		log.Fatal(err)
	}
	//if err = encoding.DecodeJSONResponse(searchRes.RawResponse, &searchResults); err != nil {
	//	log.Fatal(err)
	//}
	//if searchRes.IsRead {
	//	log.Fatal("reset of body is needed")
	//}
	//rc := io.NopCloser(searchRes.Body)
	//if err := encoding.DecodeReadCloser(rc, searchResults); err != nil {
	//	log.Fatal(err)
	//}

	//if err := encoding.DecodeJSONResponse(searchRes., &searchResults); err != nil {
	//	log.Fatal(err)
	//}

	fmt.Println("received ", len(searchResults), " items")
	for i, result := range searchResults {
		fmt.Println("index", i, "permalink:", result.Permalink)
	}

	//fmt.Println(searchRes.String())
	//
	//doc, err := goquery.NewDocumentFromReader(searchRes.Body)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//// Find the review items
	//selection := doc.Find(".video-item")
	//fmt.Println("found", selection.Length(), "items")
	//selection.Each(func(i int, s *goquery.Selection) {
	//	// For each item found, get the title
	//	//title := s.Find("a").Text()
	//	href, ok := s.Attr("href")
	//	if !ok {
	//		href = "unknown"
	//	}
	//	fmt.Printf("Review %d: %s\n", i, href)
	//})
}

type PostSummaryData struct {
	ID        int64  `json:"ID"`
	Permalink string `json:"permalink"`
	PostTitle string `json:"post_title"`
}
