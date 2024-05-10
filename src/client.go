package main

import (
	"fmt"
	"log"
	"net/http"
    "io/ioutil"
	"os"
    "encoding/json"
	"time"

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

// note: struct fields have to be capital inorder for the json parsing to work, its annoying and ugly I know but such is life with go
type Audio struct {
    Author     string `json:"author"`
	//Set      string
    Name       string `json:"name"`
	//Duration time.Duration
}

func makeAudioRequest() {
	response, err := http.Get("http://localhost:8080/audio?file=phantom_chizh")
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
	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		log.Fatal("Failed to initialize speaker: ", err)
	}

	// Play the audio stream
	var done = make(chan struct{})
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		speaker.Clear()
		close(done)
	})))

	// Wait for the audio to finish playing
	<-done
}

func pauseAudio(isPlaying *bool) {
    if *isPlaying {
        speaker.Unlock()
        *isPlaying = false
    } else {
        speaker.Lock()
        *isPlaying = true
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
	// var pm ProcessMan = makeProcessMan(ApplicationInterface{0})
	var playing = false

	a := app.New()
	w := a.NewWindow("VAMP")
    w.Resize(fyne.NewSize(1920, 1080))

    audioRaw := fetchAudioList()
    audioList := binding.NewUntypedList()
    for _, s := range audioRaw {
        audioList.Append(s)
    }

    list := widget.NewListWithData(
		audioList,
		func() fyne.CanvasObject {
            return widget.NewButton("Name - Author", func() {})
            //return container.NewVBox(
            //    widget.NewLabel("Name"),
            //    widget.NewLabel("Author"),
            //)
		},
		func(di binding.DataItem, o fyne.CanvasObject) {
            ctr, _ := o.(*fyne.Container)
            //s := ctr.Objects[0].(*widget.Label)
            //a := ctr.Objects[1].(*widget.Label)
            button := ctr.Objects[0].(*widget.Button)
            diu, _ := di.(binding.Untyped).Get()
            audio := diu.(Audio)

            button.SetText(fmt.Sprintf("%s - %s", audio.Name, audio.Author))
            button.OnTapped = func() {
                fmt.Printf("test\n")
            }
            //s.SetText(audio.Name)
            //a.SetText(audio.Author)
		})

    bar := container.NewVBox(
		widget.NewButton("Play Phantom", func() {
            go makeAudioRequest()
		}),
        widget.NewButton("Pause", func() {
            pauseAudio(&playing)
        }),
        widget.NewButton("Quit", func() {
            os.Exit(0)
        }),
    )

    root := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, bar, nil, nil), bar, bar, list)

	w.SetContent(root)
	w.ShowAndRun()
}
