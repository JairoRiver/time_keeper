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
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/api/v1/entries-time": {
            "get": {
                "description": "Retrieve a paginated list of entry times for a user",
                "produces": [
                    "application/json"
                ],
                "summary": "List entry times",
                "operationId": "get-list-entry-time",
                "parameters": [
                    {
                        "type": "string",
                        "description": "User ID",
                        "name": "user_id",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "Page number (must be \u003e= 1)",
                        "name": "page_number",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/handler.EntryTimeResponse"
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/entry-time": {
            "put": {
                "description": "Update an existing entry time by its ID",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Update an entry time",
                "operationId": "put-update-entry-time",
                "parameters": [
                    {
                        "description": "Entry Time Data",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.UpdateEntryTimeParams"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handler.EntryTimeResponse"
                        }
                    }
                }
            },
            "post": {
                "description": "Create a new time entry for a user",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Create an entry time",
                "operationId": "post-create-entry-time",
                "parameters": [
                    {
                        "description": "Entry Time Data",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.CreateEntryTimeParams"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/handler.EntryTimeResponse"
                        }
                    }
                }
            }
        },
        "/api/v1/entry-time/{id}": {
            "get": {
                "description": "Retrieve an entry time by its ID",
                "produces": [
                    "application/json"
                ],
                "summary": "Get an entry time",
                "operationId": "get-entry-time",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Entry Time ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handler.EntryTimeResponse"
                        }
                    }
                }
            },
            "delete": {
                "description": "Delete an entry time by its ID",
                "produces": [
                    "application/json"
                ],
                "summary": "Delete an entry time",
                "operationId": "delete-entry-time",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Entry Time ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "202": {
                        "description": "No Content"
                    }
                }
            }
        },
        "/api/v1/user": {
            "post": {
                "description": "generate a new user",
                "produces": [
                    "application/json"
                ],
                "summary": "Create a new user",
                "operationId": "post-create-user",
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/handler.ResponseUser"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "handler.CreateEntryTimeParams": {
            "type": "object",
            "required": [
                "time_start",
                "user_id"
            ],
            "properties": {
                "tag": {
                    "type": "string"
                },
                "time_end": {
                    "type": "string"
                },
                "time_start": {
                    "type": "string"
                },
                "user_id": {
                    "type": "string"
                }
            }
        },
        "handler.EntryTimeResponse": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "string"
                },
                "tag": {
                    "type": "string"
                },
                "timeEnd": {
                    "type": "string"
                },
                "timeStart": {
                    "type": "string"
                }
            }
        },
        "handler.ResponseUser": {
            "type": "object",
            "properties": {
                "accessToken": {
                    "type": "string"
                },
                "email": {
                    "type": "string"
                },
                "expiredAt": {
                    "type": "string"
                },
                "userId": {
                    "type": "string"
                }
            }
        },
        "handler.UpdateEntryTimeParams": {
            "type": "object",
            "required": [
                "id"
            ],
            "properties": {
                "id": {
                    "type": "string"
                },
                "tag": {
                    "type": "string"
                },
                "time_end": {
                    "type": "string"
                },
                "time_start": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
