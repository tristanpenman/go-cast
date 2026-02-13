# Plan

- Receiver
  + ~~Implement enough receiver functionality to advertise an app~~
  + ~~Launch apps~~
  + ~~Properly handle CONNECT messages and transport logic~~
  + ~~VP8 decoding~~
  + ~~Display content using OpenGL~~
  + ~~Backdrop and status~~
  + ~~Receive and decrypt RTP stream~~
  + RTCP nacks and various reliability improvements
  * Fix issues with initial session negotiation
  * Handle multiple clients properly
  + H.264 Decoding
- Sender
  + Basic structure for sender (**in progress**)
  + Allow sender to cast a local video (**in progress**)
  + Allow sender to cast a YouTube video
  + Test with a real receiver
- Misc
  + More protocol documentation
  + Allow Cloudflare worker to be hosted locally
  + Set up linters
  + Set up GitHub actions
  + Windows and macOS builds
