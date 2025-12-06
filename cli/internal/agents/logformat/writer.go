package logformat

import (
	"bufio"
	"bytes"
	"io"
	"sync"
)

// FormattingWriter wraps a writer and formats stream-json events as they arrive.
// It buffers incoming data and processes complete JSON lines.
type FormattingWriter struct {
	output    io.Writer
	formatter *Formatter
	buffer    bytes.Buffer
	mu        sync.Mutex
}

// NewFormattingWriter creates a new FormattingWriter that writes formatted
// output to the given writer.
func NewFormattingWriter(output io.Writer) *FormattingWriter {
	return &FormattingWriter{
		output:    output,
		formatter: NewFormatter(),
	}
}

// NewFormattingWriterWithFormatter creates a FormattingWriter with custom formatter settings.
func NewFormattingWriterWithFormatter(output io.Writer, formatter *Formatter) *FormattingWriter {
	return &FormattingWriter{
		output:    output,
		formatter: formatter,
	}
}

// Write implements io.Writer. It buffers incoming data and processes
// complete JSON lines as they arrive.
func (w *FormattingWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Add to buffer
	w.buffer.Write(p)

	// Process complete lines
	for {
		line, err := w.buffer.ReadBytes('\n')
		if err == io.EOF {
			// Incomplete line, put it back
			w.buffer.Write(line)
			break
		}
		if err != nil {
			return len(p), err
		}

		// Parse and format the line
		w.processLine(line)
	}

	return len(p), nil
}

// Flush processes any remaining buffered data.
func (w *FormattingWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.buffer.Len() > 0 {
		w.processLine(w.buffer.Bytes())
		w.buffer.Reset()
	}
	return nil
}

func (w *FormattingWriter) processLine(line []byte) {
	line = bytes.TrimSpace(line)
	if len(line) == 0 {
		return
	}

	event, err := ParseEvent(line)
	if err != nil {
		// Not valid JSON, skip silently
		return
	}

	formatted := w.formatter.Format(event)
	if formatted != "" {
		w.output.Write([]byte(formatted))
	}
}

// FormatFile reads a stream-json log file and writes formatted output.
// This is useful for reformatting existing log files.
func FormatFile(input io.Reader, output io.Writer) error {
	formatter := NewFormatter()
	scanner := bufio.NewScanner(input)

	// Handle very long lines
	const maxLineSize = 10 * 1024 * 1024 // 10MB
	buf := make([]byte, maxLineSize)
	scanner.Buffer(buf, maxLineSize)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		event, err := ParseEvent(line)
		if err != nil {
			continue
		}

		formatted := formatter.Format(event)
		if formatted != "" {
			if _, err := output.Write([]byte(formatted)); err != nil {
				return err
			}
		}
	}

	return scanner.Err()
}

// DualWriter writes to both a raw JSON file and a formatted output.
// This allows keeping the original JSON for machine processing while
// providing human-readable output.
type DualWriter struct {
	raw       io.Writer
	formatted *FormattingWriter
}

// NewDualWriter creates a writer that outputs to both raw JSON and formatted destinations.
func NewDualWriter(raw io.Writer, formatted io.Writer) *DualWriter {
	return &DualWriter{
		raw:       raw,
		formatted: NewFormattingWriter(formatted),
	}
}

// Write implements io.Writer, writing to both outputs.
func (w *DualWriter) Write(p []byte) (n int, err error) {
	// Write to raw first
	n, err = w.raw.Write(p)
	if err != nil {
		return n, err
	}

	// Then to formatted
	w.formatted.Write(p)
	return n, nil
}

// Flush flushes any buffered formatted output.
func (w *DualWriter) Flush() error {
	return w.formatted.Flush()
}
