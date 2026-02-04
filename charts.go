// WIP
package main

import (
	"context"
	"database/sql"
	"fmt"
	"image/color"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/MarDoA/gotimey/internal/database"
)

func NewChartsTab(dbq *database.Queries, fs *filters, w fyne.Window) *container.TabItem {
	chartContainer := container.NewVBox()
	ySelect := widget.NewSelect(fs.yearSelect.Options, nil)
	mSelect := widget.NewSelect(fs.monthSelect.Options, nil)
	updateChart := func() {
		month := mSelect.Selected
		year := ySelect.Selected
		y, _ := strconv.Atoi(year)
		monStart, monEnd := monthRangeInEpoch(year, month)

		allData := make(map[string][]database.Attendance)

		for _, empName := range fs.employees {
			empID := int64(fs.empMap[empName])
			entries, err := dbq.GetAttendanceByMonth(context.Background(),
				database.GetAttendanceByMonthParams{
					Start:   monStart,
					Start_2: monEnd,
					EmpID:   empID,
				})
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			allData[empName] = entries
		}

		chart := createTimelineChart(allData, y, month)
		chartContainer.Objects = []fyne.CanvasObject{chart}
		chartContainer.Refresh()
	}
	ySelect.OnChanged = func(s string) { updateChart() }
	mSelect.OnChanged = func(s string) { updateChart() }
	ySelect.SetSelected(fs.yearSelect.Selected)
	mSelect.SetSelected(fs.monthSelect.Selected)

	gr := container.NewGridWithColumns(2, ySelect, mSelect)

	content := container.NewBorder(
		gr,
		nil, nil, nil,
		chartContainer,
	)

	updateChart()

	return container.NewTabItem("Charts", content)
}

func createTimelineChart(data map[string][]database.Attendance, year int, month string) fyne.CanvasObject {
	t, _ := time.Parse("January", month)
	daysInMonth := time.Date(year, t.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()

	hourWidth := float32(30.0)
	dayHeight := float32(25)
	labelWidth := float32(30.0)
	headerHeight := float32(25.0)

	totalWidth := labelWidth + (24 * hourWidth)
	totalHeight := headerHeight + (float32(daysInMonth) * dayHeight)

	chart := container.NewWithoutLayout()

	bg := canvas.NewRectangle(color.RGBA{250, 250, 250, 255})
	bg.Resize(fyne.NewSize(totalWidth, totalHeight))
	chart.Add(bg)

	for hour := 0; hour <= 24; hour++ {
		x := labelWidth + (float32(hour) * hourWidth)
		line := canvas.NewLine(color.RGBA{200, 200, 200, 255})
		line.StrokeWidth = 1
		line.Position1 = fyne.NewPos(x, 0)
		line.Position2 = fyne.NewPos(x, totalHeight)
		chart.Add(line)

		if hour < 24 {
			label := canvas.NewText(fmt.Sprintf("%d", hour), color.Black)
			label.TextSize = 10
			label.Move(fyne.NewPos(x+8, 5))
			chart.Add(label)
		}
	}
	entriesByDay := make(map[int][]struct {
		empName string
		entry   database.Attendance
		color   color.Color
	})
	empColors := map[string]color.Color{}
	colors := []color.Color{
		color.RGBA{100, 149, 237, 200},
		color.RGBA{60, 179, 113, 200},
		color.RGBA{255, 140, 0, 200},
		color.RGBA{186, 85, 211, 200},
		color.RGBA{220, 20, 60, 200},
	}
	colorIndex := 0

	for empName, entries := range data {
		if _, ok := empColors[empName]; !ok {
			empColors[empName] = colors[colorIndex%len(colors)]
			colorIndex++
		}

		for _, entry := range entries {
			if !entry.End.Valid {
				continue
			}
			startTime := time.Unix(entry.Start, 0).UTC()
			endTime := time.Unix(entry.End.Int64, 0).UTC()
			if startTime.Day() != endTime.Day() {
				currentTime := startTime
				for currentTime.Before(endTime) {
					dayEnd := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 23, 59, 59, 0, time.UTC)
					segmentEnd := endTime
					if endTime.After(dayEnd) {
						segmentEnd = dayEnd
					}
					entriesByDay[currentTime.Day()] = append(entriesByDay[currentTime.Day()], struct {
						empName string
						entry   database.Attendance
						color   color.Color
					}{empName, database.Attendance{
						ID: entry.ID, EmpID: entry.EmpID, Start: currentTime.Unix(), End: sql.NullInt64{Int64: segmentEnd.Unix(), Valid: true}, Hours: sql.NullInt64{Int64: int64(segmentEnd.Sub(currentTime).Seconds()), Valid: true},
					}, empColors[empName]})
					currentTime = currentTime.AddDate(0, 0, 1)
				}
			} else {
				day := startTime.Day()
				entriesByDay[day] = append(entriesByDay[day], struct {
					empName string
					entry   database.Attendance
					color   color.Color
				}{empName, entry, empColors[empName]})
			}
		}
	}

	for day := 1; day <= daysInMonth; day++ {
		y := headerHeight + (float32(day-1) * dayHeight)

		nameLabel := canvas.NewText(fmt.Sprintf("%d", day), color.Black)
		nameLabel.TextSize = 12
		nameLabel.TextStyle = fyne.TextStyle{Bold: true}
		nameLabel.Move(fyne.NewPos(10, y+6))
		chart.Add(nameLabel)

		separator := canvas.NewLine(color.RGBA{180, 180, 180, 255})
		separator.StrokeWidth = 1
		separator.Position1 = fyne.NewPos(0, y)
		separator.Position2 = fyne.NewPos(totalWidth, y)
		chart.Add(separator)

		shifts := entriesByDay[day]
		barHeight := (dayHeight - 10) / float32(len(shifts)+1)
		if barHeight > 25 {
			barHeight = 25
		}

		for i, entry := range shifts {
			startTime := time.Unix(entry.entry.Start, 0).UTC()
			endTime := time.Unix(entry.entry.End.Int64, 0).UTC()

			startHour := float32(startTime.Hour()) + float32(startTime.Minute())/60.0
			endHour := float32(endTime.Hour()) + float32(endTime.Minute())/60.0

			if endHour < startHour {
				endHour = 24
			}

			x1 := labelWidth + (startHour * hourWidth)
			x2 := labelWidth + (endHour * hourWidth)
			width := x2 - x1

			barY := y + (float32(i) * (barHeight + 2))

			bar := canvas.NewRectangle(entry.color)
			bar.Move(fyne.NewPos(x1, barY))
			bar.Resize(fyne.NewSize(width, barHeight))
			chart.Add(bar)

			empText := canvas.NewText(entry.empName, color.White)
			empText.TextSize = 9
			empText.Move(fyne.NewPos(x1+3, barY+5))
			chart.Add(empText)

		}
	}
	legendY := headerHeight
	legendX := float32(750)

	legendBg := canvas.NewRectangle(color.RGBA{255, 255, 255, 230})
	legendBg.Move(fyne.NewPos(legendX-5, legendY-5))
	legendBg.Resize(fyne.NewSize(145, float32(len(empColors)*25+15)))
	chart.Add(legendBg)

	i := 0
	for empName, empColor := range empColors {
		box := canvas.NewRectangle(empColor)
		box.Move(fyne.NewPos(legendX, legendY+float32(i*25)))
		box.Resize(fyne.NewSize(15, 15))
		chart.Add(box)

		label := canvas.NewText(empName, color.Black)
		label.TextSize = 11
		label.Move(fyne.NewPos(legendX+20, legendY+float32(i*25)))
		chart.Add(label)

		i++
	}
	chart.Resize(fyne.NewSize(totalWidth, totalHeight))

	return chart
}
