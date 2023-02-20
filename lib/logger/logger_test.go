package logger

import (
	"os"
	"testing"
)

func TestLogger(t *testing.T) {
	l := NewLogger(WithLevel(TraceLevel))
	h1 := NewHelper(l).WithFields(map[string]interface{}{"key1": "val1"})
	h1.Trace("trace_msg1")
	h1.Log(WarnLevel, "warn_msg1")

	h2 := NewHelper(l).WithFields(map[string]interface{}{"key2": "val2"})
	h2.Trace("trace_msg2")
	h2.Warn("warn_msg2")

	l.Fields(map[string]interface{}{"key3": "val4"}).Log(InfoLevel, "test_msg")

	Fields(map[string]interface{}{"key1": "val1"})
	Info("info")

	h3 := NewHelper(NewLogger(WithLevel(TraceLevel), WithOutput(os.Stdout)))
	h3.Info("info => ")
}
