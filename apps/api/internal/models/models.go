package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Provider string

const (
	ProviderGoogle   Provider = "google"
	ProviderPassword Provider = "password"
	ProviderGitHub   Provider = "github"
	ProviderApple    Provider = "apple"
	ProviderFacebook Provider = "facebook"
)

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null"                           json:"email"`
	Name      string    `gorm:"not null"                                       json:"name"`
	AvatarURL string    `                                                      json:"avatar_url"`
	CreatedAt time.Time `                                                      json:"created_at"`
	UpdatedAt time.Time `                                                      json:"updated_at"`
}

type UserIdentity struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID         uuid.UUID `gorm:"type:uuid;not null;index"                       json:"user_id"`
	Provider       Provider  `gorm:"not null"                                       json:"provider"`
	ProviderUserID string    `gorm:"not null"                                       json:"provider_user_id"`
	CreatedAt      time.Time `                                                      json:"created_at"`
	UpdatedAt      time.Time `                                                      json:"updated_at"`
}

type Activity struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID         uuid.UUID  `gorm:"type:uuid;not null;index"                       json:"user_id"`
	Name           string     `gorm:"not null"                                       json:"name"`
	Description    string     `                                                      json:"description"`
	Color          string     `gorm:"not null"                                       json:"color"`
	CreatedAt      time.Time  `                                                      json:"created_at"`
	UpdatedAt      time.Time  `                                                      json:"updated_at"`
	LastLoggedDate *time.Time `gorm:"-:all"                                          json:"last_logged_date"`
}

type ActivityLog struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ActivityID uuid.UUID `gorm:"type:uuid;not null;index"                       json:"activity_id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index"                       json:"user_id"`
	LoggedDate time.Time `gorm:"type:date;not null"                             json:"logged_date"`
	CreatedAt  time.Time `                                                      json:"created_at"`
}

func (ActivityLog) TableName() string { return "activity_logs" }

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (i *UserIdentity) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}

func (a *Activity) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

func (l *ActivityLog) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}
