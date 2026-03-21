package jav

import "errors"

// ResourceNotFonud indicates the requested resource is not available.
var ResourceNotFonud = errors.New("jav: resource not found")

// Info holds basic metadata extracted from a JavBus detail page.
type Info struct {
	Title       string
	Code        string
	ReleaseUnix int64
	DurationMin int
	Tags        []string
	Actors      []string
}

// ActressInfo describes basic actress profile fields from JavDatabase.
type ActressInfo struct {
	RomanName    string
	JapaneseName string
	ChineseName string
	HeightCM     int
	Bust         int
	Waist        int
	Hips         int
	BirthDate    int
	Cup          int
	ProfileURL   string
}

// JavLookupProvider can resolve both JAV metadata and actress profiles by code.
// Return ResourceNotFonud when the code is invalid or not found.
// Return other error for retryable lookup failures.
type JavLookupProvider interface {
	LookupActressByCode(code string) (*ActressInfo, error)
	LookupActressByJapaneseName(name string) (*ActressInfo, error)
	LookupJavByCode(code string) (*Info, error)
}
