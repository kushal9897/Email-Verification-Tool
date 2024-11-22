package main

import (
	"fmt"
	"net"
	"net/smtp"
	"regexp"
	"strings"
	"time"
)

// validateEmailFormat checks if the email format is valid.
func validateEmailFormat(email string) bool {
	// Define a regex pattern for basic email format validation
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	regex := regexp.MustCompile(pattern)
	return regex.MatchString(email)
}

// getMXRecords retrieves MX records for a given domain.
func getMXRecords(domain string) ([]*net.MX, error) {
	mxRecords, err := net.LookupMX(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch MX records for domain %s: %w", domain, err)
	}
	return mxRecords, nil
}

// smtpCheck validates the email address using SMTP handshake.
func smtpCheck(mxServer, email string) error {
	client, err := smtp.Dial(mxServer + ":25")
	if err != nil {
		return fmt.Errorf("SMTP dial error: %w", err)
	}
	defer client.Close()

	// Set a 5-second timeout for SMTP operations
	client.SetDeadline(time.Now().Add(5 * time.Second))

	// Say hello to the server
	if err := client.Hello("localhost"); err != nil {
		return fmt.Errorf("SMTP Hello error: %w", err)
	}

	// Start the MAIL FROM command
	if err := client.Mail("test@example.com"); err != nil {
		return fmt.Errorf("SMTP MAIL FROM error: %w", err)
	}

	// Check the RCPT TO command for the target email
	if err := client.Rcpt(email); err != nil {
		return fmt.Errorf("SMTP RCPT TO error: %w", err)
	}

	return nil
}

// verifyEmail performs complete email verification.
func verifyEmail(email string) (bool, error) {
	// Step 1: Validate email syntax
	if !validateEmailFormat(email) {
		return false, fmt.Errorf("invalid email format")
	}

	// Step 2: Extract the domain and fetch MX records
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false, fmt.Errorf("email address must contain a single '@'")
	}
	domain := parts[1]

	mxRecords, err := getMXRecords(domain)
	if err != nil {
		return false, fmt.Errorf("domain verification failed: %w", err)
	}

	// Step 3: Perform SMTP verification
	for _, mx := range mxRecords {
		fmt.Printf("Checking SMTP server: %s\n", mx.Host)
		err := smtpCheck(mx.Host, email)
		if err == nil {
			// If SMTP check succeeds for any MX server, email is valid
			return true, nil
		}
		fmt.Printf("Error with server %s: %v\n", mx.Host, err)
	}

	return false, fmt.Errorf("email could not be verified with available MX servers")
}

func main() {
	// Example email for testing
	email := "example@domain.com"

	// Start verification
	fmt.Printf("Verifying email: %s\n", email)
	valid, err := verifyEmail(email)
	if err != nil {
		fmt.Printf("Verification failed: %v\n", err)
	} else {
		fmt.Printf("Email is valid: %v\n", valid)
	}
}
