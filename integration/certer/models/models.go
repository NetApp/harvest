package models

import (
	"fmt"
	"time"
)

type Cert struct {
	UUID                   string    `json:"uuid"`
	Scope                  string    `json:"scope"`
	Type                   string    `json:"type"`
	CommonName             string    `json:"common_name"`
	SerialNumber           string    `json:"serial_number"`
	Ca                     string    `json:"ca"`
	HashFunction           string    `json:"hash_function"`
	KeySize                int       `json:"key_size"`
	ExpiryTime             time.Time `json:"expiry_time"`
	PublicCertificate      string    `json:"public_certificate"`
	Name                   string    `json:"name"`
	AuthorityKeyIdentifier string    `json:"authority_key_identifier"`
	SubjectKeyIdentifier   string    `json:"subject_key_identifier"`
	Links                  struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"_links"`
}

type Certificates struct {
	Records    []Cert `json:"records"`
	NumRecords int    `json:"num_records"`
	Links      struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"_links"`
}

type SVMResponse struct {
	Records []struct {
		Vserver string `json:"vserver"`
		UUID    string `json:"uuid"`
		Type    string `json:"type"`
	} `json:"records"`
	NumRecords int `json:"num_records"`
}

type SVM struct {
	Name string `json:"name,omitempty"`
	UUID string `json:"uuid,omitempty"`
}

type NewCA struct {
	PrivateKey               string   `json:"private_key,omitempty"`
	IntermediateCertificates []string `json:"intermediate_certificates,omitempty"`
	ExpiryTime               string   `json:"expiry_time,omitempty"`
	CommonName               string   `json:"common_name,omitempty"`
	KeySize                  int      `json:"key_size,omitempty"`
	HashFunction             string   `json:"hash_function,omitempty"`
	Name                     string   `json:"name,omitempty"`
	PublicCertificate        string   `json:"public_certificate,omitempty"`
	Type                     string   `json:"type,omitempty"`
	Svm                      SVM      `json:"svm"`
}

type RootCA struct {
	SerialNumber string `json:"serial_number"`
	Links        struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"_links"`
	PrivateKey               string   `json:"private_key"`
	UUID                     string   `json:"uuid"`
	IntermediateCertificates []string `json:"intermediate_certificates"`
	ExpiryTime               string   `json:"expiry_time"`
	CommonName               string   `json:"common_name"`
	Scope                    string   `json:"scope"`
	AuthorityKeyIdentifier   string   `json:"authority_key_identifier"`
	KeySize                  int      `json:"key_size"`
	HashFunction             string   `json:"hash_function"`
	Name                     string   `json:"name"`
	Ca                       string   `json:"ca"`
	SubjectKeyIdentifier     string   `json:"subject_key_identifier"`
	PublicCertificate        string   `json:"public_certificate"`
	Type                     string   `json:"type"`
	Svm                      struct {
		Links struct {
			Self struct {
				Href string `json:"href"`
			} `json:"self"`
		} `json:"_links"`
		Name string `json:"name"`
		UUID string `json:"uuid"`
	} `json:"svm"`
}

type CreateRootCAResponse struct {
	Records []RootCA `json:"records"`
	Links   struct {
		Next struct {
			Href string `json:"href"`
		} `json:"next"`
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"_links"`
	NumRecords int `json:"num_records"`
}

type OErr struct {
	Message string `json:"message"`
	Code    string `json:"code"`
	Target  string `json:"target"`
}

type OntapError struct {
	Err        OErr `json:"error"`
	StatusCode int
}

func (o OntapError) Error() string {
	return fmt.Sprintf("message: %s code: %s", o.Err.Message, o.Err.Code)
}

type NewSignRequest struct {
	HashFunction   string `json:"hash_function,omitempty"`
	ExpiryTime     string `json:"expiry_time,omitempty"`
	SigningRequest string `json:"signing_request,omitempty"`
}

type SignResponse struct {
	PublicCertificate string `json:"public_certificate,omitempty"`
}

type Apps struct {
	AuthenticationMethods      []string `json:"authentication_methods,omitempty"`
	Application                string   `json:"application,omitempty"`
	SecondAuthenticationMethod string   `json:"second_authentication_method,omitempty"`
}

type Role struct {
	Name string `json:"name,omitempty"`
}

type PatchUser struct {
	Password     string `json:"password,omitempty"`
	Locked       bool   `json:"locked,omitempty"`
	Comment      string `json:"comment,omitempty"`
	Applications []Apps `json:"applications,omitempty"`
	Role         Role   `json:"role"`
}

type SecurityPermissions struct {
	Application string `json:"application,omitempty"`
	Role        string `json:"role,omitempty"`
	AuthMethod  string `json:"authentication_method,omitempty"`
	User        string `json:"user_or_group_name,omitempty"`
}
