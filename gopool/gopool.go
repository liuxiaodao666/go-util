package gopool

import (
	"fmt"
	"time"
)

// Task 定义了一个任务接口··
type Task func() (interface{}, error)

// Worker 是一个能够接收任务请求的并发实体
type Worker struct {
	workerPool chan chan Task
	taskQueue  chan Task
	quit       chan bool
}

// NewWorker 创建一个新的工作协程
func NewWorker(workerPool chan chan Task) *Worker {
	return &Worker{
		workerPool: workerPool,
		taskQueue:  make(chan Task),
		quit:       make(chan bool),
	}
}

// Start 方法用于启动工作协程，它会使工作协程进入监听状态，等待任务请求
func (w *Worker) Start() {
	go func() {
		for {
			// 将当前工作的任务通道注册到工作池中
			w.workerPool <- w.taskQueue
			select {
			case task := <-w.taskQueue:
				// 接收到任务后，执行任务
				result, err := task()
				if err != nil {
					fmt.Println("Error executing task:", err)
				} else {
					fmt.Println("Task result:", result)
				}
			case <-w.quit:
				// 接收到退出信号，停止工作
				return
			}
		}
	}()
}

// Stop 停止工作协程
func (w *Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

// TaskQueue 是一个任务请求通道
type TaskQueue chan Task

// Dispatcher 负责分配任务给空闲的工作协程
type Dispatcher struct {
	workerPool chan chan Task
	workers    []*Worker
}

// NewDispatcher 创建一个新的调度器
func NewDispatcher(maxWorkers int) *Dispatcher {
	pool := make(chan chan Task, maxWorkers)
	return &Dispatcher{workerPool: pool}
}

// Run 启动调度器
func (d *Dispatcher) Run(maxWorkers int) {
	d.workers = make([]*Worker, maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		worker := NewWorker(d.workerPool)
		worker.Start()
		d.workers[i] = worker
	}
}

// Submit 提交任务到调度器
func (d *Dispatcher) Submit(task Task) {
	go func() {
		d.workerPool <- task
	}()
}

// Stop 优雅地关闭调度器
func (d *Dispatcher) Stop() {
	// 关闭工作池
	close(d.workerPool)
	// 停止所有工作协程
	for _, worker := range d.workers {
		worker.Stop()
	}
}

func main() {
	dispatcher := NewDispatcher(4)
	dispatcher.Run(4)

	// 提交一些任务
	for i := 0; i < 10; i++ {
		task := Task(func() (interface{}, error) {
			return i * i, nil
		})
		dispatcher.Submit(task)
	}

	// 模拟长时间运行的服务
	fmt.Println("Service is running, press enter to stop...")
	fmt.Scanln()

	// 优雅地关闭调度器
	dispatcher.Stop()
	fmt.Println("Dispatcher stopped.")
}