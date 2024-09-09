package v1

import (
	"fmt"
	"time"
)

type Task func()

type WorkerPool struct {
	cap              int
	runCount         int
	goCount          int
	taskChanel       chan Task
	closeChanel      chan struct{}
	regulatoryChanel chan struct{}
}

func NewWorkerPool(cap int) *WorkerPool {
	w := &WorkerPool{
		cap:              cap,
		taskChanel:       make(chan Task, cap),
		closeChanel:      make(chan struct{}),
		regulatoryChanel: make(chan struct{}),
	}
	w.start(cap)
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

// Submit
// 提交需要协程池运行的任务
func (w *WorkerPool) Submit(task Task) {
	w.runCount++
	w.taskChanel <- task
}

func (w *WorkerPool) CloseAll() {
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
