package codegen

import (
	"quill/ast"
	"strings"
	"testing"
)

func TestGenerateFullStackServerBlock(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.ServerBlockStatement{
				Port: 4000,
				Routes: []ast.RouteDefinition{
					{
						Method: "get",
						Path:   "/api/users",
						Body: []ast.Statement{
							&ast.RespondStatement{
								Value:      &ast.ListLiteral{Elements: nil},
								StatusCode: 200,
								Line:       1,
							},
						},
						Line: 1,
					},
				},
				Line: 1,
			},
		},
	}

	js := GenerateFullStackApp(program)

	if !strings.Contains(js, "http.createServer") {
		t.Error("should generate http.createServer")
	}
	if !strings.Contains(js, "4000") {
		t.Error("should use port 4000")
	}
	if !strings.Contains(js, "/api/users") {
		t.Error("should include route path")
	}
}

func TestGenerateFullStackRoutes(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.ServerBlockStatement{
				Port: 3000,
				Routes: []ast.RouteDefinition{
					{
						Method: "get",
						Path:   "/api/items",
						Body: []ast.Statement{
							&ast.RespondStatement{
								Value: &ast.Identifier{Name: "items"},
								Line:  1,
							},
						},
						Line: 1,
					},
					{
						Method: "post",
						Path:   "/api/items",
						Body: []ast.Statement{
							&ast.RespondStatement{
								Value:      &ast.Identifier{Name: "newItem"},
								StatusCode: 201,
								Line:       2,
							},
						},
						Line: 2,
					},
				},
				Line: 1,
			},
		},
	}

	js := GenerateFullStackApp(program)

	if !strings.Contains(js, "'get'") {
		t.Error("should include GET method")
	}
	if !strings.Contains(js, "'post'") {
		t.Error("should include POST method")
	}
	if !strings.Contains(js, "'/api/items'") {
		t.Error("should include route path")
	}
}

func TestGenerateFullStackDatabase(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.DatabaseBlockStatement{
				ConnectString: "sqlite://app.db",
				Models: []ast.ModelDef{
					{
						Name: "User",
						Fields: []ast.ModelFieldDef{
							{Name: "name", Type: "text"},
							{Name: "email", Type: "text"},
							{Name: "age", Type: "number"},
						},
					},
				},
				Line: 1,
			},
			&ast.ServerBlockStatement{
				Port: 3000,
				Line: 2,
			},
		},
	}

	js := GenerateFullStackApp(program)

	if !strings.Contains(js, "DB") {
		t.Error("should generate DB object")
	}
	if !strings.Contains(js, `"User"`) {
		t.Error("should define User model")
	}
	if !strings.Contains(js, "sqlite://app.db") {
		t.Error("should include connection string")
	}
	if !strings.Contains(js, "DB.migrate") {
		t.Error("should call DB.migrate")
	}
}

func TestGenerateFullStackRespondStatement(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.ServerBlockStatement{
				Port: 3000,
				Routes: []ast.RouteDefinition{
					{
						Method: "get",
						Path:   "/api/health",
						Body: []ast.Statement{
							&ast.RespondStatement{
								Value:      &ast.ObjectLiteral{Keys: []string{"status"}, Values: []ast.Expression{&ast.StringLiteral{Value: "ok"}}},
								StatusCode: 200,
								Line:       1,
							},
						},
						Line: 1,
					},
				},
				Line: 1,
			},
		},
	}

	js := GenerateFullStackApp(program)

	if !strings.Contains(js, "res.writeHead") {
		t.Error("should generate res.writeHead")
	}
	if !strings.Contains(js, "res.end") {
		t.Error("should generate res.end")
	}
	if !strings.Contains(js, "JSON.stringify") {
		t.Error("should generate JSON.stringify")
	}
}

func TestGenerateFullStackCompleteServer(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.DatabaseBlockStatement{
				ConnectString: "sqlite://test.db",
				Models: []ast.ModelDef{
					{
						Name:   "Task",
						Fields: []ast.ModelFieldDef{{Name: "title", Type: "text"}},
					},
				},
				Line: 1,
			},
			&ast.ServerBlockStatement{
				Port: 8080,
				Routes: []ast.RouteDefinition{
					{
						Method: "get",
						Path:   "/api/tasks",
						Body: []ast.Statement{
							&ast.RespondStatement{
								Value: &ast.Identifier{Name: "tasks"},
								Line:  3,
							},
						},
						Line: 3,
					},
				},
				Line: 2,
			},
		},
	}

	js := GenerateFullStackApp(program)

	// Should be a complete runnable server
	if !strings.Contains(js, "http.createServer") {
		t.Error("should generate http.createServer")
	}
	if !strings.Contains(js, "8080") {
		t.Error("should use port 8080")
	}
	if !strings.Contains(js, "server.listen") {
		t.Error("should generate server.listen")
	}
	if !strings.Contains(js, "Quill app running") {
		t.Error("should log startup message")
	}
	if !strings.Contains(js, `"Task"`) {
		t.Error("should define Task model")
	}
}
