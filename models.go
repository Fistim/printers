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
	Id   uint   `gorm:"type:integer"`
	Name string `gorm:"type:text"`
}

type Cartridges struct {
	Id   uint   `gorm:"type:integer"`
	Name string `gorm:"type:text"`
}

type Cartridgeofprinter struct {
	Cartridgeid uint `gorm:"type:integer"`
	Printerid   uint `gorm:"type:integer"`
}
