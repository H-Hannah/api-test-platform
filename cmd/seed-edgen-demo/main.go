// 灌入 Edgen 端到端演示数据：BDD 锚点 + 接口场景（含就绪/缺口），无需调用 AI。
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"api-test-platform/internal/config"
	"api-test-platform/internal/store"
)

const edgenProductID = int64(2)
const demoMRTag = "MR-EDGEN-DEMO"

func main() {
	cfg := config.Load()
	st, err := store.Open(cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	root := findRoot()
	if err := st.MigrateAll(filepath.Join(root, "migrations")); err != nil {
		log.Fatal("migrate:", err)
	}

	seedFolders(st)
	bddID := seedBDD(st, root)
	seedAPIs(st, bddID)

	fmt.Println("✅ Edgen 演示数据已就绪")
	fmt.Println("   产品: Edgen (id=2)")
	fmt.Printf("   BDD Feature id=%d\n", bddID)
	fmt.Printf("   MR 标签: %s\n", demoMRTag)
	fmt.Println()
	printWalkthrough(root)
}

func findRoot() string {
	wd, _ := os.Getwd()
	for dir := wd; dir != "/"; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
	}
	return wd
}

func readFile(root, rel string) string {
	b, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		log.Fatalf("read %s: %v", rel, err)
	}
	return string(b)
}

func seedFolders(st *store.Store) {
	paths := [][]string{
		{"Edgen", "Portal", "Platform"},
		{"Edgen", "API"},
	}
	for _, p := range paths {
		if _, _, err := st.EnsureFolderPath(edgenProductID, p); err != nil {
			log.Printf("folder %v: %v", p, err)
		}
	}
}

func seedBDD(st *store.Store, root string) int64 {
	gherkin := readFile(root, "fixtures/edgen/bdd-platform-bind.feature")
	f := &store.BDDFeature{
		ProductID:    edgenProductID,
		Title:        "社交平台绑定状态",
		UserStory:    "US-EDGEN-042",
		PRDText:      readFile(root, "fixtures/edgen/prd-platform-bind.md"),
		UIDesignText: readFile(root, "fixtures/edgen/design-platform-bind.md"),
		Gherkin:      gherkin,
		FeatureFiles: []store.BDDFeatureFile{{Filename: "bdd-platform-bind.feature", Content: gherkin}},
		GatePassed:   false,
	}
	id, err := st.CreateBDDFeature(f)
	if err != nil {
		log.Fatal("bdd:", err)
	}
	return id
}

func seedAPIs(st *store.Store, bddID int64) {
	bddRef := fmt.Sprintf("fixtures/edgen/bdd-platform-bind.feature (id=%d)", bddID)
	folderID, _, _ := st.EnsureFolderPath(edgenProductID, []string{"Edgen", "Portal", "Platform"})

	specs := []apiSpec{
		{
			name:    "获取 Twitter 平台绑定信息",
			method:  "GET",
			path:    "/v2/platform/bind/TWITTER?reverse=false",
			fullTpl: "{{base_url_edgen}}/v2/platform/bind/TWITTER?reverse=false",
			headers: `[{"name":"Authorization","value":"Bearer {{token}}","enabled":true},{"name":"Content-Type","value":"application/json","enabled":true}]`,
			ready:   true,
			us:      "US-EDGEN-042",
			bdd:     bddRef + " :: 查看 Twitter 绑定状态（已绑定）",
			mr:      demoMRTag,
			assertions: []store.Assertion{
				{Type: "status_code", Expression: "200", Operator: "eq", Expected: "200", Enabled: true},
				{Type: "json_path", Expression: "$.code", Operator: "eq", Expected: "0", Enabled: true},
				{Type: "json_path", Expression: "$.data.platform", Operator: "eq", Expected: "TWITTER", Enabled: true},
				{Type: "json_path", Expression: "$.data.bound", Operator: "not_empty", Expected: "", Enabled: true},
			},
		},
		{
			name:    "分页查询平台绑定列表",
			method:  "GET",
			path:    "/v2/platform/bindings?page=1&pageSize=20",
			fullTpl: "{{base_url_edgen}}/v2/platform/bindings?page=1&pageSize=20",
			headers: `[{"name":"Authorization","value":"Bearer {{token}}","enabled":true}]`,
			ready:   false,
			us:      "",
			bdd:     "",
			mr:      demoMRTag,
			assertions: []store.Assertion{
				{Type: "status_code", Expression: "200", Operator: "eq", Expected: "200", Enabled: true},
			},
		},
	}

	for _, sp := range specs {
		api := &store.APIDefinition{
			ProductID:       edgenProductID,
			FolderID:        folderID,
			Name:            sp.name,
			Method:          sp.method,
			Path:            sp.path,
			FullURLTemplate: sp.fullTpl,
			Headers:         sp.headers,
			BodyType:        "json",
			Description:     "Edgen 演示 — " + sp.name,
			AIRemark:        "seed-edgen-demo",
		}
		id, err := st.CreateAPI(api)
		if err != nil {
			log.Fatal(err)
		}
		if err := st.CreateAssertions(id, sp.assertions); err != nil {
			log.Fatal(err)
		}
		_ = st.UpdateAPIMeta(id, sp.us, sp.bdd, "", sp.mr)
		status := "缺口"
		if sp.ready {
			status = "就绪"
		}
		fmt.Printf("   API #%d %s [%s] %s\n", id, sp.method, status, sp.path)
	}
}

type apiSpec struct {
	name, method, path, fullTpl, headers, us, bdd, mr string
	ready                                           bool
	assertions                                      []store.Assertion
}

func printWalkthrough(root string) {
	doc := filepath.Join(root, "docs/EDGEN-WALKTHROUGH.md")
	fmt.Println("--- 操作指引（详见 docs/EDGEN-WALKTHROUGH.md）---")
	b, _ := os.ReadFile(doc)
	if len(b) > 0 {
		fmt.Println(string(b))
		return
	}
	fmt.Println("1. 顶栏选项目 Edgen、环境 PROD，环境管理里填 token")
	fmt.Println("2. BDD 设计 — 查看已灌入的「社交平台绑定状态」")
	fmt.Println("3. MR 核查 — MR-EDGEN-DEMO，路径列表见文档，点「AI 按 BDD 核对 MR」")
	fmt.Println("4. 接口定义 — 筛选 MR-EDGEN-DEMO，补全缺口接口的追溯信息")
}
