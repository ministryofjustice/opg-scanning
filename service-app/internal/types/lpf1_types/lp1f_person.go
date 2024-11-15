package lpf1_types

import "time"

// Person represents a person involved in the form and is manually composed from various sources.
type Person struct {
	ID            *int
	CaseRecNumber *string
	UID           *string
	Email         *string
	DOB           *string
	DateOfDeath   *string
	Title         *string
	FirstName     *string
	MiddleNames   *string
	LastName      *string
	DisplayName   *string
	OtherNames    *string
	Address       PersonAddress
	PhoneNumber   *string
	Occupation    *string
	CreatedDate   *time.Time
	UpdatedDate   *time.Time
	PersonType    *string

	// Relationships
	Children       []Person
	Parent         *Person
	Addresses      []PersonAddress
	PhoneNumbers   []PersonPhoneNumber
	Tasks          []PersonTask
	Warnings       []PersonWarning
	Notes          []PersonNote
	Documents      []PersonDocument
	Investigations []PersonInvestigation
}

// PersonAddress represents an address of a person.
type PersonAddress struct {
	AddressLine1 string
	AddressLine2 string
	AddressLine3 string
	Town         string
	County       string
	Postcode     string
	Country      string
	IsAirmail    bool
}

type PersonPhoneNumber struct {
	Number string
	Type   string
}

type PersonTask struct {
	ID      int
	Details string
}

type PersonWarning struct {
	Type    string
	Message string
}

type PersonNote struct {
	Content string
}

type PersonDocument struct {
	Title     string
	Type      string
	Direction string
	IsDraft   bool
}

type PersonInvestigation struct {
	Details string
}

// Provides a unified way to access person data from the document.
func (doc *LP1FDocument) GetPersons() []Person {
	persons := []Person{}

	// Page1, Section1
	if doc.Page1.Section1.FirstName != "" || doc.Page1.Section1.LastName != "" {
		persons = append(persons, Person{
			Title:     &doc.Page1.Section1.Title,
			FirstName: &doc.Page1.Section1.FirstName,
			LastName:  &doc.Page1.Section1.LastName,
			DOB:       &doc.Page1.Section1.DOB,
			Address: PersonAddress{
				AddressLine1: doc.Page1.Section1.Address.Address1,
				AddressLine2: doc.Page1.Section1.Address.Address2,
				AddressLine3: doc.Page1.Section1.Address.Address3,
				Postcode:     doc.Page1.Section1.Address.Postcode,
			},
			Email: &doc.Page1.Section1.EmailAddress,
		})
	}

	// Page2, Section2 (Attorney1 and Attorney2)
	if doc.Page2.Section2.Attorney1.FirstName != "" || doc.Page2.Section2.Attorney1.LastName != "" {
		persons = append(persons, Person{
			Title:     &doc.Page2.Section2.Attorney1.Title,
			FirstName: &doc.Page2.Section2.Attorney1.FirstName,
			LastName:  &doc.Page2.Section2.Attorney1.LastName,
			DOB:       &doc.Page2.Section2.Attorney1.DOB,
			Address: PersonAddress{
				AddressLine1: doc.Page2.Section2.Attorney1.Address.Address1,
				AddressLine2: doc.Page2.Section2.Attorney1.Address.Address2,
				AddressLine3: doc.Page2.Section2.Attorney1.Address.Address3,
				Postcode:     doc.Page2.Section2.Attorney1.Address.Postcode,
			},
			Email: &doc.Page2.Section2.Attorney1.EmailAddress,
		})
	}

	if doc.Page2.Section2.Attorney2.FirstName != "" || doc.Page2.Section2.Attorney2.LastName != "" {
		persons = append(persons, Person{
			Title:     &doc.Page2.Section2.Attorney2.Title,
			FirstName: &doc.Page2.Section2.Attorney2.FirstName,
			LastName:  &doc.Page2.Section2.Attorney2.LastName,
			DOB:       &doc.Page2.Section2.Attorney2.DOB,
			Address: PersonAddress{
				AddressLine1: doc.Page2.Section2.Attorney2.Address.Address1,
				AddressLine2: doc.Page2.Section2.Attorney2.Address.Address2,
				AddressLine3: doc.Page2.Section2.Attorney2.Address.Address3,
				Postcode:     doc.Page2.Section2.Attorney2.Address.Postcode,
			},
			Email: &doc.Page2.Section2.Attorney2.EmailAddress,
		})
	}

	// Page7, Section6 (PeopleToNotify)
	for _, personToNotify := range doc.Page7.Section6.PeopleToNotify {
		persons = append(persons, Person{
			Title:     &personToNotify.Title,
			FirstName: &personToNotify.FirstName,
			LastName:  &personToNotify.LastName,
			Address: PersonAddress{
				AddressLine1: personToNotify.Address.Address1,
				AddressLine2: personToNotify.Address.Address2,
				AddressLine3: personToNotify.Address.Address3,
				Postcode:     personToNotify.Address.Postcode,
			},
		})
	}

	// Page10, Section9 (Donor and Witness)
	if doc.Page10.Section9.Witness.FullName != "" {
		persons = append(persons, Person{
			FirstName: &doc.Page10.Section9.Witness.FullName,
			Address: PersonAddress{
				AddressLine1: doc.Page10.Section9.Witness.Address.Address1,
				AddressLine2: doc.Page10.Section9.Witness.Address.Address2,
				AddressLine3: doc.Page10.Section9.Witness.Address.Address3,
				Postcode:     doc.Page10.Section9.Witness.Address.Postcode,
			},
		})
	}

	// Page11, Section10 (Witness)
	if doc.Page11.Section10.FirstName != "" || doc.Page11.Section10.LastName != "" {
		persons = append(persons, Person{
			Title:     &doc.Page11.Section10.Title,
			FirstName: &doc.Page11.Section10.FirstName,
			LastName:  &doc.Page11.Section10.LastName,
			Address: PersonAddress{
				AddressLine1: doc.Page11.Section10.Address.Address1,
				AddressLine2: doc.Page11.Section10.Address.Address2,
				AddressLine3: doc.Page11.Section10.Address.Address3,
				Postcode:     doc.Page11.Section10.Address.Postcode,
			},
		})
	}

	// Page12, Section11 (Attorney and Witness)
	// if doc.Page12.Section11.Attorney.FirstName != "" || doc.Page12.Section11.Attorney.LastName != "" {
	// 	persons = append(persons, Person{
	// 		Title:     &doc.Page12.Section11.Attorney.Title,
	// 		FirstName: &doc.Page12.Section11.Attorney.FirstName,
	// 		LastName:  &doc.Page12.Section11.Attorney.LastName,
	// 	})
	// }

	// if doc.Page12.Section11.Witness.FullName != "" {
	// 	persons = append(persons, Person{
	// 		FirstName: &doc.Page12.Section11.Witness.FullName,
	// 		Address: PersonAddress{
	// 			AddressLine1: doc.Page12.Section11.Witness.Address.Address1,
	// 			AddressLine2: doc.Page12.Section11.Witness.Address.Address2,
	// 			AddressLine3: doc.Page12.Section11.Witness.Address.Address3,
	// 			Postcode:     doc.Page12.Section11.Witness.Address.Postcode,
	// 		},
	// 	})
	// }

	// Page17, Section12 (Attorney)
	for _, attorney := range doc.Page17.Section12.Attorney {
		persons = append(persons, Person{
			Title:     &attorney.Title,
			FirstName: &attorney.FirstName,
			LastName:  &attorney.LastName,
			DOB:       &attorney.DOB,
			Address: PersonAddress{
				AddressLine1: attorney.Address.Address1,
				AddressLine2: attorney.Address.Address2,
				AddressLine3: attorney.Address.Address3,
				Postcode:     attorney.Address.Postcode,
			},
		})
	}

	// Page18, Section13
	if doc.Page18.Section13.FirstName != "" || doc.Page18.Section13.LastName != "" {
		persons = append(persons, Person{
			Title:     &doc.Page18.Section13.Title,
			FirstName: &doc.Page18.Section13.FirstName,
			LastName:  &doc.Page18.Section13.LastName,
			Address: PersonAddress{
				AddressLine1: doc.Page18.Section13.Address.Address1,
				AddressLine2: doc.Page18.Section13.Address.Address2,
				AddressLine3: doc.Page18.Section13.Address.Address3,
				Postcode:     doc.Page18.Section13.Address.Postcode,
			},
			Email: &doc.Page18.Section13.EmailAddress,
		})
	}

	// Continuation Pages
	// for _, attorney := range doc.ContinuationPage1.ContinuationSheet1.Attorney {
	// 	persons = append(persons, Person{
	// 		FirstName: &attorney.FirstName,
	// 		LastName:  &attorney.LastName,
	// 		Address: PersonAddress{
	// 			AddressLine1: attorney.Address.Address1,
	// 			AddressLine2: attorney.Address.Address2,
	// 			AddressLine3: attorney.Address.Address3,
	// 			Postcode:     attorney.Address.Postcode,
	// 		},
	// 	})
	// }

	// if doc.ContinuationPage2.ContinuationSheet2.Donor.FullName != "" {
	// 	persons = append(persons, Person{
	// 		FirstName: &doc.ContinuationPage2.ContinuationSheet2.Donor.FullName,
	// 		Address: PersonAddress{
	// 			AddressLine1: doc.ContinuationPage2.ContinuationSheet2.Donor.Address.Address1,
	// 			AddressLine2: doc.ContinuationPage2.ContinuationSheet2.Donor.Address.Address2,
	// 			AddressLine3: doc.ContinuationPage2.ContinuationSheet2.Donor.Address.Address3,
	// 			Postcode:     doc.ContinuationPage2.ContinuationSheet2.Donor.Address.Postcode,
	// 		},
	// 	})
	// }

	// Page20 (Applicant)
	// for _, applicant := range doc.Page20.Section15.Applicant {
	// 	persons = append(persons, Person{
	// 		Address: PersonAddress{
	// 			AddressLine1: applicant.Address.Address1,
	// 			AddressLine2: applicant.Address.Address2,
	// 			AddressLine3: applicant.Address.Address3,
	// 			Postcode:     applicant.Address.Postcode,
	// 		},
	// 	})
	// }

	return persons
}
