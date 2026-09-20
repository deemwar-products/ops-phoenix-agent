package setup

import (
	"encoding/json"
	"fmt"
	"io"
)

// Emitter is the interface for step event output.
type Emitter interface {
	StepStarted(id string)
	StepSkipped(id string)
	StepDone(id string, output map[string]string)
	StepAwaitingHuman(id, url, message string)
	StepAwaitingInput(id string, fields []string, message string)
}

// textEmitter prints human-readable progress to stdout.
type textEmitter struct {
	w io.Writer
}

func NewTextEmitter(w io.Writer) Emitter {
	return &textEmitter{w: w}
}

func (e *textEmitter) StepStarted(id string) {
	fmt.Fprintf(e.w, "  → %s...\n", id)
}

func (e *textEmitter) StepSkipped(id string) {
	fmt.Fprintf(e.w, "  ✓ %s (already done)\n", id)
}

func (e *textEmitter) StepDone(id string, output map[string]string) {
	fmt.Fprintf(e.w, "  ✓ %s done\n", id)
}

func (e *textEmitter) StepAwaitingHuman(id, url, message string) {
	fmt.Fprintf(e.w, "\n  ⚠  %s needs your action:\n", id)
	fmt.Fprintf(e.w, "     %s\n", message)
	if url != "" {
		fmt.Fprintf(e.w, "     URL: %s\n", url)
	}
}

func (e *textEmitter) StepAwaitingInput(id string, fields []string, message string) {
	fmt.Fprintf(e.w, "\n  ⚠  %s needs input:\n", id)
	fmt.Fprintf(e.w, "     %s\n", message)
	fmt.Fprintf(e.w, "     Fields: %v\n", fields)
}

// jsonEmitter emits NDJSON events for LLM-driven setup.
type jsonEmitter struct {
	w io.Writer
}

func NewJSONEmitter(w io.Writer) Emitter {
	return &jsonEmitter{w: w}
}

type jsonEvent struct {
	Event string                 `json:"event"`
	Step  string                 `json:"step,omitempty"`
	Data  map[string]interface{} `json:"data,omitempty"`
}

func (e *jsonEmitter) emit(ev jsonEvent) {
	data, _ := json.Marshal(ev)
	fmt.Fprintln(e.w, string(data))
}

func (e *jsonEmitter) StepStarted(id string) {
	e.emit(jsonEvent{Event: "step_started", Step: id, Data: map[string]interface{}{"step": id}})
}

func (e *jsonEmitter) StepSkipped(id string) {
	e.emit(jsonEvent{Event: "step_skipped", Step: id, Data: map[string]interface{}{"step": id}})
}

func (e *jsonEmitter) StepDone(id string, output map[string]string) {
	e.emit(jsonEvent{Event: "step_done", Step: id, Data: map[string]interface{}{"step": id, "output": output}})
}

func (e *jsonEmitter) StepAwaitingHuman(id, url, message string) {
	e.emit(jsonEvent{
		Event: "awaiting_human",
		Step:  id,
		Data: map[string]interface{}{
			"step":      id,
			"url":       url,
			"message":   message,
			"resume_with": fmt.Sprintf("sre-agent setup run --resume --label=%s", ""),
		},
	})
}

func (e *jsonEmitter) StepAwaitingInput(id string, fields []string, message string) {
	e.emit(jsonEvent{
		Event: "awaiting_input",
		Step:  id,
		Data: map[string]interface{}{
			"step":   id,
			"fields": fields,
			"message": message,
		},
	})
}
