// memberTrainingLink.go

package models

type MemberTrainingLink struct {
	TagID        uint32 // Foreign Key to Members (is an rfid)
	TrainingName string // Foreign Key to Trainings
}
