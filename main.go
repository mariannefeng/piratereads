package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type goodreadsRSS struct {
	Channel goodreadsChannel `xml:"channel"`
}

type goodreadsChannel struct {
	Items []goodreadsItem `xml:"item"`
}

type goodreadsItem struct {
	Title          string `xml:"title"`
	Link           string `xml:"link"`
	AuthorName     string `xml:"author_name"`
	BookSmallImage string `xml:"book_small_image_url"`
	UserRating     int    `xml:"user_rating"`
	UserReview     string `xml:"user_review"`
	Description    string `xml:"description"`
}

type review struct {
	BookTitle    string `json:"book_title"`
	BookAuthor   string `json:"book_author"`
	BookCoverImg string `json:"book_cover_img"`
	BookLink     string `json:"book_link"`
	Rating       int    `json:"rating"`
	Text         string `json:"text"`
}

type reviewsResponse struct {
	Count   int      `json:"count"`
	Reviews []review `json:"reviews"`
}

func extractBookAnchor(description string) string {
	start := strings.Index(description, "<a ")
	if start == -1 {
		return ""
	}

	rest := description[start:]
	endRel := strings.Index(rest, "</a>")
	if endRel == -1 {
		return ""
	}

	return rest[:endRel+len("</a>")]
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/{username}/reviews", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		username := vars["username"]
		if strings.TrimSpace(username) == "" {
			http.NotFound(w, r)
			return
		}

		query := r.URL.Query()

		perPage := 100
		if v := query.Get("per_page"); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil || n <= 0 {
				http.Error(w, "per_page must be a positive integer", http.StatusBadRequest)
				return
			}
			perPage = n
		}

		page := 1
		if v := query.Get("page"); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil || n <= 0 {
				http.Error(w, "page must be a positive integer", http.StatusBadRequest)
				return
			}
			page = n
		}

		rssURL := fmt.Sprintf(
			"https://www.goodreads.com/review/list_rss/%s?shelf=read&per_page=%d&page=%d",
			username,
			perPage,
			page,
		)

		resp, err := http.Get(rssURL)
		if err != nil {
			log.Printf("error fetching Goodreads RSS for %q: %v", username, err)
			http.Error(w, "failed to fetch reviews", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("unexpected Goodreads status for %q: %d", username, resp.StatusCode)
			http.Error(w, "failed to fetch reviews", http.StatusBadGateway)
			return
		}

		var rss goodreadsRSS
		if err := xml.NewDecoder(resp.Body).Decode(&rss); err != nil {
			log.Printf("error decoding Goodreads RSS for %q: %v", username, err)
			http.Error(w, "failed to parse reviews", http.StatusBadGateway)
			return
		}

		reviews := make([]review, 0, len(rss.Channel.Items))
		for _, item := range rss.Channel.Items {
			text := strings.TrimSpace(item.UserReview)
			bookLink := extractBookAnchor(item.Description)
			if bookLink == "" {
				bookLink = item.Link
			}

			reviews = append(reviews, review{
				BookTitle:    item.Title,
				BookAuthor:   item.AuthorName,
				BookCoverImg: item.BookSmallImage,
				BookLink:     bookLink,
				Rating:       item.UserRating,
				Text:         text,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		respBody := reviewsResponse{
			Count:   len(reviews),
			Reviews: reviews,
		}

		if err := json.NewEncoder(w).Encode(respBody); err != nil {
			log.Printf("error encoding reviews response: %v", err)
		}
	}).Methods(http.MethodGet)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
