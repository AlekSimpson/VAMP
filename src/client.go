package main

import (
	"fmt"
	"os"
	"github.com/faiface/beep/speaker"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/layout"
    "fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/data/binding"
)

type AppState struct {
	player *Player
}

func pauseAudio(as *AppState) {
    if as.player.playing {
        speaker.Lock()
        as.player.playing = false
    } else {
        speaker.Unlock()
        as.player.playing = true
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

	test := container.NewVBox(
		layout.NewSpacer(),
		bar,
	)

	return test
}

func makePlaybackControl() fyne.CanvasObject {
    // hstack with repeat, skip back, pause/play, skip forward, shuffle 
    repeat := widget.NewButton("Repeat", func() {
        fmt.Println("repeat song selected")
    }) 

    skipBack := widget.NewButton("Skip Back", func() {
        fmt.Println("skipping backwards")
    })

    togglePause := widget.NewButton("Pause", func() {
        fmt.Println("paused")
    })

    skipForward := widget.NewButton("Skip Forward", func() {
        fmt.Println("skipping forwards")
    })

    shuffle := widget.NewButton("Shuffle", func() {
        fmt.Println("shuffling")
    })

    hbox := container.NewHBox(
        repeat, skipBack, togglePause, skipForward, shuffle,
    )

    return hbox
}

func makePlaybackView() fyne.CanvasObject {
    // audio cover, progress bar (name and author overlayed over progress bar)
    return nil
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

    elapsedBinding := binding.NewFloat()
    progressBar := widget.NewProgressBarWithData(elapsedBinding)

    var as = AppState{newPlayer(false)}
	as.player.bindElapsedTime(elapsedBinding)

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

	bar := makeSideBar(&as)
    plbControl := makePlaybackControl()

    root := fyne.NewContainerWithLayout(layout.NewBorderLayout(plbControl, progressBar, bar, nil), plbControl, bar, progressBar, list)

	w.SetContent(root)
    w.Resize(fyne.NewSize(1920, 1080))
	w.ShowAndRun()
}



// TODO
// Update layout
// Add caching system for songs, that way the fetchAudioList() operation will speed up more after subsequent uses
