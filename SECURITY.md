# Security Policy

## Reporting a vulnerability

Please report security vulnerabilities privately, using GitHub's private
vulnerability reporting: open the "Security" tab of this repository and
click "Report a vulnerability", or go directly to
<https://github.com/eoscloud/go-eoscloud/security/advisories/new>.
Do not open a public issue for a suspected vulnerability.

If you can't use GitHub's form, write to security@eoscloud.io.

## What to expect

- Acknowledgement within 7 days.
- An initial assessment (accepted or declined, and severity) within 14 days.
- A fix or mitigation and coordinated public disclosure, via a GitHub
  Security Advisory (with a CVE where applicable), within 90 days of the
  report, earlier once a fix is released.
- If a deadline can't be met, we'll tell you why and give you a new date.

You'll be credited in the advisory unless you'd rather stay anonymous.

## Supported versions

Only the latest release is supported. Fixes land there;
please upgrade before reporting an issue that may already be fixed.

## Scope

go-eoscloud is a client library for the eoscloud API. In scope, for
example: leaking credentials or tokens (in logs or error messages),
skipping or weakening TLS verification, sending requests or credentials to
an endpoint other than the configured one, and crashes or resource
exhaustion caused by API responses.

Out of scope:

- Vulnerabilities in the eoscloud platform or its API. Report those to
  security@eoscloud.io, not through GitHub.
- Vulnerabilities in third-party dependencies. Please
  report those upstream; we update affected dependencies once a fix is
  available.
