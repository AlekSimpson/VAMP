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

func main() {
	a := app.New()
	w := a.NewWindow("VAMP")

    audioRaw := fetchAudioList()

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

    root := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, progressBar, bar, nil), bar, progressBar, list)

	w.SetContent(root)
    w.Resize(fyne.NewSize(1920, 1080))
	w.ShowAndRun()
}

// TODO
// Update layout
// Add caching system for songs, that way the fetchAudioList() operation will speed up more after subsequent uses
