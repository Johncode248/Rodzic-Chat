package groupteacherspg

import "time"

// GroupTeacher is a group <-> teacher membership row. A Group also carries
// its own single TeacherId field (the "primary" teacher, set at group
// creation or via a teacher-invite's group_id) -- this table is what lets a
// group have additional teachers beyond that one, without disturbing any
// existing code that reads Group.TeacherId. A group's full teacher set is
// therefore the union of Group.TeacherId (if set) and these rows.
type GroupTeacher struct {
	Id         string    `json:"id,omitempty" gorm:"primaryKey"`
	GroupId    string    `json:"group_id,omitempty" gorm:"not null;index"`
	TeacherId  string    `json:"teacher_id,omitempty" gorm:"not null;index"`
	CreateTime time.Time `json:"create_time,omitempty" gorm:"not null"`
}

type CreateRequest struct {
	GroupId   string `json:"group_id,omitempty"`
	TeacherId string `json:"teacher_id,omitempty"`
}

type CreateResponse struct {
	GroupTeacher GroupTeacher `json:"group_teacher,omitempty"`
}

type SelectResponse struct {
	GroupTeachers []GroupTeacher `json:"group_teachers,omitempty"`
}

type DeleteResponse struct {
	GroupTeachers []GroupTeacher `json:"group_teachers,omitempty"`
}
