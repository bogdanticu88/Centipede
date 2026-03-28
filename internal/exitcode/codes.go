package exitcode

// Exit codes for CI/CD integration
const (
	// Success
	Success = 0

	// General errors
	ErrGeneral = 1

	// Configuration errors
	ErrConfig = 2

	// Data/Input errors
	ErrData = 3

	// Execution errors
	ErrExecution = 4

	// Anomaly detection results (non-fatal)
	// Used to indicate detection status for CI/CD decision making
	WarningDetected    = 10 // Anomalies with warning-level score (2-3)
	CriticalDetected   = 11 // Anomalies with critical score (4+)
	SuspiciousDetected = 12 // Honeypot hit detected
)

// IsError returns true if code indicates an error (exit code > 1)
func IsError(code int) bool {
	return code > 1
}

// IsCritical returns true if code indicates critical findings
func IsCritical(code int) bool {
	return code == CriticalDetected || code == SuspiciousDetected
}
