package logs

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// LogOptions defines options for collecting logs
type LogOptions struct {
	// Resource type (pod, deployment, statefulset, etc.)
	ResourceType string
	// Resource name
	ResourceName string
	// Namespace
	Namespace string
	// Container name (optional)
	Container string
	// Previous terminated container logs
	Previous bool
	// Number of lines to return
	TailLines *int64
	// If true, logs are streamed as they become available
	Follow bool
	// Start time for logs
	SinceTime *metav1.Time
	// Duration from now to start returning logs
	SinceSeconds *int64
	// Time to wait if Follow=true
	Timeout time.Duration
}

// LogEntry represents a structured log entry
type LogEntry struct {
	// Timestamp of the log entry
	Timestamp time.Time
	// Source pod name
	PodName string
	// Source container name
	ContainerName string
	// Log level (info, warn, error, etc.)
	LogLevel string
	// Raw log content
	Content string
	// Structured data extracted from the log
	Data map[string]string
}

// LogCollector handles collecting logs from Kubernetes resources
type LogCollector struct {
	clientset kubernetes.Interface
}

// NewLogCollector creates a new log collector
func NewLogCollector(clientset kubernetes.Interface) *LogCollector {
	return &LogCollector{
		clientset: clientset,
	}
}

// GetPodLogs retrieves logs directly from a pod
func (c *LogCollector) GetPodLogs(ctx context.Context, options LogOptions) ([]LogEntry, error) {
	podLogOpts := &corev1.PodLogOptions{
		Container:    options.Container,
		Follow:       options.Follow,
		Previous:     options.Previous,
		SinceSeconds: options.SinceSeconds,
		SinceTime:    options.SinceTime,
		TailLines:    options.TailLines,
	}

	req := c.clientset.CoreV1().Pods(options.Namespace).GetLogs(options.ResourceName, podLogOpts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("error opening log stream for pod %s: %w", options.ResourceName, err)
	}
	defer podLogs.Close()

	var logEntries []LogEntry
	reader := bufio.NewReader(podLogs)

	for {
		select {
		case <-ctx.Done():
			return logEntries, nil
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					// Add the last line if it's not empty
					if line != "" {
						entry := parseLogLine(line, options.ResourceName, options.Container)
						logEntries = append(logEntries, entry)
					}
					return logEntries, nil
				}
				return logEntries, fmt.Errorf("error reading log stream: %w", err)
			}

			// Parse and add the log entry
			entry := parseLogLine(line, options.ResourceName, options.Container)
			logEntries = append(logEntries, entry)

			// For non-follow logs, we limit the number of entries to prevent memory issues
			if !options.Follow && len(logEntries) > 10000 {
				return logEntries, fmt.Errorf("log output too large, please use filters to reduce the log volume")
			}
		}
	}
}

// GetResourceLogs retrieves logs from a Kubernetes resource
// This handles different resource types (e.g., deployments, statefulsets)
func (c *LogCollector) GetResourceLogs(ctx context.Context, options LogOptions) ([]LogEntry, error) {
	switch options.ResourceType {
	case "pod":
		return c.GetPodLogs(ctx, options)
	case "deployment", "deploy":
		return c.getDeploymentLogs(ctx, options)
	case "statefulset", "sts":
		return c.getStatefulSetLogs(ctx, options)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", options.ResourceType)
	}
}

// getDeploymentLogs retrieves logs from all pods in a deployment
func (c *LogCollector) getDeploymentLogs(ctx context.Context, options LogOptions) ([]LogEntry, error) {
	// Get the deployment to find its selector
	deployment, err := c.clientset.AppsV1().Deployments(options.Namespace).Get(ctx, options.ResourceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting deployment %s: %w", options.ResourceName, err)
	}

	// Get the pods matching the deployment selector
	selector := metav1.FormatLabelSelector(deployment.Spec.Selector)
	pods, err := c.clientset.CoreV1().Pods(options.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, fmt.Errorf("error listing pods for deployment %s: %w", options.ResourceName, err)
	}

	return c.getLogsFromPods(ctx, pods.Items, options)
}

// getStatefulSetLogs retrieves logs from all pods in a statefulset
func (c *LogCollector) getStatefulSetLogs(ctx context.Context, options LogOptions) ([]LogEntry, error) {
	// Get the statefulset to find its selector
	statefulset, err := c.clientset.AppsV1().StatefulSets(options.Namespace).Get(ctx, options.ResourceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting statefulset %s: %w", options.ResourceName, err)
	}

	// Get the pods matching the statefulset selector
	selector := metav1.FormatLabelSelector(statefulset.Spec.Selector)
	pods, err := c.clientset.CoreV1().Pods(options.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, fmt.Errorf("error listing pods for statefulset %s: %w", options.ResourceName, err)
	}

	return c.getLogsFromPods(ctx, pods.Items, options)
}

// getLogsFromPods retrieves logs from a list of pods
func (c *LogCollector) getLogsFromPods(ctx context.Context, pods []corev1.Pod, options LogOptions) ([]LogEntry, error) {
	var allLogs []LogEntry

	// Collect logs from each pod
	for _, pod := range pods {
		podOpts := options
		podOpts.ResourceType = "pod"
		podOpts.ResourceName = pod.Name

		logs, err := c.GetPodLogs(ctx, podOpts)
		if err != nil {
			// Continue to next pod if we can't get logs from this one
			fmt.Printf("Warning: error getting logs from pod %s: %v\n", pod.Name, err)
			continue
		}

		allLogs = append(allLogs, logs...)
	}

	if len(allLogs) == 0 {
		return nil, fmt.Errorf("no logs found for %s %s", options.ResourceType, options.ResourceName)
	}

	return allLogs, nil
}

// parseLogLine parses a log line into a structured LogEntry
func parseLogLine(line string, podName, containerName string) LogEntry {
	line = strings.TrimSuffix(line, "\n")

	entry := LogEntry{
		Timestamp:     time.Now(), // Default to current time if we can't parse
		PodName:       podName,
		ContainerName: containerName,
		Content:       line,
		Data:          make(map[string]string),
	}

	// Try to extract timestamp
	if timestampEnd := strings.IndexByte(line, ' '); timestampEnd > 0 {
		if t, err := time.Parse(time.RFC3339, line[:timestampEnd]); err == nil {
			entry.Timestamp = t
			line = line[timestampEnd+1:]
		}
	}

	// Try to extract log level
	for _, level := range []string{"DEBUG", "INFO", "WARN", "WARNING", "ERROR", "FATAL"} {
		if strings.Contains(line, level) {
			entry.LogLevel = level
			break
		}
	}

	// If no explicit level is found, try to infer from content
	if entry.LogLevel == "" {
		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, "error") || strings.Contains(lowerLine, "exception") || strings.Contains(lowerLine, "fail") {
			entry.LogLevel = "ERROR"
		} else if strings.Contains(lowerLine, "warn") {
			entry.LogLevel = "WARN"
		} else {
			entry.LogLevel = "INFO"
		}
	}

	// Extract key-value pairs if present
	extractStructuredData(&entry)

	return entry
}

// extractStructuredData attempts to extract structured data from log content
func extractStructuredData(entry *LogEntry) {
	content := entry.Content

	// Look for JSON-like patterns {key=value} or key=value
	// This is a simple extraction - a more robust solution might use regex

	// Extract key-value pairs with format key=value
	parts := strings.Split(content, " ")
	for _, part := range parts {
		if strings.Contains(part, "=") {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) == 2 {
				key := strings.Trim(kv[0], `"'`)
				value := strings.Trim(kv[1], `"'`)
				entry.Data[key] = value
			}
		}
	}
}

// StreamLogs streams logs in real-time, sending each log entry to the provided channel
func (c *LogCollector) StreamLogs(ctx context.Context, options LogOptions, logChan chan<- LogEntry, errChan chan<- error) error {
	// Make sure we're streaming
	options.Follow = true

	// For non-pod resources, we need to determine the actual pods
	switch options.ResourceType {
	case "pod":
		return c.streamPodLogs(ctx, options, logChan, errChan)
	case "deployment":
		return c.streamDeploymentLogs(ctx, options, logChan, errChan)
	case "statefulset":
		return c.streamStatefulSetLogs(ctx, options, logChan, errChan)
	default:
		return fmt.Errorf("unsupported resource type for log streaming: %s", options.ResourceType)
	}
}

// streamPodLogs streams logs from a specific pod
func (c *LogCollector) streamPodLogs(ctx context.Context, options LogOptions, logChan chan<- LogEntry, errChan chan<- error) error {
	defer close(logChan)
	defer close(errChan)

	podLogOpts := &corev1.PodLogOptions{
		Container:    options.Container,
		Follow:       true,
		Previous:     options.Previous,
		SinceSeconds: options.SinceSeconds,
		SinceTime:    options.SinceTime,
		TailLines:    options.TailLines,
	}

	req := c.clientset.CoreV1().Pods(options.Namespace).GetLogs(options.ResourceName, podLogOpts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return fmt.Errorf("error opening log stream for pod %s: %w", options.ResourceName, err)
	}
	defer podLogs.Close()

	reader := bufio.NewReader(podLogs)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					// This shouldn't happen with Follow=true unless the pod terminated
					time.Sleep(1 * time.Second) // Wait a bit before retrying
					continue
				}
				if ctx.Err() != nil {
					// Context was canceled, just return
					return nil
				}
				errChan <- fmt.Errorf("error reading log stream: %w", err)
				return err
			}

			// Parse and send the log entry
			entry := parseLogLine(line, options.ResourceName, options.Container)
			select {
			case logChan <- entry:
				// Successfully sent the log entry
			case <-ctx.Done():
				return nil
			}
		}
	}
}

// streamDeploymentLogs streams logs from all pods in a deployment
func (c *LogCollector) streamDeploymentLogs(ctx context.Context, options LogOptions, logChan chan<- LogEntry, errChan chan<- error) error {
	// Get the deployment to find its selector
	deployment, err := c.clientset.AppsV1().Deployments(options.Namespace).Get(ctx, options.ResourceName, metav1.GetOptions{})
	if err != nil {
		close(logChan)
		close(errChan)
		return fmt.Errorf("error getting deployment %s: %w", options.ResourceName, err)
	}

	// Get the pods matching the deployment selector
	selector := metav1.FormatLabelSelector(deployment.Spec.Selector)
	return c.streamPodsWithSelector(ctx, options, selector, logChan, errChan)
}

// streamStatefulSetLogs streams logs from all pods in a stateful set
func (c *LogCollector) streamStatefulSetLogs(ctx context.Context, options LogOptions, logChan chan<- LogEntry, errChan chan<- error) error {
	// Get the statefulset to find its selector
	statefulset, err := c.clientset.AppsV1().StatefulSets(options.Namespace).Get(ctx, options.ResourceName, metav1.GetOptions{})
	if err != nil {
		close(logChan)
		close(errChan)
		return fmt.Errorf("error getting statefulset %s: %w", options.ResourceName, err)
	}

	// Get the pods matching the statefulset selector
	selector := metav1.FormatLabelSelector(statefulset.Spec.Selector)
	return c.streamPodsWithSelector(ctx, options, selector, logChan, errChan)
}

// streamPodsWithSelector streams logs from all pods matching a label selector
func (c *LogCollector) streamPodsWithSelector(ctx context.Context, options LogOptions, selector string, logChan chan<- LogEntry, errChan chan<- error) error {
	defer close(logChan)
	defer close(errChan)

	// List pods with the given selector
	pods, err := c.clientset.CoreV1().Pods(options.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return fmt.Errorf("error listing pods with selector %s: %w", selector, err)
	}

	if len(pods.Items) == 0 {
		return fmt.Errorf("no pods found matching selector %s", selector)
	}

	// Create a wait group to manage multiple goroutines for pod log streaming
	var wg sync.WaitGroup
	podLogChans := make([]chan LogEntry, len(pods.Items))
	podErrChans := make([]chan error, len(pods.Items))

	// Stream logs from each pod in a separate goroutine
	for i, pod := range pods.Items {
		podLogChans[i] = make(chan LogEntry)
		podErrChans[i] = make(chan error)

		wg.Add(1)
		go func(index int, podName string) {
			defer wg.Done()

			podOpts := options
			podOpts.ResourceType = "pod"
			podOpts.ResourceName = podName

			// This will close the pod's channels when done
			_ = c.streamPodLogs(ctx, podOpts, podLogChans[index], podErrChans[index])
		}(i, pod.Name)
	}

	// Merge all the pod log channels into the main channel
	go func() {
		for i := range pods.Items {
			go func(index int) {
				for entry := range podLogChans[index] {
					select {
					case logChan <- entry:
					case <-ctx.Done():
						return
					}
				}
			}(i)

			go func(index int) {
				for err := range podErrChans[index] {
					select {
					case errChan <- err:
					case <-ctx.Done():
						return
					}
				}
			}(i)
		}
	}()

	// Wait for all pod streaming to complete
	wg.Wait()

	return nil
}
