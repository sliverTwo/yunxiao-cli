package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/yunxiao-cli/yunxiao/internal/client"
	"github.com/yunxiao-cli/yunxiao/internal/zhiyi"
)

// lookupCancelReasonFieldID discovers the 取消原因 custom field id for a workitem type
// (in-process cache per type_id). Best-effort: empty id on failure.
func lookupCancelReasonFieldID(ctx context.Context, c *client.Client, spaceID, typeID string) string {
	typeID = strings.TrimSpace(typeID)
	spaceID = strings.TrimSpace(spaceID)
	if typeID == "" || spaceID == "" || c == nil {
		return ""
	}
	if id, ok := zhiyi.CachedCancelReasonFieldID(typeID); ok {
		return id
	}
	path, err := c.ProjexPath(ctx, "/projects/"+spaceID+"/workitemTypes/"+typeID+"/fields")
	if err != nil {
		zhiyi.StoreCancelReasonFieldID(typeID, "")
		return ""
	}
	var raw any
	if err := c.Get(ctx, path, nil, &raw); err != nil {
		// Do not cache hard failures so a transient blip can retry next call.
		return ""
	}
	id := zhiyi.FindFieldIDByName(raw, zhiyi.CancelReasonFieldName)
	zhiyi.StoreCancelReasonFieldID(typeID, id)
	return id
}

// applyCancelReasonFlag looks up 取消原因 and merges --cancel-reason into body.
func applyCancelReasonFlag(ctx context.Context, c *client.Client, item map[string]any, body map[string]any, cancelReason string) (fieldID string, err error) {
	cancelReason = strings.TrimSpace(cancelReason)
	if cancelReason == "" {
		return "", nil
	}
	spaceID := zhiyi.ResolveSpaceID(item, profileSpaceID(), "")
	typeID := workitemTypeID(item)
	if spaceID == "" || typeID == "" {
		return "", fmt.Errorf("--cancel-reason needs workitem space and type; could not resolve from item (space=%q type=%q)", spaceID, typeID)
	}
	fieldID = lookupCancelReasonFieldID(ctx, c, spaceID, typeID)
	if fieldID == "" {
		return "", fmt.Errorf("could not find field named %q on type %s; use --custom-fields with the field id from: workitem fields --space-id %s --type-id %s",
			zhiyi.CancelReasonFieldName, typeID, spaceID, typeID)
	}
	zhiyi.MergeCancelReasonIntoBody(body, fieldID, cancelReason)
	return fieldID, nil
}

// softCancelReasonWarning optionally checks fields when status looks cancelled.
func softCancelReasonWarning(ctx context.Context, c *client.Client, item map[string]any, status, cancelReason string) string {
	if !zhiyi.LooksLikeCancelStatus(status) || strings.TrimSpace(cancelReason) != "" {
		return ""
	}
	spaceID := zhiyi.ResolveSpaceID(item, profileSpaceID(), "")
	typeID := workitemTypeID(item)
	if spaceID == "" || typeID == "" {
		return ""
	}
	fieldID := lookupCancelReasonFieldID(ctx, c, spaceID, typeID)
	return zhiyi.CancelReasonSoftWarning(status, cancelReason, fieldID)
}

func fetchWorkItemMap(ctx context.Context, c *client.Client, id string) (map[string]any, error) {
	path, err := c.ProjexPath(ctx, "/workitems/"+id)
	if err != nil {
		return nil, err
	}
	var item map[string]any
	if err := c.Get(ctx, path, nil, &item); err != nil {
		return nil, err
	}
	return item, nil
}
