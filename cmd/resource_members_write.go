package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yunxiao-cli/yunxiao/internal/risk"
)

func rmBasePath(rtype, rid string) string {
	return "/resourceMembers/resourceTypes/" + rtype + "/resourceIds/" + rid
}

var pipelineRMCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Add a resource member (high-risk-write)",
	Long:  "Risk: high-risk-write\nHTTP: POST .../resourceMembers/...?roleName=&userId=\nSource: createResourceMemberFunc",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		rtype, _ := cmd.Flags().GetString("resource-type")
		rid, _ := cmd.Flags().GetString("resource-id")
		role, _ := cmd.Flags().GetString("role-name")
		uid, _ := cmd.Flags().GetString("user-id")
		if err := requireFlags("resource-type", rtype, "resource-id", rid, "role-name", role, "user-id", uid); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.FlowPath(cmd.Context(), rmBasePath(rtype, rid))
		if err != nil {
			handleErr(err)
			return
		}
		q := map[string]string{"roleName": role, "userId": uid}
		handleErr(runJSONMutating(cmd.Context(), c, "pipeline resource-members create", risk.HighRiskWrite, "POST", path, q, nil, nil))
	},
}

var pipelineRMUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a resource member role (high-risk-write)",
	Long:  "Risk: high-risk-write\nHTTP: PUT .../resourceMembers/...?roleName=&userId=\nSource: updateResourceMemberFunc",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		rtype, _ := cmd.Flags().GetString("resource-type")
		rid, _ := cmd.Flags().GetString("resource-id")
		role, _ := cmd.Flags().GetString("role-name")
		uid, _ := cmd.Flags().GetString("user-id")
		if err := requireFlags("resource-type", rtype, "resource-id", rid, "role-name", role, "user-id", uid); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.FlowPath(cmd.Context(), rmBasePath(rtype, rid))
		if err != nil {
			handleErr(err)
			return
		}
		q := map[string]string{"roleName": role, "userId": uid}
		handleErr(runJSONMutating(cmd.Context(), c, "pipeline resource-members update", risk.HighRiskWrite, "PUT", path, q, nil, nil))
	},
}

var pipelineRMDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Remove a resource member (high-risk-write)",
	Long:  "Risk: high-risk-write\nHTTP: DELETE .../resourceMembers/...?userId=\nSource: deleteResourceMemberFunc",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		rtype, _ := cmd.Flags().GetString("resource-type")
		rid, _ := cmd.Flags().GetString("resource-id")
		uid, _ := cmd.Flags().GetString("user-id")
		if err := requireFlags("resource-type", rtype, "resource-id", rid, "user-id", uid); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.FlowPath(cmd.Context(), rmBasePath(rtype, rid))
		if err != nil {
			handleErr(err)
			return
		}
		q := map[string]string{"userId": uid}
		handleErr(runJSONMutating(cmd.Context(), c, "pipeline resource-members delete", risk.HighRiskWrite, "DELETE", path, q, nil, nil))
	},
}

var pipelineRMTransferCmd = &cobra.Command{
	Use:   "transfer-owner",
	Short: "Transfer resource owner (high-risk-write)",
	Long:  "Risk: high-risk-write\nHTTP: POST .../transfer/owner?newOwnerId=\nSource: updateResourceOwnerFunc",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		rtype, _ := cmd.Flags().GetString("resource-type")
		rid, _ := cmd.Flags().GetString("resource-id")
		nid, _ := cmd.Flags().GetString("new-owner-id")
		if err := requireFlags("resource-type", rtype, "resource-id", rid, "new-owner-id", nid); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.FlowPath(cmd.Context(), rmBasePath(rtype, rid)+"/transfer/owner")
		if err != nil {
			handleErr(err)
			return
		}
		q := map[string]string{"newOwnerId": nid}
		handleErr(runJSONMutating(cmd.Context(), c, "pipeline resource-members transfer-owner", risk.HighRiskWrite, "POST", path, q, nil, nil))
	},
}

func init() {
	for _, c := range []*cobra.Command{pipelineRMCreateCmd, pipelineRMUpdateCmd} {
		c.Flags().String("resource-type", "", "pipeline|hostGroup (required)")
		c.Flags().String("resource-id", "", "resource id (required)")
		c.Flags().String("role-name", "", "role name (required)")
		c.Flags().String("user-id", "", "user id (required)")
	}
	pipelineRMDeleteCmd.Flags().String("resource-type", "", "pipeline|hostGroup (required)")
	pipelineRMDeleteCmd.Flags().String("resource-id", "", "resource id (required)")
	pipelineRMDeleteCmd.Flags().String("user-id", "", "user id (required)")
	pipelineRMTransferCmd.Flags().String("resource-type", "", "pipeline|hostGroup (required)")
	pipelineRMTransferCmd.Flags().String("resource-id", "", "resource id (required)")
	pipelineRMTransferCmd.Flags().String("new-owner-id", "", "new owner user id (required)")
	pipelineRMCmd.AddCommand(pipelineRMCreateCmd, pipelineRMUpdateCmd, pipelineRMDeleteCmd, pipelineRMTransferCmd)
}
