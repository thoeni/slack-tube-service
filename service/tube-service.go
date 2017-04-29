package service

import (
	"github.com/thoeni/go-tfl"
	"time"
)

type TubeService struct {
	client tfl.Client
}

type InMemoryCachedClient struct {
	client                      tfl.Client
	tubeStatus                  []tfl.Report
	lastRetrieve                time.Time
	invalidateIntervalInSeconds float64
}

func (c InMemoryCachedClient) GetTubeStatus() ([]tfl.Report, error) {
	if time.Since(c.lastRetrieve).Seconds() > c.invalidateIntervalInSeconds {
		return c.client.GetTubeStatus()
	}
	return c.tubeStatus, nil
}

func (s TubeService) GetStatusFor(lines []string) (map[string]tfl.Report, error) {
	reports, err := s.client.GetTubeStatus()
	if err != nil {
		return nil, err
	}
	reportsMap := tfl.ReportArrayToMap(reports)
	if len(lines) == 0 {
		return reportsMap, nil
	} else {
		return filter(reportsMap, lines), nil
	}
}

func filter(reportsMap map[string]tfl.Report, lines []string) map[string]tfl.Report {
	var response map[string]tfl.Report = make(map[string]tfl.Report)
	for _, line := range lines {
		if report, found := reportsMap[line]; found {
			response[report.Name] = report
		}
	}
	return response
}
