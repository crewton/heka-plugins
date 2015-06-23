package kinesis

import (
	"encoding/json"
	"fmt"
	"github.com/AdRoll/goamz/aws"
	kin "github.com/AdRoll/goamz/kinesis"
	"github.com/mozilla-services/heka/pipeline"
	"time"
)

type KinesisOutput struct {
	auth   aws.Auth
	config *KinesisOutputConfig
	Client *kin.Kinesis
}

type KinesisOutputConfig struct {
	Region          string `toml:"region"`
	Stream          string `toml:"stream"`
	AccessKeyID     string `toml:"access_key_id"`
	SecretAccessKey string `toml:"secret_access_key"`
	Token           string `toml:"token"`
	PayloadOnly     bool   `toml:"payload_only"`
}

func (k *KinesisOutput) ConfigStruct() interface{} {
	return &KinesisOutputConfig{
		Region:          "us-east-1",
		Stream:          "",
		AccessKeyID:     "",
		SecretAccessKey: "",
		Token:           "",
	}
}

func (k *KinesisOutput) Init(config interface{}) error {
	k.config = config.(*KinesisOutputConfig)
	a, err := aws.GetAuth(k.config.AccessKeyID, k.config.SecretAccessKey, k.config.Token, time.Now())
	if err != nil {
		return fmt.Errorf("error authenticating: %s", err)
	}
	k.auth = a

	region, ok := aws.Regions[k.config.Region]
	if !ok {
		return fmt.Errorf("region does not exist: %s", k.config.Region)
	}

	k.Client = kin.New(k.auth, region)

	return nil
}

func (k *KinesisOutput) Run(or pipeline.OutputRunner, helper pipeline.PluginHelper) error {
	var (
		pack     *pipeline.PipelinePack
		contents []byte
		msg      []byte
		err      error
	)

	for pack = range or.InChan() {
		msg, err = or.Encode(pack)
		if err != nil {
			or.LogError(fmt.Errorf("Error encoding message: %s", err))
			pack.Recycle()
			continue
		}
		if contents, err = json.Marshal(msg); err != nil {
			or.LogError(fmt.Errorf("Error marshalling: %s", err))
			pack.Recycle()
			continue
		} else {
			pk := fmt.Sprintf("%s-%d", pack.Message.Timestamp, pack.Message.Hostname)
			_, err = k.Client.PutRecord(k.config.Stream, pk, contents, "", "")
			if err != nil {
				or.LogError(fmt.Errorf("Error pushing message to Kinesis: %s", err))
				pack.Recyle()
				continue
			}
		}
		pack.Recycle()
	}

	return nil
}

func init() {
	pipeline.RegisterPlugin("KinesisOutput", func() interface{} { return new(KinesisOutput) })
}
