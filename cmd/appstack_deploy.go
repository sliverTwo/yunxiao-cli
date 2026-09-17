package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/yunxiao-cli/yunxiao/internal/risk"
)

var appstackDeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "AppStack host / deploy-group mutations",
	Long: `Deployment resources (operations/appstack/deploymentResources.ts).

  yunxiao appstack deploy machine-log --tunnel-id <n> --machine-sn <sn>
  yunxiao appstack deploy add-hosts|remove-hosts --instance <name> --host-sns a,b
  yunxiao appstack deploy add-hosts-to-group|remove-hosts-from-group --instance <n> --group <g> --host-sns …

Risk: machine-log=read; host list mutations=high-risk-write`,
}

var appstackDeployMachineLogCmd = &cobra.Command{
	Use:   "machine-log",
	Short: "Get machine deploy log",
	Long:  "Risk: read\nHTTP: GET .../host/deployLog?tunnelId=&machineSn=",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		tunnelID, _ := cmd.Flags().GetInt("tunnel-id")
		machineSn, _ := cmd.Flags().GetString("machine-sn")
		if tunnelID <= 0 {
			handleErr(fmt.Errorf("missing required flag --tunnel-id"))
			return
		}
		if err := requireFlags("machine-sn", machineSn); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/host/deployLog")
		if err != nil {
			handleErr(err)
			return
		}
		q := map[string]string{"tunnelId": strconv.Itoa(tunnelID), "machineSn": machineSn}
		handleErr(runRead(cmd.Context(), c, "GET", path, q, nil, map[string]any{"risk": risk.Read}, nil))
	},
}

func deployHostMutation(action, pathSuffix string, needGroup bool) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		instance, _ := cmd.Flags().GetString("instance")
		group, _ := cmd.Flags().GetString("group")
		hostSns, _ := cmd.Flags().GetString("host-sns")
		if err := requireFlags("instance", instance, "host-sns", hostSns); err != nil {
			handleErr(err)
			return
		}
		if needGroup {
			if err := requireFlags("group", group); err != nil {
				handleErr(err)
				return
			}
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		suffix := "/pools/instances/" + instance + pathSuffix
		if needGroup {
			suffix = "/pools/instances/" + instance + "/deployGroup/" + group + pathSuffix
		}
		path, err := c.AppstackPath(cmd.Context(), suffix)
		if err != nil {
			handleErr(err)
			return
		}
		body := map[string]any{"hostSns": splitCSV(hostSns)}
		handleErr(runJSONMutating(cmd.Context(), c, action, risk.HighRiskWrite, "PUT", path, nil, body, nil))
	}
}

var appstackDeployAddHostsCmd = &cobra.Command{
	Use: "add-hosts", Short: "Add hosts to host group (high-risk-write)",
	Long: "Risk: high-risk-write\nHTTP: PUT .../pools/instances/{instance}/addHostList",
	Run:  deployHostMutation("appstack deploy add-hosts", "/addHostList", false),
}
var appstackDeployRemoveHostsCmd = &cobra.Command{
	Use: "remove-hosts", Short: "Remove hosts from host group (high-risk-write)",
	Long: "Risk: high-risk-write\nHTTP: PUT .../pools/instances/{instance}/removeHostList",
	Run:  deployHostMutation("appstack deploy remove-hosts", "/removeHostList", false),
}
var appstackDeployAddHostsGroupCmd = &cobra.Command{
	Use: "add-hosts-to-group", Short: "Add hosts to deploy group (high-risk-write)",
	Long: "Risk: high-risk-write\nHTTP: PUT .../deployGroup/{group}/addHostList",
	Run:  deployHostMutation("appstack deploy add-hosts-to-group", "/addHostList", true),
}
var appstackDeployRemoveHostsGroupCmd = &cobra.Command{
	Use: "remove-hosts-from-group", Short: "Remove hosts from deploy group (high-risk-write)",
	Long: "Risk: high-risk-write\nHTTP: PUT .../deployGroup/{group}/removeHostList",
	Run:  deployHostMutation("appstack deploy remove-hosts-from-group", "/removeHostList", true),
}

func init() {
	appstackDeployMachineLogCmd.Flags().Int("tunnel-id", 0, "tunnel id (required)")
	appstackDeployMachineLogCmd.Flags().String("machine-sn", "", "machine sn (required)")
	for _, c := range []*cobra.Command{appstackDeployAddHostsCmd, appstackDeployRemoveHostsCmd, appstackDeployAddHostsGroupCmd, appstackDeployRemoveHostsGroupCmd} {
		c.Flags().String("instance", "", "host cluster instance name (required)")
		c.Flags().String("host-sns", "", "comma-separated ECS host sns (required)")
	}
	appstackDeployAddHostsGroupCmd.Flags().String("group", "", "deploy group name (required)")
	appstackDeployRemoveHostsGroupCmd.Flags().String("group", "", "deploy group name (required)")
	appstackDeployCmd.AddCommand(
		appstackDeployMachineLogCmd, appstackDeployAddHostsCmd, appstackDeployRemoveHostsCmd,
		appstackDeployAddHostsGroupCmd, appstackDeployRemoveHostsGroupCmd,
	)
	appstackCmd.AddCommand(appstackDeployCmd)
}
