package dbhelper

import (
	"encoding/json"
	"log"
	"reflect"
	"strconv"

	"../common"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// LoadTab will load the fields information to the map dicFields
func LoadTab(dbName string,
	tableName string,
	conf util.WebConf,
	dicFields util.DBFields) {

	session, err := mgo.Dial(conf.DBServer)
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

	if _, exist := dicFields[dbName]; !exist {
		dicFields[dbName] = make(map[string]map[string]util.Field)
	}

	if _, exist := dicFields[dbName][tableName]; !exist {
		dicFields[dbName][tableName] = make(map[string]util.Field)
	}

	for k, v := range record {
		dicFields[dbName][tableName][k] = util.Field{Type: reflect.TypeOf(v).String()}
	}
}

// GetItemsStr will search the mongo db using the passed fields, the dicFields is to help identify integer value
func GetItemsStr(dbName string,
	tableName string,
	fields map[string][]string,
	dicHides map[string]bool,
	conf util.WebConf,
	dicFields util.DBFields) string {
	session, err := mgo.Dial(conf.DBServer)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	c := session.DB(dbName).C(tableName)

	columns := bson.M{}
	filter := bson.M{}
	for k, v := range fields {
		if len(v) > 0 && len(v[0]) > 0 {
			if dicFields[dbName][tableName][k].Type == "int" {
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
	err = c.Find(filter).Select(columns).Limit(conf.MaxOutNum).All(&records)
	if err != nil {
		log.Fatal(err)
	}
	if len(records) == 0 {
		return "[]"
	}
	data, _ := json.Marshal(records)
	return string(data)
}
