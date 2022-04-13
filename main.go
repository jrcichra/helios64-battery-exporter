package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//https://wiki.kobol.io/helios64/ups/#main-power-status
var mainPowerStatus = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "helios64_main_power_status",
		Help: "1 means power adapter supplying the power. 0 means loss of power, power adapter no longer supplying power.",
	},
)

//https://wiki.kobol.io/helios64/ups/#charging-status
var batteryChargingStatus = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "helios64_battery_charging_status",
		Help: "0 = Not Charging, 1 = Charging, -1 = Error",
	},
)

//https://wiki.kobol.io/helios64/ups/#battery-level
var batteryLevel = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "helios64_battery_level",
		Help: "0V 114mV = No batteries, 7V 916mw Recommended threshold to force shutdown system, 8.4V 1099mV Fully Charged",
	},
)

func setMainPowerStatus() {
	bytes, err := os.ReadFile("/sys/class/power_supply/gpio-charger/online")
	if err != nil {
		log.Fatal(err)
	}
	parsed, err := strconv.ParseFloat(strings.TrimSpace(string(bytes)), 64)
	if err != nil {
		log.Fatal(err)
	}
	mainPowerStatus.Set(parsed)
}

func setChargingStaus() {
	bytes, err := os.ReadFile("/sys/class/power_supply/gpio-charger/status")
	if err != nil {
		log.Fatal(err)
	}

	val := -1.0

	switch strings.ToLower(strings.TrimSpace(string(bytes))) {
	case "charging":
		val = 1
	case "not charging":
		val = 0
	}

	batteryChargingStatus.Set(val)
}

func setBatteryLevel() {
	rawBytes, err := os.ReadFile("/sys/bus/iio/devices/iio:device0/in_voltage2_raw")
	if err != nil {
		log.Fatal(err)
	}

	scaleBytes, err := os.ReadFile("/sys/bus/iio/devices/iio:device0/in_voltage_scale")
	if err != nil {
		log.Fatal(err)
	}

	raw, err := strconv.ParseFloat(strings.TrimSpace(string(rawBytes)), 64)
	if err != nil {
		log.Fatal(err)
	}

	scale, err := strconv.ParseFloat(strings.TrimSpace(string(scaleBytes)), 64)
	if err != nil {
		log.Fatal(err)
	}

	adc := raw * scale

	batteryLevel.Set(adc)

}

func collect() {
	for {
		setMainPowerStatus()
		setChargingStaus()
		setBatteryLevel()
		time.Sleep(time.Second * 5)
	}
}

func main() {
	go collect()
	prometheus.MustRegister(mainPowerStatus)
	prometheus.MustRegister(batteryChargingStatus)
	prometheus.MustRegister(batteryLevel)

	r := mux.NewRouter()
	r.Handle("/metrics", promhttp.Handler())

	addr := ":8080"
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	log.Print("Starting server at ", addr)
	log.Fatal(srv.ListenAndServe())
}
