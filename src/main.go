package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"time"
)

var mapPrefixCountryVN = map[string]bool{
	"vietnam":  true,
	"việt nam": true,
	"việtnam":  true,
	"vn":       true,
}

var mapPrefixCityHCM = map[string]bool{
	"hồ chí minh":           true,
	"hồchíminh":             true,
	"hcm":                   true,
	"tphcm":                 true,
	"thanhphohcm":           true,
	"tphochiminh":           true,
	"thành phố hồ chí minh": true,
	"thànhphốhồchíminh":     true,
	"tp.hồ chí minh":        true,
	"tp. hồ chí minh":       true,
	"tp.hồchíminh":          true,
	"tp. hồchíminh":         true,
	"tp. hcm":               true,
	"tp.hcm":                true,
	"tp hcm":                true,
}

var mapPrefixDistrictVN = map[string]bool{
	"quận": true, // len: 6
	"quân": true, // len: 5
	"quan": true, // len: 4
	"q.":   true, //len: 2
	"q":    true, // len: 1
}

var mapPrefixWardVN = map[string]bool{
	"phường": true, // len: 9
	"phương": true, // len: 8
	"phuong": true, // len: 6
	"p.":     true, // len: 2
	"p":      true, // len: 1
}

var listSeparated = []string{
	",", // Sample Address: XXX, YYY, ZZZ
	//"-", // Sample Address: XXX - YYY - ZZZ
	//"/", // Sample Address: XXX / YYY / ZZZ
}

type AddressDetail struct {
	FullAddress string
	Street      string
	Ward        string
	District    string
	City        string
	Country     string
}

func main() {
	// Read data
	csvPath := "test.csv"
	data, err := GetDataCsv(csvPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(data) == 0 {
		fmt.Print("DONE ALLL.!!!!!!!!!!!!\n")
		return
	}
	fullAddress := []string{}
	for _, row := range data {
		fullAddress = append(fullAddress, row[0])

	}
	addressDetails := []AddressDetail{}
	for _, value := range fullAddress {
		addressDetail := GetAddressDetailFromFullAddress(value)
		addressDetails = append(addressDetails, addressDetail)
	}
	//Write data
	ExportFileCSV(addressDetails)
}

func GetDataCsv(csvPath string) ([][]string, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return [][]string{}, err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return [][]string{}, err
	}
	return rows, nil
}

func GetAddressDetailFromFullAddress(fullAddress string) (addressDetail AddressDetail) {
	if len(fullAddress) == 0 {
		return
	}
	var arrFullAddress = []string{}
	for _, value := range listSeparated {
		arrFullAddress = splitFullAddress(fullAddress, value)
		if len(arrFullAddress) > 1 {
			break
		}
	}
	// Format others: 45 Đinh Tiên Hoàng phường bến nghé quận 1
	if len(arrFullAddress) <= 1 {
		addressDetail = getAddressDetailFormatOthers(fullAddress)
		addressDetail.FullAddress = fullAddress
		return
	}

	addressDetail = getAddressDetail(arrFullAddress)
	addressDetail.FullAddress = fullAddress
	return
}

// Sample format: 45 Đinh Tiên Hoàng phường bến nghé quận 1
func getAddressDetailFormatOthers(fullAddress string) (addressDetail AddressDetail) {
	// Get postition country
	country, _ := getPositionConstainString(fullAddress, mapPrefixCountryVN)
	if country != "" {
		addressDetail.Country = country
	}
	city, positionCity := getPositionConstainString(fullAddress, mapPrefixCityHCM)
	if city != "" {
		addressDetail.City = city
	}
	_, positionDistrict := getPositionConstainString(fullAddress, mapPrefixDistrictVN)
	_, positionWard := getPositionConstainString(fullAddress, mapPrefixWardVN)
	if positionDistrict != -1 {
		if positionCity != -1 && positionCity > positionDistrict {
			addressDetail.District = fullAddress[positionDistrict : positionCity-1]
		} else if positionCity == -1 {
			addressDetail.District = fullAddress[positionDistrict:]
		}
		addressDetail.Street = fullAddress[:positionDistrict-1]
	}
	if positionWard != -1 {
		if positionDistrict != -1 && positionDistrict > positionWard {
			addressDetail.Ward = fullAddress[positionWard : positionDistrict-1]
		} else if positionDistrict == -1 {
			addressDetail.Ward = fullAddress[positionWard:]
		}
		addressDetail.Street = fullAddress[:positionWard-1]
	}
	if addressDetail.Street == "" {
		addressDetail.Street = fullAddress
	}
	return
}

func getAddressDetail(arrAddress []string) (newAddressDetail AddressDetail) {
	position := len(arrAddress)
	for {
		if position == 0 {
			break
		}
		address := arrAddress[position-1]
		if isPrefixCountryVN(strings.ToLower(address)) {
			newAddressDetail.Country = address
		} else if isPrefixCityHCM(strings.ToLower(address)) {
			newAddressDetail.City = address
		} else if isPrefixDistrictVN(strings.ToLower(address)) {
			newAddressDetail.District = address
		} else if isPrefixWardVN(strings.ToLower(address)) {
			newAddressDetail.Ward = address
		} else {
			if newAddressDetail.Street != "" {
				newAddressDetail.Street = address + " - " + newAddressDetail.Street
			} else {
				newAddressDetail.Street = address
			}
		}
		position--
	}
	return
}

func ExportFileCSV(addressDetails []AddressDetail) {
	// Prepare file
	exportDirectoryFilePath := "."
	filePath := fmt.Sprintf("%s/test-%s-%d.csv", exportDirectoryFilePath, time.Now().Format("2006-01-02"), time.Now().Unix())
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("ERROR! Can not create file, err ", err)
		return
	}
	defer file.Close()

	// Write on defer
	writer := csv.NewWriter(file)
	defer writer.Flush()
	// Get headers
	rows := [][]string{}
	headers := []string{
		"Address Origin",
		"Street",
		"Ward",
		"District",
		"City",
		"Country",
	}
	rows = append(rows, headers)

	// Collect data bulks limit bulk size
	limit := 100
	low := 0
	for low < len(addressDetails) {
		high := low + limit
		if high > len(addressDetails) {
			high = len(addressDetails)
		}
		bulks := addressDetails[low:high]
		var bulkRows [][]string
		bulkRows, err = collectCSVData(bulks, headers)
		if err != nil {
			fmt.Println("Error can not collect csv data, err ", err)
			return
		}
		rows = append(rows, bulkRows...)

		// next
		low = high
	}
	// Write
	for _, value := range rows {
		err = writer.Write(value)
		if err != nil {
			fmt.Println("Error can not write file, err ", err)
			return
		}
	}
	return
}

func collectCSVData(addressDetails []AddressDetail, headers []string) (rows [][]string, err error) {
	rows = [][]string{}
	for _, addressDetail := range addressDetails {
		row := make([]string, len(headers))
		row[0] = addressDetail.FullAddress
		row[1] = addressDetail.Street
		row[2] = addressDetail.Ward
		row[3] = addressDetail.District
		row[4] = addressDetail.City
		row[5] = addressDetail.Country

		rows = append(rows, row)
	}
	return
}

func isPrefixCountryVN(value string) bool {
	return mapPrefixCountryVN[value]
}

func isPrefixCityHCM(value string) bool {
	return mapPrefixCityHCM[value]
}

func isPrefixDistrictVN(value string) bool {
	if (len(value) > 5 && mapPrefixDistrictVN[value[0:6]]) ||
		(len(value) > 4 && mapPrefixDistrictVN[value[0:5]]) ||
		(len(value) > 3 && mapPrefixDistrictVN[value[0:4]]) ||
		(len(value) > 1 && mapPrefixDistrictVN[value[0:2]]) ||
		(len(value) > 0 && mapPrefixDistrictVN[value[0:1]]) {
		return true
	}
	return false
}

func isPrefixWardVN(value string) bool {
	if (len(value) > 8 && mapPrefixWardVN[value[0:9]]) ||
		(len(value) > 7 && mapPrefixWardVN[value[0:8]]) ||
		(len(value) > 5 && mapPrefixWardVN[value[0:6]]) ||
		(len(value) > 1 && mapPrefixWardVN[value[0:2]]) ||
		(len(value) > 0 && mapPrefixWardVN[value[0:1]]) {
		return true
	}
	return false
}

func splitFullAddress(fullAddress, separated string) (result []string) {
	splitAddress := strings.Split(fullAddress, separated)
	for _, value := range splitAddress {
		result = append(result, strings.TrimSpace(value))
	}
	return result
}

func getPositionConstainString(fullAddress string, mapListCompare map[string]bool) (value string, position int) {
	for key, _ := range mapListCompare {
		if len(key) > 1 && strings.Contains(fullAddress, key) {
			return key, strings.Index(fullAddress, key)
		}
	}
	return "", -1
}
