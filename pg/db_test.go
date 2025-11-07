package pg

import (
    "testing"

    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

// UserM 为迁移示例模型，仅用于测试；包含主键与一个普通字段。
type UserM struct {
    ID       string `gorm:"primaryKey"`
    Username string
}

// TestAutoMigrateEnabled 验证在 autoMigrate=true 时，AutoMigrate 能正常执行并不返回错误。
func TestAutoMigrateEnabled(t *testing.T) {
    gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        t.Fatalf("failed to open sqlite: %v", err)
    }
    d := &DB{DB: gdb, autoMigrate: true}
    if err := d.AutoMigrate(&UserM{}); err != nil {
        t.Fatalf("auto migrate failed: %v", err)
    }
}

// TestAutoMigrateDisabled 验证在 autoMigrate=false 时，AutoMigrate 应直接返回 nil（不执行迁移）。
func TestAutoMigrateDisabled(t *testing.T) {
    gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        t.Fatalf("failed to open sqlite: %v", err)
    }
    d := &DB{DB: gdb, autoMigrate: false}
    if err := d.AutoMigrate(&UserM{}); err != nil {
        t.Fatalf("expected nil when autoMigrate disabled, got: %v", err)
    }
}