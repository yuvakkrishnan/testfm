package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const (
	lastFMAPIKey         = "ba20a7c69f0bd21fee290eef9eba2c02"
	musixmatchAPIKey     = "c4ae057a06d121be45b0e597d5139ae5"
	googleAPIKey         = "AIzaSyDQ5E73rafz3N_b-MtkS-S_PX4zZY3jKrk"
	googleCustomSearchID = "c208ac6713e194d91"
)

func main() {
	http.HandleFunc("/artist", artistHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func artistHandler(w http.ResponseWriter, r *http.Request) {
	region := r.URL.Query().Get("region")
	if region == "" {
		http.Error(w, "Missing 'region' parameter", http.StatusBadRequest)
		return
	}

	topTrack, err := getTopTrack(region)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get top track: %s", err), http.StatusInternalServerError)
		log.Printf("Failed to get top track for region %s: %s", region, err)
		return
	}

	lyrics, err := getLyrics(topTrack)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get lyrics: %s", err), http.StatusInternalServerError)
		log.Printf("Failed to get lyrics for track %s by %s: %s", topTrack.Name, topTrack.Artist, err)
		return
	}

	artistInfo, err := getArtistInfo(topTrack.Artist)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get artist information: %s", err), http.StatusInternalServerError)
		log.Printf("Failed to get artist information for %s: %s", topTrack.Artist, err)
		return
	}

	imageURL, err := getArtistImage(topTrack.Artist)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get artist image: %s", err), http.StatusInternalServerError)
		log.Printf("Failed to get artist image for %s: %s", topTrack.Artist, err)
		return
	}

	response := struct {
		TopTrack    Track      `json:"top_track"`
		Lyrics      string     `json:"lyrics"`
		ArtistInfo  ArtistInfo `json:"artist_info"`
		ArtistImage string     `json:"artist_image"`
	}{
		TopTrack:    topTrack,
		Lyrics:      lyrics,
		ArtistInfo:  artistInfo,
		ArtistImage: imageURL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type Track struct {
	Name   string `json:"name"`
	Artist string `json:"artist"`
}

type ArtistInfo struct {
	Name        string `json:"name"`
	Bio         string `json:"bio"`
	YearsActive string `json:"years_active"`
}

func getTopTrack(region string) (Track, error) {
	url := fmt.Sprintf("http://ws.audioscrobbler.com/2.0/?method=geo.gettoptracks&country=%s&api_key=%s&format=json", region, lastFMAPIKey)
	resp, err := http.Get(url)
	if err != nil {
		return Track{}, err
	}
	defer resp.Body.Close()

	var result struct {
		Tracks struct {
			Track []struct {
				Name   string `json:"name"`
				Artist struct {
					Name string `json:"name"`
				} `json:"artist"`
			} `json:"track"`
		} `json:"tracks"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Track{}, err
	}

	if len(result.Tracks.Track) == 0 {
		return Track{}, fmt.Errorf("no tracks found for region %s", region)
	}

	return Track{Name: result.Tracks.Track[0].Name, Artist: result.Tracks.Track[0].Artist.Name}, nil
}

func getLyrics(track Track) (string, error) {
	url := fmt.Sprintf("https://api.musixmatch.com/ws/1.1/matcher.lyrics.get?format=json&apikey=%s&q_track=%s&q_artist=%s", musixmatchAPIKey, url.PathEscape(track.Name), url.PathEscape(track.Artist))
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Message struct {
			Body struct {
				Lyrics struct {
					Body string `json:"lyrics_body"`
				} `json:"lyrics"`
			} `json:"body"`
		} `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	lyrics := result.Message.Body.Lyrics.Body
	if strings.Contains(lyrics, "Unfortunately, we are not licensed") {
		return "", fmt.Errorf("lyrics not available for track %s by %s", track.Name, track.Artist)
	}

	return lyrics, nil
}

func getArtistInfo(artist string) (ArtistInfo, error) {
	return ArtistInfo{}, nil
}

func getArtistImage(artist string) (string, error) {
	return "", nil
}
