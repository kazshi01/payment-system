package docs

const SwaggerHTML = `<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<title>Swagger UI</title>
		<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist/swagger-ui.css">
		<style>body { margin:0 } #swagger-ui { height: 100vh; }</style>
	</head>
	<body>
		<div id="swagger-ui"></div>
		<script src="https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js"></script>
		<script>
			window.onload = () => {
				window.ui = SwaggerUIBundle({
					url: '/openapi.yaml',
					dom_id: '#swagger-ui',
					presets: [SwaggerUIBundle.presets.apis],
					layout: "BaseLayout"
				});
			};
		</script>
	</body>
	</html>`
