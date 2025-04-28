package course

import (
	"sync/atomic"
	"time"
)
import "golang.org/x/time/rate"

var TotalQuery int32

func Hander() {
	atomic.AddInt32(&TotalQuery, 1)
	time.Sleep(50 * time.Millisecond)
}

func CallHandler() {
	limiter := rate.NewLimiter(rate.Every(100*time.Millisecond), 10)
	n := 3

	for {
		//ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		//defer cancel()
		//if err := limiter.Wait(ctx); err != nil {
		//	println("rate limit")
		//	return
		//}
		//if limiter.AllowN(time.Now(), n) {
		//	go Hander()
		//}

		reserve := limiter.ReserveN(time.Now(), n)
		time.Sleep(reserve.Delay())
		Hander()
	}
}
