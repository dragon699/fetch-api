package slack

import (
	"embed"
	"encoding/json"
	"fmt"

	"common/utils"
)

//go:embed templates/*.tpl
var templateFiles embed.FS

func RenderTemplate(templateName string, templateVars any) (map[string]any, error) {
	tplPath := fmt.Sprintf("templates/%s.tpl", templateName)
	tplBytes, err := templateFiles.ReadFile(tplPath)

	if err != nil {
		return nil, fmt.Errorf("[%s] Invalid template file: %s", err, tplPath)
	}

	tplContent, err := utils.RenderTemplateContent(tplPath, string(tplBytes), templateVars)

	if err != nil {
		return nil, fmt.Errorf("[%s] Failed to render embedded template file: %w", tplPath, err)
	}

	var slackBody map[string]any

	if err = json.Unmarshal([]byte(tplContent), &slackBody); err != nil {
		return nil, fmt.Errorf("[%s] Failed to unmarshal content after templating: %w", tplPath, err)
	}

	return slackBody, nil
}
