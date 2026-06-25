package main

import (
	"fmt"
	"log"
	"path/filepath"

	"api-test-platform/internal/config"
	"api-test-platform/internal/store"
)

func main() {
	cfg := config.Load()

	st, err := store.Open(cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	if err := st.MigrateAll(filepath.Join("migrations")); err != nil {
		log.Fatal("migrate:", err)
	}

	// 预置目录树（供 AI 分组复用测试）
	seedFolders(st)

	fmt.Println("✅ Seed 完成")
	fmt.Printf("   数据库: %s\n", cfg.DBPath)
	printSummary(st)
}

func seedFolders(st *store.Store) {
	type spec struct {
		productID int64
		paths     [][]string
	}
	specs := []spec{
		{
			productID: 1,
			paths: [][]string{
				{"T-Rex", "Portal", "Badge"},
				{"T-Rex", "Portal", "Persona"},
				{"T-Rex", "Portal", "Quest"},
				{"T-Rex", "Auth"},
			},
		},
		{
			productID: 2,
			paths: [][]string{
				{"Edgen", "API"},
				{"Edgen", "Auth"},
				{"Edgen", "Portal"},
			},
		},
		{
			productID: 3,
			paths: [][]string{
				{"example", "Demo"},
				{"example", "API"},
			},
		},
	}
	for _, s := range specs {
		for _, p := range s.paths {
			if _, _, err := st.EnsureFolderPath(s.productID, p); err != nil {
				log.Printf("warn: folder %v product %d: %v", p, s.productID, err)
			}
		}
	}
}

func printSummary(st *store.Store) {
	for _, pid := range []int64{1, 2, 3} {
		envs, _ := st.ListEnvironments(pid)
		tree, _ := st.BuildFolderTree(pid)
		fmt.Printf("\n--- 产品 %d ---\n", pid)
		fmt.Printf("  环境: %d 个\n", len(envs))
		for _, e := range envs {
			def := ""
			if e.IsDefault {
				def = " (default)"
			}
			fmt.Printf("    - %s%s → %s\n", e.Name, def, e.BaseURL)
		}
		fmt.Printf("  目录树节点: %d 个顶层\n", len(tree))
		paths, _ := st.FlatFolderPaths(pid)
		for _, p := range paths {
			fmt.Printf("    - %s\n", p)
		}
	}
}
