package service

import (
	"sort"
	"time"

	"hrattendance/internal/model"
	"hrattendance/pkg/idgen"
)

func (s *Service) CreateAttendance(attendance model.Attendance) (*model.Attendance, error) {
	if err := attendance.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetEmployee(attendance.EmployeeID); err != nil {
		return nil, model.NewValidationError("employee_id", "关联的员工不存在")
	}
	attendance.Status = s.evaluateStatus(attendance.CheckIn, attendance.CheckOut)
	now := time.Now()
	attendance.ID = idgen.Hex()
	attendance.CreatedAt = now
	attendance.UpdatedAt = now
	if err := s.store.CreateAttendance(&attendance); err != nil {
		return nil, err
	}
	return &attendance, nil
}

// CheckIn 为员工签到，若当日记录不存在则创建，存在则补签到时间。
func (s *Service) CheckIn(employeeID, date, checkIn string) (*model.Attendance, error) {
	if _, err := s.store.GetEmployee(employeeID); err != nil {
		return nil, err
	}
	if existing, err := s.store.GetAttendanceByEmployeeDate(employeeID, date); err == nil {
		existing.CheckIn = checkIn
		existing.Status = s.evaluateStatus(existing.CheckIn, existing.CheckOut)
		existing.UpdatedAt = time.Now()
		if err := s.store.UpdateAttendance(existing); err != nil {
			return nil, err
		}
		return existing, nil
	}
	return s.CreateAttendance(model.Attendance{
		EmployeeID: employeeID,
		Date:       date,
		CheckIn:    checkIn,
	})
}

// CheckOut 为员工签退，若当日记录不存在则创建，存在则补签退时间。
func (s *Service) CheckOut(employeeID, date, checkOut string) (*model.Attendance, error) {
	if _, err := s.store.GetEmployee(employeeID); err != nil {
		return nil, err
	}
	if existing, err := s.store.GetAttendanceByEmployeeDate(employeeID, date); err == nil {
		existing.CheckOut = checkOut
		existing.Status = s.evaluateStatus(existing.CheckIn, existing.CheckOut)
		existing.UpdatedAt = time.Now()
		if err := s.store.UpdateAttendance(existing); err != nil {
			return nil, err
		}
		return existing, nil
	}
	return s.CreateAttendance(model.Attendance{
		EmployeeID: employeeID,
		Date:       date,
		CheckOut:   checkOut,
	})
}

func (s *Service) GetAttendance(id string) (*model.Attendance, error) {
	return s.store.GetAttendance(id)
}

func (s *Service) ListAttendances(filter model.AttendanceFilter, page, size int) ([]*model.Attendance, int, error) {
	all := s.store.ListAttendances()
	matched := make([]*model.Attendance, 0, len(all))
	for _, a := range all {
		if filter.Match(a) {
			matched = append(matched, a)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		if matched[i].Date == matched[j].Date {
			return matched[i].EmployeeID < matched[j].EmployeeID
		}
		return matched[i].Date < matched[j].Date
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Attendance{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// UpdateAttendance 补卡：修正签到/签退时间并重新判定状态。
func (s *Service) UpdateAttendance(id string, input model.Attendance) (*model.Attendance, error) {
	existing, err := s.store.GetAttendance(id)
	if err != nil {
		return nil, err
	}
	if input.CheckIn != "" {
		existing.CheckIn = input.CheckIn
	}
	if input.CheckOut != "" {
		existing.CheckOut = input.CheckOut
	}
	if input.Note != "" {
		existing.Note = input.Note
	}
	if err := existing.Validate(); err != nil {
		return nil, err
	}
	existing.Status = s.evaluateStatus(existing.CheckIn, existing.CheckOut)
	existing.UpdatedAt = time.Now()
	if err := s.store.UpdateAttendance(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *Service) DeleteAttendance(id string) error {
	return s.store.DeleteAttendance(id)
}

// AttendanceBatchResult 批量导入考勤结果。
type AttendanceBatchResult struct {
	Succeeded []*model.Attendance `json:"succeeded"`
	Failed    map[int]string      `json:"failed"`
}

// BatchCreateAttendances 批量导入考勤记录，逐条校验并汇总成功与失败。
func (s *Service) BatchCreateAttendances(attendances []model.Attendance) *AttendanceBatchResult {
	result := &AttendanceBatchResult{
		Succeeded: make([]*model.Attendance, 0),
		Failed:    make(map[int]string),
	}
	for i, a := range attendances {
		created, err := s.CreateAttendance(a)
		if err != nil {
			result.Failed[i] = err.Error()
			continue
		}
		result.Succeeded = append(result.Succeeded, created)
	}
	return result
}

// evaluateStatus 根据考勤规则判定打卡状态。
func (s *Service) evaluateStatus(checkIn, checkOut string) string {
	rule := s.ActiveRule()
	workStart := rule.WorkStart
	workEnd := rule.WorkEnd
	lateThreshold := rule.LateThreshold

	if checkIn == "" && checkOut == "" {
		return model.AttendanceAbsent
	}

	if checkIn != "" {
		start, err := time.Parse("15:04", workStart)
		if err == nil {
			lateAfter := start.Add(time.Duration(lateThreshold) * 0)
			if t, err := time.Parse("15:04", checkIn); err == nil && t.After(lateAfter) {
				return model.AttendanceLate
			}
		}
	}

	if checkOut != "" {
		end, err := time.Parse("15:04", workEnd)
		if err == nil {
			if t, err := time.Parse("15:04", checkOut); err == nil && t.Before(end) {
				return model.AttendanceEarlyLeave
			}
		}
	}

	return model.AttendanceNormal
}
