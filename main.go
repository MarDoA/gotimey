package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	"github.com/MarDoA/gotimey/internal/database"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

//go:embed sql/schema/*.sql
var embedMigrations embed.FS

func initDB(dbPath string) (*sql.DB, error) {
	// Open/create database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Set goose to use embedded migrations
	goose.SetBaseFS(embedMigrations)

	// Run migrations
	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, err
	}

	if err := goose.Up(db, "sql/schema"); err != nil {
		return nil, err
	}

	return db, nil
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("goTimey")
	myWindow.Resize(fyne.NewSize(250, 120))
	db, err := initDB("./clockin.db")
	if err != nil {
		dialog.ShowError(err, myWindow)
	}

	defer db.Close()
	dbq := database.New(db)

	settingsdb, err := dbq.GetSettings(context.Background())
	if err != nil {
		dialog.ShowError(err, myWindow)
		return
	}
	settings := make(map[string]string)
	for _, s := range settingsdb {
		settings[s.Key] = s.Value
	}
	_, ok := settings["password"]
	if !ok {
		err := dbq.AddSet(context.Background(), database.AddSetParams{Key: "password", Value: "1234"})
		if err != nil {
			dialog.ShowError(err, myWindow)
		}
		settings["password"] = "1234"
	}
	dbEmployees, err := dbq.GetEmployees(context.Background())
	if err != nil {
		dialog.ShowError(err, myWindow)
	}
	var fs filters
	fs.employees = make([]string, len(dbEmployees))
	fs.empMap = make(map[string]int64)
	for i, emp := range dbEmployees {
		fs.employees[i] = emp.Name
		fs.empMap[emp.Name] = emp.ID
	}
	currentId := int64(0)
	currentEmp := ""
	sessionID := int64(0)
	clockInButton := widget.NewButton("clock in", nil)
	clockOutButton := widget.NewButton("clock out", nil)
	employeeSelect := widget.NewSelect(fs.employees, nil)

	clockInButton.OnTapped = func() {
		_, err := dbq.CreateAttendanceStart(context.Background(), database.CreateAttendanceStartParams{
			Start: time.Now().UTC().Unix(), EmpID: currentId})
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		employeeSelect.SetSelected(currentEmp)
	}
	clockOutButton.OnTapped = func() {
		etime := time.Now().UTC().Unix()
		s, err := dbq.AddEndToAttendaceByID(context.Background(),
			database.AddEndToAttendaceByIDParams{
				End: sql.NullInt64{Int64: etime, Valid: true}, ID: int64(sessionID), Start: etime})
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		diff := float64(s.Hours.Int64) / 3600.00
		stri := fmt.Sprintf("worked for %.2f hours", diff)
		employeeSelect.SetSelected(currentEmp)
		popup := widget.NewPopUp(widget.NewLabel(stri), myWindow.Canvas())
		popup.Show()
		time.AfterFunc(time.Second*3, popup.Hide)
	}
	active := true
	employeeSelect.OnChanged = func(value string) {
		currentId = fs.empMap[value]
		currentEmp = value
		sessionID, err = dbq.GetActiveSessionByEmpID(context.Background(), currentId)
		if err != nil {
			if err == sql.ErrNoRows {
				active = false
			} else {
				dialog.ShowError(err, myWindow)
				return
			}
		} else {
			active = true
		}
		if active {
			clockInButton.Disable()
			clockOutButton.Enable()
		} else {
			clockInButton.Enable()
			clockOutButton.Disable()
		}
	}

	clockInButton.Disable()
	clockOutButton.Disable()
	buttonsGrid := container.New(layout.NewGridLayout(2), clockInButton, clockOutButton)

	adminlabel := widget.NewLabel("Password")
	input := widget.NewPasswordEntry()
	adminBtn := widget.NewButton("Enter", func() {
		if input.Text == settings["password"] {
			adminWin := NewAdminWindow(myApp, dbq, &fs, settings)
			adminWin.Show()
			myWindow.Hide()
			adminWin.SetOnClosed(func() {
				employeeSelect.Options = fs.employees
				myWindow.Show()
			})

		} else {
			popup := widget.NewPopUp(widget.NewLabel("Wrong password"), myWindow.Canvas())
			popup.Show()
			time.AfterFunc(time.Second*3, popup.Hide)
		}
		input.SetText("")
	})
	loginHBox := container.NewVBox(employeeSelect, buttonsGrid)
	clockTab := container.NewTabItem("Clock-In", loginHBox)
	adminHBox := container.NewVBox(adminlabel, input, adminBtn)
	adminTab := container.NewTabItem("Admin", adminHBox)
	tabs := container.NewAppTabs(clockTab, adminTab)

	myWindow.SetContent(tabs)

	myWindow.ShowAndRun()
}
