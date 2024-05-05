package main

import (
	"fmt"
	"time"
)

type PROC chan struct{}
var second = time.Second

type ApplicationInterface struct {
	action int
}

type Runnable interface {
	Run()
}

type ProcessMan struct {
	processes []PROC
	runnables []Runnable
}

func makeProcessMan(runs ...Runnable) ProcessMan {
	size := len(runs)
	procs := make([]PROC, size)
	runnables := make([]Runnable, size)
	for i := 0; i < size; i++ {
		procs[i] = make(PROC)
		runnables[i] = runs[i]
	}
	return ProcessMan{procs, runnables}
}

func (pm *ProcessMan) process(id int) {
	proc := pm.processes[id]
	runnable := pm.runnables[id]
	for {
		select {
		case <- proc:
			<-proc
			fmt.Println("[LOG] worker is paused")
		default:
			runnable.Run()
			time.Sleep(second)
		}
	}
}

func (pm *ProcessMan) toggleProcess(id int) {
	pm.processes[id] <- struct{}{}
}