package jav

import (
	"strings"
	"testing"
	"time"

	"golang.org/x/net/html"
)

func TestFindAvsoxSearchResultURLMatchesExactCode(t *testing.T) {
	doc, err := html.Parse(strings.NewReader(`
<!doctype html>
<html>
<body>
  <div class="container-fluid">
    <div class="row">
      <div id="waterfall">
        <div class="item">
          <a class="movie-box" href="//avsox.click/cn/movie/498391511e8b8877">
            <div class="photo-info"><span>Movie A<br><date>030919_046</date> / <date>2019-03-09</date></span></div>
          </a>
        </div>
        <div class="item">
          <a class="movie-box" href="//avsox.click/cn/movie/23cf2dcfe67622ce">
            <div class="photo-info"><span>Movie B<br><date>030919_047</date> / <date>2019-03-09</date></span></div>
          </a>
        </div>
        <div class="item">
          <a class="movie-box" href="//avsox.click/cn/movie/304fbcb19650ebce">
            <div class="photo-info"><span>Movie C<br><date>030919-874</date> / <date>2019-03-09</date></span></div>
          </a>
        </div>
      </div>
    </div>
  </div>
</body>
</html>`))
	if err != nil {
		t.Fatalf("parse html: %v", err)
	}

	got := findAvsoxSearchResultURL(doc, "030919_047", "https://avsox.click/cn/search/030919")
	if got != "https://avsox.click/cn/movie/23cf2dcfe67622ce" {
		t.Fatalf("unexpected detail url: %q", got)
	}
}

func TestFindAvsoxSearchResultURLRequiresExactCode(t *testing.T) {
	doc, err := html.Parse(strings.NewReader(`
<!doctype html>
<html>
<body>
  <div id="waterfall">
    <div class="item"><a href="//avsox.click/cn/movie/one"><date>030919_047</date> / <date>2019-03-09</date></a></div>
    <div class="item"><a href="//avsox.click/cn/movie/two"><date>030919-874</date> / <date>2019-03-09</date></a></div>
  </div>
</body>
</html>`))
	if err != nil {
		t.Fatalf("parse html: %v", err)
	}

	got := findAvsoxSearchResultURL(doc, "030919", "https://avsox.click/cn/search/030919")
	if got != "" {
		t.Fatalf("unexpected fuzzy detail url: %q", got)
	}
}

func TestParseAvsoxMovieInfoFromFixture(t *testing.T) {
	doc, err := html.Parse(strings.NewReader(avsoxDetailFixture))
	if err != nil {
		t.Fatalf("parse html: %v", err)
	}

	info := parseAvsoxMovieInfo(doc)
	if info == nil {
		t.Fatal("expected info, got nil")
	}
	if info.Provider != ProviderAvsox {
		t.Fatalf("unexpected provider: %s", info.Provider.String())
	}
	if info.IsUncensored == nil || !*info.IsUncensored {
		t.Fatal("expected avsox info to be marked uncensored")
	}
	if info.Code != "030919_047" {
		t.Fatalf("unexpected code: %q", info.Code)
	}
	if info.Title != "自分の武器をわかっている甘え上手な人妻〜何をされてもカメラ目線〜" {
		t.Fatalf("unexpected title: %q", info.Title)
	}
	if info.Studio != "パコパコママ( pacopacomama )" {
		t.Fatalf("unexpected studio: %q", info.Studio)
	}
	wantRelease := time.Date(2019, 3, 9, 0, 0, 0, 0, time.UTC).Unix()
	if info.ReleaseUnix != wantRelease {
		t.Fatalf("unexpected release unix: got %d want %d", info.ReleaseUnix, wantRelease)
	}
	if info.DurationMin != 63 {
		t.Fatalf("unexpected duration: %d", info.DurationMin)
	}

	wantTags := []string{"素人", "口交", "萝莉"}
	if len(info.Tags) != len(wantTags) {
		t.Fatalf("unexpected tags length: got %d want %d %#v", len(info.Tags), len(wantTags), info.Tags)
	}
	for i, tag := range wantTags {
		if info.Tags[i] != tag {
			t.Fatalf("unexpected tag at %d: got %q want %q", i, info.Tags[i], tag)
		}
	}

	wantActors := []string{"小橋りえこ"}
	if len(info.Actors) != len(wantActors) {
		t.Fatalf("unexpected actors length: got %d want %d", len(info.Actors), len(wantActors))
	}
	for i, actor := range wantActors {
		if info.Actors[i] != actor {
			t.Fatalf("unexpected actor at %d: got %q want %q", i, info.Actors[i], actor)
		}
	}
}

func TestParseAvsoxCoverURLFromFixture(t *testing.T) {
	doc, err := html.Parse(strings.NewReader(avsoxDetailFixture))
	if err != nil {
		t.Fatalf("parse html: %v", err)
	}

	got := parseAvsoxCoverURL(doc, "https://avsox.click/cn/movie/23cf2dcfe67622ce")
	if got != "https://file.netcdn.space/storage/pacopacomama/moviepages/030919_047/images/l_hd.jpg" {
		t.Fatalf("unexpected cover url: %q", got)
	}
}

const avsoxDetailFixture = `
<!doctype html>
<html>
<head><title>030919_047 自分の武器をわかっている甘え上手な人妻〜何をされてもカメラ目線〜 - AVSOX</title></head>
<body>
  <nav><div class="container"></div></nav>
  <div class="container">
    <h3>030919_047 自分の武器をわかっている甘え上手な人妻〜何をされてもカメラ目線〜</h3>
    <div class="row movie">
      <div class="col-md-9 screencap">
        <a class="bigImage" href="https://file.netcdn.space/storage/pacopacomama/moviepages/030919_047/images/l_hd.jpg">
          <img src="https://file.netcdn.space/storage/pacopacomama/moviepages/030919_047/images/l_hd.jpg">
        </a>
      </div>
      <div class="col-md-3 info">
        <p><span class="header">识别码:</span> <span style="color:#CC0000;">030919_047</span></p>
        <p><span class="header">发行时间:</span> 2019-03-09</p>
        <p><span class="header">长度:</span> 63分钟</p>
        <p class="header">制作商: </p>
        <p><a href="//avsox.click/cn/studio/c14e64bf5a834c44">パコパコママ<br>( pacopacomama )</a></p>
        <p class="header">类别:</p>
        <p>
          <span class="genre"><a href="//avsox.click/cn/genre/1">素人</a></span>
          <span class="genre"><a href="//avsox.click/cn/genre/2">口交</a></span>
          <span class="genre"><a href="//avsox.click/cn/genre/3">萝莉</a></span>
        </p>
      </div>
    </div>
    <h4>演员</h4>
    <div id="avatar-waterfall">
      <a class="avatar-box" href="//avsox.click/cn/star/c111eb0b8bf5eb47"><span>小橋りえこ</span></a>
    </div>
  </div>
</body>
</html>`
