package main
// Golang - Сила!
import(
	"fmt"
	"os"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"image"
	"image/png"
	"log"
	"net/http"
	"strings"
	"gorm.io/gorm"
  	"gorm.io/driver/sqlite"
  	"html/template"
)

type ViewData struct{
	Title string
	PrinterNames []string
	CartridgeNames []string
}

type Printers struct{
	ID uint `gorm:"type:integer"`
	Name string `gorm:"type:text"`
}

type Cartridges struct{
	ID uint `gorm:"type:integer"`
	Name string `gorm:"type:text"`
}

type Cartridgeofprinter struct{
	Cartridgeid uint `gorm:"type:integer"`
	Printerid uint `gorm:"type:integer"`
}

func main(){
	http.HandleFunc("/", PrinterList)
	http.HandleFunc("/generate", GenerateQR)
	http.HandleFunc("/addprinter", AddPrinter)
	http.HandleFunc("/addcartridge", AddCartridge)
	http.HandleFunc("/cartridgeOfPrinter", AddCartridgeOfPrinter)
	http.HandleFunc("/compatible", FindCompatibleCartridges)
	fmt.Println("Server is listening...")

	http.ListenAndServe(":8888", nil)
}

func FindCompatibleCartridges(w http.ResponseWriter, r *http.Request){
	db, _ := gorm.Open(sqlite.Open("printer.db"), &gorm.Config{})

	var PrinterList []Printers
	db.Find(&PrinterList)
	var PrinterName []string
	for _, pr := range PrinterList{
		PrinterName = append(PrinterName, pr.Name)
	}

	var printer Printers
	r.ParseForm()
	selectedPrinter := strings.Join(r.Form["printer"], "")
	fmt.Println("Selected: " + selectedPrinter)

	if selectedPrinter != ""{
		db.Where("Name = ?", selectedPrinter).First(&printer)
		var cops []uint
		db.Table("CartridgeOfPrinters").Where("PrinterID = ?", printer.ID).Select("CartridgeID").Find(&cops)

		var cartridgeName []string

		for _, cop := range cops{
			var cart Cartridges
			db.Where("ID = ?", cop).First(&cart)
			cartridgeName = append(cartridgeName, cart.Name)
		}
		data := ViewData{
		Title: "Совместимость картриджей",
		PrinterNames: PrinterName,
		CartridgeNames: cartridgeName,
		}

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", selectedPrinter + ".png"))
		var cartText = strings.Join(cartridgeName, "\n")
		cartridgeQR(cartText, selectedPrinter)
		http.ServeFile(w, r, selectedPrinter)

		tmpl, _ := template.ParseFiles("compatible.html")
	    tmpl.Execute(w, data)
	    return
	}

	data := ViewData{
		Title: "Совместимость картриджей",
		PrinterNames: PrinterName,
		}
	
	tmpl, _ := template.ParseFiles("compatible.html")
	tmpl.Execute(w, data)

	
}

func AddCartridgeOfPrinter(w http.ResponseWriter, r *http.Request){
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
	cop.Cartridgeid = cartridgeID
	cop.Printerid = printerID
	db.Create(&cop)
	http.Redirect(w, r, "/", 302)
}

func AddCartridge(w http.ResponseWriter, r *http.Request){
	db, _ := gorm.Open(sqlite.Open("printer.db"), &gorm.Config{})
	r.ParseForm()
	text := strings.Join(r.Form["cartridgeName"], "")
	var newCartridge Cartridges
	newCartridge.Name = text
	db.Create(&newCartridge)
	http.Redirect(w, r, "/", 302)
}

func AddPrinter(w http.ResponseWriter, r *http.Request){
	db, _ := gorm.Open(sqlite.Open("printer.db"), &gorm.Config{})
	r.ParseForm()
	text := strings.Join(r.Form["printerName"], "")
	var newPrinter Printers
	newPrinter.Name = text
	db.Create(&newPrinter)
	http.Redirect(w, r, "/", 302)
}

func GenerateQR(w http.ResponseWriter, r *http.Request){
	r.ParseForm()
	text := strings.Join(r.Form["printer"], "")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", text + ".png"))
	generateFromText(text)
	http.ServeFile(w, r, text)
}

func PrinterList(w http.ResponseWriter, r *http.Request){
	db, _ := gorm.Open(sqlite.Open("printer.db"), &gorm.Config{})

	var PrinterList []Printers
	db.Find(&PrinterList)
	var PrinterName []string
	for _, pr := range PrinterList{
		PrinterName = append(PrinterName, pr.Name)
	}

	var CartridgeList []Cartridges
	db.Find(&CartridgeList)
	var CartridgeName []string
	for _, ca := range CartridgeList{
		CartridgeName = append(CartridgeName, ca.Name)
	}


	data := ViewData{
		Title: "Generate QR",
		PrinterNames: PrinterName,
		CartridgeNames: CartridgeName,
	}

	tmpl, _ := template.ParseFiles("generate.html")
    tmpl.Execute(w, data)
}

func generateCompatibleCartridges(printer Printers){
	db, _ := gorm.Open(sqlite.Open("printer.db"), &gorm.Config{})
	var cops []Cartridgeofprinter
	var cartridges []Cartridges
	db.Where("Printerid = ?", printer.ID).Find(&cops)
	for _, cop := range cops{
		var cartridge Cartridges
		fmt.Println(cop.Cartridgeid)
		db.Where("ID = ?", cop.Cartridgeid).First(&cartridge)
		cartridges = append(cartridges, cartridge)
	}
}

func cartridgeQR(text string, filename string){
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

func generateFromText(text string){
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

	writePng(text, code)
}

func writePng(filename string, img image.Image) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	err = png.Encode(file, img)
	if err != nil {
		log.Fatal(err)
	}
	file.Close()
}