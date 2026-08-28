package handler

import "net/http"

const openAPISpec = `{
  "openapi": "3.0.3",
  "info": {
    "title": "Task Queue API",
    "description": "A Go stdlib REST API with background task processing via goroutines and channels.",
    "version": "1.0.0",
    "contact": {
      "name": "Sovan Kumar Bera",
      "email": "sovan.kumar.bera@hanriver.in"
    }
  },
  "servers": [{"url": "http://localhost:8080"}],
  "paths": {
    "/tasks": {
      "post": {
        "summary": "Create a task",
        "description": "Creates a new task and enqueues it for background processing.",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/CreateTaskRequest" }
            }
          }
        },
        "responses": {
          "201": {
            "description": "Task created",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Task" }
              }
            }
          },
          "400": {
            "description": "Invalid request",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Error" }
              }
            }
          }
        }
      },
      "get": {
        "summary": "List all tasks",
        "description": "Returns all tasks ordered by creation time descending.",
        "responses": {
          "200": {
            "description": "Task list",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": { "$ref": "#/components/schemas/Task" }
                }
              }
            }
          }
        }
      }
    },
    "/tasks/{id}": {
      "get": {
        "summary": "Get a task by ID",
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "schema": { "type": "integer" }
          }
        ],
        "responses": {
          "200": {
            "description": "Task found",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Task" }
              }
            }
          },
          "404": {
            "description": "Task not found",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Error" }
              }
            }
          }
        }
      }
    },
    "/health": {
      "get": {
        "summary": "Health check",
        "responses": {
          "200": {
            "description": "Service healthy",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "status": { "type": "string", "example": "ok" },
                    "queue_len": { "type": "integer", "example": 0 }
                  }
                }
              }
            }
          }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "CreateTaskRequest": {
        "type": "object",
        "required": ["title"],
        "properties": {
          "title": { "type": "string", "example": "Process report" },
          "description": { "type": "string", "example": "Generate monthly report for August" }
        }
      },
      "Task": {
        "type": "object",
        "properties": {
          "id": { "type": "integer", "example": 1 },
          "title": { "type": "string", "example": "Process report" },
          "description": { "type": "string", "example": "Generate monthly report for August" },
          "status": { "type": "string", "enum": ["pending", "processing", "completed", "failed"], "example": "completed" },
          "result": { "type": "string", "example": "processed: 5 words in description" },
          "created_at": { "type": "string", "format": "date-time" },
          "updated_at": { "type": "string", "format": "date-time" }
        }
      },
      "Error": {
        "type": "object",
        "properties": {
          "error": { "type": "string", "example": "title is required" }
        }
      }
    }
  }
}`

const swaggerHTML = `<!DOCTYPE html>
<html>
<head>
  <title>Task Queue API — Swagger UI</title>
  <meta charset="utf-8"/>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"/>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    SwaggerUIBundle({ url: "/swagger.json", dom_id: "#swagger-ui" });
  </script>
</body>
</html>`

func RegisterSwagger(mux *http.ServeMux) {
	mux.HandleFunc("GET /swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(openAPISpec))
	})
	mux.HandleFunc("GET /docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(swaggerHTML))
	})
}
