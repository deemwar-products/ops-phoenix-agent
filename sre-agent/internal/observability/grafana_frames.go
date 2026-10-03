package observability

import (
	"encoding/json"
	"fmt"
)

// Grafana's /api/ds/query returns DataFrames, not raw Loki rows. A logs frame
// looks like:
//
//	schema.fields = [{labels}, {Time}, {Line}, {tsNs}, {id}]
//	data.values   = [[{container=...}], [ts...], ["log line"...], [ts...], [id...]]
//
// One array per field, aligned by row index. So the log line is values[2] and
// its timestamp values[1] — not the last array, and not a [ts, line] pair.
// The old parser expected each value to be a 2-element [ts, line] pair, which
// no Loki frame produces, so every frame was skipped and detect reported
// "0 errors" against a Loki that plainly had logs.

type lokiFrame struct {
	Schema struct {
		Fields []struct {
			Name string `json:"name"`
		} `json:"fields"`
	} `json:"schema"`
	Data struct {
		Values [][]json.RawMessage `json:"values"`
	} `json:"data"`
}

// lokiLabels matches Grafana's labels column. It is a list, one entry per row,
// and the shape inside varies: Loki frames put a map per row
// ({"container":"api","host":"app1"}), while some paths put a single map for
// the whole frame. Unmarshal either way.
type lokiLabels struct {
	// Per-row maps.
	Rows []map[string]string
	// Single map for the whole frame.
	One map[string]string
}

func (l *lokiLabels) UnmarshalJSON(b []byte) error {
	var rows []map[string]string
	if err := json.Unmarshal(b, &rows); err == nil {
		l.Rows = rows
		return nil
	}
	var one map[string]string
	if err := json.Unmarshal(b, &one); err == nil {
		l.One = one
		return nil
	}
	return nil
}

func (l *lokiLabels) forRow(i int) map[string]string {
	if l == nil {
		return nil
	}
	if i < len(l.Rows) {
		return l.Rows[i]
	}
	return l.One
}

type dsQueryEnvelope struct {
	Results map[string]struct {
		Error  string      `json:"error"`
		Frames []lokiFrame `json:"frames"`
	} `json:"results"`
}

// parseGrafanaResponse unwraps Grafana's DataFrame envelope into ErrorLogs.
func parseGrafanaResponse(data []byte) ([]ErrorLog, error) {
	var env dsQueryEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("unmarshal grafana response: %w", err)
	}

	var out []ErrorLog
	for ref, r := range env.Results {
		if r.Error != "" {
			return nil, fmt.Errorf("grafana query %s: %s", ref, r.Error)
		}
		for _, frame := range r.Frames {
			labels, tsCol, lineCol := frame.columns()
			if lineCol < 0 || lineCol >= len(frame.Data.Values) {
				continue
			}
			lines := frame.Data.Values[lineCol]
			for i, raw := range lines {
				var line string
				if err := json.Unmarshal(raw, &line); err != nil || line == "" {
					continue
				}
				e := ErrorLog{Message: line, Raw: line, Labels: map[string]string{}}
				if tsCol >= 0 && tsCol < len(frame.Data.Values) && i < len(frame.Data.Values[tsCol]) {
					var ts string
					if json.Unmarshal(frame.Data.Values[tsCol][i], &ts) == nil {
						e.Timestamp = parseGrafanaTimestamp(ts)
					}
				}
				if labels >= 0 && labels < len(frame.Data.Values) {
					col := frame.Data.Values[labels]
					if i < len(col) {
						var lbls lokiLabels
						if json.Unmarshal(col[i], &lbls) == nil {
							for k, v := range lbls.forRow(i) {
								e.Labels[k] = v
							}
							e.Container = e.Labels["container"]
						}
					}
				}
				out = append(out, e)
			}
		}
	}
	return out, nil
}

// columns locates the labels, Time and Line arrays by schema field name, so a
// Grafana version that adds or reorders fields doesn't break parsing.
func (f lokiFrame) columns() (labels, ts, line int) {
	labels, ts, line = -1, -1, -1
	for i, fld := range f.Schema.Fields {
		switch fld.Name {
		case "labels":
			labels = i
		case "Time", "tsNs":
			if ts < 0 {
				ts = i
			}
		case "Line":
			line = i
		}
	}
	return
}

func rowCount(f lokiFrame) int {
	n := 0
	for _, col := range f.Data.Values {
		if len(col) > n {
			n = len(col)
		}
	}
	return n
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}