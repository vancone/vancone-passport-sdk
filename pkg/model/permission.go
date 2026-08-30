package model

type Permission struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	ServiceId   string   `json:"serviceId"`
	TenantId    string   `json:"tenantId"`
	ApiIds      []string `json:"apiIds"`
}
