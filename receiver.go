package jsoncheckreceiver

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pipeline"
	"go.uber.org/zap"
)

const (
	ATTR_SERVER_URL = "net.server.url"
)

type jsonScraper struct {
	logger    *zap.Logger
	serverURL string
	target    string
}

func newScraper(logger *zap.Logger, serverURL string, target string) (*jsonScraper, error) {
	return &jsonScraper{
		logger:    logger,
		serverURL: serverURL,
		target:    target,
	}, nil
}

func (s *jsonScraper) ID() pipeline.Signal {
	return pipeline.SignalMetrics
}

func (s *jsonScraper) Scrape(ctx context.Context) (pmetric.Metrics, error) {
	metrics := pmetric.NewMetrics()
	scopeMetrics := metrics.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics()

	dataMetric := scopeMetrics.AppendEmpty()
	dataMetric.SetName("rtu.data")
	dataMetricDataPoints := dataMetric.SetEmptyGauge().DataPoints()

	alarmMetric := scopeMetrics.AppendEmpty()
	alarmMetric.SetName("alarm.value")
	alarmMetricDataPoints := alarmMetric.SetEmptyGauge().DataPoints()

	resp, err := http.Get(fmt.Sprintf("%s", s.serverURL))
	if err != nil {
		return pmetric.NewMetrics(), fmt.Errorf("failed to fetch data from server %q: %w", s.serverURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return pmetric.NewMetrics(), fmt.Errorf("received non-200 response code from server %q: %d", s.serverURL, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return pmetric.NewMetrics(), fmt.Errorf("failed to read response body: %w", err)
	}

	var jsonData []map[string]interface{}
	if err := json.Unmarshal(body, &jsonData); err != nil {
		return pmetric.NewMetrics(), fmt.Errorf("failed to unmarshal JSON data: %w", err)
	}
	//保留字段以后使用item["ip"].(string)
	for _, item := range jsonData {
		rtuFields := []string{"rtu1", "rtu2", "rtu3", "rtu4"}
		for _, field := range rtuFields {
			if value, ok := item[field]; ok {
				appendDataPoint(dataMetricDataPoints, field, value)
			}
		}

		if alarmValue, ok := item["alarm_value"]; ok {
			appendDataPoint(alarmMetricDataPoints, "alarm.value", alarmValue)
		}
	}

	return metrics, nil
}

func appendDataPoint(metricDataPoints pmetric.NumberDataPointSlice, key string, value interface{}) {
	dp := metricDataPoints.AppendEmpty()
	dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	dp.Attributes().PutStr(ATTR_SERVER_URL, key)

	switch v := value.(type) {
	case float64:
		dp.SetDoubleValue(v)
	case int:
		dp.SetIntValue(int64(v))
	case string:
		dp.Attributes().PutStr("value", v)
	default:
		dp.Attributes().PutStr("value", fmt.Sprintf("%v", v))
	}
}
