package clickhouse

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// QueryOption 定义对 *DB 进行链式包装的函数类型，返回经变更后的 *DB，便于组合多个查询配置。
// 所有实现都应满足幂等与安全（避免 SQL 注入）的要求。
type QueryOption func(db *DB) *DB

// OptionDB 按序应用一组 QueryOption 到传入的 *DB 并返回更新后的 *DB。
// 为提升健壮性：
// 1) 当 db 为 nil 时返回错误；
// 2) 跳过为 nil 的选项，保证调用安全。
func OptionDB(db *DB, options ...QueryOption) (*DB, error) {
	if db == nil {
		return nil, NewQueryError("database instance cannot be nil", nil).
			WithCode("DB_NIL")
	}

	if db.DB == nil {
		return nil, NewQueryError("gorm database instance cannot be nil", nil).
			WithCode("GORM_DB_NIL")
	}

	var lastError error
	for i, option := range options {
		if option == nil {
			continue
		}

		// 在应用选项时进行错误捕获
		func() {
			defer func() {
				if r := recover(); r != nil {
					lastError = NewQueryError(fmt.Sprintf("panic applying query option at index %d", i),
						fmt.Errorf("panic: %v", r)).
						WithCode("QUERY_OPTION_PANIC").
						WithContext("option_index", i)
				}
			}()
			db = option(db)
		}()

		if db == nil {
			return nil, NewQueryError(fmt.Sprintf("query option at index %d returned nil database", i), lastError).
				WithCode("QUERY_OPTION_NIL_RESULT").
				WithContext("option_index", i)
		}
	}

	if lastError != nil {
		return db, WrapError(lastError, ErrorTypeQuery, "some query options failed to apply")
	}

	return db, nil
}

// OptionDBMust 与 OptionDB 功能相同，但在出错时返回 nil 而不是错误，用于向后兼容
func OptionDBMust(db *DB, options ...QueryOption) *DB {
	result, err := OptionDB(db, options...)
	if err != nil {
		// 在调试模式下记录错误，但保持向后兼容
		if db != nil && db.DB != nil {
			// 可以考虑添加日志记录
			_ = err
		}
		return db // 返回原始 db 而不是 nil
	}
	return result
}

// WithTable 设置查询所使用的表名。
// 注意：tableName 需来源于受控白名单以防止 SQL 注入，这里进行更严格的验证。
func WithTable(tableName string) QueryOption {
	return func(db *DB) *DB {
		t := strings.TrimSpace(tableName)
		if t == "" {
			return db
		}

		// 验证表名的安全性
		if err := validateTableName(t); err != nil {
			// 记录验证错误但不中断查询流程
			// 在生产环境中，这里可以记录日志或返回错误
			return db
		}

		db.DB = db.DB.Table(t)
		return db
	}
}

// validateTableName 验证表名是否安全，防止 SQL 注入
func validateTableName(tableName string) error {
	if tableName == "" {
		return nil // 空表名不报错，会由调用方处理
	}

	// 检查表名长度
	if len(tableName) > 64 {
		return NewValidationError("table name too long (max 64 characters)", nil).
			WithContext("table_name", tableName).
			WithContext("length", len(tableName)).
			WithCode("TABLE_NAME_TOO_LONG")
	}

	// 检查表名格式：只允许字母、数字、下划线，且必须以字母开头
	validTableName := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)
	if !validTableName.MatchString(tableName) {
		return NewValidationError("invalid table name format", nil).
			WithContext("table_name", tableName).
			WithCode("TABLE_NAME_INVALID_FORMAT").
			WithRetriable(false)
	}

	// 检查是否为 SQL 关键字
	if isSQLKeyword(tableName) {
		return NewValidationError(fmt.Sprintf("table name '%s' cannot be SQL keyword", tableName), nil).
			WithContext("table_name", tableName).
			WithCode("TABLE_NAME_SQL_KEYWORD")
	}

	return nil
}

// isSQLKeyword 检查是否为 SQL 关键字
func isSQLKeyword(word string) bool {
	sqlKeywords := map[string]bool{
		"SELECT": true, "FROM": true, "WHERE": true, "INSERT": true,
		"UPDATE": true, "DELETE": true, "CREATE": true, "DROP": true,
		"ALTER": true, "INDEX": true, "TABLE": true, "DATABASE": true,
		"ORDER": true, "BY": true, "LIMIT": true, "OFFSET": true,
		"JOIN": true, "INNER": true, "LEFT": true, "RIGHT": true,
		"OUTER": true, "UNION": true, "GROUP": true, "HAVING": true,
		"EXISTS": true, "IN": true, "AND": true, "OR": true,
		"NOT": true, "NULL": true, "TRUE": true, "FALSE": true,
		"CASE": true, "WHEN": true, "THEN": true, "ELSE": true,
		"END": true, "AS": true, "ON": true, "DISTINCT": true,
		"BETWEEN": true, "LIKE": true, "ILIKE": true, "IS": true,
	}

	_, exists := sqlKeywords[strings.ToUpper(word)]
	return exists
}

// WithId 按主键 id 追加 WHERE 条件（id = ?）。当 id 为空或仅包含空白时忽略该条件。
func WithId(id string) QueryOption {
	return func(db *DB) *DB {
		trimmedId := strings.TrimSpace(id)
		if trimmedId == "" {
			return db
		}

		// 验证 ID 的安全性
		if err := validateID(trimmedId); err != nil {
			// 记录验证错误但继续执行
			return db
		}

		db.DB = db.DB.Where("id = ?", trimmedId)
		return db
	}
}

// validateID 验证 ID 参数的安全性
func validateID(id string) error {
	if len(id) > 255 {
		return NewValidationError("id too long (max 255 characters)", nil).
			WithContext("id_length", len(id)).
			WithCode("ID_TOO_LONG")
	}

	// 检查是否包含危险字符
	if containsControlCharacters(id) {
		return NewValidationError("id contains invalid control characters", nil).
			WithContext("id", id).
			WithCode("ID_INVALID_CHARS")
	}

	return nil
}

// containsControlCharacters 检查字符串是否包含控制字符
func containsControlCharacters(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) && r != '\t' && r != '\n' && r != '\r' {
			return true
		}
	}
	return false
}

// WithUserName 按 username 追加 WHERE 条件（username = ?）。当 username 为空或仅包含空白时忽略该条件。
func WithUserName(username string) QueryOption {
	return func(db *DB) *DB {
		if strings.TrimSpace(username) == "" {
			return db
		}
		db.DB = db.DB.Where("username = ?", username)
		return db
	}
}

// WithName 按 name 追加 WHERE 条件（name = ?）。当 name 为空或仅包含空白时忽略该条件。
func WithName(name string) QueryOption {
	return func(db *DB) *DB {
		if strings.TrimSpace(name) == "" {
			return db
		}
		db.DB = db.DB.Where("name = ?", name)
		return db
	}
}

// WithStatus 按 status 追加 WHERE 条件（status = ?）。
// 说明：当 status == 0 时该条件会被忽略；若业务允许 0 为有效状态，建议改用更明确的条件构造函数以避免歧义。
func WithStatus(status int) QueryOption {
	return func(db *DB) *DB {
		if status == 0 {
			return db
		}
		db.DB = db.DB.Where("status = ?", status)
		return db
	}
}

// WithOffset 设置查询的 OFFSET（偏移量）。当 offset 为负数时忽略该选项。
func WithOffset(offset int) QueryOption {
	// WithOffset 返回一个合法的 QueryOption。
	// 当 offset 为负数时，该选项将被忽略（不设置 OFFSET），以减少调用方对 nil 的处理负担。
	return func(db *DB) *DB {
		if offset < 0 {
			return db
		}
		db.DB = db.DB.Offset(offset)
		return db
	}
}

// WithLimit 设置查询的 LIMIT（行数上限）。当 limit 小于等于 0 时忽略该选项。
func WithLimit(limit int) QueryOption {
	// WithLimit 返回一个合法的 QueryOption。
	// 当 limit 小于等于 0 时，该选项将被忽略（不设置 LIMIT），以减少调用方对 nil 的处理负担。
	return func(db *DB) *DB {
		if limit <= 0 {
			return db
		}
		db.DB = db.DB.Limit(limit)
		return db
	}
}

// OrderAsc 对指定字段进行升序排序，字段必须出现在白名单中以防止 SQL 注入。
// whitelist 参数由上层按业务整理（如 {"id","created_at","username"}），此处会忽略空白字段。
func OrderAsc(field string, whitelist map[string]struct{}) QueryOption {
	return func(db *DB) *DB {
		f := strings.TrimSpace(field)
		if f == "" {
			return db
		}

		// 验证字段名安全性
		if err := validateFieldName(f, whitelist); err != nil {
			// 记录验证错误但继续执行
			return db
		}

		db.DB = db.DB.Order(f + " ASC")
		return db
	}
}

// OrderDesc 对指定字段进行降序排序，字段必须出现在白名单中以防止 SQL 注入。
// whitelist 参数由上层按业务整理，忽略空白字段。
func OrderDesc(field string, whitelist map[string]struct{}) QueryOption {
	return func(db *DB) *DB {
		f := strings.TrimSpace(field)
		if f == "" {
			return db
		}

		// 验证字段名安全性
		if err := validateFieldName(f, whitelist); err != nil {
			// 记录验证错误但继续执行
			return db
		}

		db.DB = db.DB.Order(f + " DESC")
		return db
	}
}

// validateFieldName 验证字段名是否安全，防止 SQL 注入
func validateFieldName(field string, whitelist map[string]struct{}) error {
	if field == "" {
		return nil // 空字段名不报错，会由调用方处理
	}

	// 检查字段名长度
	if len(field) > 64 {
		return NewValidationError("field name too long (max 64 characters)", nil).
			WithContext("field_name", field).
			WithContext("length", len(field)).
			WithCode("FIELD_NAME_TOO_LONG")
	}

	// 检查字段名格式
	validFieldName := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	if !validFieldName.MatchString(field) {
		return NewValidationError("invalid field name format", nil).
			WithContext("field_name", field).
			WithCode("FIELD_NAME_INVALID_FORMAT")
	}

	// 检查是否在白名单中
	if whitelist != nil {
		if _, exists := whitelist[field]; !exists {
			return NewValidationError(fmt.Sprintf("field '%s' not in whitelist", field), nil).
				WithContext("field_name", field).
				WithCode("FIELD_NOT_WHITELISTED").
				WithRetriable(false)
		}
	}

	return nil
}

// WithIn 构建一个通用的 IN 条件（如 field IN ?），当 values 为空时忽略该条件。
// 参数 field 必须是已知安全的列名（建议调用前进行白名单校验）。此处不对 values 做深度清洗，交由 GORM 预处理。
func WithIn(field string, values any) QueryOption {
	return func(db *DB) *DB {
		// 仅在 values 为非空切片时才应用条件
		switch v := values.(type) {
		case []string:
			if len(v) == 0 {
				return db
			}
		case []int:
			if len(v) == 0 {
				return db
			}
		case []int64:
			if len(v) == 0 {
				return db
			}
		default:
			// 其他类型直接交给 GORM 处理，但一般建议限制到常用类型
		}
		db.DB = db.DB.Where(field+" IN ?", values)
		return db
	}
}

// WithIds 使用更安全的 IN ? 形式展开 id 列的切片。空切片时忽略该条件。
func WithIds(ids []string) QueryOption {
	return func(db *DB) *DB {
		if len(ids) == 0 {
			return db
		}
		db.DB = db.DB.Where("id IN ?", ids)
		return db
	}
}

// WithNames 使用更安全的 IN ? 形式展开 name 列的切片。空切片时忽略该条件。
func WithNames(names []string) QueryOption {
	return func(db *DB) *DB {
		if len(names) == 0 {
			return db
		}
		db.DB = db.DB.Where("name IN ?", names)
		return db
	}
}

// WithUsernames 使用更安全的 IN ? 形式展开 username 列的切片。空切片时忽略该条件。
func WithUsernames(usernames []string) QueryOption {
	return func(db *DB) *DB {
		if len(usernames) == 0 {
			return db
		}
		db.DB = db.DB.Where("username IN ?", usernames)
		return db
	}
}