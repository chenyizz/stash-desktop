//go:build integration
// +build integration

package sqlite_test

import (
	"context"
	"testing"

	"case/backend/pkg/models"

	"github.com/stretchr/testify/assert"
)

// 场景搜索应覆盖关联的 tag / 演员 / 工作室名称。
func TestSceneQueryQ_RelatedNames(t *testing.T) {
	withTxn(func(ctx context.Context) error {
		sqb := db.Scene

		cases := []struct {
			name string
			q    string
			want int
		}{
			{"tag", getTagStringValue(tagIdx1WithScene, "Name"), sceneIdxWithTwoTags},
			{"performer", getPerformerStringValue(performerIdxWithScene, "Name"), sceneIdxWithPerformer},
			{"studio", getStudioStringValue(studioIdxWithScene, "Name"), sceneIdxWithStudio},
		}

		for _, c := range cases {
			filter := models.FindFilterType{Q: &c.q}
			scenes := queryScene(ctx, t, sqb, nil, &filter)

			found := false
			for _, s := range scenes {
				if s.ID == sceneIDs[c.want] {
					found = true
					break
				}
			}
			assert.True(t, found, "Q (%s) %q should match scene index %d", c.name, c.q, c.want)
		}

		return nil
	})
}
