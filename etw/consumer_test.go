package etw

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConsumeEvents(t *testing.T) {
	var guid = "{A68CA8B7-004F-D7B6-A698-07E2DE0F1F5D}"
	guid = "{EDD08927-9CC4-4E65-B970-C2560FB5C289}"
	err := ConsumeEvents(guid)
	assert.NoError(t, err)
}









func TestConsumerReadEvents(t *testing.T) {
	//guid := "{EDD08927-9CC4-4E65-B970-C2560FB5C289}"
	//session := "TestGoSession"
	//consumer := NewConsumer(config.Config{})
	//err := consumer.EnableTrace(guid, session)
	//assert.NoError(t, err)
	//defer consumer.CloseTrace(session)
	//err = consumer.ReadEvents(consumer.Sessions[0])
	//time.Sleep(60 * time.Second)
	//assert.NoError(t, err)
	//defer consumer.CloseSession(consumer.Sessions[0].SessionHandle)
}
