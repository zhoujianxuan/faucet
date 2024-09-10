package v1

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"
)

type Task func()

type WorkerPool struct {
	cap              int
	runCount         int
	goCount          int
	state            int32 // 0 == Open, 1 == Close
	taskChanel       chan Task
	closeChanel      chan struct{}
	regulatoryChanel chan struct{}
}

const (
	Open  = 0
	Close = 1
)

func NewWorkerPool(cap int) *WorkerPool {
	w := &WorkerPool{
		cap:              cap,
		taskChanel:       make(chan Task, cap),
		closeChanel:      make(chan struct{}),
		regulatoryChanel: make(chan struct{}),
	}
	w.start(cap)
	go w.regulatory()
	return w
}

func (w *WorkerPool) run() {
	for {
		select {
		case task := <-w.taskChanel:
			task()
		case <-w.closeChanel:
			fmt.Println("结束退出")
			return
		}
	}
}

var ErrClosed = errors.New("state is closed")

// Submit
// 提交需要协程池运行的任务
func (w *WorkerPool) Submit(task Task) error {
	// 通过使用atomic.LoadInt32，可以确保读取操作是原子的，即在同一时间只有一个操作可以访问该变量，从而避免了并发读取时可能出现的竞态条件问题。
	if atomic.LoadInt32(&w.state) == Close {
		return ErrClosed
	}

	w.runCount++
	w.taskChanel <- task
	return nil
}

func (w *WorkerPool) CloseAll() {
	if !atomic.CompareAndSwapInt32(&w.state, Open, Close) {
		return
	}

	w.close(w.goCount)
	w.regulatoryChanel <- struct{}{}
}

func (w *WorkerPool) start(cap int) {
	for i := 0; i < cap; i++ {
		w.goCount++
		go w.run()
	}
}

func (w *WorkerPool) close(num int) {
	for i := 0; i < num; i++ {
		w.goCount--
		w.closeChanel <- struct{}{}
	}
}

func (w *WorkerPool) regulatory() {
	var t = time.NewTimer(time.Second)
	for {
		select {
		case <-t.C:
			// 下面部分可以考虑用一个定时器执行
			// 达到阈值开始新的线程
			if w.runCount/w.cap > 5 {
				w.start(w.cap)
			}

			// 检查运行情况，关闭多余部分线程
			if w.runCount/w.cap < 3 && w.goCount-w.cap > 0 {
				w.close(w.goCount - w.cap)
			}
		case <-w.regulatoryChanel:
			fmt.Println("监督协程退出")
			return
		}
	}
}
