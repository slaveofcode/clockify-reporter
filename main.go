package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/slaveofcode/clockify-report-to-slack/http_client"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".") // working directory

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic(fmt.Errorf("please create the config file [config.yaml]"))
		} else {
			panic(fmt.Errorf("unable to read config file [config.yaml]: %w", err))
		}
	}

	viper.SetDefault("clockify.urlBaseEndpoint", "https://api.clockify.me/api/v1")
	viper.SetDefault("clockify.urlReportEndpoint", "https://reports.api.clockify.me/v1")
	viper.SetDefault("clockify.urlTimeoffEndpoint", "https://pto.api.clockify.me/v1")
}

type Workspace struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	HourlyRate struct {
		Amount   int    `json:"amount"`
		Currency string `json:"currency"`
	} `json:"hourlyRate"`
	Memberships             []interface{}          `json:"memberships"`
	WorkspaceSettings       map[string]interface{} `json:"workspaceSettings"`
	ImageURL                string                 `json:"imageUrl"`
	FeatureSubscriptionType string                 `json:"featureSubscriptionType"`
}

type User struct {
	ID               string                 `json:"id"`
	Email            string                 `json:"email"`
	Name             string                 `json:"name"`
	Memberships      []interface{}          `json:"memberships"`
	ProfilePicture   string                 `json:"profilePicture"`
	ActiveWorkspace  string                 `json:"activeWorkspace"`
	DefaultWorkspace string                 `json:"defaultWorkspace"`
	Settings         map[string]interface{} `json:"settings"`
	Status           string                 `json:"status"`
	CustomFields     []interface{}          `json:"customFields"`
}

type TimeEntry struct {
	ID           string   `json:"id"`
	Description  string   `json:"description"`
	TagIds       []string `json:"tagIds"`
	UserId       string   `json:"userId"`
	Billable     bool     `json:"billable"`
	TaskId       string   `json:"taskId"`
	ProjectId    string   `json:"projectId"`
	TimeInterval struct {
		Start    string `json:"start"`
		End      string `json:"end"`
		Duration string `json:"duration"`
	} `json:"timeInterval"`
	WorkspaceId       string        `json:"workspaceId"`
	IsLocked          bool          `json:"isLocked"`
	CustomFieldValues []interface{} `json:"customFieldValues"`
	Type              string        `json:"type"`
	KioskId           string        `json:"kioskId"`
}

func getAllWorkspaces(client *http_client.HttpClient) (*[]Workspace, error) {
	res, err := client.Get(viper.GetString("clockify.urlBaseEndpoint")+"/workspaces", nil)
	if err != nil {
		return nil, err
	}

	if res.Status != 200 {
		return nil, fmt.Errorf("failed get workspaces, got non 200 status")
	}

	var wrkspc []Workspace
	err = json.Unmarshal(res.Body, &wrkspc)
	if err != nil {
		return nil, err
	}

	return &wrkspc, nil
}

func getCurrentUser(client *http_client.HttpClient) (*User, error) {
	res, err := client.Get(viper.GetString("clockify.urlBaseEndpoint")+"/user", nil)
	if err != nil {
		return nil, err
	}

	if res.Status != 200 {
		return nil, fmt.Errorf("failed get user, got non 200 status")
	}

	var user User
	err = json.Unmarshal(res.Body, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func getTimeEntries(client *http_client.HttpClient, workspaceId, userId, start, end string) (*[]TimeEntry, error) {
	res, err := client.Get(viper.GetString("clockify.urlBaseEndpoint")+"/workspaces/"+workspaceId+"/user/"+userId+"/time-entries?start="+start+"&end="+end, nil)
	if err != nil {
		return nil, err
	}

	if res.Status != 200 {
		return nil, fmt.Errorf("failed get time-entries, got non 200 status")
	}

	var timeEntries []TimeEntry
	err = json.Unmarshal(res.Body, &timeEntries)
	if err != nil {
		return nil, err
	}

	return &timeEntries, nil
}

func getPreferredWorkspaceId(wrkspcs []Workspace, workspaceName string) string {
	for _, w := range wrkspcs {
		if w.Name == workspaceName {
			return w.ID
		}
	}

	return ""
}

var day *int

func init() {
	day = flag.Int("day", 0, "Day backward value")
	flag.Parse()
}

func main() {
	client := http_client.NewHTTPClient(60, nil)

	wrkspcs, err := getAllWorkspaces(client)
	if err != nil {
		log.Println(err)
		return
	}

	user, err := getCurrentUser(client)
	if err != nil {
		log.Println(err)
		return
	}

	workspaceId := getPreferredWorkspaceId(*wrkspcs, viper.GetString("workspace.name"))
	userId := user.ID

	timeToGet := time.Now().AddDate(0, 0, *day).UTC() // default today

	startDay := time.Date(timeToGet.Year(), timeToGet.Month(), timeToGet.Day(), 0, 0, 0, timeToGet.Nanosecond(), timeToGet.Location()).Format(time.RFC3339)
	endDay := time.Date(timeToGet.Year(), timeToGet.Month(), timeToGet.Day(), 23, 59, 59, timeToGet.Nanosecond(), timeToGet.Location()).Format(time.RFC3339)

	timeEntries, err := getTimeEntries(client, workspaceId, userId, startDay, endDay)
	if err != nil {
		log.Println(err)
		return
	}

	taskList := []string{}
	for _, entry := range *timeEntries {
		hasDuplicate := false
		for _, dup := range taskList {
			if dup == entry.Description {
				hasDuplicate = true
				break
			}
		}

		if hasDuplicate {
			continue
		}

		taskList = append(taskList, entry.Description)
	}

	fmt.Printf("============= %s =============\n", timeToGet.Format("January 02, 2006"))
	sort.SliceStable(taskList, func(i, j int) bool {
		return i > j
	})
	for _, task := range taskList {
		fmt.Println("-", task)
	}
}
