# heka-plugins

## kinesis
This output will take your raw message and put it into a Kinesis stream.

Example configuration:

```ini
[KinesisOut]
type = "KinesisOutput"
region = "us-east-1"
stream = "foobar"
message_matcher = "TRUE"
```