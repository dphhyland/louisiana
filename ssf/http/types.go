package main

// Subject represents a subject in a stream
type Subject struct {
	Format      string `json:"format" bson:"format"`
	Email       string `json:"email,omitempty" bson:"email,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty" bson:"phone_number,omitempty"`
}
