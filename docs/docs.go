// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag

package docs

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/alecthomas/template"
	"github.com/swaggo/swag"
)

var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{.Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/positions": {
            "post": {
                "description": "Create a new position with the input data",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "positions"
                ],
                "summary": "Create a new position",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/api.Response"
                        }
                    }
                }
            }
        },
        "/positions/{id}": {
            "get": {
                "description": "Get details of the position",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "positions"
                ],
                "summary": "Get details of the position",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Position ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/pages.Position"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "api.Response": {
            "type": "object",
            "properties": {
                "Status": {
                    "type": "integer"
                },
                "id": {
                    "type": "integer"
                },
                "message": {
                    "type": "string"
                }
            }
        },
        "pages.Position": {
            "type": "object",
            "properties": {
                "current_price": {
                    "type": "number"
                },
                "date_created": {
                    "type": "string"
                },
                "direction": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "instrument": {
                    "type": "string"
                },
                "margin": {
                    "type": "number"
                },
                "price": {
                    "type": "number"
                },
                "quantity": {
                    "type": "integer"
                },
                "result": {
                    "type": "number"
                },
                "stop_loss": {
                    "type": "string"
                },
                "take_profit": {
                    "type": "string"
                },
                "trailing_stop": {
                    "type": "string"
                }
            }
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     "",
	Host:        "",
	BasePath:    "",
	Schemes:     []string{},
	Title:       "",
	Description: "",
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register(swag.Name, &s{})
}
