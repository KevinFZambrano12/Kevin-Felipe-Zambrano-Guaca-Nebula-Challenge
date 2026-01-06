# SSL Labs Scanner (Go)

This project is a **Go-based CLI tool** that consumes the **SSL Labs API v2** to analyze the TLS/SSL configuration of a public server and present **only the most relevant security information**, avoiding overly verbose output.

---

## Project Goal

The goal of this project is to:
- Perform an SSL/TLS analysis using SSL Labs
- Correctly handle the asynchronous nature of the assessment
- Display a **clear and concise security summary**, including:
  - SSL grade (A+, A, B, etc.)
  - Supported TLS protocols
  - Certificate information (CN and expiration date)
  - Critical vulnerabilities (Heartbleed, POODLE, Logjam)
  - Forward Secrecy support

The tool intentionally does **not** display the full SSL Labs response, which can be very large and difficult to interpret.

---

##How SSL Labs Works

SSL Labs assessments are **not instantaneous**:
- Tests are executed on Qualys servers
- A single assessment may take **60 seconds or more**
- Clients must:
  1. Start a new assessment
  2. Periodically poll for progress
  3. Retrieve the results once the assessment is complete

This project implements this workflow correctly.

---

## Application Flow

1. **User Input**
   - The user enters a hostname (e.g. `github.com`)
   - The hostname is validated to ensure:
     - No protocol is included (`https://`)
     - No paths are included (`/`)
     - The hostname format is valid
     - The hostname resolves via DNS

2. **Start a New Assessment**
   - A new assessment is initiated using:
     ```
     GET /analyze?host=example.com&startNew=on&all=done
     ```
   - The `startNew=on` parameter is used **only once**

3. **Polling for Status**
   - While the assessment status is `DNS` or `IN_PROGRESS`:
     - The program waits 10 seconds
     - The status is queried again
   - Polling stops when the status becomes:
     - `READY` → assessment completed successfully
     - `ERROR` → assessment failed

4. **Processing Results**
   - All discovered endpoints (IPs) are processed
   - For each endpoint with `statusMessage == "Ready"`, the tool extracts:
     - SSL grade
     - Supported TLS protocols
     - Certificate information
     - Forward Secrecy support
     - Known critical vulnerabilities

---

## Data Modeling Strategy

The SSL Labs API returns a very large JSON response.  
This project **models only the most relevant fields**, keeping the code simple and maintainable.

### Key Objects Used
- **Host**: overall assessment status
- **Endpoint**: individual IP address results
- **EndpointDetails**: essential security details (protocols, certificate, vulnerabilities)

This approach avoids unnecessary complexity while preserving security value.

---

## Displayed Security Information

For each successfully assessed endpoint, the tool displays:

- **IP address**
- **SSL Grade** (A+, A, A-, B, etc.)
- **Supported TLS protocols** (e.g. TLS 1.2, TLS 1.3)
- **Forward Secrecy support**
- **Critical vulnerabilities**
  - Heartbleed
  - POODLE
  - Logjam
- **Certificate details**
  - Common Name (CN)
  - Expiration date
