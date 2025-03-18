package music

import (
	"bytes"
	"io"
	"log"
	"math"
	"time"

	configs "github.com/dominik-merdzik/project-starbyte/configs"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"

	// embed the main track MP3 file (this will write the file contents to a byte slice at compile time)
	_ "embed"
)

//go:embed tracks/main_track.mp3
var mainTrack []byte

// PlayBackgroundMusicFromEmbed decodes, buffers, and plays the embedded MP3 file
func PlayBackgroundMusicFromEmbed(cfg configs.MusicConfig) {

	// only play if enabled in the config
	if !cfg.Enabled {
		return
	}

	// create a reader for the embedded MP3 data
	reader := bytes.NewReader(mainTrack)

	// wrap the reader to satisfy io.ReadCloser
	dec, format, err := mp3.Decode(io.NopCloser(reader))
	if err != nil {
		log.Println("Error decoding embedded music file:", err)
		return
	}
	// buffer the entire stream to allow reliable looping
	buf := beep.NewBuffer(format)
	buf.Append(dec)
	dec.Close()

	// obtain a streamer from the buffer
	streamer := buf.Streamer(0, buf.Len())
	// loop the buffered streamer indefinitely
	looped := beep.Loop(-1, streamer)

	// initialize the speaker
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	// set up volume control
	volControl := &effects.Volume{
		Streamer: looped,
		Base:     10, // base for dB conversion
		Volume:   0,  // 0 dB (full volume) by default
		Silent:   false,
	}

	if cfg.Volume <= 0 {
		volControl.Silent = true
	} else {
		// convert the [0,100] volume to decibels - 2.0 is a scaling factor to adjust the volume level to a comfortable range
		const volumeScale = 2.0
		volControl.Volume = volumeScale * math.Log10(float64(cfg.Volume)/100.0)
	}

	// play the volume-controlled streamer in the background
	go speaker.Play(volControl)
}
