package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"
	"github.com/xuri/excelize/v2"
	"html"

)

type shop struct {
	Name       string      `xml:"name"`
	Company    string      `xml:"company"`
	URL        string      `xml:"url"`
	Currencies []Currency  `xml:"currencies>currency"`
	Categories []Category  `xml:"categories>category"`
	Offers     []Offer     `xml:"offers>offer"`
}

type Currency struct {
	ID   string `xml:"id,attr"`
	Rate string `xml:"rate,attr"`
}

type Category struct {
	ID       string `xml:"id,attr"`
	ParentID string `xml:"parentId,attr,omitempty"`
	Name     string `xml:",chardata"`
}

type Offer struct {
	ID          string    `xml:"id,attr"`
	Available   bool      `xml:"available,attr"`
	URL         string    `xml:"url"`
	Price       string    `xml:"price"`
	CurrencyID  string    `xml:"currencyId"`
	CategoryID  string    `xml:"categoryId"`
	Pictures    []Picture `xml:"picture"`
	Pickup      bool      `xml:"pickup"`
	Delivery    bool      `xml:"delivery"`
	Name        string    `xml:"name"`
	Vendor      string    `xml:"vendor"`
	VendorCode  string    `xml:"vendorCode"`
	Description CDATAText `xml:"description"`
	GTIN        []Param `xml:"param"`
}

type Param struct {
	Name  string `xml:"name,attr"`
	Unit  string `xml:"unit,attr"`
	//Unit  string `xml:"unit,attr,omitempty"`
	Value string `xml:",chardata"`
}

type Picture struct {
	Value string `xml:",chardata"`
}

// CDATAText represents a text value wrapped within a CDATA section.
type CDATAText struct {
	Value string `xml:",cdata"`
}

func main() {
	// Открытие файла Excel
	f, err := excelize.OpenFile("products.xlsx")
	if err != nil {
		log.Fatal(err)
	}

	// Чтение данных из первой страницы Excel
	rows, err := f.GetRows("Export Products Sheet")
	if err != nil {
		log.Fatal(err)
	}

	// Создание структуры магазина
	shop := shop{
		Name:    "Allegro *UA*",
		Company: "Allegro *UA*",
		URL:     "Allegro *UA*",
		Currencies: []Currency{
			{ID: "USD", Rate: "CB"},
			{ID: "PLN", Rate: "1"},
			{ID: "BYN", Rate: "CB"},
			{ID: "KZT", Rate: "CB"},
			{ID: "EUR", Rate: "CB"},
		},
		Categories: []Category{},
		Offers:     []Offer{},
	}

	// Проход по строкам Excel
	for _, row := range rows[1:] {
		// Создание оффера
		offer := Offer{
			ID:         row[0],
			Available:  true,
			URL:        "",
			Price:      row[8],
			CurrencyID: "PLN",
			CategoryID: row[27],
			Pictures:   []Picture{},
			Pickup:     false,
			Delivery:   true,
			Name:       row[2],
			Vendor:     "",
			VendorCode: row[24],
			Description: CDATAText{
				Value: html.UnescapeString(row[6]),
			},
		}

		//offer.Name = html.EscapeString(offer.Name)
		//offer.NameUkr = html.EscapeString(offer.NameUkr)

		// Добавление изображений
		pictures := row[14]
		if pictures != "" {
			imageUrls := strings.Split(pictures, ",")
			for _, imageUrl := range imageUrls {
				picture := Picture{Value: imageUrl}
				offer.Pictures = append(offer.Pictures, picture)
			}
		}
		
		// Создаем срез GTIN
		offer.GTIN = []Param{}
		
		if len(row) > 53 && row[53] != "NULL" {
			offer.GTIN = append(offer.GTIN, Param{Name: "GTIN", Unit: "", Value: row[53]})
		} 

		// Добавление оффера в список офферов
		shop.Offers = append(shop.Offers, offer)
	}

	// Чтение данных из второй страницы Excel
	rows, err = f.GetRows("Export Groups Sheet")
	if err != nil {
		log.Fatal(err)
	}

	// Проход по строкам Excel
	for _, row := range rows[1:] {
		// Создание категории
		category := Category{
			ID:   row[3],
			Name: row[2],
		}

		// Добавление родительской категории, если есть
		if len(row) > 5 && row[5] != "" {
			category.ParentID = row[5]
		}

		// Добавление категории в список категорий
		shop.Categories = append(shop.Categories, category)
	}

	// Генерация XML
	output, err := xml.MarshalIndent(shop, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	// Запись XML-файла
	header := `<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE yml_catalog>
`

	date := time.Now().Format("2006-01-02 15:04")

	footer := "\n</yml_catalog>"

	xmlData := []byte(header + fmt.Sprintf("<yml_catalog date=\"%s\">\n", date) + string(output) + footer)

	err = ioutil.WriteFile("output.xml", xmlData, 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Done!")
}
