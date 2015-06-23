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
	var contents []byte
	var err error

	for pack := range or.InChan() {
		if contents, err = json.Marshal(pack.Message); err != nil {
			or.LogError(err)
		} else {
			or.LogMessage(string(contents))
		}
		pack.Recycle()
	}

	return nil
}

func init() {
	pipeline.RegisterPlugin("KinesisOutput", func() interface{} { return new(KinesisOutput) })
}
