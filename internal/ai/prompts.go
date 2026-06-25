package ai

import (
	"encoding/json"
	"fmt"
)

func buildIngestPrompt(mode string, records []RawRecord, folderTree any, existingPaths []string, hint string) string {
	recordsJSON, _ := json.Marshal(records)
	treeJSON, _ := json.Marshal(folderTree)
	pathsJSON, _ := json.Marshal(existingPaths)

	classifyRules := `
## 自动分组规则（folder_path）
1. 根据 URL 路径、host、service 字段、接口语义，将每个接口归入合理的业务模块树。
2. folder_path 为字符串数组，表示从根到叶的路径，例如 ["T-Rex","Anchor"] 或 ["T-Rex","Quest"]。
3. 微服务划分：BETA/PRE/PROD 是运行环境，不是 folder_path；同一产品下不同后端（anchor、quest、trex 等）用 folder_path 第二级区分，优先使用录制数据中的 service 字段。
4. 每条录制的 path 字段为完整 pathname+query；输出 apis[].path 使用该完整路径（仅将动态段参数化为 {{id}} 等），不要截断为仅 /api/v1。
5. 优先复用已有目录（见 existing_folder_paths）；仅在确实无合适节点时新建路径。
6. 场景（scenario）的 folder_path 取主流程所属模块（通常是第一步所在模块或其父级）。
7. apis[].path 与 scenario.steps[].path 只输出 pathname+query（不要写 {{base_url}}）；服务端会按录制 service 自动绑定 base_url_trex / base_url_quest / base_url_anchor。
`

	outputSchema := `
## 输出 JSON Schema（不要 markdown，不要代码块）
{
  "apis": [
    {
      "name": "接口名称",
      "method": "GET",
      "path": "/api/v1/users/{{userId}}",
      "headers": [{"name":"Content-Type","value":"application/json","enabled":true}],
      "body": "",
      "body_type": "json",
      "description": "功能说明",
      "ai_remark": "测试注意点",
      "folder_path": ["模块","子模块"],
      "assertions": [
        {"type":"status_code","expression":"200","operator":"eq","expected":"200"},
        {"type":"json_path","expression":"$.code","operator":"eq","expected":"0"},
        {"type":"duration_ms","expression":"3000","operator":"lt","expected":""}
      ]
    }
  ],
  "scenario": null
}
`

	apiModeRules := ""
	if mode == "api" {
		apiModeRules = `
### 接口模式（mode=api）
- 录制数据有 N 条，则 apis 数组必须包含 N 个接口定义（每条录制对应一个元素），按录制顺序排列。
- scenario 必须为 null。
`
	}

	if mode == "scenario" {
		outputSchema = `
## 输出 JSON Schema（不要 markdown，不要代码块）
{
  "apis": [],
  "scenario": {
    "name": "场景名称",
    "description": "场景说明",
    "folder_path": ["模块","子模块"],
    "steps": [
      {
        "name": "步骤名",
        "method": "POST",
        "path": "/api/auth/login",
        "headers": [],
        "body": "{}",
        "extract_rules": [{"var":"token","jsonPath":"$.data.token"}],
        "assertions": [
          {"type":"status_code","expression":"200","operator":"eq","expected":"200"},
          {"type":"json_path","expression":"$.data.token","operator":"ne","expected":""},
          {"type":"duration_ms","expression":"3000","operator":"lt","expected":""}
        ]
      }
    ]
  }
}
`
	}

	return fmt.Sprintf(`你是资深接口自动化测试专家。根据浏览器录制的真实 HTTP 流量，生成可直接入库的接口定义与断言。

### 模式
%s

### 用户业务提示（可为空）
%s

### 录制数据
%s

### 已有目录树（优先复用）
%s

### 已有目录路径列表
%s

%s

%s

### 数据处理要求
- 请求头：优先保留录制中的业务头（Origin、Referer、X-*、Accept-Language 等）；若 apis[].headers 为空，服务端从 requestHeaders 回填。
- 请求体：POST/PUT/PATCH 必须保留录制 requestBody（可参数化 ID）；若 body 为空，服务端从录制回填。
- 仅脱敏：Cookie、User-Agent、Sec-*、Accept-Encoding 等浏览器噪声；长 Bearer 令牌改为 Bearer {{token}}。
- URL 路径：以录制 path 为准保留完整 pathname+query；仅将路径/查询中的数字 ID、UUID 替换为 {{变量名}}；不要输出完整域名或 {{base_url_*}}（执行时由环境变量注入）。
- 跨服务场景：同一场景可包含 trex/quest/anchor 等多条录制，每条步骤 path 仍为各自 pathname；执行时选择 BETA/PRE/PROD 一次即可。
- 断言 expected 优先来自真实响应；status_code、json_path、duration_ms 三类断言每个接口/步骤都必须有。
- json_path 的 expression 针对 **HTTP 响应体 JSON 根**（如 $.code、$.data.token）；禁止 $.body.xxx（body 仅为平台快照字段名，不是业务 JSON 根）。
- assertions[].expected 必须是 JSON 字符串（如 "200"、"0"、"true"），禁止输出布尔 true/false 或未加引号的数字。
- json_path 判断字段非空时用 operator **not_empty**（expected 可留空 ""）；等价写法 operator **ne** + expected **""**。
- 场景模式：按 timestamp 升序排列步骤；自动生成 extract_rules 供后续步骤使用 {{var}}。
%s
- 仅返回一个 JSON 对象，无其他文字。`,
		mode, hint, string(recordsJSON), string(treeJSON), string(pathsJSON),
		classifyRules, outputSchema, apiModeRules)
}
