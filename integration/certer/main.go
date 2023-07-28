package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"github.com/Netapp/harvest-automation/certer/models"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/carlmjohnson/requests"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

const (
	commonName    = "Harvest2"
	harvestUser   = "harvest"
	ontapRole     = "harvest"
	ontapRestRole = "harvest-rest"
)

var (
	username      = "admin"
	password      string
	ip            string
	adminSVM      string
	clientKeyDir  = "/opt/harvest/cert"
	clientKeyName = "u2"
	force         bool
)

func main() {
	utils.SetupLogging()
	parseCLI()
	begin()
}

func parseCLI() {
	flag.StringVar(&ip, "ip", "", "IP of ONTAP cluster (required)")
	flag.StringVar(&username, "username", "", "Username of ONTAP admin user (default=admin)")
	flag.StringVar(&password, "password", "", "Password of ONTAP admin user (required)")
	flag.StringVar(&clientKeyDir, "keydir", ".", "Directory to write cert files to")
	flag.StringVar(&clientKeyName, "keyname", "u2", "Prefix name to use for cert files")
	flag.BoolVar(&force, "force", false, "Always create certs even if the current ones are still valid")

	flag.Parse()
	if ip == "" {
		printRequired("ip")
	}
	if password == "" {
		printRequired("password")
	}
}

func printRequired(name string) {
	fmt.Printf("%s address is required\n", name)
	fmt.Printf("usage: \n")
	flag.PrintDefaults()
	os.Exit(1)
}

func begin() {
	log.Info().Str("ip", ip).Msg("Create certificates for ip")

	// Get admin SVM
	fetchAdminSVM()

	// Query for existing CA
	certificates, err := fetchCA()
	if err != nil {
		log.Error().Err(err).Msg("")
		return
	}

	// Check if these certs have expired
	if !force && certsAreFresh(certificates) {
		return
	}

	// Create private key and certificate signing request (CSR)
	csr, err := ensureOpenSSLInstalled()
	if err != nil {
		log.Error().Err(err).Msg("")
		return
	}

	// Delete existing
	if certificates.NumRecords > 0 {
		log.Info().
			Int("num", certificates.NumRecords).
			Str("common_name", commonName).
			Msg("Deleting matching certificates")
		err := deleteCertificates(certificates)
		if err != nil {
			log.Error().Err(err).Msg("failed to delete certificates")
			return
		}
	}

	// Create a root CA certificate that will be used to sign certificate requests for the user account(s)
	err = createRootCA()
	if err != nil {
		log.Error().Err(err).Msg("failed")
		return
	}

	// Sign the locally created certificate with the root CA generated above
	err = signCSR(csr)
	if err != nil {
		log.Error().Err(err).Msg("failed")
		return
	}

	// Add certificate auth to this ONTAP user
	err = addCertificateAuthToHarvestUser()
	if err != nil {
		log.Error().Err(err).Msg("")
	}

	fmt.Printf("Success! Test with:\n")
	fmt.Printf("curl --insecure --cert %s --key %s \"https://%s/api/cluster?fields=version\"\n",
		local(".crt"), local(".key"), ip)
	curlServer()
}

func sleep(s string) {
	duration, err := time.ParseDuration(s)
	if err != nil {
		log.Error().Err(err).Msg("failed to sleep")
	}
	log.Info().Str("sleep", s).Msg("sleep")
	time.Sleep(duration)
}

func curlServer() {
	if _, err := os.Stat(local(".crt")); errors.Is(err, os.ErrNotExist) {
		log.Panic().Str("crt", local(".crt")).Msg("does not exist")
	}
	for i := 0; i < 60; i++ {
		//nolint:gosec
		command := exec.Command("curl", "--insecure", "--cert", local(".crt"), "--key", local(".key"),
			fmt.Sprintf("https://%s/api/cluster?fields=version", ip))
		output, err := command.CombinedOutput()
		if err != nil {
			log.Error().Err(err).Str("output", string(output)).Msg("failed to exec curl")
		} else {
			fmt.Println(string(output))
			return
		}
		sleep("1s")
	}
}

func certsAreFresh(certificates models.Certificates) bool {
	cert := certificates.Records[0]
	date := cert.ExpiryTime.Format("2006-01-02")
	log.Info().Str("expire", date).Msg("Certificates are fresh. Done")
	return cert.ExpiryTime.After(time.Now().Add(8 * time.Hour))
}

func addCertificateAuthToHarvestUser() error {
	perms := []models.SecurityPermissions{
		{
			Application: "http",
			Role:        ontapRestRole,
			AuthMethod:  "cert",
			User:        harvestUser,
		},
		{
			Application: "ontapi",
			Role:        ontapRole,
			AuthMethod:  "cert",
			User:        harvestUser,
		},
	}
	for _, perm := range perms {
		p := perm
		err := newRequest().
			Pathf("/api/private/cli/security/login").
			BodyJSON(&p).
			Fetch(context.Background())

		if err != nil {
			var oe models.OntapError
			if errors.As(err, &oe) {
				if oe.StatusCode == 409 {
					// duplicate entry - that's fine, ignore
					continue
				}
			}
			return fmt.Errorf("failed to add cert auth to user=%s err=%w", harvestUser, err)
		}
	}
	return nil
}

func fetchCA() (models.Certificates, error) {
	var certificates models.Certificates
	err := newRequest().
		Pathf("/api/security/certificates").
		Param("common_name", commonName).
		Param("fields", "**").
		ToJSON(&certificates).
		Fetch(context.Background())
	if err != nil {
		return models.Certificates{}, err
	}
	return certificates, nil
}

func signCSR(csr string) error {
	certificates, err := fetchCA()
	if err != nil {
		return fmt.Errorf("failed to fetch CA err=%w", err)
	}

	ca := findRootCA(certificates)
	if ca == nil {
		return fmt.Errorf("unable to find CA")
	}
	// This is needed because you can't create a signing request with an expiry longer than the CA's expiry.
	// Use one day less than the number of days until the CA expires
	days := int(time.Until(ca.ExpiryTime).Hours()/24) - 1
	var signResponse models.SignResponse
	expiry := fmt.Sprintf("P%dDT", days)
	newCa := models.NewSignRequest{
		ExpiryTime:     expiry,
		SigningRequest: csr,
		HashFunction:   "sha256",
	}

	err = newRequest().
		Pathf("/api/security/certificates/%s/sign", ca.UUID).
		BodyJSON(&newCa).
		ToJSON(&signResponse).
		Fetch(context.Background())
	if err != nil {
		return fmt.Errorf("failed to create signed cert err=%w", err)
	}

	localCert := local(".crt")
	err = os.WriteFile(localCert, []byte(signResponse.PublicCertificate), 0600)
	if err != nil {
		return fmt.Errorf("failed to write %s err=%w", localCert, err)
	}
	return nil
}

func findRootCA(certificates models.Certificates) *models.Cert {
	for _, record := range certificates.Records {
		if record.Type == "root_ca" {
			return &record
		}
	}
	return nil
}

func ensureOpenSSLInstalled() (string, error) {
	_, err := exec.LookPath("openssl")
	if err != nil {
		return "", err
	}
	privateKey := local(".key")
	csr := local(".csr")
	command := exec.Command("openssl", "genrsa", "-out", privateKey, "2048")
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("err=%w output=%s", err, output)
	}
	log.Debug().Str("output", string(output)).Msg("created private key")
	// openssl req -days 3650 -sha256 -new -nodes -key cert/u2.key -subj /CN=harvest -out u2.csr

	command = exec.Command("openssl", "req", "-days", "3650", "-sha256", "-new", "-nodes", "-key", privateKey,
		"-subj", "/CN="+harvestUser, "-out", csr)
	output, err = command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error creating csr err=%w output=%s", err, output)
	}

	log.Debug().Str("output", string(output)).Msg("created csr")
	log.Info().Str("privateKey", privateKey).Msg("Created private key and certificate signing request (CSR)")

	data, err := os.ReadFile(csr)
	if err != nil {
		return "", fmt.Errorf("failed to read csr file=%s err=%w", csr, err)
	}

	return string(data), nil
}

func createRootCA() error {
	// 10 year expiry
	tenYears := fmt.Sprintf("P%dDT", 365*10)
	newCa := models.NewCA{
		Svm:        models.SVM{Name: adminSVM},
		Type:       "root-ca",
		CommonName: commonName,
		ExpiryTime: tenYears,
	}
	err := newRequest().
		Pathf("/api/security/certificates").
		BodyJSON(&newCa).
		Fetch(context.Background())
	if err != nil {
		return fmt.Errorf("failed to create root CA err=%w", err)
	}
	log.Info().Msg("Created Root CA")
	return nil
}

func fetchAdminSVM() {
	var svmResp models.SVMResponse
	err := newRequest().
		Pathf("/api/private/cli/vserver").
		Param("type", "admin").
		Param("fields", "type,uuid").
		ToJSON(&svmResp).
		Fetch(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("failed to fetch admin SVM")
		return
	}
	adminSVM = svmResp.Records[0].Vserver
}

func newRequest() *requests.Builder {
	return requests.
		URL(fmt.Sprintf("https://%s", ip)).
		BasicAuth(username, password).
		AddValidator(func(response *http.Response) error {
			if response.StatusCode >= 400 {
				var ontapErr models.OntapError
				//nolint:bodyclose
				err := requests.ToJSON(&ontapErr)(response)
				if err != nil {
					return err
				}
				ontapErr.StatusCode = response.StatusCode
				return ontapErr
			}
			return nil
		}).
		Client(newClient())
}

func deleteCertificates(certificates models.Certificates) error {
	// Three certificates are returned: server_ca, client_ca, root_ca
	for _, record := range certificates.Records {
		var resp string
		err := newRequest().
			Pathf("/api/security/certificates/%s", record.UUID).
			ToString(&resp).
			AddValidator(func(response *http.Response) error {
				if response.StatusCode != 200 {
					return fmt.Errorf("failed to delete ertificates. statusCode=%d status=%s", response.StatusCode, response.Status)
				}
				return nil
			}).
			Delete().
			Fetch(context.Background())

		if err != nil {
			return err
		}
	}
	return nil
}

func newClient() *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec
		},
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   2 * time.Minute,
	}

	return client
}

func local(ext string) string {
	return fmt.Sprintf("%s/%s%s", clientKeyDir, clientKeyName, ext)
}
