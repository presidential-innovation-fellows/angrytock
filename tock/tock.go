package tockPackage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/18F/angrytock/helpers"
)

// User is a struct representation of the user JSON object from tock
type User struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

// ReportingPeriod is a struct representation of the reporting_period JSON object from tock
type ReportingPeriod struct {
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
	WorkingHours int    `json:"working_hours"`
}

// APIPages is a struct representation of a API page response from tock
type APIPages struct {
	Count   int    `json:"count"`
	NextURL string `json:"next"`
	PrevURL string `json:"previous"`
}

// ReportingPeriodAuditList is a struct representation of an API response from
//the Reporting Period Audit list endpoint
type ReportingPeriodAuditList struct {
	APIPages
	ReportingPeriods []ReportingPeriod `json:"results"`
}

// ReportingPeriodAuditDetails is a struct representation of an API response
//from the Reporting Period Audit details endpoint
type ReportingPeriodAuditDetails struct {
	APIPages
	Users []User `json:"results"`
}

// Tock struct contains the audit endpoint and methods associated with Tock
type Tock struct {
	// Get Audit endpoint
	TockURL       string
	AuditEndpoint string
	DataFetcher   *helpers.DataFetcher
}

// InitTock initalizes the tock struct
func InitTock() *Tock {
	// Get the tock url
	tockURL := os.Getenv("TOCK_URL")
	if tockURL == "" {
		log.Fatal("AUDIT_ENDPOINT environment variable not found")
	}
	auditEndpoint := "/api/reporting_period_audit/"
	// Initalize a new data fetcher
	dataFetcher := helpers.NewDataFetcher(helpers.FetchData)
	return &Tock{tockURL, auditEndpoint, dataFetcher}
}

// fetchCurrentReportingPeriod gets the latest reporting time period that
// has happend
func fetchCurrentReportingPeriod(data *ReportingPeriodAuditList) string {
	currentPeriodIndex := 0
	for idx, period := range data.ReportingPeriods {
		endDate, _ := time.Parse("2006-01-02", period.EndDate)
		if endDate.Before(time.Now()) {
			currentPeriodIndex = idx
			break
		}
	}
	return data.ReportingPeriods[currentPeriodIndex].StartDate
}

// fetchCurrentReportingPeriod collects the current reporting period
func (tock *Tock) fetchReportingPeriod() string {
	var data ReportingPeriodAuditList
	URL := fmt.Sprintf(tock.AuditEndpoint)
	body := tock.DataFetcher.FetchData(URL)
	err := json.Unmarshal(body, &data)
	if err != nil {
		log.Print(err)
	}
	return fetchCurrentReportingPeriod(&data)
}

// FetchTockUsers is a function for collecting all the users who have not
// filled out thier time sheet for the current period
func (tock *Tock) FetchTockUsers() *ReportingPeriodAuditDetails {
	var data ReportingPeriodAuditDetails
	timePeriod := tock.fetchReportingPeriod()
	URL := fmt.Sprintf("%s%s", tock.AuditEndpoint, timePeriod)
	body := tock.DataFetcher.FetchData(URL)
	err := json.Unmarshal(body, &data)
	if err != nil {
		log.Print(err)
	}
	return &data
}
