package graph_templates

import "embed"

//go:embed *
var templatesFolder embed.FS

type TemplateName string

const (
	DIGRAPH  = "digraph.dot"  // Basic template
	LABEL    = "label.dot"    // Labels
	RELATION = "relation.dot" // Relations
)

func GetTemplate(templateName TemplateName) string {
	result, err := templatesFolder.ReadFile(string(templateName))
	if err != nil {
		return ""
	}
	return string(result)
}
