package mcache

import (
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"testing"
)

func BenchmarkLFU_Rand(b *testing.B) {
	l, err := NewLFU(8192)
	if err != nil {
		b.Fatalf("err: %v", err)
	}

	trace := make([]int64, b.N*2)
	for i := 0; i < b.N*2; i++ {
		trace[i] = rand.Int63() % 32768
	}

	b.ResetTimer()

	var hit, miss int
	for i := 0; i < 2*b.N; i++ {
		if i%2 == 0 {
			l.Add(trace[i], trace[i], 0)
		} else {
			_, _, ok := l.Get(trace[i])
			if ok {
				hit++
			} else {
				miss++
			}
		}
	}
	b.Logf("hit: %d miss: %d ratio: %f", hit, miss, float64(hit)/float64(miss))
}

func BenchmarkLFU_Freq(b *testing.B) {
	l, err := NewLFU(8192)
	if err != nil {
		b.Fatalf("err: %v", err)
	}

	trace := make([]int64, b.N*2)
	for i := 0; i < b.N*2; i++ {
		if i%2 == 0 {
			trace[i] = rand.Int63() % 16384
		} else {
			trace[i] = rand.Int63() % 32768
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.Add(trace[i], trace[i], 0)
	}
	var hit, miss int
	for i := 0; i < b.N; i++ {
		_, _, ok := l.Get(trace[i])
		if ok {
			hit++
		} else {
			miss++
		}
	}
	b.Logf("hit: %d miss: %d ratio: %f", hit, miss, float64(hit)/float64(miss))
}

func TestLFU(t *testing.T) {
	evictCounter := 0
	onEvicted := func(k interface{}, v interface{}, expirationTime int64) {
		if k != v {
			t.Fatalf("Evict values not equal (%v!=%v)", k, v)
		}
		evictCounter++
	}
	l, err := NewLruWithEvict(128, onEvicted)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	for i := 0; i < 256; i++ {
		l.Add(i, i, 0)
	}

	if l.Len() != 128 {
		t.Fatalf("bad len: %v", l.Len())
	}

	if evictCounter != 128 {
		t.Fatalf("bad evict count: %v", evictCounter)
	}

	for i, k := range l.Keys() {
		if v, _, ok := l.Get(k); !ok || v != k || v != i+128 {
			t.Fatalf("bad key: %v", k)
		}
	}
	for i := 0; i < 128; i++ {
		_, _, ok := l.Get(i)
		if ok {
			t.Fatalf("should be evicted")
		}
	}
	for i := 128; i < 256; i++ {
		_, _, ok := l.Get(i)
		if !ok {
			t.Fatalf("should not be evicted")
		}
	}
	for i := 128; i < 192; i++ {
		l.Remove(i)
		_, _, ok := l.Get(i)
		if ok {
			t.Fatalf("should be deleted")
		}
	}

	l.Get(192) // expect 192 to be last key in l.Keys()

	for i, k := range l.Keys() {
		if (i < 63 && k != i+193) || (i == 63 && k != 192) {
			t.Fatalf("out of order key: %v", k)
		}
	}

	l.Purge()
	if l.Len() != 0 {
		t.Fatalf("bad len: %v", l.Len())
	}
	if _, _, ok := l.Get(200); ok {
		t.Fatalf("should contain nothing")
	}
}

// test that Add returns true/false if an eviction occurred
func TestLFUAdd(t *testing.T) {
	evictCounter := 0
	onEvicted := func(k interface{}, v interface{}, expirationTime int64) {
		evictCounter++
	}

	l, err := NewLfuWithEvict(1, onEvicted)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if l.Add(1, 1, 0) == false || evictCounter != 0 {
		t.Errorf("should not have an eviction")
	}
	if l.Add(2, 2, 0) == false || evictCounter != 1 {
		t.Errorf("should have an eviction")
	}
}

// test that Contains doesn't update recent-ness
func TestLFUContains(t *testing.T) {
	l, err := NewLRU(2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	l.Add(1, 1, 0)
	l.Add(2, 2, 0)
	if !l.Contains(1) {
		t.Errorf("1 should be contained")
	}

	l.Add(3, 3, 0)
	if l.Contains(1) {
		t.Errorf("Contains should not have updated recent-ness of 1")
	}
}

// test that ContainsOrAdd doesn't update recent-ness
func TestLFUContainsOrAdd(t *testing.T) {
	l, err := NewLRU(2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	l.Add(1, 1, 0)
	l.Add(2, 2, 0)
	contains, evict := l.ContainsOrAdd(1, 1, 0)
	if !contains {
		t.Errorf("1 should be contained")
	}
	if evict {
		t.Errorf("nothing should be evicted here")
	}

	l.Add(3, 3, 0)
	contains, evict = l.ContainsOrAdd(1, 1, 0)
	if contains {
		t.Errorf("1 should not have been contained")
	}
	if !evict {
		t.Errorf("an eviction should have occurred")
	}
	if !l.Contains(1) {
		t.Errorf("now 1 should be contained")
	}
}

// test that PeekOrAdd doesn't update recent-ness
func TestLFUPeekOrAdd(t *testing.T) {
	l, err := NewLRU(2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	l.Add(1, 1, 0)
	l.Add(2, 2, 0)
	previous, contains, evict := l.PeekOrAdd(1, 1, 0)
	if !contains {
		t.Errorf("1 should be contained")
	}
	if evict {
		t.Errorf("nothing should be evicted here")
	}
	if previous != 1 {
		t.Errorf("previous is not equal to 1")
	}

	l.Add(3, 3, 0)
	contains, evict = l.ContainsOrAdd(1, 1, 0)
	if contains {
		t.Errorf("1 should not have been contained")
	}
	if !evict {
		t.Errorf("an eviction should have occurred")
	}
	if !l.Contains(1) {
		t.Errorf("now 1 should be contained")
	}
}

// test that Peek doesn't update recent-ness
func TestLFUPeek(t *testing.T) {
	l, err := NewLRU(2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	l.Add(1, 1, 0)
	l.Add(2, 2, 0)
	if v, _, ok := l.Peek(1); !ok || v != 1 {
		t.Errorf("1 should be set to 1: %v, %v", v, ok)
	}

	l.Add(3, 3, 0)
	if l.Contains(1) {
		t.Errorf("should not have updated recent-ness of 1")
	}
}

// test that Resize can upsize and downsize
func TestLFUResize(t *testing.T) {
	onEvictCounter := 0
	onEvicted := func(k interface{}, v interface{}, expirationTime int64) {
		onEvictCounter++
	}
	l, err := NewLfuWithEvict(2, onEvicted)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Downsize
	l.Add(1, 1, 0)
	l.Add(2, 2, 0)
	evicted := l.Resize(1)
	if evicted != 1 {
		t.Errorf("1 element should have been evicted: %v", evicted)
	}
	if onEvictCounter != 1 {
		t.Errorf("onEvicted should have been called 1 time: %v", onEvictCounter)
	}

	l.Add(3, 3, 0)
	if l.Contains(1) {
		t.Errorf("Element 1 should have been evicted")
	}

	// Upsize
	evicted = l.Resize(2)
	if evicted != 0 {
		t.Errorf("0 elements should have been evicted: %v", evicted)
	}

	l.Add(4, 4, 0)
	if !l.Contains(3) || !l.Contains(4) {
		t.Errorf("lruCache should have contained 2 elements")
	}
}



// LFU 性能压测
func TestLFU_Performance(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	//cpu 性能分析 go tool pprof --pdf cpu ./cpu.pprof > cpu.pdf
	//开始性能分析, 返回一个停止接口
	//stopper1 := profile.Start(profile.CPUProfile, profile.ProfilePath("."))
	////在main()结束时停止性能分析
	//defer stopper1.Stop()

	//// 查看导致阻塞同步的堆栈跟踪
	//stopper2 := profile.Start(profile.BlockProfile, profile.ProfilePath("."))
	//// 在main()结束时停止性能分析
	//defer stopper2.Stop()
	//
	//// 查看当前所有运行的 goroutines 堆栈跟踪
	//stopper3 := profile.Start(profile.GoroutineProfile, profile.ProfilePath("."))
	//// 在main()结束时停止性能分析
	//defer stopper3.Stop()
	//
	//// 查看当前所有运行的 goroutines 堆栈跟踪
	//stopper4 := profile.Start(profile.MemProfile, profile.ProfilePath("."))
	//// 在main()结束时停止性能分析
	//defer stopper4.Stop()

	count := 10000000
	l, _ := NewLFU(20000)

	wg := &sync.WaitGroup{}
	for k := 0; k < count; k++ {
		wg.Add(1)
		go lfuPerformanceOne(l, wg, k)
	}
	wg.Wait()

}


func lfuPerformanceOne(h *LfuCache, c *sync.WaitGroup, k int) {

	for i := 0; i < 5; i++ {

		var strKey string
		strKey = strconv.Itoa(k) + "_" + strconv.Itoa(i)

		h.Add(strKey, &testJsonStr, 0)

	}

	// 通知main已经结束循环(我搞定了!)
	c.Done()
}





