package main

import (
	"container/list"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	tm "github.com/buger/goterm"
)

type FlightInfo map[string]interface{}

var (
	GogoArg = `use the G flag if you are on a united 
flight operated by a 3rd party who uses gogo-inflight wifi`

	GogoInflightWifi = "http://airborne.gogoinflight.com/portal/r/getAllSessionData"

	UnitedWifi      = "https://www.unitedwifi.com/portal/r/getAllSessionData"
	lastFlightInfos = list.New().Init()
	Color           = flag.Bool("c", false, "don't print colors")
	Metric          = flag.Bool("m", false, "use the metric system")
	Gogo            = flag.Bool("g", false, GogoArg)
)

func Print() {
	el := lastFlightInfos.Front()
	if el == nil {
		return
	}
	tm.Clear()
	tm.MoveCursor(0, 0)

	flightReport := tm.NewTable(0, 2, 5, ' ', 0)

	f := el.Value.(FlightInfo)
	var flInfo map[string]interface{}
	if flInfoGen, ok := f["flifo"]; ok {
		flInfo = flInfoGen.(map[string]interface{})
	} else {
		if isPortalInitialized, ok := f["isPortalInitialized"]; ok {
			if !isPortalInitialized.(bool) {
				fmt.Println("Your flight is still initializing...")
			}
		}
		return
	}

	titleString := "United Airlines Flight " + flInfo["flightNumber"].(string)
	titleUnderline := strings.Repeat("=", len(titleString)+1)
	tm.Println(titleString)
	tm.Println(titleUnderline)

	currAltitudeFeet, _ := strconv.Atoi(flInfo["altitudeFt"].(string))
	currAltitudeMeters, _ := strconv.Atoi(flInfo["altitudeMeters"].(string))

	currGroundSpeedMPH, _ := strconv.Atoi(flInfo["groundSpeedMPH"].(string))
	currGroundSpeedKPH, _ := strconv.Atoi(flInfo["groundSpeedKPH"].(string))

	if *Metric {
		fmt.Fprintf(flightReport, "%s\t%d m\n", "Current Altitude", currAltitudeMeters)
	} else {
		fmt.Fprintf(flightReport, "%s\t%d ft\n", "Current Altitude", currAltitudeFeet)
	}

	lastEl := lastFlightInfos.Back()
	lastFlightInfo := lastEl.Value.(FlightInfo)
	var lastFlInfo map[string]interface{}
	if lastinfo, ok := lastFlightInfo["flifo"]; ok {
		lastFlInfo = lastinfo.(map[string]interface{})
	} else {
		return
	}

	lastAltitudeFeet, _ := strconv.Atoi(lastFlInfo["altitudeFt"].(string))
	lastAltitudeMeters, _ := strconv.Atoi(lastFlInfo["altitudeMeters"].(string))
	lastGroundSpeedMPH, _ := strconv.Atoi(lastFlInfo["groundSpeedMPH"].(string))
	lastGroundSpeedKPH, _ := strconv.Atoi(lastFlInfo["groundSpeedKPH"].(string))

	deltaAltitudeFeet := currAltitudeFeet - lastAltitudeFeet
	deltaAltitudeMeters := currAltitudeMeters - lastAltitudeMeters

	deltaGroundSpeedMPH := currGroundSpeedMPH - lastGroundSpeedMPH
	deltaGroundSpeedKPH := currGroundSpeedKPH - lastGroundSpeedKPH

	if *Metric {
		fmt.Fprintf(flightReport, "%s\t%d m\n", "∆ Altitude (2 min)", deltaAltitudeMeters)
		fmt.Fprintf(flightReport, "%s\t%s\n", "", "")
		fmt.Fprintf(flightReport, "%s\t%d kph\n", "Current Ground Speed", currGroundSpeedKPH)
		fmt.Fprintf(flightReport, "%s\t%d kph\n", "∆ Ground Speed (2 min)", deltaGroundSpeedKPH)
	} else {
		fmt.Fprintf(flightReport, "%s\t%d ft\n", "∆ Altitude (2 min)", deltaAltitudeFeet)
		fmt.Fprintf(flightReport, "%s\t%s\n", "", "")
		fmt.Fprintf(flightReport, "%s\t%d mph\n", "Current Ground Speed", currGroundSpeedMPH)
		fmt.Fprintf(flightReport, "%s\t%d mph\n", "∆ Ground Speed (2 min)", deltaGroundSpeedMPH)
	}

	tm.Println(flightReport)
	// Do it twice... No idea why.
	tm.Flush()
}

func main() {

	flag.Parse()

	dataAPI := UnitedWifi
	if *Gogo {
		dataAPI = GogoInflightWifi
	}
	for {
		flightRequest, err := http.Get(dataAPI)
		if err != nil {
			log.Printf("Could not get flight data: %s\n.", err.Error())
			return
		}

		flData, err := ioutil.ReadAll(flightRequest.Body)
		if err != nil {
			log.Printf("Could not read flight data: %s\n.", err.Error())
			return
		}
		flightRequest.Body.Close()

		var flightInfo FlightInfo
		err = json.Unmarshal(flData, &flightInfo)

		if err != nil {
			log.Printf("Could not unmarshal flight info: %s\n.", err.Error())
		}

		el := lastFlightInfos.PushFront(flightInfo)
		if el == nil {
			log.Println("Push to front of list failed")
		}

		// Keep the last 24 flight infos.
		if lastFlightInfos.Len() > 24 {
			backEl := lastFlightInfos.Back()
			lastFlightInfos.Remove(backEl)
		}

		Print()

		//fmt.Println(calcPressure("-30", "2000"))
		// Call out to the server every 5 seconds.
		time.Sleep(time.Second * 5)
	}

}

//func calcPressure(temp float64, altitudeMeters float64) float64 {
//	//return 1013.25((1 - ((.0065 * altitudeMeters) / (temp + .0065*altitudeMeters + 273.15))) ^ 5.257)
//}
