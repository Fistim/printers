package main

import (
	"fmt"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"html/template"
	"image"
	"image/png"
	"log"
	"net/http"
	"io/ioutil"
	"os"
	"strings"
	"math/rand"
	"time"
)

func init() {
    rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func main() {
	port := os.Getenv("PORT")
	r := mux.NewRouter()

	r.HandleFunc("/", PrinterList)
	r.HandleFunc("/generate", GenerateQR)
	r.HandleFunc("/generateCompatible", generateCompatible)
	r.HandleFunc("/addprinter", AddPrinter)
	r.HandleFunc("/addcartridge", AddCartridge)
	r.HandleFunc("/cartridgeOfPrinter", AddCartridgeOfPrinter)
	r.HandleFunc("/compatible", FindCompatibleCartridges)
	r.HandleFunc("/printer/{printerName}", PrinterPage)
	fmt.Println("Server is listening...")

	http.Handle("/", r)
	http.ListenAndServe(":" + port, nil)
}

func generateCompatible(w http.ResponseWriter, r *http.Request){
	fmt.Println("New generating request from" + r.RemoteAddr)
	data, err := ioutil.ReadAll(r.Body)
	if err != nil{
		fmt.Println("Something went wrong during generating QR code")
		fmt.Fprintf(w, "Something went wrong during generating QR code")
	}
	filename := GenerateRandomString(10)
	fmt.Println("Generated name " + filename + " for " + r.RemoteAddr)
	generateFromText(string(data), filename)
	fmt.Println("Generated file " + filename + ".png for " + r.RemoteAddr)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename +".png"))
	http.ServeFile(w, r, filename)
	fmt.Println("Served file " + filename + ".png for " + r.RemoteAddr)
	os.Remove(filename + ".png")
	fmt.Println("Removed file " + filename + ".png")
}

func PrinterPage(w http.ResponseWriter, r *http.Request) {
	db, err := gorm.Open(sqlite.Open("printer.db"), &gorm.Config{})

	if err != nil {
		fmt.Println("Error during opening DB")
	} else {
		name := mux.Vars(r)["printerName"]
		var printer Printers
		db.Where("Name = ?", name).First(&printer)

		var printers []Printers
		db.Find(&printers)
		fmt.Println(printer)
		if printer.ID > 0 {
			var cartridges []string
			var cot []Cartridgeofprinter
			db.Find(&cot)
			db.Where("Printer_Id = ?", printer.ID).Find(&cot)
			for _, v := range cot {
				var cart Cartridges
				if v.Printer_Id == printer.ID{
					db.Where("ID = ?", v.Cartridge_Id).First(&cart)
					cartridges = append(cartridges, cart.Name)
				}
			}

			data := Output{
				PrinterName: name,
				Cartridges:  cartridges,
			}

			tmpl, _ := template.ParseFiles("output.html")
			tmpl.Execute(w, data)
		} else {
			http.Redirect(w, r, "https://www.youtube.com/watch?v=dQw4w9WgXcQ", 301)
		}
	}

	// w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", name + ".png"))
	// var cartText = strings.Join(cartridgeName, "\n")
	// cartridgeQR(cartText, name)
	// http.ServeFile(w, r, name)
}

func FindCompatibleCartridges(w http.ResponseWriter, r *http.Request) {
	db, _ := gorm.Open(sqlite.Open("printer.db"), &gorm.Config{})

	var PrinterList []Printers
	db.Find(&PrinterList)
	var PrinterName []string
	for _, pr := range PrinterList {
		PrinterName = append(PrinterName, pr.Name)
	}

	// var printer Printers
	r.ParseForm()
	selectedPrinter := strings.Join(r.Form["printer"], "")

	if selectedPrinter != "" {
		// db.Where("Name = ?", selectedPrinter).First(&printer)
		// var cops []uint
		// db.Table("CartridgeOfPrinters").Where("PrinterID = ?", printer.ID).Select("CartridgeID").Find(&cops)

		// var cartridgeName []string

		// for _, cop := range cops {
		// 	var cart Cartridges
		// 	db.Where("ID = ?", cop).First(&cart)
		// 	cartridgeName = append(cartridgeName, cart.Name)
		// }
		// // data := ViewData{
		// // Title: "Совместимость картриджей",
		// // PrinterNames: PrinterName,
		// // CartridgeNames: cartridgeName,
		// // }

		http.Redirect(w, r, "/printer/"+selectedPrinter, 301)
		return
	}

	data := ViewData{
		Title:        "Совместимость картриджей",
		PrinterNames: PrinterName,
	}

	tmpl, _ := template.ParseFiles("compatible.html")
	tmpl.Execute(w, data)
}

func AddCartridgeOfPrinter(w http.ResponseWriter, r *http.Request) {
	db, _ := gorm.Open(sqlite.Open("printer.db"), &gorm.Config{})
	r.ParseForm()
	cartridgeName := strings.Join(r.Form["cartridges"], "")
	printerName := strings.Join(r.Form["printers"], "")
	// printerName := strings.Join(r.Form["printers"], "")
	var cartridge Cartridges
	db.Where("Name = ?", cartridgeName).First(&cartridge)
	cartridgeID := cartridge.ID
	var printer Printers
	db.Where("Name = ?", printerName).First(&printer)
	printerID := printer.ID
	var cop Cartridgeofprinter
	cop.Cartridge_Id = cartridgeID
	cop.Printer_Id = printerID
	db.Create(&cop)
	http.Redirect(w, r, "/", 302)
}

func AddCartridge(w http.ResponseWriter, r *http.Request) {
	db, _ := gorm.Open(sqlite.Open("printer.db"), &gorm.Config{})
	r.ParseForm()
	text := strings.Join(r.Form["cartridgeName"], "")
	var newCartridge Cartridges
	newCartridge.Name = text
	db.Create(&newCartridge)
	http.Redirect(w, r, "/", 302)
}

func AddPrinter(w http.ResponseWriter, r *http.Request) {
	db, _ := gorm.Open(sqlite.Open("printer.db"), &gorm.Config{})
	r.ParseForm()
	text := strings.Join(r.Form["printerName"], "")
	var newPrinter Printers
	newPrinter.Name = text
	db.Create(&newPrinter)
	http.Redirect(w, r, "/", 302)
}

func GenerateQR(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() // Получение имени принтера
	text := strings.Join(r.Form["printer"], "")
	GenerateFromWeb(w, r, text)
}

func GenerateFromWeb(w http.ResponseWriter, r *http.Request, printerName string){
	qrtext := "https://printers-ttit.herokuapp.com/printer/" + printerName // Генерация URL
	fmt.Println("Generating QR code with text: " + qrtext + " for " + r.RemoteAddr)
	filename := GenerateRandomString(10) // Генерация имени файла
	fmt.Println("Generated filename " + filename + " for " + r.RemoteAddr)
	err := generateFromText(qrtext, filename) // qrtext - текст в QR, filename - имя файла

	if err != nil{
		fmt.Println(err)
	}

	fmt.Println("Setting header Content-Disposition for " + r.RemoteAddr)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename +".png"))
	fmt.Println("Starting serving file for " + r.RemoteAddr)
	http.ServeFile(w, r, filename + ".png")
}

func generateFromText(text string, filename string) error{
	code, err := qr.Encode(text, qr.L, qr.Auto)
	if err != nil {
		return fmt.Errorf("Error during generating QR code")
	}
	if text != code.Content() {
		return fmt.Errorf("data differs")
	}
	code, err = barcode.Scale(code, 512, 512)
	if err != nil {
		return fmt.Errorf("Error during scaling QR code")
	}

	err = writePng(filename, code)
	if err != nil{
		return err
	}
	return nil
}

func writePng(filename string, img image.Image) error{
	file, err := os.Create(filename + ".png")
	if err != nil {
		return fmt.Errorf("Creating file error")
	}
	err = png.Encode(file, img)
	if err != nil {
		return fmt.Errorf("PNG file error")
	}
	file.Close()
	return nil
}

func GenerateRandomString(n int) string {
    b := make([]rune, n)
    for i := range b {
        b[i] = letterRunes[rand.Intn(len(letterRunes))]
    }
    return string(b)
}

func PrinterList(w http.ResponseWriter, r *http.Request) {
	db, _ := gorm.Open(sqlite.Open("printer.db"), &gorm.Config{})

	var PrinterList []Printers
	db.Find(&PrinterList)
	var PrinterName []string
	for _, pr := range PrinterList {
		PrinterName = append(PrinterName, pr.Name)
	}

	var CartridgeList []Cartridges
	db.Find(&CartridgeList)
	var CartridgeName []string
	for _, ca := range CartridgeList {
		CartridgeName = append(CartridgeName, ca.Name)
	}

	data := ViewData{
		Title:          "Generate QR",
		PrinterNames:   PrinterName,
		CartridgeNames: CartridgeName,
	}

	tmpl, _ := template.ParseFiles("generate.html")
	tmpl.Execute(w, data)
}

func cartridgeQR(text string, filename string) {
	code, err := qr.Encode(text, qr.L, qr.Auto)
	if err != nil {
		fmt.Println("Something went wrong...")
	}
	if text != code.Content() {
		log.Fatal("data differs")
	}
	code, err = barcode.Scale(code, 512, 512)
	if err != nil {
		log.Fatal(err)
	}

	writePng(filename, code)
}

