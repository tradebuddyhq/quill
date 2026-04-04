package stdlib

import (
	"strings"
	"testing"
)

func TestOrmRuntimeNonEmpty(t *testing.T) {
	runtime := GetOrmRuntime()
	if len(runtime) == 0 {
		t.Fatal("ORM runtime should not be empty")
	}
}

func TestOrmRuntimeContainsDBObject(t *testing.T) {
	runtime := GetOrmRuntime()
	if !strings.Contains(runtime, "var DB = {}") {
		t.Error("ORM runtime should define DB object")
	}
}

func TestOrmRuntimeContainsConnect(t *testing.T) {
	runtime := GetOrmRuntime()
	if !strings.Contains(runtime, "DB.connect") {
		t.Error("ORM runtime should contain DB.connect function")
	}
}

func TestOrmRuntimeContainsModel(t *testing.T) {
	runtime := GetOrmRuntime()
	if !strings.Contains(runtime, "DB.model") {
		t.Error("ORM runtime should contain DB.model function")
	}
}

func TestOrmRuntimeContainsMigrate(t *testing.T) {
	runtime := GetOrmRuntime()
	if !strings.Contains(runtime, "DB.migrate") {
		t.Error("ORM runtime should contain DB.migrate function")
	}
}

func TestOrmRuntimeContainsRaw(t *testing.T) {
	runtime := GetOrmRuntime()
	if !strings.Contains(runtime, "DB.raw") {
		t.Error("ORM runtime should contain DB.raw function")
	}
}

func TestOrmRuntimeContainsTransaction(t *testing.T) {
	runtime := GetOrmRuntime()
	if !strings.Contains(runtime, "DB.transaction") {
		t.Error("ORM runtime should contain DB.transaction function")
	}
}

func TestOrmRuntimeContainsModelCreate(t *testing.T) {
	runtime := GetOrmRuntime()
	if !strings.Contains(runtime, "Model.prototype.create") {
		t.Error("ORM runtime should contain Model.prototype.create")
	}
}

func TestOrmRuntimeContainsModelFind(t *testing.T) {
	runtime := GetOrmRuntime()
	if !strings.Contains(runtime, "Model.prototype.find") {
		t.Error("ORM runtime should contain Model.prototype.find")
	}
}

func TestOrmRuntimeContainsModelUpdate(t *testing.T) {
	runtime := GetOrmRuntime()
	if !strings.Contains(runtime, "Model.prototype.update") {
		t.Error("ORM runtime should contain Model.prototype.update")
	}
}

func TestOrmRuntimeContainsModelDelete(t *testing.T) {
	runtime := GetOrmRuntime()
	if !strings.Contains(runtime, "Model.prototype.delete") {
		t.Error("ORM runtime should contain Model.prototype.delete")
	}
}

func TestOrmRuntimeContainsQueryBuilder(t *testing.T) {
	runtime := GetOrmRuntime()
	if !strings.Contains(runtime, "QueryBuilder") {
		t.Error("ORM runtime should contain QueryBuilder class")
	}
}

func TestOrmRuntimeContainsQueryWhere(t *testing.T) {
	runtime := GetOrmRuntime()
	if !strings.Contains(runtime, "QueryBuilder.prototype.where") {
		t.Error("ORM runtime should contain QueryBuilder.prototype.where")
	}
}

func TestOrmRuntimeContainsQueryOrderBy(t *testing.T) {
	runtime := GetOrmRuntime()
	if !strings.Contains(runtime, "QueryBuilder.prototype.orderBy") {
		t.Error("ORM runtime should contain QueryBuilder.prototype.orderBy")
	}
}

func TestOrmRuntimeContainsQueryLimit(t *testing.T) {
	runtime := GetOrmRuntime()
	if !strings.Contains(runtime, "QueryBuilder.prototype.limit") {
		t.Error("ORM runtime should contain QueryBuilder.prototype.limit")
	}
}
