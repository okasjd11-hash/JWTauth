package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func connectDB() {
	MongoURI := "mongodb://localhost:27017"
	err := mgm.SetDefaultConfig(nil, "E-commerce", options.Client().ApplyURI(MongoURI))
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("Connected to MongoDB!")

}

type Products struct {
	mgm.DefaultModel `bson:",inline"`
	ID               primitive.ObjectID `json:"id" bson:"_id,omitempty"` // Capitalized ID
	Name             string             `json:"name" bson:"name"`
	Price            float64            `json:"price" bson:"price"`
	Description      string             `json:"description" bson:"description"`
}
type User struct {
	mgm.DefaultModel `bson:",inline"`
	ID               primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name             string             `json:"name" bson:"name"`
	Email            string             `json:"email" bson:"email"`
	Password         string             `json:"password" bson:"password"`
}

func HarshPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
func CheckUserdata() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ValidateUsers := User{}
		err := c.BodyParser(&ValidateUsers)
		if err != nil {
			return err
		}
		if ValidateUsers.Email == "" || ValidateUsers.Password == "" || ValidateUsers.Name == "" {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"message": "user data is empty",
			})
		}
		return c.Next()
	}

}
func CheckProductsData() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ValidateProducts := Products{}
		if err := c.BodyParser(&ValidateProducts); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})

		}

		if ValidateProducts.Price < 0 {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"error":   "Price Can't be negative",
			})

		}

		if ValidateProducts.Name == "" || ValidateProducts.Description == "" {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"error":   "Product Name and Description can't be empty",
			})

		}
		return c.Next()
	}

}
func main() {
	connectDB()

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Backend Server is running!")
	})
	app.Post("/SendProducts", CheckProductsData(), func(c *fiber.Ctx) error {
		NewProducts := new(Products)
		if err := c.BodyParser(NewProducts); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"error":   "Error parsing request body" + err.Error(),
			})
		}

		mgm.Coll(NewProducts).Create(NewProducts)

		return c.Status(200).JSON(fiber.Map{
			"success":  true,
			"products": NewProducts,
		})
	})
	app.Post("/register", CheckUserdata(), func(c *fiber.Ctx) error {
		NewUser := new(User)

		if err := c.BodyParser(NewUser); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"error":   "Error parsing request body" + err.Error(),
			})
		}
		//Harshing the userPass before Saving it to Database
		hashedPassword, err := HarshPassword(NewUser.Password)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"error":   "Failed to hash password",
			})
		}

		NewUser.Password = hashedPassword

		mgm.Coll(NewUser).Create(NewUser)
		return c.Status(200).JSON(fiber.Map{
			"success": true,
			"data":    "user User Registered" + NewUser.Name,
		})

	})

	log.Fatal(app.Listen(":3000"))

}
