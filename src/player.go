package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
    "strings"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"

	"fyne.io/fyne/v2/data/binding"
)

// note: struct fields have to be capital inorder for the json parsing to work, its annoying and ugly I know but such is life with go
type Audio struct {
    Author     string `json:"author"`
	//Set      string
    Name       string `json:"name"`
    Duration   int64  `json:"duration"`
}

// TODO maybe setup a requests queue in this later to handle song queue and stuff
type Player struct {
    playing          bool
    currentlyPlaying *Audio
    elapsedTime      binding.Float
    killThreadSignal bool
    verbose          bool
}

func removeMP3Tag(str string) string {
    var delimited = strings.Split(str, ".")
    return delimited[0]
}

func codifyAudioEntry(name string, authorRaw string) string {
    // remove .mp3
    var author = removeMP3Tag(authorRaw)

    // begin splicing
    var splice = fmt.Sprintf("%s_%s", name, author)
    splice = strings.ReplaceAll(splice, " ", "-")
    splice = strings.ToUpper(splice)
    return splice
}

func (p *Player) makeGetRequest(url string) *http.Response {
    response, err := http.Get(url)
    if err != nil {
        log.Fatal("Cannot make empty request")
        panic(1)
    }

	// Check if the response status code is OK
	if response.StatusCode != http.StatusOK {
		fmt.Println(response.Body)
		log.Fatal("Server responded with status: ", response.Status)
        panic(1)
	}

    return response
}

func (p *Player) loadStreamerFormat(response *http.Response) (*mp3.decoder, beep.Format) {
	// Decode the MP3 audio stream
	decoder, format, err := mp3.Decode(response.Body)
	if err != nil {
		log.Fatal("Failed to decode audio: ", err)
        panic(1)
	}

    return decoder, format
}

func (p *Player) makeAudioRequest(name string, author string, duration int64) {
    var request = codifyAudioEntry(name, author)
    var url = fmt.Sprintf("http://localhost:8080/audio?file=%s", request)
    if (request == "") {
        log.Fatal("Cannot make empty request")
        panic(1)
    }

    response := p.makeGetRequest(url)
	defer response.Body.Close()

    decoder, format := p.loadStreamerFormat(response)
    defer decoder.Close()

	// Initialize the speaker with the audio format
    speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	// Play the audio stream
	var done = make(chan bool)
	speaker.Play(beep.Seq(decoder, beep.Callback(func() {
	    speaker.Clear()
	    close(done)
	})))

    for {
        if (p.killThreadSignal) {
            fmt.Println("[client: makeAudioRequest] finished playback")
            speaker.Clear()
            <-done
            close(done)
            return
        }

        // calculate elapsed time in seconds
        var elapsed int64 = int64(format.SampleRate.D(decoder.Position()).Round(time.Second) / 1000000000)
        var percent = float64(elapsed) / float64(duration)

        if p.verbose {
            fmt.Printf("[client] elapsed time: %d, duration: %d, percentage: %.2f\n", elapsed, duration, percent)
        }

        p.elapsedTime.Set(percent)
        if percent >= 1.0 {
            p.killThreadSignal = true; // break the loop
        }
        time.Sleep(time.Second / 10) // update only every 10th of a second
    }
}

func (p *Player) playSong(name string, author string, duration int64) {
    // first clear the speaker if anything is currently playing
    speaker.Clear()
    // next signal to the previous song playing thread that it needs to shutdown
    p.killThreadSignal = true
    // give that thread time to shutdown before the new one (needs to be more than 1/10 a second)
    time.Sleep(time.Second / 5)
    // turn shutdown signal off so that new thread can play next song
    p.killThreadSignal = false
    go func() {
        p.playing = true
        p.makeAudioRequest(name, author, duration) // false at the end is for turning off verbose
    }()
}

func main() {
    fmt.Println("test")
}
