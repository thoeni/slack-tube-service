package main

import (
	"strings"
	"time"

	"github.com/thoeni/go-tfl"
)

type TflService interface {
	GetStatusFor(lines []string) (map[string]tfl.Report, error)
}

type TubeService struct {
	Client tfl.Client
}

func (s TubeService) GetStatusFor(lines []string) (map[string]tfl.Report, error) {
	start := time.Now()
	reports, err := s.Client.GetTubeStatus()
	if err != nil {
		return nil, err
	}
	go func() {
		elapsed := time.Since(start)
		msElapsed := elapsed / time.Millisecond
		tflResponseLatencies.Set(float64(msElapsed))
	}()
	reportsMap := tfl.ReportArrayToMap(reports)
	if len(lines) == 0 {
		return reportsMap, nil
	}
	filteredReportsMap := filter(reportsMap, lines)
	return filteredReportsMap, nil
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
