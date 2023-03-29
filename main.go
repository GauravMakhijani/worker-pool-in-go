package main

import (
	"fmt"
	"log"

	"sync"
	"time"
)


var ErrNoWorkers = fmt.Errorf("attempting to create worker pool with less than 1 worker")
var ErrNegativeChannelSize = fmt.Errorf("attempting to create worker pool with a negative channel size")


type WorkerPool interface{
    Start()
    Stop()
    AddTask(task func())

}




type workerPool struct{
    maxWorkers  int
    queuedTasks chan func()
    // start once
    start sync.Once
    stop sync.Once
    //quit to signal worker to stop working
    quit chan struct{}
}

func NewWorkerPool(numWorkers int) (WorkerPool,error){
    if numWorkers<=0 {
        return nil,ErrNoWorkers
    }
    taskChan := make(chan func())

    return &workerPool{
        maxWorkers: numWorkers,
        queuedTasks: taskChan,
        start: sync.Once{},
        stop: sync.Once{},
        quit: make(chan struct{}),
    },nil
}

func (wp * workerPool) Start(){
    wp.start.Do(func() {
        log.Print("starting worker pool")
        wp.Run()
    })
}

func (wp *workerPool) Stop(){
    wp.stop.Do(func() {
        log.Print("stopping worker pool")
        close(wp.quit)
    })
}



func (w* workerPool) Run(){
    for i := 0; i < w.maxWorkers; i++ {
        go func (workerId int)  {
            log.Printf("starting worker %d",workerId)
            for{
                select{
                case <-w.quit:
                    log.Printf("stopping worker %d with quit channel",workerId)
                    return
                case task,ok := <-w.queuedTasks:
                    if !ok{
                        log.Printf("stopping worker %d with closed tasks channel\n", workerId)
						return
                    }

                    task()
                }
            }
        }(i+1)
    }
}


func (w *workerPool) AddTask(task func()){
    select {
    case w.queuedTasks <-task:
    case <-w.quit:
    }
}



func main(){
    var wg sync.WaitGroup


    totalWorker := 3
    wp,err := NewWorkerPool(totalWorker)
    if err != nil {
        fmt.Println(err)
        return
    }
    wp.Start()
    type result struct{
        id int
        value int
    }

    totalTask := 5
    resultC := make(chan result,totalTask)

    for i := 0; i < totalTask; i++ {
        id := i+1
        wp.AddTask(func() {
            wg.Add(1)
            log.Printf("starting task %d",id)
            time.Sleep(5 * time.Second)
            resultC<-result{id, id*2}
            log.Printf("task finished %d",id)
            wg.Done()
        })
    }
    wg.Wait()
    wp.Stop()

    for i := 0; i < totalTask; i++ {
        res := <-resultC
        log.Printf("[main] Task %d has been finished ,result = %d",res.id,res.value)
    }

}
