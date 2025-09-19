package internal

import (
	"context"
	"log/slog"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// TflogHandler implements slog.Handler to redirect logs to tflog
type TflogHandler struct {
	ctx    context.Context
	attrs  []slog.Attr
	groups []string
}

// NewTflogHandler creates a new handler that redirects logs to tflog
func NewTflogHandler(ctx context.Context) *TflogHandler {
	return &TflogHandler{
		ctx: ctx,
	}
}

func (h *TflogHandler) clone() *TflogHandler {
	return &TflogHandler{
		ctx:    h.ctx,
		attrs:  append([]slog.Attr{}, h.attrs...),
		groups: append([]string{}, h.groups...),
	}
}

func (h *TflogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	// Always enabled - tflog will handle filtering
	return true
}

// Handle processes a log record
func (h *TflogHandler) Handle(ctx context.Context, record slog.Record) error {
	fields := make(map[string]interface{})

	// Add handler attributes
	for _, attr := range h.attrs {
		key := attr.Key
		fields[key] = attr.Value.Any()
	}
	// Add groups as structured fields rather than key prefixes
	if len(h.groups) > 0 {
		fields["slog_groups"] = h.groups
	}

	// Add record attributes
	record.Attrs(func(attr slog.Attr) bool {
		fields[attr.Key] = attr.Value.Any()
		return true
	})

	logCtx := h.ctx

	message := record.Message

	// Map slog groups to tflog subsystems
	// If we have groups, use the first group as subsystem
	if len(h.groups) > 0 {
		logCtx = tflog.NewSubsystem(logCtx, h.groups[0])

		// If we have nested groups, include them in the message or fields
		if len(h.groups) > 1 {
			fields["slog_subgroups"] = h.groups[1:]
		}
	}

	// No groups, use regular tflog functions
	switch record.Level {
	case slog.LevelDebug:
		tflog.Debug(logCtx, message, fields)
	case slog.LevelInfo:
		tflog.Info(logCtx, message, fields)
	case slog.LevelWarn:
		tflog.Warn(logCtx, message, fields)
	case slog.LevelError:
		tflog.Error(logCtx, message, fields)
	default:
		tflog.Info(logCtx, message, fields)
	}

	return nil
}

// WithAttrs returns a new handler with added attributes
func (h *TflogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h2 := h.clone()
	h2.attrs = append(h2.attrs, attrs...)
	return h2
}

// WithGroup returns a new handler with the specified group
func (h *TflogHandler) WithGroup(name string) slog.Handler {
	h2 := h.clone()
	h2.groups = append(h2.groups, name)
	return h2
}
