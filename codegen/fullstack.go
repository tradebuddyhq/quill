package codegen

import (
	"quill/ast"
	"quill/tools"
)

// GenerateFullStackApp takes an AST program and separates it into server, database,
// auth, and frontend sections, then generates a complete Node.js server.
func GenerateFullStackApp(program *ast.Program) string {
	app := &tools.FullStackApp{}

	var frontendStmts []ast.Statement

	for _, stmt := range program.Statements {
		switch s := stmt.(type) {
		case *ast.ServerBlockStatement:
			app.Server = &tools.ServerBlock{
				Port:   s.Port,
				Routes: convertRoutes(s.Routes),
				Line:   s.Line,
			}
		case *ast.DatabaseBlockStatement:
			app.Database = &tools.DatabaseBlock{
				ConnectString: s.ConnectString,
				Models:        convertModels(s.Models),
				Line:          s.Line,
			}
		case *ast.AuthBlockStatement:
			app.Auth = &tools.AuthBlock{
				Secret: s.Secret,
				Routes: convertRoutes(s.Routes),
				Line:   s.Line,
			}
		case *ast.ComponentStatement, *ast.MountStatement:
			frontendStmts = append(frontendStmts, stmt)
		default:
			// Non-fullstack statements go to frontend
			frontendStmts = append(frontendStmts, stmt)
		}
	}

	// Generate frontend JS if there are component statements
	frontendJS := ""
	if len(frontendStmts) > 0 {
		frontendProg := &ast.Program{Statements: frontendStmts}
		gen := NewBrowser()
		frontendJS = gen.Generate(frontendProg)
	}

	return tools.GenerateFullStack(app, frontendJS)
}

func convertRoutes(astRoutes []ast.RouteDefinition) []tools.RouteHandler {
	var routes []tools.RouteHandler
	for _, r := range astRoutes {
		routes = append(routes, tools.RouteHandler{
			Method: r.Method,
			Path:   r.Path,
			Body:   r.Body,
			Line:   r.Line,
		})
	}
	return routes
}

func convertModels(astModels []ast.ModelDef) []tools.ModelDefinition {
	var models []tools.ModelDefinition
	for _, m := range astModels {
		var fields []tools.ModelField
		for _, f := range m.Fields {
			fields = append(fields, tools.ModelField{
				Name: f.Name,
				Type: f.Type,
			})
		}
		models = append(models, tools.ModelDefinition{
			Name:   m.Name,
			Fields: fields,
		})
	}
	return models
}
