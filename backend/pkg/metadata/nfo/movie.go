// Package nfo provides parsing of local NFO metadata files (the Kodi/Emby
// "movie" schema) into plain Go types, plus a mapping into scene metadata.
//
// The parser is intentionally free of any database or repository dependency so
// that it can be tested in isolation and reused by the scan pipeline.
package nfo

import "encoding/xml"

// Movie mirrors the standard Kodi NFO <movie> schema. Numeric values are kept
// as strings so that a missing element and a present-but-empty element are
// indistinguishable; callers decide how to interpret them.
type Movie struct {
	XMLName xml.Name `xml:"movie"`

	Title         string `xml:"title"`
	OriginalTitle string `xml:"originaltitle"`
	SortTitle     string `xml:"sorttitle"`

	// Num is the non-standard release number/code used by many JAV scrapers.
	Num string `xml:"num"`

	Tagline string `xml:"tagline"`
	Plot    string `xml:"plot"`
	Outline string `xml:"outline"`

	Premiered   string `xml:"premiered"`
	ReleaseDate string `xml:"releasedate"`
	Year        string `xml:"year"`
	Runtime     string `xml:"runtime"`

	Rating       string `xml:"rating"`
	MPAA         string `xml:"mpaa"`
	CustomRating string `xml:"customrating"`
	CountryCode  string `xml:"countrycode"`

	Studio    string `xml:"studio"`
	Maker     string `xml:"maker"`
	Publisher string `xml:"publisher"`
	Label     string `xml:"label"`
	Series    string `xml:"series"`
	Director  string `xml:"director"`

	Actors []Actor  `xml:"actor"`
	Sets   []Set    `xml:"set"`
	Tags   []string `xml:"tag"`
	Genres []string `xml:"genre"`

	// Non-standard identifiers used by JAV metadata tools.
	JavbusID      string `xml:"javbusid"`
	JavdbSearchID string `xml:"javdbsearchid"`
}

// Actor is a single <actor> entry. Type defaults to Actor when omitted.
type Actor struct {
	Name string `xml:"name"`
	Type string `xml:"type"`
}

// Set is a single <set> entry, used to model collections/series.
type Set struct {
	Name string `xml:"name"`
}
