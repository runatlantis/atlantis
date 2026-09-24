package persistence

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/runatlantis/atlantis/internal/domain/project"
	"go.etcd.io/bbolt"
)

// ProjectRepository implements project.Repository using BoltDB
type ProjectRepository struct {
	db *bbolt.DB
}

var projectBucket = []byte("projects")

func NewProjectRepository(db *bbolt.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Save(ctx context.Context, proj *project.Project) error {
	return r.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(projectBucket)
		if err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}

		// Convert domain model to persistence model
		persistenceModel := projectToPersistence(proj)
		
		data, err := json.Marshal(persistenceModel)
		if err != nil {
			return fmt.Errorf("marshal project: %w", err)
		}

		key := []byte(string(proj.ID()))
		return bucket.Put(key, data)
	})
}

func (r *ProjectRepository) FindByID(ctx context.Context, id project.ProjectID) (*project.Project, error) {
	var persistenceModel ProjectPersistenceModel
	
	err := r.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(projectBucket)
		if bucket == nil {
			return fmt.Errorf("project not found")
		}

		key := []byte(string(id))
		data := bucket.Get(key)
		if data == nil {
			return fmt.Errorf("project not found")
		}

		return json.Unmarshal(data, &persistenceModel)
	})
	
	if err != nil {
		return nil, err
	}

	// Convert persistence model back to domain model
	return persistenceToProject(persistenceModel)
}

func (r *ProjectRepository) FindByRepositoryAndPR(ctx context.Context, repoFullName string, pullNumber int) ([]*project.Project, error) {
	var projects []*project.Project
	
	err := r.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(projectBucket)
		if bucket == nil {
			return nil // No projects exist yet
		}

		cursor := bucket.Cursor()
		for key, value := cursor.First(); key != nil; key, value = cursor.Next() {
			var persistenceModel ProjectPersistenceModel
			if err := json.Unmarshal(value, &persistenceModel); err != nil {
				continue // Skip corrupted entries
			}

			if persistenceModel.Repository.FullName == repoFullName {
				domainProject, err := persistenceToProject(persistenceModel)
				if err != nil {
					continue // Skip corrupted entries
				}
				projects = append(projects, domainProject)
			}
		}
		return nil
	})

	return projects, err
}

func (r *ProjectRepository) Delete(ctx context.Context, id project.ProjectID) error {
	return r.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(projectBucket)
		if bucket == nil {
			return nil // Nothing to delete
		}

		key := []byte(string(id))
		return bucket.Delete(key)
	})
}

func (r *ProjectRepository) List(ctx context.Context, filters project.ListFilters) ([]*project.Project, error) {
	var projects []*project.Project
	
	err := r.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(projectBucket)
		if bucket == nil {
			return nil // No projects exist yet
		}

		cursor := bucket.Cursor()
		for key, value := cursor.First(); key != nil; key, value = cursor.Next() {
			var persistenceModel ProjectPersistenceModel
			if err := json.Unmarshal(value, &persistenceModel); err != nil {
				continue // Skip corrupted entries
			}

			// Apply filters
			if filters.Repository != "" && persistenceModel.Repository.FullName != filters.Repository {
				continue
			}
			if filters.Status != "" && persistenceModel.Status != string(filters.Status) {
				continue
			}
			if filters.Workspace != "" && persistenceModel.Workspace != filters.Workspace {
				continue
			}

			domainProject, err := persistenceToProject(persistenceModel)
			if err != nil {
				continue // Skip corrupted entries
			}
			projects = append(projects, domainProject)
		}
		return nil
	})

	return projects, err
}

// ProjectPersistenceModel represents how project is stored
type ProjectPersistenceModel struct {
	ID          string                         `json:"id"`
	Name        string                         `json:"name"`
	Directory   string                         `json:"directory"`
	Workspace   string                         `json:"workspace"`
	Repository  RepositoryPersistenceModel     `json:"repository"`
	Status      string                         `json:"status"`
	Requirements RequirementsPersistenceModel  `json:"requirements"`
}

type RepositoryPersistenceModel struct {
	FullName string `json:"full_name"`
	Owner    string `json:"owner"`
	Name     string `json:"name"`
}

type RequirementsPersistenceModel struct {
	PlanRequirements  []string `json:"plan_requirements"`
	ApplyRequirements []string `json:"apply_requirements"`
}

func projectToPersistence(proj *project.Project) ProjectPersistenceModel {
	return ProjectPersistenceModel{
		ID:        string(proj.ID()),
		Name:      proj.Name(),
		Directory: proj.Directory(),
		Workspace: proj.Workspace(),
		Repository: RepositoryPersistenceModel{
			FullName: "", // Will need to add getter to domain model
			Owner:    "",
			Name:     "",
		},
		Status: string(proj.Status()),
		Requirements: RequirementsPersistenceModel{
			PlanRequirements:  []string{},
			ApplyRequirements: []string{},
		},
	}
}

func persistenceToProject(model ProjectPersistenceModel) (*project.Project, error) {
	repo := project.RepositoryInfo{
		FullName: model.Repository.FullName,
		Owner:    model.Repository.Owner,
		Name:     model.Repository.Name,
	}

	proj, err := project.NewProject(
		project.ProjectID(model.ID),
		model.Name,
		model.Directory,
		model.Workspace,
		repo,
	)
	if err != nil {
		return nil, fmt.Errorf("create project from persistence: %w", err)
	}

	return proj, nil
} 