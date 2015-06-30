package kinesis

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	kin "github.com/aws/aws-sdk-go/service/kinesis"
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

	providers := make([]credentials.Provider)
	role := credentials.EC2RoleProvider{
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
		Endpoint:     "",
		ExpiryWindow: 0,
	}
	providers = append(providers, role)

	if k.config.AccessKeyID != "" && k.config.SecretAccessKey != "" {
		static := credentials.StaticProvider{
			AccessKeyID:     k.config.AccessKeyID,
			SecretAccessKey: k.config.SecretAccessKey,
		}
		providers = append(providers, static)
	}
	creds := credentials.NewChainCredentials(providers)
	conf := &aws.Config{
		Region:      k.config.Region,
		Credentials: creds,
	}
	k.Client = kin.New(conf)

	return nil
}

func (k *KinesisOutput) Run(or pipeline.OutputRunner, helper pipeline.PluginHelper) error {
	var (
		pack   *pipeline.PipelinePack
		msg    []byte
		pk     string
		err    error
		params *kin.PutRecordInput
	)

	if or.Encoder() == nil {
		return fmt.Errorf("Encoder required.")
	}

	for pack = range or.InChan() {
		msg, err = or.Encode(pack)
		if err != nil {
			or.LogError(fmt.Errorf("Error encoding message: %s", err))
			pack.Recycle()
			continue
		}
		pk = fmt.Sprintf("%d-%s", pack.Message.Timestamp, pack.Message.Hostname)
		if k.config.PayloadOnly {
			msg = pack.Message.GetPayload()
		}
		params = &kinesis.PutRecordInput{
			Data:         []byte(msg),
			PartitionKey: aws.String(pk),
			StreamName:   aws.String(k.config.Stream),
		}
		_, err = k.Client.PutRecord(params)
		if err != nil {
			or.LogError(fmt.Errorf("Error pushing message to Kinesis: %s", err))
			pack.Recycle()
			continue
		}
		pack.Recycle()
	}

	return nil
}

func init() {
	pipeline.RegisterPlugin("KinesisOutput", func() interface{} { return new(KinesisOutput) })
}
