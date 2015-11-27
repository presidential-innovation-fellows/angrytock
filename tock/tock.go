package tockPackage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/geramirez/tock-bot/helpers"
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
	tockURL string
}

// InitTock initalizes the tock struct
func InitTock() *Tock {
	tockURL := os.Getenv("TOCK_URL")
	if tockURL == "" {
		log.Fatal("AUDIT_ENDPOINT environment variable not found")
	}
	tockURL += "/api/reporting_period_audit/"
	return &Tock{tockURL}
}

// fetchCurrentReportingPeriod collects the current reporting period
func (tock *Tock) fetchCurrentReportingPeriod() string {
	var data ReportingPeriodAuditList
	URL := fmt.Sprintf(tock.tockURL)
	body := helpers.FetchData(URL)
	err := json.Unmarshal(body, &data)
	if err != nil {
		log.Print(err)
	}
	return data.ReportingPeriods[0].StartDate
}

// FetchTockUsers is a function for collecting all the users who have not
// filled out thier time sheet for the current period
func (tock *Tock) FetchTockUsers() *ReportingPeriodAuditDetails {
	var data ReportingPeriodAuditDetails
	timePeriod := tock.fetchCurrentReportingPeriod()
	URL := fmt.Sprintf("%s%s/", tock.tockURL, timePeriod)
	body := helpers.FetchData(URL)
	err := json.Unmarshal(body, &data)
	if err != nil {
		log.Print(err)
	}
	return &data
}
