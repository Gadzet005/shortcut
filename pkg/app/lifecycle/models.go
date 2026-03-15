package lifecycle

import "context"

type App interface {
	// Неблокирующий запуск приложения
	OnRun(ctx Context) error
	// Вызывается при ошибке запуска приложения
	OnRunFailed(ctx Context, err error)
	// Вызывается при начале завершения работы приложения
	OnStopStarted(ctx Context)
	// Вызывается при ошибке завершения работы приложения
	OnStopFailed(ctx Context, err error)
	// Вызывается при успешном завершении работы приложения
	OnStopCompleted(ctx Context)
	// Возвращает контекст и функцию отмены для завершения работы приложения
	ShutdownContext(ctx Context) (context.Context, context.CancelFunc)
}

// Блокирующий запуск задачи
type Runner func(ctx context.Context)

// Остановка задачи
type Stopper func(ctx context.Context) error
