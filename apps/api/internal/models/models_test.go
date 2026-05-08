package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestUser_BeforeCreate_GeneratesUUID(t *testing.T) {
	u := &User{}
	assert.NoError(t, u.BeforeCreate(&gorm.DB{}))
	assert.NotEqual(t, uuid.Nil, u.ID)
}

func TestUser_BeforeCreate_KeepsExisting(t *testing.T) {
	existing := uuid.New()
	u := &User{ID: existing}
	assert.NoError(t, u.BeforeCreate(&gorm.DB{}))
	assert.Equal(t, existing, u.ID)
}

func TestActivity_BeforeCreate_GeneratesUUID(t *testing.T) {
	a := &Activity{}
	assert.NoError(t, a.BeforeCreate(&gorm.DB{}))
	assert.NotEqual(t, uuid.Nil, a.ID)
}

func TestActivity_BeforeCreate_KeepsExisting(t *testing.T) {
	existing := uuid.New()
	a := &Activity{ID: existing}
	assert.NoError(t, a.BeforeCreate(&gorm.DB{}))
	assert.Equal(t, existing, a.ID)
}

func TestActivityLog_BeforeCreate_GeneratesUUID(t *testing.T) {
	l := &ActivityLog{}
	assert.NoError(t, l.BeforeCreate(&gorm.DB{}))
	assert.NotEqual(t, uuid.Nil, l.ID)
}

func TestActivityLog_BeforeCreate_KeepsExisting(t *testing.T) {
	existing := uuid.New()
	l := &ActivityLog{ID: existing}
	assert.NoError(t, l.BeforeCreate(&gorm.DB{}))
	assert.Equal(t, existing, l.ID)
}

func TestActivityLog_TableName(t *testing.T) {
	assert.Equal(t, "activity_logs", ActivityLog{}.TableName())
}
