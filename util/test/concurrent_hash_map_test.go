package main

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"zcygo/util"
)

var conMap = util.NewConcurrentHashMap(1000, 8)
var synMap = sync.Map{}

func readConMap() {
	for i := 0; i < 10000; i++ {
		key := strconv.Itoa(int(rand.Int63()))
		conMap.Get(key)
	}
}

func writeConMap() {
	for i := 0; i < 10000; i++ {
		key := strconv.Itoa(int(rand.Int63()))
		conMap.Set(key, i)
	}
}

func readSynMap() {
	for i := 0; i < 10000; i++ {
		key := strconv.Itoa(int(rand.Int63()))
		synMap.Load(key)
	}
}

func writeSynMap() {
	for i := 0; i < 10000; i++ {
		key := strconv.Itoa(int(rand.Int63()))
		synMap.Store(key, i)
	}
}
func BenchmarkConMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		const P = 300
		var wg sync.WaitGroup
		wg.Add(2 * P)
		for i := 0; i < P; i++ {
			go func() {
				defer wg.Done()
				readConMap()
			}()
		}
		for i := 0; i < P; i++ {
			go func() {
				defer wg.Done()
				writeConMap()
			}()
		}
	}
}

func BenchmarkSynMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		const P = 300
		var wg sync.WaitGroup
		wg.Add(2 * P)
		for i := 0; i < P; i++ {
			go func() {
				defer wg.Done()
				readSynMap()
			}()
		}

		for i := 0; i < P; i++ {
			go func() {
				defer wg.Done()
				writeSynMap()
			}()
		}
	}
}
