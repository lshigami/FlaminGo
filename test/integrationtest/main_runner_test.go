package integrationtest

import (
	"fmt"
	"os"
	"testing"
)

type testMainLogger struct{}

func (tml *testMainLogger) Logf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}
func (tml *testMainLogger) Fatalf(format string, args ...interface{}) {
	fmt.Printf("FATAL: "+format+"\n", args...)
	os.Exit(1)
}

func TestMain(m *testing.M) {
	logger := &testMainLogger{}
	var err error
	globalTestApp, err = SetupTestApp(logger)

	if err != nil {
		logger.Logf("FATAL: Test application setup failed in TestMain: %v", err)
		os.Exit(1)
	}

	if globalTestApp == nil {
		logger.Logf("FATAL: globalTestApp is nil after setup without an error, something is wrong.")
		os.Exit(1)
	}

	defer func() {
		if globalTestApp != nil && globalTestApp.DB != nil {
			sqlDB, dbErr := globalTestApp.DB.DB()
			if dbErr == nil && sqlDB != nil {
				closeErr := sqlDB.Close()
				if closeErr != nil {
					logger.Logf("Error closing database connection during teardown: %v", closeErr)
				}
			}
			logger.Logf("Test suite finished. Test DB (%s) connection potentially closed.", globalTestApp.Config.Database.Name)
		} else {
			logger.Logf("Test suite finished, but globalTestApp or DB was nil during teardown.")
		}
	}()

	exitCode := m.Run()
	os.Exit(exitCode)
}
