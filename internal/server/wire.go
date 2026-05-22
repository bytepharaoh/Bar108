package server

import (
	"bar108/internal/handlers"
	jwtpkg "bar108/internal/jwt"
	"bar108/internal/repository"
	"bar108/internal/services"
	"database/sql"
)

func newMenuRepository(db *sql.DB) *repository.MenuRepository {
	return repository.NewMenuRepository(db)
}

func newUserRepository(db *sql.DB) *repository.UsersRepository {
	return repository.NewUserRepository(db)
}

func newMenuService(repo *repository.MenuRepository) services.MenuService {
	return services.NewMenuService(repo)
}

func newUserService(repo *repository.UsersRepository) services.UserService {
	return services.NewUserService(repo)
}

func newOrderRepository(db *sql.DB) repository.OrderRepository {
	return repository.NewOrderRepository(db)
}

func newOrderService(repo repository.OrderRepository) services.OrderService {
	return services.NewOrderService(repo)
}

func newOrderHandler(svc services.OrderService) *handlers.OrderHandler {
	return handlers.NewOrderHandler(svc)
}

func newAuthService(store *repository.UsersRepository, jwtManager *jwtpkg.Manager) services.AuthService {
	return services.NewAuthService(store, jwtManager)
}
