package store

import (
	"database/sql"
	"strings"
)

func (s *Store) ListFolders(productID int64) ([]Folder, error) {
	rows, err := s.db.Query(`
		SELECT id, product_id, parent_id, name, path, created_at
		FROM folders WHERE product_id = ? ORDER BY path`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Folder
	for rows.Next() {
		var f Folder
		if err := rows.Scan(&f.ID, &f.ProductID, &f.ParentID, &f.Name, &f.Path, &f.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, f)
	}
	return list, rows.Err()
}

func (s *Store) BuildFolderTree(productID int64) ([]FolderTreeNode, error) {
	if productID > 0 {
		return s.buildFolderTreeForProduct(productID)
	}
	products, err := s.ListProducts()
	if err != nil {
		return nil, err
	}
	var roots []FolderTreeNode
	for _, p := range products {
		sub, err := s.buildFolderTreeForProduct(p.ID)
		if err != nil {
			return nil, err
		}
		roots = append(roots, sub...)
	}
	return roots, nil
}

func (s *Store) buildFolderTreeForProduct(productID int64) ([]FolderTreeNode, error) {
	flat, err := s.ListFolders(productID)
	if err != nil {
		return nil, err
	}
	byParent := map[int64][]Folder{}
	for _, f := range flat {
		byParent[f.ParentID] = append(byParent[f.ParentID], f)
	}
	var build func(parentID int64) []FolderTreeNode
	build = func(parentID int64) []FolderTreeNode {
		items := byParent[parentID]
		out := make([]FolderTreeNode, 0, len(items))
		for _, f := range items {
			out = append(out, FolderTreeNode{
				ID:       f.ID,
				Name:     f.Name,
				Path:     f.Path,
				Children: build(f.ID),
			})
		}
		return out
	}
	return build(0), nil
}

// EnsureFolderPath creates folders along path segments and returns leaf folder id.
// pathSegments e.g. ["用户中心", "认证"]
func (s *Store) EnsureFolderPath(productID int64, pathSegments []string) (int64, string, error) {
	if len(pathSegments) == 0 {
		return 0, "", nil
	}
	parentID := int64(0)
	fullPath := ""
	for _, seg := range pathSegments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		if fullPath == "" {
			fullPath = seg
		} else {
			fullPath = fullPath + "/" + seg
		}
		id, err := s.findOrCreateFolder(productID, parentID, seg, fullPath)
		if err != nil {
			return 0, "", err
		}
		parentID = id
	}
	return parentID, fullPath, nil
}

func (s *Store) findOrCreateFolder(productID, parentID int64, name, path string) (int64, error) {
	var id int64
	err := s.db.QueryRow(`
		SELECT id FROM folders WHERE product_id = ? AND parent_id = ? AND name = ?`,
		productID, parentID, name).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	res, err := s.db.Exec(`
		INSERT INTO folders (product_id, parent_id, name, path) VALUES (?, ?, ?, ?)`,
		productID, parentID, name, path)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) GetFolderPath(folderID int64) (string, error) {
	if folderID == 0 {
		return "", nil
	}
	var path string
	err := s.db.QueryRow(`SELECT path FROM folders WHERE id = ?`, folderID).Scan(&path)
	return path, err
}

func (s *Store) CreateFolder(productID, parentID int64, name string) (*Folder, error) {
	parentPath := ""
	if parentID > 0 {
		var err error
		parentPath, err = s.GetFolderPath(parentID)
		if err != nil {
			return nil, err
		}
	}
	path := name
	if parentPath != "" {
		path = parentPath + "/" + name
	}
	res, err := s.db.Exec(`
		INSERT INTO folders (product_id, parent_id, name, path) VALUES (?, ?, ?, ?)`,
		productID, parentID, name, path)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &Folder{ID: id, ProductID: productID, ParentID: parentID, Name: name, Path: path}, nil
}

func (s *Store) DeleteFolder(id int64) error {
	_, err := s.db.Exec(`DELETE FROM folders WHERE id = ?`, id)
	return err
}

// FlatFolderPaths returns existing paths for AI context.
func (s *Store) FlatFolderPaths(productID int64) ([]string, error) {
	rows, err := s.db.Query(`SELECT path FROM folders WHERE product_id = ? AND path != '' ORDER BY path`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var paths []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		paths = append(paths, p)
	}
	return paths, rows.Err()
}
