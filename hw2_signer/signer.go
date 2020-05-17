package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// сюда писать код

const maxGoroutines = 6

func ExecutePipeline(jobs ...job) {
	var wg = &sync.WaitGroup{}
	var in = make(chan interface{})
	var f = func(job job, in, out chan interface{}) {
		job(in, out)
		wg.Done()
		close(out)
	}

	for _, job := range jobs {
		wg.Add(1)
		out := make(chan interface{})
		go f(job, in, out)
		in = out
	}

	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	var waiter = &sync.WaitGroup{}
	hashCh := make(chan string)

	for data := range in {
		waiter.Add(1)
		info := fmt.Sprintf("%v", data)
		hashMD5 := DataSignerMd5(info)
		go func(s1, s2 string) {

			crc32Ch := calcCRC32(s1, DataSignerCrc32)
			right32 := DataSignerCrc32(s2)
			left32 := <-crc32Ch
			hashCh <- left32 + "~" + right32

			waiter.Done()
		}(info, hashMD5)
	}

	go func() {
		waiter.Wait()
		close(hashCh)
	}()

	for hash := range hashCh {
		out <- hash
	}

}

func calcCRC32(s string, fn func(string) string) <-chan string {
	out := make(chan string, 1)
	go func(s string) { out <- fn(s) }(s)
	return out

}

func MultiHash(in, out chan interface{}) {

	var wg = &sync.WaitGroup{}

	for i := range in {
		wg.Add(1)
		go func() { defer wg.Done(); multiHash(i, out) }()
	}

	wg.Wait()
}

func multiHash(x interface{}, out chan interface{}) {
	var (
		wg  = &sync.WaitGroup{}
		arr = make([]string, maxGoroutines)
		fn  = func(i interface{}, s int, wg *sync.WaitGroup) {
			data := fmt.Sprintf("%s%v", strconv.Itoa(s),i)
			arr[s] = DataSignerCrc32(data)
			wg.Done()
		}
	)

	wg.Add(maxGoroutines)
	for i := 0; i < maxGoroutines; i++ {
		go  fn(x, i, wg)
	}
	wg.Wait()

	out <- strings.Join(arr, "")

}

func CombineResults(in, out chan interface{}) {
	var arr []string

	for i := range in {
		arr = append(arr, i.(string))
	}

	sort.Strings(arr)
	out <- strings.Join(arr, "_")

}
