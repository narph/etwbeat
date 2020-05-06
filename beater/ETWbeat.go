package beater

import (
	"fmt"
	"sync"

	"github.com/narph/etwbeat/etw"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"

	"github.com/narph/etwbeat/config"
)

type ETWbeat struct {
	beat     *beat.Beat       // Common beat information.
	config   config.ETWConfig // Configuration settings.
	consumer etw.Consumer
	done     chan struct{} // Channel to initiate shutdown of main event loop.
	pipeline beat.Pipeline // Interface to publish event.
}

// New creates an instance of etwbeat.
func New(b *beat.Beat, conf *common.Config) (beat.Beater, error) {
	var c config.ETWConfig
	err := b.BeatConfig.Unpack(&c)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file: %v", err)
	}
	consumer, err := etw.NewConsumer(conf, b.Info)
	if err != nil {
		return nil, fmt.Errorf("error initializing the consumer: %v", err)
	}
	eb := &ETWbeat{
		beat:     b,
		config:   c,
		done:     make(chan struct{}),
		consumer: *consumer,
	}
	return eb, nil
}

// Run starts etwbeat.
func (bt *ETWbeat) Run(b *beat.Beat) error {
	bt.pipeline = b.Publisher
	// setup global event ACK handler
	err := bt.pipeline.SetACKHandler(beat.PipelineACKHandler{
		//ACKEvents: acker.ACKEvents,
	})
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	for _, session := range bt.config.Sessions {

		// Start a goroutine for each event log.
		wg.Add(1)
		//go eb.processEventLog(&wg, provider, state, acker)
		go bt.process(&wg, session, &bt.consumer)
	}
	wg.Wait()
	//defer bt.checkpoint.Shutdown()

	//if eb.config.ShutdownTimeout > 0 {
	//	logp.Info("Shutdown will wait max %v for the remaining %v events to publish.",
	//		eb.config.ShutdownTimeout, acker.Active())
	//	ctx, cancel := context.WithTimeout(context.Background(), eb.config.ShutdownTimeout)
	//	defer cancel()
	//	acker.Wait(ctx)
	//}
	return nil
}

// Stop stops etwbeat.
func (etb *ETWbeat) Stop() {
	logp.Info("Stopping ETWbeat")
	if etb.done != nil {
		close(etb.done)
	}
}

func (eb *ETWbeat) process(
	wg *sync.WaitGroup,
	session config.Session,
	consumer *etw.Consumer,
) {
	defer wg.Done()
	consumer.Run(eb.done, eb.pipeline, session)
}
