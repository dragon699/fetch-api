package swagger

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// Handler serves a custom Swagger UI page with enhanced options.
func Handler(docURL string) fiber.Handler {
	page := fmt.Sprintf(`<!doctype html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Kube Connector - Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  <style>
    :root { color-scheme: light; }
    body { margin: 0; padding: 0; background: #f5f7fc; }
    /* hide the top bar and tighten spacing */
    .swagger-ui .topbar { display: none; }
    .swagger-ui .swagger-ui { padding: 12px 28px 48px; max-width: 1200px; margin: 0 auto; }
    /* hide base URL line */
    .swagger-ui .info .base-url, .swagger-ui .info .url { display: none !important; }
    /* panel styling similar to the screenshot */
    .swagger-ui .opblock { border: 1px solid #d5e2f3; background: #eef3fb; }
    .swagger-ui .opblock-summary { background: #e3ecfb; }
    .swagger-ui .opblock-summary-method { background: #4a90e2; color: #fff; }
    .swagger-ui .opblock .opblock-section-header { background: #f4f7fd; border-color: #d5e2f3; }
    .swagger-ui .opblock .opblock-section-header h4 span { color: #2b3748; }
    .swagger-ui .responses-wrapper { background: #f4f7fd; }
    .swagger-ui section.models { border: 1px solid #d5e2f3; background: #fff; }
    .swagger-ui .btn.execute { background: linear-gradient(90deg, #4a90e2, #5ba2ff); border: none; color: #fff; }
    .swagger-ui .btn.clear { background: #e8edf5; color: #2b3748; }
    .swagger-ui .btn { box-shadow: none; }
    .swagger-ui .model-box { background: #fff; }
    /* code blocks to match dark snippet style */
    .swagger-ui .highlight-code, .swagger-ui .microlight, .swagger-ui .model-box .model-javadoc { background: #2c2f36 !important; color: #f8f8f2 !important; }
    .swagger-ui .response-col_status { color: #2b3748; }
    .swagger-ui .request-url { background: #2c2f36; color: #f8f8f2; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = () => {
      const ui = SwaggerUIBundle({
        url: "%s",
        dom_id: '#swagger-ui',
        deepLinking: true,
        docExpansion: "list",
        defaultModelsExpandDepth: -1,
        defaultModelExpandDepth: 2,
        displayRequestDuration: true,
        filter: true,
        tryItOutEnabled: true,
        persistAuthorization: true,
        requestSnippetsEnabled: true,
        syntaxHighlight: { theme: "github" },
        layout: "BaseLayout",
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
      });
      window.ui = ui;
    };
  </script>
</body>
</html>`, docURL)

	return func(ctx *fiber.Ctx) error {
		ctx.Type("html", "utf-8")
		return ctx.SendString(page)
	}
}
