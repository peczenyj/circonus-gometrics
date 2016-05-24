package circonusgometrics

import (
	"net/http"
	"time"
)

// TrackHTTPLatency wraps Handler functions registered with an http.ServerMux tracking latencies.
// Metrics are of the for go`HTTP`<method>`<name>`latency and are tracked in a histogram in units
// of seconds (as a float64) providing nanosecond ganularity.
func TrackHTTPLatency(name string, handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		start := time.Now().UnixNano()
		handler(rw, req)
		elapsed := time.Now().UnixNano() - start
		hist := NewHistogram("go`HTTP`" + req.Method + "`" + name + "`latency")
		hist.RecordValue(float64(elapsed) / float64(time.Second))
	}
}
