package main

import (
	"fmt"
	"log"
	"net/http"
    "io/ioutil"
	"os"
    "encoding/json"
	"time"
    "strings"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/layout"
    "fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/data/binding"
)

type AppState struct {
    playing       bool
    audio         Audio
    elapsedTime   binding.Float
    audioSelected bool 
}

// note: struct fields have to be capital inorder for the json parsing to work, its annoying and ugly I know but such is life with go
type Audio struct {
    Author     string `json:"author"`
	//Set      string
    Name       string `json:"name"`
    Duration   int64  `json:"duration"`
}

func makeAudioRequest(as *AppState, name string, author string, duration int64) {
    var request = codifyAudioEntry(name, author)
    var url = fmt.Sprintf("http://localhost:8080/audio?file=%s", request)
    if (request == "") {
        log.Fatal("Cannot make empty request")
        panic(1)
    }

    response, err := http.Get(url)
	if err != nil {
		log.Fatal("Failed to make GET request: ", err)
	}
	defer response.Body.Close()

	// Check if the response status code is OK
	if response.StatusCode != http.StatusOK {
		fmt.Println(response.Body)
		log.Fatal("Server responded with status: ", response.Status)
	}

	// Decode the MP3 audio stream
	streamer, format, err := mp3.Decode(response.Body)
	if err != nil {
		log.Fatal("Failed to decode audio: ", err)
	}
	defer streamer.Close()

	// Initialize the speaker with the audio format
    speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	// Play the audio stream
	var done = make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
	    speaker.Clear()
	    close(done)
	})))

    for {
        if (as.audioSelected) {
            fmt.Println("[client: makeAudioRequest] finished streamer")
            speaker.Clear()
            <-done
            close(done)
            return
        }

        // elapsed time in seconds
        var elapsed int64 = int64(format.SampleRate.D(streamer.Position()).Round(time.Second) / 1000000000)
        var percent = float64(elapsed) / float64(duration)

        fmt.Printf("[client] elapsed time: %d, duration: %d, percentage: %.2f\n", elapsed, duration, percent)

        as.elapsedTime.Set(percent)
        if percent >= 1.0 {
            as.audioSelected = true; // break the loop
        }
        time.Sleep(time.Second / 10)
    }
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

func pauseAudio(as *AppState) {
    if as.playing {
        speaker.Lock()
        as.playing = false
    } else {
        speaker.Unlock()
        as.playing = true
    }
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

func main() {
	a := app.New()
	w := a.NewWindow("VAMP")
    w.Resize(fyne.NewSize(1920, 1080))

    audioRaw := fetchAudioList()

    elapsedBinding := binding.NewFloat()
    progressBar := widget.NewProgressBarWithData(elapsedBinding)

    var as = AppState{false, Audio{"", "", 0}, elapsedBinding, false}

    list := widget.NewList(
        func() int {
            return len(audioRaw)
        },
        func() fyne.CanvasObject {
            return widget.NewButton("Name - Author", func() {})
        },
        func (i widget.ListItemID, o fyne.CanvasObject) {
            button := o.(*widget.Button)
            button.SetText(fmt.Sprintf("%s - %s", audioRaw[i].Name, removeMP3Tag(audioRaw[i].Author)))
            button.Alignment = widget.ButtonAlignLeading
            button.OnTapped = func() {
                // first clear the speaker if anything is currently playing
                speaker.Clear()
                // next signal to the previous song playing thread that it needs to shutdown
                as.audioSelected = true
                // give that thread time to shutdown before the new one
                time.Sleep(time.Second / 2)
                // turn shutdown signal off so that new thread can play next song
                as.audioSelected = false
                go func() {
                    as.playing = true
                    makeAudioRequest(&as, audioRaw[i].Name, audioRaw[i].Author, audioRaw[i].Duration)
                }()
            }
        })

    bar := container.NewVBox(
        progressBar,
        widget.NewButton("Pause", func() {
            pauseAudio(&as)
        }),
        widget.NewButton("Quit", func() {
            os.Exit(0)
        }),
    )

    root := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, bar, nil, nil), bar, bar, list)

	w.SetContent(root)
	w.ShowAndRun()
}


/*
TODO: 
1. Currently pause button breaks program if you press it while a song is not being played

*/
