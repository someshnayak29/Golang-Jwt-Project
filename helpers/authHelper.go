package helpers

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func CheckUserType(c *gin.Context, role string) (err error) {
	userType := c.GetString("user_type")
	err = nil

	// checks if the user has the overall authority or permission based on their role ("USER", "ADMIN").
	if userType != role {
		err = errors.New("unauthorized to access this resource")
		return err
	}
	return err
}

func MatchUserTypeToUid(c *gin.Context, userId string) (err error) {

	//userId is from url and uid is from request

	// Fetch user_type and uid from context which was set during middleware verification of token
	userType := c.GetString("user_type") // two possible ADMIN or USER check userModel.go
	uid := c.GetString("uid")            //
	err = nil

	// ensures that a user with type "USER" can only access resources
	// associated with their own userId. This is a specific access control rule.
	if userType == "USER" && uid != userId {
		err = errors.New("unauthorized to access this resource")
		return err
	}

	err = CheckUserType(c, userType)
	return err

}
