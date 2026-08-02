package session

import "testing"

func TestSelectVP8VideoStreamSelectsFirstVP8VideoOffer(t *testing.T) {
	streams := []SupportedStream{
		{Index: 0, Type: "audio_source", CodecName: "opus", Ssrc: 10},
		{Index: 1, Type: "video_source", CodecName: "h264", Ssrc: 20},
		{Index: 2, Type: "video_source", CodecName: "vp8", Ssrc: 30, RtpPayloadType: 96},
		{Index: 3, Type: "video_source", CodecName: "vp8", Ssrc: 40, RtpPayloadType: 97},
	}

	selected := selectVP8VideoStream(streams)
	if selected == nil {
		t.Fatal("expected a VP8 video stream to be selected")
	}
	if selected.Index != 2 {
		t.Fatalf("selected stream index %d, want 2", selected.Index)
	}
}

func TestSelectVP8VideoStreamRejectsUnsupportedStreams(t *testing.T) {
	tests := []struct {
		name    string
		streams []SupportedStream
	}{
		{name: "empty offer"},
		{
			name: "unsupported video codec",
			streams: []SupportedStream{
				{Index: 0, Type: "video_source", CodecName: "h264"},
				{Index: 1, Type: "video_source", CodecName: "vp9"},
			},
		},
		{
			name: "VP8 audio stream",
			streams: []SupportedStream{
				{Index: 0, Type: "audio_source", CodecName: "vp8"},
			},
		},
		{
			name: "codec name is case sensitive",
			streams: []SupportedStream{
				{Index: 0, Type: "video_source", CodecName: "VP8"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if selected := selectVP8VideoStream(test.streams); selected != nil {
				t.Fatalf("unexpectedly selected stream %#v", selected)
			}
		})
	}
}
