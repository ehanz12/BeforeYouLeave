package handlers

import (
	"github.com/ehanz12/BeforeYouLeave/databases"
	"github.com/ehanz12/BeforeYouLeave/dtos/request"
	"github.com/ehanz12/BeforeYouLeave/mappers"
	"github.com/ehanz12/BeforeYouLeave/models"
	"github.com/ehanz12/BeforeYouLeave/utils"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *fiber.Ctx) error {
	var r request.RegisterRequest
	if err := c.BodyParser(&r); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "terjadi kesalahan request"})
	}
	if r.Name == "" || r.Email == "" || r.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "name, email, password wajib di isi !"})
	}
	var u models.User
	if err := databases.DB.Select("id", "email").Where("email = ?", r.Email).First(&u).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "email sudah digunakan"})
	}

	passwordHansh, err := bcrypt.GenerateFromPassword([]byte(r.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "terjadi kesalahan konversi password"})
	}
	user := models.User{
		Name:     r.Name,
		Email:    r.Email,
		Password: string(passwordHansh),
		Phone:    &r.Phone,
	}

	if err := databases.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "terjadi kesalahan sistem"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "berhasil registrasi",
		"data":    mappers.ToDoUser(user),
	})
}

func Login(c *fiber.Ctx) error {
	var r request.LoginRequest
	if err := c.BodyParser(&r); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "terjadi kesalahan request"})
	}
	var exits models.User
	if err := databases.DB.Select("id", "email", "password").Where("email = ?", r.Email).First(&exits).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "email tidak tersedia harap registrasi !"})
	}
	if err := bcrypt.CompareHashAndPassword([]byte(exits.Password), []byte(r.Password)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "password salah !"})
	}

	access_token, err := utils.GenerateJWT(exits.ID, exits.Email)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "terjadi kesalahan sistem"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "message": "Berhasil Login", "access_token": access_token})
}

func Me(c *fiber.Ctx) error {
	userID := c.Locals("user_id")

	var user models.User
	if err := databases.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "user tidak ditemukan !"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "Message": "user berhasil di dapat !", "data": user})
}
