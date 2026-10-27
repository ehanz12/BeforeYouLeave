package services

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ehanz12/BeforeYouLeave/databases"
	"github.com/ehanz12/BeforeYouLeave/dtos/request"
	"github.com/ehanz12/BeforeYouLeave/models"
)

// normalizeTime = terima "07:00" atau "07:00:00", selalu simpan sebagai "15:04:05".
func normalizeTime(field, input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("%s harus diisi !", field)
	}
	for _, layout := range []string{"15:04", "15:04:05"} {
		if t, err := time.Parse(layout, input); err == nil {
			return t.Format("15:04:05"), nil
		}
	}
	return "", fmt.Errorf("format %s %q tidak valid (HH:MM atau HH:MM:SS) !", field, input)
}

func validateDayOfWeek(day int8) error {
	if day < 0 || day > 6 {
		return errors.New("day_of_week harus antara 0 (Minggu) dan 6 (Sabtu) !")
	}
	return nil
}

func validateRemindDMinus(d int8) error {
	if d != models.RemindSameDay && d != models.RemindHMinus1 {
		return errors.New("remind_d_minus harus 0 (hari-H) atau 1 (H-1 malam) !")
	}
	return nil
}

// checkSameDayOrder = untuk remind_d_minus=0 (hari-H), jam reminder harus lebih awal
// dari jam berangkat, kalau tidak user baru ingat.pack setelah berangkat.
// Kedua jam sudah ternormalisasi "15:04:05", jadi compare string cukup.
func checkSameDayOrder(dMinus int8, remind, departure string) error {
	if dMinus == models.RemindSameDay && remind >= departure {
		return fmt.Errorf(
			"remind_time (%s) harus lebih awal dari departure_time (%s) saat remind_d_minus = 0 (hari-H) !",
			remind, departure,
		)
	}
	return nil
}

// scheduleInput = hasil validasi request sebelum disimpan.
type scheduleInput struct {
	DayOfWeek     int8
	DepartureTime string
	RemindDMinus  int8
	RemindTime    string
	IsActive      bool
}

func buildScheduleInput(day int8, departureTime string, remindDMinus *int8, remindTime string, isActive *bool) (scheduleInput, error) {
	if err := validateDayOfWeek(day); err != nil {
		return scheduleInput{}, err
	}

	departure, err := normalizeTime("departure_time", departureTime)
	if err != nil {
		return scheduleInput{}, err
	}
	remind, err := normalizeTime("remind_time", remindTime)
	if err != nil {
		return scheduleInput{}, err
	}

	dMinus := int8(models.RemindHMinus1)
	if remindDMinus != nil {
		dMinus = *remindDMinus
	}
	if err := validateRemindDMinus(dMinus); err != nil {
		return scheduleInput{}, err
	}
	if err := checkSameDayOrder(dMinus, remind, departure); err != nil {
		return scheduleInput{}, err
	}

	// Untuk hari-H, reminder harus bunyi sebelum berangkat. Kalau tidak, user baru
	// ingat.pack setelah berangkat. Kedua jam sudah ternormalisasi "15:04:05",
	// jadi perbandingan string sama dengan perbandingan waktu.
	if dMinus == models.RemindSameDay && remind >= departure {
		return scheduleInput{}, fmt.Errorf(
			"remind_time (%s) harus lebih awal dari departure_time (%s) saat remind_d_minus = 0 (hari-H) !",
			remind, departure,
		)
	}

	active := true
	if isActive != nil {
		active = *isActive
	}

	return scheduleInput{
		DayOfWeek:     day,
		DepartureTime: departure,
		RemindDMinus:  dMinus,
		RemindTime:    remind,
		IsActive:      active,
	}, nil
}

// dayTaken = satu destination hanya boleh punya satu baris per hari.
func dayTaken(destinationID uint64, day int8, excludeID uint64) bool {
	var s models.DestinationSchedule
	query := databases.DB.
		Select("id").
		Where("destination_id = ? AND day_of_week = ?", destinationID, day)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	return query.First(&s).Error == nil
}

func CreateSchedule(req request.CreateScheduleRequest, destinationID, userID uint64) (models.DestinationSchedule, error) {
	if _, err := getOwnedDestination(destinationID, userID); err != nil {
		return models.DestinationSchedule{}, err
	}

	in, err := buildScheduleInput(req.DayOfWeek, req.DepartureTime, req.RemindDMinus, req.RemindTime, req.IsActive)
	if err != nil {
		return models.DestinationSchedule{}, err
	}
	if dayTaken(destinationID, in.DayOfWeek, 0) {
		return models.DestinationSchedule{}, errors.New("jadwal untuk hari ini sudah ada")
	}

	schedule := models.DestinationSchedule{
		DestinationID: destinationID,
		DayOfWeek:     in.DayOfWeek,
		DepartureTime: in.DepartureTime,
		RemindDMinus:  in.RemindDMinus,
		RemindTime:    in.RemindTime,
		IsActive:      in.IsActive,
	}
	if err := databases.DB.Create(&schedule).Error; err != nil {
		return models.DestinationSchedule{}, errors.New("terjadi kesalahan sistem")
	}
	return schedule, nil
}

func ListSchedules(req request.ListScheduleRequest, destinationID, userID uint64, page, limit int) ([]models.DestinationSchedule, int64, error) {
	if _, err := getOwnedDestination(destinationID, userID); err != nil {
		return nil, 0, err
	}

	query := databases.DB.Model(&models.DestinationSchedule{}).Where("destination_id = ?", destinationID)
	if req.IsActive != nil {
		query = query.Where("is_active = ?", *req.IsActive)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.New("terjadi kesalahan sistem !")
	}

	var schedules []models.DestinationSchedule
	offset := (page - 1) * limit
	if err := query.
		Order("day_of_week ASC").
		Offset(offset).
		Limit(limit).
		Find(&schedules).Error; err != nil {
		return nil, 0, errors.New("terjadi kesalahan sistem !")
	}
	return schedules, total, nil
}

// getOwnedSchedule = schedule harus milik destination itu, dan destination-nya milik user.
func getOwnedSchedule(scheduleID, destinationID, userID uint64) (models.DestinationSchedule, error) {
	if _, err := getOwnedDestination(destinationID, userID); err != nil {
		return models.DestinationSchedule{}, err
	}
	var s models.DestinationSchedule
	if err := databases.DB.
		Where("id = ? AND destination_id = ?", scheduleID, destinationID).
		First(&s).Error; err != nil {
		return models.DestinationSchedule{}, errors.New("jadwal tidak ditemukan !")
	}
	return s, nil
}

func DetailSchedule(scheduleID, destinationID, userID uint64) (models.DestinationSchedule, error) {
	return getOwnedSchedule(scheduleID, destinationID, userID)
}

func UpdateSchedule(req request.UpdateScheduleRequest, scheduleID, destinationID, userID uint64) (models.DestinationSchedule, error) {
	schedule, err := getOwnedSchedule(scheduleID, destinationID, userID)
	if err != nil {
		return models.DestinationSchedule{}, err
	}

	updates := map[string]interface{}{}

	if req.DayOfWeek != nil {
		if err := validateDayOfWeek(*req.DayOfWeek); err != nil {
			return models.DestinationSchedule{}, err
		}
		if dayTaken(destinationID, *req.DayOfWeek, schedule.ID) {
			return models.DestinationSchedule{}, errors.New("jadwal untuk hari ini sudah ada")
		}
		updates["day_of_week"] = *req.DayOfWeek
	}
	if req.DepartureTime != nil {
		departure, err := normalizeTime("departure_time", *req.DepartureTime)
		if err != nil {
			return models.DestinationSchedule{}, err
		}
		updates["departure_time"] = departure
	}
	if req.RemindTime != nil {
		remind, err := normalizeTime("remind_time", *req.RemindTime)
		if err != nil {
			return models.DestinationSchedule{}, err
		}
		updates["remind_time"] = remind
	}
	if req.RemindDMinus != nil {
		if err := validateRemindDMinus(*req.RemindDMinus); err != nil {
			return models.DestinationSchedule{}, err
		}
		updates["remind_d_minus"] = *req.RemindDMinus
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) == 0 {
		return models.DestinationSchedule{}, errors.New("tidak ada field yang diupdate !")
	}

	// Validasi ulang memakai nilai akhir, bukan cuma field yang dikirim.
	// Contoh: remind_time diubah jadi 20:00 pada jadwal hari-H yang berangkat 07:00.
	finalDMinus := schedule.RemindDMinus
	if req.RemindDMinus != nil {
		finalDMinus = *req.RemindDMinus
	}
	finalRemind := schedule.RemindTime
	if req.RemindTime != nil {
		finalRemind = updates["remind_time"].(string)
	}
	finalDeparture := schedule.DepartureTime
	if req.DepartureTime != nil {
		finalDeparture = updates["departure_time"].(string)
	}
	if err := checkSameDayOrder(finalDMinus, finalRemind, finalDeparture); err != nil {
		return models.DestinationSchedule{}, err
	}

	if err := databases.DB.Model(&schedule).Updates(updates).Error; err != nil {
		return models.DestinationSchedule{}, errors.New("terjadi kesalahan sistem !")
	}
	if err := databases.DB.First(&schedule, schedule.ID).Error; err != nil {
		return models.DestinationSchedule{}, errors.New("terjadi kesalahan sistem !")
	}
	return schedule, nil
}

func DeleteSchedule(scheduleID, destinationID, userID uint64) error {
	schedule, err := getOwnedSchedule(scheduleID, destinationID, userID)
	if err != nil {
		return err
	}
	if err := databases.DB.Delete(&schedule).Error; err != nil {
		return errors.New("terjadi kesalahan sistem !")
	}
	return nil
}

// ReplaceSchedules = set jadwal sepekan dalam satu kali call (replace all).
// Dipakai form FE yang mencentang hari aktif sekaligus, misal Senin-Jumat = 5 rows.
// Semua-or-nothing: kalau ada satu baris invalid, tidak ada yang tersimpan.
func ReplaceSchedules(req request.ReplaceSchedulesRequest, destinationID, userID uint64) ([]models.DestinationSchedule, error) {
	if _, err := getOwnedDestination(destinationID, userID); err != nil {
		return nil, err
	}
	if len(req.Schedules) == 0 {
		return nil, errors.New("schedules tidak boleh kosong !")
	}

	inputs := make([]scheduleInput, 0, len(req.Schedules))
	seen := make(map[int8]bool, len(req.Schedules))
	for i, s := range req.Schedules {
		if err := validateDayOfWeek(s.DayOfWeek); err != nil {
			return nil, fmt.Errorf("schedules[%d]: %w", i, err)
		}
		if seen[s.DayOfWeek] {
			return nil, fmt.Errorf("schedules[%d]: hari %d dobel di dalam request !", i, s.DayOfWeek)
		}
		seen[s.DayOfWeek] = true

		in, err := buildScheduleInput(s.DayOfWeek, s.DepartureTime, s.RemindDMinus, s.RemindTime, s.IsActive)
		if err != nil {
			return nil, fmt.Errorf("schedules[%d]: %w", i, err)
		}
		inputs = append(inputs, in)
	}

	tx := databases.DB.Begin()
	if tx.Error != nil {
		return nil, errors.New("terjadi kesalahan sistem !")
	}

	if err := tx.Where("destination_id = ?", destinationID).Delete(&models.DestinationSchedule{}).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("terjadi kesalahan sistem !")
	}

	schedules := make([]models.DestinationSchedule, 0, len(inputs))
	for _, in := range inputs {
		s := models.DestinationSchedule{
			DestinationID: destinationID,
			DayOfWeek:     in.DayOfWeek,
			DepartureTime: in.DepartureTime,
			RemindDMinus:  in.RemindDMinus,
			RemindTime:    in.RemindTime,
			IsActive:      in.IsActive,
		}
		if err := tx.Create(&s).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("terjadi kesalahan sistem !")
		}
		schedules = append(schedules, s)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("terjadi kesalahan sistem !")
	}

	sort.Slice(schedules, func(i, j int) bool {
		return schedules[i].DayOfWeek < schedules[j].DayOfWeek
	})
	return schedules, nil
}
