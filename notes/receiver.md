# Receiver

This note documents deficiencies in the current receiver implementation. TLDR is that the receiver is incomplete enough that current Chrome mirroring is unlikely to produce a displayed frame.

## Codec negotiation

The main problem is codec negotiation. The receiver accepts every `video_source` offer without reading `codecName` or `rtpPayloadType` in [`internal/session/session.go`](internal/session/session.go#L137), yet it always initializes a VP8 decoder in [`internal/session/session.go`](internal/session/session.go#L385).

If Chrome selects H.264, VP9, or AV1, the receiver advertises acceptance and then feeds that stream into libvpx as VP8. This explains why streaming stopped working at some point after the original implementation.

## Correctness issues

- The RTP assembler accesses `packet.Payload[0]` and `packet.Payload[1]` without validating the payload length in [`internal/session/stream.go`](internal/session/stream.go#L33).
- Packet and frame ordering uses ordinary integer comparisons even though Cast frame IDs and RTP sequence numbers wrap around in [`internal/session/stream.go`](internal/session/stream.go#L38).
- One missing packet stalls the queue indefinitely. No NACK, loss recovery, or keyframe resynchronization is implemented.
- The answer declares RTCP event-log support, but the receiver does not implement it in [`internal/session/session.go`](internal/session/session.go#L142).
- RTCP parsing assumes every RTCP packet can first be parsed as RTP and identifies RTCP through the masked payload type `72` in [`internal/session/session.go`](internal/session/session.go#L261). This is a brittle RTP/RTCP multiplexing heuristic.
- The receiver report writes the RTP timestamp into `LastSenderReport`. That field should contain the middle 32 bits of the sender report's NTP timestamp.
- Every session binds fixed UDP port `50000`. A bind failure returns `nil`, which callers immediately dereference in [`internal/server/device.go`](internal/server/device.go#L144).
- Offer key and IV decoding errors are ignored, so malformed or changed encryption parameters can produce a nil decrypter or meaningless decode failures.
- The streaming answer omits constraints and display information. Those fields are optional but strongly recommended in the current [Cast streaming protocol](https://chromium.googlesource.com/openscreen/+/refs/heads/main/cast/protocol/streaming_session_protocol.md).

## How to identify the immediate failure from logs

- No `read ... bytes`: Chrome never started UDP. Investigate the `ANSWER`, selected codec and index, firewall, and advertised port.
- `read ... bytes`, followed by `stream not found`: selected SSRC or stream negotiation is wrong.
- `enqueued packet`, but no `frame` or `decoding frame`: the assembler, ordering, or loss handling is stuck.
- `decoding frame`, followed by `failed to decode buffer`: codec selection or AES/frame-counter handling is wrong.

## Recommended improvements

Parse the complete stream description, select exactly one codec that the receiver genuinely supports—currently VP8—and reject unsupported offers. Then add safe Cast packet parsing and frame collection keyed by frame ID instead of using a single strictly sequential queue.
