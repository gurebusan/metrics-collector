package zapdiscard

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewDiscardLogger создает zap.Logger, который игнорирует все логи
func NewDiscardLogger() *zap.Logger {
	return zap.New(NewDiscardCore())
}

// DiscardCore реализует zapcore.Core для игнорирования логов
type DiscardCore struct{}

func NewDiscardCore() *DiscardCore {
	return &DiscardCore{}
}

func (c *DiscardCore) Enabled(zapcore.Level) bool {
	return false // Всегда отключаем логирование
}

func (c *DiscardCore) With([]zap.Field) zapcore.Core {
	return c // Возвращаем тот же core, так как ничего не сохраняем
}

func (c *DiscardCore) Check(zapcore.Entry, *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return nil // Ничего не записываем
}

func (c *DiscardCore) Write(zapcore.Entry, []zap.Field) error {
	return nil // Игнорируем запись
}

func (c *DiscardCore) Sync() error {
	return nil // Ничего не нужно синхронизировать
}
