package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/someshnayak29/golang-jwt-project/database"
	helper "github.com/someshnayak29/golang-jwt-project/helpers"
	"github.com/someshnayak29/golang-jwt-project/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validate = validator.New() // To validate whether the user matches the description and fields of user struct

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14) // Converts the password string into a byte slice, as GenerateFromPassword
	// 14 is cost parameter or the "work factor" of bcrypt. It determines how computationally expensive the hashing will be
	// A higher cost value means more iterations of the bcrypt hashing algorithm
	if err != nil {
		log.Panic(err)
	}

	return string(bytes)
}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {

	// CompareHashAndPassword compares a bcrypt hashed password with its possible plaintext equivalent. Returns nil on success, or an error on failure.
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = fmt.Sprintf("password of email is incorrect")
		check = false
	}
	return check, msg
}

func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {

		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// To validate whether the user matches the description and fields of user struct
		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		// checking if email already exists in dB or not
		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		defer cancel() // to stop searching after 100 sec

		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the email"})
		}

		password := HashPassword(*user.Password)
		user.Password = &password

		// checking if phone already exists in dB or not
		count, err = userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		defer cancel()

		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for thr phone number"})
		}

		// if user already exists
		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this email or phone number already exists"})
		}

		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID() // generate a new unique ObjectID

		hex := user.ID.Hex() // direct not working, therefore first stored it int string, then use below
		user.User_id = &hex

		token, refreshToken, _ := helper.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, *user.User_type, *user.User_id)
		user.Token = &token
		user.Refresh_token = &refreshToken

		// Use fmt.Sprintf to format strings and capture the result.
		// Use fmt.Printf to format strings and print them directly to standard output.
		// mongodb will take _id and will return it,  if not then will create it using generate ObjectID and will be stored in resultInsertionNumber

		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {

			msg := fmt.Sprintf("User item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, resultInsertionNumber)

	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {

		var user models.User
		var foundUser models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "email or password is incorrect"})
			return
		}
		// email matched successfully now we will check if password is correct or not

		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		defer cancel()

		if !passwordIsValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		// we can add many more checks like this one: if user is not found in dB, then we can also prompt it to signup
		if foundUser.Email == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found!!! Kindly Sign Up "})
		}

		// now will generate a new token for the login users new session
		token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, *foundUser.User_type, *foundUser.User_id)

		// update both token and refreshToken in the user profile
		helper.UpdateAllTokens(token, refreshToken, *foundUser.User_id)

		// once again fetch user with updated values from dB using user_id
		err = userCollection.FindOne(ctx, bson.M{"user_id": *foundUser.User_id}).Decode(&foundUser)
		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, foundUser)

	}
}

// GetUsers can only be accessed by ADMIN

// strconv.Atoi to convert string to int
// Page number less than 1 doesn't make sense that's why default value is 1 and in error case also its also set to 1
// If parsing "startIndex" fails (err != nil), startIndex retains its value calculated from (page - 1) * recordPerPage.

func GetUsers() gin.HandlerFunc {

	return func(c *gin.Context) {

		if err := helper.CheckUserType(c, "ADMIN"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage")) // fetch from context request and convert to int
		// retrieve the value of a query parameter named "recordPerPage" from HTTP request handled by gin context c

		if err != nil || recordPerPage < 1 {
			recordPerPage = 10 // if we dont mention in context and default and error case also we take it as 10
		}

		page, err1 := strconv.Atoi(c.Query("page"))
		if err1 != nil || page < 1 {
			page = 1
		}

		// similar to skip and limit of node.js
		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex")) // if we specify in context starting index we will take it otherwise above calculated default page-1 * record will be used

		// MongoDB Aggregation Pipeline
		matchStage := bson.D{{Key: "$match", Value: bson.D{{}}}} // $match: Filters documents to pass only those that match the specified conditions.

		// _id : on what field we wants data to be grouped on; null value means all data in one single group
		// if we give value to _id, the it will group all unique ids together and give total docs under it using $sum
		// $group Stage: This is a MongoDB aggregation pipeline stage that groups documents
		// data Field: This stage creates a field data where each document in the group is added to an array.
		// Value: 1 specifies that for each document in the group, the accumulator should add 1 to the total count.

		groupStage := bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{{Key: "_id", Value: "null"}}},
			{Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}}}}}

		// $project allows you to manipulate and transform data within your MongoDB collection i.e. which all data points we want
		// Here, 0 means Excludes the _id field. 1 meanIncludes a field total_count. and Includes an array field user_items that is sliced from an array field data, based on startIndex and recordPerPage.
		// []interface This is an array of interfaces containing the arguments for the $slice operator:
		// The $ symbol is used in MongoDB query operators to indicate that the following string is a field name or a reference to a field within a document.

		projectStage := bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "total_count", Value: 1},
				{Key: "user_items", Value: bson.D{{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}}}}}}}

		result, err := userCollection.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage, projectStage})
		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing user items"})
		}

		// M is an unordered representation of a BSON document.
		// mongo driver returns mongo.Cursor, therefore we will store this in bson.M
		// result.All: Retrieves all documents from the aggregation result (result) and decodes them into the slice allUsers
		// All iterates the cursor and decodes each document into results.

		var allUsers []bson.M
		if err = result.All(ctx, &allUsers); err != nil {
			// Fatal is equivalent to [Print] followed by a call to os.Exit(1).
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, allUsers[0])

	}

}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("user_id") // context have every info regarding http request and user_id bcoz its used in url users/user_id

		// function to check if id is of admin or not
		if err := helper.MatchUserTypeToUid(c, userId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var user models.User
		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user) // user_id is from json of models
		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user) // send user details
	}

}
