package pg

import (
    "testing"
    "strings"

    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

// newTestDB 返回一个启用 DryRun 的测试用 *DB，使用 sqlite 内存驱动以便无真实数据库也可生成 SQL。
// 作用：为单元测试构造查询链并提取生成的 SQL/绑定参数（Statement.SQL / Statement.Vars）。
func newTestDB(t *testing.T) *DB {
    t.Helper()
    gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DryRun: true})
    if err != nil {
        t.Fatalf("failed to open dryrun sqlite: %v", err)
    }
    return &DB{DB: gdb}
}

// execFind 触发一次干运行查询以让 GORM 构建 SQL；
// 返回构建完成的 *gorm.DB（其中 Statement.SQL 为最终 SQL，Statement.Vars 为绑定参数）。
func execFind(t *testing.T, db *DB) *gorm.DB {
    t.Helper()
    tx := db.DB.Session(&gorm.Session{DryRun: true}).Find(&[]struct{}{})
    if tx.Error != nil {
        t.Fatalf("dryrun find failed: %v", tx.Error)
    }
    return tx
}

// TestOptionDB_NilSafety 验证 OptionDB 对 nil 的安全处理，以及忽略 nil 选项的行为。
func TestOptionDB_NilSafety(t *testing.T) {
    // db 为 nil 时应直接返回 nil
    if OptionDB(nil, WithId("1")) != nil {
        t.Fatalf("OptionDB should return nil when input db is nil")
    }

    // 忽略 nil 的 QueryOption
    db := newTestDB(t)
    updated := OptionDB(db, nil, WithId("abc"))
    tx := execFind(t, updated)
    sql := tx.Statement.SQL.String()
    if !containsAll(sql, []string{"WHERE id = ?"}) {
        t.Fatalf("expected SQL to contain WHERE id = ?, got: %s", sql)
    }
    if len(tx.Statement.Vars) != 1 || tx.Statement.Vars[0] != "abc" {
        t.Fatalf("expected vars [abc], got: %#v", tx.Statement.Vars)
    }
}

// 已移除 WithTable，相关行为由 WithTableSafe 覆盖。

// TestWithTableWhitelist 验证 WithTableSafe 的白名单校验逻辑：
// 仅允许在 whitelist 中的表名生效，非白名单表名应被忽略。
func TestWithTableWhitelist(t *testing.T) {
    db := newTestDB(t)

    wl := map[string]struct{}{"users": {}, "accounts": {}}

    // 允许的表名
    updated := OptionDB(db, WithTableSafe("users", wl))
    tx := execFind(t, updated)
    sql := tx.Statement.SQL.String()
    if !containsAll(sql, []string{"FROM `users`"}) {
        t.Fatalf("expected SQL to use table users with whitelist, got: %s", sql)
    }

    // 非白名单表名应被忽略
    db2 := newTestDB(t)
    updated2 := OptionDB(db2, WithTableSafe("not_allowed", wl))
    tx2 := execFind(t, updated2)
    sql2 := tx2.Statement.SQL.String()
    if containsAll(sql2, []string{"FROM `not_allowed`"}) {
        t.Fatalf("non-whitelist table should be ignored, got: %s", sql2)
    }

    // 空白表名应被忽略
    db3 := newTestDB(t)
    db3.DB = db3.DB.Table("accounts")
    updated3 := OptionDB(db3, WithTableSafe("   ", wl))
    tx3 := execFind(t, updated3)
    sql3 := tx3.Statement.SQL.String()
    if !containsAll(sql3, []string{"FROM `accounts`"}) {
        t.Fatalf("blank table name should be ignored, expect accounts to remain, got: %s", sql3)
    }
}

// TestWhereBasic 验证 WithId/WithUserName/WithName/WithStatus 的 WHERE 条件拼接与 0 值忽略逻辑。
func TestWhereBasic(t *testing.T) {
    db := newTestDB(t)
    twl := map[string]struct{}{"users": {}, "accounts": {}}
    updated := OptionDB(db,
        WithTableSafe("users", twl),
        WithId("id123"),
        WithUserName("u1"),
        WithName("n1"),
        WithStatus(1), // 应生效
    )
    tx := execFind(t, updated)
    sql := tx.Statement.SQL.String()
    // 期望包含所有 WHERE 片段
    if !containsAll(sql, []string{"FROM `users`", "WHERE id = ?", "username = ?", "name = ?", "status = ?"}) {
        t.Fatalf("unexpected SQL: %s", sql)
    }
    // 绑定参数顺序应与追加顺序一致
    if len(tx.Statement.Vars) != 4 {
        t.Fatalf("expected 4 vars, got: %d, vars: %#v", len(tx.Statement.Vars), tx.Statement.Vars)
    }
    want := []any{"id123", "u1", "n1", 1}
    for i := range want {
        if tx.Statement.Vars[i] != want[i] {
            t.Fatalf("vars[%d] expected %v, got %v", i, want[i], tx.Statement.Vars[i])
        }
    }

    // status == 0 应忽略
    db2 := newTestDB(t)
    updated2 := OptionDB(db2, WithTableSafe("users", twl), WithStatus(0))
    tx2 := execFind(t, updated2)
    sql2 := tx2.Statement.SQL.String()
    if containsAll(sql2, []string{"status = ?"}) {
        t.Fatalf("status==0 should be ignored, got SQL: %s", sql2)
    }
}

// TestLimitOffset 验证 Limit/Offset 的设置与忽略逻辑。
func TestLimitOffset(t *testing.T) {
    db := newTestDB(t)
    twl := map[string]struct{}{"users": {}}
    updated := OptionDB(db, WithTableSafe("users", twl), WithLimit(10), WithOffset(5))
    tx := execFind(t, updated)
    sql := tx.Statement.SQL.String()
    if !containsAll(sql, []string{"LIMIT 10", "OFFSET 5"}) {
        t.Fatalf("expected LIMIT 10 and OFFSET 5, got: %s", sql)
    }

    // 非法值应忽略
    db2 := newTestDB(t)
    updated2 := OptionDB(db2, WithTableSafe("users", twl), WithLimit(0), WithOffset(-1))
    tx2 := execFind(t, updated2)
    sql2 := tx2.Statement.SQL.String()
    if containsAll(sql2, []string{"LIMIT", "OFFSET"}) {
        t.Fatalf("limit<=0 or offset<0 should be ignored, got: %s", sql2)
    }
}

// TestOrderWhitelist 验证升降序排序在白名单内才生效的逻辑。
func TestOrderWhitelist(t *testing.T) {
    db := newTestDB(t)
    twl := map[string]struct{}{"users": {}}
    wl := map[string]struct{}{"created_at": {}, "username": {}}
    updated := OptionDB(db, WithTableSafe("users", twl), OrderAsc("created_at", wl), OrderDesc("username", wl))
    tx := execFind(t, updated)
    sql := tx.Statement.SQL.String()
    if !containsAll(sql, []string{"ORDER BY created_at ASC", "username DESC"}) {
        t.Fatalf("expected order clauses, got: %s", sql)
    }

    // 非白名单字段应被忽略
    db2 := newTestDB(t)
    updated2 := OptionDB(db2, WithTableSafe("users", twl), OrderAsc("not_allowed", wl))
    tx2 := execFind(t, updated2)
    sql2 := tx2.Statement.SQL.String()
    if containsAll(sql2, []string{"ORDER BY not_allowed ASC"}) {
        t.Fatalf("non-whitelist field should be ignored, got: %s", sql2)
    }
}

// TestWithIn 验证通用 IN 条件在不同类型切片上的生效与空切片忽略逻辑。
func TestWithIn(t *testing.T) {
    twl := map[string]struct{}{"users": {}}
    // string 切片
    db := newTestDB(t)
    updated := OptionDB(db, WithTableSafe("users", twl), WithIn("id", []string{"a", "b"}))
    tx := execFind(t, updated)
    sql := tx.Statement.SQL.String()
    if !containsAll(sql, []string{"id IN"}) { // GORM 会展开为 (?,?)，这里只断言 IN 片段存在
        t.Fatalf("expected IN clause, got: %s", sql)
    }
    if len(tx.Statement.Vars) != 2 {
        t.Fatalf("expected 2 vars for IN, got: %d, vars: %#v", len(tx.Statement.Vars), tx.Statement.Vars)
    }

    // 空切片应忽略
    db2 := newTestDB(t)
    updated2 := OptionDB(db2, WithTableSafe("users", twl), WithIn("id", []string{}))
    tx2 := execFind(t, updated2)
    sql2 := tx2.Statement.SQL.String()
    if containsAll(sql2, []string{" IN "}) {
        t.Fatalf("empty slice should skip IN clause, got: %s", sql2)
    }

    // int 切片
    db3 := newTestDB(t)
    updated3 := OptionDB(db3, WithTableSafe("users", twl), WithIn("id", []int{1, 2, 3}))
    tx3 := execFind(t, updated3)
    sql3 := tx3.Statement.SQL.String()
    if !containsAll(sql3, []string{"id IN"}) || len(tx3.Statement.Vars) != 3 {
        t.Fatalf("expected IN clause with 3 int vars, got SQL: %s, vars: %#v", sql3, tx3.Statement.Vars)
    }

    // 空 int 切片应忽略
    db4 := newTestDB(t)
    updated4 := OptionDB(db4, WithTableSafe("users", twl), WithIn("id", []int{}))
    tx4 := execFind(t, updated4)
    sql4 := tx4.Statement.SQL.String()
    if containsAll(sql4, []string{" IN "}) {
        t.Fatalf("empty int slice should skip IN clause, got: %s", sql4)
    }

    // int64 切片
    db5 := newTestDB(t)
    updated5 := OptionDB(db5, WithTableSafe("users", twl), WithIn("id", []int64{9}))
    tx5 := execFind(t, updated5)
    sql5 := tx5.Statement.SQL.String()
    if !containsAll(sql5, []string{"id IN"}) || len(tx5.Statement.Vars) != 1 {
        t.Fatalf("expected IN clause with 1 int64 var, got SQL: %s, vars: %#v", sql5, tx5.Statement.Vars)
    }

    // nil string 切片应忽略
    var nilStrs []string
    db6 := newTestDB(t)
    updated6 := OptionDB(db6, WithTableSafe("users", twl), WithIn("id", nilStrs))
    tx6 := execFind(t, updated6)
    sql6 := tx6.Statement.SQL.String()
    if containsAll(sql6, []string{" IN "}) {
        t.Fatalf("nil string slice should skip IN clause, got: %s", sql6)
    }
}

// TestWithIdsNamesUsernames 验证针对固定列名的 IN 条件构造与空切片忽略。
func TestWithIdsNamesUsernames(t *testing.T) {
    // WithIds
    db := newTestDB(t)
    twl := map[string]struct{}{"users": {}}
    updated := OptionDB(db, WithTableSafe("users", twl), WithIds([]string{"i1", "i2"}))
    tx := execFind(t, updated)
    sql := tx.Statement.SQL.String()
    if !containsAll(sql, []string{"id IN"}) || len(tx.Statement.Vars) != 2 {
        t.Fatalf("WithIds expected IN with 2 vars, got SQL: %s, vars: %#v", sql, tx.Statement.Vars)
    }

    // WithNames
    db2 := newTestDB(t)
    updated2 := OptionDB(db2, WithTableSafe("users", twl), WithNames([]string{"n1"}))
    tx2 := execFind(t, updated2)
    sql2 := tx2.Statement.SQL.String()
    if !containsAll(sql2, []string{"name IN"}) || len(tx2.Statement.Vars) != 1 {
        t.Fatalf("WithNames expected IN with 1 var, got SQL: %s, vars: %#v", sql2, tx2.Statement.Vars)
    }

    // WithUsernames 空切片忽略
    db3 := newTestDB(t)
    updated3 := OptionDB(db3, WithTableSafe("users", twl), WithUsernames([]string{}))
    tx3 := execFind(t, updated3)
    sql3 := tx3.Statement.SQL.String()
    if containsAll(sql3, []string{"username IN"}) {
        t.Fatalf("empty usernames should skip IN clause, got: %s", sql3)
    }
}

// containsAll 判断 s 是否同时包含所有子串 parts。
func containsAll(s string, parts []string) bool {
    for _, p := range parts {
        if !contains(s, p) {
            return false
        }
    }
    return true
}

// contains 简单的子串检查，独立封装便于未来替换为更健壮的匹配。
func contains(s, sub string) bool {
    return len(sub) == 0 || (len(s) >= len(sub) && (func() bool { return (stringIndex(s, sub) >= 0) })())
}

// stringIndex 返回子串在父串中的起始索引，不存在则返回 -1。
// 之所以不直接使用 strings.Index，是为了在本文件中提供一个可替换的实现，避免过多依赖。
func stringIndex(s, sub string) int {
    // 简单实现：委托标准库，仍然保留函数级封装便于未来扩展
    return func() int { return indexStd(s, sub) }()
}

// indexStd 使用标准库实现的 Index 封装。
func indexStd(s, sub string) int {
    // 直接调用 strings.Index
    return func() int {
        return stringsIndex(s, sub)
    }()
}

// stringsIndex 是对 strings.Index 的轻量封装，避免直接在测试逻辑中出现过多标准库符号。
func stringsIndex(s, sub string) int {
    return Index(s, sub)
}

// Index 使用标准库 strings.Index 的别名实现。
// 注意：此别名只为可读性服务，不建议在生产代码中使用这种间接封装。
func Index(s, sub string) int {
    return indexImpl(s, sub)
}

// indexImpl 最终委托标准库 strings.Index。
// 将此函数独立出来以满足“函数级注释”的要求，同时保持可测试性。
func indexImpl(s, sub string) int {
    return stdIndex(s, sub)
}

// stdIndex 直接调用标准库 strings.Index。
func stdIndex(s, sub string) int { return strings.Index(s, sub) }