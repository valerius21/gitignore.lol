// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "license": {
            "name": "MIT",
            "url": "https://github.com/valerius21/gitignore.lol/blob/main/LICENSE"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/api/list": {
            "get": {
                "description": "Returns a list of all available .gitignore templates",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "templates"
                ],
                "summary": "List available templates",
                "responses": {
                    "200": {
                        "description": "List of available templates",
                        "schema": {
                            "$ref": "#/definitions/pkg_server.TemplateResponse"
                        }
                    },
                    "500": {
                        "description": "Internal server error",
                        "schema": {
                            "$ref": "#/definitions/pkg_server.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/api/{templateList}": {
            "get": {
                "description": "Returns combined .gitignore file for specified templates",
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "templates"
                ],
                "summary": "Get gitignore templates",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Comma-separated list of templates (e.g., go,node,python)",
                        "name": "templateList",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Combined .gitignore file content",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Template not found",
                        "schema": {
                            "$ref": "#/definitions/pkg_server.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal server error",
                        "schema": {
                            "$ref": "#/definitions/pkg_server.ErrorResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "pkg_server.ErrorResponse": {
            "type": "object",
            "properties": {
                "error": {
                    "description": "Error message",
                    "type": "string",
                    "example": "Template not found"
                }
            }
        },
        "pkg_server.TemplateResponse": {
            "type": "object",
            "properties": {
                "files": {
                    "description": "List of available gitignore templates",
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "[\"go\"",
                        "\"node\"",
                        "\"python\"]"
                    ]
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:4444",
	BasePath:         "/",
	Schemes:          []string{"http"},
	Title:            "gitignore.lol API",
	Description:      "A service to generate .gitignore files for your projects. An implementation inspired by the previously known gitignore.io.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
