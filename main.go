package main

import (
	"crypto/sha256"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/mariannefeng/piratereads/docs"
	posthog "github.com/posthog/posthog-go"
	httpSwagger "github.com/swaggo/http-swagger"
)

type goodreadsRSS struct {
	Channel goodreadsChannel `xml:"channel"`
}

type goodreadsChannel struct {
	Items []goodreadsItem `xml:"item"`
}

type goodreadsItem struct {
	Title              string  `xml:"title"`
	Link               string  `xml:"link"`
	AuthorName         string  `xml:"author_name"`
	BookSmallImageURL  string  `xml:"book_small_image_url"`
	BookMediumImageURL string  `xml:"book_medium_image_url"`
	BookLargeImageURL  string  `xml:"book_large_image_url"`
	UserRating         int     `xml:"user_rating"`
	AverageRating      float64 `xml:"average_rating"`
	UserReview         string  `xml:"user_review"`
	Description        string  `xml:"description"`
	PubDate            string  `xml:"pubDate"`
}

type review struct {
	BookTitle       string  `json:"book_title"`
	BookAuthor      string  `json:"book_author"`
	BookCoverSmall  string  `json:"book_cover_small"`
	BookCoverMedium string  `json:"book_cover_medium"`
	BookCoverLarge  string  `json:"book_cover_large"`
	BookLink        string  `json:"book_link"`
	Rating          int     `json:"rating"`
	AverageRating   float64 `json:"avg_rating"`
	Text            string  `json:"text"`
	PublishedOn     string  `json:"published_on"`
}

type reviewsResponse struct {
	Count   int      `json:"count"`
	Reviews []review `json:"reviews"`
}

// @title           piratereads API
// @BasePath        /

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func analyticsMiddleware(client posthog.Client) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := newResponseWriter(w)
			next.ServeHTTP(rw, r)

			ip := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0])
			if ip == "" {
				ip = r.RemoteAddr
			}
			distinctID := fmt.Sprintf("%x", sha256.Sum256([]byte(ip)))

			props := posthog.NewProperties().
				Set("$ip", ip).
				Set("endpoint", r.URL.Path).
				Set("method", r.Method).
				Set("status_code", rw.statusCode)

			if userID := mux.Vars(r)["user_id"]; userID != "" {
				props.Set("goodreads_user_id", userID)
			}

			if err := client.Enqueue(posthog.Capture{
				DistinctId: distinctID,
				Event:      "api_request",
				Properties: props,
			}); err != nil {
				log.Printf("posthog enqueue error: %v", err)
			}
		})
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func extractBookURL(description string) string {
	const hrefPrefix = `href="`
	start := strings.Index(description, hrefPrefix)
	if start == -1 {
		return ""
	}
	start += len(hrefPrefix)
	end := strings.Index(description[start:], `"`)
	if end == -1 {
		return ""
	}
	return description[start : start+end]
}

// getReviewsHandler godoc
// @Summary      Get goodreads reviews for a user
// @Description  Returns a paginated list of reviews
// @Tags         reviews
// @Param        user_id  path   string  true   "goodreads id"
// @Param        per_page  query  int     false  "number of reviews per page"
// @Param        page      query  int     false  "page number"
// @Produce      json
// @Success      200  {object}  reviewsResponse
// @Failure      400  {string}  string  "invalid request"
// @Failure      404  {string}  string  "user not found"
// @Failure      502  {string}  string  "goodreads error"
// @Router       /{user_id}/reviews [get]
func getReviewsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["user_id"]
	if strings.TrimSpace(userId) == "" {
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
		userId,
		perPage,
		page,
	)

	resp, err := http.Get(rssURL)
	if err != nil {
		log.Printf("error fetching goodreads RSS for %q: %v", userId, err)
		http.Error(w, "failed to fetch reviews", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("unexpected goodreads status for %q: %d", userId, resp.StatusCode)
		http.Error(w, "failed to fetch reviews", http.StatusBadGateway)
		return
	}

	var rss goodreadsRSS
	if err := xml.NewDecoder(resp.Body).Decode(&rss); err != nil {
		log.Printf("error decoding goodreads RSS for %q: %v", userId, err)
		http.Error(w, "failed to parse reviews", http.StatusBadGateway)
		return
	}

	reviews := make([]review, 0, len(rss.Channel.Items))
	for _, item := range rss.Channel.Items {
		text := strings.TrimSpace(item.UserReview)
		bookLink := extractBookURL(item.Description)
		if bookLink == "" {
			bookLink = item.Link
		}
		reviews = append(reviews, review{
			BookTitle:       item.Title,
			BookAuthor:      item.AuthorName,
			BookCoverSmall:  item.BookSmallImageURL,
			BookCoverMedium: item.BookMediumImageURL,
			BookCoverLarge:  item.BookLargeImageURL,
			BookLink:        bookLink,
			Rating:          item.UserRating,
			AverageRating:   item.AverageRating,
			Text:            text,
			PublishedOn:     strings.TrimSpace(item.PubDate),
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
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("no .env file found: %v", err)
	}

	r := mux.NewRouter()

	r.HandleFunc("/{user_id}/reviews", getReviewsHandler).Methods(http.MethodGet)
	r.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	r.Use(corsMiddleware)

	if apiKey := os.Getenv("POSTHOG_API_KEY"); apiKey != "" {
		phClient, err := posthog.NewWithConfig(apiKey, posthog.Config{
			Endpoint: "https://us.i.posthog.com",
		})
		if err != nil {
			log.Printf("posthog init error: %v", err)
		} else {
			defer phClient.Close()
			r.Use(analyticsMiddleware(phClient))
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
