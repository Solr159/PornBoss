package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"pornboss/internal/jav"
	"pornboss/internal/util"
)

func main() {
	var filename string
	flag.StringVar(&filename, "file", "", "video filename (e.g. MBMA-143.mp4)")
	flag.Parse()

	if filename == "" && flag.NArg() > 0 {
		filename = flag.Arg(0)
	}
	if filename == "" {
		log.Fatal("usage: go run ./cmd/javbuslookup -file MBMA-143.mp4")
	}

	possibleCodes := util.ExtractCodeFromName(filename)
	var (
		info    *jav.Info
		code    string
		lastErr error
	)
	for _, candidate := range possibleCodes {
		candidateInfo, err := jav.JavBusProvider.LookupJavByCode(candidate)
		if err != nil {
			if errors.Is(err, jav.ResourceNotFonud) {
				continue
			}
			lastErr = err
			continue
		}
		if candidateInfo != nil {
			info = candidateInfo
			code = candidate
			break
		}
	}
	if info == nil {
		if lastErr != nil {
			log.Fatalf("lookup error: %v", lastErr)
		}
		fmt.Println("not a JavBus title or not found")
		return
	}
	if info.Code == "" && code != "" {
		info.Code = code
	}

	tags := strings.Join(info.Tags, ", ")
	if tags == "" {
		tags = "None"
	}
	actors := strings.Join(info.Actors, ", ")
	if actors == "" {
		actors = "None"
	}
	duration := "unknown"
	if info.DurationMin > 0 {
		duration = fmt.Sprintf("%d min", info.DurationMin)
	}
	release := "unknown"
	if info.ReleaseUnix > 0 {
		release = time.Unix(info.ReleaseUnix, 0).UTC().Format("2006-01-02")
	}

	fmt.Printf("Code: %s\n", info.Code)
	fmt.Printf("Title: %s\n", info.Title)
	fmt.Printf("Release: %s\n", release)
	fmt.Printf("Duration: %s\n", duration)
	fmt.Printf("Tags: %s\n", tags)
	fmt.Printf("Actors: %s\n", actors)
}
