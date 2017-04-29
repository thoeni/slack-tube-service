package service

import "github.com/thoeni/go-tfl"

type TflService interface {
	getStatusFor([]string) (map[string]tfl.Report, error)
}

type HttpTubeService struct {
	client tfl.Client
}

func (s HttpTubeService) getStatusFor(lines []string) (map[string]tfl.Report, error) {
	reports, _ := s.client.GetTubeStatus()
	return filter(tfl.ReportArrayToMap(reports), lines), nil
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
