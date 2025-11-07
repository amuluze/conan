package sqlite

import "strings"

// QueryOption 定义对 *DB 进行链式包装的函数类型，返回经变更后的 *DB，便于组合多个查询配置。
// 所有实现都应满足幂等与安全（避免 SQL 注入）的要求。
type QueryOption func(db *DB) *DB

// OptionDB 按序应用一组 QueryOption 到传入的 *DB 并返回更新后的 *DB。
// 为提升健壮性：
// 1) 当 db 为 nil 时直接返回 nil，避免后续调用产生 panic；
// 2) 跳过为 nil 的选项，保证调用安全。
func OptionDB(db *DB, options ...QueryOption) *DB {
	if db == nil {
		return nil
	}
	for _, option := range options {
		if option == nil {
			continue
		}
		db = option(db)
	}
	return db
}

// WithTable 设置查询所使用的表名。
// 注意：tableName 需来源于受控白名单以防止 SQL 注入，这里仅进行空值与空白过滤。
func WithTable(tableName string) QueryOption {
	return func(db *DB) *DB {
		t := strings.TrimSpace(tableName)
		if t == "" {
			return db
		}
		db.DB = db.DB.Table(t)
		return db
	}
}

// WithId 按主键 id 追加 WHERE 条件（id = ?）。当 id 为空或仅包含空白时忽略该条件。
func WithId(id string) QueryOption {
	return func(db *DB) *DB {
		if strings.TrimSpace(id) == "" {
			return db
		}
		db.DB = db.DB.Where("id = ?", id)
		return db
	}
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
		if _, ok := whitelist[f]; !ok {
			// 非白名单字段直接忽略或记录告警
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
		if _, ok := whitelist[f]; !ok {
			return db
		}
		db.DB = db.DB.Order(f + " DESC")
		return db
	}
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