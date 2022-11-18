package chanhelper

import (
	"reflect"
	"sort"
	"strconv"
	"testing"
)

func TestPeekAndReceive_int(t *testing.T) {
	ch0 := make(chan int)
	defer close(ch0)
	ch1 := make(chan int, 1)
	ch1 <- 1
	defer close(ch1)
	ch2 := make(chan int, 2)
	ch2 <- 1
	ch2 <- 2
	close(ch2)

	type args struct {
		ch <-chan int
	}
	tests := []struct {
		name      string
		args      args
		wantValue int
		wantOk    bool
		wantOpen  bool
	}{
		{name: "01", args: args{ch: nil}, wantValue: 0, wantOk: false, wantOpen: true},
		{name: "02", args: args{ch: ch0}, wantValue: 0, wantOk: false, wantOpen: true},
		{name: "11", args: args{ch: ch1}, wantValue: 1, wantOk: true, wantOpen: true},
		{name: "12", args: args{ch: ch1}, wantValue: 0, wantOk: false, wantOpen: true},
		{name: "21", args: args{ch: ch2}, wantValue: 1, wantOk: true, wantOpen: true},
		{name: "22", args: args{ch: ch2}, wantValue: 2, wantOk: true, wantOpen: true},
		{name: "23", args: args{ch: ch2}, wantValue: 0, wantOk: false, wantOpen: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotOk, gotOpen := PeekAndReceive(tt.args.ch)
			if gotValue != tt.wantValue {
				t.Errorf("PeekAndReceive() gotValue = %v, want %v", gotValue, tt.wantValue)
			}
			if gotOk != tt.wantOk {
				t.Errorf("PeekAndReceive() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			if gotOpen != tt.wantOpen {
				t.Errorf("PeekAndReceive() gotOpen = %v, want %v", gotOpen, tt.wantOpen)
			}
		})
	}
}

func TestMerge4_string(t *testing.T) {
	in1 := make(chan string)
	in2 := make(chan string)
	go func() {
		for i := 0; i < 4; i++ {
			in1 <- "one"
		}
		close(in1)
	}()
	go func() {
		for i := 0; i < 4; i++ {
			in2 <- "two"
		}
		close(in2)
	}()
	var out []string
	for v := range Merge4(in1, in2, nil, nil) {
		out = append(out, v)
	}
	sort.Strings(out)
	want := []string{"one", "one", "one", "one", "two", "two", "two", "two"}
	if !reflect.DeepEqual(out, want) {
		t.Errorf("Merge4() = %v, want %v", out, want)
	}
}

func TestMerge_0_int(t *testing.T) {
	_, ok := <-Merge[int]()
	want := false
	if ok != want {
		t.Errorf("Merge() = %v, want %v", ok, want)
	}
}

func TestMerge_1_int(t *testing.T) {
	in1 := make(chan int)
	go func() {
		for i := 0; i < 4; i++ {
			in1 <- i
		}
		close(in1)
	}()
	var out []int
	for v := range Merge(in1) {
		out = append(out, v)
	}
	sort.Ints(out)
	want := []int{0, 1, 2, 3}
	if !reflect.DeepEqual(out, want) {
		t.Errorf("Merge() = %v, want %v", out, want)
	}
}

func TestMergeInt(t *testing.T) {
	type args struct {
		ins []chan int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{name: "2",
			args: args{ins: func() []chan int {
				cc := []chan int{make(chan int), make(chan int)}
				for i, c := range cc {
					go func(i int, c chan<- int) {
						for n := 0; n < 4; n++ {
							c <- i
						}
						close(c)
					}(i, c)
				}
				return cc
			}(),
			},
			want: []int{0, 0, 0, 0, 1, 1, 1, 1},
		},
		{name: "3",
			args: args{ins: func() []chan int {
				cc := []chan int{make(chan int), make(chan int), make(chan int)}
				for i, c := range cc {
					go func(i int, c chan<- int) {
						for n := 0; n < 2; n++ {
							c <- i
						}
						close(c)
					}(i, c)
				}
				return cc
			}(),
			},
			want: []int{0, 0, 1, 1, 2, 2},
		},
		{name: "4",
			args: args{ins: func() []chan int {
				cc := []chan int{make(chan int), make(chan int), make(chan int), make(chan int)}
				for i, c := range cc {
					go func(i int, c chan<- int) {
						for n := 0; n < 2; n++ {
							c <- i
						}
						close(c)
					}(i, c)
				}
				return cc
			}(),
			},
			want: []int{0, 0, 1, 1, 2, 2, 3, 3},
		},
		{name: "5",
			args: args{ins: func() []chan int {
				cc := []chan int{make(chan int), make(chan int), make(chan int), make(chan int), make(chan int)}
				for i, c := range cc {
					go func(i int, c chan<- int) {
						for n := 0; n < 2; n++ {
							c <- i * i
						}
						close(c)
					}(i, c)
				}
				return cc
			}(),
			},
			want: []int{0, 0, 1, 1, 4, 4, 9, 9, 16, 16},
		},
		{name: "23",
			args: args{ins: func() []chan int {
				var cc []chan int
				for i := 0; i < 23; i++ {
					cc = append(cc, make(chan int))
				}
				for i, c := range cc {
					go func(i int, c chan<- int) {
						c <- i
						close(c)
					}(i, c)
				}
				return cc
			}(),
			},
			want: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []int
			for v := range Merge(tt.args.ins...) {
				got = append(got, v)
			}
			sort.Ints(got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Merge() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeBuf_int(t *testing.T) {
	c1 := make(chan int)
	go func() {
		for i := 0; i < 4; i++ {
			c1 <- i
		}
		close(c1)
	}()
	cc4 := []chan int{make(chan int), make(chan int), make(chan int), make(chan int)}
	for i, c4 := range cc4 {
		go func(i int, c chan<- int) {
			for n := 0; n < 2; n++ {
				c <- i
			}
			close(c)
		}(i, c4)
	}

	type args struct {
		cc []chan int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{name: "00", args: args{cc: nil}, want: []int{}},
		{name: "01", args: args{cc: []chan int{}}, want: []int{}},
		{name: "1", args: args{cc: []chan int{c1}}, want: []int{0, 1, 2, 3}},
		{name: "4", args: args{cc: cc4}, want: []int{0, 0, 1, 1, 2, 2, 3, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := make([]int, 0)
			for v := range MergeBuf(tt.args.cc...) {
				out = append(out, v)
			}
			sort.Ints(out)
			if !reflect.DeepEqual(out, tt.want) {
				t.Errorf("MergeBuf() = %v, want %v", out, tt.want)
			}
		})
	}
}

func TestTToAny_string(t *testing.T) {
	c0 := make(chan string)
	close(c0)
	c1 := make(chan string)
	go func() {
		for i := 0; i < 4; i++ {
			c1 <- strconv.Itoa(i)
		}
		close(c1)
	}()
	type args struct {
		chT chan string
	}
	tests := []struct {
		name string
		args args
		want []any
	}{
		{name: "0", args: args{chT: c0}, want: []any{}},
		{name: "1", args: args{chT: c1}, want: []any{"0", "1", "2", "3"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := make([]any, 0)
			for a := range TToAny(tt.args.chT) {
				out = append(out, a)
			}
			if !reflect.DeepEqual(out, tt.want) {
				t.Errorf("TToAny() = %v, want %v", out, tt.want)
			}
		})
	}
}

func TestAnyToT_int(t *testing.T) {
	c0 := make(chan any)
	close(c0)
	c1 := make(chan any)
	go func() {
		for i := 0; i < 4; i++ {
			c1 <- i
		}
		close(c1)
	}()
	type args struct {
		chAny chan any
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{name: "0", args: args{chAny: c0}, want: []int{}},
		{name: "1", args: args{chAny: c1}, want: []int{0, 1, 2, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := make([]int, 0)
			for i := range AnyToT[int](tt.args.chAny) {
				out = append(out, i)
			}
			if !reflect.DeepEqual(out, tt.want) {
				t.Errorf("AnyToT() = %v, want %v", out, tt.want)
			}
		})
	}
}
