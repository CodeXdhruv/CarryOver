package events

import (
	"encoding/json"
	"io"
	"os"
	"sync"
)

// EventType defines the type of event emitted on stdout
type EventType string

const (
	TypePhase         EventType = "phase"
	TypeProgress      EventType = "progress"
	TypeWarning       EventType = "warning"
	TypeNeedPrivilege EventType = "need_privilege"
	TypeDone          EventType = "done"
	TypeError         EventType = "error"
)

// Event is the generic structure for newline-delimited JSON events
type Event struct {
	Type           EventType `json:"type"`
	Phase          string    `json:"phase,omitempty"`
	Name           string    `json:"name,omitempty"`
	Status         string    `json:"status,omitempty"`
	Percent        int       `json:"percent,omitempty"`
	BytesDone      int64     `json:"bytes_done,omitempty"`
	BytesTotal     int64     `json:"bytes_total,omitempty"`
	Current        string    `json:"current,omitempty"`
	Code           string    `json:"code,omitempty"`
	Message        string    `json:"message,omitempty"`
	Item           string    `json:"item,omitempty"`
	Reason         string    `json:"reason,omitempty"`
	CommandPreview string    `json:"command_preview,omitempty"`
	Report         string    `json:"report,omitempty"`
}

// Emitter sends events in JSON format to a writer
type Emitter struct {
	w       io.Writer
	mu      sync.Mutex
	enabled bool
}

// NewEmitter creates an event emitter. If enabled is false, events are silenced.
func NewEmitter(w io.Writer, enabled bool) *Emitter {
	if w == nil {
		w = os.Stdout
	}
	return &Emitter{
		w:       w,
		enabled: enabled,
	}
}

// Emit sends an event as a single line JSON
func (e *Emitter) Emit(ev Event) error {
	if !e.enabled {
		return nil
	}
	e.mu.Lock()
	defer e.mu.Unlock()

	data, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = e.w.Write(data)
	return err
}

func (e *Emitter) Phase(name, status string) {
	_ = e.Emit(Event{
		Type:   TypePhase,
		Name:   name,
		Status: status,
	})
}

func (e *Emitter) Progress(phase string, percent int, bytesDone, bytesTotal int64, current string) {
	_ = e.Emit(Event{
		Type:       TypeProgress,
		Phase:      phase,
		Percent:    percent,
		BytesDone:  bytesDone,
		BytesTotal: bytesTotal,
		Current:    current,
	})
}

func (e *Emitter) Warning(code, message, item string) {
	_ = e.Emit(Event{
		Type:    TypeWarning,
		Code:    code,
		Message: message,
		Item:    item,
	})
}

func (e *Emitter) NeedPrivilege(reason, commandPreview string) {
	_ = e.Emit(Event{
		Type:           TypeNeedPrivilege,
		Reason:         reason,
		CommandPreview: commandPreview,
	})
}

func (e *Emitter) Done(status, report string) {
	_ = e.Emit(Event{
		Type:   TypeDone,
		Status: status,
		Report: report,
	})
}
