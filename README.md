JWT Authentication Backend using GoLang and Gin Gonic

This project demonstrates a full-fledged JWT authentication backend using GoLang and Gin Gonic framework with MongoDB Atlas for database storage.

Features

Signup: Register new users with unique usernames and encrypted passwords. JWT token is generated upon successful signup.
Login: Authenticate users and generate JWT tokens with expiration time.
Show Users: Fetch a list of all registered users (ADMIN & USER feature).
Show User by ID: Retrieve details of a specific user based on their ID (admin feature).

Technologies Used

GoLang: Backend language
Gin Gonic: HTTP web framework
MongoDB Atlas: NoSQL database
bcrypt: Hashing passwords securely
JWT: JSON Web Tokens for authentication with HMAC SHA-256 algorithm
Installation

Clone the repository:

git clone <repository_url>
cd <repository_directory>

Install dependencies:

Ensure you have Go installed. Then, install the required dependencies:

go mod tidy


Setup Environment Variables:

Create a .env file in the root directory and configure the following variables:

# MongoDB Atlas
MONGO_URL = < COPY_YOUR_URL>
PORT = 9000
SECRET_KEY = <COPY_YOUR_SECRET_KEY>

# Token Expiry

JWT_EXPIRE_MINUTES=60  # Example: Token expires in 60 minutes


Run the Application:

Start the application server:

go run main.go


The server will start running at http://localhost:9000.

API Endpoints

Authentication
POST /users/signup

Create a new user account and generate JWT token.


Body parameters: username, password
Header Parameter: token

POST /users/login

Authenticate and login a user. Generates JWT token.
Body parameters: username, password
User Management (Admin)
GET /users

Retrieve all users.
Requires admin privileges (JWT token with appropriate role).
GET /users/

Retrieve a user by ID.
Requires admin privileges (JWT token with appropriate role).
Contributing

Contributions are welcome! Fork the repository and submit a pull request for any enhancements.

License

This project is licensed under the MIT License - see the LICENSE file for details.
