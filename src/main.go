package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Story struct {
	Title string `json:"title"`
	ID    string `json:"id"`
	URL   string `json:"url"`
}

// 🔍 Scraping depuis le profil d’un auteur
func getStoriesFromAuthor(author string) ([]Story, error) {
	url := fmt.Sprintf("https://www.wattpad.com/user/%s", author)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var stories []Story
	doc.Find("a.title.meta.on-story-preview").Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Text())
		id, _ := s.Attr("data-story-id")
		href, _ := s.Attr("href")
		stories = append(stories, Story{
			Title: title,
			ID:    id,
			URL:   "https://www.wattpad.com" + href,
		})
	})

	return stories, nil
}

// 📚 Route /api/v1/stories
func storiesHandler(w http.ResponseWriter, r *http.Request) {
	author := r.Header.Get("Author")
	if author == "" {
		http.Error(w, "Missing 'Author' header", http.StatusBadRequest)
		return
	}

	stories, err := getStoriesFromAuthor(author)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stories)
}

// 👀 Route /api/v1/views
func viewsHandler(w http.ResponseWriter, r *http.Request) {
	author := r.Header.Get("Author")
	title := r.Header.Get("Title")

	if author == "" || title == "" {
		http.Error(w, "Missing 'Author' or 'Title' header", http.StatusBadRequest)
		return
	}

	stories, err := getStoriesFromAuthor(author)
	if err != nil {
		http.Error(w, "Failed to fetch stories: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var storyURL string
	for _, story := range stories {
		if strings.EqualFold(story.Title, title) {
			storyURL = story.URL
			break
		}
	}

	if storyURL == "" {
		http.Error(w, "Story not found", http.StatusNotFound)
		return
	}

	res, err := http.Get(storyURL)
	if err != nil {
		http.Error(w, "Failed to fetch story page", http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		http.Error(w, "Failed to parse story page", http.StatusInternalServerError)
		return
	}
	
	var views int
	doc.Find("span.sr-only").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if strings.HasPrefix(text, "Reads ") {
			clean := strings.ReplaceAll(text[6:], " ", "")
			clean = strings.ReplaceAll(clean, ",", "")
			fmt.Sscanf(clean, "%d", &views)
		}
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"title": title,
		"views": fmt.Sprintf("%d", views),
	})
}

func main() {
	http.HandleFunc("/api/v1/stories", storiesHandler)
	http.HandleFunc("/api/v1/views", viewsHandler)

	fmt.Println("🚀 API running on http://localhost:5000")
	log.Fatal(http.ListenAndServe(":5000", nil))
}