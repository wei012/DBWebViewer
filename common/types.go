package util

// WebConf is the configuration for the web server
type WebConf struct {
	DBServer  string
	MaxOutNum int
	WebPort   string
	DB        map[string]map[string]map[string]string
}

//Field is the struct of data
type Field struct {
	Type  string
	Alias string
	Hide  bool
}

//DBFields contains the fields' name and type information in a DB
type DBFields = map[string]map[string]map[string]Field
