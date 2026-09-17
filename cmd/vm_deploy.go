package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yunxiao-cli/yunxiao/internal/risk"
)

var pipelineVMDeployCmd = &cobra.Command{
	Use:     "vm-deploy",
	Aliases: []string{"deploy-order"},
	Short:   "Flow VM deploy orders",
	Long: `VM deploy order ops (operations/flow/vmDeployOrder.ts).

  yunxiao pipeline vm-deploy get|stop|resume --pipeline-id <id> --deploy-id <id>
  yunxiao pipeline vm-deploy machine skip|retry|log --pipeline-id <id> --deploy-id <id> --machine-sn <sn>

Risk: get/log=read; stop/resume/skip/retry=high-risk-write`,
}

func vmDeployPath(pid, did, suffix string) string {
	return "/pipelines/" + pid + "/deploy/" + did + suffix
}

var pipelineVMDeployGetCmd = &cobra.Command{
	Use: "get", Short: "Get VM deploy order",
	Long: "Risk: read\nHTTP: GET .../pipelines/{id}/deploy/{deployId}",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		pid, _ := cmd.Flags().GetString("pipeline-id")
		did, _ := cmd.Flags().GetString("deploy-id")
		if err := requireFlags("pipeline-id", pid, "deploy-id", did); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.FlowPath(cmd.Context(), vmDeployPath(pid, did, ""))
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runRead(cmd.Context(), c, "GET", path, nil, nil, map[string]any{"risk": risk.Read}, nil))
	},
}

func vmDeployPut(action, suffix string) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		pid, _ := cmd.Flags().GetString("pipeline-id")
		did, _ := cmd.Flags().GetString("deploy-id")
		if err := requireFlags("pipeline-id", pid, "deploy-id", did); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.FlowPath(cmd.Context(), vmDeployPath(pid, did, suffix))
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runJSONMutating(cmd.Context(), c, action, risk.HighRiskWrite, "PUT", path, nil, nil, nil))
	}
}

var pipelineVMDeployStopCmd = &cobra.Command{
	Use: "stop", Short: "Stop VM deploy order (high-risk-write)",
	Run: vmDeployPut("pipeline vm-deploy stop", "/stop"),
}
var pipelineVMDeployResumeCmd = &cobra.Command{
	Use: "resume", Short: "Resume VM deploy order (high-risk-write)",
	Run: vmDeployPut("pipeline vm-deploy resume", "/resume"),
}

func vmMachineAction(action, suffix string, mutating bool) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		pid, _ := cmd.Flags().GetString("pipeline-id")
		did, _ := cmd.Flags().GetString("deploy-id")
		msn, _ := cmd.Flags().GetString("machine-sn")
		if err := requireFlags("pipeline-id", pid, "deploy-id", did, "machine-sn", msn); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.FlowPath(cmd.Context(), vmDeployPath(pid, did, "/machine/"+msn+suffix))
		if err != nil {
			handleErr(err)
			return
		}
		if !mutating {
			handleErr(runRead(cmd.Context(), c, "GET", path, nil, nil, map[string]any{"risk": risk.Read}, nil))
			return
		}
		handleErr(runJSONMutating(cmd.Context(), c, action, risk.HighRiskWrite, "PUT", path, nil, nil, nil))
	}
}

var pipelineVMMachineSkipCmd = &cobra.Command{
	Use: "skip", Short: "Skip a VM deploy machine (high-risk-write)",
	Run: vmMachineAction("pipeline vm-deploy machine skip", "/skip", true),
}
var pipelineVMMachineRetryCmd = &cobra.Command{
	Use: "retry", Short: "Retry a VM deploy machine (high-risk-write)",
	Run: vmMachineAction("pipeline vm-deploy machine retry", "/retry", true),
}
var pipelineVMMachineLogCmd = &cobra.Command{
	Use: "log", Short: "Get VM deploy machine log",
	Run: vmMachineAction("pipeline vm-deploy machine log", "/log", false),
}

var pipelineVMMachineCmd = &cobra.Command{Use: "machine", Short: "Per-machine VM deploy actions"}

func init() {
	for _, c := range []*cobra.Command{pipelineVMDeployGetCmd, pipelineVMDeployStopCmd, pipelineVMDeployResumeCmd} {
		c.Flags().String("pipeline-id", "", "pipeline id (required)")
		c.Flags().String("deploy-id", "", "deploy order id (required)")
	}
	for _, c := range []*cobra.Command{pipelineVMMachineSkipCmd, pipelineVMMachineRetryCmd, pipelineVMMachineLogCmd} {
		c.Flags().String("pipeline-id", "", "pipeline id (required)")
		c.Flags().String("deploy-id", "", "deploy order id (required)")
		c.Flags().String("machine-sn", "", "machine sn (required)")
	}
	pipelineVMMachineCmd.AddCommand(pipelineVMMachineSkipCmd, pipelineVMMachineRetryCmd, pipelineVMMachineLogCmd)
	pipelineVMDeployCmd.AddCommand(pipelineVMDeployGetCmd, pipelineVMDeployStopCmd, pipelineVMDeployResumeCmd, pipelineVMMachineCmd)
	pipelineCmd.AddCommand(pipelineVMDeployCmd)
}
