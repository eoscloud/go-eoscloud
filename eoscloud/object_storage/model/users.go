package model

import "time"

// User models (tag: user)

type User struct {
	Self        string     `json:"_self,omitempty" yaml:"_self,omitempty"`
	Actions     []string   `json:"_actions,omitempty" yaml:"_actions,omitempty"`
	ID          string     `json:"id,omitempty" yaml:"id,omitempty"`
	Name        string     `json:"name,omitempty" yaml:"name,omitempty"`
	Description *string    `json:"description,omitempty" yaml:"description,omitempty"`
	DistrictID  string     `json:"district_id,omitempty" yaml:"district_id,omitempty"`
	AccountID   string     `json:"account_id,omitempty" yaml:"account_id,omitempty"`
	IAMHandle   string     `json:"iam_handle,omitempty" yaml:"iam_handle,omitempty"`
	AccessKey   string     `json:"access_key,omitempty" yaml:"access_key,omitempty"`
	SecretKey   string     `json:"secret_key,omitempty" yaml:"secret_key,omitempty"`
	Root        bool       `json:"root,omitempty" yaml:"root,omitempty"`
	Status      StatusEnum `json:"status,omitempty" yaml:"status,omitempty"`
	CreatedAt   time.Time  `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
}

type UserRequest struct {
	Name        string  `json:"name" yaml:"name"`
	Description *string `json:"description,omitempty" yaml:"description,omitempty" shortflag:"D"`
	DistrictID  string  `json:"district_id" yaml:"district_id"`
}

type PatchedUserRequest struct {
	Name        string  `json:"name,omitempty" yaml:"name,omitempty"`
	Description *string `json:"description,omitempty" yaml:"description,omitempty" shortflag:"D"`
	DistrictID  string  `json:"district_id,omitempty" yaml:"district_id,omitempty"`
}
