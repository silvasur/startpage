package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"time"
)

var porn EarthPorn
var weather Weather
var sun Sun

func trylater(ch chan<- bool) {
	log.Println("Will try again later...")
	time.Sleep(1 * time.Minute)
	ch <- true
}

func earthPornUpdater(ch chan bool) {
	for _ = range ch {
		newporn, err := GetEarthPorn()
		if err != nil {
			log.Printf("Failed getting fap material: %s", err)
			go trylater(ch)
		}

		porn = newporn
		log.Println("New fap material!")
	}
}

func weatherUpdater(ch chan bool) {
	for _ = range ch {
		newW, newS, err := CurrentWeather()
		if err != nil {
			log.Printf("Failed getting latest weather data: %s", err)
			go trylater(ch)
		}

		weather = newW
		sun = newS
		log.Println("New weather data")
	}
}

func intervalUpdates(d time.Duration, stopch <-chan bool, chans ...chan<- bool) {
	send := func(chans ...chan<- bool) {
		for _, ch := range chans {
			go func(ch chan<- bool) {
				ch <- true
			}(ch)
		}
	}

	send(chans...)

	tick := time.NewTicker(d)
	for {
		select {
		case <-tick.C:
			send(chans...)
		case <-stopch:
			tick.Stop()
			for _, ch := range chans {
				close(ch)
			}
			return
		}
	}
}

func main() {
	laddr := flag.String("laddr", ":25145", "Listen on this port")
	flag.Parse()

	pornch := make(chan bool)
	weatherch := make(chan bool)
	stopch := make(chan bool)

	go intervalUpdates(30*time.Minute, stopch, pornch, weatherch)
	go weatherUpdater(weatherch)
	go earthPornUpdater(pornch)

	defer func(stopch chan<- bool) {
		stopch <- true
	}(stopch)

	http.HandleFunc("/", startpage)
	log.Fatal(http.ListenAndServe(*laddr, nil))
}

var tpl = template.Must(template.ParseFiles("template.html"))

type TplData struct {
	Porn    *EarthPorn
	Weather *Weather
	Links   []Link
	LCols   int
}

func startpage(rw http.ResponseWriter, req *http.Request) {
	links := GetLinks()
	lcols := len(links) / 3
	if lcols < 1 {
		lcols = 1
	} else if lcols > 4 {
		lcols = 4
	}

	if err := tpl.Execute(rw, &TplData{&porn, &weather, links, lcols}); err != nil {
		log.Printf("Failed executing template: %s\n", err)
	}
}
