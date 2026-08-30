package model

type AccountInfo struct {
	TenantId      string  `json:"tenantId"`
	AccountId     string  `json:"accountId"`
	UserId        string  `json:"userId"`
	PermissionIds []int64 `json:"permissionIds"`
}
