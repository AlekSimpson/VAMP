package main

import (
	"fmt"
	"os"
	"bufio"
	"log"
	"io/ioutil"
	"github.com/faiface/beep/speaker"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/layout"
    "fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/data/binding"
)

type AppState struct {
	player   *Player
}

func pauseAudio(as *AppState) {
	if as.player.playbackActive {
		if as.player.paused {
			// resume playback
			speaker.Unlock()
			as.player.paused = false
		} else {
			// pause playback
			speaker.Lock()
			as.player.paused = true
		}
	}
}

func makeSideBar(as *AppState) fyne.CanvasObject {
	pauseButton := widget.NewButton("Pause", func() {
		pauseAudio(as)
	})

	bar := container.NewHBox(
        pauseButton,
        widget.NewButton("Quit", func() {
            os.Exit(0)
        }),
    )

	barContainer := container.NewVBox(
		layout.NewSpacer(),
		bar,
	)

	return barContainer
}

func readImageFile(filepath string) []byte {
	var appendedPath = fmt.Sprintf("../assets/app-assets/%s", filepath)
	iconFile, err := os.Open(appendedPath)
	if err != nil {
		log.Fatal(err)
	}

	r := bufio.NewReader(iconFile)

	b, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}
	return b
}

func makePlaybackControl(as *AppState) fyne.CanvasObject {
    // hstack with repeat, skip back, pause/play, skip forward, shuffle
	// repeat button setup
    repeatResource := fyne.  NewStaticResource("../assets/app-assets/repeat.svg", []byte(repeatContent))
    repeat         := widget.NewButtonWithIcon("", repeatResource, func() {
        fmt.Println("repeat song selected")
    }) 

	// skip back button setup
	skipBackResource := fyne.  NewStaticResource("../assets/app-assets/skip-back.svg", []byte(skipBackContent))
    skipBack         := widget.NewButtonWithIcon("", skipBackResource, func() {
        fmt.Println("skipping backwards")
    })

	// pause button setup
	resources := []fyne.Resource{
		fyne.NewStaticResource("../assets/app-assets/play.svg", []byte(playContent)),
		fyne.NewStaticResource("../assets/app-assets/paused.svg", []byte(pauseContent)),
	}
    togglePause := widget.NewButtonWithIcon("", resources[0], func() {})
	counter     := 0
	togglePause.OnTapped = func() {
		if as.player.playbackActive {
			counter += 1
			nextIndex := counter % len(resources)
			togglePause.SetIcon(resources[nextIndex])
		}
    }

	// skip forward setup
	skipForwardResource := fyne.  NewStaticResource("../assets/app-assets/skip-forward.svg", []byte(skipForwardContent))
    skipForward         := widget.NewButtonWithIcon("", skipForwardResource, func() {
        fmt.Println("skipping forwards")
    })

	// shuffle button setup
	shuffleResource := fyne.  NewStaticResource("../assets/app-assets/shuffle.svg", []byte(shuffleContent))
    shuffle         := widget.NewButtonWithIcon("", shuffleResource, func() {
        fmt.Println("shuffle songs")
    })

    hbox := container.NewHBox(
        repeat, skipBack, togglePause, skipForward, shuffle,
    )

    return hbox
}

func makePlaybackView(as *AppState) fyne.CanvasObject {
    // audio cover, progress bar (name and author overlayed over progress bar)

	// TODO: make it check for a saved audio icon/image, if that is null then just display the default
	audioIcon      := widget. NewIcon(fyne.NewStaticResource("../assets/app-assets/audio-default.svg", []byte(audioDefaultContent)))
    elapsedBinding := binding.NewFloat()
    progressBar    := widget. NewProgressBarWithData(elapsedBinding)

	as.player.bindElapsedTime(elapsedBinding)

	hbox := container.NewHBox(
		audioIcon,
		progressBar,
	)

    return hbox
}

func makePlayback(as *AppState) fyne.CanvasObject {
	plbControl := makePlaybackControl(as)

    elapsedBinding := binding.NewFloat()
    progressBar    := widget. NewProgressBarWithData(elapsedBinding)

	as.player.bindElapsedTime(elapsedBinding)

	hbox := container.NewHBox(
		plbControl,
		progressBar,
	)

	return hbox
}

// Parts of the layout
// Top Bar
//   - playback controls: skip song, pause/play, repeat, shuffle
//      + makePlaybackControl() -> fyne.CanvasObject
//   - playback view: audio cover, progress bar, name of audio and author
//      + makePlaybackView() -> fyne.CanvasObject
//  + makePlayback() -> fyne.CanvasObject
// Left Side Bar
//   - button to go to sign in/sign out
//      + makeAccountButton() -> fyne.CanvasObject
//   - Lists all audio sets
//      + makeAudioSetsListView() -> fyne.CanvasObject
// Middle Content: audio set
//  + makeAudioListView() -> fyne.CanvasObject

func main() {
	a := app.New()
	w := a.NewWindow("VAMP")

    audioRaw := fetchAudioList() // TODO: MAKE THIS FASTER (caching?, lazying loading?)

    //elapsedBinding := binding.NewFloat()
    //progressBar    := widget.NewProgressBarWithData(elapsedBinding)

    var as = AppState{newPlayer(false)}
	//as.player.bindElapsedTime(elapsedBinding)

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
                as.player.playSong(audioRaw[i].Name, audioRaw[i].Author, audioRaw[i].Duration)
            }
        })

	bar        := makeSideBar(&as)
	playback   := makePlayback(&as)

    root := fyne.NewContainerWithLayout(layout.NewBorderLayout(playback, nil, bar, nil), playback, bar, list)

	w.SetContent(root)
    w.Resize(fyne.NewSize(1920, 1080))
	w.ShowAndRun()
}



// TODO
// Update layout
// Add caching system for songs, that way the fetchAudioList() operation will speed up more after subsequent uses
