package grpc

import (
	"net/http"
)

// SwaggerHandler handles serving Swagger UI and API documentation
type SwaggerHandler struct {
	swaggerJSONPath string
}

// NewSwaggerHandler creates a new SwaggerHandler
func NewSwaggerHandler(swaggerJSONPath string) *SwaggerHandler {
	return &SwaggerHandler{
		swaggerJSONPath: swaggerJSONPath,
	}
}

// ServeSwaggerUI serves the Swagger UI using CDN
func (h *SwaggerHandler) ServeSwaggerUI(w http.ResponseWriter, r *http.Request) {
	const swaggerHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go Clean DDD ES Template - API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
        .loading {
            text-align: center;
            padding: 50px;
            font-family: Arial, sans-serif;
        }
        .error {
            color: red;
            text-align: center;
            padding: 50px;
            font-family: Arial, sans-serif;
        }
    </style>
</head>
<body>
    <div id="swagger-ui">
        <div class="loading">
            <h2>Loading Swagger UI...</h2>
            <p>If this doesn't load, check the browser console for errors.</p>
        </div>
    </div>
    <script>
        // Add error handling for CDN loading
        function loadScript(src, callback) {
            var script = document.createElement('script');
            script.src = src;
            script.onload = callback;
            script.onerror = function() {
                console.error('Failed to load:', src);
                document.getElementById('swagger-ui').innerHTML = 
                    '<div class="error"><h2>Failed to load Swagger UI</h2><p>Error loading: ' + src + '</p><p>Please check your internet connection or try refreshing the page.</p></div>';
            };
            document.head.appendChild(script);
        }

        // Load CSS with error handling
        var link = document.createElement('link');
        link.rel = 'stylesheet';
        link.type = 'text/css';
        link.href = 'https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css';
        link.onerror = function() {
            console.error('Failed to load CSS');
        };
        document.head.appendChild(link);

        // Load scripts in sequence
        loadScript('https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js', function() {
            loadScript('https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js', function() {
                // Initialize Swagger UI
                try {
                    const ui = SwaggerUIBundle({
                        url: '/swagger.json',
                        dom_id: '#swagger-ui',
                        deepLinking: true,
                        presets: [
                            SwaggerUIBundle.presets.apis,
                            SwaggerUIStandalonePreset
                        ],
                        plugins: [
                            SwaggerUIBundle.plugins.DownloadUrl
                        ],
                        layout: "StandaloneLayout",
                        validatorUrl: null,
                        onComplete: function() {
                            console.log('Swagger UI loaded successfully');
                        },
                        onFailure: function(data) {
                            console.error('Swagger UI failed to load:', data);
                            document.getElementById('swagger-ui').innerHTML = 
                                '<div class="error"><h2>Failed to load API documentation</h2><p>Error: ' + JSON.stringify(data) + '</p></div>';
                        }
                    });
                } catch (error) {
                    console.error('Error initializing Swagger UI:', error);
                    document.getElementById('swagger-ui').innerHTML = 
                        '<div class="error"><h2>Error initializing Swagger UI</h2><p>Error: ' + error.message + '</p></div>';
                }
            });
        });
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(swaggerHTML))
}

// ServeSwaggerJSON serves the merged swagger JSON
func (h *SwaggerHandler) ServeSwaggerJSON(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, h.swaggerJSONPath)
}

// ServeSwaggerIndex serves a custom index page with links to Swagger UI
func (h *SwaggerHandler) ServeSwaggerIndex(w http.ResponseWriter, r *http.Request) {
	const indexHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go Clean DDD ES Template - API Documentation</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            text-align: center;
            margin-bottom: 30px;
        }
        .api-links {
            display: grid;
            gap: 20px;
            margin-top: 30px;
        }
        .api-link {
            display: block;
            padding: 20px;
            background: #f8f9fa;
            border: 1px solid #dee2e6;
            border-radius: 6px;
            text-decoration: none;
            color: #495057;
            transition: all 0.2s ease;
        }
        .api-link:hover {
            background: #e9ecef;
            border-color: #adb5bd;
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(0,0,0,0.15);
        }
        .api-link h3 {
            margin: 0 0 10px 0;
            color: #212529;
        }
        .api-link p {
            margin: 0;
            color: #6c757d;
        }
        .badge {
            display: inline-block;
            padding: 4px 8px;
            background: #007bff;
            color: white;
            border-radius: 4px;
            font-size: 12px;
            margin-left: 10px;
        }
        .info {
            background: #d1ecf1;
            border: 1px solid #bee5eb;
            color: #0c5460;
            padding: 15px;
            border-radius: 6px;
            margin-bottom: 20px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Go Clean DDD ES Template API</h1>
        
        <div class="info">
            <strong>📚 API Documentation</strong><br>
            This is a comprehensive API for the Go Clean DDD ES Template with Event Sourcing architecture.
            Choose from the options below to explore the API documentation.
        </div>

        <div class="api-links">
            <a href="/swagger/" class="api-link">
                <h3>📖 Interactive API Documentation <span class="badge">Recommended</span></h3>
                <p>Explore the API with Swagger UI - interactive documentation with try-it-out functionality</p>
            </a>
            
            <a href="/swagger.json" class="api-link">
                <h3>📄 Raw Swagger JSON</h3>
                <p>Download or view the complete OpenAPI/Swagger specification in JSON format</p>
            </a>
        </div>

        <div style="margin-top: 30px; text-align: center; color: #6c757d;">
            <p>Built with ❤️ using Clean Architecture, DDD, and Event Sourcing</p>
        </div>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(indexHTML))
}
