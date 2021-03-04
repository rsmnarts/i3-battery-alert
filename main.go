package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	nameService      = "i3-battery-alert"
	batWarnPercent   = 15
	batDangerPercent = 10
)

// 0 normal, 1 low, 2 critical
func sendNotify(urgency int, msg string) {
	var urgencyStr string
	switch urgency {
	case 1:
		urgencyStr = "low"
	case 2:
		urgencyStr = "critical"
	default:
		urgencyStr = "normal"
	}

	err := exec.Command("notify-send", "--urgency="+urgencyStr, "--expire-time=3000",
		nameService, msg).Start()
	if err != nil {
		log.Println(err)
	}
}

func sendNotifyErr(num int, err error) {
	msgErr := fmt.Sprintf("arts[%d]:%s", num, err)
	sendNotify(2, msgErr)
}

func sendNagbar(msg string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	msg = "\nWarning: " + msg
	if err := exec.CommandContext(ctx, "i3-nagbar", "-t", "warning", "-m", msg).Run(); err != nil {
		log.Println(err)
	}
}

func batteryAlert() {
	getDischarging, err := exec.Command("acpi", "-b").Output()
	if err != nil {
		sendNotifyErr(0, err)
		return
	}

	if strings.Contains(strings.ToLower(string(getDischarging)), "discharging") {
		percentage := regexp.MustCompile(`(\d+(\.\d+)?%)`).Find(getDischarging)
		// remaining := regexp.MustCompile(`(\d{1,2}:\d{1,2}:\d{1,2})`).Find(getDischarging)
		// fmt.Printf("%s %s\n", percentage, remaining)

		percentInt, _ := strconv.Atoi(strings.Trim(string(percentage), "%"))
		if percentInt <= batDangerPercent {
			if err := exec.Command("i3exit", "suspend").Run(); err != nil {
				sendNotifyErr(1, err)
			}
			return
		} else if percentInt <= batWarnPercent {
			sendNagbar(string(getDischarging))
			return
		} else if percentInt <= 5 {
			if err := exec.Command("i3exit", "hibernate").Run(); err != nil {
				sendNotifyErr(1, err)
			}
			return
		}
	}
}

func main() {
	sleepTime, _ := time.ParseDuration("30s")

	for {
		batteryAlert()
		time.Sleep(sleepTime)
	}
}
