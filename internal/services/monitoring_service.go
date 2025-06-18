package services

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nguyendkn/nginx-manager/pkg/logger"
)

// MonitoringService handles system monitoring and real-time metrics
type MonitoringService struct {
	startTime    time.Time
	connections  map[string]*websocket.Conn
	upgrader     websocket.Upgrader
	nginxService *NginxService
}

// SystemMetrics represents comprehensive system metrics
type SystemMetrics struct {
	Timestamp time.Time `json:"timestamp"`
	CPU       CPUStats  `json:"cpu"`
	Memory    MemStats  `json:"memory"`
	Disk      DiskStats `json:"disk"`
	Network   NetStats  `json:"network"`
	Process   ProcStats `json:"process"`
}

// CPUStats represents CPU usage statistics
type CPUStats struct {
	Usage     float64 `json:"usage"`
	LoadAvg1  float64 `json:"load_avg_1"`
	LoadAvg5  float64 `json:"load_avg_5"`
	LoadAvg15 float64 `json:"load_avg_15"`
}

// MemStats represents memory usage statistics
type MemStats struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
	GoAlloc     uint64  `json:"go_alloc"`
	GoTotal     uint64  `json:"go_total"`
	GoSys       uint64  `json:"go_sys"`
}

// DiskStats represents disk usage statistics
type DiskStats struct {
	Total       uint64  `json:"total"`
	Free        uint64  `json:"free"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
}

// NetStats represents network statistics
type NetStats struct {
	BytesRecv   uint64 `json:"bytes_recv"`
	BytesSent   uint64 `json:"bytes_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
	PacketsSent uint64 `json:"packets_sent"`
}

// ProcStats represents process statistics
type ProcStats struct {
	Goroutines int           `json:"goroutines"`
	GCRuns     uint32        `json:"gc_runs"`
	Uptime     time.Duration `json:"uptime"`
	GoVersion  string        `json:"go_version"`
	PID        int           `json:"pid"`
}

// NginxStatus represents nginx service status
type NginxStatus struct {
	Running     bool      `json:"running"`
	PID         int       `json:"pid"`
	Version     string    `json:"version"`
	ConfigTest  bool      `json:"config_test"`
	LastReload  time.Time `json:"last_reload"`
	Connections int       `json:"connections"`
}

// ActivityEvent represents a system activity event
type ActivityEvent struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Level     string    `json:"level"`
	Details   gin.H     `json:"details"`
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService(nginxService *NginxService) *MonitoringService {
	return &MonitoringService{
		startTime:    time.Now(),
		connections:  make(map[string]*websocket.Conn),
		nginxService: nginxService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow connections from any origin in development
				// In production, this should be more restrictive
				return true
			},
		},
	}
}

// GetSystemMetrics collects comprehensive system metrics
func (s *MonitoringService) GetSystemMetrics() (*SystemMetrics, error) {
	metrics := &SystemMetrics{
		Timestamp: time.Now(),
	}

	// Collect CPU stats
	cpuStats, err := s.getCPUStats()
	if err != nil {
		logger.Warn("Failed to get CPU stats", logger.Err(err))
	} else {
		metrics.CPU = cpuStats
	}

	// Collect memory stats
	memStats, err := s.getMemoryStats()
	if err != nil {
		logger.Warn("Failed to get memory stats", logger.Err(err))
	} else {
		metrics.Memory = memStats
	}

	// Collect disk stats
	diskStats, err := s.getDiskStats()
	if err != nil {
		logger.Warn("Failed to get disk stats", logger.Err(err))
	} else {
		metrics.Disk = diskStats
	}

	// Collect network stats
	netStats, err := s.getNetworkStats()
	if err != nil {
		logger.Warn("Failed to get network stats", logger.Err(err))
	} else {
		metrics.Network = netStats
	}

	// Collect process stats
	metrics.Process = s.getProcessStats()

	return metrics, nil
}

// getCPUStats gets CPU usage statistics
func (s *MonitoringService) getCPUStats() (CPUStats, error) {
	stats := CPUStats{}

	// For Windows and cross-platform compatibility, we'll use simulated data
	// In a production environment, you would use platform-specific APIs
	stats.Usage = 15.5 + float64(time.Now().Unix()%20)
	stats.LoadAvg1 = 1.2
	stats.LoadAvg5 = 1.5
	stats.LoadAvg15 = 1.8

	// For Linux/Unix systems, try to read from /proc/loadavg
	if runtime.GOOS != "windows" {
		if data, err := os.ReadFile("/proc/loadavg"); err == nil {
			fields := strings.Fields(string(data))
			if len(fields) >= 3 {
				if val, err := strconv.ParseFloat(fields[0], 64); err == nil {
					stats.LoadAvg1 = val
				}
				if val, err := strconv.ParseFloat(fields[1], 64); err == nil {
					stats.LoadAvg5 = val
				}
				if val, err := strconv.ParseFloat(fields[2], 64); err == nil {
					stats.LoadAvg15 = val
				}
			}
		}
	}

	return stats, nil
}

// getMemoryStats gets memory usage statistics
func (s *MonitoringService) getMemoryStats() (MemStats, error) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	stats := MemStats{
		GoAlloc: memStats.Alloc,
		GoTotal: memStats.TotalAlloc,
		GoSys:   memStats.Sys,
	}

	// Cross-platform memory stats - simplified for demo
	if runtime.GOOS == "windows" {
		// Simplified memory stats for Windows
		stats.Total = 8 * 1024 * 1024 * 1024 // 8GB
		stats.Used = stats.Total / 3         // ~33% usage
		stats.Available = stats.Total - stats.Used
		stats.UsedPercent = float64(stats.Used) / float64(stats.Total) * 100
		return stats, nil
	}

	// For Linux, read from /proc/meminfo
	if data, err := os.ReadFile("/proc/meminfo"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				value, _ := strconv.ParseUint(fields[1], 10, 64)
				value *= 1024 // Convert from KB to bytes

				switch fields[0] {
				case "MemTotal:":
					stats.Total = value
				case "MemAvailable:":
					stats.Available = value
				}
			}
		}
		stats.Used = stats.Total - stats.Available
		if stats.Total > 0 {
			stats.UsedPercent = float64(stats.Used) / float64(stats.Total) * 100
		}
	}

	return stats, nil
}

// getDiskStats gets disk usage statistics
func (s *MonitoringService) getDiskStats() (DiskStats, error) {
	stats := DiskStats{}

	// Cross-platform disk stats - simplified for demo
	// In production, use platform-specific APIs
	stats.Total = 500 * 1024 * 1024 * 1024 // 500GB
	stats.Used = stats.Total / 2           // 50% usage
	stats.Free = stats.Total - stats.Used
	stats.UsedPercent = 50.0

	return stats, nil
}

// getNetworkStats gets network usage statistics
func (s *MonitoringService) getNetworkStats() (NetStats, error) {
	stats := NetStats{}

	// Cross-platform network stats - simplified for demo
	// In production, use platform-specific APIs
	stats.BytesRecv = 1024 * 1024 * 100 // 100MB
	stats.BytesSent = 1024 * 1024 * 50  // 50MB
	stats.PacketsRecv = 10000
	stats.PacketsSent = 8000

	// For Linux, try to read from /proc/net/dev
	if runtime.GOOS != "windows" {
		if data, err := os.ReadFile("/proc/net/dev"); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.Contains(line, ":") {
					fields := strings.Fields(line)
					if len(fields) >= 10 {
						if recv, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
							stats.BytesRecv += recv
						}
						if sent, err := strconv.ParseUint(fields[9], 10, 64); err == nil {
							stats.BytesSent += sent
						}
					}
				}
			}
		}
	}

	return stats, nil
}

// getProcessStats gets process-specific statistics
func (s *MonitoringService) getProcessStats() ProcStats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return ProcStats{
		Goroutines: runtime.NumGoroutine(),
		GCRuns:     memStats.NumGC,
		Uptime:     time.Since(s.startTime),
		GoVersion:  runtime.Version(),
		PID:        os.Getpid(),
	}
}

// GetNginxStatus gets nginx service status
func (s *MonitoringService) GetNginxStatus() (*NginxStatus, error) {
	status := &NginxStatus{
		Running:    false,
		ConfigTest: false,
		LastReload: time.Now(),
	}

	// Check if nginx is running and get basic status
	if running := s.isNginxRunning(); running {
		status.Running = true
	}

	// Test nginx configuration
	if configValid := s.testNginxConfig(); configValid {
		status.ConfigTest = true
	}

	// Get nginx version
	if version, err := s.getNginxVersion(); err == nil {
		status.Version = version
	}

	// Get nginx PID
	if pid, err := s.getNginxPID(); err == nil {
		status.PID = pid
	}

	return status, nil
}

// isNginxRunning checks if nginx is running
func (s *MonitoringService) isNginxRunning() bool {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("tasklist", "/fi", "imagename eq nginx.exe")
		output, err := cmd.Output()
		if err != nil {
			return false
		}
		return strings.Contains(string(output), "nginx.exe")
	}

	// For Linux/Unix
	cmd := exec.Command("pgrep", "nginx")
	err := cmd.Run()
	return err == nil
}

// testNginxConfig tests nginx configuration
func (s *MonitoringService) testNginxConfig() bool {
	cmd := exec.Command("nginx", "-t")
	err := cmd.Run()
	return err == nil
}

// getNginxVersion gets nginx version
func (s *MonitoringService) getNginxVersion() (string, error) {
	cmd := exec.Command("nginx", "-v")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	// Parse version from output like "nginx version: nginx/1.18.0"
	version := strings.TrimSpace(string(output))
	if strings.Contains(version, "nginx/") {
		parts := strings.Split(version, "nginx/")
		if len(parts) > 1 {
			return parts[1], nil
		}
	}

	return version, nil
}

// getNginxPID gets nginx master process PID
func (s *MonitoringService) getNginxPID() (int, error) {
	// For Windows
	if runtime.GOOS == "windows" {
		cmd := exec.Command("tasklist", "/fi", "imagename eq nginx.exe", "/fo", "csv")
		output, err := cmd.Output()
		if err != nil {
			return 0, err
		}

		lines := strings.Split(string(output), "\n")
		if len(lines) > 1 {
			fields := strings.Split(lines[1], ",")
			if len(fields) > 1 {
				pidStr := strings.Trim(fields[1], `"`)
				return strconv.Atoi(pidStr)
			}
		}
		return 0, fmt.Errorf("nginx not found")
	}

	// For Linux/Unix
	cmd := exec.Command("pgrep", "-f", "nginx: master")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	pidStr := strings.TrimSpace(string(output))
	return strconv.Atoi(pidStr)
}

// HandleWebSocket handles WebSocket connections for real-time updates
func (s *MonitoringService) HandleWebSocket(c *gin.Context) {
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("Failed to upgrade to WebSocket", logger.Err(err))
		return
	}
	defer conn.Close()

	clientID := c.Query("client_id")
	if clientID == "" {
		clientID = fmt.Sprintf("client_%d", time.Now().UnixNano())
	}

	s.connections[clientID] = conn
	defer delete(s.connections, clientID)

	logger.Info("WebSocket client connected", logger.String("client_id", clientID))

	// Send initial metrics
	if metrics, err := s.GetSystemMetrics(); err == nil {
		s.sendToClient(conn, "metrics", metrics)
	}

	if nginxStatus, err := s.GetNginxStatus(); err == nil {
		s.sendToClient(conn, "nginx_status", nginxStatus)
	}

	// Keep connection alive and handle incoming messages
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			logger.Info("WebSocket client disconnected", logger.String("client_id", clientID))
			break
		}
	}
}

// sendToClient sends data to a specific WebSocket client
func (s *MonitoringService) sendToClient(conn *websocket.Conn, eventType string, data interface{}) {
	message := gin.H{
		"type":      eventType,
		"timestamp": time.Now(),
		"data":      data,
	}

	if err := conn.WriteJSON(message); err != nil {
		logger.Error("Failed to send WebSocket message", logger.Err(err))
	}
}

// BroadcastMetrics broadcasts system metrics to all connected clients
func (s *MonitoringService) BroadcastMetrics() {
	metrics, err := s.GetSystemMetrics()
	if err != nil {
		logger.Error("Failed to get system metrics", logger.Err(err))
		return
	}

	nginxStatus, err := s.GetNginxStatus()
	if err != nil {
		logger.Error("Failed to get nginx status", logger.Err(err))
		return
	}

	for clientID, conn := range s.connections {
		s.sendToClient(conn, "metrics", metrics)
		s.sendToClient(conn, "nginx_status", nginxStatus)

		// Remove disconnected clients
		if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			logger.Info("Removing disconnected client", logger.String("client_id", clientID))
			delete(s.connections, clientID)
		}
	}
}

// GetRecentActivity gets recent system activity events
func (s *MonitoringService) GetRecentActivity(limit int) ([]ActivityEvent, error) {
	// In a real implementation, this would read from a database or log file
	// For demo purposes, we'll generate some sample events
	events := []ActivityEvent{
		{
			ID:        "evt_001",
			Timestamp: time.Now().Add(-5 * time.Minute),
			Type:      "certificate",
			Message:   "Certificate renewed successfully",
			Level:     "info",
			Details:   gin.H{"domain": "example.com", "provider": "letsencrypt"},
		},
		{
			ID:        "evt_002",
			Timestamp: time.Now().Add(-10 * time.Minute),
			Type:      "proxy_host",
			Message:   "Proxy host created",
			Level:     "info",
			Details:   gin.H{"domain": "api.example.com", "target": "localhost:3000"},
		},
		{
			ID:        "evt_003",
			Timestamp: time.Now().Add(-15 * time.Minute),
			Type:      "nginx",
			Message:   "Nginx configuration reloaded",
			Level:     "info",
			Details:   gin.H{"config_test": true},
		},
	}

	if limit > 0 && limit < len(events) {
		events = events[:limit]
	}

	return events, nil
}

// StartMetricsBroadcast starts periodic metrics broadcasting
func (s *MonitoringService) StartMetricsBroadcast(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			s.BroadcastMetrics()
		}
	}()

	logger.Info("Started metrics broadcasting", logger.Duration("interval", interval))
}
