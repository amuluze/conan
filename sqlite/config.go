package sqlite

type Config struct {
	Debug        bool   // 是否开启调试模式，默认 false
	AutoMigrate  bool   // 是否自动迁移数据库结构，默认 false
	Type         string // 数据库类型，默认 sqlite
	DatabasePath string // 数据库文件路径，默认 :memory:
	MaxLifetime  int    // 连接最大生命周期，默认 5 分钟
	MaxOpenConns int    // 最大打开连接数，默认 100
	MaxIdleConns int    // 最大空闲连接数，默认 100
	BusyTimeout  int    // SQLite 忙等待超时时间，默认 5 秒
}