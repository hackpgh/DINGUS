// member.go

package models

type Member struct {
	TagId           uint32 // corresponds to RFIDFieldName in config
	MembershipLevel int
}
