package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	helper "github.com/someshnayak29/golang-jwt-project/helpers"
)

// gin.HandlerFunc is used to define middleware and route handlers.

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {

		clientToken := c.Request.Header.Get("token")
		if clientToken == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("No Authorization header provided")})
			c.Abort() // Abort prevents pending handlers from being called.
			return
		}
		claims, err := helper.ValidateToken(clientToken)
		if err != "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			c.Abort()
			return
		}

		// Now we will set logged users details in context

		c.Set("email", claims.Email)
		c.Set("first_name", claims.First_name)
		c.Set("last_name", claims.Last_name)
		c.Set("uid", claims.Uid)
		c.Set("user_type", claims.User_type)
		c.Next() // Next used only inside middleware. It executes the pending handlers in the chain inside the calling handler.
	}
}
