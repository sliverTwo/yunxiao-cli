package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yunxiao-cli/yunxiao/internal/client"
	"github.com/yunxiao-cli/yunxiao/internal/risk"
	"github.com/yunxiao-cli/yunxiao/internal/zhiyi"
)

var codeupMrsPlusCreateCmd = &cobra.Command{
	Use:   "+create",
	Short: "Shortcut: create MR with repo alias, WIP title, work-item link",
	Long: `Risk: high-risk-write (requires --yes after confirmation; prefer --dry-run first)

Zhiyi-oriented wrapper around Codeup changeRequests. Does not replace typed
"codeup mrs create".

  yunxiao codeup mrs +create --profile zhiyi \
    --repo iipmes_gy --source feat/x --target master \
    --title "fix" --work-item ZYPT-5768 --wip --dry-run

  yunxiao codeup mrs +create --repo <repo-id> --source feat/x --title "fix" --yes

--repo accepts numeric id or profile.repositories alias. Profile optional when --repo is numeric.
--target defaults to master. --wip prefixes "WIP: " when target is master.
--reviewer is comma-separated userIds (OpenAPI reviewerUserIds), same as typed mrs create.`,
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)

		repo, _ := cmd.Flags().GetString("repo")
		source, _ := cmd.Flags().GetString("source")
		target, _ := cmd.Flags().GetString("target")
		title, _ := cmd.Flags().GetString("title")
		desc, _ := cmd.Flags().GetString("description")
		workItem, _ := cmd.Flags().GetString("work-item")
		reviewer, _ := cmd.Flags().GetString("reviewer")
		wip, _ := cmd.Flags().GetBool("wip")

		if err := requireFlags("repo", repo, "source", source, "title", title); err != nil {
			handleErr(err)
			return
		}
		if strings.TrimSpace(target) == "" {
			target = "master"
		}

		repositoryID, err := resolveCodeupRepo(repo)
		if err != nil {
			handleErr(err)
			return
		}
		pf, err := applyActiveProfileOrg()
		if err != nil {
			handleErr(err)
			return
		}

		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}

		var workItemIDs []string
		if strings.TrimSpace(workItem) != "" {
			wpath, err := c.ProjexPath(cmd.Context(), "/workitems/"+strings.TrimSpace(workItem))
			if err != nil {
				handleErr(err)
				return
			}
			var item map[string]any
			if err := c.Get(cmd.Context(), wpath, nil, &item); err != nil {
				handleErr(fmt.Errorf("resolve --work-item: %w", err))
				return
			}
			internal := zhiyi.InternalID(item)
			if internal == "" {
				handleErr(fmt.Errorf("无法解析工作项内部 id: %s", workItem))
				return
			}
			workItemIDs = []string{internal}
		}

		title = zhiyi.WithWipTitle(title, target, wip)
		reviewerIDs := zhiyi.SplitUserIDs(reviewer)

		body := map[string]any{
			"title":           title,
			"sourceBranch":    source,
			"targetBranch":    target,
			"sourceProjectId": repositoryID,
			"targetProjectId": repositoryID,
			"createFrom":      "WEB",
			"description":     desc,
			"reviewerUserIds": reviewerIDs,
			"workItemIds":     workItemIDs,
		}

		repoID := client.EncodeRepoID(repositoryID)
		path, err := c.CodeupPath(cmd.Context(), "/repositories/"+repoID+"/changeRequests")
		if err != nil {
			handleErr(err)
			return
		}

		meta := map[string]any{"risk": risk.HighRiskWrite, "repository_id": repositoryID}
		if pf != nil {
			meta["profile"] = pf.Name
		}

		handleErr(runJSONMutating(cmd.Context(), c, "codeup mrs +create", risk.HighRiskWrite, "POST", path, nil, body, func(out any, m map[string]any) (any, map[string]any) {
			for k, v := range meta {
				m[k] = v
			}
			zhiyi.EnrichMergeRequestMeta(m, asStringMap(out))
			return out, m
		}))
	},
}

func init() {
	codeupMrsPlusCreateCmd.Flags().String("repo", "", "numeric repositoryId or alias (required)")
	codeupMrsPlusCreateCmd.Flags().String("source", "", "source branch (required)")
	codeupMrsPlusCreateCmd.Flags().String("target", "master", "target branch (default master)")
	codeupMrsPlusCreateCmd.Flags().String("title", "", "MR title (required)")
	codeupMrsPlusCreateCmd.Flags().String("description", "", "MR description")
	codeupMrsPlusCreateCmd.Flags().String("work-item", "", "ZYPT serial or internal id (resolved to workItemIds)")
	codeupMrsPlusCreateCmd.Flags().String("reviewer", "", "optional reviewer userId(s), comma-separated (OpenAPI reviewerUserIds; same as mrs create)")
	codeupMrsPlusCreateCmd.Flags().Bool("wip", false, "prefix WIP: when target is master")
	codeupMrsCmd.AddCommand(codeupMrsPlusCreateCmd)
}
