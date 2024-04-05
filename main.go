package main

import (
	"fmt"
	"log"

	"github.com/boltdb/bolt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Worker struct {
	ID   int
	Name string
	Type string
}

type Device struct {
	ID   int
	Host string
	Name string
}

var (
	app     = tview.NewApplication()
	pages   = tview.NewPages()
	workers = tview.NewTable()
	devices = tview.NewTable()
	db      *bolt.DB
)

func main() {
	var err error
	db, err = bolt.Open("data.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Initialize the database with dummy data
	initData()

	menu := tview.NewTextView().SetTextColor(tcell.ColorGreen).
		SetText("(a) Show Worker List\n(b) Show Device List\n(q) Quit")

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(menu, 3, 1, false).
		AddItem(tview.NewTextView(), 2, 1, false). // Add an empty text view for margin
		AddItem(pages, 0, 10, true)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'q' {
			app.Stop()
		} else if event.Rune() == 'a' {
			showWorkers()
		} else if event.Rune() == 'b' {
			showDevices()
		}
		return event
	})

	pages.AddPage("Workers", workers, true, false)
	pages.AddPage("Devices", devices, true, false)

	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func initData() {
	err := db.Update(func(tx *bolt.Tx) error {
		// Create a bucket for workers
		workersBucket, err := tx.CreateBucketIfNotExists([]byte("workers"))
		if err != nil {
			return err
		}

		// Insert dummy worker data
		for i := 1; i <= 3; i++ {
			worker := Worker{ID: i, Name: fmt.Sprintf("Worker %d", i), Type: fmt.Sprintf("Type %c", 'A'+i-1)}
			data := fmt.Sprintf("%d,%s,%s", worker.ID, worker.Name, worker.Type)
			if err := workersBucket.Put([]byte(fmt.Sprintf("w%d", i)), []byte(data)); err != nil {
				return err
			}
		}

		// Create a bucket for devices
		devicesBucket, err := tx.CreateBucketIfNotExists([]byte("devices"))
		if err != nil {
			return err
		}

		// Insert dummy device data
		for i := 1; i <= 3; i++ {
			device := Device{ID: i, Host: fmt.Sprintf("Host %d", i), Name: fmt.Sprintf("Device %d", i)}
			data := fmt.Sprintf("%d,%s,%s", device.ID, device.Host, device.Name)
			if err := devicesBucket.Put([]byte(fmt.Sprintf("d%d", i)), []byte(data)); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func showWorkers() {
	workersData := fetchWorkers()
	displayTable(workers, workersData)
	pages.SwitchToPage("Workers")
}

func showDevices() {
	devicesData := fetchDevices()
	displayTable(devices, devicesData)
	pages.SwitchToPage("Devices")
}

func fetchWorkers() []Worker {
	var workersData []Worker
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("workers"))
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}
		return bucket.ForEach(func(k, v []byte) error {
			var worker Worker
			fmt.Sscanf(string(v), "%d,%s,%s", &worker.ID, &worker.Name, &worker.Type)
			workersData = append(workersData, worker)
			return nil
		})
	})
	if err != nil {
		log.Fatal(err)
	}
	return workersData
}

func fetchDevices() []Device {
	var devicesData []Device
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("devices"))
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}
		return bucket.ForEach(func(k, v []byte) error {
			var device Device
			fmt.Sscanf(string(v), "%d,%s,%s", &device.ID, &device.Host, &device.Name)
			devicesData = append(devicesData, device)
			return nil
		})
	})
	if err != nil {
		log.Fatal(err)
	}
	return devicesData
}

func displayTable(table *tview.Table, data interface{}) {
	table.Clear()

	switch data := data.(type) {
	case []Worker:
		for i, col := range []string{"ID", "Name", "Type"} {
			table.SetCell(0, i, tview.NewTableCell(col).SetTextColor(tcell.ColorYellow))
		}
		for r, w := range data {
			table.SetCell(r+1, 0, tview.NewTableCell(fmt.Sprintf("%d", w.ID)))
			table.SetCell(r+1, 1, tview.NewTableCell(w.Name))
			table.SetCell(r+1, 2, tview.NewTableCell(w.Type))
		}
	case []Device:
		for i, col := range []string{"ID", "Host", "Name"} {
			table.SetCell(0, i, tview.NewTableCell(col).SetTextColor(tcell.ColorYellow))
		}
		for r, d := range data {
			table.SetCell(r+1, 0, tview.NewTableCell(fmt.Sprintf("%d", d.ID)))
			table.SetCell(r+1, 1, tview.NewTableCell(d.Host))
			table.SetCell(r+1, 2, tview.NewTableCell(d.Name))
		}
	}
}
