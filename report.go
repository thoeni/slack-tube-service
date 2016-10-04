package main

type report struct {
	Name         string
	LineStatuses []status
}

func mapTflLineToResponse(tflLine report) report {
	return tflLine
}
