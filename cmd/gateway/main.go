package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"grpc/gateway/internal/config"
	"grpc/gateway/internal/logger"
	"grpc/gateway/internal/middleware"
	"grpc/gateway/pb"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const swaggerUIHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>Swagger UI</title>
	<link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
	<style>
		html { box-sizing: border-box; overflow-y: scroll; }
		*, *:before, *:after { box-sizing: inherit; }
		body { margin: 0; background: #fafafa; }
	</style>
</head>
<body>
	<div id="swagger-ui"></div>
	<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
	<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
	<script>
		window.onload = function() {
			window.ui = SwaggerUIBundle({
				urls: {{URLS}},
				dom_id: "#swagger-ui",
				deepLinking: true,
				presets: [
					SwaggerUIBundle.presets.apis,
					SwaggerUIStandalonePreset
				],
				plugins: [
					SwaggerUIBundle.plugins.DownloadUrl
				],
				layout: "StandaloneLayout"
			});
		};
	</script>
</body>
</html>`

type swaggerURL struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

func findProtoDir() string {
	dirs := []string{"proto", "../proto", "../../proto"}
	for _, dir := range dirs {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}
	return ""
}

func getSwaggerFiles(protoDir string) []string {
	var files []string
	entries, err := os.ReadDir(protoDir)
	if err != nil {
		return files
	}
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			files = append(files, entry.Name())
		}
	}
	return files
}

func formatSwaggerName(filename string) string {
	name := strings.TrimSuffix(filename, ".json")
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, "-", " ")
	if len(name) > 0 {
		return strings.ToUpper(name[:1]) + name[1:]
	}
	return name
}

func swaggerHandler(mux *runtime.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/docs/") && strings.HasSuffix(r.URL.Path, ".json") {
			filename := filepath.Base(r.URL.Path)
			protoDir := findProtoDir()
			if protoDir != "" {
				data, err := os.ReadFile(filepath.Join(protoDir, filename))
				if err == nil {
					w.Header().Set("Content-Type", "application/json")
					w.Write(data)
					return
				}
			}
			http.NotFound(w, r)
			return
		}
		if r.URL.Path == "/docs" || r.URL.Path == "/docs/" {
			protoDir := findProtoDir()
			var urls []swaggerURL
			if protoDir != "" {
				files := getSwaggerFiles(protoDir)
				for _, file := range files {
					urls = append(urls, swaggerURL{
						URL:  "/docs/" + file,
						Name: formatSwaggerName(file),
					})
				}
			}
			urlsJSON, err := json.Marshal(urls)
			if err != nil {
				urlsJSON = []byte("[]")
			}
			html := strings.Replace(swaggerUIHTMLTemplate, "{{URLS}}", string(urlsJSON), 1)
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(html))
			return
		}
		mux.ServeHTTP(w, r)
	})
}

func main() {
	ctx := context.Background()

	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	}

	userPort := config.EnvCfg.UserPort
	gatewayPort := config.EnvCfg.GatewayPort

	err := pb.RegisterUserServiceHandlerFromEndpoint(
		ctx,
		mux,
		"localhost:"+userPort,
		opts,
	)

	if err != nil {
		logger.Log.Fatal("gateway", "gateway failed to connect to user-service "+err.Error())
	}

	limiter := middleware.NewIPRateLimiter(5.0, 10.0)

	logger.Log.Info("gateway", "gateway running "+gatewayPort)

	if err := http.ListenAndServe(
		":"+gatewayPort,
		middleware.RateLimitMiddleware(limiter, swaggerHandler(mux)),
	); err != nil {
		logger.Log.Fatal("gateway", err.Error())
	}
}
