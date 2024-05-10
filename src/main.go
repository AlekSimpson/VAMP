package main

import (
    "encoding/json"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
    "log"
)

type FileModel struct {
    Name   string `json:"name"`
    Author string `json:"author"`
}

func audioHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("handling song request...")

	// process uri to get requested song
	var uri string = r.RequestURI

	var delimited = strings.Split(uri, "?file=") // [audio, <file name>]
	if len(delimited) <= 1 {
		http.Error(w, "Sorry please provide a song to play", http.StatusInternalServerError)
		return
	}

	var audioName = strings.ToUpper(delimited[1]) // index 1 will always have the song name
	var filepath = "../assets/" + audioName + ".mp3"

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

// makes the inputted string a bit more readable (as a title)
// Ex: DRIVING-MY-LOVE ==> Driving My Love
func prettify(name string) string {
    var retval = name
    retval = strings.ReplaceAll(retval, "-", " ")
    retval = strings.ToLower(retval)
    retval = strings.Title(retval)
    return retval
}

func processFilename(filename string) FileModel {
    var delimited = strings.Split(filename, "_")
    name := delimited[0];
    author := delimited[1];

    // process strings to be a bit more readable
    name = prettify(name)
    author = prettify(author)

    return FileModel{name, author}
}

func getAvailableAudio(w http.ResponseWriter, r *http.Request) {
    // get audio file names
    files, err := ioutil.ReadDir("../assets/")
    if err != nil {
        log.Fatalf("Error reading directory: %s\n", err);
    }

    // process file names into go model and then into json
    var fileModels = make([]FileModel, len(files))
    for i, file := range files {
        fileModels[i] = processFilename(file.Name())
    }

    jsonData, err := json.Marshal(fileModels)
	if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
	    return
	}

    // send response
    w.Header().Set("Content-Type", "application/json")
    w.Write(jsonData)
}

func main() {
	http.HandleFunc("/audio", audioHandler)
    http.HandleFunc("/availableAudio", getAvailableAudio)
	http.ListenAndServe(":8080", nil)
}
