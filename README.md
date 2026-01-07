# SSL Labs TLS Scanner (Go)

This project is a **command-line tool written in Go** that uses the **SSL Labs API v2** to evaluate the TLS/SSL security configuration of a given public domain.

The implementation focuses on **correct API usage, simplicity, and meaningful security output**, rather than reproducing the complete (and very large) SSL Labs report.

---

## Objective

The purpose of this project is to:

- Consume the SSL Labs API correctly
- Handle the **asynchronous assessment workflow**
- Extract and present **high-impact TLS security indicators**
- Keep the implementation simple and easy to reason about

This tool is designed as a technical exercise and not as a full replacement for the official SSL Labs web interface.

---

## How the SSL Labs API Works

SSL Labs assessments are **asynchronous**:

- Tests are executed remotely by Qualys
- An assessment can take **over a minute** to complete
- Clients are expected to:
    1. Start an assessment
    2. Poll its status periodically
    3. Retrieve the results once the status is `READY`

This project follows this workflow as described in the API documentation.

---

## Program Flow

1. **Input Validation**
    - The user provides a hostname (e.g. `github.com`)
    - The program validates that:
        - No protocol is included
        - No paths are included
        - The hostname has a valid format
        - The hostname resolves via DNS

2. **Assessment Initialization**
    - A new assessment is started using:
      ```
      GET /analyze?host=example.com&startNew=on&all=done
      ```
    - The `startNew=on` parameter is used only once to avoid unnecessary reassessments

3. **Polling**
    - While the assessment status is `DNS` or `IN_PROGRESS`:
        - The program waits 10 seconds
        - The status is queried again
    - Polling stops when the status becomes:
        - `READY` → assessment completed
        - `ERROR` → assessment failed

4. **Result Processing**
    - Each discovered endpoint (IP address) is processed
    - Only endpoints with `statusMessage == "Ready"` are considered
    - Relevant security information is extracted and displayed

---

## Data Modeling

The SSL Labs API returns a large and deeply nested JSON response.

This project **models only the fields required for analysis**, keeping the codebase minimal and maintainable.

### Main Structures

- **Host**: overall assessment status and metadata
- **Endpoint**: results per IP address
- **EndpointDetails**: TLS protocols, certificate data, and vulnerability flags

This selective modeling avoids unnecessary complexity while preserving security relevance.

---

## Displayed Security Information

For each successfully analyzed endpoint, the tool displays:

- IP address
- SSL Grade
- Supported TLS protocol versions
- Forward Secrecy support
- Critical vulnerabilities:
    - Heartbleed
    - POODLE
    - Logjam
- Certificate information:
    - Common Name (CN)
    - Expiration date
