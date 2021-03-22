package main

import (
	"bufio"
	"flag"
	"log"
	"os"

	exapi "github.com/zaneGittins/exapi"
	"gopkg.in/ini.v1"
)

var (
	exabeamSection string = "exabeam"
	APIUsername    string = ""
	APIPassword    string = ""
	ContextTable   string = ""
	AAHost         string = ""
)

func UploadResultsToExabeam(data []string) {
	// Initialize authentication struct and api struct.
	auth := exapi.ExabeamAuth{Username: APIUsername, Password: APIPassword}
	api := exapi.ExabeamAAApi{Auth: auth, Tablename: ContextTable, Host: AAHost}
	api.Initialize()

	// Authenticate to the API.
	result := api.Authenticate()
	if result != 200 {
		log.Printf("auth status code %d\n", result)
		return
	}

	// Create struct of keys to upload to context table.
	keys := []exapi.NewKey{}
	for _, v := range data {
		newKey := exapi.NewKey{Key: string(v)}
		keys = append(keys, newKey)
	}
	newRecords := exapi.NewRecords{ContextTableName: api.Tablename, Records: keys}

	// Upload new records to the context table.
	result, resultJSON := api.AddRecords(newRecords)
	if result != 200 {
		log.Printf("add status code %d\n", result)
		return
	}

	// Commit changes to the context table.
	commit := exapi.CommitChangesData{SessionId: resultJSON.SessionId, Replace: true}
	result = api.CommitChanges(commit)
	if result != 200 {
		log.Printf("commit status code %d\n", result)
		return
	}
}

func parseConfig(configFile string) {

	// Load and parse configuration file.
	cfg, err := ini.Load(configFile)
	if err != nil {
		log.Fatalln(err)
		return
	}

	// Parse Exabeam settings.

	// Get username for API from config.
	APIUsername = cfg.Section(exabeamSection).Key("username").String()

	// Get password for API from config.
	APIPassword = cfg.Section(exabeamSection).Key("password").String()

	// Get context table name from config.
	ContextTable = cfg.Section(exabeamSection).Key("context_table").String()

	// Get AA host name from config.
	AAHost = cfg.Section(exabeamSection).Key("host").String()

}

func main() {
	config := flag.String("config", "config.ini", "path to config.ini file, optional.")
	flag.Parse()

	parseConfig(*config)

	scanner := bufio.NewScanner(os.Stdin)

	data := []string{}
	for scanner.Scan() {
		data = append(data, scanner.Text())
	}

	UploadResultsToExabeam(data)
}
