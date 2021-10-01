package ratelimiter

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestLimiter_Limit(t *testing.T) {
	limit := uint32(5)
	l := NewLimiter(context.TODO(), limit, 5*time.Second)
	defer l.Close()
	for i:=0;i<10;i++{
		l.Limit(func() {
			fmt.Println("PASSED", i)
			time.Sleep(100*time.Millisecond)
		})
	}
}