package client

import (
	"net/http"
	"net/url"
)

type QuayPermission string
type QuayTeamRole string

const (
	QuayPermissionAdmin QuayPermission = "admin"
	QuayPermissionRead  QuayPermission = "read"
	QuayPermissionWrite QuayPermission = "write"
	QuayTeamRoleAdmin   QuayTeamRole   = "admin"
	QuayTeamRoleCreator QuayTeamRole   = "creator"
	QuayTeamRoleMember  QuayTeamRole   = "member"
)

type QuayClient struct {
	baseURL    *url.URL
	httpClient *http.Client
	authToken  string
}

type PrototypesResponse struct {
	Prototypes []Prototype `json:"prototypes"`
}

type RobotAccount struct {
	Description  string `json:"description"`
	Created      string `json:"created"`
	LastAccessed string `json:"last_accessed"`
	Token        string `json:"token"`
	Name         string `json:"name"`
}

type Prototype struct {
	ID       string            `json:"id"`
	Role     string            `json:"role"`
	Delegate PrototypeDelegate `json:"delegate"`
}

type PrototypeDelegate struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Robot     bool   `json:"is_robot"`
	OrgMember bool   `json:"is_org_member"`
}

type Team struct {
	Name string       `json:"name"`
	Role QuayTeamRole `json:"role"`
}

// StringValue represents an object containing a single string
type StringValue struct {
	Value string
}

type QuayApiError struct {
	Error error
}
