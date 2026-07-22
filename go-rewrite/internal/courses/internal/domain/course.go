package domain

import "time"

type Course struct {
	id int64
	title string
	tenantID int64
	isActive bool
	isDirty bool
	updatedAt time.Time
}

func NewCourse(title string, tenantID int64, isActive bool, isDirty bool, updatedAt time.Time) *Course {
	return &Course{
		title: title,
		tenantID: tenantID,
		isActive: isActive,
		isDirty: isDirty,
		updatedAt: updatedAt,
	}
}

func (c *Course) ID() int64 {
	return c.id
}

func (c *Course) Title() string {
	return c.title
}

func (c *Course) IsActive() bool {
	return c.isActive
}

func (c *Course) TenantID() int64 {
	return c.tenantID
}

func (c *Course) IsDirty() bool {
	return c.isDirty
}

func (c *Course) UpdatedAt() time.Time {
	return c.updatedAt
}