# Bugs

Issues identified by gpt-5.6-sol.

## Main Findings

1. **Concurrent sends can corrupt Cast framing.**

   `internal/transport/channel.go` performs separate header and payload writes in `CastChannel.Send` without synchronization. Calls from different goroutines can interleave and produce invalid frames. Protect the complete framed write with a per-channel mutex.

2. **Incoming frame lengths are unbounded.**

   `internal/transport/channel.go` trusts the peer-supplied 32-bit message length and immediately allocates a buffer of that size. A malformed or malicious peer can request an allocation approaching 4 GiB. Enforce a reasonable maximum Cast-message size before allocating.

3. **Launching two different mirroring applications can panic.**

   Every session binds UDP port `50000` in `internal/session/session.go`. Starting another session makes `NewSession` return `nil`, after which `internal/server/device.go` immediately calls `activeSession.Start()`. Prefer an ephemeral port, return `(*Session, error)`, and propagate the failure.

4. **Malformed RTP payloads can crash the receiver.**

   `internal/session/stream.go` indexes `Payload[1]` without first checking the payload length. It later slices using an extension-derived offset without confirming that the offset is within the payload. Treat UDP input as untrusted and validate every index, field length, and slice boundary.

5. **Malformed encryption parameters can cause a nil-pointer panic.**

   AES key and IV decoding errors are ignored in `internal/session/session.go`. `NewDecrypter` can consequently return `nil`, which the decode closure later dereferences. Reject invalid hex, AES key sizes, and IV lengths while processing the offer.

6. **Receiver state is not concurrency-safe.**

   `Device.Sessions`, `transports`, subscriptions, and `nextPid` in `internal/server/device.go` are read and modified by per-client goroutines without synchronization. Multiple connected senders can cause data races or concurrent-map panics.

7. **Sender waits can return stale responses.**

   `RequestStatus` and `RequestAppAvailability` do not reset or version their stored results. `WaitForStatus` and `WaitForAvailability` only wait for a non-nil value, so after the first response, later request/wait pairs can immediately return old data. Correlate responses with request IDs or track response generations.

## Smaller deficiencies

- `DownloadManifest` in `internal/common/manifest.go` never closes its HTTP response body and uses `http.Get` without a timeout.
- `CastFeedback.Unmarshal` in `internal/session/feedback.go` accepts an eight-byte minimum but reads through byte twelve, allowing short input to panic.
- `ParseYouTubeVideoID` in `internal/client/youtube.go` accepts malformed IDs such as `https://youtu.be/foo/bar`; validate extracted IDs.
- Stopped sessions are removed from `Device.Sessions`, but their entries remain in `device.transports`, leaving stale transports and subscriptions.

## Tests

At the time of review:

- `go test ./...` passed.
- `go vet ./...` passed.
- Race-enabled tests passed for `internal/client`, `internal/session`, `internal/server`, and `internal/discovery`.

The existing tests do not substantially exercise the network and concurrent paths described above. Test coverage is particularly sparse in `transport`, `session`, and `server`.
