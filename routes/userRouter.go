package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/someshnayak29/golang-jwt-project/controllers"
	"github.com/someshnayak29/golang-jwt-project/middleware"
)

func UserRoutes(incomingRoutes *gin.Engine) {

	incomingRoutes.Use(middleware.Authenticate())
	/* we are using middleware because they are protected routes, earlier while login signup we didnt had token, bt after
	logging in we have token, therefore we have used middleware, user should not be allowed to use userRoutes without token*/
	incomingRoutes.GET("/users", controller.GetUsers())
	incomingRoutes.GET("/users/:user_id", controller.GetUser())

}
