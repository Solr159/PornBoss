package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"pornboss/internal/manager"
)

func main() {
	code := flag.String("code", "", "JAV code to fetch cover URL for (required)")
	timeout := flag.Duration("timeout", 30*time.Second, "Request timeout")
	flag.Parse()

	if *code == "" {
		log.Fatal("code is required (e.g. -code IPX-228)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	url, err := manager.FetchCoverURL(ctx, *code)
	if err != nil {
		log.Fatalf("fetch cover url: %v", err)
	}
	if url == "" {
		log.Fatalf("cover url not found for code %s", *code)
	}
	fmt.Println(url)
}
