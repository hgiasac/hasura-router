package tracing

import (
	"time"

	"github.com/google/uuid"
)

// TimeMeasurement store tracing checkpoints of logs ingestion process
type TimeMeasurement struct {
	StartTime time.Time
}

// Calculate calculate duration metrics of logs ingestion
func (tm *TimeMeasurement) Calculate() map[string]interface{} {
	timeM := map[string]interface{}{}
	now := time.Now()

	if tm.StartTime.IsZero() {
		return timeM
	}
	timeM["start"] = tm.StartTime
	timeM["end"] = now
	timeM["total_time"] = durationToMilliseconds(now.Sub(tm.StartTime))

	return timeM
}

// Tracing store tracing infomation for logging
type Tracing struct {
	requestId   string
	fields      map[string]interface{}
	measurement *TimeMeasurement
}

// New create new tracing context instance
func New(requestId string) *Tracing {

	if requestId == "" {
		requestId = uuid.New().String()
	}

	return &Tracing{
		requestId: requestId,
		fields:    make(map[string]interface{}),
		measurement: &TimeMeasurement{
			StartTime: time.Now(),
		},
	}
}

func (t Tracing) GetRequestId() string {
	return t.requestId
}

func (t *Tracing) SetRequestId(requestId string) {
	t.requestId = requestId
}

func (t *Tracing) WithField(key string, value interface{}) *Tracing {
	t.fields[key] = value
	return t
}

func (t *Tracing) WithFields(fields map[string]interface{}) *Tracing {
	for k, v := range fields {
		t.fields[k] = v
	}
	return t
}

func (t *Tracing) Values() map[string]interface{} {
	values := make(map[string]interface{})
	for k, v := range t.fields {
		values[k] = v
	}
	values["request_id"] = t.requestId
	values["measurement"] = t.measurement.Calculate()
	return values
}

// durationToMilliseconds convert duration to milliseconds
func durationToMilliseconds(d time.Duration) float64 {
	return float64(d) / float64(time.Millisecond)
}
