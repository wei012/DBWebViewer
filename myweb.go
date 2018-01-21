package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"./common"
	db "./dblayer"
)

// DicFields is the detailed schema information for each table
var DicFields util.DBFields

// DicWebConfs is the detailed configuration of each field
var DicWebConfs util.WebConf

func getSchema(dbName string, tableName string, w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	data := util.GetHeaderStr(dbName, tableName, DicFields[dbName][tableName], DicWebConfs)
	w.Header().Set("Content-Length", fmt.Sprint(len(data)))
	fmt.Fprint(w, string(data))
}

func getItems(dbName string, tableName string, w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)

	dicHides, _ := util.GetDisplayColumns(dbName, tableName, DicWebConfs)
	data := db.GetItemsStr(dbName, tableName, r.URL.Query(), dicHides, DicWebConfs, DicFields)
	w.Header().Set("Content-Length", fmt.Sprint(len(data)))
	fmt.Fprint(w, data)
}

func callAPI(paras []string, w http.ResponseWriter, r *http.Request) {
	num := len(paras)
	if num < 2 {
		return
	}

	dbName := paras[0]
	tableName := paras[1]
	db.LoadTab(dbName, tableName, DicWebConfs, DicFields)
	if num == 2 {
		http.ServeFile(w, r, "./static/grid.html")
	} else if num == 3 && paras[2] == "items" {
		getItems(dbName, tableName, w, r)
	} else if num == 3 && paras[2] == "schema" {
		getSchema(dbName, tableName, w, r)
	}
}

func dispatch(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	num := len(parts)

	if num == 2 && (parts[1] == "" || parts[1] == "index.html") {
		http.ServeFile(w, r, "./static/index.html")
	} else if num > 1 && parts[1] == "db" {
		callAPI(parts[2:], w, r)
	} else {
		http.ServeFile(w, r, "./static/"+r.URL.Path[1:])
	}
}

func loadConf() bool {
	content, err := ioutil.ReadFile("./conf/web.conf.json")
	if err != nil {
		fmt.Println(err)
		return false
	}

	if err := json.Unmarshal(content, &DicWebConfs); err != nil {
		fmt.Println(err)
		return false
	}

	DicFields = make(map[string]map[string]map[string]util.Field)
	return true
}

func main() {
	fmt.Println("Loading configuration...")
	if !loadConf() {
		fmt.Println("Failed to load configuration information!")
		return
	}
	fmt.Println("Starting server...")
	http.HandleFunc("/", dispatch)
	log.Fatal(http.ListenAndServe(":"+DicWebConfs.WebPort, nil))
}
