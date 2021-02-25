package exapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"strings"
)

// ExabeamAuth - Data to authenticate to the AA API.
type ExabeamAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Jar      *cookiejar.Jar
}

// ExabeamAAApi - Primary struct for holding auth and using API.
type ExabeamAAApi struct {
	Auth              ExabeamAuth
	Tablename         string
	Host              string
	BaseURL           string
	ContextTableBase  string
	LoginEndpoint     string
	TableTemplate     string
	RecordsEndpoint   string
	AddRecordEndpoint string
}

func (e *ExabeamAAApi) Initialize() {
	e.BaseURL = e.Host + "/api/"
	e.ContextTableBase = "setup/contextTables"
	e.LoginEndpoint = e.BaseURL + "auth/login"
	e.TableTemplate = "{table_name}"
	e.RecordsEndpoint = e.BaseURL + e.ContextTableBase + "/" + e.TableTemplate + "/records"
	e.AddRecordEndpoint = e.BaseURL + e.ContextTableBase + "/" + e.TableTemplate + "/" + "changes/add"
}

// ContextRecord - A record received when reading records from a current table.
type ContextRecord struct {
	ID         string `json:"id"`
	Key        string `json:"key"`
	Position   int    `json:"position"`
	SourceType string `json:"sourceType"`
}

// ContextTable - Context table returned when getting context table via api.
type ContextTable struct {
	AdRecordsSize     int             `json:"adRecordsSize"`
	ManualRecordsSize int             `json:"manualRecordsSize"`
	Records           []ContextRecord `json:"records"`
	TisRecordsSize    int             `json:"tisRecordsSize"`
}

// NewKey - ney key to add to a context table.
type NewKey struct {
	Key string `json:"key"`
}

// NewRecords - table name and slice of keys to submit to context table api.
type NewRecords struct {
	ContextTableName string   `json:"contextTableName"`
	Records          []NewKey `json:"records"`
}

// AddRecordChange - Data returned when adding a context table.
type AddRecordChange struct {
	RecordChanges string `json:"recordChanges"`
	ChangeType    string `json:"changeType"`
	Record        NewKey `json:"record"`
}

// AddRecordMetadata - Metadata returned when adding a context table.
type AddRecordMetadata struct {
	CreatedSize int `json:"createdSize"`
	RemovedSize int `json:"removedSize"`
	UpdatedSize int `json:"updatedSize"`
}

// AddRecordResult - Result returned when adding a context table.
type AddRecordResult struct {
	Metadata      AddRecordMetadata `json:"metadata"`
	RecordChanges []AddRecordChange `json:"recordChanges"`
	SessionId     string            `json:"sessionId"`
}

// CommitChangesData - information necessary to commit changes to table.
type CommitChangesData struct {
	SessionId string `json:"sessionId"`
	Replace   bool   `json:"replace"`
}

// Authenticate - authenticates to advanced analytics and stores auth cookie.
func (e *ExabeamAAApi) Authenticate() int {

	e.Auth.Jar, _ = cookiejar.New(nil)
	var client = &http.Client{
		Jar: e.Auth.Jar,
	}

	logonJSON, _ := json.Marshal(&e.Auth)
	req, err := http.NewRequest("POST", e.LoginEndpoint, bytes.NewBuffer(logonJSON))
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}

	// body, _ := ioutil.ReadAll(resp.Body)
	return resp.StatusCode
}

// GetRecords - gets all current records.
func (e *ExabeamAAApi) GetRecords() (int, ContextTable) {

	var client = &http.Client{
		Jar: e.Auth.Jar,
	}

	endpoint := strings.Replace(e.RecordsEndpoint, e.TableTemplate, e.Tablename, -1)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		fmt.Println(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}

	var contextTable ContextTable
	err = json.Unmarshal([]byte(body), &contextTable)
	if err != nil {
		log.Println(err)
	}

	return resp.StatusCode, contextTable
}

// AddRecords - Adds new records to a context table.
func (e *ExabeamAAApi) AddRecords(records NewRecords) (int, AddRecordResult) {
	var client = &http.Client{
		Jar: e.Auth.Jar,
	}

	recordsJSON, _ := json.Marshal(records)
	endpoint := strings.Replace(e.AddRecordEndpoint, e.TableTemplate, e.Tablename, -1)
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(recordsJSON))
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}

	body, _ := ioutil.ReadAll(resp.Body)

	var results AddRecordResult
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		log.Println(err)
	}

	return resp.StatusCode, results
}

// CommitChanges - Commits all changes made to a context table.
func (e *ExabeamAAApi) CommitChanges(commit CommitChangesData) int {
	var client = &http.Client{
		Jar: e.Auth.Jar,
	}

	commitJSON, _ := json.Marshal(commit)
	endpoint := strings.Replace(e.RecordsEndpoint, e.TableTemplate, e.Tablename, -1)
	req, err := http.NewRequest("PUT", endpoint, bytes.NewBuffer(commitJSON))
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}

	return resp.StatusCode
}
