package model

// Shared enums and types used across multiple models

type StatusEnum string

const (
	StatusCreating   StatusEnum = "creating"
	StatusReady      StatusEnum = "ready"
	StatusDestroying StatusEnum = "destroying"
	StatusUpdating   StatusEnum = "updating"
)

type UnitEnum string

const (
	UnitGigabyteHours UnitEnum = "gigabytehours"
)
