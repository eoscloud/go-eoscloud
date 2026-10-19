package model

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

// VM models (tag: vms)

type ListVM struct {
	Self                 string         `json:"_self,omitempty" yaml:"_self,omitempty"`
	ID                   string         `json:"id,omitempty" yaml:"id,omitempty"`
	Block                string         `json:"block_id" yaml:"block_id"`
	Product              string         `json:"product_id" yaml:"product_id"`
	Network              string         `json:"network_id" yaml:"network_id"`
	Image                string         `json:"image_id" yaml:"image_id"`
	Name                 string         `json:"name" yaml:"name"`
	Description          string         `json:"description,omitempty" yaml:"description,omitempty"`
	DiskSize             int            `json:"disk_size" yaml:"disk_size"`
	IPv4                 string         `json:"ipv4,omitempty" yaml:"ipv4,omitempty"`
	IPv6                 string         `json:"ipv6,omitempty" yaml:"ipv6,omitempty"`
	Status               StatusEnum     `json:"status,omitempty" yaml:"status,omitempty"`
	PowerState           PowerStateEnum `json:"power_state,omitempty" yaml:"power_state,omitempty"`
	CpuUtilization5m     int            `json:"cpu_utilization_5m" yaml:"cpu_utilization_5m"`
	MemoryUtilization5m  int            `json:"memory_utilization_5m" yaml:"memory_utilization_5m"`
	UtilizationUpdatedAt time.Time      `json:"utilization_updated_at,omitempty" yaml:"utilization_updated_at,omitempty"`
	CreatedAt            time.Time      `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt            time.Time      `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
	StartedAt            time.Time      `json:"started_at,omitempty" yaml:"started_at,omitempty"`
	Actions              []string       `json:"_actions,omitempty" yaml:"_actions,omitempty"`
}

type VMDetail struct {
	Self                 string          `json:"_self,omitempty" yaml:"_self,omitempty"`
	ID                   string          `json:"id,omitempty" yaml:"id,omitempty"`
	Block                string          `json:"block_id" yaml:"block_id"`
	Product              string          `json:"product_id" yaml:"product_id"`
	Network              string          `json:"network_id" yaml:"network_id"`
	Image                string          `json:"image_id" yaml:"image_id"`
	Name                 string          `json:"name" yaml:"name"`
	Description          string          `json:"description,omitempty" yaml:"description,omitempty"`
	DiskSize             int             `json:"disk_size" yaml:"disk_size"`
	IPv4                 string          `json:"ipv4,omitempty" yaml:"ipv4,omitempty"`
	IPv6                 string          `json:"ipv6,omitempty" yaml:"ipv6,omitempty"`
	Status               StatusEnum      `json:"status,omitempty" yaml:"status,omitempty"`
	PowerState           PowerStateEnum  `json:"power_state,omitempty" yaml:"power_state,omitempty"`
	AgentReady           bool            `json:"agent_ready,omitempty" yaml:"agent_ready,omitempty"`
	CpuUtilization5m     int             `json:"cpu_utilization_5m" yaml:"cpu_utilization_5m"`
	MemoryUtilization5m  int             `json:"memory_utilization_5m" yaml:"memory_utilization_5m"`
	Utilization          json.RawMessage `json:"utilization,omitempty" yaml:"utilization,omitempty"`
	UtilizationUpdatedAt time.Time       `json:"utilization_updated_at,omitempty" yaml:"utilization_updated_at,omitempty"`
	CreatedAt            time.Time       `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt            time.Time       `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
	StartedAt            time.Time       `json:"started_at,omitempty" yaml:"started_at,omitempty"`
	Actions              []string        `json:"_actions,omitempty" yaml:"_actions,omitempty"`
}

type VMRequest struct {
	Block       string `json:"block_id" yaml:"block_id" shortflag:"B"`
	Product     string `json:"product_id" yaml:"product_id" shortflag:"P"`
	Network     string `json:"network_id" yaml:"network_id" shortflag:"N"`
	Image       string `json:"image_id,omitempty" yaml:"image_id,omitempty" shortflag:"I"`
	Template    string `json:"template_id,omitempty" yaml:"template_id,omitempty" shortflag:"T"`
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty" shortflag:"D"`
	DiskSize    int    `json:"disk_size" yaml:"disk_size" shortflag:"S"`
	UserData    string `json:"user_data,omitempty" yaml:"user_data,omitempty" shortflag:"U"`
}

// MarshalJSON base64-encodes UserData before the request is sent. The
// field is populated with the raw cloud-init payload; the API expects it as a
// base64-encoded string.
func (r VMRequest) MarshalJSON() ([]byte, error) {
	type alias VMRequest
	out := alias(r)
	if out.UserData != "" {
		out.UserData = base64.StdEncoding.EncodeToString([]byte(out.UserData))
	}
	return json.Marshal(out)
}

type PatchedVMRequest struct {
	Description string `json:"description,omitempty" yaml:"description,omitempty" shortflag:"D"`
}

type VMPowerActionRequest struct {
	Timeout int `json:"timeout,omitempty" yaml:"timeout,omitempty" shortflag:"T"`
}

// VMConsole models and requests (from vms tag)

type VMConsole struct {
	ID         string    `json:"id,omitempty" yaml:"id,omitempty"`
	ConsoleURL string    `json:"console_url,omitempty" yaml:"console_url,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty" yaml:"created_at,omitempty"`
}

// VMConsoleRequestRequest represents the request payload to obtain a VM console URL.

type VMConsoleRequestRequest struct {
	Force bool `json:"force,omitempty" yaml:"force,omitempty" shortflag:"F"`
}

// VMConsoleError represents error response for VM console request

type VMConsoleError struct {
	Detail string `json:"detail" yaml:"detail"`
}

// VMUpscaleRequest represents the payload for VM upscale

type VMUpscaleRequest struct {
	ProductID string `json:"product_id" yaml:"product_id" shortflag:"P"`
}

// VMUpscaleError represents error response for VM upscale

type VMUpscaleError struct {
	Detail string `json:"detail" yaml:"detail"`
}

// VMChangeNetworkRequest represents the payload for VM upscale

type VMChangeNetworkRequest struct {
	NetworkID string `json:"network_id" yaml:"network_id" shortflag:"N"`
}

// VMChangeNetworkError represents error response for VM upscale

type VMChangeNetworkError struct {
	Detail string `json:"detail" yaml:"detail"`
}

// VMExec models and requests (from vms tag)

type VMExec struct {
	ID         string    `json:"id,omitempty" yaml:"id,omitempty"`
	ConsoleURL string    `json:"console_url,omitempty" yaml:"console_url,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty" yaml:"created_at,omitempty"`
}

// VMExecRequest represents the request payload to execute a command on a VM.
// An existing session can be forcefully replaced. Supports interactive mode.

type VMExecRequest struct {
	Force       bool              `json:"force,omitempty" yaml:"force,omitempty" shortflag:"F"`
	Interactive bool              `json:"interactive,omitempty" yaml:"interactive,omitempty" shortflag:"I"`
	Command     string            `json:"command,omitempty" yaml:"command,omitempty" shortflag:"C"`
	Environment map[string]string `json:"environment,omitempty" yaml:"environment,omitempty" shortflag:"E"`
}

// VMExecError represents error response for VM exec request

type VMExecError struct {
	Detail string `json:"detail" yaml:"detail"`
}
