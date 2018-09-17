package parser

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type DatadogMetricSuite struct {
	suite.Suite
}

func (s *DatadogMetricSuite) Test_DatadogMetric() {
	m := &DatadogMetric{
		Name:  "foo",
		Value: "bar",
		Type:  MetricCount,
		Tags:  []string{"baz"},
	}
	s.EqualValues("foo", m.Name)
	s.EqualValues("bar", m.Value)
	s.EqualValues(MetricCount, m.Type)
	s.EqualValues([]string{"baz"}, m.Tags)
	s.EqualValues("COUNT foo bar [baz]", m.String())
}

type DatadogParserSuite struct {
	suite.Suite
	p datadogParser
}

func (s *DatadogParserSuite) SetupSuite() {
	s.p = datadogParser{}
}

func (s *DatadogParserSuite) Test_Parse_Empty() {
	input := []byte("")
	m, err := s.p.Parse(input)
	s.Nil(m)
	s.EqualValues(ErrEmptyPayload, err)
}

func (s *DatadogParserSuite) Test_ParseMulti() {
	input := []byte("foo:1|c|#baz,zap\n\nnotavalidmetric\nbar:2|c")
	ms, errs := s.p.ParseMulti(input)

	s.Require().Len(errs, 3)
	s.Require().Len(ms, 3)

	s.Require().Nil(errs[0])
	s.Require().Nil(errs[2])
	s.Require().NotNil(errs[1])

	s.Require().NotNil(ms[0])
	s.Require().NotNil(ms[2])
	s.Require().Nil(ms[1])

	s.Equal(MetricCount, ms[0].Type)
	s.Equal(MetricCount, ms[2].Type)

	s.Equal("foo", ms[0].Name)
	s.Equal("bar", ms[2].Name)

	s.Equal("1", ms[0].Value)
	s.Equal("2", ms[2].Value)
}

func (s *DatadogParserSuite) Test_GoStatsd_Valid_Metric() {
	input := []byte("modprox-registry.heartbeat-accepted:1|c")
	m, err := s.p.Parse(input)
	s.Require().NoError(err)
	s.Require().NotNil(m)
	s.Equal(MetricCount, m.Type)
	s.Equal("modprox-registry.heartbeat-accepted", m.Name)
	s.Equal("1", m.Value)
	s.Empty(m.Tags)
}

func (s *DatadogParserSuite) Test_Parse_Metric_ValidWithTags() {
	input := []byte("foo:bar|c|#baz,zap")
	m, err := s.p.Parse(input)
	s.EqualValues(&DatadogMetric{
		Name:  "foo",
		Value: "bar",
		Type:  MetricCount,
		Tags:  []string{"baz", "zap"},
	}, m)
	s.NoError(err)
}

func (s *DatadogParserSuite) Test_Parse_ServiceCheck_ValidWithTags() {
	input := []byte("_sc|foobar|0|#baz,zap")
	m, err := s.p.Parse(input)
	s.EqualValues(&DatadogMetric{
		Name:  "foobar",
		Value: string(ServiceCheckOK),
		Type:  MetricServiceCheck,
		Tags:  []string{"baz", "zap"},
	}, m)
	s.NoError(err)
}

func (s *DatadogParserSuite) Test_Parse_Event_ValidWithTags() {
	input := []byte("_e{3,6}:foo|barbaz|#baz,zap")
	m, err := s.p.Parse(input)
	s.EqualValues(&DatadogMetric{
		Name:  "foo",
		Value: "barbaz",
		Type:  MetricEvent,
		Tags:  []string{"baz", "zap"},
	}, m)
	s.NoError(err)
}

func (s *DatadogParserSuite) Test_parseMetric_Empty() {
	payload := []byte("")
	tags := []string(nil)
	m, err := s.p.parseMetric(payload, tags)
	s.Nil(m)
	s.EqualValues(ErrEmptyPayload, err)
}

func (s *DatadogParserSuite) Test_parseMetric_NoValue() {
	payload := []byte("foobar|c")
	tags := []string(nil)
	m, err := s.p.parseMetric(payload, tags)
	s.Nil(m)
	s.EqualValues(ErrNoValSep, err)
}

func (s *DatadogParserSuite) Test_parseMetric_TrailingPipe() {
	payload := []byte("foo:bar|")
	tags := []string(nil)
	m, err := s.p.parseMetric(payload, tags)
	s.Nil(m)
	s.EqualValues(ErrInvalidTrailingPipe, err)
}

func (s *DatadogParserSuite) Test_parseMetric_NoTypeSep() {
	payload := []byte("foo:barc")
	tags := []string(nil)
	m, err := s.p.parseMetric(payload, tags)
	s.Nil(m)
	s.EqualValues(ErrNoTypeSep, err)
}

func (s *DatadogParserSuite) Test_parseMetric_ValidNoTags() {
	payload := []byte("foo:bar|c")
	tags := []string(nil)
	m, err := s.p.parseMetric(payload, tags)
	s.EqualValues(&DatadogMetric{
		Name:  "foo",
		Value: "bar",
		Type:  MetricCount,
		Tags:  []string(nil),
	}, m)
	s.NoError(err)
}

func (s *DatadogParserSuite) Test_parseMetric_InvalidTrailingTags() {
	payload := []byte("foo:bar|c|#baz")
	tags := []string(nil)
	m, err := s.p.parseMetric(payload, tags)
	s.Nil(m)
	s.EqualValues(ErrInvalidMetricType, err)
}

func (s *DatadogParserSuite) Test_parseServiceCheck_Empty() {
	payload := []byte("")
	tags := []string(nil)
	m, err := s.p.parseServiceCheck(payload, tags)
	s.Nil(m)
	s.EqualValues(ErrEmptyPayload, err)
}

func (s *DatadogParserSuite) Test_parseServiceCheck_TrailingPipe() {
	payload := []byte("foo.bar|")
	tags := []string(nil)
	m, err := s.p.parseServiceCheck(payload, tags)
	s.Nil(m)
	s.EqualValues(ErrInvalidTrailingPipe, err)
}

func (s *DatadogParserSuite) Test_parseServiceCheck_NoType() {
	payload := []byte("foo.bar")
	tags := []string(nil)
	m, err := s.p.parseServiceCheck(payload, tags)
	s.Nil(m)
	s.EqualValues(ErrNoTypeSep, err)
}

func (s *DatadogParserSuite) Test_parseServiceCheck_InvalidType() {
	payload := []byte("foo.bar|baz")
	tags := []string(nil)
	m, err := s.p.parseServiceCheck(payload, tags)
	s.Nil(m)
	s.EqualValues(ErrInvalidServiceCheckType, err)
}

func (s *DatadogParserSuite) Test_parseEvent_Empty() {
	input := []byte("")
	e, err := s.p.parseEvent(input, []string(nil))
	s.Nil(e)
	s.EqualValues(ErrEmptyPayload, err)
}

func (s *DatadogParserSuite) Test_parseEvent_MissingValSep() {
	input := []byte("{3,6}foo|barbaz")
	e, err := s.p.parseEvent(input, []string(nil))
	s.Nil(e)
	s.EqualValues(ErrNoValSep, err)
}

func (s *DatadogParserSuite) Test_parseEvent_MissingMsgSep() {
	input := []byte("{3,6}:foobarbaz")
	e, err := s.p.parseEvent(input, []string(nil))
	s.Nil(e)
	s.EqualValues(ErrNoMsgSep, err)
}

func (s *DatadogParserSuite) Test_parseEvent_ValidNoTags() {
	input := []byte("{3,6}:foo|barbaz")
	e, err := s.p.parseEvent(input, []string(nil))
	s.EqualValues(&DatadogMetric{
		Name:  "foo",
		Value: "barbaz",
		Type:  MetricEvent,
		Tags:  []string(nil),
	}, e)
	s.NoError(err)
}

func (s *DatadogParserSuite) Test_parseEvent_ValidEmpty() {
	input := []byte("{0,0}:|")
	e, err := s.p.parseEvent(input, []string(nil))
	s.EqualValues(&DatadogMetric{
		Name:  "",
		Value: "",
		Type:  MetricEvent,
		Tags:  []string(nil),
	}, e)
	s.NoError(err)
}

func (s *DatadogParserSuite) Test_parseTags_Empty() {
	input := []byte("")
	tags, idx := s.p.parseTags(input)
	s.Empty(tags)
	s.EqualValues(0, idx)
}

func (s *DatadogParserSuite) Test_parseTags_ValidOneTag() {
	input := []byte("k:v|c|#foo:1")
	tags, idx := s.p.parseTags(input)
	s.EqualValues([]string{"foo:1"}, tags)
	s.EqualValues(6, idx)
}

func (s *DatadogParserSuite) Test_parseTags_ValidTwoTags() {
	input := []byte("k:v|c|#foo:1,bar:2")
	tags, idx := s.p.parseTags(input)
	s.EqualValues([]string{"foo:1", "bar:2"}, tags)
	s.EqualValues(6, idx)
}

func (s *DatadogParserSuite) Test_typeOfMetric() {
	m, err := s.p.typeOfMetric(typeGauge)
	s.EqualValues(MetricGauge, m)
	s.NoError(err)

	m, err = s.p.typeOfMetric(typeSet)
	s.EqualValues(MetricSet, m)
	s.NoError(err)

	m, err = s.p.typeOfMetric(typeCount)
	s.EqualValues(MetricCount, m)
	s.NoError(err)

	m, err = s.p.typeOfMetric(typeHistogram)
	s.EqualValues(MetricHist, m)
	s.NoError(err)

	m, err = s.p.typeOfMetric(typeSet)
	s.EqualValues(MetricSet, m)
	s.NoError(err)

	m, err = s.p.typeOfMetric(typeTiming)
	s.EqualValues(MetricTiming, m)
	s.NoError(err)

	m, err = s.p.typeOfMetric([]byte(""))
	s.Empty(m)
	s.EqualValues(err, ErrInvalidMetricType)

	m, err = s.p.typeOfMetric([]byte("foo"))
	s.Empty(m)
	s.EqualValues(err, ErrInvalidMetricType)

}

func (s *DatadogParserSuite) Test_typeOfServiceCheck() {
	t, err := s.p.typeOfServiceCheck(typeServiceCheckOK)
	s.EqualValues(ServiceCheckOK, t)
	s.NoError(err)

	t, err = s.p.typeOfServiceCheck(typeServiceCheckWarn)
	s.EqualValues(ServiceCheckWarn, t)
	s.NoError(err)

	t, err = s.p.typeOfServiceCheck(typeServiceCheckCritical)
	s.EqualValues(ServiceCheckCritical, t)
	s.NoError(err)

	t, err = s.p.typeOfServiceCheck(typeServiceCheckUnknown)
	s.EqualValues(ServiceCheckUnknown, t)
	s.NoError(err)

	t, err = s.p.typeOfServiceCheck([]byte(""))
	s.Zero(t)
	s.EqualValues(ErrInvalidServiceCheckType, err)

	t, err = s.p.typeOfServiceCheck([]byte("foo"))
	s.Zero(t)
	s.EqualValues(ErrInvalidServiceCheckType, err)
}

func TestDatadogParserSuite(t *testing.T) {
	suite.Run(t, new(DatadogMetricSuite))
	suite.Run(t, new(DatadogParserSuite))
}
