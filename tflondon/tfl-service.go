package tflondon

import (
	"strings"
	"github.com/thoeni/go-tfl"
	"net/http"
	"time"
)

var httpTimeout = 5 * time.Second

type Service interface {
	GetStatusFor(lines []string) (map[string]tfl.Report, error)
}

type service struct {
	Client tfl.Client
}

func NewService() *service {
	return &service{
		tfl.NewCachedClient(&http.Client{Timeout: httpTimeout}, 120),
	}
}

func (s service) GetStatusFor(lines []string) (map[string]tfl.Report, error) {
	reports, err := s.Client.GetTubeStatus()
	if err != nil {
		return nil, err
	}
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
