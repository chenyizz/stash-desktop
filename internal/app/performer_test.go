package app

import (
	"testing"

	"case/backend/pkg/models"

	"github.com/stretchr/testify/assert"
)

func TestToPerformerDetailDTO(t *testing.T) {
	gender := models.GenderEnum("FEMALE")
	rating := 80
	height := 160

	p := &models.Performer{
		ID:             3,
		Name:           "A",
		Disambiguation: "d",
		Gender:         &gender,
		Birthdate:      mustDate(t, "1990-01-02"),
		Country:        "JP",
		Height:         &height,
		Measurements:   "B",
		Favorite:       true,
		Rating:         &rating,
		Details:        "det",
		Tattoos:        "tt",
		Aliases:        models.NewRelatedStrings([]string{"a1"}),
		URLs:           models.NewRelatedStrings([]string{"u1"}),
		TagIDs:         models.NewRelatedIDs([]int{1}),
	}

	dto := toPerformerDetailDTO(p, []*models.Tag{{ID: 1, Name: "t1"}}, true)

	assert.Equal(t, 3, dto.ID)
	assert.Equal(t, "A", dto.Name)
	assert.Equal(t, "d", dto.Disambiguation)
	assert.Equal(t, "FEMALE", dto.Gender)
	assert.Equal(t, "1990-01-02", dto.Birthdate)
	assert.Equal(t, "JP", dto.Country)
	assert.Equal(t, 80, dto.Rating)
	assert.True(t, dto.Favorite)
	assert.Equal(t, []string{"a1"}, dto.Aliases)
	assert.Equal(t, []string{"u1"}, dto.URLs)
	assert.Equal(t, []TagDTO{{ID: 1, Name: "t1"}}, dto.Tags)
	assert.Equal(t, "/performers/3/image", dto.ImageURL)
}

func TestToPerformerDetailDTO_NilAndNoImage(t *testing.T) {
	p := &models.Performer{
		ID:      1,
		Name:    "B",
		Aliases: models.NewRelatedStrings([]string{}),
		URLs:    models.NewRelatedStrings([]string{}),
		TagIDs:  models.NewRelatedIDs([]int{}),
	}

	dto := toPerformerDetailDTO(p, nil, false)

	assert.Equal(t, "", dto.Gender)
	assert.Equal(t, "", dto.Birthdate)
	assert.Equal(t, 0, dto.Rating)
	assert.Equal(t, "", dto.ImageURL)
	assert.Nil(t, dto.Tags)
}
