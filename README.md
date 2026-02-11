# Gotimey
Gotimey is a simple clock-in app build with fyne(https://fyne.io/) in Go backed by a local SQLite database.

![Clockin](https://github.com/user-attachments/assets/0820c9d5-d4db-4c2a-8272-1331fcad9827)


## Motivation
This project was made to save time for me and my employees, instead of manually writing down hours and calculating the total monthly hours of each one.

## Quick Start
Download the latest realse
Unzip and run gitimey.exe
the app will create its own database in the same dir as the exe.

## Usage
Admin access by default is 1234 you can change it from the setting tab.
You can edit any saved record by clicking on any cell of that record row.
You can also easily visualize monthly hours of everyone
![Chart](https://github.com/user-attachments/assets/afc69a96-5885-47c7-b5c7-060df00228dd)

## Contributing

### Clone the repo

```bash
git clone https://github.com/MarDoA/gotimey@latest
cd gotimey
```

### Build the compiled binary

```bash
go build -o gotimey
```
### Cross-compile for another OS (requires Docker + fyne-cross)
fyne-cross windows
fyne-cross linux
fyne-cross darwin
