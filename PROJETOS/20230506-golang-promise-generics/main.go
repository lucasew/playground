package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Runnable interface {
    Run(context.Context) error
}

type Promise[ResultType any] interface {
    IsDone() bool
    Await(context.Context) (ResultType, error)
}

type PromiseFactory[ArgumentType any, ResultType any] func (f func(ArgumentType) (ResultType, error)) func(ArgumentType) Promise[ResultType]


type RuntimeJob[ArgumentType any, ResultType any] struct {
    sync.Mutex
    sync.WaitGroup

    done bool
    fn func(ArgumentType) (ResultType, error)
    argument ArgumentType
    result ResultType
    err error
}

func (rj *RuntimeJob[ArgumentType, ResultType]) Run(ctx context.Context) error {
    rj.Lock()
    defer rj.Unlock()
    if (rj.done) {
        return nil
    }
    rj.result, rj.err = rj.fn(rj.argument)
    rj.done = true
    return rj.err
}

func (rj *RuntimeJob[ArgumentType, ResultType]) IsDone() bool {
    return rj.done
}

func (rj *RuntimeJob[ArgumentType, ResultType]) Await(ctx context.Context) (ResultType, error) {
    _ = rj.Run(ctx) // if didn't started run right now otherwise wait for the background runner to run
    return rj.result, rj.err
}

func GetPromiseFactoryFromRuntime[ArgumentType any, ResultType any](runtime *Runtime) PromiseFactory[ArgumentType, ResultType] {
    return func(fn func(ArgumentType) (ResultType, error)) (func(argument ArgumentType) Promise[ResultType]) {
        return func (argument ArgumentType) Promise[ResultType] {
            promise := &RuntimeJob[ArgumentType, ResultType]{
                fn: fn,
                argument: argument,
            }
            runtime.SubmitRunnable(promise)
            return promise
        }
    }
}


type Runtime struct {
    jobs chan Runnable
    ctx context.Context
}

func NewRuntime(ctx context.Context, queueSize int) *Runtime {
    return &Runtime{
        jobs: make(chan Runnable, queueSize),
        ctx: ctx,
    }
}

func (rt *Runtime) Run(ctx context.Context) error {
    select {
        case <-rt.ctx.Done():
            return nil
        case job:=<-rt.jobs:
            job.Run(ctx)
            return nil
        case <-ctx.Done():
            return nil

    }
}

func (rt *Runtime) RunStep() bool {
    select {
        case <-rt.ctx.Done():
            return false
        case job:=<-rt.jobs:
            job.Run(rt.ctx)
            return true
    }
}

func (rt *Runtime) RunForever() {
    for rt.RunStep() {} // runs until context cancels
}

func (rt *Runtime) SubmitRunnable(runnable Runnable) {
    rt.jobs<-runnable
}

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    runtime := NewRuntime(ctx, 10)
    for i := 0; i < 5; i++ {
        go runtime.RunForever()
    }
    promiseFactory := GetPromiseFactoryFromRuntime[int, int](runtime)
    sum2actor := promiseFactory(sum2)
    sleepActor := promiseFactory(randomSleep)

    promises := []Promise[int]{
        sum2actor(2),
        sleepActor(100),
        sleepActor(200),
        sleepActor(300),
        sleepActor(300),
        sleepActor(300),
        sleepActor(300),
        sleepActor(400),
        sleepActor(500),
    }
    for _, promise := range promises {
        value, err := promise.Await(ctx)
        if err != nil {
            panic(err)
        }
        fmt.Printf("%d\n", value)
    }
}

func sum2(x int) (int, error) {
    return x + 2, nil
}

func randomSleep(x int) (int, error) {
    time.Sleep(time.Millisecond*time.Duration(x))
    return x, nil
}

