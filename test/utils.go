package test

import (
	"os"
	"path"
	"runtime"
	"testing"

	"go.uber.org/zap"
)

func ProjectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	return path.Join(filename, "..", "..")
}

func PathForAsset(t *testing.T, name string) string {
	filePath := path.Join(ProjectRoot(), "test/assets", name)
	info, err := os.Stat(filePath)
	if err != nil {
		t.Errorf("error getting file info: %v", err)
	}
	if info.Mode().Perm()&0444 != 0444 {
		t.Errorf("Cannot read error getting file info: %v", err)
	}
	return filePath
}

func GetLogger(t *testing.T) *zap.Logger {
	log, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	return log
}
