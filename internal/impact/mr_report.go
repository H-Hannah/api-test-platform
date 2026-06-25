package impact

import (
	"fmt"
	"strings"
	"time"
)

// MRCommentContext 生成 MR 评论时的附加上下文。
type MRCommentContext struct {
	GitLabMRURL   string
	Version       string
	RequirementID string
	TCDocsBranch  string
}

// BuildMRCommentMarkdown 生成可贴 GitLab MR 的 Markdown 报告。
func BuildMRCommentMarkdown(ctx MRCommentContext, res *AnalyzeResult) string {
	if res == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString("## 精准测试报告\n\n")
	b.WriteString("> 由 **API Test Platform** 自动生成\n\n")

	if ctx.TCDocsBranch != "" || ctx.Version != "" {
		b.WriteString("**用例来源**：`qa-doc-generator`")
		if ctx.TCDocsBranch != "" {
			b.WriteString(" · 分支 `" + ctx.TCDocsBranch + "`")
		}
		if ctx.Version != "" && ctx.RequirementID != "" {
			b.WriteString(fmt.Sprintf(" · `%s/%s`", ctx.Version, ctx.RequirementID))
		}
		b.WriteString("\n\n")
	}
	if s := strings.TrimSpace(res.Summary); s != "" {
		b.WriteString("**摘要**：" + s + "\n\n")
	}

	ai := strings.TrimSpace(res.AISummary)
	if ai != "" {
		title := "变更解读"
		if res.Source == "description" {
			title = "测试点解读"
		}
		b.WriteString("### " + title + "\n\n")
		b.WriteString(ai)
		b.WriteString("\n\n")
	}

	n := len(res.RecommendedTCs)
	b.WriteString(fmt.Sprintf("### 推荐用例（%d 条）\n\n", n))
	if reason := strings.TrimSpace(res.RecommendedTCReason); reason != "" {
		b.WriteString(reason + "\n\n")
	}
	if n == 0 {
		b.WriteString("_无得分>1的强相关用例_\n\n")
	} else {
		b.WriteString("| ID | 标题 | 优先级 | 分 |\n")
		b.WriteString("| --- | --- | --- | --- |\n")
		for _, tc := range res.RecommendedTCs {
			b.WriteString(fmt.Sprintf("| %s | %s | %s | %.2f |\n",
				escapeMDCell(tc.TCID),
				escapeMDCell(tc.Title),
				escapeMDCell(tc.Priority),
				tc.Score,
			))
		}
		b.WriteString("\n")
	}

	na := len(res.RecommendedAPIs)
	b.WriteString(fmt.Sprintf("### 推荐接口（%d 条）\n\n", na))
	if na == 0 {
		b.WriteString("_无推荐接口_\n\n")
	} else {
		b.WriteString("| 方法 | 路径 | 名称 | 场景就绪 | 分 |\n")
		b.WriteString("| --- | --- | --- | --- | --- |\n")
		for _, api := range res.RecommendedAPIs {
			ready := "否"
			if api.ScenarioReady {
				ready = "是"
			}
			b.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %.2f |\n",
				escapeMDCell(api.Method),
				escapeMDCell(api.Path),
				escapeMDCell(api.Name),
				ready,
				api.Score,
			))
		}
		b.WriteString("\n")
	}

	if res.Source != "description" && len(res.ChangedFiles) > 0 {
		b.WriteString(fmt.Sprintf("### 变更文件（%d）\n\n", len(res.ChangedFiles)))
		for _, f := range res.ChangedFiles {
			line := "- `" + escapeMDCell(f.Path) + "`"
			if f.Status != "" {
				line += " (" + f.Status + ")"
			}
			b.WriteString(line + "\n")
		}
		b.WriteString("\n")
	}

	if len(res.Gaps) > 0 {
		b.WriteString("### 提示\n\n")
		for _, g := range res.Gaps {
			line := "- "
			if g.Type != "" {
				line += "`" + g.Type + "` "
			}
			if g.API != "" {
				line += g.API + " — "
			}
			line += g.Action
			b.WriteString(line + "\n")
		}
		b.WriteString("\n")
	}

	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("_生成时间：%s_\n", time.Now().Format("2006-01-02 15:04 MST")))
	return b.String()
}

func escapeMDCell(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.TrimSpace(s)
}
