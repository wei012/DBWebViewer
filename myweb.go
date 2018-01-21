package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//Field is the struct of data
type Field struct {
	Type  string
	Alias string
	Hide  bool
}

// WebConf is the configuration for the web server
type WebConf struct {
	DBServer  string
	MaxOutNum int
	WebPort   string
	DB        map[string]map[string]map[string]string
}

// DicFields is the detailed schema information for each table
var DicFields map[string]map[string]map[string]Field

// DicWebConfs is the detailed configuration of each field
var DicWebConfs WebConf

func getHeaderStr(dbName string, tableName string) string {
	record := DicFields[dbName][tableName]
	data := ""
	dicHides := getDisplayColumns(dbName, tableName)
	for k := range record {
		if _, has := dicHides[k]; !has {
			data += `{ "name": "` + k + `", "type": "text" },`
		}
	}
	if len(data) > 0 {
		data = data[0 : len(data)-1]
	}
	return "[" + data + "]"
}

func getSchema(dbName string, tableName string, w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	data := getHeaderStr(dbName, tableName)
	w.Header().Set("Content-Length", fmt.Sprint(len(data)))
	fmt.Fprint(w, string(data))
}

//Load the information of fields to hide
func getDisplayColumns(dbName string, tableName string) map[string]bool {
	dicHide := make(map[string]bool)
	if db, hasDB := DicWebConfs.DB[dbName]; hasDB {
		if tab, hasTab := db[tableName]; hasTab {
			parts := strings.Split(tab["Hides"], ",")
			for _, s := range parts {
				if len(s) > 0 {
					dicHide[s] = true
				}
			}
		}
	}
	return dicHide
}

func getItemsStr(dbName string, tableName string, fields map[string][]string) string {
	session, err := mgo.Dial(DicWebConfs.DBServer)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	c := session.DB(dbName).C(tableName)

	columns := bson.M{}
	filter := bson.M{}

	dicHides := getDisplayColumns(dbName, tableName)
	for k, v := range fields {
		if len(v) > 0 && len(v[0]) > 0 {
			if DicFields[dbName][tableName][k].Type == "int" {
				filter[k], _ = strconv.Atoi(v[0])
			} else {
				filter[k] = v[0]
			}
		}

		if _, has := dicHides[k]; !has {
			columns[k] = 1
		}
	}

	var records []bson.M
	err = c.Find(filter).Select(columns).Limit(DicWebConfs.MaxOutNum).All(&records)
	if err != nil {
		log.Fatal(err)
	}
	data, _ := json.Marshal(records)
	return string(data)
}

func getItems(dbName string, tableName string, w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	data := getItemsStr(dbName, tableName, r.URL.Query())
	w.Header().Set("Content-Length", fmt.Sprint(len(data)))
	fmt.Fprint(w, data)
}

func loadTab(dbName string, tableName string) {
	session, err := mgo.Dial(DicWebConfs.DBServer)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	c := session.DB(dbName).C(tableName)

	var record bson.M
	err = c.Find(nil).One(&record)
	if err != nil {
		log.Fatal(err)
	}

	if _, exist := DicFields[dbName]; !exist {
		DicFields[dbName] = make(map[string]map[string]Field)
	}

	if _, exist := DicFields[dbName][tableName]; !exist {
		DicFields[dbName][tableName] = make(map[string]Field)
	}

	for k, v := range record {
		DicFields[dbName][tableName][k] = Field{Type: reflect.TypeOf(v).String()}
	}
}

func callAPI(paras []string, w http.ResponseWriter, r *http.Request) {
	num := len(paras)
	if num < 2 {
		return
	}

	dbName := paras[0]
	tableName := paras[1]
	loadTab(dbName, tableName)
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

	DicFields = make(map[string]map[string]map[string]Field)
	return true
}

func main() {

	fmt.Println("Loading configuration...")
	if !loadConf() {
		fmt.Println("Failed to load configuration file")
		return
	}
	fmt.Println("Starting server...")
	http.HandleFunc("/", dispatch)
	log.Fatal(http.ListenAndServe(":"+DicWebConfs.WebPort, nil))

}
