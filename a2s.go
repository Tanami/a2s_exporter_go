package main

import (
	"net/http"
	"sync"
	"time"
	"log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rumblefrog/go-a2s"
)

var servers = map[string]string{
	"ANIME LUCHSHE":    "46.174.52.164:27015",
	"ANIME LUCHSHE WL": "46.174.50.224:27015",
	"Baguette":         "148.251.11.171:27215",
	"BHOP.RIP":         "144.48.37.114:27015",
	"CN BHOP":          "183.131.85.109:27018",
	"BhopParadise":     "198.50.245.166:27015",
	"Blank WL":         "92.38.148.25:27016",
	"DHC BHOP":         "193.192.58.145:27115",
	"Dystopia AU":      "27.100.36.6:27015",
	"Dystopia NA":      "74.91.126.160:27015",
	"EUFrag BHOP":      "87.98.174.44:27017",
	"Exalted WL":       "54.38.72.50:27015",
	"fluffytail":       "74.91.119.40:27015",
	"FuckItHops":       "95.217.200.57:27015",
	"FuckItHops WL":    "95.217.200.57:27016",
	"FR ONLY BHOP":     "92.222.116.243:27015",
	"GAME4X.RU":        "46.174.51.172:27015",
	"GFLClan BHOP":     "72.5.195.96:27015",
	"H.D.F. BHOP":      "213.202.212.239:27015",
	"HyperHops":        "177.54.144.126:27622",
	"Jiminy Jilikers":  "118.210.183.239:27016",
	"JP Climb":         "126.87.115.250:27015",
	"KR AKB BHOP":      "125.186.11.253:27015",
	"KR AKB KZ":        "125.186.11.253:27016",
	"Kana":             "140.143.166.15:27018",
	"Kawaii BHOP":      "74.201.72.19:27015",
	"Kawaii WL":        "74.201.72.19:27016",
	"KwikHops NA":      "74.91.124.75:27015",
	"LacunaHops":       "104.153.108.15:27015",
	"Lets Bhop!":       "193.192.58.156:27015",
	"Mac-Infectus":     "45.235.99.134:27055",
	"mahtava":          "95.217.39.250:27060",
	"MarcoPlay BHOP":   "37.230.228.27:35000",
	"Omega-Portal":     "46.174.53.79:27015",
	"TheSourceElite":   "192.223.29.6:27015",
	"TrikzTime Cafe":   "62.122.215.209:2020",
	"Ultima Auto":      "148.251.11.171:1338",
	"Ultima Scroll":    "148.251.11.171:1336",
	"Ultima WL":        "148.251.11.171:27319",
	"XC makes me XD":   "74.91.124.58:27015",
	"]HeLL[ BHOP":      "178.32.58.203:27026",
	"]HeLL[ EZ BHOP":   "178.32.58.203:27035",
	"bhop.pro":         "148.251.234.158:27000",
	"freakhops wl":     "47.102.45.144:27115",
	"freakhops":        "47.102.45.144:27015",
	"gotta go faste":   "74.91.113.110:27015",
	"slowhops":         "162.248.88.24:27015",
	"strafersonly":     "136.243.94.194:27015",
	"strafersonly wl":  "136.243.94.194:27055",
	"STRAFE.EXPERT":    "144.48.37.118:27015",
	"StrafesOnFleek":   "192.223.27.114:27015",
	"Tarik Bhop":       "178.233.163.31:27015",
	"TRUBHOPING AUTO":  "176.212.185.161:27018",
	"UA-DV1ZH BHOP":    "176.104.57.115:27120",
	"sqf wl":           "60.111.209.149:27018",
	"Yandere BHOP":     "121.146.157.174:27017",
	"zammyhop":         "86.2.204.93:27015",
}

var players_online = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "players_online",
	Help: "The total number of players online.",
},
	[]string{
		"server",
		"address",
	},
)

var last_updated sync.Map

// for some reason running the goroutines in the for loop causes it to repeatedly run the first?
func process(name string, addr string) {
	go func() {
		client, err := a2s.NewClient(addr)
		last_updated.Store(name, time.Now())

		if err != nil {
			// don't care
		}

		defer client.Close()

		info, err := client.QueryInfo()

		if err != nil {
			// todo: do something here
			log.Printf("%s crapped itself\n", name)
			return
		}

		var real_players = float64(info.Players - info.Bots)
		players_online.With(prometheus.Labels{"server": name, "address": addr}).Set(real_players)
	}()
}

func resetOld() {
	// should probably just iterate across servers
	last_updated.Range(func(key interface{}, value interface{}) bool {
		name := key.(string)
		last_time := value.(time.Time)
		threshold := time.Now().Add(-time.Second * 30)
		if (last_time.Before(threshold)) {
			players_online.With(prometheus.Labels{"server": name, "address": servers[name]}).Set(0)
		}
		return true
	})
}

func main() {
	go func() {
		ticker := time.NewTicker(time.Second * 10)
		for ; true; <-ticker.C {
			for name, addr := range servers {
				process(name, addr)
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Second * 10)
		for ; true; <-ticker.C {
			resetOld()
		}
	}()

	// just to make promhttp shut up
	r := prometheus.NewRegistry()
	r.MustRegister(players_online)
	handler := promhttp.HandlerFor(r, promhttp.HandlerOpts{})

	http.Handle("/metrics", handler)
	http.ListenAndServe(":9110", nil)
}
