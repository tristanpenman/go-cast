# Plan

Goal is to provide a desktop sender app that discovers Cast devices, shows available apps and their status for the selected device, and controls YouTube when the receiver exposes the YouTube app.

## Sender Implementation

1. ~~Extract sender protocol logic from `cmd/sender` into an internal sender package:~~
    - ~~Connect/authenticate to a receiver.~~
    - ~~Send `CONNECT`, `GET_STATUS`, `LAUNCH`, and app namespace messages.~~
    - ~~Track receiver status, running sessions, app IDs, transport IDs, and errors.~~
2. Merge discovery into the sender workflow:
    - Reuse the mDNS discovery command logic as a library API.
    - List nearby receivers with friendly name, model, host, port, and connection state.
    - Allow manual host entry for networks where discovery fails.
3. Define the app browsing model:
    - Show currently running receiver apps from `RECEIVER_STATUS`.
    - Show known launchable apps from a local app registry, starting with:
      - YouTube: `233637DE`
      - Chrome mirroring: `0F5096E8`
      - Android mirroring: `674A0243`
4. Add YouTube control support:
    - Parse common YouTube URLs into video IDs.
    - Launch YouTube if needed.
    - Wait for the YouTube transport ID.
    - Send `flingVideo` on `urn:x-cast:com.google.youtube.mdx`.
    - Add basic controls once namespace messages are confirmed against a real receiver: play/pause, seek, stop, and current video state.
5. Introduce Wails as the desktop shell:
    - Keep Go as the backend and expose discovery, connect, launch, and YouTube commands to the frontend.
    - Build a compact UI with device list, app list, receiver status, YouTube URL input, transport logs, and error state.
    - Keep the existing CLI sender as a testable protocol/debug entry point.
6. Validate against real devices:
    - Test discovery, status, YouTube launch, and YouTube playback on Chromecast/Google TV hardware.
    - Record receiver payloads needed to fill gaps in app status and media control.
    - Add integration notes and fixtures where real-device behavior differs from the local receiver.

## Local Video Sender

- Keep local video casting behind the CLI until the sender service layer is stable.
- Finish launch flow for the mirroring receiver app.
- Implement media transport only after sender UI can reliably select a device and launch an app.

## Receiver App

This project was originally developed around a custom Chromecast receiver implementation:

- ~~Implement enough receiver functionality to advertise an app~~
- ~~Launch apps~~
- ~~Properly handle CONNECT messages and transport logic~~
- ~~VP8 decoding~~
- ~~Display content using OpenGL~~
- ~~Backdrop and status~~
- ~~Receive and decrypt RTP stream~~
- ~~H.264 decoding~~
- Fix issues with initial session negotiation
- Handle multiple clients properly
- Add RTCP NACKs and reliability improvements

## Misc

- More protocol documentation
- Allow Cloudflare worker to be hosted locally
- Windows and macOS builds
  - Once Wails support has been implemented
