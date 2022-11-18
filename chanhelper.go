// Package chanhelper contains channel helpers.
package chanhelper

import (
	"sync"
)

// PeekAndReceive checks whether the channel has a value and receives the value if possible.
//
// If the channel has a value, the function receives the value from the channel and returns the value and 'true', 'true'.
// If the channel is open and has no value or the channel is nil, the function returns a zero value and 'false', 'true'.
// If the channel is closed and empty, the function returns a zero value and 'false', 'false'.
func PeekAndReceive[T any](ch <-chan T) (value T, ok, open bool) {
	select {
	case value, open = <-ch:
		ok = open
	default:
		ok = false
		open = true
	}
	return
}

func merge4[T any](in1, in2, in3, in4 <-chan T, out chan<- T) {
	defer close(out)
	for in1 != nil || in2 != nil || in3 != nil || in4 != nil {
		select {
		case v1, ok1 := <-in1:
			if ok1 {
				out <- v1
			} else {
				in1 = nil
			}
		case v2, ok2 := <-in2:
			if ok2 {
				out <- v2
			} else {
				in2 = nil
			}
		case v3, ok3 := <-in3:
			if ok3 {
				out <- v3
			} else {
				in3 = nil
			}
		case v4, ok4 := <-in4:
			if ok4 {
				out <- v4
			} else {
				in4 = nil
			}
		}
	}
}

// Merge4 merges four channels into a single channel to receive the values from.
// Pass corresponding nils if there are only two or three channels to merge (look at Merge function as an example).
func Merge4[T any](in1, in2, in3, in4 <-chan T) <-chan T {
	out := make(chan T)
	go merge4(in1, in2, in3, in4, out)
	return out
}

// Merge merges several channels into a single channel to receive the values from.
func Merge[T any](ins ...chan T) <-chan T {
	switch len(ins) {
	case 0:
		out := make(chan T)
		close(out)
		return out
	case 1:
		return ins[0]
	case 2:
		return Merge4(ins[0], ins[1], nil, nil)
	case 3:
		return Merge4(ins[0], ins[1], ins[2], nil)
	case 4:
		return Merge4(ins[0], ins[1], ins[2], ins[3])
	default:
		q2 := len(ins) / 2
		q1 := q2 / 2
		q3 := q2 + (len(ins)-q2)/2
		return Merge4(Merge(ins[:q1]...), Merge(ins[q1:q2]...), Merge(ins[q2:q3]...), Merge(ins[q3:]...))
	}
}

// MergeBuf merges several channels into a single channel with buffering values from the input channels
// (see https://stackoverflow.com/questions/19992334/how-to-listen-to-n-channels-dynamic-select-statement/32381409#32381409).
func MergeBuf[T any](ins ...chan T) <-chan T {
	switch len(ins) {
	case 0:
		out := make(chan T)
		close(out)
		return out
	case 1:
		return ins[0]
	default:
		out := make(chan T)
		// https://stackoverflow.com/questions/19992334/how-to-listen-to-n-channels-dynamic-select-statement/32381409#32381409
		// https://play.golang.org/p/v4M14CL7OL
		var wg sync.WaitGroup
		wg.Add(len(ins))
		for _, in := range ins {
			go func(ch <-chan T) {
				for v := range ch {
					out <- v
				}
				wg.Done()
			}(in)
		}
		go func() {
			wg.Wait()
			close(out)
		}()
		return out
	}
}

// TToAny converts chan T to chan any to receive the values from.
func TToAny[T any](chT <-chan T) <-chan any {
	chAny := make(chan any)
	go func() {
		defer close(chAny)
		for t := range chT {
			chAny <- t
		}
	}()
	return chAny
}

// AnyToT converts chan any to chan T to receive the values from.
// If some element from 'chAny' cannot be casted to type T a run-time panic occurs during receiving from the result channel.
func AnyToT[T any](chAny <-chan any) <-chan T {
	chT := make(chan T)
	go func() {
		defer close(chT)
		for a := range chAny {
			chT <- a.(T)
		}
	}()
	return chT
}
