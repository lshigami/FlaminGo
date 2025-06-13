package enums

type AppointmentStatus string

const (
	Pending   AppointmentStatus = "pending"
	Confirmed AppointmentStatus = "confirmed"
	Cancelled AppointmentStatus = "cancelled"
	Completed AppointmentStatus = "completed"
)

func (s AppointmentStatus) IsValid() bool {
	switch s {
	case Pending, Confirmed, Cancelled, Completed:
		return true
	}
	return false
}
