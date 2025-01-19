package routes

import (
	"net/http"

	"github.com/mysterybee07/go-sessions-auth/controllers"
)

func Setup() {
	http.HandleFunc("/login", controllers.Login)

}
