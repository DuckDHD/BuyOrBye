package models

import (
	"time"

	"github.com/google/uuid"
)

type Referral struct {
	ID            uuid.UUID      `json:"id" db:"id"`
	InviterUserID uuid.UUID      `json:"inviter_user_id" db:"inviter_user_id"`
	InviteeEmail  string         `json:"invitee_email" db:"invitee_email"`
	Status        ReferralStatus `json:"status" db:"status"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
	CompletedAt   *time.Time     `json:"completed_at,omitempty" db:"completed_at"`
}

type ReferralStatus string

const (
	ReferralPending   ReferralStatus = "pending"
	ReferralCompleted ReferralStatus = "completed"
	ReferralExpired   ReferralStatus = "expired"
)
