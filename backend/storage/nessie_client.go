package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"bronze-backend/config"
)

type NessieClient struct {
	client    *http.Client
	config    *config.NessieConfig
	baseURL   string
	namespace string
	authToken string
}

type NessieConfig struct {
	Endpoint  string `json:"endpoint"`
	Namespace string `json:"namespace"`
	AuthToken string `json:"auth_token"`
	DefaultDB string `json:"default_database"`
	BatchSize int    `json:"batch_size"`
}

type NessieTable struct {
	Name       string                 `json:"name"`
	Database   string                 `json:"database"`
	Columns    []NessieColumn         `json:"columns"`
	Properties map[string]interface{} `json:"properties"`
}

type NessieColumn struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
	Comment  string `json:"comment,omitempty"`
}

type NessieColumnMismatch struct {
	ColumnName   string `json:"column_name"`
	MismatchType string `json:"mismatch_type"` // "missing", "extra", "type_mismatch"
	SourceType   string `json:"source_type,omitempty"`
	TargetType   string `json:"target_type,omitempty"`
	Severity     string `json:"severity"` // "error", "warning", "info"
}

type CreateOperation string

const (
	CreateNewTable CreateOperation = "create"
	AppendTable    CreateOperation = "append"
)

func NewNessieClient(cfg *config.NessieConfig) (*NessieClient, error) {
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("Nessie endpoint is required")
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Remove trailing slash from endpoint
	endpoint := strings.TrimRight(cfg.Endpoint, "/")
	baseURL := fmt.Sprintf("%s/api/v1/namespaces/%s", endpoint, cfg.Namespace)

	nessieClient := &NessieClient{
		client:    client,
		config:    cfg,
		baseURL:   baseURL,
		namespace: cfg.Namespace,
		authToken: cfg.AuthToken,
	}

	// Test connection
	if err := nessieClient.testConnection(); err != nil {
		return nil, fmt.Errorf("failed to connect to Nessie: %w", err)
	}

	log.Printf("Nessie client initialized for namespace: %s", cfg.Namespace)
	return nessieClient, nil
}

func (n *NessieClient) testConnection() error {
	req, err := http.NewRequest("GET", n.baseURL+"/config", nil)
	if err != nil {
		return fmt.Errorf("failed to create test request: %w", err)
	}

	n.addAuthHeader(req)

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Nessie connection failed with status: %d", resp.StatusCode)
	}

	log.Printf("Successfully connected to Nessie")
	return nil
}

func (n *NessieClient) TableExists(ctx context.Context, database, tableName string) (bool, error) {
	tableURL := fmt.Sprintf("%s/databases/%s/tables/%s", n.baseURL, database, tableName)

	req, err := http.NewRequest("GET", tableURL, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create table exists request: %w", err)
	}

	n.addAuthHeader(req)
	req = req.WithContext(ctx)

	resp, err := n.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to check table existence: %w", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode < 400, nil
}

func (n *NessieClient) GetTableSchema(ctx context.Context, database, tableName string) (*NessieTable, error) {
	tableURL := fmt.Sprintf("%s/databases/%s/tables/%s", n.baseURL, database, tableName)

	req, err := http.NewRequest("GET", tableURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create table schema request: %w", err)
	}

	n.addAuthHeader(req)
	req = req.WithContext(ctx)

	resp, err := n.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get table schema: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		if resp.StatusCode == 404 {
			return nil, nil // Table doesn't exist
		}
		return nil, fmt.Errorf("failed to get table schema, status: %d", resp.StatusCode)
	}

	var table NessieTable
	if err := json.NewDecoder(resp.Body).Decode(&table); err != nil {
		return nil, fmt.Errorf("failed to decode table schema: %w", err)
	}

	return &table, nil
}

func (n *NessieClient) CreateTable(ctx context.Context, table *NessieTable) error {
	createURL := fmt.Sprintf("%s/databases/%s/tables", n.baseURL, table.Database)

	jsonData, err := json.Marshal(table)
	if err != nil {
		return fmt.Errorf("failed to marshal table schema: %w", err)
	}

	req, err := http.NewRequest("POST", createURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create table request: %w", err)
	}

	n.addAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed to create table, status: %d", resp.StatusCode)
	}

	log.Printf("Successfully created Nessie table: %s.%s", table.Database, table.Name)
	return nil
}

func (n *NessieClient) AppendToTable(ctx context.Context, database, tableName string, rows []map[string]interface{}) error {
	appendURL := fmt.Sprintf("%s/databases/%s/tables/%s/data", n.baseURL, database, tableName)

	requestData := map[string]interface{}{
		"rows": rows,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("failed to marshal append data: %w", err)
	}

	req, err := http.NewRequest("POST", appendURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create append request: %w", err)
	}

	n.addAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to append to table: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed to append to table, status: %d", resp.StatusCode)
	}

	log.Printf("Successfully appended %d rows to Nessie table: %s.%s", len(rows), database, tableName)
	return nil
}

func (n *NessieClient) ValidateSchema(sourceColumns []string, targetTable *NessieTable) []NessieColumnMismatch {
	var mismatches []NessieColumnMismatch

	// Create map of target columns for faster lookup
	targetCols := make(map[string]NessieColumn)
	for _, col := range targetTable.Columns {
		targetCols[strings.ToLower(col.Name)] = col
	}

	// Check each source column
	for _, sourceCol := range sourceColumns {
		sourceColLower := strings.ToLower(sourceCol)
		targetCol, exists := targetCols[sourceColLower]

		if !exists {
			// Source column not in target table
			mismatches = append(mismatches, NessieColumnMismatch{
				ColumnName:   sourceCol,
				MismatchType: "extra",
				SourceType:   "VARCHAR", // Assume string for source
				TargetType:   "",
				Severity:     "warning",
			})
		} else if !strings.EqualFold(sourceCol, targetCol.Name) {
			// Case difference
			mismatches = append(mismatches, NessieColumnMismatch{
				ColumnName:   sourceCol,
				MismatchType: "case_diff",
				SourceType:   "VARCHAR",
				TargetType:   targetCol.Type,
				Severity:     "info",
			})
		}
	}

	// Check for missing target columns
	sourceColMap := make(map[string]bool)
	for _, col := range sourceColumns {
		sourceColMap[strings.ToLower(col)] = true
	}

	for _, targetCol := range targetTable.Columns {
		if !sourceColMap[strings.ToLower(targetCol.Name)] {
			mismatches = append(mismatches, NessieColumnMismatch{
				ColumnName:   targetCol.Name,
				MismatchType: "missing",
				SourceType:   "",
				TargetType:   targetCol.Type,
				Severity:     "warning",
			})
		}
	}

	return mismatches
}

func (n *NessieClient) addAuthHeader(req *http.Request) {
	if n.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+n.authToken)
	}
}

func (n *NessieClient) InferNessieType(value interface{}) string {
	if value == nil {
		return "VARCHAR(255)"
	}

	switch v := value.(type) {
	case string:
		// Try to detect if it's a number or date
		if _, err := time.Parse(time.RFC3339, v); err == nil {
			return "TIMESTAMP"
		}
		if _, err := time.Parse("2006-01-02", v); err == nil {
			return "DATE"
		}
		if _, err := fmt.Sscanf(v, "%f", make([]interface{}, 1)...); err == nil {
			if strings.Contains(v, ".") {
				return "DECIMAL(20,8)"
			}
			return "BIGINT"
		}
		return "VARCHAR(255)"
	case int, int32, int64:
		return "BIGINT"
	case float32, float64:
		return "DECIMAL(20,8)"
	case bool:
		return "BOOLEAN"
	default:
		return "VARCHAR(255)"
	}
}
