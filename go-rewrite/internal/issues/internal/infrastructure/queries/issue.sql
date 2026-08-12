-- name: DeleteIssuesByContentItemIDs :exec
DELETE FROM issue WHERE content_item_id IN (sqlc.slice('content_item_ids'));

-- name: CreateIssue :exec
INSERT INTO issue (content_item_id, scan_rule, content_xpath, issue_status, issue_severity, fixed_by, fixed_on, details)
VALUES (
  sqlc.arg('content_item_id'),
  sqlc.arg('scan_rule'),
  sqlc.arg('content_xpath'),
  sqlc.arg('issue_status'),
  sqlc.arg('issue_severity'),
  sqlc.arg('fixed_by'),
  sqlc.arg('fixed_on'),
  sqlc.arg('details')
);