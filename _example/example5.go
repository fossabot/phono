package example

import (
	"github.com/dudk/phono"
	"github.com/dudk/phono/mixer"
	"github.com/dudk/phono/pipe"
	"github.com/dudk/phono/test"
	"github.com/dudk/phono/vst2"
	"github.com/dudk/phono/wav"
	vst2sdk "github.com/dudk/vst2"
)

// Example:
//		Read two .wav files
//		Mix them
// 		Process with vst2
//		Save result into new .wav file
//
// NOTE: For example both wav files have same characteristics i.e: sample rate, bit depth and number of channels.
// In real life implicit conversion will be needed.
func five() {
	bs := phono.BufferSize(512)

	// wav pump 1
	wavPump1, err := wav.NewPump(test.Data.Wav1, bs)
	check(err)

	// wav pump 2
	wavPump2, err := wav.NewPump(test.Data.Wav2, bs)
	check(err)

	sampleRate := wavPump1.WavSampleRate()
	// mixer
	mixer := mixer.New(bs, wavPump1.WavNumChannels())

	// track 1
	track1 := pipe.New(
		sampleRate,
		pipe.WithPump(wavPump1),
		pipe.WithSinks(mixer),
	)
	defer track1.Close()
	// track 2
	track2 := pipe.New(
		sampleRate,
		pipe.WithPump(wavPump2),
		pipe.WithSinks(mixer),
	)
	defer track2.Close()

	// vst2 processor
	vst2lib, err := vst2sdk.Open(test.Vst)
	check(err)
	defer vst2lib.Close()

	vst2plugin, err := vst2lib.Open()
	check(err)
	defer vst2plugin.Close()

	vst2processor := vst2.NewProcessor(
		vst2plugin,
		bs,
		wavPump1.WavSampleRate(),
		wavPump1.WavNumChannels(),
	)

	// wav sink
	wavSink, err := wav.NewSink(
		test.Out.Example5,
		wavPump1.WavSampleRate(),
		wavPump1.WavNumChannels(),
		wavPump1.WavBitDepth(),
		wavPump1.WavAudioFormat(),
	)
	check(err)

	// out pipe
	out := pipe.New(
		sampleRate,
		pipe.WithPump(mixer),
		pipe.WithProcessors(vst2processor),
		pipe.WithSinks(wavSink),
	)
	defer out.Close()

	// run all
	track1Done, err := track1.Begin(pipe.Run)
	check(err)
	track2Done, err := track2.Begin(pipe.Run)
	check(err)
	outDone, err := out.Begin(pipe.Run)
	check(err)

	// wait results
	err = track1.Wait(track1Done)
	check(err)
	err = track2.Wait(track2Done)
	check(err)
	err = out.Wait(outDone)
	check(err)
}
