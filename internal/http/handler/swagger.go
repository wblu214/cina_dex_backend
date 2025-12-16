package handler

import "github.com/gin-gonic/gin"

// SwaggerUI serves the Swagger UI HTML page.
func SwaggerUI(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.File("docs/swagger.html")
}

// SwaggerSpec serves the OpenAPI specification.
func SwaggerSpec(c *gin.Context) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.File("docs/openapi.json")
}
