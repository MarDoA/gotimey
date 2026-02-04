package main

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/MarDoA/gotimey/internal/database"
)

func loadData(table *widget.Table, fs *filters, data *[]tableData, dbq *database.Queries, w fyne.Window) {
	month := fs.monthSelect.Selected
	year := fs.yearSelect.Selected
	emp := fs.empSelect.Selected
	monStart, monEnd := monthRangeInEpoch(year, month)

	dbAttebdance, err := dbq.GetAttendanceByMonth(context.Background(),
		database.GetAttendanceByMonthParams{Start: monStart, Start_2: monEnd, EmpID: int64(fs.empMap[emp])})
	if err != nil {
		dialog.ShowError(err, w)
		return
	}
	tdata := []tableData{}
	var total float64

	for _, d := range dbAttebdance {
		stime := time.Unix(d.Start, 0).UTC()
		end := "--:--"
		endEpoch := int64(0)
		hours := "0.00"
		if d.End.Valid {
			shift := float64(d.Hours.Int64) / 3600
			end = time.Unix(d.End.Int64, 0).UTC().Format("15:04")
			endEpoch = d.End.Int64
			hours = fmt.Sprintf("%.2f", shift)
			total += shift
		}
		tdata = append(tdata, tableData{ID: d.ID, Day: fmt.Sprintf("%d", stime.Day()), Start: stime.Format("15:04"), End: end, Hours: hours, StartTimeEpoch: d.Start, EndTimeEpoch: endEpoch})
	}
	fs.totalHours.SetText(fmt.Sprintf("Total Hours: %.2f", total))
	*data = tdata
	table.Refresh()
}

func NewViewTab(dbq *database.Queries, w fyne.Window, fs *filters) *container.TabItem {
	var data []tableData
	var currentEmp string

	shiftsTable := widget.NewTable(
		func() (int, int) {
			return len(data), 4
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("wide content")
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			if i.Row >= len(data) {
				return
			}
			dat := data[i.Row]
			switch i.Col {
			case 0:
				label.SetText(dat.Day)
			case 1:
				label.SetText(dat.Start)
			case 2:
				label.SetText(dat.End)
			case 3:
				label.SetText(dat.Hours)
			}

		},
	)
	shiftsTable.OnSelected = func(id widget.TableCellID) {
		if id.Row >= len(data) {
			return
		}
		e := data[id.Row]
		st := time.Unix(e.StartTimeEpoch, 0).UTC()

		empEditSelect := widget.NewSelect(fs.employees, nil)
		startEntry := widget.NewEntry()
		startDatePick := widget.NewDateEntry()
		endEntry := widget.NewEntry()
		endDatePick := widget.NewDateEntry()
		hoursEntry := widget.NewEntry()
		empEditSelect.Selected = currentEmp
		startEntry.Text = e.Start
		startDatePick.SetDate(&st)
		if e.EndTimeEpoch != 0 {
			et := time.Unix(e.EndTimeEpoch, 0).UTC()
			endDatePick.SetDate(&et)
		}
		endEntry.Text = e.End
		hoursEntry.Text = e.Hours

		startEntry.Validator = onlyDigitsValidator
		endEntry.Validator = onlyDigitsValidator
		startEntry.OnChanged = func(s string) {
			if len(s) == 2 && !strings.Contains(s, ":") {
				startEntry.SetText(s + ":")
				startEntry.CursorColumn = len(startEntry.Text)
			}
			changeHours(startEntry, endEntry, hoursEntry, startDatePick, endDatePick)
		}
		endEntry.OnChanged = func(s string) {
			if len(s) == 2 && !strings.Contains(s, ":") {
				endEntry.SetText(s + ":")
				endEntry.CursorColumn = len(endEntry.Text)
			}
			changeHours(startEntry, endEntry, hoursEntry, startDatePick, endDatePick)
		}
		startDatePick.OnChanged = func(*time.Time) { changeHours(startEntry, endEntry, hoursEntry, startDatePick, endDatePick) }
		endDatePick.OnChanged = func(*time.Time) { changeHours(startEntry, endEntry, hoursEntry, startDatePick, endDatePick) }
		var d dialog.Dialog
		deletBtn := widget.NewButton("Delete", func() {
			confirm := dialog.NewConfirm("Delete Entry", "sure?", func(ok bool) {
				if ok {
					err := dbq.DeleteEntry(context.Background(), e.ID)
					if err != nil {
						dialog.ShowError(err, w)
						return
					}
					loadData(shiftsTable, fs, &data, dbq, w)
				}
			}, w)
			confirm.SetOnClosed(func() {
				d.Hide()
			})
			confirm.Show()

		})

		entryForm := &widget.Form{Items: []*widget.FormItem{
			{Text: "Name", Widget: empEditSelect},
			{Text: "Start Time (hh:mm)", Widget: startEntry},
			{Text: "Start Date", Widget: startDatePick},
			{Text: "End Time (hh:mm)", Widget: endEntry},
			{Text: "End Date", Widget: endDatePick},
			{Text: "shift in hours", Widget: hoursEntry},
		}, OnSubmit: func() {
			updateParams := database.UpdateAttendaceByIDParams{}

			st := startDatePick.Date
			sparts := strings.Split(startEntry.Text, ":")
			shour, _ := strconv.Atoi(sparts[0])
			smin, _ := strconv.Atoi(sparts[1])
			startTime := time.Date(st.Year(), st.Month(), st.Day(), shour, smin, 0, 0, time.UTC)
			updateParams.Start = startTime.Unix()
			if endEntry.Text != "--:--" {
				et := endDatePick.Date
				eparts := strings.Split(endEntry.Text, ":")
				ehour, _ := strconv.Atoi(eparts[0])
				emin, _ := strconv.Atoi(eparts[1])
				endTime := time.Date(et.Year(), et.Month(), et.Day(), ehour, emin, 0, 0, time.UTC)
				updateParams.End = sql.NullInt64{Int64: endTime.Unix(), Valid: true}
			}
			hours, _ := strconv.ParseFloat(hoursEntry.Text, 64)
			h := hours * 3600
			updateParams.Hours = sql.NullInt64{Int64: int64(h), Valid: true}
			updateParams.ID = e.ID
			updateParams.EmpID = int64(fs.empMap[empEditSelect.Selected])
			err := dbq.UpdateAttendaceByID(context.Background(), updateParams)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			d.Hide()
			loadData(shiftsTable, fs, &data, dbq, w)

		}}

		content := container.NewBorder(nil, deletBtn, nil, nil, entryForm)

		d = dialog.NewCustom("Edit Entry", "Cancel & Close", content, w)

		d.Resize(fyne.NewSize(320, 200))
		d.Show()
		shiftsTable.UnselectAll()
	}

	fs.monthSelect.OnChanged = func(string) {
		loadData(shiftsTable, fs, &data, dbq, w)
	}
	fs.yearSelect.OnChanged = func(v string) {
		loadData(shiftsTable, fs, &data, dbq, w)
	}
	fs.empSelect.OnChanged = func(v string) {
		loadData(shiftsTable, fs, &data, dbq, w)
		currentEmp = v
	}
	if len(fs.employees) != 0 {
		currentEmp = fs.employees[0]
		fs.empSelect.SetSelected(currentEmp)
	}
	selectorGrid := container.NewGridWithColumns(3, fs.monthSelect, fs.yearSelect, fs.empSelect)
	viewBorder := container.NewBorder(selectorGrid, fs.totalHours, nil, nil, shiftsTable)

	return container.NewTabItem("View", viewBorder)
}
