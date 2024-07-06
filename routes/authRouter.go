package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/someshnayak29/golang-jwt-project/controllers"
)

func AuthRoutes(incomingRoutes *gin.Engine) {

	// these are public routes and accessible by everyone, therefore no middleware needed to check token validacy
	incomingRoutes.POST("users/signup", controller.Signup()) // same as usual routes =>  endpt., function()
	incomingRoutes.POST("users/login", controller.Login())
}
