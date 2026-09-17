package cmd

// Yunxiao PAT console (one-click primary) and help doc (secondary).
// Console UI uses coarse per-module checkboxes (读/写), not Feishu/OAuth scopes.
const (
	yunxiaoPATConsoleURL = "https://account-devops.aliyun.com/settings/personalAccessToken"
	yunxiaoPATHelpURL    = "https://help.aliyun.com/zh/yunxiao/user-guide/personal-access-token"
)

// patHintShort is suitable for JSON error.hint / one-line prompts.
func patHintShort() string {
	return "Create PAT: " + yunxiaoPATConsoleURL + " (help: " + yunxiaoPATHelpURL + ")"
}

// patPermissionsGuideZH is printed on missing-token / empty-token paths.
func patPermissionsGuideZH() string {
	return `推荐权限（控制台按「模块」勾选读/写，不是飞书式 OAuth scope）：
  - 组织/成员：读（whoami、organization）
  - 项目管理(Projex)：读+写（工作项；只读试用可只开读）
  - 代码管理(Codeup)：读+写（仓库/分支/文件/MR；只读试用可只开读）
  - 流水线(Flow)：读+写（list/run；只读试用可只开读）
  - 按需：制品 Packages、测试管理 Testhub、应用交付 AppStack
新建令牌名称建议：yunxiao-cli；设合理到期；令牌只显示一次；勿把完整 PAT 贴到聊天。
当前支持路径：控制台 PAT（无飞书式一键 OAuth）。`
}

// patPermissionsGuideEN is a short English checklist for README / optional stderr.
func patPermissionsGuideEN() string {
	return `Recommended PAT module checkboxes (not OAuth scopes):
  - Organization/members: read (whoami, organization)
  - Projex: read+write (work items; read-only trial: read only)
  - Codeup: read+write (repos/branches/files/MRs; trial: read only)
  - Flow/pipelines: read+write (list/run; trial: read only)
  - As needed: Packages, Testhub, AppStack
Token name tip: yunxiao-cli; set a sensible expiry; shown once; never paste raw PAT into chat.
Auth today: PAT console only (no Feishu-style one-click OAuth in this CLI yet).`
}

// patPermissionsGuide defaults to ZH for CLI stderr (Asia/Shanghai users).
func patPermissionsGuide() string { return patPermissionsGuideZH() }
