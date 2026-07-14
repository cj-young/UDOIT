package domain

import "time"

type Course struct {
	id int64
	title string
	isActive bool
	isDirty bool
	updatedAt time.Time
}