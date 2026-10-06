// Command boss 是贡献分服务的场景入口。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"example.com/boss-contribution/bossscore"
)

func emit(v any) {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func main() {
	sample := flag.String("sample", "", "dupe|cap|neg|phase|work")
	flag.Parse()

	if *sample == "" {
		flag.Usage()
		os.Exit(2)
	}

	srv := bossscore.NewServer()
	ts := httptest.NewServer(srv)
	defer ts.Close()
	client := ts.Client()

	send := func(body string) {
		req, err := http.NewRequest(http.MethodPost, ts.URL+"/hit", strings.NewReader(body))
		if err != nil {
			panic(err)
		}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		resp.Body.Close()
	}
	contrib := func(player string) map[string]any {
		resp, err := client.Get(ts.URL + "/contrib?player=" + player)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		var out map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			panic(err)
		}
		return out
	}

	switch *sample {
	case "dupe":
		send(`{"player":"p1","damage":100,"hit_id":"h1","phase":1}`)
		send(`{"player":"p1","damage":100,"hit_id":"h1","phase":1}`)
		emit(contrib("p1"))
	case "cap":
		send(`{"player":"p1","damage":900,"hit_id":"h1","phase":1}`)
		emit(contrib("p1"))
	case "neg":
		send(`{"player":"p1","damage":200,"hit_id":"h1","phase":1}`)
		send(`{"player":"p1","damage":-50,"hit_id":"h2","phase":1}`)
		emit(contrib("p1"))
	case "phase":
		send(`{"player":"p1","damage":100,"hit_id":"h1","phase":1}`)
		send(`{"player":"p1","damage":200,"hit_id":"h2","phase":2}`)
		emit(contrib("p1"))
	case "work":
		for i := 0; i < 2000; i++ {
			send(fmt.Sprintf(`{"player":"p1","damage":1,"hit_id":"h%d","phase":1}`, i))
		}
		sum := 0
		for i := 0; i < 2000; i++ {
			sum += int(contrib("p1")["scanned"].(float64))
		}
		emit(map[string]int{"hits": 2000, "queries": 2000, "scanned": sum})
	default:
		fmt.Fprintln(os.Stderr, "未知场景："+*sample)
		os.Exit(2)
	}
}
