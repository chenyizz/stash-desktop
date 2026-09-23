package nfo

import (
	"context"
	"fmt"
	"strings"

	"case/backend/pkg/logger"
	"case/backend/pkg/models"
)

// Applier applies NFO metadata to a scene. It is safe for concurrent use: it
// holds no mutable state and performs all database work in a single writable
// transaction. SQLite serialises writable transactions (_txlock=immediate), so
// the resolve-or-create and fill-only read-modify-write sequences are atomic.
type Applier struct {
	Repo models.Repository
}

// Apply reads the NFO sidecar for videoPath and fills in scene metadata that is
// still empty. A missing or unparseable NFO is logged and skipped, so a bad
// sidecar can never abort a scan. A non-nil error is only returned for
// unexpected database failures; the caller is expected to log it.
func (a *Applier) Apply(ctx context.Context, sceneID int, videoPath string) (err error) {
	defer func() {
		if p := recover(); p != nil {
			logger.Errorf("panic while applying NFO metadata for %s: %v", videoPath, p)
			err = nil
		}
	}()

	nfoPath, ok := FindForVideo(videoPath)
	if !ok {
		logger.Debugf("No NFO sidecar found for %s", videoPath)
		return nil
	}

	movie, parseErr := ParseFile(nfoPath)
	if parseErr != nil {
		logger.Warnf("Skipping NFO %q for %s: %v", nfoPath, videoPath, parseErr)
		return nil
	}

	// All file I/O is complete; only the parsed struct enters the transaction.
	meta := movie.SceneMetadata()

	if err := a.Repo.WithTxn(ctx, func(ctx context.Context) error {
		return a.applyTxn(ctx, sceneID, meta)
	}); err != nil {
		return fmt.Errorf("applying NFO metadata to scene %d: %w", sceneID, err)
	}

	return nil
}

func (a *Applier) applyTxn(ctx context.Context, sceneID int, meta SceneMetadata) error {
	scene, err := a.Repo.Scene.Find(ctx, sceneID)
	if err != nil {
		return fmt.Errorf("finding scene: %w", err)
	}

	studioID, err := a.resolveStudioIfMissing(ctx, scene, meta.StudioName)
	if err != nil {
		return err
	}

	tagIDs, err := a.resolveTags(ctx, meta.TagNames)
	if err != nil {
		return err
	}

	performerIDs, err := a.resolvePerformers(ctx, meta.PerformerNames)
	if err != nil {
		return err
	}

	partial, hasChanges := buildPartial(scene, meta, studioID, tagIDs, performerIDs)
	if !hasChanges {
		return nil
	}

	if _, err := a.Repo.Scene.UpdatePartial(ctx, sceneID, partial); err != nil {
		return fmt.Errorf("updating scene: %w", err)
	}

	return nil
}

// buildPartial fills only the fields that are still empty on the scene, so that
// existing (including user-edited) metadata is never overwritten.
func buildPartial(scene *models.Scene, meta SceneMetadata, studioID *int, tagIDs, performerIDs []int) (models.ScenePartial, bool) {
	var partial models.ScenePartial
	var changed bool

	if scene.Title == "" && meta.Title != "" {
		partial.Title = models.NewOptionalString(meta.Title)
		changed = true
	}
	if scene.Code == "" && meta.Code != "" {
		partial.Code = models.NewOptionalString(meta.Code)
		changed = true
	}
	if scene.Details == "" && meta.Details != "" {
		partial.Details = models.NewOptionalString(meta.Details)
		changed = true
	}
	if scene.Director == "" && meta.Director != "" {
		partial.Director = models.NewOptionalString(meta.Director)
		changed = true
	}
	if scene.Date == nil && meta.Date != nil {
		partial.Date = models.NewOptionalDate(*meta.Date)
		changed = true
	}
	if scene.ProductionDate == nil && meta.ProductionDate != nil {
		partial.ProductionDate = models.NewOptionalDate(*meta.ProductionDate)
		changed = true
	}
	if scene.Rating == nil && meta.Rating != nil {
		partial.Rating = models.NewOptionalInt(*meta.Rating)
		changed = true
	}
	if scene.StudioID == nil && studioID != nil {
		partial.StudioID = models.NewOptionalInt(*studioID)
		changed = true
	}

	if len(tagIDs) > 0 {
		partial.TagIDs = &models.UpdateIDs{IDs: tagIDs, Mode: models.RelationshipUpdateModeAdd}
		changed = true
	}
	if len(performerIDs) > 0 {
		partial.PerformerIDs = &models.UpdateIDs{IDs: performerIDs, Mode: models.RelationshipUpdateModeAdd}
		changed = true
	}
	if len(meta.URLs) > 0 {
		partial.URLs = &models.UpdateStrings{Values: meta.URLs, Mode: models.RelationshipUpdateModeAdd}
		changed = true
	}

	return partial, changed
}

func (a *Applier) resolveStudioIfMissing(ctx context.Context, scene *models.Scene, name string) (*int, error) {
	if scene.StudioID != nil {
		return nil, nil
	}

	return a.resolveStudio(ctx, name)
}

func (a *Applier) resolveStudio(ctx context.Context, name string) (*int, error) {
	if name == "" {
		return nil, nil
	}

	studio, err := a.Repo.Studio.FindByName(ctx, name, true)
	if err != nil {
		return nil, fmt.Errorf("finding studio %q: %w", name, err)
	}
	if studio != nil {
		return &studio.ID, nil
	}

	newStudio := models.NewCreateStudioInput()
	newStudio.Name = name
	if err := a.Repo.Studio.Create(ctx, &newStudio); err != nil {
		return nil, fmt.Errorf("creating studio %q: %w", name, err)
	}

	return &newStudio.ID, nil
}

func (a *Applier) resolveTags(ctx context.Context, names []string) ([]int, error) {
	if len(names) == 0 {
		return nil, nil
	}

	existing, err := a.Repo.Tag.FindByNames(ctx, names, true)
	if err != nil {
		return nil, fmt.Errorf("finding tags: %w", err)
	}

	resolved := make(map[string]int, len(names))
	for _, t := range existing {
		resolved[strings.ToLower(t.Name)] = t.ID
	}

	ids := make([]int, 0, len(names))
	for _, name := range names {
		key := strings.ToLower(name)
		if id, ok := resolved[key]; ok {
			ids = append(ids, id)
			continue
		}

		newTag := models.NewTag()
		newTag.Name = name
		if err := a.Repo.Tag.Create(ctx, &models.CreateTagInput{Tag: &newTag}); err != nil {
			return nil, fmt.Errorf("creating tag %q: %w", name, err)
		}

		resolved[key] = newTag.ID
		ids = append(ids, newTag.ID)
	}

	return ids, nil
}

func (a *Applier) resolvePerformers(ctx context.Context, names []string) ([]int, error) {
	if len(names) == 0 {
		return nil, nil
	}

	existing, err := a.Repo.Performer.FindByNames(ctx, names, true)
	if err != nil {
		return nil, fmt.Errorf("finding performers: %w", err)
	}

	resolved := make(map[string]int, len(names))
	for _, p := range existing {
		resolved[strings.ToLower(p.Name)] = p.ID
	}

	ids := make([]int, 0, len(names))
	for _, name := range names {
		key := strings.ToLower(name)
		if id, ok := resolved[key]; ok {
			ids = append(ids, id)
			continue
		}

		newPerformer := models.NewPerformer()
		newPerformer.Name = name
		createErr := a.Repo.Performer.Create(ctx, &models.CreatePerformerInput{Performer: &newPerformer})
		if createErr != nil {
			// Performers have a unique name index. If another writer created the
			// same performer between our find and create, use the existing row.
			reFound, findErr := a.Repo.Performer.FindByNames(ctx, []string{name}, true)
			if findErr == nil && len(reFound) > 0 {
				resolved[key] = reFound[0].ID
				ids = append(ids, reFound[0].ID)
				continue
			}
			return nil, fmt.Errorf("creating performer %q: %w", name, createErr)
		}

		resolved[key] = newPerformer.ID
		ids = append(ids, newPerformer.ID)
	}

	return ids, nil
}
