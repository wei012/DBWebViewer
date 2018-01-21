package util

import (
	"sort"
	"strings"
)

//GetDisplayColumns load the information of fields to hide
func GetDisplayColumns(dbName string,
	tableName string,
	conf WebConf) (map[string]bool, []string) {

	dicHide := make(map[string]bool)
	var orders []string
	if db, hasDB := conf.DB[dbName]; hasDB {
		if tab, hasTab := db[tableName]; hasTab {
			parts := strings.Split(tab["Hides"], ",")
			for _, s := range parts {
				if len(s) > 0 {
					dicHide[s] = true
				}
			}
			orders = strings.Split(tab["Orders"], ",")
		}
	}
	return dicHide, orders
}

//GetHeaderStr return the header columns of the grid
func GetHeaderStr(dbName string,
	tableName string,
	record map[string]Field,
	conf WebConf) string {

	data := ""
	dicHides, orders := GetDisplayColumns(dbName, tableName, conf)

	dicOrders := make(map[string]bool)
	for _, v := range orders {
		dicOrders[v] = true
	}

	dicExist := make(map[string]bool)
	var strs []string
	for k := range record {
		if _, has := dicOrders[k]; !has {
			strs = append(strs, k)
		} else {
			dicExist[k] = true
		}
	}

	var headers []string
	for _, v := range orders {
		if _, has := dicExist[v]; has {
			headers = append(headers, v)
		}
	}

	sort.Strings(strs)
	headers = append(headers, strs...)
	for _, k := range headers {
		if _, has := dicHides[k]; !has {
			data += `{ "name": "` + k + `", "type": "text" },`
		}
	}
	if len(data) > 0 {
		data = data[0 : len(data)-1]
	}
	return "[" + data + "]"
}
