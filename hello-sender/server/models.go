package server

import (
	"time"
)

type Command struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	CmdName string       `json:"cmdName" binding:"required"`
	Args    string       `json:"args" binding:"required"`
	Runs    []CommandRun `json:"runs"`
}

type CommandRun struct {
	ID          uint `json:"id"`
	CommandID   uint `gorm:"not null" sql:"type:uint REFERENCES commands(id)" json:"commandId" binding:"required"`
	StartedAt   *time.Time
	CompletedAt *time.Time
}
