package server

import (
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
