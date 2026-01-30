// Package utils provides utility functions for idempotency, error handling, and logging.
package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Logger handles logging to file and console
type Logger struct {
	logFile   *os.File
	logPath   string
	verbosity int
}

// NewLogger creates a new logger
func NewLogger(logDir string) (*Logger, error) {
	if logDir == "" {
		return &Logger{verbosity: 1}, nil
	}

	// Ensure log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logPath := filepath.Join(logDir, "servctl.log")
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &Logger{
		logFile:   file,
		logPath:   logPath,
		verbosity: 1,
	}, nil
}

// Close closes the log file
func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// log writes to log file with timestamp
func (l *Logger) log(level, message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	entry := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level, message)

	if l.logFile != nil {
		l.logFile.WriteString(entry)
	}
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.log("INFO", msg)
}

// Warn logs a warning
func (l *Logger) Warn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.log("WARN", msg)
}

// Error logs an error
func (l *Logger) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.log("ERROR", msg)
}

// Debug logs debug info
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.verbosity > 1 {
		msg := fmt.Sprintf(format, args...)
		l.log("DEBUG", msg)
	}
}

// FileExists checks if a file or directory exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DirExists checks if a directory exists
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// BackupFile creates a backup of a file before overwriting
func BackupFile(path string) (string, error) {
	if !FileExists(path) {
		return "", nil // Nothing to backup
	}

	timestamp := time.Now().Format("20060102-150405")
	backupPath := fmt.Sprintf("%s.bak.%s", path, timestamp)

	// Copy file
	src, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open source: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to copy: %w", err)
	}

	return backupPath, nil
}

// SafeWriteFile writes a file, backing up existing if present
func SafeWriteFile(path string, content []byte, perm os.FileMode, backup bool) error {
	if backup && FileExists(path) {
		backupPath, err := BackupFile(path)
		if err != nil {
			return fmt.Errorf("backup failed: %w", err)
		}
		if backupPath != "" {
			fmt.Printf("Backed up: %s → %s\n", path, backupPath)
		}
	}

	if err := os.WriteFile(path, content, perm); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ContainsLine checks if a file contains a specific line
func ContainsLine(filePath, line string) (bool, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	lines := strings.Split(string(content), "\n")
	for _, l := range lines {
		if strings.TrimSpace(l) == strings.TrimSpace(line) {
			return true, nil
		}
	}
	return false, nil
}

// AppendLineIfMissing appends a line to file if not already present (idempotent)
func AppendLineIfMissing(filePath, line string) (bool, error) {
	exists, err := ContainsLine(filePath, line)
	if err != nil {
		return false, err
	}
	if exists {
		return false, nil // Already exists
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return false, err
	}
	defer file.Close()

	if _, err := file.WriteString(line + "\n"); err != nil {
		return false, err
	}

	return true, nil
}

// EnsureDir creates directory if it doesn't exist (idempotent)
func EnsureDir(path string, perm os.FileMode) (created bool, err error) {
	if DirExists(path) {
		return false, nil
	}
	if err := os.MkdirAll(path, perm); err != nil {
		return false, err
	}
	return true, nil
}

// CheckFstabEntry checks if an fstab entry already exists
func CheckFstabEntry(device, mountPoint string) (bool, error) {
	return ContainsLine("/etc/fstab", device)
}

// ServctlError represents a servctl-specific error with context
type ServctlError struct {
	Phase       string   // Which phase failed
	Operation   string   // What operation failed
	Err         error    // Underlying error
	Remediation []string // Steps to fix
	IsCritical  bool     // Should we stop execution?
}

func (e *ServctlError) Error() string {
	return fmt.Sprintf("[%s] %s: %v", e.Phase, e.Operation, e.Err)
}

// NewCriticalError creates a critical error that should stop execution
func NewCriticalError(phase, operation string, err error, remediation ...string) *ServctlError {
	return &ServctlError{
		Phase:       phase,
		Operation:   operation,
		Err:         err,
		Remediation: remediation,
		IsCritical:  true,
	}
}

// NewWarningError creates a non-critical error
func NewWarningError(phase, operation string, err error, remediation ...string) *ServctlError {
	return &ServctlError{
		Phase:       phase,
		Operation:   operation,
		Err:         err,
		Remediation: remediation,
		IsCritical:  false,
	}
}

// FormatError formats an error for display
func FormatError(err *ServctlError) string {
	var b strings.Builder

	if err.IsCritical {
		b.WriteString("🚨 CRITICAL ERROR\n")
	} else {
		b.WriteString("⚠️  WARNING\n")
	}

	b.WriteString(fmt.Sprintf("\nPhase: %s\n", err.Phase))
	b.WriteString(fmt.Sprintf("Operation: %s\n", err.Operation))
	b.WriteString(fmt.Sprintf("Error: %v\n", err.Err))

	if len(err.Remediation) > 0 {
		b.WriteString("\nSuggested fixes:\n")
		for i, step := range err.Remediation {
			b.WriteString(fmt.Sprintf("  %d. %s\n", i+1, step))
		}
	}

	return b.String()
}

// IsRoot checks if running as root
func IsRoot() bool {
	return os.Geteuid() == 0
}

// RequireRoot returns error if not running as root
func RequireRoot(operation string) error {
	if !IsRoot() {
		return fmt.Errorf("%s requires root privileges. Run with sudo.", operation)
	}
	return nil
}

// ExpandHome expands ~ to home directory
func ExpandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}
