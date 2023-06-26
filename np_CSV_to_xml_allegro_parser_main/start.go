package main

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Shop struct {
	XMLName    xml.Name   `xml:"yml_catalog"`
	Date       string     `xml:"date,attr"`
	ShopName   string     `xml:"shop>name"`
	Company    string     `xml:"shop>company"`
	URL        string     `xml:"shop>url"`
	Currencies []Currency `xml:"shop>currencies>currency"`
	Categories []Category `xml:"shop>categories>category"`
	Offers     []Offer    `xml:"shop>offers>offer"`
}

type Currency struct {
	ID   string `xml:"id,attr"`
	Rate string `xml:"rate,attr"`
}

type Category struct {
	ID    string `xml:"id,attr"`
	Value string `xml:",chardata"`
}

type Offer struct {
	ID          string `xml:"id,attr"`
	Available   string `xml:"available,attr"` 
	URL         string `xml:"url"`
	Price       string `xml:"price"`
	CurrencyID  string `xml:"currencyId"`
	CategoryID  string `xml:"categoryId"`
	Picture     string `xml:"picture"`
	Pickup      string `xml:"pickup"`
	Delivery    string `xml:"delivery"`
	Name        string `xml:"name"`
	// NameUA      string `xml:"name_ua"`
	Vendor      string `xml:"vendor"`
	VendorCode  string `xml:"vendorCode"`
	Description string `xml:"description"`
	// DescriptionUA string `xml:"description_ua"`
	PackWeight  string `xml:"pack_weight"`
	Condition   string `xml:"condition"`
	GTIN        []Param `xml:"param"`
}

type Param struct {
	Name  string `xml:"name,attr"`
	Unit  string `xml:"unit,attr"`
	//Unit  string `xml:"unit,attr,omitempty"`
	Value string `xml:",chardata"`
}


func main() {
	// Открываем CSV-файл для чтения
	file, err := os.Open("products.csv")
	if err != nil {
		fmt.Println("Ошибка при открытии файла:", err)
		return
	}
	defer file.Close()

	// Создаем CSV Reader
	reader := csv.NewReader(file)
	reader.Comma = '\t' // Задаем разделитель как табуляцию
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	// Читаем заголовок CSV-файла
	_, err = reader.Read()
	if err != nil {
		fmt.Println("Ошибка при чтении заголовка CSV:", err)
		return
	}

	// Создаем структуру магазина
	shop := Shop{
		Date:     time.Now().Format("2006-01-02 15:04"),
		ShopName: "Allegro *UA*",
		Company:  "Allegro *UA*",
		URL:      "Allegro *UA*",
		Currencies: []Currency{
			{ID: "USD", Rate: "CB"},
			{ID: "PLN", Rate: "1"},
			{ID: "BYN", Rate: "CB"},
			{ID: "KZT", Rate: "CB"},
			{ID: "EUR", Rate: "CB"},
		},
	}

	// Создаем карту для хранения уникальных категорий
	categoryMap := make(map[string]string)

	// Создаем список для хранения товаров
	offers := []Offer{}

	// Читаем данные товаров из CSV-файла
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Ошибка при чтении данных CSV:", err)
			return
		}

		// Проверяем количество полей
		if len(record) != 22 {
			fmt.Println("Неправильное количество полей в строке CSV")
			continue
		}

		// Обрабатываем значения и пропускаем пустые значения
		for i := range record {
			if record[i] == "" {
				record[i] = "NULL"
			}
		}

		// Создаем структуру товара
			offer := Offer{
				ID:         record[0],
				Available:  "true",
				URL:        record[10],
				Price:      record[6],
				CurrencyID: "PLN",
				CategoryID: record[19],
				Picture:    record[11],
				Pickup:     "false",
				Delivery:   "true",
				Name:       "",
				Vendor:     record[2],
				VendorCode: "",
				PackWeight: record[14],
			}
			
	
		// Создаем срез GTIN
		offer.GTIN = []Param{}
		
		if record[4] != "NULL" {
			offer.GTIN = append(offer.GTIN, Param{Name: "GTIN", Unit: "", Value: record[4]})
		} else {
			// offer.GTIN = append(offer.GTIN, Param{Name: "GTIN", Unit: "", Value: ""})
		}
				
		if offer.CategoryID == "0"{
			offer.CategoryID = "1000"
		}

		// Проверяем и формируем поле Name
		name := record[3]
		if record[1] != "" && strings.Contains(name, record[1]) {
			name = strings.Replace(name, record[1], "", 1) + " " + record[2] + " " + record[1]
		}
		offer.Name = name + " " + record[2] + " " + record[1]
		
		offer.Name = offer.Name;

		// Формируем поле Description с barcode
		description := record[16]
		if record[4] != "" && record[4] != "NULL" {
			description += "\nbarcode=\"" + record[4] + "\""
		}
		offer.Description = description
		
		if record[18] == "Новый" {
			offer.Condition = "Новый"
		}

		// Добавляем товар в список
		offers = append(offers, offer)

		// Добавляем категорию в карту
		// 	categoryMap[record[19]] = record[5]
		categoryID := record[19]
		categoryName := record[5]
		
		if categoryID == "0" {
		categoryName = "Основная"
		categoryID = "1000"
		}

		categoryMap[categoryID] = categoryName
	}

	// Сортируем товары по полю CategoryID
	sort.Slice(offers, func(i, j int) bool {
		return offers[i].CategoryID < offers[j].CategoryID
	})

	// Создаем список уникальных категорий из карты
	var categories []Category
	for id, value := range categoryMap {
		categories = append(categories, Category{ID: id, Value: value})
	}

	// Сортируем категории по ID
	sort.Slice(categories, func(i, j int) bool {
		id1, _ := strconv.Atoi(categories[i].ID)
		id2, _ := strconv.Atoi(categories[j].ID)
		return id1 < id2
	})

	// Добавляем категории в магазин
	shop.Categories = categories

	// Максимальное количество товаров в одном файле
	maxProductsPerFile := 2000

	// Создаем XML-файлы с товарами
	for i := 0; i < len(offers); i += maxProductsPerFile {
		end := i + maxProductsPerFile
		if end > len(offers) {
			end = len(offers)
		}

		// Создаем подмножество товаров
		subOffers := offers[i:end]

		// Создаем новый магазин с подмножеством товаров
		subShop := shop
		subShop.Offers = subOffers

		// Конвертируем структуру магазина в XML
		xmlData, err := xml.MarshalIndent(subShop, "", "  ")
		if err != nil {
			fmt.Println("Ошибка при преобразовании в XML:", err)
			return
		}

		// Сохраняем XML в файл
		outputDir := "output" // Путь к папке "output"
		batchFilename := fmt.Sprintf("%s/products_%d.xml", outputDir, i+1)
		xmlFile, err := os.Create(batchFilename)
		if err != nil {
			fmt.Println("Ошибка при создании файла XML:", err)
			return
		}
		defer xmlFile.Close()

		_, err = xmlFile.Write([]byte(xml.Header))
		if err != nil {
			fmt.Println("Ошибка при записи в XML-файл:", err)
			return
		}

		_, err = xmlFile.Write(xmlData)
		if err != nil {
			fmt.Println("Ошибка при записи в XML-файл:", err)
			return
		}

		fmt.Println("Файл", batchFilename, "создан.")
	}

	fmt.Println("Конвертация завершена.")
}
