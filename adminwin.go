package main

import (
	"context"
	"fmt"
	"slices"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/MarDoA/gotimey/internal/database"
)

func NewAdminWindow(a fyne.App, dbq *database.Queries, fs *filters, settings map[string]string) fyne.Window {
	w := a.NewWindow("Admin")
	w.Resize(fyne.NewSize(440, 840))

	now := time.Now()
	currentMonth := now.Month()
	currentYear := now.Year()
	months := []string{"January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December"}

	fs.monthSelect = widget.NewSelect(months, nil)
	fs.monthSelect.SetSelected(months[currentMonth-1])

	fs.totalHours = widget.NewLabel("Total Hours: 0.00")

	years := []string{}
	for i := currentYear - 5; i <= currentYear+1; i++ {
		years = append(years, fmt.Sprintf("%d", i))
	}
	fs.yearSelect = widget.NewSelect(years, nil)
	fs.yearSelect.SetSelected(fmt.Sprintf("%d", currentYear))

	fs.empSelect = widget.NewSelect(fs.employees, nil)

	viewTab := NewViewTab(dbq, w, fs)
	chartsTab := NewChartsTab(dbq, fs, w)
	employeeTab := NewEmployeeTab(dbq, fs, w)
	settingsTab := NewSettingsTab(dbq, w, settings)
	tabs := container.NewAppTabs(viewTab, employeeTab, chartsTab, settingsTab)
	tabs.OnSelected = func(ti *container.TabItem) {
		switch ti.Text {
		case "Charts":
			w.Resize(fyne.NewSize(900, 900))
		case "View":
			w.Resize(fyne.NewSize(440, 840))
		default:
			w.Resize(fyne.NewSize(150, 120))
		}
	}
	w.SetContent(tabs)
	return w
}

func NewEmployeeTab(dbq *database.Queries, fs *filters, w fyne.Window) *container.TabItem {
	newEmpLabel := widget.NewLabel("emp name")
	newEmpEntry := widget.NewEntry()
	addButton := widget.NewButton("Add", func() {
		empToAdd := newEmpEntry.Text
		if empToAdd != "" && !slices.Contains(fs.employees, empToAdd) {
			id, err := dbq.CreateEmployee(context.Background(), empToAdd)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			fs.empMap[empToAdd] = id
			fs.employees = append(fs.employees, empToAdd)
			fs.empSelect.Options = fs.employees
			newEmpEntry.SetText("")
			popup := widget.NewPopUp(widget.NewLabel(fmt.Sprintf("succesfully added %s", empToAdd)), w.Canvas())
			popup.Show()
			time.AfterFunc(time.Second*3, popup.Hide)
		}
	})
	addContainer := container.NewBorder(nil, nil, nil, addButton, newEmpEntry)

	spe := widget.NewSeparator()

	removeEmpLabel := widget.NewLabel("select to remove")
	removeButton := widget.NewButton("Delete", func() {
		confirm := dialog.NewConfirm("Remove Empolyee", "sure?", func(ok bool) {
			if ok {
				empToDelet := fs.empSelect.Selected
				id := fs.empMap[empToDelet]
				err := dbq.DeleteEmployee(context.Background(), int64(id))
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				delete(fs.empMap, empToDelet)
				ind := slices.Index(fs.employees, empToDelet)
				fs.employees = slices.Delete(fs.employees, ind, ind+1)
				fs.empSelect.Options = fs.employees
				fs.empSelect.SetSelectedIndex(0)
				popup := widget.NewPopUp(widget.NewLabel(fmt.Sprintf("succesfully deleted %s", empToDelet)), w.Canvas())
				popup.Show()
				time.AfterFunc(time.Second*3, popup.Hide)
			}
		}, w)
		confirm.Show()
	})
	removeContainer := container.NewBorder(nil, nil, nil, removeButton, fs.empSelect)
	vb := container.NewVBox(newEmpLabel, addContainer, spe, spe, removeEmpLabel, removeContainer)
	return container.NewTabItem("Employees", vb)
}

func NewSettingsTab(dbq *database.Queries, w fyne.Window, settings map[string]string) *container.TabItem {
	passwordLabel := widget.NewLabel("Change password")
	passwordEntry := widget.NewPasswordEntry()
	passwordButton := widget.NewButton("Confirm", func() {
		if passwordEntry.Text != "" {
			err := dbq.UpdateSet(context.Background(), database.UpdateSetParams{Value: passwordEntry.Text, Key: "password"})
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			settings["password"] = passwordEntry.Text
			popup := widget.NewPopUp(widget.NewLabel("successfully changed the password"), w.Canvas())
			popup.Show()
			time.AfterFunc(time.Second*3, popup.Hide)

		}
	})
	pasBorder := container.NewBorder(nil, nil, nil, passwordButton, passwordEntry)
	vb := container.NewVBox(passwordLabel, pasBorder)
	return container.NewTabItem("Settings", vb)
}
