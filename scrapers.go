package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Nyaa blocks plain Go/Python scrapers, so we spoof a browser UA.
func fetchPage(url string) (*goquery.Document, int, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "+
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, resp.StatusCode, fmt.Errorf("upstream returned %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return doc, resp.StatusCode, nil
}

func scrapeList(url string) ([]Torrent, error) {
	doc, _, err := fetchPage(url)
	if err != nil {
		return nil, err
	}

	var torrents []Torrent

	doc.Find("tbody tr").Each(func(_ int, row *goquery.Selection) {
		torrentPath, _ := row.Find("td:nth-child(2) a").Last().Attr("href")
		filePath, _ := row.Find("td:nth-child(3) a:nth-child(1)").Attr("href")
		magnet, _ := row.Find("td:nth-child(3) a:nth-child(2)").Attr("href")
		category, _ := row.Find("td:nth-child(1) a").Attr("title")

		// torrentPath is like /view/12345, grab the last segment
		idStr := ""
		parts := strings.Split(torrentPath, "/")
		if len(parts) >= 3 {
			idStr = parts[2]
		}
		id, _ := strconv.Atoi(idStr)

		seeders, _ := strconv.Atoi(strings.TrimSpace(row.Find("td:nth-child(6)").Text()))
		leechers, _ := strconv.Atoi(strings.TrimSpace(row.Find("td:nth-child(7)").Text()))
		completed, _ := strconv.Atoi(strings.TrimSpace(row.Find("td:nth-child(8)").Text()))

		torrents = append(torrents, Torrent{
			ID:        id,
			Title:     strings.TrimSpace(row.Find("td:nth-child(2) a").Last().Text()),
			Link:      NyaaBaseURL + torrentPath,
			File:      NyaaBaseURL + filePath,
			Category:  category,
			Size:      strings.TrimSpace(row.Find("td:nth-child(4)").Text()),
			Uploaded:  strings.TrimSpace(row.Find("td:nth-child(5)").Text()),
			Seeders:   seeders,
			Leechers:  leechers,
			Completed: completed,
			Magnet:    magnet,
		})
	})

	return torrents, nil
}

func scrapeDetail(url string) (*File, error) {
	doc, _, err := fetchPage(url)
	if err != nil {
		return nil, err
	}

	container := doc.Find("body div.container").Last()

	parts := strings.Split(url, "/")
	id, _ := strconv.Atoi(parts[len(parts)-1])

	fileHref, _ := container.Find("div.panel-footer a").First().Attr("href")
	magnet, _ := container.Find("div.panel-footer a:nth-child(2)").Attr("href")

	// rows: 1=category+date, 2=submitter+seeders, 3=info_hash+leechers, 4=size+completed, 5=hash
	// col-md-5:nth-child(2) = left value, col-md-5:nth-child(4) = right value
	seeders, _ := strconv.Atoi(strings.TrimSpace(
		container.Find("div.panel-body div.row:nth-child(2) .col-md-5:nth-child(4)").Text()))
	leechers, _ := strconv.Atoi(strings.TrimSpace(
		container.Find("div.panel-body div.row:nth-child(3) .col-md-5:nth-child(4)").Text()))
	completed, _ := strconv.Atoi(strings.TrimSpace(
		container.Find("div.panel-body div.row:nth-child(4) .col-md-5:nth-child(4)").Text()))

	torrent := Torrent{
		ID:        id,
		Title:     strings.TrimSpace(container.Find("h3.panel-title").First().Text()),
		File:      NyaaBaseURL + fileHref,
		Link:      fmt.Sprintf("%s/view/%d", NyaaBaseURL, id),
		Magnet:    magnet,
		Size:      strings.TrimSpace(container.Find("div.panel-body div.row:nth-child(4) .col-md-5:nth-child(2)").Text()),
		Category:  strings.TrimSpace(container.Find("div.panel-body div.row:nth-child(1) .col-md-5:nth-child(2)").Text()),
		Uploaded:  strings.TrimSpace(container.Find("div.panel-body div.row:nth-child(1) .col-md-5:nth-child(4)").Text()),
		Seeders:   seeders,
		Leechers:  leechers,
		Completed: completed,
	}

	// header is "Comments - N"
	commentHeader := container.Find("div#comments h3.panel-title").Text()
	commentParts := strings.Split(commentHeader, "-")
	commentCount := 0
	if len(commentParts) > 0 {
		commentCount, _ = strconv.Atoi(strings.TrimSpace(commentParts[len(commentParts)-1]))
	}

	var comments []Comment
	if commentCount > 0 {
		container.Find("div#comments div.comment-panel div.panel-body").Each(func(_ int, el *goquery.Selection) {
			img, exists := el.Find("img.avatar").Attr("src")
			if !exists || img == "" {
				// deleted/anonymous accounts have no avatar
				img = DefaultProfilePic
			}
			comments = append(comments, Comment{
				Name:      strings.TrimSpace(el.Find("a").First().Text()),
				Content:   el.Find("div.comment-body div.comment-content").Text(),
				Image:     img,
				Timestamp: el.Find("a").Children().First().Text(),
			})
		})
	}
	if comments == nil {
		// return an empty slice rather than null in the JSON
		comments = []Comment{}
	}

	return &File{
		Torrent:     torrent,
		Description: container.Find("div.panel-body#torrent-description").Text(),
		SubmittedBy: strings.TrimSpace(container.Find("div.panel-body div.row:nth-child(2) .col-md-5:nth-child(2)").Text()),
		InfoHash:    strings.TrimSpace(container.Find("div.panel-body div.row:nth-child(5) .col-md-5:nth-child(2)").Text()),
		CommentInfo: Comments{
			Count:    commentCount,
			Comments: comments,
		},
	}, nil
}
