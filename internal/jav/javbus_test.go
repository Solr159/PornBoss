package jav

import (
	"context"
	"errors"
	"testing"
	"time"
)

func resetJavBusRateLimiterForTest() {
	javBusRateLimiter.Lock()
	javBusRateLimiter.next = time.Time{}
	javBusRateLimiter.Unlock()
}

func TestJavBusRateLimiterSpacesRequests(t *testing.T) {
	resetJavBusRateLimiterForTest()
	t.Cleanup(resetJavBusRateLimiterForTest)

	start := time.Now()
	for i := 0; i < 5; i++ {
		if err := waitForJavBusRateLimit(context.Background()); err != nil {
			t.Fatalf("waitForJavBusRateLimit() request %d: %v", i+1, err)
		}
	}

	if elapsed := time.Since(start); elapsed < (4*javBusRequestInterval - 50*time.Millisecond) {
		t.Fatalf("rate limiter allowed 5 requests in %s", elapsed)
	}
}

func TestJavBusRateLimiterHonorsContext(t *testing.T) {
	resetJavBusRateLimiterForTest()
	t.Cleanup(resetJavBusRateLimiterForTest)

	javBusRateLimiter.Lock()
	javBusRateLimiter.next = time.Now().Add(time.Hour)
	javBusRateLimiter.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := waitForJavBusRateLimit(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("waitForJavBusRateLimit() err = %v, want context deadline exceeded", err)
	}
}
