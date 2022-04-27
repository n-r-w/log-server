// Package domain
package domain

import "github.com/n-r-w/log-server/internal/domain/usecase"

// Domain - контейнер для набора текущих реализаций интерфейсов сценариев
// Инициализируется на старте путем выбора нужных реализаций в зависимости
// от необходимости обычной работы, юнит-тестов и т.п.
type Domain struct {
	LogUsecase  usecase.LogInterface
	UserUsecase usecase.UserInterface
}

// NewDomain - Создание объекта Domain
func NewDomain(
	logUsecase usecase.LogInterface,
	userUsecase usecase.UserInterface) *Domain {
	return &Domain{
		LogUsecase:  logUsecase,
		UserUsecase: userUsecase,
	}
}
