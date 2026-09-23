package nfo

import (
	"math"
	"strconv"
	"strings"

	"golang.org/x/text/unicode/norm"

	"case/backend/pkg/logger"
	"case/backend/pkg/models"
)

// SceneMetadata is the database-independent result of mapping an NFO Movie onto
// scene concepts. It intentionally holds names rather than IDs; resolving names
// to performers/tags/studios and persisting them is the job of the caller.
type SceneMetadata struct {
	Title          string
	Code           string
	Details        string
	Director       string
	Date           *models.Date
	ProductionDate *models.Date
	Rating         *int

	StudioName     string
	PerformerNames []string
	TagNames       []string
	GroupNames     []string
	URLs           []string
	CustomFields   map[string]any
}

// SceneMetadata maps the parsed NFO onto scene metadata. The mapping is a pure
// function of the Movie; the only side effect is logging unparseable dates.
func (m *Movie) SceneMetadata() SceneMetadata {
	meta := SceneMetadata{
		Title:          firstNonEmpty(m.Title, m.OriginalTitle),
		Code:           strings.TrimSpace(m.Num),
		Details:        firstNonEmpty(m.Plot, m.Outline, m.Tagline),
		Director:       firstNonEmpty(m.Director, directorFromActors(m.Actors)),
		Date:           firstDate(m.Premiered, m.ReleaseDate, m.Year, metaTitle(m)),
		ProductionDate: parseNFODate(m.ReleaseDate, "releasedate", metaTitle(m)),
		Rating:         parseRating(m.Rating),
		StudioName:     firstNonEmpty(m.Studio, m.Maker, m.Publisher, m.Label),
		PerformerNames: performerNames(m.Actors),
		TagNames:       dedupNames(append(append([]string{}, m.Genres...), m.Tags...)),
		GroupNames:     dedupNames(append(setNames(m.Sets), m.Series)),
		URLs:           nonEmptyValues(m.JavbusID),
		CustomFields:   customFields(m),
	}

	return meta
}

func metaTitle(m *Movie) string {
	return firstNonEmpty(m.Title, m.OriginalTitle)
}

func performerNames(actors []Actor) []string {
	var names []string
	for _, a := range actors {
		if t := strings.TrimSpace(a.Type); t != "" && !strings.EqualFold(t, "Actor") {
			continue
		}
		names = append(names, a.Name)
	}

	return dedupNames(names)
}

func directorFromActors(actors []Actor) string {
	for _, a := range actors {
		if strings.EqualFold(strings.TrimSpace(a.Type), "Director") {
			return normalizeName(a.Name)
		}
	}

	return ""
}

func setNames(sets []Set) []string {
	names := make([]string, 0, len(sets))
	for _, s := range sets {
		names = append(names, s.Name)
	}

	return names
}

// firstDate returns the first parseable date from the provided values, in
// priority order. Unparseable values are logged and skipped. logContext is
// included in log messages only.
func firstDate(premiered, releaseDate, year, logContext string) *models.Date {
	for _, v := range []string{premiered, releaseDate, year} {
		if d := parseNFODate(v, "date", logContext); d != nil {
			return d
		}
	}

	return nil
}

func parseNFODate(value, field, title string) *models.Date {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	d, err := models.ParseDate(value)
	if err != nil {
		logger.Warnf("跳过无法解析的 NFO %s %q (title %q): %v", field, value, title, err)
		return nil
	}

	return &d
}

func parseRating(value string) *int {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	f, err := strconv.ParseFloat(value, 64)
	if err != nil || f <= 0 {
		return nil
	}

	// Kodi ratings are 0-10; scene ratings are expressed on a 1-100 scale.
	if f <= 10 {
		f *= 10
	}

	r := int(math.Round(f))
	switch {
	case r < 1:
		r = 1
	case r > 100:
		r = 100
	}

	return &r
}

// dedupNames trims, NFC-normalizes and case-insensitively de-duplicates names,
// preserving the first-seen spelling and order.
func dedupNames(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))

	for _, s := range in {
		s = normalizeName(s)
		if s == "" {
			continue
		}

		key := strings.ToLower(s)
		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}
		out = append(out, s)
	}

	return out
}

func normalizeName(s string) string {
	return strings.TrimSpace(norm.NFKC.String(s))
}

func nonEmptyValues(values ...string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, v)
		}
	}

	return out
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v = strings.TrimSpace(v); v != "" {
			return v
		}
	}

	return ""
}

func customFields(m *Movie) map[string]any {
	fields := map[string]any{
		"mpaa":          m.MPAA,
		"customrating":  m.CustomRating,
		"countrycode":   m.CountryCode,
		"originaltitle": m.OriginalTitle,
		"sorttitle":     m.SortTitle,
		"runtime":       m.Runtime,
		"javdbsearchid": m.JavdbSearchID,
	}

	for k, v := range fields {
		s, ok := v.(string)
		if !ok || strings.TrimSpace(s) == "" {
			delete(fields, k)
		}
	}

	if len(fields) == 0 {
		return nil
	}

	return fields
}
