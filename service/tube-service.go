package service

import (
	"github.com/thoeni/go-tfl"
	"log"
	"strings"
	"time"
)

type TubeService struct {
	Client tfl.Client
}

type InMemoryCachedClient struct {
	Client                      tfl.Client
	TubeStatus                  []tfl.Report
	LastUpdated                 time.Time
	InvalidateIntervalInSeconds float64
}

func (c *InMemoryCachedClient) GetTubeStatus() ([]tfl.Report, error) {
	log.Printf("Called GetTubeStatus on InMemoryCachedClient!")
	if time.Since(c.LastUpdated).Seconds() > c.InvalidateIntervalInSeconds {
		log.Printf("Calling TFL...")
		r, e := c.Client.GetTubeStatus()
		c.TubeStatus = r
		c.LastUpdated = time.Now()
		return c.TubeStatus, e
	}
	log.Printf("Returning cached value...")
	return c.TubeStatus, nil
}

func (c *InMemoryCachedClient) SetBaseURL(newURL string) {
	c.Client.SetBaseURL(newURL)
}

func (s TubeService) GetStatusFor(lines []string) (map[string]tfl.Report, error) {
	reports, err := s.Client.GetTubeStatus()
	if err != nil {
		return nil, err
	}
	log.Printf("Retrieved %d reports!", len(reports))
	reportsMap := tfl.ReportArrayToMap(reports)
	log.Printf("Mapped %d out of %d reports retrieved.", len(reportsMap), len(reports))
	log.Printf("Lines requested %d", len(lines))
	if len(lines) == 0 {
		log.Printf("Returning %d reports from the service.", len(reportsMap))
		return reportsMap, nil
	} else {
		filteredReportsMap := filter(reportsMap, lines)
		log.Printf("Returning %d filtered reports from the service.", len(filteredReportsMap))
		return filteredReportsMap, nil
	}
}

func filter(reportsMap map[string]tfl.Report, lines []string) map[string]tfl.Report {
	var response map[string]tfl.Report = make(map[string]tfl.Report)
	for _, line := range lines {
		if report, found := reportsMap[strings.ToLower(line)]; found {
			response[report.Name] = report
		}
	}
	return response
}
