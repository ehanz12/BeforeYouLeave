package handlers

import (
	"context"
	"log"
	"time"

	"github.com/ehanz12/BeforeYouLeave/configs"
	"github.com/ehanz12/BeforeYouLeave/databases"
	"github.com/ehanz12/BeforeYouLeave/dtos/request"
	"github.com/ehanz12/BeforeYouLeave/mappers"
	"github.com/ehanz12/BeforeYouLeave/models"
	"github.com/ehanz12/BeforeYouLeave/utils"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
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
		// Jangan bocorkan detail internal ke client, cukup log di server.
		log.Printf("⚠️ gagal membuat JWT untuk user %d: %v", exits.ID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "gagal membuat token"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "message": "Berhasil Login", "access_token": access_token})
}

func GoogleAuth(c *fiber.Ctx) error {
	var r request.GoogleAuthRequest
	if err := c.BodyParser(&r); err != nil || r.IDToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "id_token google wajib diisi"})
	}

	if configs.AppConfig.GoogleClientID == "" {
		log.Println("⚠️ GOOGLE_CLIENT_ID belum di-set, verifikasi id_token tidak bisa dilakukan")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "google auth belum dikonfigurasi"})
	}

	// Validasi tanda tangan + audience id_token ke Google.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	payload, err := idtoken.Validate(ctx, r.IDToken, configs.AppConfig.GoogleClientID)
	if err != nil {
		log.Printf("⚠️ id_token google tidak valid: %v", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "id_token google tidak valid"})
	}

	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)
	if email == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "email tidak ditemukan di id_token"})
	}
	if verified, ok := payload.Claims["email_verified"].(bool); ok && !verified {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "email google belum diverifikasi"})
	}
	if name == "" {
		name = email
	}

	// Auto-login / auto-register berdasarkan email dari token yang sudah tervalidasi.
	var user models.User
	if err := databases.DB.Where("email = ?", email).First(&user).Error; err != nil {
		user = models.User{
			Name:     name,
			Email:    email,
			Password: "",
		}
		if err := databases.DB.Create(&user).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "gagal mendaftarkan user google"})
		}
	}

	token, err := utils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		log.Printf("⚠️ gagal membuat JWT untuk user google %d: %v", user.ID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "gagal membuat token jwt"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success":      true,
		"message":      "Berhasil Login via Google",
		"access_token": token,
		"data":         mappers.ToDoUser(user),
	})
}

func Me(c *fiber.Ctx) error {
	userID := c.Locals("user_id")

	var user models.User
	if err := databases.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "user tidak ditemukan !"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "Message": "user berhasil di dapat !", "data": user})
}
