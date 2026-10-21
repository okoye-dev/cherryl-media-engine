package models

// EditorialCopy is the editable text on the public site. It is stored as one setting
// ("editorial_copy") and always returned in full: anything not saved yet falls back to the defaults.
type EditorialCopy struct {
	Hero    HeroCopy    `json:"hero"`
	Ticker  TickerCopy  `json:"ticker"`
	Dossier DossierCopy `json:"dossier"`
	Footer  FooterCopy  `json:"footer"`
}

// HeroCopy is the text at the top of the homepage. The headline is stored in three parts so the
// accent word can be styled: Lead + Accent + Trail read as one sentence.
type HeroCopy struct {
	EditionBadge string `json:"editionBadge"`
	TitleLead    string `json:"titleLead"`
	TitleAccent  string `json:"titleAccent"`
	TitleTrail   string `json:"titleTrail"`
	Description  string `json:"description"`
}

// TickerCopy is the scrolling news ticker. CustomNotice, when set, shows as the first item.
type TickerCopy struct {
	Label        string `json:"label"`
	CustomNotice string `json:"customNotice"`
}

// DossierCopy is the heading block of the Culture section.
type DossierCopy struct {
	IssueBadge string `json:"issueBadge"`
	Heading    string `json:"heading"`
	Subheading string `json:"subheading"`
}

// FooterCopy is the text in the site footer.
type FooterCopy struct {
	Tagline         string `json:"tagline"`
	Locations       string `json:"locations"`
	ContactEmail    string `json:"contactEmail"`
	CopyrightNotice string `json:"copyrightNotice"`
}

// DefaultEditorialCopy is what the site shows until something is saved. It matches the website's own defaults.
func DefaultEditorialCopy() EditorialCopy {
	return EditorialCopy{
		Hero: HeroCopy{
			EditionBadge: "DAILY EDITION",
			TitleLead:    "Independent ",
			TitleAccent:  "journalism",
			TitleTrail:   " & digital media network.",
			Description:  "Leading digital newsroom and new-age media distribution network covering state affairs, business, culture, politics, and dispatches across Africa and globally.",
		},
		Ticker: TickerCopy{
			Label:        "LIVE WIRE",
			CustomNotice: "",
		},
		Dossier: DossierCopy{
			IssueBadge: "WEEKEND EDITION · ISSUE #18",
			Heading:    "Culture, Cinema & In-Depth Features",
			Subheading: "Features, music chronicles, food heritage, and modern narratives across Africa and the diaspora.",
		},
		Footer: FooterCopy{
			Tagline:         "New-Age Media Distribution & Independent Journalism",
			Locations:       "Abuja · Lagos · London",
			ContactEmail:    "newsroom@cherrylmedia.com",
			CopyrightNotice: "© 2026 Cherryl Media Enterprise. All rights reserved.",
		},
	}
}
