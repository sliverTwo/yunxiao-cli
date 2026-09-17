package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/yunxiao-cli/yunxiao/internal/risk"
)

var appstackRWCmd = &cobra.Command{
	Use:     "release-workflows",
	Aliases: []string{"rw"},
	Short:   "AppStack release workflows / stages",
	Long: `App release workflows (operations/appstack/releaseWorkflows.ts).

  yunxiao appstack release-workflows list|briefs --app <name>
  yunxiao appstack release-workflows stage get|briefs|runs|execute|…
  yunxiao appstack release-workflows system list|create --system <name>

Risk: list/get=read; execute/cancel/retry/skip/pass/refuse/update=high-risk-write`,
}

func rwBase(app, wf, stage string) (string, error) {
	return "/apps/" + app + "/releaseWorkflows/" + wf + "/releaseStages/" + stage, nil
}

var appstackRWListCmd = &cobra.Command{
	Use: "list", Short: "List app release workflows",
	Long: "Risk: read\nHTTP: GET .../apps/{app}/releaseWorkflows",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		app, _ := cmd.Flags().GetString("app")
		if err := requireFlags("app", app); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/apps/"+app+"/releaseWorkflows")
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runRead(cmd.Context(), c, "GET", path, nil, nil, map[string]any{"risk": risk.Read}, nil))
	},
}

var appstackRWBriefsCmd = &cobra.Command{
	Use: "briefs", Short: "List app release workflow briefs",
	Long: "Risk: read\nHTTP: GET .../apps/{app}/releaseWorkflowBriefs",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		app, _ := cmd.Flags().GetString("app")
		if err := requireFlags("app", app); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/apps/"+app+"/releaseWorkflowBriefs")
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runRead(cmd.Context(), c, "GET", path, nil, nil, map[string]any{"risk": risk.Read}, nil))
	},
}

var appstackRWStageCmd = &cobra.Command{Use: "stage", Short: "Release stage operations"}

func requireRWStage(cmd *cobra.Command) (app, wf, stage string, err error) {
	app, _ = cmd.Flags().GetString("app")
	wf, _ = cmd.Flags().GetString("workflow-sn")
	stage, _ = cmd.Flags().GetString("stage-sn")
	err = requireFlags("app", app, "workflow-sn", wf, "stage-sn", stage)
	return
}

var appstackRWStageGetCmd = &cobra.Command{
	Use: "get", Short: "Get a release stage",
	Long: "Risk: read\nHTTP: GET .../releaseWorkflow/{wf}/releaseStage/{stage}",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		app, wf, stage, err := requireRWStage(cmd)
		if err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		// Note singular path segments per MCP getReleaseWorkflowStage
		path, err := c.AppstackPath(cmd.Context(), "/apps/"+app+"/releaseWorkflow/"+wf+"/releaseStage/"+stage)
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runRead(cmd.Context(), c, "GET", path, nil, nil, map[string]any{"risk": risk.Read}, nil))
	},
}

var appstackRWStageBriefsCmd = &cobra.Command{
	Use: "briefs", Short: "List stage briefs for a workflow",
	Long: "Risk: read\nHTTP: GET .../releaseWorkflow/{wf}/releaseStageBriefs",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		app, _ := cmd.Flags().GetString("app")
		wf, _ := cmd.Flags().GetString("workflow-sn")
		if err := requireFlags("app", app, "workflow-sn", wf); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/apps/"+app+"/releaseWorkflow/"+wf+"/releaseStageBriefs")
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runRead(cmd.Context(), c, "GET", path, nil, nil, map[string]any{"risk": risk.Read}, nil))
	},
}

var appstackRWStageRunsCmd = &cobra.Command{
	Use: "runs", Short: "List stage execution runs",
	Long: "Risk: read\nHTTP: GET .../releaseStages/{stage}/executions",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		app, wf, stage, err := requireRWStage(cmd)
		if err != nil {
			handleErr(err)
			return
		}
		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		base, _ := rwBase(app, wf, stage)
		path, err := c.AppstackPath(cmd.Context(), base+"/executions")
		if err != nil {
			handleErr(err)
			return
		}
		q := map[string]string{}
		if page > 0 {
			q["page"] = strconv.Itoa(page)
		}
		if perPage > 0 {
			q["perPage"] = strconv.Itoa(perPage)
		}
		handleErr(runRead(cmd.Context(), c, "GET", path, q, nil, map[string]any{"risk": risk.Read}, nil))
	},
}

var appstackRWStageExecuteCmd = &cobra.Command{
	Use: "execute", Short: "Execute a release stage (high-risk-write)",
	Long: "Risk: high-risk-write\nHTTP: POST .../releaseStages/{stage}:execute\nBody: execution object (use --data or --app-release-sn)",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		app, wf, stage, err := requireRWStage(cmd)
		if err != nil {
			handleErr(err)
			return
		}
		dataStr, _ := cmd.Flags().GetString("data")
		dataFile, _ := cmd.Flags().GetString("data-file")
		appReleaseSn, _ := cmd.Flags().GetString("app-release-sn")
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		base, _ := rwBase(app, wf, stage)
		path, err := c.AppstackPath(cmd.Context(), base+":execute")
		if err != nil {
			handleErr(err)
			return
		}
		body := map[string]any{}
		extra, err := loadJSONBodyFromFlags(dataStr, dataFile)
		if err != nil {
			handleErr(err)
			return
		}
		if extra != nil {
			m := asStringMap(extra)
			if m == nil {
				handleErr(fmt.Errorf("--data/--data-file must be a JSON object"))
				return
			}
			for k, v := range m {
				body[k] = v
			}
		}
		if appReleaseSn != "" {
			body["appReleaseSn"] = appReleaseSn
		}
		handleErr(runJSONMutating(cmd.Context(), c, "appstack release-workflows stage execute", risk.HighRiskWrite, "POST", path, nil, body, nil))
	},
}

func rwExecAction(action, suffix string, needJob bool) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		app, wf, stage, err := requireRWStage(cmd)
		if err != nil {
			handleErr(err)
			return
		}
		execNum, _ := cmd.Flags().GetString("execution")
		jobID, _ := cmd.Flags().GetString("job-id")
		if err := requireFlags("execution", execNum); err != nil {
			handleErr(err)
			return
		}
		if needJob {
			if err := requireFlags("job-id", jobID); err != nil {
				handleErr(err)
				return
			}
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		base, _ := rwBase(app, wf, stage)
		path, err := c.AppstackPath(cmd.Context(), base+"/executions/"+execNum+":"+suffix)
		if err != nil {
			handleErr(err)
			return
		}
		q := map[string]string{}
		if jobID != "" {
			q["jobId"] = jobID
		}
		handleErr(runJSONMutating(cmd.Context(), c, action, risk.HighRiskWrite, "POST", path, q, nil, nil))
	}
}

var appstackRWStageCancelCmd = &cobra.Command{
	Use: "cancel", Short: "Cancel a stage execution (high-risk-write)",
	Long: "Risk: high-risk-write\nHTTP: POST .../executions/{n}:cancel",
	Run:  rwExecAction("appstack rw stage cancel", "cancel", false),
}
var appstackRWStageRetryCmd = &cobra.Command{
	Use: "retry", Short: "Retry a stage pipeline job (high-risk-write)",
	Long: "Risk: high-risk-write\nHTTP: POST .../executions/{n}:retry?jobId=",
	Run:  rwExecAction("appstack rw stage retry", "retry", true),
}
var appstackRWStageSkipCmd = &cobra.Command{
	Use: "skip", Short: "Skip a stage pipeline job (high-risk-write)",
	Long: "Risk: high-risk-write\nHTTP: POST .../executions/{n}:skip?jobId=",
	Run:  rwExecAction("appstack rw stage skip", "skip", true),
}
var appstackRWStagePassCmd = &cobra.Command{
	Use: "pass", Short: "Pass stage pipeline validate (high-risk-write)",
	Long: "Risk: high-risk-write\nHTTP: POST .../executions/{n}:passPipelineValidate?jobId=",
	Run:  rwExecAction("appstack rw stage pass", "passPipelineValidate", true),
}
var appstackRWStageRefuseCmd = &cobra.Command{
	Use: "refuse", Short: "Refuse stage pipeline validate (high-risk-write)",
	Long: "Risk: high-risk-write\nHTTP: POST .../executions/{n}:refusePipelineValidate?jobId=",
	Run:  rwExecAction("appstack rw stage refuse", "refusePipelineValidate", true),
}

var appstackRWStagePipelineRunCmd = &cobra.Command{
	Use: "pipeline-run", Short: "Get stage execution pipeline run",
	Long: "Risk: read\nHTTP: GET .../executions/{n}:getPipelineRun",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		app, wf, stage, err := requireRWStage(cmd)
		if err != nil {
			handleErr(err)
			return
		}
		execNum, _ := cmd.Flags().GetString("execution")
		if err := requireFlags("execution", execNum); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		base, _ := rwBase(app, wf, stage)
		path, err := c.AppstackPath(cmd.Context(), base+"/executions/"+execNum+":getPipelineRun")
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runRead(cmd.Context(), c, "GET", path, nil, nil, map[string]any{"risk": risk.Read}, nil))
	},
}

var appstackRWStageJobLogCmd = &cobra.Command{
	Use: "job-log", Short: "Get stage execution job log",
	Long: "Risk: read\nHTTP: GET .../executions/{n}:pipelineJobLog?jobId=",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		app, wf, stage, err := requireRWStage(cmd)
		if err != nil {
			handleErr(err)
			return
		}
		execNum, _ := cmd.Flags().GetString("execution")
		jobID, _ := cmd.Flags().GetString("job-id")
		if err := requireFlags("execution", execNum, "job-id", jobID); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		base, _ := rwBase(app, wf, stage)
		path, err := c.AppstackPath(cmd.Context(), base+"/executions/"+execNum+":pipelineJobLog")
		if err != nil {
			handleErr(err)
			return
		}
		q := map[string]string{"jobId": jobID}
		handleErr(runRead(cmd.Context(), c, "GET", path, q, nil, map[string]any{"risk": risk.Read}, nil))
	},
}

var appstackRWStageMetaCmd = &cobra.Command{
	Use: "integrated-metadata", Short: "List stage execution integrated metadata",
	Long: "Risk: read\nHTTP: GET .../executions/{n}/integratedMetadata",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		app, wf, stage, err := requireRWStage(cmd)
		if err != nil {
			handleErr(err)
			return
		}
		execNum, _ := cmd.Flags().GetString("execution")
		if err := requireFlags("execution", execNum); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		base, _ := rwBase(app, wf, stage)
		path, err := c.AppstackPath(cmd.Context(), base+"/executions/"+execNum+"/integratedMetadata")
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runRead(cmd.Context(), c, "GET", path, nil, nil, map[string]any{"risk": risk.Read}, nil))
	},
}

var appstackRWStageUpdateCmd = &cobra.Command{
	Use: "update", Short: "Update app release stage (high-risk-write)",
	Long: "Risk: high-risk-write\nHTTP: PUT .../releaseStages/{stage}\nBody via --data JSON (stage fields).",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		app, wf, stage, err := requireRWStage(cmd)
		if err != nil {
			handleErr(err)
			return
		}
		dataStr, _ := cmd.Flags().GetString("data")
		dataFile, _ := cmd.Flags().GetString("data-file")
		raw, err := loadJSONBodyFromFlags(dataStr, dataFile)
		if err != nil {
			handleErr(err)
			return
		}
		if raw == nil {
			handleErr(fmt.Errorf("missing --data or --data-file"))
			return
		}
		body := asStringMap(raw)
		if body == nil {
			handleErr(fmt.Errorf("--data/--data-file must be a JSON object"))
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		base, _ := rwBase(app, wf, stage)
		path, err := c.AppstackPath(cmd.Context(), base)
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runJSONMutating(cmd.Context(), c, "appstack rw stage update", risk.HighRiskWrite, "PUT", path, nil, body, nil))
	},
}

var appstackRWSystemCmd = &cobra.Command{Use: "system", Short: "System-level release workflows"}

var appstackRWSystemListCmd = &cobra.Command{
	Use: "list", Short: "List system release workflows",
	Long: "Risk: read\nHTTP: GET .../systems/{system}/releaseWorkflows",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		sys, _ := cmd.Flags().GetString("system")
		if err := requireFlags("system", sys); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/systems/"+sys+"/releaseWorkflows")
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runRead(cmd.Context(), c, "GET", path, nil, nil, map[string]any{"risk": risk.Read}, nil))
	},
}

var appstackRWSystemCreateCmd = &cobra.Command{
	Use: "create", Short: "Create system release workflow (high-risk-write)",
	Long: "Risk: high-risk-write\nHTTP: POST .../systems/{system}/releaseWorkflows\nBody: workflow {name, note?, templateSn?}",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		sys, _ := cmd.Flags().GetString("system")
		name, _ := cmd.Flags().GetString("name")
		note, _ := cmd.Flags().GetString("note")
		tpl, _ := cmd.Flags().GetString("template-sn")
		dataStr, _ := cmd.Flags().GetString("data")
		dataFile, _ := cmd.Flags().GetString("data-file")
		if err := requireFlags("system", sys); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/systems/"+sys+"/releaseWorkflows")
		if err != nil {
			handleErr(err)
			return
		}
		var body map[string]any
		raw, err := loadJSONBodyFromFlags(dataStr, dataFile)
		if err != nil {
			handleErr(err)
			return
		}
		if raw != nil {
			body = asStringMap(raw)
			if body == nil {
				handleErr(fmt.Errorf("--data/--data-file must be a JSON object"))
				return
			}
		} else {
			if err := requireFlags("name", name); err != nil {
				handleErr(err)
				return
			}
			wf := map[string]any{"name": name}
			if note != "" {
				wf["note"] = note
			}
			if tpl != "" {
				wf["templateSn"] = tpl
			}
			body = map[string]any{"workflow": wf}
		}
		handleErr(runJSONMutating(cmd.Context(), c, "appstack rw system create", risk.HighRiskWrite, "POST", path, nil, body, nil))
	},
}

func init() {
	appstackRWListCmd.Flags().String("app", "", "app name (required)")
	appstackRWBriefsCmd.Flags().String("app", "", "app name (required)")
	for _, c := range []*cobra.Command{
		appstackRWStageGetCmd, appstackRWStageRunsCmd, appstackRWStageExecuteCmd,
		appstackRWStageCancelCmd, appstackRWStageRetryCmd, appstackRWStageSkipCmd,
		appstackRWStagePassCmd, appstackRWStageRefuseCmd, appstackRWStagePipelineRunCmd,
		appstackRWStageJobLogCmd, appstackRWStageMetaCmd, appstackRWStageUpdateCmd,
	} {
		c.Flags().String("app", "", "app name (required)")
		c.Flags().String("workflow-sn", "", "release workflow sn (required)")
		c.Flags().String("stage-sn", "", "release stage sn (required)")
	}
	appstackRWStageBriefsCmd.Flags().String("app", "", "app name (required)")
	appstackRWStageBriefsCmd.Flags().String("workflow-sn", "", "release workflow sn (required)")
	appstackRWStageRunsCmd.Flags().Int("page", 1, "page")
	appstackRWStageRunsCmd.Flags().Int("per-page", 20, "per page")
	appstackRWStageExecuteCmd.Flags().String("data", "", "execution JSON body (or @file)")
	appstackRWStageExecuteCmd.Flags().String("data-file", "", "read JSON body from file (alternative to --data)")
	appstackRWStageExecuteCmd.Flags().String("app-release-sn", "", "optional appReleaseSn")
	appstackRWStageUpdateCmd.Flags().String("data", "", "stage JSON body (required; or @file)")
	appstackRWStageUpdateCmd.Flags().String("data-file", "", "read JSON body from file (alternative to --data)")
	for _, c := range []*cobra.Command{
		appstackRWStageCancelCmd, appstackRWStageRetryCmd, appstackRWStageSkipCmd,
		appstackRWStagePassCmd, appstackRWStageRefuseCmd, appstackRWStagePipelineRunCmd,
		appstackRWStageJobLogCmd, appstackRWStageMetaCmd,
	} {
		c.Flags().String("execution", "", "execution number (required)")
	}
	for _, c := range []*cobra.Command{
		appstackRWStageRetryCmd, appstackRWStageSkipCmd, appstackRWStagePassCmd,
		appstackRWStageRefuseCmd, appstackRWStageJobLogCmd,
	} {
		c.Flags().String("job-id", "", "job id (required for this action)")
	}
	appstackRWSystemListCmd.Flags().String("system", "", "system name (required)")
	appstackRWSystemCreateCmd.Flags().String("system", "", "system name (required)")
	appstackRWSystemCreateCmd.Flags().String("name", "", "workflow name")
	appstackRWSystemCreateCmd.Flags().String("note", "", "note")
	appstackRWSystemCreateCmd.Flags().String("template-sn", "", "template sn")
	appstackRWSystemCreateCmd.Flags().String("data", "", "full JSON body override (or @file)")
	appstackRWSystemCreateCmd.Flags().String("data-file", "", "read JSON body from file (alternative to --data)")

	appstackRWStageCmd.AddCommand(
		appstackRWStageGetCmd, appstackRWStageBriefsCmd, appstackRWStageRunsCmd,
		appstackRWStageExecuteCmd, appstackRWStageCancelCmd, appstackRWStageRetryCmd,
		appstackRWStageSkipCmd, appstackRWStagePassCmd, appstackRWStageRefuseCmd,
		appstackRWStagePipelineRunCmd, appstackRWStageJobLogCmd, appstackRWStageMetaCmd,
		appstackRWStageUpdateCmd,
	)
	appstackRWSystemCmd.AddCommand(appstackRWSystemListCmd, appstackRWSystemCreateCmd)
	appstackRWCmd.AddCommand(appstackRWListCmd, appstackRWBriefsCmd, appstackRWStageCmd, appstackRWSystemCmd)
	appstackCmd.AddCommand(appstackRWCmd)
}
