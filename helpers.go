package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2/widget"
)

type tableData struct {
	ID             int64
	Day            string
	Start          string
	End            string
	Hours          string
	StartTimeEpoch int64
	EndTimeEpoch   int64
}
type filters struct {
	yearSelect  *widget.Select
	monthSelect *widget.Select
	empSelect   *widget.Select
	totalHours  *widget.Label

	employees []string
	empMap    map[string]int64
}

func monthRangeInEpoch(yearstr string, monthstr string) (int64, int64) {
	year, _ := strconv.Atoi(yearstr)
	t, _ := time.Parse("January", monthstr)
	date := time.Date(year, t.Month(), 1, 0, 0, 0, 0, time.UTC)
	start := date.Unix()
	end := date.AddDate(0, 1, 0).Unix() - 1
	return start, end
}

func changeHours(startEntry, endEntry, hoursEntry *widget.Entry, startDatePick, endDatePick *widget.DateEntry) {

	if endEntry.Text == "--:--" || !strings.Contains(endEntry.Text, ":") {
		return
	}
	parts := strings.Split(endEntry.Text, ":")
	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		return
	}
	minute, err := strconv.Atoi(parts[1])
	if err != nil {
		return
	}
	if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return
	}
	sparts := strings.Split(startEntry.Text, ":")
	shour, err := strconv.Atoi(sparts[0])
	if err != nil {
		return
	}
	sminute, err := strconv.Atoi(sparts[1])
	if err != nil {
		return
	}
	if shour < 0 || shour > 23 || sminute < 0 || sminute > 59 {
		return
	}
	stime := time.Date(startDatePick.Date.Year(), startDatePick.Date.Month(), startDatePick.Date.Day(), shour, sminute, 0, 0, time.UTC)
	etime := time.Date(endDatePick.Date.Year(), endDatePick.Date.Month(), endDatePick.Date.Day(), hour, minute, 0, 0, time.UTC)
	dif := etime.Sub(stime)
	hoursEntry.SetText(fmt.Sprintf("%.2f", dif.Hours()))
}
func onlyDigitsValidator(s string) error {
	if s == "" {
		return nil
	}
	matched, _ := regexp.MatchString(`^\d{1,2}:\d{2}$`, s)
	if !matched {
		return fmt.Errorf("time must be in HH:MM format")
	}
	parts := strings.Split(s, ":")
	hour, _ := strconv.Atoi(parts[0])
	minute, _ := strconv.Atoi(parts[1])
	if hour < 0 || hour > 23 {
		return fmt.Errorf("hour must be 0–23")
	}
	if minute < 0 || minute > 59 {
		return fmt.Errorf("minute must be 0–59")
	}
	return nil
}
