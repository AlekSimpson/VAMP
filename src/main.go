package main

import (
	// "io/ioutil"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func audioHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("handling song request...")

	// process uri to get requested song
	var uri string = r.RequestURI

	var delimited = strings.Split(uri, "?song=") // [audio, <song name>]
	if len(delimited) <= 1 {
		http.Error(w, "Sorry please provide a song to play", http.StatusInternalServerError)
		return
	}

	var songName = strings.ToUpper(delimited[1]) // index 1 will always have the song name
	var filepath = "../assets/" + songName + ".mp3"

	// Open the audio file
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		http.Error(w, "Could not open audio file", http.StatusInternalServerError)
		return
	}

	// Set the content type header
	w.Header().Set("Content-Type", "audio/mpeg")

	// Copy the audio file to the response writer
	binary.Write(w, binary.BigEndian, &data)
}

func main() {
	http.HandleFunc("/audio", audioHandler)
	http.ListenAndServe(":8080", nil)
}
