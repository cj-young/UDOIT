package application

import (
	"context"
	"rewritetest/internal/content"
	"rewritetest/internal/courses"
	"rewritetest/internal/lms"
	"rewritetest/internal/scanner/internal/domain"
	"rewritetest/internal/shared/auth"
)

type CourseRetriever interface {
	GetCourse(ctx context.Context, courseID int64) (courses.Course, error)
}

type ContentItemService interface {
	GetByCourse(ctx context.Context, courseID int64) ([]content.ContentItem, error)
	CreateManyContentItems(ctx context.Context, contentItems []content.ContentItem) error
}

type ExternalContentRetriever interface {
	GetContent(ctx context.Context, res lms.GetContentRequest) ([]lms.ContentItemDTO, error)
}

type ContentHasher interface {
	HashContent(content string) (string, error)
}

type HashedContentItem struct {
	ExternalID  	string
	ContentHash		string
	Type					string
}

type ScanCourseUseCase struct {
	courseRetriever 					CourseRetriever
	contentItemService 			ContentItemService
	externalContentRetriever 	ExternalContentRetriever
	contentHasher 						ContentHasher
	scanner 									domain.Scanner
}

func NewScanCourseUseCase(
	courseRetriever CourseRetriever,
	contentItemRetriever ContentItemService,
	externalContentRetriever ExternalContentRetriever,
	contentHasher ContentHasher,
	scanner domain.Scanner,
) *ScanCourseUseCase {
	return &ScanCourseUseCase{
		courseRetriever:          courseRetriever,
		contentItemService:     contentItemRetriever,
		externalContentRetriever: externalContentRetriever,
		contentHasher:            contentHasher,
		scanner:                  scanner,
	}
}

type FullContentItem struct {
	ExternalID 	string
	HTML 				string
	ContentHash	string
	Type        string
}

func (u *ScanCourseUseCase) Execute(ctx context.Context, principal auth.Principal, courseID int64) error {
	// Get stored content items
	// Get content items
	// update / create content items based on external IDs
	// register content items
	// scan content items
	// create report

	currentContentItems, err := u.contentItemService.GetByCourse(ctx, courseID)
	if err != nil {
		return err
	}

	course, err := u.courseRetriever.GetCourse(ctx, courseID)
	if err != nil {
		return err
	}

	externalContentItems, err := u.externalContentRetriever.GetContent(ctx, lms.GetContentRequest{
		CourseID:            courseID,
		TenantID:            principal.TenantID,
		UserID:              principal.AgentID,
		ExternalCourseID:    course.ExternalID,
		ExternalCourseData:  course.ExternalData,
	})
	if err != nil {
		return err
	}

	changedContentItems, changedContentItemsHashed, err := u.getChangedContentItems(currentContentItems, externalContentItems, courseID)
	if err != nil {
		return err
	}

	err = u.contentItemService.CreateManyContentItems(ctx, changedContentItemsHashed)
	if err != nil {
		return err
	}

	err = u.scanner.ScanContent(ctx, changedContentItems)
	if err != nil {
		return err
	}

	return nil
}

func (u *ScanCourseUseCase) getChangedContentItems(currentContentItems []content.ContentItem, externalContentItems []lms.ContentItemDTO, courseID int64) ([]lms.ContentItemDTO, []content.ContentItem, error) {
	currentContentItemMap := make(map[string]content.ContentItem)
	for _, item := range currentContentItems {
		currentContentItemMap[item.ExternalID] = item
	}

	changedContentItems := []lms.ContentItemDTO{}
	changedContentItemsHashed := []content.ContentItem{}
	for _, item := range externalContentItems {
		contentHash, err := u.contentHasher.HashContent(item.HTML)
		if err != nil {
			return nil, nil, err
		}
		if currentItem, exists := currentContentItemMap[item.ExternalID]; !exists || currentItem.ContentHash != contentHash {
			changedContentItems = append(changedContentItems, item)
			changedContentItemsHashed = append(changedContentItemsHashed, content.ContentItem{
				ExternalID:   item.ExternalID,
				ContentHash:  contentHash,
				CourseID:    	courseID,
			})
		}
	}

	return changedContentItems, changedContentItemsHashed, nil
}