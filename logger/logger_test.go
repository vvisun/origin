package cherryLogger

import (
	"testing"
	"time"
)

func BenchmarkWrite(b *testing.B) {
	config := defaultConsoleConfig()
	config.EnableConsole = true
	config.EnableWriteFile = true
	config.FileLinkPath = "logs/log1.log"
	config.FilePathFormat = "logs/log1_%Y%m%d%H%M.log"

	log1 := NewConfigLogger(config)

	for i := 0; i < b.N; i++ {
		log1.Debug(time.Now().UnixMilli(), "test")
	}
}
