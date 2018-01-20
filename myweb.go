package main

import (
	"encoding/json"
	"fmt"
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

// ServerName is Mongo DB server host address
var ServerName string

// DicTab is the detailed schema information for each table
var DicTab map[string]map[string]map[string]Field

// MaxOutNum is the max returned records number
var MaxOutNum int

func getHeaderStr(dbName string, tableName string) string {
	record := DicTab[dbName][tableName]
	data := ""
	for k := range record {
		data += `{ "name": "` + k + `", "type": "text" },`
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

func getItemsStr(dbName string, tableName string, fields map[string][]string) string {
	session, err := mgo.Dial(ServerName)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	c := session.DB(dbName).C(tableName)

	filter := bson.M{}
	for k, v := range fields {
		if len(v) > 0 && len(v[0]) > 0 {
			if DicTab[dbName][tableName][k].Type == "int" {
				filter[k], _ = strconv.Atoi(v[0])
			} else {
				filter[k] = v[0]
			}
		}
	}

	var records []bson.M
	err = c.Find(filter).Limit(MaxOutNum).All(&records)
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
	session, err := mgo.Dial(ServerName)
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

	if _, exist := DicTab[dbName]; !exist {
		DicTab[dbName] = make(map[string]map[string]Field)
	}

	if _, exist := DicTab[dbName][tableName]; !exist {
		DicTab[dbName][tableName] = make(map[string]Field)
	}

	for k, v := range record {
		DicTab[dbName][tableName][k] = Field{Type: reflect.TypeOf(v).String()}
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

func main() {

	ServerName = "127.0.0.1"
	MaxOutNum = 100
	DicTab = make(map[string]map[string]map[string]Field)

	fmt.Println("Starting server...")
	http.HandleFunc("/", dispatch)

	log.Fatal(http.ListenAndServe(":80", nil))

}
