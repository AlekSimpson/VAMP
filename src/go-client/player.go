package main

import (
	"fmt"
	"log"
	"io/ioutil"
	"encoding/json"
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
    paused           bool
    playbackActive   bool
    currentlyPlaying *Audio
    elapsedTime      binding.Float
    killThreadSignal bool
    verbose          bool
}

func (p *Player) guardedPrint(format string, args ...any) {
	if p.verbose {
		fmt.Printf(format, args)
	}
}

func newPlayer(v bool) (*Player) {
	return &Player{false, false, nil, binding.NewFloat(), false, v}
}

func (p *Player) bindElapsedTime(et binding.Float) {
	p.elapsedTime = et;
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

func (p *Player) loadStreamerFormat(response *http.Response) (beep.StreamSeekCloser, beep.Format) {
	// Decode the MP3 audio stream
	streamer, format, err := mp3.Decode(response.Body)
	if err != nil {
		log.Fatal("Failed to decode audio: ", err)
        panic(1)
	}

    return streamer, format
}

func (p *Player) getURL(request string) string {
    var url = fmt.Sprintf("http://localhost:8080/audio?file=%s", request)
    if (request == "") {
        log.Fatal("Cannot make empty request")
        panic(1)
    }
	return url
}

func (p *Player) makeAudioRequest(name string, author string, duration int64) {
    var request = codifyAudioEntry(name, author)
	fmt.Printf("[client] Requesting: %s by %s\n", name, author)

	var url = p.getURL(request)

    response := p.makeGetRequest(url)
	defer response.Body.Close()

    streamer, format := p.loadStreamerFormat(response)
    defer streamer.Close()

	// Initialize the speaker with the audio format
    speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	// Play the audio stream
	var done = make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
	    speaker.Clear()
		p.playbackActive = false
	    close(done)
	})))

    for {
        if (p.killThreadSignal) {
			fmt.Printf("[client: makeAudioRequest] finished playback\n")
            speaker.Clear()
			p.playbackActive = false
            <-done
            close(done)
            return
        }

        // calculate elapsed time in seconds
        var elapsed int64 = int64(format.SampleRate.D(streamer.Position()).Round(time.Second) / 1000000000)
        var percent       = float64(elapsed) / float64(duration)

		//p.guardedPrint("[client] elapsed time: %d, duration: %d, percentage: %.2f\n", elapsed, duration, percent)
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
        p.paused         = false
		p.playbackActive = true
        p.makeAudioRequest(name, author, duration) // false at the end is for turning off verbose
    }()
}

func removeMP3Tag(str string) string {
    var delimited = strings.Split(str, ".")
    return delimited[0]
}

func codifyAudioEntry(name string, authorRaw string) string {
    // remove .mp3
    var author = removeMP3Tag(authorRaw)

    // begin splicing
    var splice = fmt.    Sprintf("%s_%s", name, author)
    splice     = strings.ReplaceAll(splice, " ", "-")
    splice     = strings.ToUpper(splice)
    return splice
}

func fetchAudioList() []Audio {
    response, err := http.Get("http://localhost:8080/availableAudio")
    if err != nil {
        log.Fatal("Failed to make GET request: ", err)
    }
    defer response.Body.Close()

    if response.StatusCode != http.StatusOK {
        log.Fatal("Server responded with status: ", response.Status)
    }

    // parse response.Body
    var result []Audio
    body, err := ioutil.ReadAll(response.Body)
    if err != nil {
        panic(err)
    }

    err = json.Unmarshal(body, &result)
    if err != nil {
        panic(err)
    }

    return result
}
