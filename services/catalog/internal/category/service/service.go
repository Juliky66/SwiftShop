package service

import (
	"context"

	catrepo "shopkuber/catalog/internal/category/repository"
)

// CategoryTree is a nested category structure.
type CategoryTree struct {
	catrepo.Category
	Children []*CategoryTree `json:"children,omitempty"`
}

// Service handles category business logic.
type Service struct {
	repo *catrepo.Repository
}

// New creates a new category service.
func New(repo *catrepo.Repository) *Service {
	return &Service{repo: repo}
}

// Tree returns all categories as a nested tree.
func (s *Service) Tree(ctx context.Context) ([]*CategoryTree, error) {
	flat, err := s.repo.Tree(ctx)
	if err != nil {
		return nil, err
	}
	return buildTree(flat), nil
}

// GetBySlug returns a category and its immediate children.
func (s *Service) GetBySlug(ctx context.Context, slug string) (*CategoryTree, error) {
	cat, err := s.repo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	children, err := s.repo.Children(ctx, cat.ID)
	if err != nil {
		return nil, err
	}

	node := &CategoryTree{Category: *cat}
	for _, c := range children {
		node.Children = append(node.Children, &CategoryTree{Category: c})
	}
	return node, nil
}

// buildTree converts a flat list (ordered by parent) into a nested tree.
func buildTree(flat []catrepo.Category) []*CategoryTree {
	nodeMap := make(map[string]*CategoryTree, len(flat))
	for i := range flat {
		nodeMap[flat[i].ID] = &CategoryTree{Category: flat[i]}
	}

	var roots []*CategoryTree
	for _, cat := range flat {
		node := nodeMap[cat.ID]
		if cat.ParentID == nil {
			roots = append(roots, node)
		} else {
			if parent, ok := nodeMap[*cat.ParentID]; ok {
				parent.Children = append(parent.Children, node)
			}
		}
	}
	return roots
}
