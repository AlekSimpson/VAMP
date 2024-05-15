// VAMP! - VAMP (is not) a Music Platform!
 package main

 import (
 	"fmt"
 	"time"
 )

 type State struct {
     cs int
     pm *ProcessMan
 }

 type Runnable interface {
 	Run()
 }

 type MyFunc struct {
 	message string
 }

 type PROC chan struct{}
 var second = time.Second

 // process manager
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
 		case <-proc:
 			<-proc
 			fmt.Println("worker is paused")
 		default:
 			runnable.Run()
 			time.Sleep(second)
 		}
 	}
 }

 func (pm *ProcessMan) toggleProcess(id int) {
 	pm.processes[id] <- struct{}{}
 }

 func (mf MyFunc) Run() {
 	fmt.Println(mf.message)
 }

 func spawn(s *State) {
     if (s.cs == 0) {
         s.pm.toggleProcess(0)
         s.pm.toggleProcess(1)
         s.cs = 1
     } else if (s.cs == 1) {
         s.cs = 0
     }

 }

 func main() {
 	var pm ProcessMan = makeProcessMan(MyFunc{message: "proc zero"}, MyFunc{message: "proc one"})
    var state = State{&pm}

    spawn(state)
    time.Sleep(3 * time.Second)
    spawn(state)

 	//go state.pm.process(0)
 	//go state.pm.process(1)
 	//state.pm.toggleProcess(1)

 	//time.Sleep(5 * time.Second)
 	//fmt.Println("Pausing proc 0 // unpausing proc 1")
 	//state.pm.toggleProcess(0)
 	//state.pm.toggleProcess(1)

 	//time.Sleep(5 * time.Second)

 	//state.pm.toggleProcess(1)
 	//fmt.Println("unpausing worker")
 	//state.pm.toggleProcess(0)
 }
