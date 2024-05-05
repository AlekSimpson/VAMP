package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Player struct {
	elapsed int
	total   int
	quit    chan struct{}
}

type Song struct {
	artist   string
	album    string
	track    string
	duration time.Duration
}

func (af ApplicationInterface) Run() {
	switch af.action {
	case 0:
		makeSongRequest()
	default:
		fmt.Println("Invalid action")
	}
}

func makeSongRequest() {
	response, err := http.Get("http://localhost:8080/audio?song=phantom_chizh")
	if err != nil {
		log.Fatal("Failed to make GET request:", err)
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
		fmt.Println("test")
		speaker.Clear()
		close(done)
	})))

	// Wait for the audio to finish playing
	<-done
}

func main() {
	// var pm ProcessMan = makeProcessMan(ApplicationInterface{0})
	// var playing = false

	a := app.New()
	w := a.NewWindow("Hello")

	hello := widget.NewLabel("Hello Fyne!")
	w.SetContent(container.NewVBox(
		hello,
		widget.NewButton("Hi!", func() {
			hello.SetText("Welcome :)")
		}),
	))

	w.ShowAndRun()

	// begin application
	//for true {
	//	fmt.Println(tuiMessage)
	//	fmt.Scanln(&selection)
	//	switch selection {
	//	case "q":
	//		return
	//	case "p":
	//		// play song
	//		go pm.process(0)
	//	case "d":
	//		pm.toggleProcess(0)
	//	case "x":
	//		// stop song
	//		if paused {
	//			speaker.Unlock()
	//			paused = false
	//		}else {
	//			speaker.Lock()
	//			paused = true
	//		}
	//	default:
	//		fmt.Println("Sorry that input was invalid")
	//		return
	//	}
	//}
}
