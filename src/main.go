package main

import (
    "encoding/json"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
    "bytes"

	"github.com/hajimehoshi/go-mp3"
)

func serverLog(message string, funcOfOrigin string) {
    fmt.Printf("[server: %s] %s\n", funcOfOrigin, message)
}

func fatalServerLog(message string, funcOfOrigin string) {
    serverLog(message, funcOfOrigin)
    panic(1)
}

type FileModel struct {
    Name     string `json:"name"`
    Author   string `json:"author"`
    Duration int64  `json:"duration"`
}

func readAudio(filename string) []byte {
    var f = filename
    if (!strings.Contains(filename, ".mp3")) {
        f = filename + ".mp3"
    }

    var filepath = fmt.Sprintf("../assets/%s", f)
    data, err := ioutil.ReadFile(filepath)
    if err != nil {
        fatalServerLog("cannot read audio file", "readAudio")
    }
    return data
}

func readAudioDuration(filename string) int64 {
    var data = readAudio(filename)

    decoder, err := mp3.NewDecoder(bytes.NewReader(data))
    if err != nil {
        fatalServerLog("can't get audio length", "readAudioDuration")
    }

    // for some reason multiplying by a sample rate of 52000 will result an approximate calculation of the length of the audio
    // this is despite two go audio libraries telling me the sample rate is 48000, 48000 just does not calculate the right times
    duration := (decoder.Length() * 52000) / 10000000000

    return duration
}

func audioHandler(w http.ResponseWriter, r *http.Request) {
    serverLog("handling song request...", "audioHandler")

	// process uri to get requested song
	var uri string = r.RequestURI

	var delimited = strings.Split(uri, "?file=") // [audio, <file name>]
	if len(delimited) <= 1 {
		http.Error(w, "Sorry please provide a song to play", http.StatusInternalServerError)
		return
	}

	var audioName = strings.ToUpper(delimited[1]) // index 1 will always have the song name

	// Open the audio file
    var data = readAudio(audioName)

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

    // get duration 
    var duration = readAudioDuration(filename)

    return FileModel{name, author, duration}
}

func getAvailableAudio(w http.ResponseWriter, r *http.Request) {
    // get audio file names
    files, err := ioutil.ReadDir("../assets/")
    if err != nil {
        fatalServerLog(fmt.Sprintf("Error reading directory %s", err), "getAvailableAudio")
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
