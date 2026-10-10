# Successful CBL login capture

Source: user-recorded Amazon traffic from `alexa auth login --record-dir`,
2026-10-10 09:10:12 UTC. The user completed browser authorization and reported
successful credential persistence. These three ordered HTTP 200 pairs cover
code creation, device registration, and access-token refresh in the NA region.

Sanitization replaced the serial with `ABCDEF1234567`, preserving the observed
13 uppercase alphanumeric character format and its identity across all three
requests. Public/private codes, access/refresh/session tokens, customer IDs,
request IDs, and age classification were replaced with fixed synthetic values.
Response headers were reduced to Content-Type, removing timestamps, request
identifiers, server metadata, and obsolete body lengths. JSON bodies were
re-encoded after sanitization; request fields and request headers are retained.

The observed device name is `go-alexa CLI`; the profile uses Client SDK 1.0,
iPhone device type A2IVLV5VM2W81, domain Device, model Client SDK, OS version 0,
and manufacturer Amazon on refresh. Replay matches every request exactly.
Negative controls reject changed or malformed serials, changed/empty device
names, changed device types, and additional device-name metadata on refresh.

This capture confirms this profile succeeded for one account at this time.
Negative matcher tests enforce the captured contract; they do not establish
Amazon's complete validation rules or prove that every different name is invalid.
Known failure response fixtures are separately labeled synthetic.
