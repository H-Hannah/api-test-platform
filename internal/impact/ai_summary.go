package impact

import (
	"fmt"
	"strings"
)

const maxChangeDigestBytes = 14000

// buildChangeDigestForAI 将变更文件与 diff 节选整理为 AI 可读文本。
func buildChangeDigestForAI(files []prFile) string {
	if len(files) == 0 {
		return "（无变更文件）"
	}
	var b strings.Builder
	used := 0
	for i, f := range files {
		if i >= 40 {
			fmt.Fprintf(&b, "\n... 另有 %d 个文件未展开\n", len(files)-40)
			break
		}
		status := f.Status
		if status == "" {
			status = "modified"
		}
		header := fmt.Sprintf("### %s [%s]\n", f.Path, status)
		if used+len(header) > maxChangeDigestBytes {
			break
		}
		b.WriteString(header)
		used += len(header)

		patch := strings.TrimSpace(f.Patch)
		if patch == "" {
			line := "（无 diff 内容，仅路径变更）\n"
			if used+len(line) > maxChangeDigestBytes {
				break
			}
			b.WriteString(line)
			used += len(line)
			continue
		}
		remaining := maxChangeDigestBytes - used
		if len(patch) > remaining {
			patch = patch[:remaining] + "\n...(diff truncated)"
		}
		b.WriteString(patch)
		b.WriteByte('\n')
		used += len(patch) + 1
	}
	return b.String()
}

func buildMRContextForAI(req AnalyzeRequest) string {
	if strings.TrimSpace(req.ChangeDescription) != "" {
		return fmt.Sprintf(`- 来源: 口述变更
- 内容:
%s`, truncateForAI(req.ChangeDescription, 4000))
	}
	if u := strings.TrimSpace(req.GitLabMRURL); u != "" {
		meta, err := fetchGitLabMRMeta(u)
		if err == nil && meta != nil {
			return fmt.Sprintf(`- 来源: GitLab MR
- 标题: %s
- 分支: %s → %s
- 描述:
%s`,
				meta.Title, meta.SourceBranch, meta.TargetBranch,
				truncateForAI(meta.Description, 2000))
		}
	}
	if u := strings.TrimSpace(req.GitHubPRURL); u != "" {
		return fmt.Sprintf("- 来源: GitHub PR\n- 链接: %s", u)
	}
	if repo := strings.TrimSpace(req.Repo); repo != "" {
		return fmt.Sprintf("- 来源: 分支对比 %s\n- %s → %s", repo, req.BaseRef, req.HeadRef)
	}
	return "- 来源: 手动文件/接口列表"
}

func truncateForAI(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "\n...(truncated)"
}

func (s *Service) aiSummary(req AnalyzeRequest, files []prFile, res *AnalyzeResult) (string, error) {
	descMode := strings.TrimSpace(req.ChangeDescription) != ""
	taskIntro := "请阅读以下 Merge Request / 代码变更，用中文 Markdown 说明 **本次具体改了什么**。"
	notes := `- 只描述变更内容，不要写回归策略或推荐用例列表。
- 从 diff 归纳，不要臆造未出现的功能。`
	outputSections := `1. **变更概览**（1-2 句）
2. **接口/API 变更**（无则写「未检出路由变更」）
3. **核心业务改动**（按模块分点）
4. **其他变更**（配置、依赖等）`
	detail := buildChangeDigestForAI(files)

	if descMode {
		taskIntro = "请根据测试同学**口述的变更说明**，归纳 **本次可能改了什么、应关注哪些测试点**。"
		notes = `- 口述可能不完整，用「可能」「建议关注」等措辞。
- 配置/中间件/环境类变更要单独点出。`
		outputSections = `1. **变更概览**（1-2 句）
2. **接口/API**（有则列，无则写「未明确提及」）
3. **配置与环境**（Redis/Nacos/开关等）
4. **建议关注的测试点**（3-6 条）`
		detail = truncateForAI(req.ChangeDescription, maxChangeDigestBytes)
	}

	prompt := fmt.Sprintf(`你是资深测试架构师。%s

注意：
%s

## 输出结构（Markdown）
%s

## 变更来源
%s

## 推断接口路径
%v

## 变更详情
%s`,
		taskIntro, notes, outputSections,
		buildMRContextForAI(req), res.InferredAPIs, detail)
	return s.ai.Complete(prompt)
}
