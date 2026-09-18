package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/yunxiao-cli/yunxiao/internal/risk"
)

var appstackCmd = &cobra.Command{
	Use:   "appstack",
	Short: "AppStack applications and change orders",
	Long: `AppStack domain.

Typed:
  yunxiao appstack apps list
  yunxiao appstack apps get --name <appName>
  yunxiao appstack change-orders versions|get|by-origin|job-logs|create|execute-job
  yunxiao appstack orchestrations list|get
  yunxiao appstack tags search|create|update|delete|bind
  yunxiao appstack variable-groups list|get|revision|create|update|delete
  yunxiao appstack change-requests list|create|cancel|close|audit-items|work-items
  yunxiao appstack global-vars list|get
  yunxiao appstack apps create|update|sources

Risk: reads=read; tags/VG mutations + CO create/execute=high-risk-write`,
}

var appstackAppsCmd = &cobra.Command{Use: "apps", Short: "Applications"}

var appstackAppsListCmd = &cobra.Command{
	Use:   "list",
	Short: "Search / list applications",
	Long:  "Risk: read\nHTTP: GET .../appstack/.../apps:search",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		perPage, _ := cmd.Flags().GetInt("per-page")
		nextToken, _ := cmd.Flags().GetString("next-token")
		orderBy, _ := cmd.Flags().GetString("order-by")
		sort, _ := cmd.Flags().GetString("sort")
		tags, _ := cmd.Flags().GetString("tags")
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/apps:search")
		if err != nil {
			handleErr(err)
			return
		}
		q := appstackAppsListQuery(perPage, nextToken, orderBy, sort, tags)
		handleErr(runRead(cmd.Context(), c, "GET", path, q, nil, map[string]any{"risk": risk.Read}, nil))
	},
}

var appstackAppsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get application by name",
	Long:  "Risk: read\nHTTP: GET .../appstack/.../apps/{name}",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		name, _ := cmd.Flags().GetString("name")
		if err := requireFlags("name", name); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/apps/"+name)
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runRead(cmd.Context(), c, "GET", path, nil, nil, map[string]any{"risk": risk.Read}, nil))
	},
}

var appstackChangeOrdersCmd = &cobra.Command{Use: "change-orders", Aliases: []string{"co"}, Short: "Deploy / change orders"}

var appstackCOVersionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "List change order versions for an app",
	Long:  "Risk: read\nHTTP: GET .../apps/{app}/changeOrders/versions",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		app, _ := cmd.Flags().GetString("app")
		current, _ := cmd.Flags().GetInt("current")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		envNames, _ := cmd.Flags().GetString("env-names")
		creators, _ := cmd.Flags().GetString("creators")
		if err := requireFlags("app", app); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/apps/"+app+"/changeOrders/versions")
		if err != nil {
			handleErr(err)
			return
		}
		q := map[string]string{}
		if current > 0 {
			q["current"] = strconv.Itoa(current)
		}
		if pageSize > 0 {
			q["pageSize"] = strconv.Itoa(pageSize)
		}
		if envNames != "" {
			q["envNames"] = envNames
		}
		if creators != "" {
			q["creators"] = creators
		}
		handleErr(runRead(cmd.Context(), c, "GET", path, q, nil, map[string]any{"risk": risk.Read}, nil))
	},
}

var appstackCOGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a change order by serial number",
	Long:  "Risk: read\nHTTP: GET .../apps/{app}/changeOrders/{sn}",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		app, _ := cmd.Flags().GetString("app")
		sn, _ := cmd.Flags().GetString("sn")
		if err := requireFlags("app", app, "sn", sn); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/apps/"+app+"/changeOrders/"+sn)
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runRead(cmd.Context(), c, "GET", path, nil, nil, map[string]any{"risk": risk.Read}, nil))
	},
}

var appstackCOByOriginCmd = &cobra.Command{
	Use:   "by-origin",
	Short: "List change orders by origin",
	Long:  "Risk: read\nHTTP: GET .../changeOrders:byOrigin",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		originType, _ := cmd.Flags().GetString("origin-type")
		originID, _ := cmd.Flags().GetString("origin-id")
		app, _ := cmd.Flags().GetString("app")
		envName, _ := cmd.Flags().GetString("env-name")
		if err := requireFlags("origin-type", originType, "origin-id", originID); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/changeOrders:byOrigin")
		if err != nil {
			handleErr(err)
			return
		}
		q := map[string]string{"originType": originType, "originId": originID}
		if app != "" {
			q["appName"] = app
		}
		if envName != "" {
			q["envName"] = envName
		}
		handleErr(runRead(cmd.Context(), c, "GET", path, q, nil, map[string]any{"risk": risk.Read}, nil))
	},
}

var appstackCOCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a change order / deploy order (high-risk-write)",
	Long: `Risk: high-risk-write

HTTP: POST .../apps/{app}/changeOrders

Pass full body via --data JSON (or --data-file / --data @file.json), e.g.:
  {"changeOrderName":"deploy-1","type":"Deploy","envs":{"prod":{"values":{}}},"orchestrationRevisionSha":"..."}`,
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		app, _ := cmd.Flags().GetString("app")
		dataStr, _ := cmd.Flags().GetString("data")
		dataFile, _ := cmd.Flags().GetString("data-file")
		if err := requireFlags("app", app); err != nil {
			handleErr(err)
			return
		}
		body, err := loadJSONBodyFromFlags(dataStr, dataFile)
		if err != nil {
			handleErr(err)
			return
		}
		if body == nil {
			handleErr(fmt.Errorf("missing --data or --data-file"))
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/apps/"+app+"/changeOrders")
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runJSONMutating(cmd.Context(), c, "appstack change-orders create", risk.HighRiskWrite, "POST", path, nil, body, nil))
	},
}

var appstackCOExecuteJobCmd = &cobra.Command{
	Use:   "execute-job",
	Short: "Execute a job action on a change order (high-risk-write)",
	Long: `Risk: high-risk-write

HTTP: PUT .../changeOrders/{sn}/jobs/{jobSn}:execute
action-type: SUSPEND|RESUME|ROLLBACK|STOP`,
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		app, _ := cmd.Flags().GetString("app")
		sn, _ := cmd.Flags().GetString("sn")
		jobSn, _ := cmd.Flags().GetString("job-sn")
		actionType, _ := cmd.Flags().GetString("action-type")
		comment, _ := cmd.Flags().GetString("comment")
		if err := requireFlags("app", app, "sn", sn, "job-sn", jobSn, "action-type", actionType); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/apps/"+app+"/changeOrders/"+sn+"/jobs/"+jobSn+":execute")
		if err != nil {
			handleErr(err)
			return
		}
		body := map[string]any{"actionType": actionType}
		if comment != "" {
			body["context"] = map[string]any{"comment": comment}
		}
		handleErr(runJSONMutating(cmd.Context(), c, "appstack change-orders execute-job", risk.HighRiskWrite, "PUT", path, nil, body, nil))
	},
}

var appstackCOJobLogsCmd = &cobra.Command{
	Use:   "job-logs",
	Short: "List change-order job logs",
	Long:  "Risk: read\nHTTP: GET .../changeOrders/{sn}/jobs/{jobSn}/logs",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		app, _ := cmd.Flags().GetString("app")
		sn, _ := cmd.Flags().GetString("sn")
		jobSn, _ := cmd.Flags().GetString("job-sn")
		current, _ := cmd.Flags().GetInt("current")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		if err := requireFlags("app", app, "sn", sn, "job-sn", jobSn); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/apps/"+app+"/changeOrders/"+sn+"/jobs/"+jobSn+"/logs")
		if err != nil {
			handleErr(err)
			return
		}
		q := map[string]string{}
		if current > 0 {
			q["current"] = strconv.Itoa(current)
		}
		if pageSize > 0 {
			q["pageSize"] = strconv.Itoa(pageSize)
		}
		handleErr(runRead(cmd.Context(), c, "GET", path, q, nil, map[string]any{"risk": risk.Read}, nil))
	},
}

var appstackOrchestrationsCmd = &cobra.Command{Use: "orchestrations", Short: "Application orchestrations"}

var appstackOrchestrationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List application orchestrations",
	Long:  "Risk: read\nHTTP: GET .../apps/{app}/orchestrations",
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
		path, err := c.AppstackPath(cmd.Context(), "/apps/"+app+"/orchestrations")
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runRead(cmd.Context(), c, "GET", path, nil, nil, map[string]any{"risk": risk.Read}, nil))
	},
}

var appstackOrchestrationsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an application orchestration",
	Long:  "Risk: read\nHTTP: GET .../apps/{app}/orchestrations/{sn}",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		app, _ := cmd.Flags().GetString("app")
		sn, _ := cmd.Flags().GetString("sn")
		tagName, _ := cmd.Flags().GetString("tag-name")
		sha, _ := cmd.Flags().GetString("sha")
		if err := requireFlags("app", app, "sn", sn); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/apps/"+app+"/orchestrations/"+sn)
		if err != nil {
			handleErr(err)
			return
		}
		q := map[string]string{}
		if tagName != "" {
			q["tagName"] = tagName
		}
		if sha != "" {
			q["sha"] = sha
		}
		handleErr(runRead(cmd.Context(), c, "GET", path, q, nil, map[string]any{"risk": risk.Read}, nil))
	},
}

var appstackTagsCmd = &cobra.Command{Use: "tags", Short: "Application tags (AppStack)"}

var appstackTagsSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search application tags",
	Long:  "Risk: read\nHTTP: POST .../appTags:search\nSource: operations/appstack/appTags.ts searchAppTag",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		current, _ := cmd.Flags().GetInt("current")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		search, _ := cmd.Flags().GetString("search")
		orderBy, _ := cmd.Flags().GetString("order-by")
		sortOrder, _ := cmd.Flags().GetString("sort")
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/appTags:search")
		if err != nil {
			handleErr(err)
			return
		}
		q := map[string]string{}
		if current > 0 {
			q["current"] = strconv.Itoa(current)
		}
		if pageSize > 0 {
			q["pageSize"] = strconv.Itoa(pageSize)
		}
		body := map[string]any{}
		if search != "" {
			body["search"] = search
		}
		if orderBy != "" {
			body["orderBy"] = orderBy
		}
		if sortOrder != "" {
			body["sort"] = sortOrder
		}
		handleErr(runRead(cmd.Context(), c, "POST", path, q, body, map[string]any{"risk": risk.Read}, nil))
	},
}

var appstackTagsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an application tag (high-risk-write)",
	Long:  "Risk: high-risk-write\nHTTP: POST .../appTags\nSource: operations/appstack/appTags.ts createAppTag",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		name, _ := cmd.Flags().GetString("name")
		color, _ := cmd.Flags().GetString("color")
		if err := requireFlags("name", name, "color", color); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/appTags")
		if err != nil {
			handleErr(err)
			return
		}
		body := map[string]any{"name": name, "color": color}
		handleErr(runJSONMutating(cmd.Context(), c, "appstack tags create", risk.HighRiskWrite, "POST", path, nil, body, nil))
	},
}

var appstackTagsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an application tag (high-risk-write)",
	Long:  "Risk: high-risk-write\nHTTP: PUT .../appTags/updateTag?name=\nSource: operations/appstack/appTags.ts updateAppTag",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		name, _ := cmd.Flags().GetString("name")
		newName, _ := cmd.Flags().GetString("new-name")
		color, _ := cmd.Flags().GetString("color")
		if err := requireFlags("name", name, "new-name", newName); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/appTags/updateTag")
		if err != nil {
			handleErr(err)
			return
		}
		q := map[string]string{"name": name}
		body := map[string]any{"newName": newName}
		if color != "" {
			body["color"] = color
		}
		handleErr(runJSONMutating(cmd.Context(), c, "appstack tags update", risk.HighRiskWrite, "PUT", path, q, body, nil))
	},
}

var appstackTagsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an application tag (high-risk-write)",
	Long:  "Risk: high-risk-write\nHTTP: DELETE .../appTags/deleteTag?name=\nSource: operations/appstack/appTags.ts deleteAppTag",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		name, _ := cmd.Flags().GetString("name")
		if err := requireFlags("name", name); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/appTags/deleteTag")
		if err != nil {
			handleErr(err)
			return
		}
		q := map[string]string{"name": name}
		handleErr(runJSONMutating(cmd.Context(), c, "appstack tags delete", risk.HighRiskWrite, "DELETE", path, q, nil, nil))
	},
}

var appstackTagsBindCmd = &cobra.Command{
	Use:   "bind",
	Short: "Set tag bindings on an app (high-risk-write; replaces list)",
	Long:  "Risk: high-risk-write\nHTTP: PUT .../apps/{app}/appTags\nBody: {tagNames:[...]}\nEmpty --tag-names clears bindings.\nSource: operations/appstack/appTags.ts updateAppTagBind",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		app, _ := cmd.Flags().GetString("app")
		tagNames, _ := cmd.Flags().GetString("tag-names")
		if err := requireFlags("app", app); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/apps/"+app+"/appTags")
		if err != nil {
			handleErr(err)
			return
		}
		names := splitCSV(tagNames)
		if names == nil {
			names = []string{}
		}
		body := map[string]any{"tagNames": names}
		handleErr(runJSONMutating(cmd.Context(), c, "appstack tags bind", risk.HighRiskWrite, "PUT", path, nil, body, nil))
	},
}

var appstackVGCmd = &cobra.Command{Use: "variable-groups", Aliases: []string{"vg"}, Short: "AppStack variable groups"}

var appstackVGListCmd = &cobra.Command{
	Use:   "list",
	Short: "List variable groups for an app",
	Long:  "Risk: read\nHTTP: GET .../apps/{app}/variableGroups\nSource: operations/appstack/variableGroups.ts getAppVariableGroups",
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
		path, err := c.AppstackPath(cmd.Context(), "/apps/"+app+"/variableGroups")
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runRead(cmd.Context(), c, "GET", path, nil, nil, map[string]any{"risk": risk.Read}, nil))
	},
}

var appstackVGRevisionCmd = &cobra.Command{
	Use:   "revision",
	Short: "Get variable-groups revision for an app",
	Long:  "Risk: read\nHTTP: GET .../apps/{app}/variableGroups:revision\nSource: getAppVariableGroupsRevision",
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
		path, err := c.AppstackPath(cmd.Context(), "/apps/"+app+"/variableGroups:revision")
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runRead(cmd.Context(), c, "GET", path, nil, nil, map[string]any{"risk": risk.Read}, nil))
	},
}

var appstackVGGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a variable group",
	Long:  "Risk: read\nHTTP: GET .../apps/{app}/variableGroup/{name}",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		app, _ := cmd.Flags().GetString("app")
		name, _ := cmd.Flags().GetString("name")
		if err := requireFlags("app", app, "name", name); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/apps/"+app+"/variableGroup/"+name)
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runRead(cmd.Context(), c, "GET", path, nil, nil, map[string]any{"risk": risk.Read}, nil))
	},
}

var appstackVGCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a variable group (high-risk-write)",
	Long: `Risk: high-risk-write
HTTP: POST .../apps/{app}/variableGroup
Source: operations/appstack/variableGroups.ts createVariableGroup
--vars is JSON array: [{"key":"K","value":"V","description":"..."}]
--from-revision-sha is required by OpenAPI (use variable-groups revision).`,
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		app, _ := cmd.Flags().GetString("app")
		name, _ := cmd.Flags().GetString("name")
		fromSha, _ := cmd.Flags().GetString("from-revision-sha")
		displayName, _ := cmd.Flags().GetString("display-name")
		message, _ := cmd.Flags().GetString("message")
		branch, _ := cmd.Flags().GetString("branch-name")
		varsJSON, _ := cmd.Flags().GetString("vars")
		if err := requireFlags("app", app, "from-revision-sha", fromSha); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/apps/"+app+"/variableGroup")
		if err != nil {
			handleErr(err)
			return
		}
		body := map[string]any{"fromRevisionSha": fromSha}
		if name != "" {
			body["name"] = name
		}
		if displayName != "" {
			body["displayName"] = displayName
		}
		if message != "" {
			body["message"] = message
		}
		if branch != "" {
			body["branchName"] = branch
		}
		if varsJSON != "" {
			var vars any
			if err := json.Unmarshal([]byte(varsJSON), &vars); err != nil {
				handleErr(fmt.Errorf("invalid --vars JSON: %w", err))
				return
			}
			body["vars"] = vars
		}
		handleErr(runJSONMutating(cmd.Context(), c, "appstack variable-groups create", risk.HighRiskWrite, "POST", path, nil, body, nil))
	},
}

var appstackVGUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a variable group (high-risk-write)",
	Long:  "Risk: high-risk-write\nHTTP: PUT .../apps/{app}/variableGroup/{name}\nSource: updateVariableGroup",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		app, _ := cmd.Flags().GetString("app")
		name, _ := cmd.Flags().GetString("name")
		fromSha, _ := cmd.Flags().GetString("from-revision-sha")
		displayName, _ := cmd.Flags().GetString("display-name")
		message, _ := cmd.Flags().GetString("message")
		branch, _ := cmd.Flags().GetString("branch-name")
		newName, _ := cmd.Flags().GetString("new-name")
		varsJSON, _ := cmd.Flags().GetString("vars")
		if err := requireFlags("app", app, "name", name, "from-revision-sha", fromSha); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/apps/"+app+"/variableGroup/"+name)
		if err != nil {
			handleErr(err)
			return
		}
		body := map[string]any{"fromRevisionSha": fromSha}
		if displayName != "" {
			body["displayName"] = displayName
		}
		if message != "" {
			body["message"] = message
		}
		if branch != "" {
			body["branchName"] = branch
		}
		if newName != "" {
			body["name"] = newName
		}
		if varsJSON != "" {
			var vars any
			if err := json.Unmarshal([]byte(varsJSON), &vars); err != nil {
				handleErr(fmt.Errorf("invalid --vars JSON: %w", err))
				return
			}
			body["vars"] = vars
		}
		handleErr(runJSONMutating(cmd.Context(), c, "appstack variable-groups update", risk.HighRiskWrite, "PUT", path, nil, body, nil))
	},
}

var appstackVGDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a variable group (high-risk-write)",
	Long:  "Risk: high-risk-write\nHTTP: DELETE .../apps/{app}/variableGroup/{name}",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		app, _ := cmd.Flags().GetString("app")
		name, _ := cmd.Flags().GetString("name")
		if err := requireFlags("app", app, "name", name); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/apps/"+app+"/variableGroup/"+name)
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runJSONMutating(cmd.Context(), c, "appstack variable-groups delete", risk.HighRiskWrite, "DELETE", path, nil, nil, nil))
	},
}

var appstackAppsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an application (high-risk-write)",
	Long:  "Risk: high-risk-write\nHTTP: POST .../appstack/.../apps\nSource: createApplication",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		name, _ := cmd.Flags().GetString("name")
		tpl, _ := cmd.Flags().GetString("template")
		desc, _ := cmd.Flags().GetString("description")
		owner, _ := cmd.Flags().GetString("owner-id")
		tags, _ := cmd.Flags().GetString("tags")
		if err := requireFlags("name", name); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/apps")
		if err != nil {
			handleErr(err)
			return
		}
		body := map[string]any{"name": name}
		if tpl != "" {
			body["appTemplateName"] = tpl
		}
		if desc != "" {
			body["description"] = desc
		}
		if owner != "" {
			body["ownerId"] = owner
		}
		if tags != "" {
			body["tags"] = splitCSV(tags)
		}
		handleErr(runJSONMutating(cmd.Context(), c, "appstack apps create", risk.HighRiskWrite, "POST", path, nil, body, nil))
	},
}

var appstackAppsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an application (high-risk-write)",
	Long:  "Risk: high-risk-write\nHTTP: PUT .../apps/{name}\nSource: updateApplication",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		name, _ := cmd.Flags().GetString("name")
		desc, _ := cmd.Flags().GetString("description")
		owner, _ := cmd.Flags().GetString("owner-id")
		dataStr, _ := cmd.Flags().GetString("data")
		dataFile, _ := cmd.Flags().GetString("data-file")
		if err := requireFlags("name", name); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/apps/"+name)
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
		if desc != "" {
			body["description"] = desc
		}
		if owner != "" {
			body["ownerId"] = owner
		}
		if len(body) == 0 {
			handleErr(fmt.Errorf("provide --description/--owner-id and/or --data/--data-file JSON"))
			return
		}
		handleErr(runJSONMutating(cmd.Context(), c, "appstack apps update", risk.HighRiskWrite, "PUT", path, nil, body, nil))
	},
}

var appstackAppsSourcesCmd = &cobra.Command{
	Use:   "sources",
	Short: "List application code sources",
	Long:  "Risk: read\nHTTP: GET .../apps/{name}/sources\nSource: listApplicationSources",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		name, _ := cmd.Flags().GetString("name")
		if err := requireFlags("name", name); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/apps/"+name+"/sources")
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runRead(cmd.Context(), c, "GET", path, nil, nil, map[string]any{"risk": risk.Read}, nil))
	},
}

var appstackCRCmd = &cobra.Command{Use: "change-requests", Aliases: []string{"cr"}, Short: "AppStack change requests"}

var appstackCRListCmd = &cobra.Command{
	Use:   "list",
	Short: "Search/list change requests for an app",
	Long:  "Risk: read\nHTTP: POST .../apps/{app}/changeRequests:search\nSource: listAppChangeRequests",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		app, _ := cmd.Flags().GetString("app")
		name, _ := cmd.Flags().GetString("name")
		state, _ := cmd.Flags().GetString("state")
		current, _ := cmd.Flags().GetInt("current")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		if err := requireFlags("app", app); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/apps/"+app+"/changeRequests:search")
		if err != nil {
			handleErr(err)
			return
		}
		body := map[string]any{}
		if name != "" {
			body["name"] = name
		}
		if state != "" {
			body["state"] = splitCSV(state)
		}
		if current > 0 {
			body["current"] = current
		}
		if pageSize > 0 {
			body["pageSize"] = pageSize
		}
		handleErr(runRead(cmd.Context(), c, "POST", path, nil, body, map[string]any{"risk": risk.Read}, nil))
	},
}

var appstackCRCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a change request (high-risk-write)",
	Long:  "Risk: high-risk-write\nHTTP: POST .../apps/{app}/changeRequests\nOr pass full body via --data JSON.",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		app, _ := cmd.Flags().GetString("app")
		dataStr, _ := cmd.Flags().GetString("data")
		dataFile, _ := cmd.Flags().GetString("data-file")
		if err := requireFlags("app", app); err != nil {
			handleErr(err)
			return
		}
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
		path, err := c.AppstackPath(cmd.Context(), "/apps/"+app+"/changeRequests")
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runJSONMutating(cmd.Context(), c, "appstack change-requests create", risk.HighRiskWrite, "POST", path, nil, body, nil))
	},
}

var appstackCRCancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "Cancel a change request (high-risk-write)",
	Long:  "Risk: high-risk-write\nHTTP: POST .../changeRequests/{sn}:cancel",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		app, _ := cmd.Flags().GetString("app")
		sn, _ := cmd.Flags().GetString("sn")
		if err := requireFlags("app", app, "sn", sn); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/apps/"+app+"/changeRequests/"+sn+":cancel")
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runJSONMutating(cmd.Context(), c, "appstack change-requests cancel", risk.HighRiskWrite, "POST", path, nil, nil, nil))
	},
}

var appstackCRCloseCmd = &cobra.Command{
	Use:   "close",
	Short: "Finish/close a change request (high-risk-write)",
	Long:  "Risk: high-risk-write\nHTTP: POST .../changeRequests/{sn}:finish",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		app, _ := cmd.Flags().GetString("app")
		sn, _ := cmd.Flags().GetString("sn")
		if err := requireFlags("app", app, "sn", sn); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/apps/"+app+"/changeRequests/"+sn+":finish")
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runJSONMutating(cmd.Context(), c, "appstack change-requests close", risk.HighRiskWrite, "POST", path, nil, nil, nil))
	},
}

var appstackCRAuditCmd = &cobra.Command{
	Use:   "audit-items",
	Short: "List change-request audit items",
	Long:  "Risk: read\nHTTP: GET .../changeRequests/{sn}/auditItems?refType=",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		app, _ := cmd.Flags().GetString("app")
		sn, _ := cmd.Flags().GetString("sn")
		refType, _ := cmd.Flags().GetString("ref-type")
		if err := requireFlags("app", app, "sn", sn, "ref-type", refType); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/apps/"+app+"/changeRequests/"+sn+"/auditItems")
		if err != nil {
			handleErr(err)
			return
		}
		q := map[string]string{"refType": refType}
		handleErr(runRead(cmd.Context(), c, "GET", path, q, nil, map[string]any{"risk": risk.Read}, nil))
	},
}

var appstackCRWorkItemsCmd = &cobra.Command{
	Use:   "work-items",
	Short: "List work items on a change request",
	Long:  "Risk: read\nHTTP: GET .../changeRequests/{sn}/workItems",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		app, _ := cmd.Flags().GetString("app")
		sn, _ := cmd.Flags().GetString("sn")
		if err := requireFlags("app", app, "sn", sn); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/apps/"+app+"/changeRequests/"+sn+"/workItems")
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runRead(cmd.Context(), c, "GET", path, nil, nil, map[string]any{"risk": risk.Read}, nil))
	},
}

var appstackGVCmd = &cobra.Command{Use: "global-vars", Short: "AppStack global variables"}

var appstackGVListCmd = &cobra.Command{
	Use:   "list",
	Short: "Search/list global vars",
	Long:  "Risk: read\nHTTP: POST .../globalVars:search\nSource: listGlobalVars",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		dataStr, _ := cmd.Flags().GetString("data")
		dataFile, _ := cmd.Flags().GetString("data-file")
		current, _ := cmd.Flags().GetInt("current")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		search, _ := cmd.Flags().GetString("search")
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/globalVars:search")
		if err != nil {
			handleErr(err)
			return
		}
		q := map[string]string{"current": strconv.Itoa(current), "pageSize": strconv.Itoa(pageSize)}
		body := map[string]any{}
		if search != "" {
			body["search"] = search
		}
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
		handleErr(runRead(cmd.Context(), c, "POST", path, q, body, map[string]any{"risk": risk.Read}, nil))
	},
}

var appstackGVGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a global var",
	Long:  "Risk: read\nHTTP: GET .../globalVars/{name}",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		name, _ := cmd.Flags().GetString("name")
		if err := requireFlags("name", name); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.AppstackPath(cmd.Context(), "/globalVars/"+name)
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runRead(cmd.Context(), c, "GET", path, nil, nil, map[string]any{"risk": risk.Read}, nil))
	},
}


// appstackAppsListQuery builds query for GET .../apps:search.
// Yunxiao requires pagination=keyset (keyset pagination); orderBy defaults to id.
func appstackAppsListQuery(perPage int, nextToken, orderBy, sort, tags string) map[string]string {
	q := map[string]string{"pagination": "keyset"}
	if perPage > 0 {
		q["perPage"] = strconv.Itoa(perPage)
	}
	if nextToken != "" {
		q["nextToken"] = nextToken
	}
	if orderBy == "" {
		orderBy = "id"
	}
	q["orderBy"] = orderBy
	if sort != "" {
		q["sort"] = sort
	}
	if tags != "" {
		q["tags"] = tags
	}
	return q
}

func init() {
	appstackAppsListCmd.Flags().Int("per-page", 20, "page size")
	appstackAppsListCmd.Flags().String("next-token", "", "pagination token")
	appstackAppsListCmd.Flags().String("order-by", "id", "orderBy (API required; default id)")
	appstackAppsListCmd.Flags().String("sort", "", "asc|desc")
	appstackAppsListCmd.Flags().String("tags", "", "comma-separated tags")
	appstackAppsGetCmd.Flags().String("name", "", "application name (required)")
	appstackCOVersionsCmd.Flags().String("app", "", "application name (required)")
	appstackCOVersionsCmd.Flags().Int("current", 1, "page current")
	appstackCOVersionsCmd.Flags().Int("page-size", 20, "page size")
	appstackCOVersionsCmd.Flags().String("env-names", "", "comma-separated env names")
	appstackCOVersionsCmd.Flags().String("creators", "", "comma-separated creators")
	appstackCOGetCmd.Flags().String("app", "", "application name (required)")
	appstackCOGetCmd.Flags().String("sn", "", "change order serial number (required)")
	appstackCOByOriginCmd.Flags().String("origin-type", "", "origin type (required)")
	appstackCOByOriginCmd.Flags().String("origin-id", "", "origin id (required)")
	appstackCOByOriginCmd.Flags().String("app", "", "optional app name filter")
	appstackCOByOriginCmd.Flags().String("env-name", "", "optional env name filter")
	appstackCOCreateCmd.Flags().String("app", "", "application name (required)")
	appstackCOCreateCmd.Flags().String("data", "", "JSON change order body (required; or @file)")
	appstackCOCreateCmd.Flags().String("data-file", "", "read JSON body from file (alternative to --data)")
	appstackCOExecuteJobCmd.Flags().String("app", "", "application name (required)")
	appstackCOExecuteJobCmd.Flags().String("sn", "", "change order sn (required)")
	appstackCOExecuteJobCmd.Flags().String("job-sn", "", "job sn (required)")
	appstackCOExecuteJobCmd.Flags().String("action-type", "", "SUSPEND|RESUME|ROLLBACK|STOP (required)")
	appstackCOExecuteJobCmd.Flags().String("comment", "", "optional action comment")
	appstackCOJobLogsCmd.Flags().String("app", "", "app name (required)")
	appstackCOJobLogsCmd.Flags().String("sn", "", "change order sn (required)")
	appstackCOJobLogsCmd.Flags().String("job-sn", "", "job sn (required)")
	appstackCOJobLogsCmd.Flags().Int("current", 1, "page current")
	appstackCOJobLogsCmd.Flags().Int("page-size", 20, "page size")
	appstackOrchestrationsListCmd.Flags().String("app", "", "app name (required)")
	appstackOrchestrationsGetCmd.Flags().String("app", "", "app name (required)")
	appstackOrchestrationsGetCmd.Flags().String("sn", "", "orchestration sn (required)")
	appstackOrchestrationsGetCmd.Flags().String("tag-name", "", "optional tagName")
	appstackOrchestrationsGetCmd.Flags().String("sha", "", "optional sha")
	appstackTagsSearchCmd.Flags().Int("current", 1, "page current")
	appstackTagsSearchCmd.Flags().Int("page-size", 10, "page size")
	appstackTagsSearchCmd.Flags().String("search", "", "fuzzy tag name")
	appstackTagsSearchCmd.Flags().String("order-by", "", "tagName|id")
	appstackTagsSearchCmd.Flags().String("sort", "", "asc|desc")
	appstackTagsCreateCmd.Flags().String("name", "", "tag name (required)")
	appstackTagsCreateCmd.Flags().String("color", "", "color hex e.g. #4676e5 (required)")
	appstackTagsUpdateCmd.Flags().String("name", "", "current tag name (required)")
	appstackTagsUpdateCmd.Flags().String("new-name", "", "new tag name (required; same as name if unchanged)")
	appstackTagsUpdateCmd.Flags().String("color", "", "optional new color")
	appstackTagsDeleteCmd.Flags().String("name", "", "tag name (required)")
	appstackTagsBindCmd.Flags().String("app", "", "app name (required)")
	appstackTagsBindCmd.Flags().String("tag-names", "", "comma-separated tag names (empty clears)")
	appstackVGListCmd.Flags().String("app", "", "app name (required)")
	appstackVGRevisionCmd.Flags().String("app", "", "app name (required)")
	appstackVGGetCmd.Flags().String("app", "", "app name (required)")
	appstackVGGetCmd.Flags().String("name", "", "variable group name (required)")
	appstackVGCreateCmd.Flags().String("app", "", "app name (required)")
	appstackVGCreateCmd.Flags().String("name", "", "variable group unique name")
	appstackVGCreateCmd.Flags().String("from-revision-sha", "", "base revision sha (required)")
	appstackVGCreateCmd.Flags().String("display-name", "", "display name")
	appstackVGCreateCmd.Flags().String("message", "", "commit message")
	appstackVGCreateCmd.Flags().String("branch-name", "", "branch, default master")
	appstackVGCreateCmd.Flags().String("vars", "", "JSON array of {key,value,description}")
	appstackVGUpdateCmd.Flags().String("app", "", "app name (required)")
	appstackVGUpdateCmd.Flags().String("name", "", "variable group name (required)")
	appstackVGUpdateCmd.Flags().String("from-revision-sha", "", "base revision sha (required)")
	appstackVGUpdateCmd.Flags().String("new-name", "", "optional rename")
	appstackVGUpdateCmd.Flags().String("display-name", "", "display name")
	appstackVGUpdateCmd.Flags().String("message", "", "commit message")
	appstackVGUpdateCmd.Flags().String("branch-name", "", "branch")
	appstackVGUpdateCmd.Flags().String("vars", "", "JSON array of vars")
	appstackVGDeleteCmd.Flags().String("app", "", "app name (required)")
	appstackVGDeleteCmd.Flags().String("name", "", "variable group name (required)")
	appstackChangeOrdersCmd.AddCommand(appstackCOVersionsCmd, appstackCOGetCmd, appstackCOByOriginCmd, appstackCOJobLogsCmd, appstackCOCreateCmd, appstackCOExecuteJobCmd)
	appstackOrchestrationsCmd.AddCommand(appstackOrchestrationsListCmd, appstackOrchestrationsGetCmd)
	appstackTagsCmd.AddCommand(appstackTagsSearchCmd, appstackTagsCreateCmd, appstackTagsUpdateCmd, appstackTagsDeleteCmd, appstackTagsBindCmd)
	appstackVGCmd.AddCommand(appstackVGListCmd, appstackVGGetCmd, appstackVGRevisionCmd, appstackVGCreateCmd, appstackVGUpdateCmd, appstackVGDeleteCmd)
	appstackAppsCreateCmd.Flags().String("name", "", "app name (required)")
	appstackAppsCreateCmd.Flags().String("template", "", "appTemplateName")
	appstackAppsCreateCmd.Flags().String("description", "", "description")
	appstackAppsCreateCmd.Flags().String("owner-id", "", "owner id")
	appstackAppsCreateCmd.Flags().String("tags", "", "comma-separated tags")
	appstackAppsUpdateCmd.Flags().String("name", "", "app name (required)")
	appstackAppsUpdateCmd.Flags().String("description", "", "description")
	appstackAppsUpdateCmd.Flags().String("owner-id", "", "owner id")
	appstackAppsUpdateCmd.Flags().String("data", "", "extra JSON body fields (or @file)")
	appstackAppsUpdateCmd.Flags().String("data-file", "", "read extra JSON from file (alternative to --data)")
	appstackAppsSourcesCmd.Flags().String("name", "", "app name (required)")
	appstackCRListCmd.Flags().String("app", "", "app name (required)")
	appstackCRListCmd.Flags().String("name", "", "filter by CR name")
	appstackCRListCmd.Flags().String("state", "", "comma states DEVELOPING|INTEGRATING|RELEASED|CLOSED")
	appstackCRListCmd.Flags().Int("current", 1, "page")
	appstackCRListCmd.Flags().Int("page-size", 10, "page size")
	appstackCRCreateCmd.Flags().String("app", "", "app name (required)")
	appstackCRCreateCmd.Flags().String("data", "", "JSON body (required; or @file)")
	appstackCRCreateCmd.Flags().String("data-file", "", "read JSON body from file (alternative to --data)")
	appstackCRCancelCmd.Flags().String("app", "", "app (required)")
	appstackCRCancelCmd.Flags().String("sn", "", "CR sn (required)")
	appstackCRCloseCmd.Flags().String("app", "", "app (required)")
	appstackCRCloseCmd.Flags().String("sn", "", "CR sn (required)")
	appstackCRAuditCmd.Flags().String("app", "", "app (required)")
	appstackCRAuditCmd.Flags().String("sn", "", "CR sn (required)")
	appstackCRAuditCmd.Flags().String("ref-type", "", "refType (required)")
	appstackCRWorkItemsCmd.Flags().String("app", "", "app (required)")
	appstackCRWorkItemsCmd.Flags().String("sn", "", "CR sn (required)")
	appstackGVListCmd.Flags().Int("current", 1, "page current (required by API)")
	appstackGVListCmd.Flags().Int("page-size", 20, "page size (required by API)")
	appstackGVListCmd.Flags().String("search", "", "search keyword")
	appstackGVListCmd.Flags().String("data", "", "optional extra JSON merged into body (or @file)")
	appstackGVListCmd.Flags().String("data-file", "", "read extra JSON from file (alternative to --data)")
	appstackGVGetCmd.Flags().String("name", "", "var name (required)")
	appstackAppsCmd.AddCommand(appstackAppsListCmd, appstackAppsGetCmd, appstackAppsCreateCmd, appstackAppsUpdateCmd, appstackAppsSourcesCmd)
	appstackCRCmd.AddCommand(appstackCRListCmd, appstackCRCreateCmd, appstackCRCancelCmd, appstackCRCloseCmd, appstackCRAuditCmd, appstackCRWorkItemsCmd)
	appstackGVCmd.AddCommand(appstackGVListCmd, appstackGVGetCmd)
	appstackCmd.AddCommand(appstackAppsCmd, appstackChangeOrdersCmd, appstackOrchestrationsCmd, appstackTagsCmd, appstackVGCmd, appstackCRCmd, appstackGVCmd)
}
