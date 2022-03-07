# Plan

- Receiver
  + ~~Implement enough receiver functionality to advertise an app~~
  + ~~Launch apps~~
  + Receive and decrypt RTP stream
  + Forward video content to ffmpeg or mpv
  - Relay functionality
    + ~~Add relay mode command line argument~~
    + ~~Connect to a remote host when a connection is received~~
    + ~~Perform device authentication~~
    + ~~Forward messages from client to remote host~~
    + ~~Forward messages from remote host to client~~
    + Allow messages and content to be captured (MITM style)
- Sender
  + Allow sender to cast a video
  + Allow sender to cast screen (feasible in Go?)
  + ...
- Misc
  + Integrated video output
  + More protocol documentation
  + ...
