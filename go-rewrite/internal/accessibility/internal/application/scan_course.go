package application

import (
	"context"

	"rewritetest/internal/accessibility/internal/domain"
	"rewritetest/internal/content"
	"rewritetest/internal/courses"
	"rewritetest/internal/lms"
	"rewritetest/internal/shared/auth"
)

type ScanCourseIssue struct {
	ContentItemID int64
	ScanRule      string
	Status        string
	Severity      string
	ContentXPath  string
	Details       map[string]any
}

type CourseRetriever interface {
	GetCourse(ctx context.Context, courseID int64) (courses.Course, error)
}

type ContentItemService interface {
	GetByCourse(ctx context.Context, courseID int64) ([]content.ContentItem, error)
	CreateManyContentItems(ctx context.Context, contentItems []content.ContentItem) (map[string]int64, error)
}

type ExternalContentRetriever interface {
	GetContent(ctx context.Context, res lms.GetContentRequest) (lms.CourseSyncDataDTO, error)
}

type CourseFileSyncService interface {
	SyncCourseFiles(ctx context.Context, courseID int64, files []lms.FileItemDTO) error
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
	issueRepository						domain.IssueRepository
	courseRetriever          CourseRetriever
	contentItemService       ContentItemService
	externalContentRetriever ExternalContentRetriever
	contentHasher            ContentHasher
	courseFileSyncService    CourseFileSyncService
	scanner                  domain.Scanner
}

func NewScanCourseUseCase(
	issueRepository	domain.IssueRepository,
	courseRetriever CourseRetriever,
	contentItemService ContentItemService,
	externalContentRetriever ExternalContentRetriever,
	contentHasher ContentHasher,
	courseFileSyncService CourseFileSyncService,
	scanner domain.Scanner,
) *ScanCourseUseCase {
	return &ScanCourseUseCase{
		issueRepository: 					issueRepository,
		courseRetriever:          courseRetriever,
		contentItemService:       contentItemService,
		externalContentRetriever: externalContentRetriever,
		contentHasher:            contentHasher,
		courseFileSyncService:    courseFileSyncService,
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

	externalCourseData, err := u.externalContentRetriever.GetContent(ctx, lms.GetContentRequest{
		CourseID:           courseID,
		TenantID:           principal.TenantID,
		UserID:             principal.AgentID,
		ExternalCourseID:   course.ExternalID,
		ExternalCourseData: course.ExternalData,
	})
	if err != nil {
		return err
	}

	err = u.courseFileSyncService.SyncCourseFiles(ctx, courseID, externalCourseData.Files)
	if err != nil {
		return err
	}

	changedContentItems, err := u.getChangedContentItems(currentContentItems, externalCourseData.ContentItems, courseID)
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
	err = u.issueRepository.DeleteByContentItemIDs(ctx, changedContentItemIDs)
	if err != nil {
		return err
	}

	var newIssues []*domain.Issue
	for _, result := range scanResults {
		newIssues = append(newIssues, domain.NewIssue(result.ContentItemID, result.ScanRule, result.ContentXPath, domain.IssueStatusActive, result.Severity, result.Details))
	}

	err = u.issueRepository.CreateMany(ctx, newIssues)
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


// getChangedContentItems compares the current content items with the external
// content items and returns a list of content items that have a different hash
// (i.e., have changed)
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
