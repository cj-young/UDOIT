package application

import (
	"context"

	"rewritetest/internal/content"
	"rewritetest/internal/courses"
	"rewritetest/internal/issues"
	"rewritetest/internal/lms"
	"rewritetest/internal/scanner/internal/domain"
	"rewritetest/internal/shared/auth"
)

type CourseRetriever interface {
	GetCourse(ctx context.Context, courseID int64) (courses.Course, error)
}

type ContentItemService interface {
	GetByCourse(ctx context.Context, courseID int64) ([]content.ContentItem, error)
	CreateManyContentItems(ctx context.Context, contentItems []content.ContentItem) (map[string]int64, error)
}

type IssueService interface {
	RegisterNewIssues(ctx context.Context, newIssues []issues.Issue, contentItemIDs []int64) error
	DeleteByContentItemIDs(ctx context.Context, contentItemIDs []int64) error
}

type ExternalContentRetriever interface {
	GetContent(ctx context.Context, res lms.GetContentRequest) ([]lms.ContentItemDTO, error)
}

type ContentHasher interface {
	HashContent(content string) (string, error)
}

type HashedContentItem struct {
	ExternalID   string
	ExternalData map[string]any
	ContentHash  string
	Type         string
}

type ScanCourseUseCase struct {
	courseRetriever          CourseRetriever
	contentItemService       ContentItemService
	externalContentRetriever ExternalContentRetriever
	contentHasher            ContentHasher
	issueService             IssueService
	scanner                  domain.Scanner
}

func NewScanCourseUseCase(
	courseRetriever CourseRetriever,
	contentItemService ContentItemService,
	externalContentRetriever ExternalContentRetriever,
	contentHasher ContentHasher,
	issueService IssueService,
	scanner domain.Scanner,
) *ScanCourseUseCase {
	return &ScanCourseUseCase{
		courseRetriever:          courseRetriever,
		contentItemService:       contentItemService,
		externalContentRetriever: externalContentRetriever,
		contentHasher:            contentHasher,
		issueService:             issueService,
		scanner:                  scanner,
	}
}

type FullContentItem struct {
	ExternalID   string
	ExternalData map[string]any
	HTML         string
	ContentHash  string
	Type         string
}

func (u *ScanCourseUseCase) Execute(ctx context.Context, principal auth.Principal, courseID int64) error {
	currentContentItems, err := u.contentItemService.GetByCourse(ctx, courseID)
	if err != nil {
		return err
	}

	course, err := u.courseRetriever.GetCourse(ctx, courseID)
	if err != nil {
		return err
	}

	externalContentItems, err := u.externalContentRetriever.GetContent(ctx, lms.GetContentRequest{
		CourseID:           courseID,
		TenantID:           principal.TenantID,
		UserID:             principal.AgentID,
		ExternalCourseID:   course.ExternalID,
		ExternalCourseData: course.ExternalData,
	})
	if err != nil {
		return err
	}

	changedContentItems, err := u.getChangedContentItems(currentContentItems, externalContentItems, courseID)
	if err != nil {
		return err
	}

	changedContentItemsHashed := make([]content.ContentItem, len(changedContentItems))
	for i, item := range changedContentItems {
		changedContentItemsHashed[i] = content.ContentItem{
			ID:           item.ID,
			ExternalID:   item.ExternalID,
			ExternalData: item.ExternalData,
			ContentHash:  item.ContentHash,
			CourseID:     item.CourseID,
		}
	}
	idMap, err := u.contentItemService.CreateManyContentItems(ctx, changedContentItemsHashed)
	if err != nil {
		return err
	}

	// Update the IDs of the changed content items with the upserted IDs returned
	// from the database
	for _, item := range changedContentItems {
		if id, exists := idMap[item.ExternalID]; exists {
			item.ID = id
		}
	}

	scanItems := make([]domain.ScanItem, len(changedContentItems))
	for i, item := range changedContentItems {
		scanItems[i] = domain.ScanItem{
			ContentItemID: item.ID,
			HTML:          item.HTML,
			Type:          item.Type,
		}
	}
	scanResults, err := u.scanner.ScanContent(ctx, scanItems)
	if err != nil {
		return err
	}

	changedContentItemIDs := make([]int64, len(changedContentItems))
	for i, item := range changedContentItems {
		changedContentItemIDs[i] = item.ID
	}
	err = u.issueService.DeleteByContentItemIDs(ctx, changedContentItemIDs)
	if err != nil {
		return err
	}

	newIssues := make([]issues.Issue, len(scanResults))
	for i, result := range scanResults {
		newIssues[i] = issues.Issue{
			ContentItemID: result.ContentItemID,
			ScanRule:      result.ScanRule,
			Status:        domain.ScanIssueStatusActive.String(),
			Severity:      result.Severity.String(),
			ContentXPath:  result.ContentXPath,
			Details:       result.Details,
		}
	}

	err = u.issueService.RegisterNewIssues(ctx, newIssues, changedContentItemIDs)
	if err != nil {
		return err
	}

	return nil
}

type ChangedContentItem struct {
	ID           int64
	CourseID     int64
	ExternalID   string
	ExternalData map[string]any
	ContentHash  string
	Type         string
	HTML         string
}

func (u *ScanCourseUseCase) getChangedContentItems(
	currentContentItems []content.ContentItem,
	externalContentItems []lms.ContentItemDTO,
	courseID int64,
) ([]*ChangedContentItem, error) {
	currentContentItemMap := make(map[string]content.ContentItem)
	for _, item := range currentContentItems {
		currentContentItemMap[item.ExternalID] = item
	}

	changedContentItems := []*ChangedContentItem{}
	for _, item := range externalContentItems {
		contentHash, err := u.contentHasher.HashContent(item.HTML)
		if err != nil {
			return nil, err
		}
		if currentItem, exists := currentContentItemMap[item.ExternalID]; !exists || currentItem.ContentHash != contentHash {
			var id int64
			if exists {
				id = currentItem.ID
			}
			changedContentItems = append(changedContentItems, &ChangedContentItem{
				ID:           id,
				CourseID:     courseID,
				ExternalID:   item.ExternalID,
				ExternalData: item.ExternalData,
				ContentHash:  contentHash,
				Type:         item.Type,
				HTML:         item.HTML,
			})
		}
	}

	return changedContentItems, nil
}
