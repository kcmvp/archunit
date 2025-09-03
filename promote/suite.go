package promote

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/kcmvp/archunit/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var (
	targetDir  string
	modulePath string
)

var _ assert.TestingT = (*T)(nil)

func init() {
	targetDir, modulePath = utils.ProjectInfo()
	targetDir = filepath.Join(targetDir, "target")

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		log.Printf("archunit-testing: failed to create target directory at %s: %v", targetDir, err)
	}
}

// T is a wrapper around the standard *testing.T that provides additional functionality, such as logging to a file.
type T struct {
	*testing.T
	logWriter io.Writer
}

// NewT creates a new test context that extends the standard *testing.T,
// using the provided writer for logging. The writer's lifecycle is managed elsewhere.
func NewT(t *testing.T, writer io.Writer) *T {
	return &T{
		T:         t,
		logWriter: writer,
	}
}

// Log formats its arguments using default formats and records the text in the error log.
// It also writes the text to the test's dedicated log file.
func (t *T) Log(args ...any) {
	t.T.Helper()
	t.T.Log(args...)
	if t.logWriter != nil {
		_, _ = fmt.Fprintln(t.logWriter, args...)
	}
}

// Logf formats its arguments and records the text in the error log.
// It also writes the text to the test's dedicated log file.
func (t *T) Logf(format string, args ...any) {
	t.T.Helper()
	t.T.Logf(format, args...)
	if t.logWriter != nil {
		_, _ = fmt.Fprintf(t.logWriter, format+"\n", args...)
	}
}

// Error is equivalent to Log followed by Fail.
func (t *T) Error(args ...any) {
	t.T.Helper()
	t.T.Error(args...)
	if t.logWriter != nil {
		_, _ = fmt.Fprintln(t.logWriter, args...)
	}
}

// Errorf is equivalent to Logf followed by Fail.
func (t *T) Errorf(format string, args ...any) {
	t.T.Helper()
	t.T.Errorf(format, args...)
	if t.logWriter != nil {
		_, _ = fmt.Fprintf(t.logWriter, format+"\n", args...)
	}
}

// Fatal is equivalent to Log followed by FailNow.
func (t *T) Fatal(args ...any) {
	t.T.Helper()
	if t.logWriter != nil {
		_, _ = fmt.Fprintln(t.logWriter, args...) // Write to log before failing
	}
	t.T.Fatal(args...)
}

// Fatalf is equivalent to Logf followed by FailNow.
func (t *T) Fatalf(format string, args ...any) {
	t.T.Helper()
	if t.logWriter != nil {
		_, _ = fmt.Fprintf(t.logWriter, format+"\n", args...) // Write to log before failing
	}
	t.T.Fatalf(format, args...)
}

// ArchSuite is a base for test suites that wish to leverage the custom T type
// for automatic logging and other features.
// By embedding this ArchSuite, your
// test methods will have access to enhanced assertion methods (like Equal,
// NoError, etc.) that automatically log to both the console and a file.
type ArchSuite struct {
	suite.Suite
	mu             sync.RWMutex
	tt             *T
	require        *require.Assertions
	logPrefixCache string // Cache for the log file prefix.
}

func (suite *ArchSuite) TT() *T {
	return suite.tt
}

func (suite *ArchSuite) SetS(s suite.TestingSuite) {
	suite.Suite.SetS(s)
	suite.mu.Lock()
	defer suite.mu.Unlock()

	// Compute and cache the log prefix once, using the concrete suite 's'
	// provided by the hook. We don't need to store the suite instance itself.
	suiteType := reflect.TypeOf(s).Elem()
	suitePkgPath := suiteType.PkgPath()
	relativePkgPath := strings.TrimPrefix(suitePkgPath, modulePath)
	relativePkgPath = strings.TrimPrefix(relativePkgPath, "/")
	folderPart := strings.ReplaceAll(relativePkgPath, "/", "_")
	suiteName := suiteType.Name()

	if folderPart != "" {
		suite.logPrefixCache = fmt.Sprintf("%s_%s_", folderPart, suiteName)
	} else {
		suite.logPrefixCache = fmt.Sprintf("%s_", suiteName)
	}
}

// SetT is called by the suite runner. We override it to wrap the standard
// *testing.T with our custom T, making it available to all test methods.
// It also re-initializes the embedded Assertions and Require fields to use
// our custom T, ensuring that calls like `s.Assertions.Equal(...)` are
// also logged correctly.
func (suite *ArchSuite) SetT(t *testing.T) {
	suite.Suite.SetT(t)
	// Only create log files for the actual test methods, not the suite runner itself.
	// Test methods are run as sub-tests, so their names will contain a "/".
	if !strings.Contains(t.Name(), "/") {
		// This is the top-level suite setup. We initialize the helpers
		// but do not create a log file.
		suite.mu.Lock()
		defer suite.mu.Unlock()
		suite.tt = NewT(t, nil) // Pass a nil writer
		suite.Assertions = assert.New(suite.tt)
		suite.require = require.New(suite.tt)
		return
	}

	// For test methods, create an in-memory buffer to capture logs.
	// The log file will be created inside a suite-specific directory for better organization.
	logBuffer := new(bytes.Buffer)

	t.Cleanup(func() {
		// After the test, check if it failed.
		if t.Failed() {
			// If it failed, write the buffered logs to a file.
			prefix := suite.logPrefix()
			if prefix == "" {
				// Should not happen in a normal test method run, but as a safeguard.
				t.Errorf("promote: could not determine log prefix")
				return
			}
			testMethodName := filepath.Base(t.Name())

			logFileName := fmt.Sprintf("%s%s.log", prefix, testMethodName)
			logFilePath := filepath.Join(targetDir, logFileName)

			// Create the file and write the FAIL banner as the first line.
			file, err := os.Create(logFilePath)
			if err != nil {
				t.Errorf("promote: failed to write log file %s: %v", logFilePath, err)
				return
			}
			defer file.Close()

			_, _ = fmt.Fprintf(file, "--- FAIL: %s ---\n", t.Name())
			_, _ = file.Write(logBuffer.Bytes())
		}
	})

	suite.mu.Lock()
	defer suite.mu.Unlock()
	suite.tt = NewT(t, logBuffer)
	suite.Assertions = assert.New(suite.tt)
	suite.require = require.New(suite.tt)
}

// logPrefix generates a sanitized prefix for log files based on the suite's package path and name.
// e.g., a suite 'ArtifactSuite' in "github.com/kcmvp/archunit/internal" will produce "internal_ArtifactSuite_".
func (suite *ArchSuite) logPrefix() string {
	suite.mu.RLock()
	defer suite.mu.RUnlock()
	return suite.logPrefixCache
}

// Require returns a require context for suite.
func (suite *ArchSuite) Require() *require.Assertions {
	suite.mu.Lock()
	defer suite.mu.Unlock()
	if suite.require == nil {
		// This behavior matches the original testify/suite, which panics
		// if Require() is called before the test context is set.
		panic("'Require' must not be called before 'Run' or 'SetT'")
	}
	return suite.require
}

// Errorf makes ArchSuite satisfy the assert.TestingT interface.
// This allows the suite instance to be passed directly to assert functions,
// ensuring that both console and file logging are triggered.
func (suite *ArchSuite) Errorf(format string, args ...any) {
	suite.mu.RLock()
	defer suite.mu.RUnlock()
	suite.tt.Errorf(format, args...)
}

// --- Shadowed Assertion Methods ---
// By shadowing the most common assertion methods from the embedded testify suite,
// we can ensure that calls like `suite.Equal(...)` are directed to our
// custom `tt` object, which provides file logging.

// Equal asserts that two objects are equal.
func (suite *ArchSuite) Equal(expected, actual any, msgAndArgs ...any) bool {
	suite.mu.RLock()
	defer suite.mu.RUnlock()
	suite.tt.Helper()
	return assert.Equal(suite.tt, expected, actual, msgAndArgs...)
}

// ElementsMatch asserts that the specified listA and listB have the same elements.
func (suite *ArchSuite) ElementsMatch(listA, listB any, msgAndArgs ...any) bool {
	suite.mu.RLock()
	defer suite.mu.RUnlock()
	suite.tt.Helper()
	return assert.ElementsMatch(suite.tt, listA, listB, msgAndArgs...)
}

// NoError asserts that a function returned no error (i.e. `nil`).
func (suite *ArchSuite) NoError(err error, msgAndArgs ...any) bool {
	suite.mu.RLock()
	defer suite.mu.RUnlock()
	suite.tt.Helper()
	return assert.NoError(suite.tt, err, msgAndArgs...)
}

// NotNil asserts that a pointer is not nil.
func (suite *ArchSuite) NotNil(object any, msgAndArgs ...any) bool {
	suite.mu.RLock()
	defer suite.mu.RUnlock()
	suite.tt.Helper()
	return assert.NotNil(suite.tt, object, msgAndArgs...)
}

// True asserts that the specified value is true.
func (suite *ArchSuite) True(value bool, msgAndArgs ...any) bool {
	suite.mu.RLock()
	defer suite.mu.RUnlock()
	suite.tt.Helper()
	return assert.True(suite.tt, value, msgAndArgs...)
}

// NotEmpty asserts that the specified object is not empty.
func (suite *ArchSuite) NotEmpty(object any, msgAndArgs ...any) bool {
	suite.mu.RLock()
	defer suite.mu.RUnlock()
	suite.tt.Helper()
	return assert.NotEmpty(suite.tt, object, msgAndArgs...)
}

// Run is a convenience function that runs a test suite that embeds promote.ArchSuite.
func Run(t *testing.T, as suite.TestingSuite) {
	cleanTestLog(as)
	suite.Run(t, as)
}

func cleanTestLog(as suite.TestingSuite) {
	// Get type info directly from the concrete suite instance passed in.
	// This is more direct and avoids manually calling SetS.
	suiteType := reflect.TypeOf(as)
	if suiteType.Kind() == reflect.Ptr {
		suiteType = suiteType.Elem()
	}

	if suiteType.Kind() != reflect.Struct {
		// Not a suite struct we can analyze, so we can't clean up.
		return
	}

	// Check if this suite embeds an ArchSuite. If not, there's nothing to clean.
	if _, ok := suiteType.FieldByName("ArchSuite"); !ok {
		return
	}

	suitePkgPath := suiteType.PkgPath()
	relativePkgPath := strings.TrimPrefix(suitePkgPath, modulePath)
	relativePkgPath = strings.TrimPrefix(relativePkgPath, "/")
	folderPart := strings.ReplaceAll(relativePkgPath, "/", "_")
	suiteName := suiteType.Name()

	var filePrefix string
	if folderPart != "" {
		filePrefix = fmt.Sprintf("%s_%s_", folderPart, suiteName)
	} else {
		filePrefix = fmt.Sprintf("%s_", suiteName)
	}

	if entries, err := os.ReadDir(targetDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasPrefix(entry.Name(), filePrefix) && strings.HasSuffix(entry.Name(), ".log") {
				_ = os.Remove(filepath.Join(targetDir, entry.Name()))
			}
		}
	}
}
