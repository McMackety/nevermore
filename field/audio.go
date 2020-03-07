package field

import (
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"log"
	"os"
	"time"
)

var sounds map[string]Sound = make(map[string]Sound)

type Sound struct {
	Streamer beep.StreamSeekCloser
	Format beep.Format
}

func LoadWAVFile(url string) {
	f, err := os.Open(url)
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := wav.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	sounds[url] = Sound{
		Streamer: streamer,
		Format: format,
	}
}

func PlayWAV(url string, playFor time.Duration) {
	sound := sounds[url]

	speaker.Init(sound.Format.SampleRate, sound.Format.SampleRate.N(time.Second/10))
	done := make(chan bool)
	speaker.Play(beep.Seq(sound.Streamer, beep.Callback(func() {
		done <- true
	})))
	time.Sleep(playFor)
	<-done
}