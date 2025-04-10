package logs

import (
	"encoding/json"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Common log patterns
var (
	// Regex to find common error messages
	errorRegex = regexp.MustCompile(`(?i)(error|exception|failed|failure|fatal|panic)`)

	// Regex to find common warning patterns
	warningRegex = regexp.MustCompile(`(?i)(warning|warn|deprecated)`)
)

// LogSummary provides a summary of analyzed logs
type LogSummary struct {
	// Total number of log entries
	TotalEntries int
	// Total number of error entries
	ErrorCount int
	// Total number of warning entries
	WarningCount int
	// Most common error messages
	CommonErrors []LogPattern
	// Most common warning messages
	CommonWarnings []LogPattern
	// Resources with the most errors
	ErrorHotspots []ResourceErrorCount
	// Potential issues detected
	PotentialIssues []string
	// Time range of logs
	TimeRange LogTimeRange
}

// LogPattern represents a recurring pattern in logs
type LogPattern struct {
	// The pattern or message
	Pattern string
	// Number of occurrences
	Count int
	// Example log entries
	Examples []LogEntry
}

// ResourceErrorCount tracks resources with the most errors
type ResourceErrorCount struct {
	// Resource name (typically pod name)
	ResourceName string
	// Error count for this resource
	ErrorCount int
}

// LogTimeRange represents the time span of analyzed logs
type LogTimeRange struct {
	// Start time of the log range
	Start time.Time
	// End time of the log range
	End time.Time
	// Duration of the log range
	Duration time.Duration
}

// ParseLogs processes a collection of log entries and extracts useful information
func ParseLogs(logs []LogEntry) LogSummary {
	if len(logs) == 0 {
		return LogSummary{}
	}

	// Initialize summary
	summary := LogSummary{
		TotalEntries: len(logs),
		TimeRange: LogTimeRange{
			Start: logs[0].Timestamp,
			End:   logs[0].Timestamp,
		},
	}

	// Maps to track unique error and warning messages
	errorMap := make(map[string]int)
	warningMap := make(map[string]int)

	// Map to track error counts by resource
	resourceErrorMap := make(map[string]int)

	// Maps to store example entries
	errorExamples := make(map[string][]LogEntry)
	warningExamples := make(map[string][]LogEntry)

	// Analyze each log entry
	for _, entry := range logs {
		// Update time range
		if entry.Timestamp.Before(summary.TimeRange.Start) {
			summary.TimeRange.Start = entry.Timestamp
		}
		if entry.Timestamp.After(summary.TimeRange.End) {
			summary.TimeRange.End = entry.Timestamp
		}

		// Process based on log level
		content := normalizeLogMessage(entry.Content)

		switch entry.LogLevel {
		case "ERROR", "FATAL":
			summary.ErrorCount++
			resourceErrorMap[entry.PodName]++

			// Extract key part of the error message
			errorKey := extractErrorKey(content)
			errorMap[errorKey]++

			// Store example (up to 3 per unique error)
			if examples, ok := errorExamples[errorKey]; ok && len(examples) < 3 {
				errorExamples[errorKey] = append(examples, entry)
			} else if !ok {
				errorExamples[errorKey] = []LogEntry{entry}
			}

		case "WARN", "WARNING":
			summary.WarningCount++

			// Extract key part of the warning message
			warningKey := extractWarningKey(content)
			warningMap[warningKey]++

			// Store example (up to 3 per unique warning)
			if examples, ok := warningExamples[warningKey]; ok && len(examples) < 3 {
				warningExamples[warningKey] = append(examples, entry)
			} else if !ok {
				warningExamples[warningKey] = []LogEntry{entry}
			}
		}
	}

	// Calculate duration
	summary.TimeRange.Duration = summary.TimeRange.End.Sub(summary.TimeRange.Start)

	// Convert error maps to sorted slices
	summary.CommonErrors = convertToPatterns(errorMap, errorExamples)
	summary.CommonWarnings = convertToPatterns(warningMap, warningExamples)

	// Convert resource error map to sorted slice
	summary.ErrorHotspots = convertToResourceErrors(resourceErrorMap)

	// Detect potential issues
	summary.PotentialIssues = detectIssues(logs, summary)

	return summary
}

// normalizeLogMessage cleans up a log message for better pattern matching
func normalizeLogMessage(message string) string {
	// Remove timestamps
	message = regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})?`).ReplaceAllString(message, "")

	// Remove UUIDs
	message = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`).ReplaceAllString(message, "UUID")

	// Remove IP addresses
	message = regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`).ReplaceAllString(message, "IP_ADDR")

	// Remove numbers that look like counts, durations, etc.
	message = regexp.MustCompile(`\b\d+\b`).ReplaceAllString(message, "N")

	// Normalize whitespace
	message = regexp.MustCompile(`\s+`).ReplaceAllString(message, " ")

	return strings.TrimSpace(message)
}

// extractErrorKey extracts a normalized key from an error message
func extractErrorKey(message string) string {
	// Look for known error patterns
	if errorMatch := errorRegex.FindString(message); errorMatch != "" {
		// Try to get the error and some context around it
		errorStart := strings.Index(strings.ToLower(message), strings.ToLower(errorMatch))
		if errorStart >= 0 {
			// Get up to 10 words after the error keyword
			messageParts := strings.Split(message[errorStart:], " ")
			if len(messageParts) > 10 {
				messageParts = messageParts[:10]
			}
			return strings.Join(messageParts, " ")
		}
	}

	// If no specific pattern, just use the first 10 words
	parts := strings.Split(message, " ")
	if len(parts) > 10 {
		parts = parts[:10]
	}
	return strings.Join(parts, " ")
}

// extractWarningKey extracts a normalized key from a warning message
func extractWarningKey(message string) string {
	// Similar logic to extractErrorKey
	if warningMatch := warningRegex.FindString(message); warningMatch != "" {
		warningStart := strings.Index(strings.ToLower(message), strings.ToLower(warningMatch))
		if warningStart >= 0 {
			messageParts := strings.Split(message[warningStart:], " ")
			if len(messageParts) > 10 {
				messageParts = messageParts[:10]
			}
			return strings.Join(messageParts, " ")
		}
	}

	parts := strings.Split(message, " ")
	if len(parts) > 10 {
		parts = parts[:10]
	}
	return strings.Join(parts, " ")
}

// convertToPatterns converts a map of message counts to a sorted slice of LogPattern
func convertToPatterns(countMap map[string]int, examplesMap map[string][]LogEntry) []LogPattern {
	patterns := make([]LogPattern, 0, len(countMap))

	for pattern, count := range countMap {
		patterns = append(patterns, LogPattern{
			Pattern:  pattern,
			Count:    count,
			Examples: examplesMap[pattern],
		})
	}

	// Sort by count, descending
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Count > patterns[j].Count
	})

	// Limit to top 10
	if len(patterns) > 10 {
		patterns = patterns[:10]
	}

	return patterns
}

// convertToResourceErrors converts a map of resource error counts to a sorted slice
func convertToResourceErrors(resourceMap map[string]int) []ResourceErrorCount {
	resources := make([]ResourceErrorCount, 0, len(resourceMap))

	for resource, count := range resourceMap {
		resources = append(resources, ResourceErrorCount{
			ResourceName: resource,
			ErrorCount:   count,
		})
	}

	// Sort by error count, descending
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].ErrorCount > resources[j].ErrorCount
	})

	// Limit to top 5
	if len(resources) > 5 {
		resources = resources[:5]
	}

	return resources
}

// detectIssues analyzes logs to find potential issues
func detectIssues(logs []LogEntry, summary LogSummary) []string {
	var issues []string

	// Check for high error rate
	errorRate := float64(summary.ErrorCount) / float64(summary.TotalEntries)
	if errorRate > 0.1 { // More than 10% errors
		issues = append(issues, "High error rate detected in logs")
	}

	// Check for error spikes
	if hasErrorSpikes(logs) {
		issues = append(issues, "Error spikes detected - possible service disruption")
	}

	// Check for repeated restarts
	if hasRepeatedRestarts(logs) {
		issues = append(issues, "Pod restart pattern detected - possible crash loop")
	}

	// Check for resource issues
	if hasResourceIssues(logs) {
		issues = append(issues, "Resource constraint issues detected (OOM, CPU throttling)")
	}

	// Check for network issues
	if hasNetworkIssues(logs) {
		issues = append(issues, "Network connectivity issues detected")
	}

	// Check for auth issues
	if hasAuthIssues(logs) {
		issues = append(issues, "Authentication or authorization issues detected")
	}

	return issues
}

// hasErrorSpikes checks if there are sudden spikes in error frequency
func hasErrorSpikes(logs []LogEntry) bool {
	if len(logs) < 100 {
		return false
	}

	// Group errors by minute
	errorsByMinute := make(map[int]int)

	// Get the baseline time
	baseTime := logs[0].Timestamp

	for _, entry := range logs {
		if entry.LogLevel == "ERROR" || entry.LogLevel == "FATAL" {
			minuteOffset := int(entry.Timestamp.Sub(baseTime).Minutes())
			errorsByMinute[minuteOffset]++
		}
	}

	// Check for any minutes with unusually high error counts
	var errorCounts []int
	for _, count := range errorsByMinute {
		errorCounts = append(errorCounts, count)
	}

	// Need at least a few minutes of data
	if len(errorCounts) < 3 {
		return false
	}

	// Calculate average and standard deviation
	avg := average(errorCounts)
	stdDev := standardDeviation(errorCounts, avg)

	// Check for any minute with error count > avg + 2*stdDev
	for _, count := range errorCounts {
		if float64(count) > avg+2*stdDev && count > 5 {
			return true
		}
	}

	return false
}

// hasRepeatedRestarts checks for patterns indicating frequent restarts
func hasRepeatedRestarts(logs []LogEntry) bool {
	restartCount := 0

	for _, entry := range logs {
		content := strings.ToLower(entry.Content)
		if strings.Contains(content, "started container") ||
			strings.Contains(content, "starting container") ||
			strings.Contains(content, "restarting container") {
			restartCount++
		}
	}

	// More than 3 restarts might indicate a problem
	return restartCount > 3
}

// hasResourceIssues checks for resource-related problems
func hasResourceIssues(logs []LogEntry) bool {
	for _, entry := range logs {
		content := strings.ToLower(entry.Content)
		if strings.Contains(content, "out of memory") ||
			strings.Contains(content, "oom killed") ||
			strings.Contains(content, "memory limit") ||
			strings.Contains(content, "cpu throttling") {
			return true
		}
	}
	return false
}

// hasNetworkIssues checks for network-related problems
func hasNetworkIssues(logs []LogEntry) bool {
	for _, entry := range logs {
		content := strings.ToLower(entry.Content)
		if strings.Contains(content, "connection refused") ||
			strings.Contains(content, "connection timeout") ||
			strings.Contains(content, "unable to connect") ||
			strings.Contains(content, "network error") {
			return true
		}
	}
	return false
}

// hasAuthIssues checks for authentication-related problems
func hasAuthIssues(logs []LogEntry) bool {
	for _, entry := range logs {
		content := strings.ToLower(entry.Content)
		if strings.Contains(content, "unauthorized") ||
			strings.Contains(content, "forbidden") ||
			strings.Contains(content, "permission denied") ||
			strings.Contains(content, "access denied") {
			return true
		}
	}
	return false
}

// average calculates the average of an integer slice
func average(values []int) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0
	for _, v := range values {
		sum += v
	}

	return float64(sum) / float64(len(values))
}

// standardDeviation calculates the standard deviation
func standardDeviation(values []int, avg float64) float64 {
	if len(values) < 2 {
		return 0
	}

	variance := 0.0
	for _, v := range values {
		variance += (float64(v) - avg) * (float64(v) - avg)
	}

	variance /= float64(len(values) - 1)
	return variance
}

// LogsToJSON converts log entries to JSON format
func LogsToJSON(logs []LogEntry) (string, error) {
	jsonData, err := json.MarshalIndent(logs, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// SummaryToJSON converts a log summary to JSON format
func SummaryToJSON(summary LogSummary) (string, error) {
	jsonData, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
