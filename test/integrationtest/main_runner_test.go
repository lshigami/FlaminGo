package integrationtest

import (
	"fmt"
	"os"
	"testing"
)

// TestMain will be run by the 'go test' command before any other tests in this package.
func TestMain(m *testing.M) {
	// Optional: Check env var here to skip ALL integration tests early
	// if os.Getenv("RUN_INTEGRATION_TESTS") == "" {
	// 	fmt.Println("Skipping all integration tests in this package: RUN_INTEGRATION_TESTS env var not set")
	// 	os.Exit(0) // Exit with success code 0 as tests are skipped, not failed.
	// }

	// Use a dummy *testing.T for SetupTestApp as it's called outside a specific test func.
	// SetupTestApp should ideally not call t.Fatal itself but return an error.
	// However, for simplicity with require.NoError, this pattern is common.
	var success bool
	defer func() {
		if !success && globalTestApp == nil {
			fmt.Println("SetupTestApp likely failed before initializing globalTestApp")
		}
	}()

	// We need a *testing.T to pass to SetupTestApp
	// Create a "master" test object for setup/teardown
	// This is a bit of a workaround for TestMain not having its own T
	// More robust would be for SetupTestApp to return an error.
	masterT := new(testing.T) // This T will not report test results, it's for setup logic

	globalTestApp = SetupTestApp(masterT) // Initialize the global instance
	if masterT.Failed() {                 // Check if SetupTestApp called Fatal on masterT
		fmt.Fprintln(os.Stderr, "FATAL: Test application setup failed in TestMain.")
		// Perform any minimal cleanup if possible before exiting
		if globalTestApp != nil && globalTestApp.DB != nil {
			sqlDB, _ := globalTestApp.DB.DB()
			sqlDB.Close()
		}
		os.Exit(1)
	}
	success = true // Mark setup as successful

	// Run all tests in the package
	exitCode := m.Run()

	// Teardown
	if globalTestApp != nil && globalTestApp.DB != nil {
		sqlDB, _ := globalTestApp.DB.DB()
		err := sqlDB.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error closing database connection: %v\n", err)
		}
		fmt.Printf("Test suite finished. Test DB (%s) connection closed.\n", globalTestApp.Config.Database.Name)
		// Add logic here to drop the test database if it's persistent and you want to clean it up.
		// e.g., dropDB(globalTestApp.Config)
	} else {
		fmt.Println("Test suite finished, but globalTestApp or DB was nil during teardown.")
	}

	os.Exit(exitCode)
}
