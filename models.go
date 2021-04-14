package main

type ViewData struct {
	Title          string
	PrinterNames   []string
	CartridgeNames []string
}

type Output struct {
	PrinterName string
	Cartridges  []string
}

type Printers struct {
	ID   uint   `gorm:"AUTO_INCREMENT;type:integer"`
	Name string `gorm:"type:text"`
}

type Cartridges struct {
	ID   uint   `gorm:"AUTO_INCREMENT;type:integer"`
	Name string `gorm:"type:text"`
}

type Cartridgeofprinter struct {
	Cartridge_Id uint `gorm:"type:integer"`
	Printer_Id   uint `gorm:"type:integer"`
}
