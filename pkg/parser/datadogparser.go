// Package parser provides methods for parsing DataDog events from raw UDP packets.
package parser

import (
	"bytes"
	"fmt"
)

// MetricType is stored as a string.
// Can be one of:
// - MetricGauge ("G") - gauge
// - MetricCount ("C") - count
// - MetricHist ("H") - histogram
// - MetricSet ("S") - set
// - MetricTiming ("T") - timing
// - MetricServiceCheck ("_SC") - service check
// - MetricEvent ("_E") - event
type MetricType string

// String implements Stringer for a MetricType to provide more human-friendly metric type descriptions.
func (m MetricType) String() string {
	switch m {
	case MetricEvent:
		return "EVENT"
	case MetricGauge:
		return "GAUGE"
	case MetricCount:
		return "COUNT"
	case MetricHist:
		return "HISTOGRAM"
	case MetricSet:
		return "SET"
	case MetricTiming:
		return "TIMING"
	case MetricServiceCheck:
		return "SERVICE_CHECK"
	default:
		return "UNKNOWN"
	}
}

// ServiceCheckStatus is stored as a string.
// Can be one of:
// - ServiceCheckOK ("OK")
// - ServiceCheckWarn ("WARN")
// - ServiceCheckCritical ("CRITICAL")
// - ServiceCheckUnknown ("UNKNOWN")
type ServiceCheckStatus string

const (
	// MetricGauge is a Gauge metric
	MetricGauge MetricType = "G"
	// MetricCount is a Count metric. Increment and Decrement are also Count.
	MetricCount MetricType = "C"
	// MetricHist is a Histogram metric.
	MetricHist MetricType = "H"
	// MetricSet is a Set metric.
	MetricSet MetricType = "S"
	// MetricTiming is a Timing metric in ms.
	MetricTiming MetricType = "T"
	// MetricServiceCheck is a Service check. Not strictly a metric.
	MetricServiceCheck MetricType = "_SC"
	// MetricEvent is an Event. Again, not strictly a metric.
	MetricEvent MetricType = "_E"

	// ServiceCheckOK is an OK ServiceCheckStatus.
	ServiceCheckOK ServiceCheckStatus = "OK"
	// ServiceCheckWarn is a Warn ServiceCheckStatus
	ServiceCheckWarn ServiceCheckStatus = "WARN"
	// ServiceCheckCritical is a Critical ServiceCheckStatus
	ServiceCheckCritical ServiceCheckStatus = "CRITICAL"
	// ServiceCheckUnknown is an Unknown ServiceCheckStatus.
	ServiceCheckUnknown ServiceCheckStatus = "UNKNOWN"
)

// ErrEmptyPayload is returned upon encountering a payload containing only tags, e.g. `#foo,bar`.
var ErrEmptyPayload = fmt.Errorf("empty payload after stripping tags")

// ErrInvalidTrailingPipe is returned upon encountering an extra trailing pipe before metric tags.
var ErrInvalidTrailingPipe = fmt.Errorf("payload should have exactly one trailing pipe before tag start")

// ErrNoTrailingPipe is returned if there is no trailing pipe before metric tags.
var ErrNoTrailingPipe = fmt.Errorf("missing trailing pipe")

// ErrNoTypeSep is returned if a metric payload has no separator between metric type and tags.
var ErrNoTypeSep = fmt.Errorf("missing type separator")

// ErrNoValSep is returned if there is no separator between metric name and metric value.
var ErrNoValSep = fmt.Errorf("missing value separator")

// ErrInvalidMetricType is returned if an unknown metric type is encountered.
var ErrInvalidMetricType = fmt.Errorf("invalid metric type")

// ErrInvalidServiceCheckType is returned if an unknown service check type is encountered.
var ErrInvalidServiceCheckType = fmt.Errorf("invalid service check type")

// ErrNoMsgSep is returned upon parsing an event with no separator between the event name and event body.
var ErrNoMsgSep = fmt.Errorf("missing pipe between event name and body")

var prefixServiceCheck = []byte("_sc|")
var prefixEvent = []byte("_e")

var sepColon = []byte(":")
var sepComma = []byte(",")
var sepHash = []byte("#")
var sepPipe = []byte("|")

var typeGauge = []byte("g")
var typeCount = []byte("c")
var typeHistogram = []byte("h")
var typeSet = []byte("s")
var typeTiming = []byte("ms")

var typeServiceCheckOK = []byte("0")
var typeServiceCheckWarn = []byte("1")
var typeServiceCheckCritical = []byte("2")
var typeServiceCheckUnknown = []byte("3")

// DatadogMetric is a single DataDog metric.
type DatadogMetric struct {
	Name  string
	Value string
	Type  MetricType
	Tags  []string
}

// String returns a string representation of a Datadog metric.
func (d *DatadogMetric) String() string {
	return fmt.Sprintf("%s %s %s %v", d.Type, d.Name, d.Value, d.Tags)
}

// DatadogParser parses datadog metrics.
type DatadogParser interface {
	Parse(payload []byte) (*DatadogMetric, error)
}

// datadogParser implements DatadogParser
type datadogParser struct{}

var _ DatadogParser = (*datadogParser)(nil)

// NewDatadogParser returns a new instance of DatadogParser
func NewDatadogParser() DatadogParser {
	return &datadogParser{}
}

// Parse parses a raw UDP message and returns a DatadogMetric or an error if parsing unsuccessful.
func (p *datadogParser) Parse(payload []byte) (*DatadogMetric, error) {
	var m *DatadogMetric
	var err error
	metricTags, tagStart := p.parseTags(payload)

	trimmed := payload[:tagStart]

	if len(trimmed) == 0 {
		return nil, ErrEmptyPayload
	}

	// trim trailing pipe if it exists
	trimmed = bytes.TrimSuffix(trimmed, sepPipe)

	if bytes.HasPrefix(trimmed, prefixServiceCheck) {
		body := trimmed[len(prefixServiceCheck):]
		m, err = p.parseServiceCheck(body, metricTags)
	} else if bytes.HasPrefix(trimmed, prefixEvent) {
		body := trimmed[len(prefixEvent):]
		m, err = p.parseEvent(body, metricTags)
	} else {
		m, err = p.parseMetric(trimmed, metricTags)
	}

	if err != nil {
		return nil, err
	}

	return m, nil
}

// parseMetric parses a Datadog metric from trimmed, assuming tags have already been stripped.
func (p *datadogParser) parseMetric(trimmed []byte, tags []string) (*DatadogMetric, error) {
	// metric.name:value|type
	if len(trimmed) < 1 {
		return nil, ErrEmptyPayload
	}

	// if trimmed ends with a pipe then no metric type is present
	if bytes.HasSuffix(trimmed, sepPipe) {
		return nil, ErrInvalidTrailingPipe
	}

	typeStart := bytes.LastIndex(trimmed, sepPipe)
	if typeStart == -1 {
		return nil, ErrNoTypeSep
	}

	rawMetricType := trimmed[typeStart+1:]
	metricType, err := p.typeOfMetric(rawMetricType)
	if err != nil {
		return nil, err
	}

	sepIdx := bytes.LastIndex(trimmed[:typeStart-1], sepColon)
	if sepIdx == -1 {
		return nil, ErrNoValSep
	}

	metricName := string(trimmed[0:sepIdx])
	metricValue := string(trimmed[sepIdx+1 : typeStart])

	return &DatadogMetric{
		Name:  metricName,
		Value: metricValue,
		Type:  metricType,
		Tags:  tags,
	}, nil
}

// parseServiceCheck parses a service check from trimmed, assuming tags have already been stripped.
func (p *datadogParser) parseServiceCheck(trimmed []byte, tags []string) (*DatadogMetric, error) {
	// servicecheck.name|value
	if len(trimmed) < 1 {
		return nil, ErrEmptyPayload
	}

	// if trimmed ends with a pipe then no service check value is present
	if bytes.HasSuffix(trimmed, sepPipe) {
		return nil, ErrInvalidTrailingPipe
	}

	typeEnd := len(trimmed)
	typeStart := bytes.LastIndex(trimmed[:typeEnd], sepPipe)
	if typeStart == -1 {
		return nil, ErrNoTypeSep
	}

	rawMetricType := trimmed[typeStart+1 : typeEnd]
	scType, err := p.typeOfServiceCheck(rawMetricType)
	if err != nil {
		return nil, err
	}

	scName := string(trimmed[:typeStart])

	return &DatadogMetric{
		Name:  scName,
		Value: string(scType),
		Type:  MetricServiceCheck,
		Tags:  tags,
	}, nil
}

// parseEvent parses a Datadog event from trimmed, assuming tags have already been stripped.
func (p *datadogParser) parseEvent(trimmed []byte, tags []string) (*DatadogMetric, error) {
	// _e{name_length,message_length}:name|message
	if len(trimmed) == 0 {
		return nil, ErrEmptyPayload
	}

	nameStart := bytes.Index(trimmed, sepColon)
	if nameStart == -1 {
		return nil, ErrNoValSep
	}

	nameEnd := bytes.Index(trimmed, sepPipe)
	if nameEnd == -1 {
		return nil, ErrNoMsgSep
	}

	evtName := string(trimmed[nameStart+1 : nameEnd])
	evtBody := string(trimmed[nameEnd+1:])
	return &DatadogMetric{
		Name:  evtName,
		Value: evtBody,
		Type:  MetricEvent,
		Tags:  tags,
	}, nil

}

// parseTags returns the tags of payload and the starting position of tags in payload.
// Tags are assumed to begin from the last index of '#' in payload.
func (p *datadogParser) parseTags(payload []byte) ([]string, int) {
	tags := make([]string, 0)
	tagStart := bytes.LastIndex(payload, sepHash)
	if tagStart == -1 {
		return tags, len(payload)
	}

	tagBytes := bytes.Split(payload[tagStart+1:], sepComma)
	for i := 0; i < len(tagBytes); i++ {
		tags = append(tags, string(tagBytes[i]))
	}
	return tags, tagStart
}

func (p *datadogParser) typeOfMetric(b []byte) (MetricType, error) {
	if bytes.Equal(b, typeGauge) {
		return MetricGauge, nil
	}
	if bytes.Equal(b, typeCount) {
		return MetricCount, nil
	}
	if bytes.Equal(b, typeHistogram) {
		return MetricHist, nil
	}
	if bytes.Equal(b, typeSet) {
		return MetricSet, nil
	}
	if bytes.Equal(b, typeTiming) {
		return MetricTiming, nil
	}
	return "", ErrInvalidMetricType
}

func (p *datadogParser) typeOfServiceCheck(b []byte) (ServiceCheckStatus, error) {
	if bytes.Equal(b, typeServiceCheckOK) {
		return ServiceCheckOK, nil
	}

	if bytes.Equal(b, typeServiceCheckWarn) {
		return ServiceCheckWarn, nil
	}

	if bytes.Equal(b, typeServiceCheckCritical) {
		return ServiceCheckCritical, nil
	}

	if bytes.Equal(b, typeServiceCheckUnknown) {
		return ServiceCheckUnknown, nil
	}

	return "", ErrInvalidServiceCheckType
}
