package main

import (
	tfl "github.com/thoeni/go-tfl"
	"strings"
	"time"
)

type TflService interface {
	GetStatusFor(lines []string) (map[string]tfl.Report, error)
}

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
	if time.Since(c.LastUpdated).Seconds() > c.InvalidateIntervalInSeconds {
		start := time.Now()
		r, e := c.Client.GetTubeStatus()
		c.TubeStatus = r
		c.LastUpdated = time.Now()
		elapsed := time.Since(start)
		msElapsed := elapsed / time.Millisecond
		tflResponseLatencies.WithLabelValues("getTubeStatus").Observe(float64(msElapsed))
		return c.TubeStatus, e
	}
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
	reportsMap := tfl.ReportArrayToMap(reports)
	if len(lines) == 0 {
		return reportsMap, nil
	} else {
		filteredReportsMap := filter(reportsMap, lines)
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
