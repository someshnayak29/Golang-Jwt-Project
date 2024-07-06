package helpers

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/someshnayak29/golang-jwt-project/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct {
	Email      string
	First_name string
	Last_name  string
	Uid        string
	User_type  string
	jwt.StandardClaims
}

// embedding jwt.StandardClaims into your custom claims struct, you can include both standard and custom claim data in your JWT tokens.

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var SECRET_KEY string = os.Getenv("SECRET_KEY")

func GenerateAllTokens(email string, firstName string, lastName string, userType string, uid string) (signedToken string, signedRefreshToken string, err error) {
	// claims is the detail with which token will be made from
	// expiresAt => time after which token expires, i.e. in our case 24 hrs after creation
	// refresh token to create new token i.e. after 168 hrs
	// newwithclaims func to create token, SigningMethodHS256 algo to create token, signed using secret key
	// Unix returns t as a Unix time, the number of seconds elapsed since January 1, 1970 UTC.

	claims := &SignedDetails{
		Email:      email,
		First_name: firstName,
		Last_name:  lastName,
		Uid:        uid,
		User_type:  userType,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}

	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		log.Panic(err)
		return
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		log.Panic(err)
		return
	}
	return token, refreshToken, err
}

func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {

	// It parses the signedToken and validates its signature using the SECRET_KEY.
	// If the token is valid and the signature is verified, it decodes the token payload into the SignedDetails
	// It validates the token's signature using the SECRET_KEY provided in the callback function.
	// If successful, it decodes the token payload into the provided SignedDetails struct.

	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)

	if err != nil {
		msg = err.Error()
		return
	}

	// The Claims field of *jwt.Token holds the decoded claims of the JWT token.
	// To access it use a type assertion to convert token.Claims to *SignedDetails.
	// ok: This boolean variable indicates whether the type assertion was successful

	claims, ok := token.Claims.(*SignedDetails)

	if !ok {
		msg = fmt.Sprintf("The Token is Invalid")
		//msg = err.Error()
		return
	}

	// check if token is expired or not
	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = fmt.Sprintf("Token is Expired")
		//msg = err.Error()
		return
	}

	// Otherwise token is correctly validated
	return claims, msg

}

// primitive.D in Go provides a convenient way to work with BSON documents in mongoDB
// primitive.E is key value pair

// The Upsert option in options. UpdateOptions specifies whether to perform an upsert operation.
// An upsert operation updates a document if it exists, or inserts it if it does not exist.

/* error : go.mongodb.org/mongo-driver/bson/primitive.E struct literal uses unkeyed fields
   solved using Key : "", Value:"" in bson.E and inside bson.D as all elements inside bson.D is bson.E */

func UpdateAllTokens(signedToken string, signedRefreshToken string, userId string) {

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	var updateObj primitive.D

	updateObj = append(updateObj, bson.E{Key: "token", Value: signedToken})
	updateObj = append(updateObj, bson.E{Key: "refresh_token", Value: signedRefreshToken})

	Updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{Key: "updated_at", Value: Updated_at})

	upsert := true
	filter := bson.M{"user_id": userId}
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	_, err := userCollection.UpdateOne(ctx, filter, bson.D{{Key: "$set", Value: updateObj}}, &opt)

	defer cancel()

	if err != nil {
		log.Panic(err)
		return
	}
	//return

}
