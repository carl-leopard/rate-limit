package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"

	rate "github.com/carl-leopard/rate-limit/src/rate"
	gin "github.com/gin-gonic/gin"
)

var (
	rat = rate.Limit(2)
	b   = int32(5)
	lim = rate.NewLimiter(rat, int(b))

	delay = time.Millisecond * 250

	errTooManyRequest   = errors.New("too many request")
	errUnknownAlgorithm = errors.New("unknow rate limit algorithm")

	total  int32
	errNum int32
)

func main() {
	runtime.GOMAXPROCS(1)

	r := gin.New()

	r.POST("/rate-limit", func(c *gin.Context) {
		atomic.StoreInt32(&total, atomic.AddInt32(&total, 1))

		m := c.PostForm("alg")
		rateLib(m)

		fmt.Printf("%s total: %d, err:%d, correct:%d\n", time.Now().String(), atomic.LoadInt32(&total), atomic.LoadInt32(&errNum), atomic.LoadInt32(&total)-atomic.LoadInt32(&errNum))
		time.Sleep(time.Millisecond * 200)
		c.JSON(http.StatusOK, "ok")
	})

	r.Run(":3000")
}

func rateLib(method string) {
	var err error
	switch method {
	case "reserve":
		err = reserve()
	case "allow":
		err = allow()
	case "wait":
		err = wait()
	default:
		err = errUnknownAlgorithm
	}

	if err == errTooManyRequest {
		atomic.StoreInt32(&errNum, atomic.AddInt32(&errNum, 1))
	}

	if atomic.LoadInt32(&total)%b == 0 {
		fmt.Println()
		time.Sleep(time.Second)
	}
}

func reserve() error {
	res := lim.Reserve()
	if res.Delay() <= delay {
		time.Sleep(res.Delay())
		return nil
	}

	return errTooManyRequest
}

func wait() error {
	ctx, cancel := context.WithTimeout(context.Background(), delay)
	defer cancel()
	return lim.Wait(ctx)
}

func allow() error {
	if !lim.Allow() {
		return errTooManyRequest
	}

	return nil
}
