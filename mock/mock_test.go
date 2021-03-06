package mock_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/dudk/phono"
	"github.com/dudk/phono/mock"
	"github.com/dudk/phono/pipe"
)

var (
	tests = []struct {
		phono.BufferSize
		phono.NumChannels
		mock.Limit
		value    float64
		messages int64
		samples  int64
	}{
		{
			NumChannels: 1,
			Limit:       10,
			value:       0.5,
			messages:    10,
			samples:     100,
		},
		{
			NumChannels: 2,
			Limit:       100,
			value:       0.7,
			messages:    100,
			samples:     1000,
		},
	}
	bufferSize = phono.BufferSize(10)
	sampleRate = phono.SampleRate(10)
)

func TestPipe(t *testing.T) {
	pump := &mock.Pump{
		UID:        phono.NewUID(),
		Limit:      1,
		BufferSize: bufferSize,
	}
	processor := &mock.Processor{UID: phono.NewUID()}
	sink := &mock.Sink{UID: phono.NewUID()}
	p := pipe.New(
		sampleRate,
		pipe.WithName("Mock"),
		pipe.WithPump(pump),
		pipe.WithProcessors(processor),
		pipe.WithSinks(sink),
	)

	for _, test := range tests {
		p.Push(
			pump.LimitParam(test.Limit),
			pump.NumChannelsParam(test.NumChannels),
			pump.ValueParam(test.value),
		)
		err := pipe.Wait(p.Run())
		assert.Nil(t, err)

		messageCount, samplesCount := pump.Count()
		assert.Equal(t, test.messages, messageCount)
		assert.Equal(t, test.samples, samplesCount)

		messageCount, samplesCount = processor.Count()
		assert.Equal(t, test.messages, messageCount)
		assert.Equal(t, test.samples, samplesCount)

		messageCount, samplesCount = sink.Count()
		assert.Equal(t, test.messages, messageCount)
		assert.Equal(t, test.samples, samplesCount)

		assert.Equal(t, test.NumChannels, sink.Buffer.NumChannels())
		assert.Equal(t, phono.BufferSize(test.samples), sink.Buffer.Size())
	}
}

func TestComponentsReuse(t *testing.T) {
	pump := &mock.Pump{
		UID:         phono.NewUID(),
		Limit:       5,
		Interval:    10 * time.Microsecond,
		BufferSize:  10,
		NumChannels: 1,
	}
	proc1 := &mock.Processor{UID: phono.NewUID()}
	proc2 := &mock.Processor{UID: phono.NewUID()}
	sink1 := &mock.Sink{UID: phono.NewUID()}
	sink2 := &mock.Sink{UID: phono.NewUID()}

	p := pipe.New(
		sampleRate,
		pipe.WithName("Pipe"),
		pipe.WithPump(pump),
		pipe.WithProcessors(proc1, proc2),
		pipe.WithSinks(sink1, sink2),
	)

	err := pipe.Wait(p.Run())
	assert.Nil(t, err)
	err = pipe.Wait(p.Close())
	assert.Nil(t, err)
	p = pipe.New(
		sampleRate,
		pipe.WithName("Pipe"),
		pipe.WithPump(pump),
		pipe.WithProcessors(proc1, proc2),
		pipe.WithSinks(sink1, sink2),
	)
	err = pipe.Wait(p.Run())
	assert.Nil(t, err)
	err = pipe.Wait(p.Close())
	assert.Nil(t, err)
}
