package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/l4go/task"
	"github.com/l4go/vqueue"
)

func free(i interface{}) {
	v := i.(int)
	log.Println("Free:", v)
}

func main() {
	signal_ch := make(chan os.Signal, 1)
	signal.Notify(signal_ch, syscall.SIGINT, syscall.SIGTERM)
	cc := task.NewCancel()
	defer cc.Cancel()

	go func() {
		select {
		case <-signal_ch:
			cc.Cancel()
		case <-cc.RecvCancel():
		}
	}()

	que := vqueue.New(free)
	defer que.Close()

	log.Printf("start PopOrTimeout()")
	res, ok, tout := que.PopOrTimeout(2 * time.Second)
	log.Printf("PopOrTimeout(): %v %v %v", res, ok, tout)

	do := func(id int) {
		for {
			res, ok := que.PopWithCancel(cc)
			if task.IsCanceled(cc) {
				return
			}
			if !ok {
				break
			}

			log.Printf("Pop(%d): %v", id, res.(int))
			free(res)
			time.Sleep(500 * time.Millisecond)
		}
	}

	do_tout := func(id int) {
		for {
			res, ok, _ := que.PopOrTimeout(time.Second)
			if task.IsCanceled(cc) {
				break
			}
			if !ok {
				break
			}

			log.Printf("PopOrTimeout(%d): %v", id, res.(int))
			free(res)
			time.Sleep(500 * time.Millisecond)
		}
	}

	go do(1)
	go do(2)
	go do_tout(3)

	for i := 0; i < 10; i++ {
		if task.IsCanceled(cc) {
			break
		}
		que.Push(i)
		log.Println("Push:", i)
		time.Sleep(100 * time.Millisecond)
	}
}
