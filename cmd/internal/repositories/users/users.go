package users

import "project_pay/cmd/internal/models"

type Users struct {
	users []models.User
}

func New() *Users {
	//toda vez que alguém inicializar essa camada
	return &Users{users: make([]models.User, 0)}
}

func (u Users) GetAll() []models.User {
	// pega todos os usuários armazenados dentro de users
	return u.users
}

//tem que alterar a estrutura, tem que falar que está alterando a ref(*)
func (u *Users) Add(newUser models.User) {
	u.users = append(u.users, newUser)
}

func (u Users) EmailExists(email string) bool {
	for _, v := range u.users {
		if v.Email == email {
			return true
		}
	}
	return false
}
