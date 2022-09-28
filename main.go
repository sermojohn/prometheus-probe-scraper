package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/common/expfmt"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("missing arguments")
		os.Exit(1)
	}
	endpoint := args[1]

	resp, err := http.Get(endpoint)
	if err != nil {
		fmt.Printf("failed to call prometheus endpoint, %v\n", err)
		os.Exit(1)
	}

	defer resp.Body.Close()
	parser := expfmt.TextParser{}
	metricFamilies, err := parser.TextToMetricFamilies(resp.Body)
	if err != nil {
		fmt.Printf("failed to decode prometheus endpoint output, %v\n", err)
		os.Exit(1)
	}

	normalisedMetrics := []NormalisedMetric{}
	for _, mf := range metricFamilies {
		for _, m := range mf.Metric {
			if counter := m.GetCounter(); counter != nil {
				value := counter.GetValue()
				normalisedMetrics = append(normalisedMetrics, NormalisedMetric{
					Name:  *mf.Name,
					Value: value,
				})
			}
			if gauge := m.GetGauge(); gauge != nil {
				value := gauge.GetValue()
				normalisedMetrics = append(normalisedMetrics, NormalisedMetric{
					Name:  *mf.Name,
					Value: value,
				})
			}
			if summary := m.GetSummary(); summary != nil {
				value := summary.GetSampleSum()
				normalisedMetrics = append(normalisedMetrics, NormalisedMetric{
					Name:  *mf.Name,
					Value: value,
				})
			}
			if histogram := m.GetHistogram(); histogram != nil {
				value := histogram.GetSampleSum()
				normalisedMetrics = append(normalisedMetrics, NormalisedMetric{
					Name:  *mf.Name,
					Value: value,
				})
			}
		}
	}

	jsonOutput, err := json.MarshalIndent(normalisedMetrics, "", "  ")
	if err != nil {
		log.Printf("failed to serialize output, due to %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonOutput))
}

type NormalisedMetric struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value,omitempty"`
}
