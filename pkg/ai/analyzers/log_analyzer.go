package analyzers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"kube-ai/pkg/ai"
	"kube-ai/pkg/k8s/logs"
)

// LogAnalysisResult represents the AI-generated analysis of logs
type LogAnalysisResult struct {
	// Summary of the logs
	Summary string `json:"summary"`

	// Identified root causes
	RootCauses []string `json:"rootCauses"`

	// Potential solutions
	Solutions []string `json:"solutions"`

	// Additional information that might be helpful
	AdditionalInfo []string `json:"additionalInfo"`

	// Severity level (Low, Medium, High, Critical)
	Severity string `json:"severity"`
}

// LogAnalyzer handles AI analysis of Kubernetes logs
type LogAnalyzer struct {
	aiService *ai.Service
}

// NewLogAnalyzer creates a new log analyzer
func NewLogAnalyzer(aiService *ai.Service) *LogAnalyzer {
	return &LogAnalyzer{
		aiService: aiService,
	}
}

// AnalyzeLogs uses AI to analyze log entries and provide insights
func (a *LogAnalyzer) AnalyzeLogs(ctx context.Context, logEntries []logs.LogEntry, summary logs.LogSummary) (*LogAnalysisResult, error) {
	// Prepare the AI prompt with log information
	prompt := a.buildLogAnalysisPrompt(logEntries, summary)

	// Call the AI service for analysis
	response, err := a.aiService.Query(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("error getting AI analysis: %w", err)
	}

	// Parse the AI response into a structured result
	result, err := parseAIResponse(response)
	if err != nil {
		return nil, fmt.Errorf("error parsing AI response: %w", err)
	}

	return result, nil
}

// buildLogAnalysisPrompt creates a prompt for the AI to analyze logs
func (a *LogAnalyzer) buildLogAnalysisPrompt(logEntries []logs.LogEntry, summary logs.LogSummary) string {
	var sb strings.Builder

	// System context
	sb.WriteString("You are an expert Kubernetes troubleshooter. Analyze these logs to identify issues, ")
	sb.WriteString("determine root causes, and suggest solutions.\n\n")

	// Add summary statistics
	sb.WriteString("## Log Summary\n")
	sb.WriteString(fmt.Sprintf("- Total log entries: %d\n", summary.TotalEntries))
	sb.WriteString(fmt.Sprintf("- Error count: %d\n", summary.ErrorCount))
	sb.WriteString(fmt.Sprintf("- Warning count: %d\n", summary.WarningCount))
	sb.WriteString(fmt.Sprintf("- Time range: %s to %s (%s)\n\n",
		summary.TimeRange.Start.Format(time.RFC3339),
		summary.TimeRange.End.Format(time.RFC3339),
		summary.TimeRange.Duration.String()))

	// Add error hotspots
	if len(summary.ErrorHotspots) > 0 {
		sb.WriteString("## Error Hotspots\n")
		for _, hotspot := range summary.ErrorHotspots {
			sb.WriteString(fmt.Sprintf("- %s: %d errors\n", hotspot.ResourceName, hotspot.ErrorCount))
		}
		sb.WriteString("\n")
	}

	// Add common errors
	if len(summary.CommonErrors) > 0 {
		sb.WriteString("## Common Errors\n")
		for _, pattern := range summary.CommonErrors {
			sb.WriteString(fmt.Sprintf("- Pattern: %s (count: %d)\n", pattern.Pattern, pattern.Count))
			if len(pattern.Examples) > 0 {
				sb.WriteString(fmt.Sprintf("  Example: %s\n", pattern.Examples[0].Content))
			}
		}
		sb.WriteString("\n")
	}

	// Add potential issues already detected
	if len(summary.PotentialIssues) > 0 {
		sb.WriteString("## Detected Issues\n")
		for _, issue := range summary.PotentialIssues {
			sb.WriteString(fmt.Sprintf("- %s\n", issue))
		}
		sb.WriteString("\n")
	}

	// Add representative log samples
	// We'll include a mix of errors, warnings, and regular logs
	sb.WriteString("## Log Samples\n")

	// Add error samples (up to 10)
	errorCount := 0
	for _, entry := range logEntries {
		if entry.LogLevel == "ERROR" || entry.LogLevel == "FATAL" {
			sb.WriteString(fmt.Sprintf("[%s] [%s] %s\n",
				entry.Timestamp.Format(time.RFC3339),
				entry.LogLevel,
				entry.Content))
			errorCount++
			if errorCount >= 10 {
				break
			}
		}
	}

	// Add warning samples (up to 5)
	warningCount := 0
	for _, entry := range logEntries {
		if entry.LogLevel == "WARN" || entry.LogLevel == "WARNING" {
			sb.WriteString(fmt.Sprintf("[%s] [%s] %s\n",
				entry.Timestamp.Format(time.RFC3339),
				entry.LogLevel,
				entry.Content))
			warningCount++
			if warningCount >= 5 {
				break
			}
		}
	}

	// Add some regular logs for context (up to 5)
	infoCount := 0
	for _, entry := range logEntries {
		if entry.LogLevel == "INFO" {
			sb.WriteString(fmt.Sprintf("[%s] [%s] %s\n",
				entry.Timestamp.Format(time.RFC3339),
				entry.LogLevel,
				entry.Content))
			infoCount++
			if infoCount >= 5 {
				break
			}
		}
	}

	sb.WriteString("\n")

	// Add request for specific analysis
	sb.WriteString("## Analysis Request\n")
	sb.WriteString("Based on the logs and summary provided, please analyze the following:\n")
	sb.WriteString("1. Provide a brief summary of the issues observed in the logs\n")
	sb.WriteString("2. Identify the most likely root causes of the issues\n")
	sb.WriteString("3. Suggest specific solutions to address the problems\n")
	sb.WriteString("4. Add any additional information or context that might be helpful\n")
	sb.WriteString("5. Assess the severity (Low, Medium, High, Critical)\n\n")

	sb.WriteString("Format your response as JSON with the following structure:\n")
	sb.WriteString("```json\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"summary\": \"Brief description of the issues\",\n")
	sb.WriteString("  \"rootCauses\": [\"Cause 1\", \"Cause 2\", ...],\n")
	sb.WriteString("  \"solutions\": [\"Solution 1\", \"Solution 2\", ...],\n")
	sb.WriteString("  \"additionalInfo\": [\"Info 1\", \"Info 2\", ...],\n")
	sb.WriteString("  \"severity\": \"Low|Medium|High|Critical\"\n")
	sb.WriteString("}\n")
	sb.WriteString("```\n")

	return sb.String()
}

// parseAIResponse parses the AI response into a structured LogAnalysisResult
func parseAIResponse(response string) (*LogAnalysisResult, error) {
	// Extract JSON object from the response
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")

	if jsonStart < 0 || jsonEnd < 0 || jsonEnd <= jsonStart {
		// If no JSON object found, try to create a structured response based on the text
		lines := strings.Split(response, "\n")

		// Create a default result
		result := &LogAnalysisResult{
			Summary:        "The AI provided an unstructured response.",
			RootCauses:     []string{},
			Solutions:      []string{},
			AdditionalInfo: []string{response},
			Severity:       "Medium",
		}

		// Try to extract meaningful information from the text
		var section string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Check if this is a section header
			if strings.HasPrefix(line, "##") || strings.HasPrefix(line, "#") {
				section = strings.ToLower(strings.TrimLeft(line, "# "))
				continue
			}

			// Based on sections, populate the result
			switch {
			case strings.Contains(section, "summar"):
				result.Summary += line + " "
			case strings.Contains(section, "root") || strings.Contains(section, "cause"):
				if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
					result.RootCauses = append(result.RootCauses, strings.TrimLeft(line, "- *"))
				}
			case strings.Contains(section, "solution") || strings.Contains(section, "recommend"):
				if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
					result.Solutions = append(result.Solutions, strings.TrimLeft(line, "- *"))
				}
			case strings.Contains(section, "additional") || strings.Contains(section, "info"):
				if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
					result.AdditionalInfo = append(result.AdditionalInfo, strings.TrimLeft(line, "- *"))
				}
			case strings.Contains(section, "sever"):
				lower := strings.ToLower(line)
				if strings.Contains(lower, "critical") {
					result.Severity = "Critical"
				} else if strings.Contains(lower, "high") {
					result.Severity = "High"
				} else if strings.Contains(lower, "medium") {
					result.Severity = "Medium"
				} else if strings.Contains(lower, "low") {
					result.Severity = "Low"
				}
			}
		}

		return result, nil
	}

	// Extract the JSON part
	jsonStr := response[jsonStart : jsonEnd+1]

	// Try to parse the JSON
	var result LogAnalysisResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("error parsing response JSON: %w", err)
	}

	// Ensure we have valid values for required fields
	if result.Summary == "" {
		result.Summary = "No summary provided by AI analysis."
	}

	if len(result.RootCauses) == 0 {
		result.RootCauses = []string{"No root causes identified in AI analysis."}
	}

	if len(result.Solutions) == 0 {
		result.Solutions = []string{"No solutions provided by AI analysis."}
	}

	if result.Severity == "" {
		result.Severity = "Medium"
	}

	return &result, nil
}

// AnalyzeErrorLogs focuses analysis specifically on error logs
func (a *LogAnalyzer) AnalyzeErrorLogs(ctx context.Context, logEntries []logs.LogEntry) (*LogAnalysisResult, error) {
	// Filter for error logs only
	errorLogs := make([]logs.LogEntry, 0)
	for _, entry := range logEntries {
		if entry.LogLevel == "ERROR" || entry.LogLevel == "FATAL" {
			errorLogs = append(errorLogs, entry)
		}
	}

	if len(errorLogs) == 0 {
		return &LogAnalysisResult{
			Summary:        "No error logs found",
			RootCauses:     []string{"No errors detected in logs"},
			Solutions:      []string{"No action needed"},
			AdditionalInfo: []string{"The logs contain no error or fatal level entries"},
			Severity:       "Low",
		}, nil
	}

	// Create a summary just for the error logs
	summary := logs.ParseLogs(errorLogs)

	// Build a specialized prompt for error analysis
	prompt := a.buildErrorAnalysisPrompt(errorLogs, summary)

	// Call the AI service for analysis
	response, err := a.aiService.Query(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("error getting AI error analysis: %w", err)
	}

	// Parse the AI response
	result, err := parseAIResponse(response)
	if err != nil {
		return nil, fmt.Errorf("error parsing AI error analysis response: %w", err)
	}

	return result, nil
}

// buildErrorAnalysisPrompt creates a specialized prompt for error analysis
func (a *LogAnalyzer) buildErrorAnalysisPrompt(errorLogs []logs.LogEntry, summary logs.LogSummary) string {
	var sb strings.Builder

	// System context
	sb.WriteString("You are an expert Kubernetes troubleshooter. Analyze these error logs to identify issues, ")
	sb.WriteString("determine root causes, and suggest solutions. Focus specifically on the errors.\n\n")

	// Add summary statistics
	sb.WriteString("## Error Log Summary\n")
	sb.WriteString(fmt.Sprintf("- Total error entries: %d\n", summary.TotalEntries))
	sb.WriteString(fmt.Sprintf("- Time range: %s to %s (%s)\n\n",
		summary.TimeRange.Start.Format(time.RFC3339),
		summary.TimeRange.End.Format(time.RFC3339),
		summary.TimeRange.Duration.String()))

	// Add error hotspots
	if len(summary.ErrorHotspots) > 0 {
		sb.WriteString("## Error Hotspots\n")
		for _, hotspot := range summary.ErrorHotspots {
			sb.WriteString(fmt.Sprintf("- %s: %d errors\n", hotspot.ResourceName, hotspot.ErrorCount))
		}
		sb.WriteString("\n")
	}

	// Add common errors
	if len(summary.CommonErrors) > 0 {
		sb.WriteString("## Common Errors\n")
		for _, pattern := range summary.CommonErrors {
			sb.WriteString(fmt.Sprintf("- Pattern: %s (count: %d)\n", pattern.Pattern, pattern.Count))
			if len(pattern.Examples) > 0 {
				sb.WriteString(fmt.Sprintf("  Example: %s\n", pattern.Examples[0].Content))
			}
		}
		sb.WriteString("\n")
	}

	// Add error log samples (up to 20)
	sb.WriteString("## Error Log Samples\n")
	sampleCount := 0
	for _, entry := range errorLogs {
		sb.WriteString(fmt.Sprintf("[%s] [%s] [%s] %s\n",
			entry.Timestamp.Format(time.RFC3339),
			entry.PodName,
			entry.LogLevel,
			entry.Content))
		sampleCount++
		if sampleCount >= 20 {
			break
		}
	}
	sb.WriteString("\n")

	// Add request for specific analysis
	sb.WriteString("## Analysis Request\n")
	sb.WriteString("Based on the error logs provided, please analyze the following:\n")
	sb.WriteString("1. Provide a brief summary of the errors observed\n")
	sb.WriteString("2. Identify the most likely root causes of the errors\n")
	sb.WriteString("3. Suggest specific solutions to address the problems\n")
	sb.WriteString("4. Add any additional information or context that might be helpful\n")
	sb.WriteString("5. Assess the severity (Low, Medium, High, Critical)\n\n")

	sb.WriteString("Format your response as JSON with the following structure:\n")
	sb.WriteString("```json\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"summary\": \"Brief description of the errors\",\n")
	sb.WriteString("  \"rootCauses\": [\"Cause 1\", \"Cause 2\", ...],\n")
	sb.WriteString("  \"solutions\": [\"Solution 1\", \"Solution 2\", ...],\n")
	sb.WriteString("  \"additionalInfo\": [\"Info 1\", \"Info 2\", ...],\n")
	sb.WriteString("  \"severity\": \"Low|Medium|High|Critical\"\n")
	sb.WriteString("}\n")
	sb.WriteString("```\n")

	return sb.String()
}
