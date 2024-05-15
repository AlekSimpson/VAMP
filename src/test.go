package main

import (
    "fmt"
    "sync"
    "time"
)

// Song represents a song to be played
type Song struct {
    Author     string `json:"author"`
	//Set      string
    Name       string `json:"name"`
    Duration   int64  `json:"duration"`
}

// MusicPlayer represents a music player
type MusicPlayer struct {
    currentSong *Song
    control     chan struct{}
    wg          sync.WaitGroup
}

// NewMusicPlayer creates a new music player
func NewMusicPlayer() *MusicPlayer {
    return &MusicPlayer{
        control: make(chan struct{}),
    }
}

// Play starts playing the given song
func (mp *MusicPlayer) Play(song *Song) {
}

// Stop stops the current playback
func (mp *MusicPlayer) Stop() {
}

// playback simulates playing a song
func (mp *MusicPlayer) playback() {
    fmt.Printf("Playing %s\n", song.Name)
    for {
        select {
        case <-mp.control:
            return
        default:
            // Simulate playing the song
        }
    }
}

func main() {

}

