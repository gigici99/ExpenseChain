package model

import "time"

// Role defines what a user is allowed to do.
type Role string

const (
	RoleAdmin    Role = "ADMIN"    // can verify/read everything
	RoleCompany  Role = "COMPANY"  // can manage policies and view the ledger
	RoleEmployee Role = "EMPLOYEE" // can submit expenses
)

// User is an authentication principal.
// CompanyID scopes COMPANY/EMPLOYEE users to their company.
// EmployeeID links an EMPLOYEE user to its Employee record.
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username" gorm:"uniqueIndex"`
	PasswordHash string    `json:"-"` // never serialized to JSON
	Role         Role      `json:"role"`
	CompanyID    string    `json:"company_id,omitempty"`
	EmployeeID   string    `json:"employee_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
